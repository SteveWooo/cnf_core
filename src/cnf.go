package src

import (
	cnfNet "github.com/cnf_core/src/modules/net"
	config "github.com/cnf_core/src/utils/config"
	// logger "github.com/cnf_core/src/utils/logger"
)

func Build() {
	// 首先把配置文件, 全局对象之类的初始化好
	config.Load()

	// 网络层入口构建
	cnfNet.Build()
}

func Run() {
	go cnfNet.Run();

	// 干一些常驻的事情，比如监听各个协程的状态
	for {}
}