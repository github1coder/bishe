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
	case "getAesKey":
		return f.getAesKey()
	case "updateQueryLog":
		return f.updateQueryLog()
	case "getAllQueryLogByUid":
		return f.getAllQueryLogByUid()
	case "getAllQueryLogByTimestamp":
		return f.getAllQueryLogByTimestamp()
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
	Uid       string //用户ID
	TimeStamp string //时间戳,日志中仍然按照原始类型的时间戳存储,例如：1735693850
	Pos       string //文件位置
	Envelop   string //数字信封(JSON)
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

func main() {
	err := sandbox.Start(new(ChainQA))
	if err != nil {
		log.Fatal(err)
	}
}
