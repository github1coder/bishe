package controller

import (
	"bytes"
	"chainqa_offchain_demo/models"
	"chainqa_offchain_demo/service"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/gin-gonic/gin"
)

type ApiUrlDTO struct {
	IpfsServiceUrl  string `json:"ipfsServiceUrl"`  // IPFS服务地址
	ContractName    string `json:"contractName"`    // 合约名
	ChainServiceUrl string `json:"chainServiceUrl"` // 链服务地址
}

type UploadFileDTO struct {
	ApiUrl      ApiUrlDTO `json:"apiUrl"`      // API地址
	AesKey      string    `json:"aesKey"`      // AES密钥
	FileContent string    `json:"fileContent"` // 文件内容
	FileName    string    `json:"fileName"`
	Uid         string    `json:"uId"`
}

type UploadFileDTOWithDomain struct {
	ApiUrl      ApiUrlDTO `json:"apiUrl"`      // API地址
	AesKey      string    `json:"aesKey"`      // AES密钥
	FileContent string    `json:"fileContent"` // 文件内容
	FileName    string    `json:"fileName"`
	Uid         string    `json:"uId"`
	Name        string    `json:"name"`        //用户姓名
	Age         int       `json:"age"`         //年龄
	Gender      string    `json:"gender"`      //性别
	Hospital    string    `json:"hospital"`    //医院
	Department  string    `json:"department"`  //科室
	DiseaseCode string    `json:"diseaseCode"` //疾病代码
	DomainName  string    `json:"domainName"`  // 数据域名称
	OrgId       string    `json:"orgId"`       // 组织ID
	Role        string    `json:"role"`        // 角色
}

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

// hello
func HelloHandler(c *gin.Context) {
	// 延时300ms
	// time.Sleep(30000 * time.Millisecond)
	models.ResponseOK(c, "success", "hello world")
}

// upload

func GetAesKeyHandler(c *gin.Context) {
	type GetAesKeyDTO struct {
		KeyNum int `json:"keyNum"`
	}
	var req GetAesKeyDTO

	// 绑定JSON数据到结构体
	if err := c.ShouldBindJSON(&req); err != nil {
		models.ResponseError400(c, http.StatusBadRequest, "请求格式错误", err)
		return
	}

	// 检查keyNum是否为正数，如果不是，可以设置默认值或者返回错误
	if req.KeyNum <= 0 {
		req.KeyNum = 1 // 设置默认值
	}

	// 调用service层的GetAesKey方法，传入keyNum参数
	KeyArr := make([]string, 0)
	for i := 0; i < req.KeyNum; i++ { // 使用var i声明循环变量
		key, err := service.GenerateAESKey()
		if err != nil {
			models.ResponseError400(c, http.StatusBadRequest, "获取AES密钥失败", err)
			return
		}
		KeyArr = append(KeyArr, key)
	}

	// 假设你需要返回KeyArr，以下是返回的示例代码
	models.ResponseOK(c, fmt.Sprintf("获取%d个AES密钥成功", req.KeyNum), KeyArr)
}

func UploadFileHandler(c *gin.Context) {
	// 打印日志
	fmt.Println("UploadFileHandler")
	var req UploadFileDTO

	// 绑定JSON数据到结构体
	if err := c.ShouldBindJSON(&req); err != nil {
		models.ResponseError400(c, http.StatusBadRequest, "请求格式错误", err)
		return
	}
	req.Uid = strings.TrimSpace(req.Uid)                 // 去除空格
	req.FileName = strings.TrimSpace(req.FileName)       // 去除空格
	req.FileContent = strings.TrimSpace(req.FileContent) // 去除空格
	req.AesKey = strings.TrimSpace(req.AesKey)           // 去除空格

	uploadHandler(req, c)
}

func uploadHandler(req UploadFileDTO, c *gin.Context) {
	// 1. 使用AES密钥加密文件内容
	cipherText, err := service.AesEncrypt(req.FileContent, req.AesKey)

	type UploadFileErrorVO struct {
		FileName string `json:"fileName"` // 文件名
		Error    string `json:"error"`    // 错误信息
	}
	var errorVO UploadFileErrorVO
	errorVO.FileName = req.FileName
	if err != nil {
		errorVO.Error = err.Error()
		models.ResponseError400(c, http.StatusBadRequest, fmt.Sprintf("加密文件内容失败: %s", req.FileName), errorVO)
		return
	}
	fmt.Println("cipherText:", cipherText)

	// 2. 将加密后的文件内容上传到IPFS
	cid, err := service.HandleUploadIPFSFile(cipherText, req.ApiUrl.IpfsServiceUrl) // 调用HandleUploadIPFSFile方法上传文件
	if err != nil {
		errorVO.Error = err.Error()
		models.ResponseError400(c, http.StatusBadRequest, fmt.Sprintf("上传文件到IPFS失败: %s", req.FileName), errorVO)
		return
	}
	fmt.Println("cid:", cid)

	// 3. 上传数字信封
	// 从区块链中获取公钥
	publicKey, err := service.GetPublicKeyFromBlockchain(req.ApiUrl.ContractName, req.ApiUrl.ChainServiceUrl)
	if err != nil {
		errorVO.Error = err.Error()
		models.ResponseError400(c, http.StatusBadRequest, "获取公钥失败", errorVO)
		return
	}

	Envelope, err := service.RSAEncryptAndReturnEnvelop(req.ApiUrl.ContractName, req.AesKey, publicKey)
	if err != nil {
		errorVO.Error = err.Error()
		models.ResponseError400(c, http.StatusBadRequest, fmt.Sprintf("生成数字信封失败: %s", req.FileName), errorVO)
		return
	}
	fmt.Println("Envelope:", Envelope)
	// 等待1s
	time.Sleep(2000 * time.Millisecond) // 延时1s
	fmt.Println("开始上传数字信封到区块链")
	// 将数字信封上传到区块链
	err = service.UploadEnvelopeToBlockchain(req.ApiUrl.ContractName, req.ApiUrl.ChainServiceUrl, Envelope, cid, req.Uid)
	if err != nil {
		errorVO.Error = err.Error()
		models.ResponseError400(c, http.StatusBadRequest, fmt.Sprintf("上传数字信封到区块链失败: %s", req.FileName), errorVO)
		return
	}
	fmt.Println("上传数字信封到区块链成功")

	// 4. 返回文件上传结果
	type UploadFileSuccessVO struct {
		FileName string `json:"fileName"` // 文件名
		Pos      string `json:"pos"`      // ipfs地址
	}

	var successVO UploadFileSuccessVO
	successVO.FileName = req.FileName
	successVO.Pos = cid
	models.ResponseOK(c, fmt.Sprintf("上传文件%s到IPFS成功", req.FileName), successVO)
}

func UploadFileWithDomainHandler(c *gin.Context) {
	fmt.Println("UploadFileWithDomainHandler")
	var req UploadFileDTOWithDomain
	// 绑定JSON数据到结构体
	if err := c.ShouldBindJSON(&req); err != nil {
		models.ResponseError400(c, http.StatusBadRequest, "请求格式错误", err)
		return
	}
	req.Uid = strings.TrimSpace(req.Uid)                 // 去除空格
	req.FileName = strings.TrimSpace(req.FileName)       // 去除空格
	req.FileContent = strings.TrimSpace(req.FileContent) // 去除空格
	req.AesKey = strings.TrimSpace(req.AesKey)           // 去除空格
	req.Name = strings.TrimSpace(req.Name)               // 去除空格
	req.Age = int(req.Age)                               // 去除空格
	req.Gender = strings.TrimSpace(req.Gender)           // 去除空格
	req.Hospital = strings.TrimSpace(req.Hospital)       // 去除空格
	req.Department = strings.TrimSpace(req.Department)   // 去除空格
	req.DiseaseCode = strings.TrimSpace(req.DiseaseCode) // 去除空格
	req.DomainName = strings.TrimSpace(req.DomainName)   // 去除空格
	req.OrgId = strings.TrimSpace(req.OrgId)             // 去除空格
	req.Role = strings.TrimSpace(req.Role)               // 去除空格
	uploadHandlerWithDomain(req, c)
}

func uploadHandlerWithDomain(req UploadFileDTOWithDomain, c *gin.Context) {
	// 1. 调用CheckAccess方法，检查是否具有权限
	allowed, err := service.CheckAccess(req.ApiUrl.ContractName, req.ApiUrl.ChainServiceUrl, req.DomainName, "write", req.OrgId, req.Role)
	if err != nil {
		models.ResponseError400(c, http.StatusBadRequest, "检查权限失败", err)
		return
	}
	if !allowed {
		models.ResponseError400(c, http.StatusBadRequest, "没有权限", nil)
		return
	}
	// 2. 使用AES密钥加密文件内容
	cipherText, err := service.AesEncrypt(req.FileContent, req.AesKey)

	type UploadFileErrorVO struct {
		FileName string `json:"fileName"` // 文件名
		Error    string `json:"error"`    // 错误信息
	}
	var errorVO UploadFileErrorVO
	errorVO.FileName = req.FileName
	if err != nil {
		errorVO.Error = err.Error()
		models.ResponseError400(c, http.StatusBadRequest, fmt.Sprintf("加密文件内容失败: %s", req.FileName), errorVO)
		return
	}
	fmt.Println("cipherText:", cipherText)

	// 3. 将加密后的文件内容上传到IPFS
	cid, err := service.HandleUploadIPFSFile(cipherText, req.ApiUrl.IpfsServiceUrl) // 调用HandleUploadIPFSFile方法上传文件
	if err != nil {
		errorVO.Error = err.Error()
		models.ResponseError400(c, http.StatusBadRequest, fmt.Sprintf("上传文件到IPFS失败: %s", req.FileName), errorVO)
		return
	}
	fmt.Println("cid:", cid)

	// 4. 上传数字信封
	// 从区块链中获取公钥
	publicKey, err := service.GetPublicKeyFromBlockchain(req.ApiUrl.ContractName, req.ApiUrl.ChainServiceUrl)
	if err != nil {
		errorVO.Error = err.Error()
		models.ResponseError400(c, http.StatusBadRequest, "获取公钥失败", errorVO)
		return
	}

	Envelope, err := service.RSAEncryptAndReturnEnvelop(req.ApiUrl.ContractName, req.AesKey, publicKey)
	if err != nil {
		errorVO.Error = err.Error()
		models.ResponseError400(c, http.StatusBadRequest, fmt.Sprintf("生成数字信封失败: %s", req.FileName), errorVO)
		return
	}
	fmt.Println("Envelope:", Envelope)
	// 等待1s
	time.Sleep(2000 * time.Millisecond) // 延时1s
	fmt.Println("开始上传数字信封到区块链")
	// 将数字信封上传到区块链
	var dataDigtalEnvelop DataDigtalEnvelop
	dataDigtalEnvelop = DataDigtalEnvelop{
		Uid:         req.Uid,
		Pos:         cid,
		Envelop:     Envelope,
		Name:        req.Name,
		Age:         int(req.Age),
		Gender:      req.Gender,
		Hospital:    req.Hospital,
		Department:  req.Department,
		DiseaseCode: req.DiseaseCode,
		DomainID:    "DOMAIN_" + req.DomainName,
	}

	envelopJsonBytes, err := json.Marshal(dataDigtalEnvelop)
	if err != nil {
		errorVO.Error = err.Error()
		models.ResponseError400(c, http.StatusBadRequest, fmt.Sprintf("序列化数字信封失败: %s", req.FileName), errorVO)
		return
	}
	envelopJsonStr := string(envelopJsonBytes)
	fmt.Println("dataDigtalEnvelop json:", envelopJsonStr)

	err = service.UploadEnvelopeToBlockchainWithDomain(req.ApiUrl.ContractName, req.ApiUrl.ChainServiceUrl, envelopJsonStr)
	if err != nil {
		errorVO.Error = err.Error()
		models.ResponseError400(c, http.StatusBadRequest, fmt.Sprintf("上传数字信封到区块链失败: %s", req.FileName), errorVO)
		return
	}
	fmt.Println("上传数字信封到区块链成功")

	// 4. 返回文件上传结果
	type UploadFileSuccessVO struct {
		FileName string `json:"fileName"` // 文件名
		Pos      string `json:"pos"`      // ipfs地址
	}

	var successVO UploadFileSuccessVO
	successVO.FileName = req.FileName
	successVO.Pos = cid
	models.ResponseOK(c, fmt.Sprintf("上传文件%s到IPFS成功", req.FileName), successVO)
}

func UploadDataFileWithoutCheckHandler(c *gin.Context) {
	fmt.Println("UploadDataFileWithoutCheckHandler")

	// 1. 解析form-data字段
	apiUrl := ApiUrlDTO{
		IpfsServiceUrl:  c.PostForm("ipfsServiceUrl"),
		ContractName:    c.PostForm("contractName"),
		ChainServiceUrl: c.PostForm("chainServiceUrl"),
	}
	aesKey := c.PostForm("aesKey")
	uId := c.PostForm("uId")

	// 2. 获取文件内容
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		models.ResponseError400(c, http.StatusBadRequest, "文件上传失败", err)
		return
	}
	defer file.Close()
	fileName := header.Filename

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		models.ResponseError400(c, http.StatusBadRequest, "读取文件内容失败", err)
		return
	}

	// 3. 根据文件类型解析
	ext := strings.ToLower(filepath.Ext(fileName))
	var records [][]string
	switch ext {
	case ".csv":
		records, err = parseCSVFile(fileBytes)
		if err != nil {
			models.ResponseError400(c, http.StatusBadRequest, "CSV文件解析失败: "+err.Error(), err)
			return
		}
		if len(records) == 0 {
			models.ResponseError400(c, http.StatusBadRequest, "CSV文件内容为空: "+err.Error(), nil)
			return
		}
	case ".xlsx":
		records, err = parseExcelFile(fileBytes)
		if err != nil {
			models.ResponseError400(c, http.StatusBadRequest, "Excel文件解析失败: "+err.Error(), nil)
			return
		}
		if len(records) == 0 {
			models.ResponseError400(c, http.StatusBadRequest, "Excel文件内容为空"+err.Error(), nil)
			return
		}
	case ".txt":
		records, err = parseTxtFile(fileBytes)
		if err != nil {
			models.ResponseError400(c, http.StatusBadRequest, "TXT文件解析失败: "+err.Error(), nil)
			return
		}
		if len(records) == 0 {
			models.ResponseError400(c, http.StatusBadRequest, "TXT文件内容为空"+err.Error(), nil)
			return
		}
	default:
		models.ResponseError400(c, http.StatusBadRequest, "仅支持CSV（.csv）或Excel文件（.xlsx）", nil)
		return
	}

	fileContent := changeToFileContentString(records)

	uploadHandler(UploadFileDTO{
		ApiUrl: ApiUrlDTO{
			IpfsServiceUrl:  apiUrl.IpfsServiceUrl,
			ContractName:    apiUrl.ContractName,
			ChainServiceUrl: apiUrl.ChainServiceUrl,
		},
		AesKey:      aesKey,
		FileContent: fileContent,
		FileName:    fileName,
		Uid:         uId,
	}, c)
}

func UploadCSVFileHandler(c *gin.Context) {
	fmt.Println("UploadCSVFileHandler")

	// 1. 解析form-data字段
	apiUrl := ApiUrlDTO{
		IpfsServiceUrl:  c.PostForm("ipfsServiceUrl"),
		ContractName:    c.PostForm("contractName"),
		ChainServiceUrl: c.PostForm("chainServiceUrl"),
	}
	aesKey := c.PostForm("aesKey")
	uId := c.PostForm("uId")

	// 2. 获取文件内容
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		models.ResponseError400(c, http.StatusBadRequest, "文件上传失败", err)
		return
	}
	defer file.Close()
	fileName := header.Filename

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		models.ResponseError400(c, http.StatusBadRequest, "读取文件内容失败", err)
		return
	}

	// 3. 根据文件类型解析
	ext := strings.ToLower(filepath.Ext(fileName))
	var records [][]string
	switch ext {
	case ".csv":
		records, err = parseCSVFile(fileBytes)
		if err != nil {
			models.ResponseError400(c, http.StatusBadRequest, "CSV文件解析失败: "+err.Error(), err)
			return
		}
		if len(records) == 0 {
			models.ResponseError400(c, http.StatusBadRequest, "CSV文件内容为空: "+err.Error(), nil)
			return
		}
	case ".xlsx":
		records, err = parseExcelFile(fileBytes)
		if err != nil {
			models.ResponseError400(c, http.StatusBadRequest, "Excel文件解析失败: "+err.Error(), nil)
			return
		}
		if len(records) == 0 {
			models.ResponseError400(c, http.StatusBadRequest, "Excel文件内容为空"+err.Error(), nil)
			return
		}
	case ".txt":
		records, err = parseTxtFile(fileBytes)
		if err != nil {
			models.ResponseError400(c, http.StatusBadRequest, "TXT文件解析失败: "+err.Error(), nil)
			return
		}
		if len(records) == 0 {
			models.ResponseError400(c, http.StatusBadRequest, "TXT文件内容为空"+err.Error(), nil)
			return
		}
	default:
		models.ResponseError400(c, http.StatusBadRequest, "仅支持CSV（.csv）或Excel文件（.xlsx）", nil)
		return
	}

	// 检查文件内容是否合法
	records, err = checkAndParseFile(records)
	if err != nil {
		models.ResponseError400(c, http.StatusBadRequest, "文件内容不合法，请检查 | "+err.Error(), nil)
		return
	}

	fileContent := changeToFileContentString(records)

	uploadHandler(UploadFileDTO{
		ApiUrl: ApiUrlDTO{
			IpfsServiceUrl:  apiUrl.IpfsServiceUrl,
			ContractName:    apiUrl.ContractName,
			ChainServiceUrl: apiUrl.ChainServiceUrl,
		},
		AesKey:      aesKey,
		FileContent: fileContent,
		FileName:    fileName,
		Uid:         uId,
	}, c)
}

// 解析CSV
func parseCSVFile(fileBytes []byte) ([][]string, error) {
	csvReader := csv.NewReader(strings.NewReader(string(fileBytes)))
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}
	return records, nil
}

// 解析TXT
func parseTxtFile(fileBytes []byte) ([][]string, error) {
	var records [][]string
	content := string(fileBytes)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		// lines去除首尾空格
		line = strings.TrimSpace(line)
		cols := strings.Split(line, " ")
		records = append(records, cols)
	}
	return records, nil
}

// 解析Excel
func parseExcelFile(fileBytes []byte) ([][]string, error) {
	f, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("Excel文件无工作表")
	}
	firstSheet := sheets[0]
	rows, err := f.GetRows(firstSheet)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// 检查文件是否合法
func checkAndParseFile(fileContent [][]string) ([][]string, error) {

	if len(fileContent) == 0 {
		return nil, fmt.Errorf("文件内容为空")
	}

	// 检查表头
	header := fileContent[0]
	if len(header) < 1 {
		return nil, fmt.Errorf("表头信息为空")
	}

	// 检查表头是否全为非空+（只允许英文、下划线、字母）+不允许纯为*
	colSet := make(map[string]bool)

	for idx, col := range header {
		col = strings.TrimSpace(col) // 去除首尾空格
		// col不能为空
		if col == "" {
			return nil, fmt.Errorf("表头存在空列，位置: %d", idx+1)
		}
		// col不能为*
		if col == "*" {
			return nil, fmt.Errorf("表头存在*列，位置: %d", idx+1)
		}
		// col不能包含中文
		for _, r := range col {
			if r >= 0x4E00 && r <= 0x9FA5 {
				return nil, fmt.Errorf("表头列名 %s 不能包含中文，位置: %d", col, idx+1)
			}
		}
		// col不能以_开头或结尾
		if strings.HasPrefix(col, "_") || strings.HasSuffix(col, "_") {
			return nil, fmt.Errorf("表头列名 %s 不能以_开头或结尾，位置: %d", col, idx+1)
		}
		// col不能重复
		// 检查是否重复
		if _, exists := colSet[col]; exists {
			return nil, fmt.Errorf("表头列名存在重复: %s，位置: %d", col, idx+1)
		}
		colSet[col] = true
	}

	// 将header的首尾空格去掉，空格转为_，
	for i, col := range header {
		header[i] = strings.TrimSpace(col)
		header[i] = strings.ReplaceAll(header[i], " ", "_")
	}

	// 检查每一行
	for i, row := range fileContent[1:] {
		// 如果row的列数为0，那么删除这个row
		if len(row) != len(header) {
			return nil, fmt.Errorf("第 %d 行数据列数不匹配", i+2)
		}
		// 列内容不能纯空
		for j, val := range row {
			val = strings.TrimSpace(val)
			if val == "" {
				return nil, fmt.Errorf("第 %d 行第 %d 列数据为空", i+2, j+1)
			}
		}
	}

	// 将单元格内容的首尾空格去掉，并将内容中的空格替换为_
	for i, row := range fileContent {
		for j, val := range row {
			val = strings.TrimSpace(val)
			val = strings.ReplaceAll(val, " ", "_")
			row[j] = val
		}
		fileContent[i] = row
	}

	return fileContent, nil
}

func changeToFileContentString(fileContent [][]string) string {
	var sb strings.Builder
	for idx, row := range fileContent {
		for i, col := range row {
			sb.WriteString(col)
			if i < len(row)-1 {
				sb.WriteString(" ")
			}
		}
		if idx < len(fileContent)-1 {
			sb.WriteString("\n")
		}
	}
	s := sb.String()
	// 去除空格
	s = strings.TrimSpace(s)
	return s
}
