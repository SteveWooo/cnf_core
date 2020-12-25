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

	subConf, _ := config.LoadByPath("../config/config.json")

	var mainCnf cnf.Cnf
	var subCnf cnf.Cnf

	// 首先把cnf对象构建好, 里面包含了配置文件的引入
	mainCnf.Build(conf)
	subCnf.Build(subConf)

	// // 获取cnf对象的公共chanel
	// mainNodeID, mainPublicChanel := mainCnf.GetPublicChanel()
	// subNodeID, subPublicChanel := subCnf.GetPublicChanel()

	// mainPublicChanels := map[string]interface{}{
	// 	mainNodeID: mainPublicChanel,
	// 	// subNodeID:  subPublicChanel,
	// }

	// subPublicChanels := map[string]interface{}{
	// 	// mainNodeID: mainPublicChanel,
	// 	subNodeID: subPublicChanel,
	// }

	// // 启动对象的所有服务
	// // mainCnf.Run()
	// go mainCnf.RunWithPublicChanel(mainPublicChanels)
	// go subCnf.RunWithPublicChanel(subPublicChanels)

	go mainCnf.Run()
	go subCnf.Run()

	// 挂起主协程
	c := make(chan bool)
	d := <-c
	if d {
		return
	}
}
