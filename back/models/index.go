package models

// OnChainRecord 链上记录结构，用于索引服务
type OnChainRecord struct {
	TxID     string         `json:"txId"`     // 交易ID
	Metadata RecordMetadata `json:"metadata"` // 元数据
}

// RecordMetadata 记录元数据
type RecordMetadata struct {
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
