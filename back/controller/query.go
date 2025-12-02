package controller

import (
	"chainqa_offchain_demo/indexer"
	"chainqa_offchain_demo/models"
	"chainqa_offchain_demo/service"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func QueryDataHandler(c *gin.Context) {
	type QueryDataDTO struct {
		Uid       string    `json:"uId"`       // 用户ID
		QueryItem string    `json:"queryItem"` // 查询项
		ApiUrl    ApiUrlDTO `json:"apiUrl"`    // API地址
		// FilePoses []string  `json:"filePoses"` // 文件位置（废弃，直接从QueryItem解析）
	}

	var queryDataDTO QueryDataDTO
	// 绑定JSON数据到结构体
	if err := c.ShouldBindJSON(&queryDataDTO); err != nil {
		models.ResponseError400(c, http.StatusBadRequest, "请求格式错误", err)
		return
	}
	queryDataDTO.Uid = strings.TrimSpace(queryDataDTO.Uid)             // 去除空格
	queryDataDTO.QueryItem = strings.TrimSpace(queryDataDTO.QueryItem) // 去除空格

	// 提取filePos
	// 解析 QueryItem 字段中的 JSON
	type QueryItemDTO struct {
		FilePos [][]string `json:"filePos"` //  filePos 是一个二维数组
	}
	var queryItemDTO QueryItemDTO
	if err := json.Unmarshal([]byte(queryDataDTO.QueryItem), &queryItemDTO); err != nil {
		models.ResponseError400(c, http.StatusBadRequest, "QueryItem 格式错误", err)
		return
	}
	// 提取 filePos 字段
	filePos2Dim := queryItemDTO.FilePos
	if len(filePos2Dim) == 0 {
		models.ResponseError400(c, http.StatusBadRequest, "filePos 不能为空", nil)
		return
	}
	// 将filePos变为一维数组
	FilePoses := make([]string, 0) // 初始化切片
	for _, pos := range filePos2Dim {
		for _, p := range pos {
			FilePoses = append(FilePoses, p)
		}
	}

	// 查询每个文件位置的数据，用map[string]string存储
	fileDataMap := make(map[string]string)
	for _, filePos := range FilePoses {
		fileCiperData, err := service.HandleGetIPFSFile(filePos, queryDataDTO.ApiUrl.IpfsServiceUrl)
		if err != nil {
			models.ResponseError400(c, http.StatusBadRequest, "请求IPFS错误:"+err.Error(), err)
			return
		}
		// 等待2s
		time.Sleep(2000 * time.Millisecond) // 延时2s
		// 从区块链中获取AES密钥
		aesKey, err := service.GetAesKeyFromBlockchain(queryDataDTO.ApiUrl.ContractName, queryDataDTO.ApiUrl.ChainServiceUrl, filePos)
		if err != nil {
			models.ResponseError400(c, http.StatusBadRequest, "获取AES密钥失败:"+err.Error(), err)
			return
		}
		// 解密
		fileData, err := service.AesDecrypt(fileCiperData, aesKey) // 解密文件
		if err != nil {
			models.ResponseError400(c, http.StatusBadRequest, "解密失败:"+err.Error(), err)
			return
		}
		fileDataMap[filePos] = fileData
	}

	// 调用查询函数
	queryResult, code := service.GetQueryResult(queryDataDTO.QueryItem, fileDataMap) // 返回查询结果

	time.Sleep(1000 * time.Millisecond)                                                                                                                               // 延时1s
	err := service.UpdateQueryLog(queryDataDTO.ApiUrl.ContractName, queryDataDTO.ApiUrl.ChainServiceUrl, queryDataDTO.Uid, queryDataDTO.QueryItem, code, queryResult) // 上链查询日志
	if err != nil {
		fmt.Println("上链查询日志失败", err)
	}
	if code == -1 {
		models.ResponseOK(c, "查询完成，但出现错误", queryResult)
		return
	} else {
		models.ResponseOK(c, "查询成功", queryResult)
	}

	// TODO: 缓存

}

// QueryByFieldsHandler 通过字段查询数据
func QueryByFieldsHandler(c *gin.Context) {
	type QueryByFieldsDTO struct {
		Uid         string    `json:"uId"`         // 用户ID
		ApiUrl      ApiUrlDTO `json:"apiUrl"`      // API地址
		DomainName  string    `json:"domainName"`  // 必须：数据域名称
		OrgId       string    `json:"orgId"`       // 组织ID
		Role        string    `json:"role"`        // 角色
		Name        string    `json:"name"`        // 选填：姓名
		AgeStart    int       `json:"ageStart"`    // 选填：起始年龄
		AgeEnd      int       `json:"ageEnd"`      // 选填：结束年龄
		Gender      string    `json:"gender"`      // 选填：性别
		Hospital    string    `json:"hospital"`    // 选填：医院
		Department  string    `json:"department"`  // 选填：科室
		DiseaseCode string    `json:"diseaseCode"` // 选填：疾病代码
	}

	var queryDTO QueryByFieldsDTO
	// 绑定JSON数据到结构体
	if err := c.ShouldBindJSON(&queryDTO); err != nil {
		models.ResponseError400(c, http.StatusBadRequest, "请求格式错误", err)
		return
	}

	// 验证必填字段
	queryDTO.DomainName = strings.TrimSpace(queryDTO.DomainName)
	if queryDTO.DomainName == "" {
		models.ResponseError400(c, http.StatusBadRequest, "domainName为必填项", nil)
		return
	}

	// 调用CheckAccess方法，检查是否具有权限
	allowed, err := service.CheckAccess(queryDTO.ApiUrl.ContractName, queryDTO.ApiUrl.ChainServiceUrl, queryDTO.DomainName, "read", queryDTO.OrgId, queryDTO.Role)
	if err != nil {
		models.ResponseError400(c, http.StatusBadRequest, "检查权限失败", err)
		return
	}
	if !allowed {
		models.ResponseError400(c, http.StatusBadRequest, "没有权限", nil)
		return
	}

	// 将domainName转换为domainID（格式：DOMAIN_ + domainName）
	domainID := "DOMAIN_" + queryDTO.DomainName

	// 清理其他字段的空格
	queryDTO.Name = strings.TrimSpace(queryDTO.Name)
	queryDTO.Gender = strings.TrimSpace(queryDTO.Gender)
	queryDTO.Hospital = strings.TrimSpace(queryDTO.Hospital)
	queryDTO.Department = strings.TrimSpace(queryDTO.Department)
	queryDTO.DiseaseCode = strings.TrimSpace(queryDTO.DiseaseCode)

	// 检查索引服务是否可用
	if indexer.GlobalIndexerService == nil {
		models.ResponseError400(c, http.StatusInternalServerError, "索引服务未初始化", nil)
		return
	}

	// 构建查询请求
	searchReq := indexer.SearchRequest{
		DomainID:    domainID,
		Name:        queryDTO.Name,
		AgeStart:    queryDTO.AgeStart,
		AgeEnd:      queryDTO.AgeEnd,
		Gender:      queryDTO.Gender,
		Hospital:    queryDTO.Hospital,
		Department:  queryDTO.Department,
		DiseaseCode: queryDTO.DiseaseCode,
	}

	// 如果年龄范围未设置，则设为0表示不设置年龄条件
	if queryDTO.AgeStart <= 0 && queryDTO.AgeEnd <= 0 {
		searchReq.AgeStart = 0
		searchReq.AgeEnd = 0
	} else if queryDTO.AgeStart > 0 && queryDTO.AgeEnd <= 0 {
		// 只设置了起始年龄，作为精确匹配
		searchReq.AgeEnd = queryDTO.AgeStart
	} else if queryDTO.AgeStart <= 0 && queryDTO.AgeEnd > 0 {
		// 只设置了结束年龄，起始年龄设为0
		searchReq.AgeStart = 0
	} else if queryDTO.AgeStart > queryDTO.AgeEnd {
		// 起始年龄大于结束年龄，交换它们
		searchReq.AgeStart, searchReq.AgeEnd = queryDTO.AgeEnd, queryDTO.AgeStart
	}

	// 执行查询
	searchResult, err := indexer.GlobalIndexerService.ExecuteQuery(searchReq)
	if err != nil {
		models.ResponseError400(c, http.StatusInternalServerError, "查询失败: "+err.Error(), err)
		return
	}

	// 根据txIDs查询区块链上的pos列表，并对filePoses去重
	FilePoses := make([]string, 0)
	for _, txID := range searchResult.TxIDs {
		pos, err := indexer.GlobalIndexerService.GetPosByTxID(txID)
		if err != nil {
			models.ResponseError400(c, http.StatusInternalServerError, "获取pos信息失败: "+err.Error(), err)
			return
		}
		if !slices.Contains(FilePoses, pos) {
			FilePoses = append(FilePoses, pos)
		}
	}

	// 检查FilePoses是否为空
	if len(FilePoses) == 0 {
		models.ResponseError400(c, http.StatusBadRequest, "未找到匹配的文件位置", nil)
		return
	}

	// 查询每个文件位置的数据，用map[string]string存储
	fileDataMap := make(map[string]string)
	for _, filePos := range FilePoses {
		fileCiperData, err := service.HandleGetIPFSFile(filePos, queryDTO.ApiUrl.IpfsServiceUrl)
		if err != nil {
			models.ResponseError400(c, http.StatusBadRequest, "请求IPFS错误:"+err.Error(), err)
			return
		}
		// 等待2s
		time.Sleep(2000 * time.Millisecond) // 延时2s
		// 从区块链中获取AES密钥
		aesKey, err := service.GetAesKeyFromBlockchain(queryDTO.ApiUrl.ContractName, queryDTO.ApiUrl.ChainServiceUrl, filePos)
		if err != nil {
			models.ResponseError400(c, http.StatusBadRequest, "获取AES密钥失败:"+err.Error(), err)
			return
		}
		// 解密
		fileData, err := service.AesDecrypt(fileCiperData, aesKey) // 解密文件
		if err != nil {
			models.ResponseError400(c, http.StatusBadRequest, "解密失败:"+err.Error(), err)
			return
		}
		fileDataMap[filePos] = fileData
	}

	// 构建queryItemDTO查询项结构体, 示例：
	// {
	//   "queryConcatType" : "single",
	//   "filePos" : [ [ "QmbPxKceAFixY3Kn4DHUokVVvmXzy1p2iA5nSQ89118TaW" ] ],
	//   "returnField" : [ "QmbPxKceAFixY3Kn4DHUokVVvmXzy1p2iA5nSQ89118TaW_*" ],
	//   "queryConditions" : [ [ {
	//     "field" : "name",
	//     "pos" : "QmbPxKceAFixY3Kn4DHUokVVvmXzy1p2iA5nSQ89118TaW",
	//     "compare" : "eq",
	//     "val" : "1",
	//     "type" : "string"
	//   }, {
	//     "field" : "age",
	//     "pos" : "QmbPxKceAFixY3Kn4DHUokVVvmXzy1p2iA5nSQ89118TaW",
	//     "compare" : "ge",
	//     "val" : "2",
	//     "type" : "int"
	//   }, {
	//     "field" : "age",
	//     "pos" : "QmbPxKceAFixY3Kn4DHUokVVvmXzy1p2iA5nSQ89118TaW",
	//     "compare" : "le",
	//     "val" : "10",
	//     "type" : "int"
	//   } ] ]
	// }

	queryItemDTO := service.QueryItem{
		QueryConcatType: "single",
		FilePos:         [][]string{FilePoses},
		ReturnField:     []string{FilePoses[0] + "_*"},
		QueryConditions: [][]service.QueryCondition{
			{
				{
					Field:   "name",
					Pos:     FilePoses[0],
					Compare: "eq",
					Val:     queryDTO.Name,
					Type:    "string",
				},
				{
					Field:   "age",
					Pos:     FilePoses[0],
					Compare: "ge",
					Val:     strconv.Itoa(queryDTO.AgeStart),
					Type:    "int",
				},
				{
					Field:   "age",
					Pos:     FilePoses[0],
					Compare: "le",
					Val:     strconv.Itoa(queryDTO.AgeEnd),
					Type:    "int",
				},
				{
					Field:   "gender",
					Pos:     FilePoses[0],
					Compare: "eq",
					Val:     queryDTO.Gender,
					Type:    "string",
				},
				{
					Field:   "hospital",
					Pos:     FilePoses[0],
					Compare: "eq",
					Val:     queryDTO.Hospital,
					Type:    "string",
				},
				{
					Field:   "department",
					Pos:     FilePoses[0],
					Compare: "eq",
					Val:     queryDTO.Department,
					Type:    "string",
				},
				{
					Field:   "diseaseCode",
					Pos:     FilePoses[0],
					Compare: "eq",
					Val:     queryDTO.DiseaseCode,
					Type:    "string",
				},
			},
		},
	}

	queryItemJSON, err := json.Marshal(queryItemDTO)
	if err != nil {
		models.ResponseError400(c, http.StatusInternalServerError, "构建查询项结构体失败: "+err.Error(), err)
		return
	}

	// 调用查询函数
	queryResult, code := service.GetQueryResult(string(queryItemJSON), fileDataMap) // 返回查询结果

	time.Sleep(1000 * time.Millisecond)                                                                                                                 // 延时1s
	err = service.UpdateQueryLog(queryDTO.ApiUrl.ContractName, queryDTO.ApiUrl.ChainServiceUrl, queryDTO.Uid, string(queryItemJSON), code, queryResult) // 上链查询日志
	if err != nil {
		fmt.Println("上链查询日志失败", err)
	}
	if code == -1 {
		models.ResponseOK(c, "查询完成，但出现错误", queryResult)
		return
	} else {
		models.ResponseOK(c, "查询成功", queryResult)
	}

}
