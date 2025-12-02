package service

import (
	"errors"

	"chainqa_offchain_demo/chain"
	"strconv"
	"strings"
)

// ExecBlockchain4Chainmaker 执行区块链操作
func ExecBlockchain4Chainmaker(contractName string, method string, params map[string]interface{}) (string, error) {
	// 将params map[string]interface{}转换为map[string]string
	mapParams := make(map[string]string)
	for k, v := range params {
		mapParams[k] = v.(string)
	}
	resp := chain.InvokeContract(contractName, method, mapParams)
	// resp是chainmakerResult
	// 判断resp.ContractResult.Message和resp.ContractResult.Result是否同时存在，若有一方不存在，返回错误
	if resp.ContractResult.Message == "" || resp.ContractResult.Result == "" {
		return "", errors.New("调用区块链错误：未获取到任何有效结果")
	}

	// 判断resp.message是否为success（大小写均可）
	if strings.ToLower(resp.ContractResult.Message) != "success" {
		return "", errors.New(resp.ContractResult.Result)
	} else {

		return string(resp.ContractResult.Result), nil
	}
}

// ------------------------ 以下为业务函数 ------------------------
func GetPublicKeyFromBlockchain4Chainmaker(contractName string) (string, error) {
	// ====================== 构造响应 ======================

	// 创建请求数据
	data := chainDTO{
		ContractName: contractName,
		MethodName:   "getPk",
		Args:         map[string]interface{}{
			// "file_prefix_path": contractName,
		},
	}

	// ======================= 发送请求 ======================

	return ExecBlockchain4Chainmaker(data.ContractName, data.MethodName, data.Args)

}

func UploadEnvelopeToBlockchain4Chainmaker(contractName string, envelope string, cid string, uId string) error {
	// ====================== 构造响应 ======================

	// 创建请求数据
	data := chainDTO{
		ContractName: contractName,
		MethodName:   "updateDataDigtalEnvelop",
		Args: map[string]interface{}{
			// "file_prefix_path": contractName,
			"uId":     uId,
			"pos":     cid,
			"envelop": envelope,
		},
	}

	// ======================= 发送请求 ======================
	_, err := ExecBlockchain4Chainmaker(data.ContractName, data.MethodName, data.Args)
	if err != nil {
		return err
	}
	return nil
}

func UploadEnvelopeToBlockchainWithDomain4Chainmaker(contractName string, envelopJsonStr string) error {
	// ====================== 构造响应 ======================
	// 创建请求数据
	data := chainDTO{
		ContractName: contractName,
		MethodName:   "updateDataDigtalEnvelopWithDomain",
		Args: map[string]interface{}{
			"envelopJsonStr": envelopJsonStr,
		},
	}

	_, err := ExecBlockchain4Chainmaker(data.ContractName, data.MethodName, data.Args)
	if err != nil {
		return err
	}
	return nil
}

func GetAesKeyFromBlockchain4Chainmaker(contractName string, chainServiceUrl string, pos string) (string, error) {
	// ====================== 构造响应 ======================

	// 创建请求数据
	data := chainDTO{
		ContractName: contractName,
		MethodName:   "getAesKey",
		Args: map[string]interface{}{
			"pos": pos,
		},
	}

	// ======================= 发送请求 =====================
	return ExecBlockchain4Chainmaker(data.ContractName, data.MethodName, data.Args)
}

// UpdateQueryLog 更新查询日志
func UpdateQueryLog4Chainmaker(contractName string, chainServiceUrl string, uId string, queryItem string, queryStatus int, queryResult string) error {
	// ====================== 构造响应 ======================

	queryStatusStr := strconv.Itoa(queryStatus)

	// 创建请求数据
	data := chainDTO{
		ContractName: contractName,
		MethodName:   "updateQueryLog",
		Args: map[string]interface{}{
			"uId":         uId,
			"queryItem":   queryItem,
			"queryStatus": queryStatusStr,
			"queryResult": queryResult,
		},
	}

	// ======================= 发送请求 ======================

	_, err := ExecBlockchain4Chainmaker(data.ContractName, data.MethodName, data.Args)
	if err != nil {
		return err
	}
	return nil
}

func GetAllQueryLogByUid4Chainmaker(contractName string, chainServiceUrl string, uId string) (string, error) {
	// ====================== 构造响应 ======================

	// 创建请求数据
	data := chainDTO{
		ContractName: contractName,
		MethodName:   "getAllQueryLogByUid",
		Args: map[string]interface{}{
			"uId": uId,
		},
	}

	// ======================= 发送请求 =====================

	return ExecBlockchain4Chainmaker(data.ContractName, data.MethodName, data.Args)
}

func GetAllQueryLogByTimestamp4Chainmaker(contractName string, chainServiceUrl string, startTime string, endTime string) (string, error) {
	// ====================== 构造响应 ======================

	// 创建请求数据
	data := chainDTO{
		ContractName: contractName,
		MethodName:   "getAllQueryLogByTimestamp",
		Args: map[string]interface{}{
			"startTime": startTime,
			"endTime":   endTime,
		},
	}

	// ======================= 发送请求 ======================
	return ExecBlockchain4Chainmaker(data.ContractName, data.MethodName, data.Args)
}

func CreateDomain4Chainmaker(contractName string, chainServiceUrl string, name string, orgId string) error {
	// ====================== 构造响应 ======================

	// 创建请求数据
	data := chainDTO{
		ContractName: contractName,
		MethodName:   "createDomain",
		Args: map[string]interface{}{
			"name":  name,
			"orgId": orgId,
		},
	}

	// ======================= 发送请求 ======================

	_, err := ExecBlockchain4Chainmaker(data.ContractName, data.MethodName, data.Args)
	if err != nil {
		return err
	}
	return nil
}

func UpdateDomainMetadata4Chainmaker(contractName string, chainServiceUrl string, name string, newMembers string, newPolicy string, orgId string) error {
	// ====================== 构造响应 ======================

	// 创建请求数据
	data := chainDTO{
		ContractName: contractName,
		MethodName:   "updateDomainMetadata",
		Args: map[string]interface{}{
			"name":       name,
			"newMembers": newMembers,
			"newPolicy":  newPolicy,
			"orgId":      orgId,
		},
	}

	// ======================= 发送请求 ======================

	_, err := ExecBlockchain4Chainmaker(data.ContractName, data.MethodName, data.Args)
	if err != nil {
		return err
	}
	return nil
}

func CheckAccess4Chainmaker(contractName string, chainServiceUrl string, name string, action string, orgId string, role string) (bool, error) {
	// ====================== 构造响应 ======================

	// 创建请求数据
	data := chainDTO{
		ContractName: contractName,
		MethodName:   "checkAccess",
		Args: map[string]interface{}{
			"name":   name,
			"action": action,
			"orgId":  orgId,
			"role":   role,
		},
	}

	// ======================= 发送请求 ======================

	result, err := ExecBlockchain4Chainmaker(data.ContractName, data.MethodName, data.Args)
	if err != nil {
		return false, err
	}
	// 解析结果，判断是否为 true
	return result == "true" || result == "True" || result == "1", nil
}

func QueryMyDomains4Chainmaker(contractName string, chainServiceUrl string, orgId string) (string, error) {
	// ====================== 构造响应 ======================

	// 创建请求数据
	data := chainDTO{
		ContractName: contractName,
		MethodName:   "queryMyDomains",
		Args: map[string]interface{}{
			"orgId": orgId,
		},
	}

	// ======================= 发送请求 ======================

	return ExecBlockchain4Chainmaker(data.ContractName, data.MethodName, data.Args)
}

func QueryMyManagedDomains4Chainmaker(contractName string, chainServiceUrl string, orgId string) (string, error) {
	// ====================== 构造响应 ======================

	// 创建请求数据
	data := chainDTO{
		ContractName: contractName,
		MethodName:   "queryMyManagedDomains",
		Args: map[string]interface{}{
			"orgId": orgId,
		},
	}
	return ExecBlockchain4Chainmaker(data.ContractName, data.MethodName, data.Args)
}

func QueryDomainInfo4Chainmaker(contractName string, chainServiceUrl string, name string, orgId string) (string, error) {
	// ====================== 构造响应 ======================

	// 创建请求数据
	data := chainDTO{
		ContractName: contractName,
		MethodName:   "queryDomainInfo",
		Args: map[string]interface{}{
			"name":  name,
			"orgId": orgId,
		},
	}
	return ExecBlockchain4Chainmaker(data.ContractName, data.MethodName, data.Args)
}
