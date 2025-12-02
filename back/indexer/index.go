package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"strconv"
	"strings"
	"time"

	"chainqa_offchain_demo/models"

	"chainmaker.org/chainmaker/pb-go/v2/common"
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
	"github.com/go-redis/redis/v8"
)

// IndexerService 索引服务核心结构
type IndexerService struct {
	chainClient *sdk.ChainClient
	redisClient *redis.Client
	ctx         context.Context
}

// GlobalIndexerService 全局索引服务实例
var GlobalIndexerService *IndexerService

// Global Bucket Config
const (
	AgeBucketSize   = 10  // 年龄分桶大小，[0-10) -> Bucket 0
	HashBucketNum   = 100 // 字符串哈希分桶模数
	GenderBucketNum = 2   // 性别分桶模数
)

// StartBlockListener 启动监听并处理索引更新
func (s *IndexerService) StartBlockListener() {
	// 1. 订阅区块事件 (SDK调用)
	// startBlock: -1 代表从最新区块开始，也可指定高度回放
	eventChan, err := s.chainClient.SubscribeBlock(s.ctx, -1, -1, true, false)
	if err != nil {
		log.Fatalf("Failed to subscribe to block: %v", err)
	}

	log.Println("Start listening to ChainMaker blocks...")

	for {
		select {
		case blockInfo, ok := <-eventChan:
			if !ok {
				log.Println("Block channel closed, attempting reconnect...")
				// 重连逻辑：等待一段时间后重新订阅
				time.Sleep(5 * time.Second)
				eventChan, err = s.chainClient.SubscribeBlock(s.ctx, -1, -1, true, false)
				if err != nil {
					log.Printf("Failed to reconnect: %v", err)
					return
				}
				log.Println("Reconnected to block subscription")
				continue
			}
			// 强转为 BlockInfo 结构 (视SDK版本具体实现而定)
			blk, ok := blockInfo.(*common.BlockInfo)
			if !ok || blk == nil {
				log.Printf("Invalid block type or nil block, skipping...")
				continue
			}
			s.processBlock(blk)
		case <-s.ctx.Done():
			return
		}
	}
}

// processBlock 处理单个区块，构建三层索引
func (s *IndexerService) processBlock(block *common.BlockInfo) {
	blockHeight := block.Block.Header.BlockHeight
	txs := block.Block.Txs

	// 使用 Redis Pipeline 保证写入性能，但不保证严格的跨Key事务原子性(Cluster模式下)，
	// 但索引构建通常是幂等的，可重试。
	pipe := s.redisClient.Pipeline()

	// 缓存当前区块涉及的数据域和桶，避免重复 SetBit
	activeDomains := make(map[string]bool)
	activeBuckets := make(map[string]bool) // Key format: "field:bucketID"
	hashIndexCache := make(map[string]map[string][]string)

	for _, tx := range txs {
		// 0. 过滤不相关的交易：只处理医疗数据相关的交易
		if !isMedicalDataTx(tx) {
			continue
		}

		// 1. 解析交易 Payload 获取医疗数据
		// 假设 Payload 是 JSON 序列化的 OnChainRecord
		record, err := parseTxPayload(tx)
		if err != nil {
			log.Printf("Skip invalid tx %s: %v", tx.Payload.TxId, err)
			continue
		}

		// 2. 收集 Layer 1 (数据域) 信息
		activeDomains[record.Metadata.DomainID] = true

		// 3. 收集 Layer 2 (字段分桶) 信息 & 构建 Layer 3 (区块内索引)

		// --- 处理 Age (范围型/数值型) ---
		ageBucket := record.Metadata.Age / AgeBucketSize
		activeBuckets[fmt.Sprintf("age:%d", ageBucket)] = true

		// Layer 3: 区块内 B+树 (Redis ZSet)
		zsetKey := fmt.Sprintf("idx:blk:%d:age:zset", blockHeight)
		pipe.ZAdd(s.ctx, zsetKey, &redis.Z{
			Score:  float64(record.Metadata.Age),
			Member: record.TxID,
		})

		// --- 处理 DiseaseCode\Name\Gender\Hospital\Department\Uid\DomainID (等值型/字符型) ---
		diseaseBucket := hashBucket(record.Metadata.DiseaseCode)
		nameBucket := hashBucket(record.Metadata.Name)
		genderBucket := hashBucket(record.Metadata.Gender)
		hospitalBucket := hashBucket(record.Metadata.Hospital)
		departmentBucket := hashBucket(record.Metadata.Department)
		uidBucket := hashBucket(record.Metadata.Uid)
		domainIDBucket := hashBucket(record.Metadata.DomainID)
		activeBuckets[fmt.Sprintf("disease:%d", diseaseBucket)] = true
		activeBuckets[fmt.Sprintf("name:%d", nameBucket)] = true
		activeBuckets[fmt.Sprintf("gender:%d", genderBucket)] = true
		activeBuckets[fmt.Sprintf("hospital:%d", hospitalBucket)] = true
		activeBuckets[fmt.Sprintf("department:%d", departmentBucket)] = true
		activeBuckets[fmt.Sprintf("uid:%d", uidBucket)] = true
		activeBuckets[fmt.Sprintf("domainID:%d", domainIDBucket)] = true
		// Layer 3: 区块内哈希索引 (Redis Hash)
		// 为了支持 "Find Tx by Disease\Name\Gender\Hospital\Department\Uid\DomainID", 结构应为 Hash: Key=Code, Value=List[TxID]
		attributeValues := map[string]string{
			"disease":    record.Metadata.DiseaseCode,
			"name":       record.Metadata.Name,
			"gender":     record.Metadata.Gender,
			"hospital":   record.Metadata.Hospital,
			"department": record.Metadata.Department,
			"uid":        record.Metadata.Uid,
			"domainID":   record.Metadata.DomainID,
		}
		for attr, val := range attributeValues {
			hashKey := fmt.Sprintf("idx:blk:%d:%s:hash", blockHeight, attr)
			updateBlockHashIndex(hashIndexCache, hashKey, val, record.TxID)
		}
	}

	// 将缓存的哈希索引写入 Redis
	for hashKey, fieldMap := range hashIndexCache {
		for field, txIDs := range fieldMap {
			if len(txIDs) == 0 {
				continue
			}
			valueBytes, err := json.Marshal(txIDs)
			if err != nil {
				log.Printf("Failed to marshal txIDs for key %s field %s: %v", hashKey, field, err)
				continue
			}
			pipe.HSet(s.ctx, hashKey, field, valueBytes)
		}
	}

	// 4. 提交 Layer 1 (Data Domain Bitmap) 更新
	for domainID := range activeDomains {
		key := fmt.Sprintf("idx:domain:%s", domainID)
		pipe.SetBit(s.ctx, key, int64(blockHeight), 1)
	}

	// 5. 提交 Layer 2 (Field Bucket Bitmap) 更新
	for bucketKey := range activeBuckets {
		key := fmt.Sprintf("idx:bucket:%s", bucketKey)
		pipe.SetBit(s.ctx, key, int64(blockHeight), 1)
	}

	// 执行 Redis 批量操作
	_, err := pipe.Exec(s.ctx)
	if err != nil {
		log.Printf("Error updating index for block %d: %v", blockHeight, err)
		// 异常处理：此处应加入重试队列或报警
	} else {
		log.Printf("Indexed block %d successfully", blockHeight)
	}
}

// 辅助函数：判断交易是否与医疗数据相关
// 判断标准：针对 updateDataDigtalEnvelopWithDomain 方法
// 该方法接收 envelopJsonStr 参数（JSON字符串），其中包含 domainID 字段
// domainID 的值必须以 DOMAIN_ 开头
func isMedicalDataTx(tx *common.Transaction) bool {
	if tx.Payload == nil || len(tx.Payload.Parameters) == 0 {
		return false
	}

	// 查找 envelopJsonStr 或 envelop 参数（updateDataDigtalEnvelopWithDomain 的特征参数）
	var envelopJsonStr string
	for _, param := range tx.Payload.Parameters {
		// 检查参数名（可能是 envelopJsonStr 或 envelop，取决于实际调用）
		if param.Key == "envelopJsonStr" || param.Key == "envelop" {
			envelopJsonStr = string(param.Value)
			break
		}
	}

	// 如果没有找到 envelopJsonStr 参数，检查是否有直接的 domainID 参数
	if envelopJsonStr == "" {
		for _, param := range tx.Payload.Parameters {
			if param.Key == "domainID" {
				domainID := string(param.Value)
				// 检查 domainID 是否以 DOMAIN_ 开头
				if strings.HasPrefix(domainID, "DOMAIN_") {
					return true
				}
			}
		}
		return false
	}

	// 解析 envelopJsonStr JSON 字符串，检查其中的 domainID
	type DataDigtalEnvelop struct {
		DomainID string `json:"domainID"`
	}

	var dataEnvelop DataDigtalEnvelop
	err := json.Unmarshal([]byte(envelopJsonStr), &dataEnvelop)
	if err != nil {
		// 如果解析失败，可能不是医疗数据交易
		return false
	}

	// 检查 domainID 是否以 DOMAIN_ 开头
	if strings.HasPrefix(dataEnvelop.DomainID, "DOMAIN_") {
		return true
	}

	return false
}

// 辅助函数：解析交易
func parseTxPayload(tx *common.Transaction) (*models.OnChainRecord, error) {
	// 实际项目中需根据合约调用的参数结构解析
	// updateDataDigtalEnvelopWithDomain 方法接收 envelopJsonStr 参数（JSON字符串）

	// 方法1: 优先处理 updateDataDigtalEnvelopWithDomain 的参数
	// 查找 envelopJsonStr 或 envelop 参数
	for _, param := range tx.Payload.Parameters {
		if param.Key == "envelopJsonStr" || param.Key == "envelop" {
			envelopJsonStr := string(param.Value)

			// 解析 JSON 字符串为 DataDigtalEnvelop 结构
			type DataDigtalEnvelop struct {
				Uid         string `json:"uId"`
				TimeStamp   string `json:"timeStamp"`
				Pos         string `json:"pos"`
				Envelop     string `json:"envelop"`
				Name        string `json:"name"`
				Age         int    `json:"age"`
				Gender      string `json:"gender"`
				Hospital    string `json:"hospital"`
				Department  string `json:"department"`
				DiseaseCode string `json:"diseaseCode"`
				DomainID    string `json:"domainID"`
			}

			var dataEnvelop DataDigtalEnvelop
			err := json.Unmarshal([]byte(envelopJsonStr), &dataEnvelop)
			if err != nil {
				return nil, fmt.Errorf("failed to parse envelopJsonStr: %v", err)
			}

			// 转换为 OnChainRecord
			record := &models.OnChainRecord{
				TxID: tx.Payload.TxId,
				Metadata: models.RecordMetadata{
					Uid:         dataEnvelop.Uid,
					TimeStamp:   dataEnvelop.TimeStamp,
					Pos:         dataEnvelop.Pos,
					Envelop:     dataEnvelop.Envelop,
					Name:        dataEnvelop.Name,
					Age:         dataEnvelop.Age,
					Gender:      dataEnvelop.Gender,
					Hospital:    dataEnvelop.Hospital,
					Department:  dataEnvelop.Department,
					DiseaseCode: dataEnvelop.DiseaseCode,
					DomainID:    dataEnvelop.DomainID,
				},
			}

			// 验证 domainID 是否存在
			if record.Metadata.DomainID == "" {
				return nil, fmt.Errorf("domainID is empty in tx %s", tx.Payload.TxId)
			}

			return record, nil
		}
	}

	// 方法2: 从交易参数解析（如果合约直接传递了元数据）
	for _, param := range tx.Payload.Parameters {
		if param.Key == "medical_data" || param.Key == "metadata" {
			var data models.OnChainRecord
			err := json.Unmarshal(param.Value, &data)
			if err == nil && data.TxID == "" {
				data.TxID = tx.Payload.TxId
			}
			return &data, err
		}
	}

	// 方法3: 尝试从其他参数构建（兼容旧格式）
	record := &models.OnChainRecord{
		TxID:     tx.Payload.TxId,
		Metadata: models.RecordMetadata{},
	}

	for _, param := range tx.Payload.Parameters {
		switch param.Key {
		case "domainID":
			record.Metadata.DomainID = string(param.Value)
		case "age":
			if age, err := strconv.Atoi(string(param.Value)); err == nil {
				record.Metadata.Age = age
			}
		case "diseaseCode":
			record.Metadata.DiseaseCode = string(param.Value)
		case "gender":
			record.Metadata.Gender = string(param.Value)
		case "name":
			record.Metadata.Name = string(param.Value)
		case "uId":
			record.Metadata.Uid = string(param.Value)
		case "pos":
			record.Metadata.Pos = string(param.Value)
		case "envelop":
			record.Metadata.Envelop = string(param.Value)
		case "hospital":
			record.Metadata.Hospital = string(param.Value)
		case "department":
			record.Metadata.Department = string(param.Value)
		case "timeStamp":
			record.Metadata.TimeStamp = string(param.Value)
		}
	}

	// 如果至少有一个元数据字段，认为解析成功
	if record.Metadata.DomainID != "" {
		return record, nil
	}

	return nil, fmt.Errorf("no medical data found in tx %s", tx.Payload.TxId)
}

// 辅助函数：字符串哈希分桶
func hashBucket(val string) int {
	h := fnv.New32a()
	h.Write([]byte(val))
	return int(h.Sum32()) % HashBucketNum
}

// 辅助函数：缓存区块内Hash索引，待批量写入
func updateBlockHashIndex(cache map[string]map[string][]string, key, field, txID string) {
	if key == "" || field == "" || txID == "" {
		return
	}
	if _, ok := cache[key]; !ok {
		cache[key] = make(map[string][]string)
	}
	cache[key][field] = append(cache[key][field], txID)
}

// SearchRequest 查询请求参数
type SearchRequest struct {
	DomainID    string // 必须：数据域
	AgeStart    int    // 范围查询条件
	AgeEnd      int
	DiseaseCode string // 等值查询条件
	Name        string // 等值查询条件
	Gender      string // 等值查询条件
	Hospital    string // 等值查询条件
	Department  string // 等值查询条件
	Uid         string // 等值查询条件
}

// SearchResult 返回结果
type SearchResult struct {
	TxIDs []string `json:"tx_ids"`
}

// ExecuteQuery 执行多层索引查询
func (s *IndexerService) ExecuteQuery(req SearchRequest) (*SearchResult, error) {
	// --- 阶段一：粗粒度筛选 (Layer 1 & 2) ---
	// 1. 获取数据域位图 Key
	domainKey := fmt.Sprintf("idx:domain:%s", req.DomainID)

	// 2. 计算属性分桶位图 Keys
	// 2.1 Age (Range): 涉及多个桶
	var ageBucketKeys []string
	if req.AgeStart > 0 || req.AgeEnd > 0 {
		startBucket := req.AgeStart / AgeBucketSize
		endBucket := req.AgeEnd / AgeBucketSize
		for i := startBucket; i <= endBucket; i++ {
			ageBucketKeys = append(ageBucketKeys, fmt.Sprintf("idx:bucket:age:%d", i))
		}
	}

	// 2.2 收集所有等值查询字段的分桶位图 Keys
	var equalityBucketKeys []string
	if req.DiseaseCode != "" {
		equalityBucketKeys = append(equalityBucketKeys, fmt.Sprintf("idx:bucket:disease:%d", hashBucket(req.DiseaseCode)))
	}
	if req.Name != "" {
		equalityBucketKeys = append(equalityBucketKeys, fmt.Sprintf("idx:bucket:name:%d", hashBucket(req.Name)))
	}
	if req.Gender != "" {
		equalityBucketKeys = append(equalityBucketKeys, fmt.Sprintf("idx:bucket:gender:%d", hashBucket(req.Gender)))
	}
	if req.Hospital != "" {
		equalityBucketKeys = append(equalityBucketKeys, fmt.Sprintf("idx:bucket:hospital:%d", hashBucket(req.Hospital)))
	}
	if req.Department != "" {
		equalityBucketKeys = append(equalityBucketKeys, fmt.Sprintf("idx:bucket:department:%d", hashBucket(req.Department)))
	}
	if req.Uid != "" {
		equalityBucketKeys = append(equalityBucketKeys, fmt.Sprintf("idx:bucket:uid:%d", hashBucket(req.Uid)))
	}

	// 3. 位运算 (BITOP)
	// 逻辑：Result = DomainBitmap AND (AgeBucket1 OR AgeBucket2...) AND (EqualityBucket1 AND EqualityBucket2...)
	// Redis BITOP 支持 AND, OR, XOR, NOT
	var tempKeys []string
	defer func() {
		// 清理所有临时Key
		for _, key := range tempKeys {
			s.redisClient.Del(s.ctx, key)
		}
	}()

	// 3.1 合并 Age Buckets (OR) - 如果有年龄条件
	var currentResultKey string
	if len(ageBucketKeys) > 0 {
		ageUnionKey := "tmp:search:age_union:" + generateRequestID()
		tempKeys = append(tempKeys, ageUnionKey)
		_, err := s.redisClient.BitOpOr(s.ctx, ageUnionKey, ageBucketKeys...).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to union age buckets: %v", err)
		}
		currentResultKey = ageUnionKey
	} else {
		// 如果没有年龄条件，从domainKey开始
		currentResultKey = domainKey
	}

	// 3.2 合并所有等值查询字段的桶 (AND)
	for _, bucketKey := range equalityBucketKeys {
		nextKey := "tmp:search:and_" + generateRequestID()
		tempKeys = append(tempKeys, nextKey)
		_, err := s.redisClient.BitOpAnd(s.ctx, nextKey, currentResultKey, bucketKey).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to AND bucket: %v", err)
		}
		// 如果currentResultKey是临时key，需要清理
		if currentResultKey != domainKey && strings.HasPrefix(currentResultKey, "tmp:search:") {
			s.redisClient.Del(s.ctx, currentResultKey)
		}
		currentResultKey = nextKey
	}

	// 3.3 最终与Domain位图合并 (如果还没有合并)
	if currentResultKey != domainKey {
		finalDestKey := "tmp:search:final:" + generateRequestID()
		tempKeys = append(tempKeys, finalDestKey)
		_, err := s.redisClient.BitOpAnd(s.ctx, finalDestKey, domainKey, currentResultKey).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to final AND: %v", err)
		}
		// 清理中间结果
		if currentResultKey != domainKey && strings.HasPrefix(currentResultKey, "tmp:search:") {
			s.redisClient.Del(s.ctx, currentResultKey)
		}
		currentResultKey = finalDestKey
	}

	// 4. 提取位图中为 1 的位置 (即 Block Heights)
	candidateBlocks, err := s.getSetBits(currentResultKey)
	if err != nil {
		return nil, err
	}

	log.Printf("Phase 1 filtered down to %d blocks", len(candidateBlocks))

	// --- 阶段二：细粒度定位 (Layer 3) ---
	var resultTxIDs []string
	for _, height := range candidateBlocks {
		// 收集该区块内满足所有条件的交易ID列表
		var candidateTxLists [][]string

		// 1. 查询 Age ZSet (B+ Tree) - 如果有年龄条件
		if req.AgeStart > 0 || req.AgeEnd > 0 {
			zsetKey := fmt.Sprintf("idx:blk:%d:age:zset", height)
			txsByAge, err := s.redisClient.ZRangeByScore(s.ctx, zsetKey, &redis.ZRangeBy{
				Min: strconv.Itoa(req.AgeStart),
				Max: strconv.Itoa(req.AgeEnd),
			}).Result()
			if err == nil && len(txsByAge) > 0 {
				candidateTxLists = append(candidateTxLists, txsByAge)
			} else {
				// 如果年龄条件不满足，跳过该区块
				continue
			}
		}

		// 2. 查询所有等值查询字段的 Hash 索引
		equalityFields := []struct {
			field string
			value string
		}{
			{"disease", req.DiseaseCode},
			{"name", req.Name},
			{"gender", req.Gender},
			{"hospital", req.Hospital},
			{"department", req.Department},
			{"uid", req.Uid},
			{"domainID", req.DomainID},
		}

		for _, fieldInfo := range equalityFields {
			if fieldInfo.value == "" {
				continue
			}
			hashKey := fmt.Sprintf("idx:blk:%d:%s:hash", height, fieldInfo.field)
			val, err := s.redisClient.HGet(s.ctx, hashKey, fieldInfo.value).Result()
			if err != nil {
				// 如果某个字段查询失败，说明该区块不满足条件
				continue
			}
			var txsByField []string
			if err := json.Unmarshal([]byte(val), &txsByField); err == nil && len(txsByField) > 0 {
				candidateTxLists = append(candidateTxLists, txsByField)
			} else {
				// 如果字段值不匹配，跳过该区块
				continue
			}
		}

		// 3. 对所有候选列表求交集
		if len(candidateTxLists) == 0 {
			// 如果没有查询条件，返回该区块所有交易（这种情况不应该发生）
			continue
		}

		// 从第一个列表开始，依次与其他列表求交集
		intersection := candidateTxLists[0]
		for i := 1; i < len(candidateTxLists); i++ {
			intersection = intersectSlices(intersection, candidateTxLists[i])
			if len(intersection) == 0 {
				break
			}
		}

		resultTxIDs = append(resultTxIDs, intersection...)
	}

	return &SearchResult{TxIDs: resultTxIDs}, nil
}

// GetPosByTxID 根据交易ID获取链上pos信息
func (s *IndexerService) GetPosByTxID(txID string) (string, error) {
	if s.chainClient == nil {
		return "", fmt.Errorf("链客户端未初始化")
	}

	// 通过交易ID获取交易信息
	txInfo, err := s.chainClient.GetTxByTxId(txID)
	if err != nil {
		return "", fmt.Errorf("获取交易信息失败: %v", err)
	}

	if txInfo == nil || txInfo.Transaction == nil {
		return "", fmt.Errorf("交易不存在: %s", txID)
	}

	record, err := parseTxPayload(txInfo.Transaction)
	if err != nil {
		return "", fmt.Errorf("解析交易payload失败: %v", err)
	}
	return record.Metadata.Pos, nil
}

// Helper: 获取 Bitmap 中所有为 1 的 Offset
// 采用将位图获取到本地，本地运算获取所有为1的位
func (s *IndexerService) getSetBits(key string) ([]int64, error) {
	var result []int64

	// 从Redis获取整个位图
	bitmapBytes, err := s.redisClient.Get(s.ctx, key).Bytes()
	if err != nil {
		// 如果key不存在，返回空结果而不是错误
		if err == redis.Nil {
			return result, nil
		}
		return nil, fmt.Errorf("failed to get bitmap: %v", err)
	}

	// 遍历每个字节
	for byteIndex, byteVal := range bitmapBytes {
		// 如果字节为0，跳过（该字节内没有为1的位）
		if byteVal == 0 {
			continue
		}

		// 检查该字节内的每一位
		for bitIndex := 0; bitIndex < 8; bitIndex++ {
			// 检查第bitIndex位是否为1
			if byteVal&(1<<bitIndex) != 0 {
				// 计算该位在整个位图中的位置
				bitPos := int64(byteIndex*8 + 7 - bitIndex)
				result = append(result, bitPos)
			}
		}
	}

	return result, nil
}

// Helper: 切片交集
func intersectSlices(a, b []string) []string {
	m := make(map[string]bool)
	for _, item := range a {
		m[item] = true
	}
	var res []string
	for _, item := range b {
		if m[item] {
			res = append(res, item)
		}
	}
	return res
}

func generateRequestID() string {
	h := fnv.New64()
	h.Write([]byte(strconv.FormatInt(time.Now().UnixNano(), 10)))
	return strconv.FormatUint(h.Sum64(), 10)
}
