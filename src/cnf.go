package src

import (
	cnfNet "github.com/cnf_core/src/modules/net"
	config "github.com/cnf_core/src/utils/config"
	// logger "github.com/cnf_core/src/utils/logger"
)

// Build 主程序配置入口
func Build(conf interface{}) {
	// 首先把配置文件, 全局对象之类的初始化好
	config.SetConfig(conf)

	// 网络层入口构建
	cnfNet.Build()
}

// Run 主程序入口
func Run() {
	go cnfNet.Run()

	// 常驻，监听各个协程的状态

	// 挂起主协程
	c := make(chan bool)
	d := <-c
	if d {
		return
	}
}
