//+build linux darwin windows
package chain

import (
    "fmt"
    
	"chainmaker.org/chainmaker/pb-go/v2/common"
    // "chainmaker.org/chainmaker/common/v2/log"
    sdk "chainmaker.org/chainmaker/sdk-go/v2"
    "go.uber.org/zap"
    "chainmaker.org/chainmaker/common/v2/log"
)



const configPathDefault = "./chain/chainConfig/sdk_config.yml"

func getLogger() *zap.SugaredLogger {
	config := log.LogConfig{
		Module:       "[SDK]",
		LogPath:      "./log/sdk.log",
		LogLevel:     log.LEVEL_DEBUG,
		MaxAge:       30,
		JsonFormat:   false,
		ShowLine:     true,
		LogInConsole: false,
	}

	logger, _ := log.InitSugarLogger(&config)
	return logger
}


// 用于解析QueryClaimContract返回的TxResponse的结构体
type QueryResult struct {
    ContractResult struct {
        Result string `json:"result"`
        Message string `json:"message"`
    } `json:"contract_result"`
    TxId    string `json:"tx_id"`
}
// 将TxResponse转为QueryResult并打印为JSON
func getQueryResultAsJSON(resp *common.TxResponse) (QueryResult) {
    if resp == nil {
        fmt.Println("resp is nil")
        return QueryResult{}
    }
    result := QueryResult{
        TxId:    resp.TxId,
    }
    if resp.ContractResult != nil {
        result.ContractResult.Result = string(resp.ContractResult.Result)
        result.ContractResult.Message = resp.ContractResult.Message
    }
    return result
}
func getFinalConfigPath(configPath string) string {
    if configPath == "" {
        return configPathDefault
    }
    return configPath
}

// 调用
func InvokeContract(contractName, method string, params map[string]string) (QueryResult){
    resp, err := InvokeClaimContract("", contractName, method, params)
    if err != nil {
        fmt.Printf("InvokeClaimContract failed: %v\n", err)
        return QueryResult{
            ContractResult: struct {
                Result  string `json:"result"`
                Message string `json:"message"`
            }{
                Message: "Fail",
                Result:  "调用智能合约出现失败: " + err.Error(),
            },
            TxId: "",
        }
    }
    return getQueryResultAsJSON(resp)
}

// 调用合约（invoke）
func InvokeClaimContract(configPath_, contractName , method string, params map[string]string) (*common.TxResponse, error) {
    configPath := getFinalConfigPath(configPath_)
    client, err := sdk.NewChainClient(sdk.WithConfPath(configPath), sdk.WithChainClientLogger(getLogger()))
    if err != nil {
        return nil, err
    }
    var kvs []*common.KeyValuePair
    for k, v := range params {
        kvs = append(kvs, &common.KeyValuePair{Key: k, Value: []byte(v)})
    }

    // //   - contractName: 合约名称
	//   - method: 合约方法
	//   - txId: 交易ID
	//           格式要求：长度为64字节，字符在a-z0-9
	//           可为空，若为空字符串，将自动生成txId
	//   - kvs: 合约参数
	//   - timeout: 超时时间，单位：s，若传入-1，将使用默认超时时间：10s
	//   - withSyncResult: 是否同步获取交易执行结果
	//            当为true时，若成功调用，common.TxResponse.ContractResult.Result为common.TransactionInfo
	//            当为false时，若成功调用，common.TxResponse.ContractResult为空，可以通过common.TxResponse.TxId查询交易结果

    resp, err := client.InvokeContract(contractName, method, "", kvs, -1, true)
    if err != nil {
        return nil, err
    }
    if resp.Code != common.TxStatusCode_SUCCESS {
        return resp, fmt.Errorf("invoke contract failed, [code:%d]/[msg:%v]", resp.Code, resp)
    }
    
    return resp, nil
}

// 查询合约（query）
func QueryClaimContract(configPath_, contractName, method string, params map[string]string) (*common.TxResponse, error) {
    configPath := getFinalConfigPath(configPath_)
    client, err := sdk.NewChainClient(sdk.WithConfPath(configPath), sdk.WithChainClientLogger(getLogger()))
    if err != nil {
        return nil, err
    }
    
    var kvs []*common.KeyValuePair
    for k, v := range params {
        kvs = append(kvs, &common.KeyValuePair{Key: k, Value: []byte(v)})
    }
    resp, err := client.QueryContract(contractName, method, kvs, -1)
    if err != nil {
        return nil, err
    }
    
    return resp, nil
}

// 2. 其他函数（如 updateTrustRoot、enableSyncCanonicalTxResult、oneTransactionPerSecond）可按类似方式封装

// 工具函数
func panicErr(err error) {
    if err != nil {
        fmt.Println(err)
    }
}


func TestChaincode() {
    fmt.Println("Testing chaincode invocation and query...")
    // 构建一个map
    params := make(map[string]string)
    params["file_hash"] = "example_file_hash_1701307200"
    rsp, _ := InvokeClaimContract("", "chainQA", "getAesKey", params)
    // if err != nil {
    //     fmt.Printf("QueryClaimContract failed: %v\n", err)
    //     return
    // }
    // PrintQueryResultAsJSON(rsp)
    // 打印rsp
    fmt.Println("rsp:", rsp)
    /**
    Testing chaincode invocation and query...
{
  "contract_result": {
    "result": "[chainqa getPK CONTRACT]pos参数不能为空",
    "message": "Fail"
  },
  "tx_id": "1863a7d2db970d9dcae3a40e2e72ac48f93456375f2049a1bcaec64e1a715e39"  
}
  */
}