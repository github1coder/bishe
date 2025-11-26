package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"chainmaker.org/chainmaker/contract-sdk-go/v2/pb/protogo"
	"chainmaker.org/chainmaker/contract-sdk-go/v2/sandbox"
	"chainmaker.org/chainmaker/contract-sdk-go/v2/sdk"
)

type ChainQA struct {
}

func (f *ChainQA) InitContract() protogo.Response {
	return sdk.Success([]byte("Init Success"))
}

func (f *ChainQA) UpgradeContract() protogo.Response {
	return sdk.Success([]byte("Upgrade contract success"))
}

func (f *ChainQA) InvokeContract(method string) protogo.Response {
	switch method {
	case "getPk":
		return f.getPK()
	case "updateDataDigtalEnvelop":
		return f.updateDataDigtalEnvelop()
	case "updateDataDigtalEnvelopWithDomain":
		return f.updateDataDigtalEnvelopWithDomain()
	case "getAesKey":
		return f.getAesKey()
	case "updateQueryLog":
		return f.updateQueryLog()
	case "getAllQueryLogByUid":
		return f.getAllQueryLogByUid()
	case "getAllQueryLogByTimestamp":
		return f.getAllQueryLogByTimestamp()
	case "createDomain":
		return f.createDomain()
	case "updateDomainMetadata":
		return f.updateDomainMetadata()
	case "checkAccess":
		return f.checkAccess()
	case "queryMyDomains":
		return f.queryMyDomains()
	case "queryMyManagedDomains":
		return f.queryMyManagedDomains()
	case "queryDomainInfo":
		return f.queryDomainInfo()
	default:
		return sdk.Error("invalid method")
	}
}

// ====================== 合约部分 ======================
// getPK：获取公钥
func (f *ChainQA) getPK() protogo.Response {
	pkString, err := GetPublicKeyString()
	if err != nil {
		return sdk.Error(fmt.Sprintf("[chainqa getPK CONTRACT]获取公钥失败：: %s", err))
	}
	return sdk.Success([]byte(pkString))

}

// updateDataDigtalEnvelop：更新数字信封
// 数字信封
type DataDigtalEnvelop struct {
	Uid         string `json:"uId"`         //用户ID
	TimeStamp   string `json:"timeStamp"`   //时间戳,日志中仍然按照原始类型的时间戳存储,例如：1735693850
	Pos         string `json:"pos"`         //文件位置
	Envelop     string `json:"envelop"`     //数字信封(JSON)
	Name        string `json:"name"`        //用户姓名
	Age         int    `json:"age"`         //年龄
	Gender      string `json:"gender"`      //性别
	Hospital    string `json:"hospital"`    //医院
	Department  string `json:"department"`  //科室
	DiseaseCode string `json:"diseaseCode"` //疾病代码
	DomainID    string `json:"domainID"`    //数据域ID
}

/**
 * 更新数字信封
 * @param uId 用户ID
 * @param pos 文件位置
 * @param envelop 数字信封（json字符串） 格式为:[{"AesKeyCipBase64":"密文"}] 【注意是数组哦，后期可能会扩展】
 * 数字信封是一个json数组，包含了一个“AES密钥”的密文。“AES密钥”用于解密IPFS文件，使用RSA加密“AES密钥”，将其存储在数字信封中
 */
func (f *ChainQA) updateDataDigtalEnvelop() protogo.Response {
	params := sdk.Instance.GetArgs()
	uIdStr := string(params["uId"])
	timestampNumberStr, err := sdk.Instance.GetTxTimeStamp()
	if err != nil {
		return sdk.Error("[chainqa updateDataDigtalEnvelop CONTRACT]时间戳获取失败")
	}
	posStr := string(params["pos"])
	envelopStr := string(params["envelop"])
	if uIdStr == "" || timestampNumberStr == "" || posStr == "" || envelopStr == "" {
		return sdk.Error("[chainqa updateDataDigtalEnvelop CONTRACT]参数不能为空")
	}
	dataDigtalEnvelop := DataDigtalEnvelop{
		Uid:       uIdStr,
		TimeStamp: timestampNumberStr,
		Pos:       posStr,
		Envelop:   envelopStr,
	}
	dataDigtalEnvelopBytes, err := json.Marshal(dataDigtalEnvelop)
	if err != nil {
		return sdk.Error(fmt.Sprintf("[chainqa updateDataDigtalEnvelop CONTRACT]序列化失败：: %s", err))
	}
	err = sdk.Instance.PutStateByte("chain_data_digtal_envelop", posStr, dataDigtalEnvelopBytes)
	if err != nil {
		return sdk.Error(fmt.Sprintf("[chainqa updateDataDigtalEnvelop CONTRACT]写入区块链失败：: %s", err))
	}
	return sdk.Success([]byte("上传数字信封成功"))
}

/**
 * 更新数字信封（带数据域）
 * @param uId 用户ID
 * @param pos 文件位置
 * @param envelop 数字信封（json字符串） 格式为:[{"AesKeyCipBase64":"密文"}] 【注意是数组哦，后期可能会扩展】
 * @param name 用户姓名
 * @param age 年龄
 * @param gender 性别
 * @param hospital 医院
 * @param department 科室
 * @param diseaseCode 疾病代码
 * @param domainID 数据域ID
 * 数字信封是一个json数组，包含了一个“AES密钥”的密文。“AES密钥”用于解密IPFS文件，使用RSA加密“AES密钥”，将其存储在数字信封中
 */
func (f *ChainQA) updateDataDigtalEnvelopWithDomain() protogo.Response {
	params := sdk.Instance.GetArgs()

	timestampNumberStr, err := sdk.Instance.GetTxTimeStamp()
	if err != nil {
		return sdk.Error("[chainqa updateDataDigtalEnvelop CONTRACT]时间戳获取失败")
	}

	envelopJsonStr := string(params["envelopJsonStr"])
	var dataDigtalEnvelop DataDigtalEnvelop
	err = json.Unmarshal([]byte(envelopJsonStr), &dataDigtalEnvelop)
	if err != nil {
		return sdk.Error(fmt.Sprintf("[chainqa updateDataDigtalEnvelop CONTRACT]反序列化失败：: %s", err))
	}
	dataDigtalEnvelop.TimeStamp = timestampNumberStr
	dataDigtalEnvelopBytes, err := json.Marshal(dataDigtalEnvelop)
	if err != nil {
		return sdk.Error(fmt.Sprintf("[chainqa updateDataDigtalEnvelop CONTRACT]序列化失败：: %s", err))
	}
	posStr := dataDigtalEnvelop.Pos
	err = sdk.Instance.PutStateByte("chain_data_digtal_envelop", posStr, dataDigtalEnvelopBytes)
	if err != nil {
		return sdk.Error(fmt.Sprintf("[chainqa updateDataDigtalEnvelop CONTRACT]写入区块链失败：: %s", err))
	}
	return sdk.Success([]byte("上传数字信封成功"))
}

/**
 * 解密数字信封获取AES密钥
 * @param pos 文件前缀路径(一般是合约名)
 */
func (f *ChainQA) getAesKey() protogo.Response {
	params := sdk.Instance.GetArgs()
	pos := string(params["pos"])
	if pos == "" {
		return sdk.Error("[chainqa getPK CONTRACT]pos参数不能为空")
	}
	dataDigtalEnvelopBytes, err := sdk.Instance.GetStateByte("chain_data_digtal_envelop", pos)
	if err != nil {
		return sdk.Error(fmt.Sprintf("[chainqa chainQuery CONTRACT]获取数字信封失败：: %s", err))
	}
	if dataDigtalEnvelopBytes == nil || len(dataDigtalEnvelopBytes) == 0 {
		return sdk.Error(fmt.Sprintf("[chainqa chainQuery CONTRACT]数字信封不存在"))
	}
	dataDigtalEnvelop := DataDigtalEnvelop{}
	err = json.Unmarshal(dataDigtalEnvelopBytes, &dataDigtalEnvelop)
	if err != nil {
		return sdk.Error(fmt.Sprintf("[chainqa chainQuery CONTRACT]反序列化失败：: %s", err))
	}
	aesKey, err := MatchEnvelopAndDecrpty(dataDigtalEnvelop.Envelop)
	if err != nil {
		return sdk.Error(fmt.Sprintf("[chainqa chainQuery CONTRACT]解密数字信封失败：: %s", err))
	}
	return sdk.Success([]byte(aesKey))
}

// [日志部分]：其实不如让区块链管理系统来做这个事情，这里只是一个简单的示例

// getQueryLog：获取查询日志
type QueryLog struct {
	QueryId     string //查询ID
	Uid         string //用户ID
	Timestamp   string //时间戳，日志中仍然按照原始类型的时间戳存储，例如：1735693850
	QueryItem   string //查询项
	QueryStatus int    //查询状态
	QueryResult string //查询结果
}

/**
 * 上传查询日志
 * @param uId 用户ID
 * @param queryItem 查询项
 * @param queryStatus 查询状态
 * @param queryResult 查询结果
 */
func (f *ChainQA) updateQueryLog() protogo.Response {
	params := sdk.Instance.GetArgs()
	uId := string(params["uId"])
	queryItem := string(params["queryItem"])
	queryStatus := string(params["queryStatus"])
	queryStatusInt, _ := strconv.Atoi(queryStatus)
	queryResult := string(params["queryResult"])
	timestampNumberStr, err := sdk.Instance.GetTxTimeStamp()
	if err != nil {
		return sdk.Error("[chainqa updateQueryLog CONTRACT]时间戳获取失败")
	}
	queryId := StringToMD5(uId + timestampNumberStr)
	timestampStrUnderline, err := StringTimestampToUnderlineTimeStamp(timestampNumberStr)
	if err != nil {
		return sdk.Error("[chainqa updateQueryLog CONTRACT]时间戳转换失败")
	}
	QueryLog := QueryLog{
		QueryId:     queryId,
		Uid:         uId,
		Timestamp:   timestampNumberStr,
		QueryItem:   queryItem,
		QueryStatus: queryStatusInt,
		QueryResult: queryResult,
	}
	QueryLogBytes, err := json.Marshal(QueryLog)
	if err != nil {
		return sdk.Error(fmt.Sprintf("[chainqa updateQueryLog CONTRACT]序列化失败：: %s", err))
	}
	QueryIdBytes := []byte(queryId)
	err = sdk.Instance.PutStateByte("chain_query_log", queryId, QueryLogBytes)
	if err != nil {
		return sdk.Error(fmt.Sprintf("[chainqa updateQueryLog CONTRACT]写入区块链失败：: %s", err))
	}
	err = sdk.Instance.PutStateByte("logHash_timestamp", timestampStrUnderline, QueryIdBytes)
	if err != nil {
		return sdk.Error(fmt.Sprintf("[chainqa updateQueryLog CONTRACT]按时间写入区块链失败：: %s", err))
	}
	err = sdk.Instance.PutStateByte("logHash_uId", uId+"__"+timestampStrUnderline, QueryIdBytes)
	if err != nil {
		return sdk.Error(fmt.Sprintf("[chainqa updateQueryLog CONTRACT]按用户写入区块链失败：: %s", err))
	}
	return sdk.Success([]byte("上传查询日志成功"))
}

/**
 * 根据用户ID获取查询日志
 * @param uId 用户ID
 */
// * 建议uId不要以"__"结尾，否则会引起查询到别人的查询日志
func (f *ChainQA) getAllQueryLogByUid() protogo.Response {
	params := sdk.Instance.GetArgs()
	uId := string(params["uId"])
	if uId == "" {
		return sdk.Error("[chainqa getAllQueryLogByUid CONTRACT]uId参数不能为空")
	}
	queryIdArray := []string{}
	rsKv, err := sdk.Instance.NewIteratorPrefixWithKeyField("logHash_uId", uId+"__")
	if err != nil {
		return sdk.Error(fmt.Sprintf("[chainqa getAllQueryLogByUid CONTRACT]获取查询日志失败：: %s", err))
	}
	for rsKv.HasNext() {
		kvkey, kvfeild, kvvalue, err := rsKv.Next()
		if err != nil || kvkey == "" || kvfeild == "" || kvvalue == nil {
			continue
		}
		queryIdArray = append(queryIdArray, string(kvvalue))
	}
	queryLogArray := []QueryLog{}
	counts := 0
	for _, queryId := range queryIdArray {
		queryLogBytes, err := sdk.Instance.GetStateByte("chain_query_log", queryId)
		if err != nil {
			continue
		}
		var queryLog QueryLog
		err = json.Unmarshal(queryLogBytes, &queryLog)
		if err != nil {
			continue
		}
		queryLogArray = append(queryLogArray, queryLog)
		counts++
	}
	type QueryLogArray struct {
		Count         int
		QueryLogArray []QueryLog
	}
	QueryLogSearchResult := QueryLogArray{
		Count:         counts,
		QueryLogArray: queryLogArray,
	}
	QueryLogSearchResultBytes, err := json.Marshal(QueryLogSearchResult)
	if err != nil {
		return sdk.Error(fmt.Sprintf("[chainqa getAllQueryLogByUid CONTRACT]序列化失败：: %s", err))
	}
	return sdk.Success(QueryLogSearchResultBytes)
}

/**
 * 根据时间戳获取查询日志
 * @param startTime 开始时间戳
 * @param endTime 结束时间戳（若为Now，则获取当前时间）
 */
//* 本方法使用了范围查询。但范围查询时，必须精确指定field，所以如果按时间查询，只能以时间为field。时间戳仅精确到秒，所以如果一秒内有多次并发，会覆盖！
func (f *ChainQA) getAllQueryLogByTimestamp() protogo.Response {
	params := sdk.Instance.GetArgs()
	startTimeStamp := string(params["startTime"])
	endTimeStamp := string(params["endTime"])
	if startTimeStamp == "" {
		return sdk.Error(fmt.Sprintf("[chainqa getAllQueryLogByTimestamp CONTRACT]开始时间不能为空"))
	}
	if endTimeStamp == "" || endTimeStamp == "Now" {
		var err error
		endTimeStamp, err = sdk.Instance.GetTxTimeStamp()
		if err != nil {
			return sdk.Error(fmt.Sprintf("[chainqa getAllQueryLogByTimestamp CONTRACT]时间戳获取失败：: %s", err))
		}
	}
	startTime, err := StringTimestampToUnderlineTimeStamp(startTimeStamp)
	if err != nil {
		return sdk.Error(fmt.Sprintf("[chainqa getAllQueryLogByTimestamp CONTRACT]时间戳转换失败：: %s", err))
	}
	endTime, err := StringTimestampToUnderlineTimeStamp(endTimeStamp)
	if err != nil {
		return sdk.Error(fmt.Sprintf("[chainqa getAllQueryLogByTimestamp CONTRACT]时间戳转换失败：: %s", err))
	}
	queryIdArray := []string{}
	rsKv, err := sdk.Instance.NewIteratorWithField("logHash_timestamp", startTime, endTime)
	if err != nil {
		return sdk.Error(fmt.Sprintf("[chainqa getAllQueryLogByTimestamp CONTRACT]获取查询日志失败：: %s", err))
	}
	for rsKv.HasNext() {
		kvkey, kvfeild, kvvalue, err := rsKv.Next()
		if err != nil || kvkey == "" || kvfeild == "" || kvvalue == nil {
			continue
		}
		queryIdArray = append(queryIdArray, string(kvvalue))
	}
	queryLogArray := []QueryLog{}
	counts := 0
	for _, queryId := range queryIdArray {
		queryLogBytes, err := sdk.Instance.GetStateByte("chain_query_log", queryId)
		if err != nil {
			continue
		}
		var queryLog QueryLog
		err = json.Unmarshal(queryLogBytes, &queryLog)
		if err != nil {
			continue
		}
		queryLogArray = append(queryLogArray, queryLog)
		counts++
	}
	type QueryLogArray struct {
		Count         int
		QueryLogArray []QueryLog
	}
	QueryLogSearchResult := QueryLogArray{
		Count:         counts,
		QueryLogArray: queryLogArray,
	}
	QueryLogSearchResultBytes, err := json.Marshal(QueryLogSearchResult)
	if err != nil {
		return sdk.Error(fmt.Sprintf("[chainqa getAllQueryLogByTimestamp CONTRACT]序列化失败：: %s", err))
	}
	return sdk.Success(QueryLogSearchResultBytes)
}

//////////////////////数据域//////////////////////

// ===================================================================================
// Part 1: 数据结构定义 (保持业务逻辑一致性)
// ===================================================================================

// TrustedDataDomain 定义可信数据域实体
type TrustedDataDomain struct {
	DomainID     string       `json:"domainID"`
	OwnerMSP     string       `json:"ownerMSP"` // 在长安链中通常对应 OrgId
	Name         string       `json:"name"`
	Members      []string     `json:"members"`
	AccessPolicy AccessPolicy `json:"accessPolicy"`
	Status       string       `json:"status"`
}

// AccessPolicy 访问策略容器
type AccessPolicy struct {
	Version      string                  `json:"version"`
	PolicyGrants map[string][]RolePolicy `json:"policyGrants"` // Key: OrgId
}

// RolePolicy 角色策略
type RolePolicy struct {
	Role        string       `json:"role"`
	Permissions []Permission `json:"permissions"`
}

// Permission 权限详情
type Permission struct {
	Effect  string   `json:"effect"`
	Actions []string `json:"actions"`
}

// UserAttributes 用于 CheckAccess 的入参
type UserAttributes struct {
	OrgId string `json:"orgId"`
	Role  string `json:"role"`
}

// ===================================================================================
// Part 2: 数据域新增与更新接口
// ===================================================================================

// createDomain 创建数据域
// 参数: domainID, name, accessPolicy (json)
func (f *ChainQA) createDomain() protogo.Response {
	// 1. 获取并校验参数
	args := sdk.Instance.GetArgs()
	domainID := "DOMAIN_" + string(args["name"])
	name := string(args["name"])

	if domainID == "DOMAIN_" || name == "" {
		return sdk.Error("Missing required parameters")
	}

	// 2. 获取调用者身份（组织ID）
	senderOrgId := string(args["orgId"])
	if senderOrgId == "" {
		return sdk.Error("orgId is required")
	}

	// 3. 检查是否存在
	// 长安链 PutStateByte 需要 field 参数，这里对于简单KV，field 设为空字符串或 "data"
	existingBytes, err := sdk.Instance.GetStateByte(domainID, "")
	if err != nil {
		return sdk.Error("Failed to read state")
	}
	if len(existingBytes) > 0 {
		return sdk.Error("Domain already exists")
	}

	domain := TrustedDataDomain{
		DomainID: domainID,
		OwnerMSP: senderOrgId,
		Name:     name,
		Members:  []string{senderOrgId},
		Status:   "Active",
	}

	// 5. 序列化并存储
	domainBytes, err := json.Marshal(domain)
	if err != nil {
		return sdk.Error("Marshal error")
	}

	// 存储: Key=domainID, Field="", Value=json
	if err := sdk.Instance.PutStateByte(domainID, "", domainBytes); err != nil {
		return sdk.Error("PutState error")
	}

	// // 6. 发送事件
	// sdk.Instance.EmitEvent("CreateDomainEvent", []string{domainID, senderOrgId})
	// sdk.Instance.Log(fmt.Sprintf("Domain %s created by %s", domainID, senderOrgId))

	return sdk.Success([]byte(domainID))
}

// updateDomainMetadata 更新数据域
// 参数: domainID, newMembers (json array), newPolicy (json object)
func (f *ChainQA) updateDomainMetadata() protogo.Response {
	args := sdk.Instance.GetArgs()
	domainID := "DOMAIN_" + string(args["name"])
	newMembersJson := string(args["newMembers"])
	newPolicyJson := string(args["newPolicy"])

	if domainID == "DOMAIN_" {
		return sdk.Error("Missing domainID")
	}

	// 1. 获取现有数据
	domainBytes, err := sdk.Instance.GetStateByte(domainID, "")
	if err != nil || len(domainBytes) == 0 {
		return sdk.Error("Domain not found")
	}

	var domain TrustedDataDomain
	if err := json.Unmarshal(domainBytes, &domain); err != nil {
		return sdk.Error("Unmarshal domain error")
	}

	// 2. 权限校验：仅所有者可修改
	senderOrgId := string(args["orgId"])
	if senderOrgId == "" {
		return sdk.Error("orgId is required")
	}
	// 注意：实际工程可能还需要校验具体的SenderAddr而不仅仅是OrgId，视业务需求而定
	if domain.OwnerMSP != senderOrgId {
		return sdk.Error("Unauthorized: only owner can update")
	}

	// 3. 执行更新
	if newMembersJson != "" {
		var members []string
		if err := json.Unmarshal([]byte(newMembersJson), &members); err == nil {
			domain.Members = members
		}
	}

	if newPolicyJson != "" {
		var policy AccessPolicy
		if err := json.Unmarshal([]byte(newPolicyJson), &policy); err == nil {
			domain.AccessPolicy = policy
		}
	}

	// 4. 保存
	updatedBytes, _ := json.Marshal(domain)
	if err := sdk.Instance.PutStateByte(domainID, "", updatedBytes); err != nil {
		return sdk.Error("PutState error")
	}

	return sdk.Success([]byte("Update success"))
}

// ===================================================================================
// Part 3: 数据域验证功能 (CheckAccess)
// ===================================================================================

// checkAccess 细粒度权限校验
// 参数: domainID, userAttrs (json: {orgId, role}), action
func (f *ChainQA) checkAccess() protogo.Response {
	args := sdk.Instance.GetArgs()
	domainID := "DOMAIN_" + string(args["name"])
	orgId := string(args["orgId"])
	if orgId == "" {
		return sdk.Error("orgId is required")
	}
	role := string(args["role"])
	if role == "" {
		return sdk.Error("role is required")
	}
	action := string(args["action"])

	// 1. 身份解析
	var user UserAttributes
	user.OrgId = orgId
	user.Role = role

	// 2. 策略获取
	domainBytes, err := sdk.Instance.GetStateByte(domainID, "")
	if err != nil || len(domainBytes) == 0 {
		return sdk.Error("Domain not found")
	}

	var domain TrustedDataDomain
	if err := json.Unmarshal(domainBytes, &domain); err != nil {
		return sdk.Error("Data corruption")
	}

	if domain.Status != "Active" {
		return sdk.Error("Domain is not active")
	}

	// 3. 核心验证逻辑 (三级验证)
	allowed := f.verifyPolicy(domain.AccessPolicy, user, action)

	if allowed {
		return sdk.Success([]byte("true"))
	}
	return sdk.Error("Access Denied") // 或者返回 sdk.Success([]byte("false")) 取决于网关如何处理
}

// verifyPolicy 纯逻辑校验函数
func (f *ChainQA) verifyPolicy(policy AccessPolicy, user UserAttributes, action string) bool {
	// Level 1: 机构匹配 (OrgId 匹配)
	orgPolicies, exists := policy.PolicyGrants[user.OrgId]
	if !exists {
		return false
	}

	// Level 2: 角色匹配
	var targetRolePolicy *RolePolicy
	for _, rp := range orgPolicies {
		if rp.Role == user.Role {
			targetRolePolicy = &rp
			break
		}
	}
	if targetRolePolicy == nil {
		return false
	}

	// Level 3: 权限动作匹配
	for _, perm := range targetRolePolicy.Permissions {
		if perm.Effect == "Allow" {
			for _, act := range perm.Actions {
				if act == action {
					return true
				}
			}
		}
	}

	return false
}

// ===================================================================================
// Part 4: 查询功能扩展
// ===================================================================================

// queryMyDomains 获取当前调用者所在组织（OrgId）参与的所有数据域name列表
// 逻辑：遍历账本所有数据 -> 解析JSON -> 检查Members数组是否包含当前OrgId
func (f *ChainQA) queryMyDomains() protogo.Response {
	args := sdk.Instance.GetArgs()
	// 1. 获取调用者的组织ID
	senderOrgId := string(args["orgId"])
	if senderOrgId == "" {
		return sdk.Error("orgId is required")
	}

	// 2. 创建迭代器 (NewIterator)
	iter, err := sdk.Instance.NewIteratorPrefixWithKey("DOMAIN_")
	if err != nil {
		return sdk.Error("Failed to create iterator: " + err.Error())
	}
	defer iter.Close() // 务必关闭迭代器以释放资源

	var myDomainNames []string

	// 3. 循环遍历
	for iter.HasNext() {
		kvkey, _, kvvalue, err := iter.Next()
		if err != nil || kvkey == "" || kvvalue == nil {
			continue
		}

		// 容错处理：跳过空值
		if len(kvvalue) == 0 {
			continue
		}

		// 4. 解析数据域对象
		var domain TrustedDataDomain
		// 如果合约中存储了非TrustedDataDomain结构的数据，这里会解析失败，需跳过
		if err := json.Unmarshal(kvvalue, &domain); err != nil {
			// 记录日志方便调试，但不中断流程
			sdk.Instance.Log(fmt.Sprintf("Skip invalid data for key %s", kvkey))
			continue
		}
		// 5. 检查权限：判断当前OrgId是否在Members列表中
		isMember := false
		for _, member := range domain.Members {
			if member == senderOrgId {
				isMember = true
				break
			}
		}

		// 如果是成员，或者该组织是Owner，则加入结果集
		if isMember || domain.OwnerMSP == senderOrgId {
			myDomainNames = append(myDomainNames, domain.Name)
		}
	}

	// 6. 返回结果 (JSON数组)
	respJson, err := json.Marshal(myDomainNames)
	if err != nil {
		return sdk.Error("Marshal response error")
	}

	return sdk.Success(respJson)
}

// 查询我管理的domain
func (f *ChainQA) queryMyManagedDomains() protogo.Response {
	args := sdk.Instance.GetArgs()
	// 1. 获取调用者的组织ID
	senderOrgId := string(args["orgId"])
	if senderOrgId == "" {
		return sdk.Error("orgId is required")
	}

	// 2. 创建迭代器 (NewIterator)
	iter, err := sdk.Instance.NewIteratorPrefixWithKey("DOMAIN_")
	if err != nil {
		return sdk.Error("Failed to create iterator: " + err.Error())
	}
	defer iter.Close() // 务必关闭迭代器以释放资源

	var myManagedDomainNames []string

	// 3. 循环遍历
	for iter.HasNext() {
		kvkey, _, kvvalue, err := iter.Next()
		if err != nil || kvkey == "" || kvvalue == nil {
			continue
		}

		// 容错处理：跳过空值
		if len(kvvalue) == 0 {
			continue
		}

		// 4. 解析数据域对象
		var domain TrustedDataDomain
		// 如果合约中存储了非TrustedDataDomain结构的数据，这里会解析失败，需跳过
		if err := json.Unmarshal(kvvalue, &domain); err != nil {
			// 记录日志方便调试，但不中断流程
			sdk.Instance.Log(fmt.Sprintf("Skip invalid data for key %s", kvkey))
			continue
		}

		// 5. 检查权限：判断当前OrgId是否是Owner
		if domain.OwnerMSP == senderOrgId {
			myManagedDomainNames = append(myManagedDomainNames, domain.Name)
		}
	}

	// 6. 返回结果 (JSON数组)
	respJson, err := json.Marshal(myManagedDomainNames)
	if err != nil {
		return sdk.Error("Marshal response error")
	}

	return sdk.Success(respJson)

}

// 查询domainName的详细信息，并验证权限
func (f *ChainQA) queryDomainInfo() protogo.Response {
	args := sdk.Instance.GetArgs()
	domainID := "DOMAIN_" + string(args["name"])
	if domainID == "DOMAIN_" {
		return sdk.Error("domainName is required")
	}
	senderOrgId := string(args["orgId"])
	if senderOrgId == "" {
		return sdk.Error("orgId is required")
	}
	domainBytes, err := sdk.Instance.GetStateByte(domainID, "")
	if err != nil {
		return sdk.Error("Failed to get domain bytes")
	}
	if len(domainBytes) == 0 {
		return sdk.Error("Domain not found")
	}

	var domain TrustedDataDomain
	if err := json.Unmarshal(domainBytes, &domain); err != nil {
		return sdk.Error("Data corruption")
	}
	if domain.OwnerMSP != senderOrgId {
		return sdk.Error("Unauthorized: only owner can update")
	}

	return sdk.Success(domainBytes)
}

func main() {
	err := sandbox.Start(new(ChainQA))
	if err != nil {
		log.Fatal(err)
	}
}
