package main

import (
	cnf "github.com/cnf_core/src"
	config "github.com/cnf_core/src/utils/config"
	error "github.com/cnf_core/src/utils/error"
)

func main() {
	// 首先把配置文件, 全局对象之类的初始化好
	conf, loadConfErr := config.Load()
	if loadConfErr != nil {
		error.New(map[string]interface{}{
			"message": "配置获取失败",
		})
		return
	}

	var mainCnf cnf.Cnf

	// 首先把cnf对象构建好, 里面包含了配置文件的引入
	mainCnf.Build(conf)

	// 获取cnf对象的公共chanel
	// nodeID, publicChanel := mainCnf.GetPublicChanel()

	// 启动对象的所有服务
	mainCnf.Run()
}
