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

	for i := 0; i < 1000; i++ {
		testConf, _ := config.LoadByPath("../config/test1000/node_" + strconv.Itoa(i) + ".json")
		var cnf cnf.Cnf
		cnf.Build(testConf)
		go cnf.Run()
	}

	// 挂起主协程
	c := make(chan bool)
	d := <-c
	if d {
		return
	}
}
