//go:build linux || darwin || windows
// +build linux darwin windows

package main

import (
	"chainqa_offchain_demo/chain"
	"chainqa_offchain_demo/routers"
	"chainqa_offchain_demo/setting"
	"log"

	"fmt"
	"os"
)

const defaultConfFile = "./conf/config.ini"

func main() {
	confFile := defaultConfFile
	// 判断是否有指定配置文件
	if len(os.Args) > 2 {
		fmt.Println("使用配置文件： ", os.Args[1])
		confFile = os.Args[1]
	} else {
		fmt.Println("使用默认配置文件： ", defaultConfFile)
	}
	// 加载配置文件
	if err := setting.Init(confFile); err != nil {
		fmt.Printf("加载配置文件失败，错误信息:%v\n", err)
		return
	}

	// chainClient, err := chain.InitChainClient("")
	_, err := chain.InitChainClient("")
	if err != nil {
		log.Fatalf("初始化链客户端失败: %v", err)
	}

	// // 初始化索引服务
	// indexerSvc, err := indexer.NewIndexerService(chainClient, ctx)
	// if err != nil {
	// 	log.Fatalf("初始化索引服务失败: %v", err)
	// }

	// // 启动区块监听
	// go indexerSvc.StartBlockListener()

	// 注册路由
	r := routers.SetupRouter()
	if err := r.Run(fmt.Sprintf(":%d", setting.Conf.Port)); err != nil {
		fmt.Printf("启动服务失败，错误信息:%v\n", err)
	}
}
