package main

import (
	"strconv"

	cnf "github.com/cnf_core/src"
	config "github.com/cnf_core/src/utils/config"
	error "github.com/cnf_core/src/utils/error"
)

func main() {
	// 首先把配置文件, 全局对象之类的初始化好
	_, loadConfErr := config.Load()
	if loadConfErr != nil {
		error.New(map[string]interface{}{
			"message": "配置获取失败",
		})
		return
	}

	COUNT := 600

	publicChanel := make(map[string]interface{})
	cnfObj := make([]*cnf.Cnf, COUNT)

	for i := 0; i < COUNT; i++ {
		testConf, _ := config.LoadByPath("../config/test1000Local/node_" + strconv.Itoa(i) + ".json")
		var newCnf cnf.Cnf
		newCnf.Build(testConf)
		nodeID, pChanel := newCnf.GetPublicChanel()
		publicChanel[nodeID] = pChanel
		cnfObj[i] = &newCnf
		// go newCnf.Run()
	}

	for i := 0; i < COUNT; i++ {
		go cnfObj[i].RunWithPublicChanel(publicChanel)
		// go cnfObj[i].Run()
	}

	// 挂起主协程
	c := make(chan bool)
	d := <-c
	if d {
		return
	}
}
