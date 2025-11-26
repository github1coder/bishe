package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"strconv"
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

		// --- 处理 DiseaseCode\Name\Gender\Hospital\Department\Uid (等值型/字符型) ---
		diseaseBucket := hashBucket(record.Metadata.DiseaseCode)
		nameBucket := hashBucket(record.Metadata.Name)
		genderBucket := hashBucket(record.Metadata.Gender)
		hospitalBucket := hashBucket(record.Metadata.Hospital)
		departmentBucket := hashBucket(record.Metadata.Department)
		uidBucket := hashBucket(record.Metadata.Uid)
		activeBuckets[fmt.Sprintf("disease:%d", diseaseBucket)] = true
		activeBuckets[fmt.Sprintf("name:%d", nameBucket)] = true
		activeBuckets[fmt.Sprintf("gender:%d", genderBucket)] = true
		activeBuckets[fmt.Sprintf("hospital:%d", hospitalBucket)] = true
		activeBuckets[fmt.Sprintf("department:%d", departmentBucket)] = true
		activeBuckets[fmt.Sprintf("uid:%d", uidBucket)] = true

		// Layer 3: 区块内哈希索引 (Redis Hash)
		// 为了支持 "Find Tx by Disease\Name\Gender\Hospital\Department\Uid", 结构应为 Hash: Key=Code, Value=List[TxID]
		attributeValues := map[string]string{
			"disease":    record.Metadata.DiseaseCode,
			"name":       record.Metadata.Name,
			"gender":     record.Metadata.Gender,
			"hospital":   record.Metadata.Hospital,
			"department": record.Metadata.Department,
			"uid":        record.Metadata.Uid,
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

// 辅助函数：解析交易
func parseTxPayload(tx *common.Transaction) (*models.OnChainRecord, error) {
	// 实际项目中需根据合约调用的参数结构解析
	// 根据合约，数据存储在链上状态中，这里需要从交易参数中解析
	// 或者从链上状态读取（需要知道 pos）

	// 方法1: 从交易参数解析（如果合约直接传递了元数据）
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

	// 方法2: 尝试从其他参数构建（根据实际合约参数结构调整）
	// 这里假设参数包含元数据字段
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
	if record.Metadata.DomainID != "" || record.Metadata.DiseaseCode != "" {
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
