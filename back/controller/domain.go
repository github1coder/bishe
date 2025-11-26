package controller

import (
	"chainqa_offchain_demo/models"
	"chainqa_offchain_demo/service"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DomainDTO struct {
	ApiUrl        ApiUrlDTO `json:"apiUrl"`        // API地址
	DomainName    string    `json:"domainName"`    // 数据域名称
	DomainMembers string    `json:"domainMembers"` // 数据域成员
	DomainPolicy  string    `json:"domainPolicy"`  // 数据域策略
	OrgId         string    `json:"orgId"`         // 组织ID
	Role          string    `json:"role"`          // 角色
}

func CreateDomainHandler(c *gin.Context) {
	fmt.Println("CreateDomainHandler")
	var req DomainDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		models.ResponseError400(c, http.StatusBadRequest, "请求格式错误", err)
		return
	}
	createDomain(req, c)
}

func createDomain(req DomainDTO, c *gin.Context) {
	fmt.Println("createDomain")
	err := service.CreateDomain(req.ApiUrl.ContractName, req.ApiUrl.ChainServiceUrl, req.DomainName, req.DomainPolicy, req.OrgId)
	if err != nil {
		models.ResponseError400(c, http.StatusBadRequest, "创建数据域失败", err)
		return
	}
	models.ResponseOK(c, "创建数据域成功", nil)
}

func UpdateDomainMetadataHandler(c *gin.Context) {
	fmt.Println("UpdateDomainMetadataHandler")
	var req DomainDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		models.ResponseError400(c, http.StatusBadRequest, "请求格式错误", err)
		return
	}
	updateDomainMetadata(req, c)
}

func updateDomainMetadata(req DomainDTO, c *gin.Context) {
	fmt.Println("updateDomainMetadata")
	err := service.UpdateDomainMetadata(req.ApiUrl.ContractName, req.ApiUrl.ChainServiceUrl, req.DomainName, req.DomainMembers, req.DomainPolicy, req.OrgId)
	if err != nil {
		models.ResponseError400(c, http.StatusBadRequest, "更新数据域元数据失败", err)
		return
	}
	models.ResponseOK(c, "更新数据域元数据成功", nil)
}

func QueryMyDomainsHandler(c *gin.Context) {
	fmt.Println("QueryMyDomainsHandler")
	var req DomainDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		models.ResponseError400(c, http.StatusBadRequest, "请求格式错误", err)
		return
	}
	queryMyDomains(req, c)
}

func queryMyDomains(req DomainDTO, c *gin.Context) {
	fmt.Println("queryMyDomains")
	queryResult, err := service.QueryMyDomains(req.ApiUrl.ContractName, req.ApiUrl.ChainServiceUrl, req.OrgId)
	if err != nil {
		models.ResponseError400(c, http.StatusBadRequest, "查询数据域失败", err)
		return
	}
	models.ResponseOK(c, "查询数据域成功", queryResult)
}

func QueryMyManagedDomainsHandler(c *gin.Context) {
	fmt.Println("QueryMyManagedDomainsHandler")
	var req DomainDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		models.ResponseError400(c, http.StatusBadRequest, "请求格式错误", err)
		return
	}
	queryMyManagedDomains(req, c)
}

func queryMyManagedDomains(req DomainDTO, c *gin.Context) {
	fmt.Println("queryMyManagedDomains")
	queryResult, err := service.QueryMyManagedDomains(req.ApiUrl.ContractName, req.ApiUrl.ChainServiceUrl, req.OrgId)
	if err != nil {
		models.ResponseError400(c, http.StatusBadRequest, "查询数据域失败", err)
		return
	}
	models.ResponseOK(c, "查询数据域成功", queryResult)
}

func QueryDomainInfoHandler(c *gin.Context) {
	fmt.Println("QueryDomainInfoHandler")
	var req DomainDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		models.ResponseError400(c, http.StatusBadRequest, "请求格式错误", err)
		return
	}
	queryDomainInfo(req, c)
}

func queryDomainInfo(req DomainDTO, c *gin.Context) {
	fmt.Println("queryDomainInfo")
	queryResult, err := service.QueryDomainInfo(req.ApiUrl.ContractName, req.ApiUrl.ChainServiceUrl, req.DomainName, req.OrgId)
	if err != nil {
		models.ResponseError400(c, http.StatusBadRequest, "查询数据域失败", err)
		return
	}
	models.ResponseOK(c, "查询数据域成功", queryResult)
}
