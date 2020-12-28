package main

import (
	_ "net/http/pprof"

	"log"
	"net/http"

	cnf "github.com/cnf_core/src"
	config "github.com/cnf_core/src/utils/config"
	error "github.com/cnf_core/src/utils/error"
	"github.com/cnf_core/src/utils/logger"
)

func main() {
	// 性能监控
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// 首先把配置文件, 全局对象之类的初始化好
	_, loadConfErr := config.Load()
	if loadConfErr != nil {
		error.New(map[string]interface{}{
			"message": "配置获取失败",
		})
		return
	}

	COUNT := 20000
	// 同一个端口，才用同一套公共频道
	publicChanels := make(map[string]interface{})

	cnfObj := make([]*cnf.Cnf, COUNT)

	// 用一个大JSON来存配置
	configJSONArray, _ := config.LoadByPath("../config/test1WComplex.json")

	for i := 0; i < COUNT; i++ {

		conf := configJSONArray.([]interface{})[i]
		confNet := conf.(map[string]interface{})["net"]
		confNetServicePort := confNet.(map[string]interface{})["servicePort"].(string)

		if publicChanels[confNetServicePort] == nil {
			publicChanels[confNetServicePort] = make(map[string]interface{})
		}

		var newCnf cnf.Cnf
		newCnf.Build(configJSONArray.([]interface{})[i])

		// 设置公共频道
		nodeID, pChanel := newCnf.GetPublicChanel()
		publicChanels[confNetServicePort].(map[string]interface{})[nodeID] = pChanel
		cnfObj[i] = &newCnf
		// go newCnf.Run()
	}

	for i := 0; i < COUNT; i++ {
		conf := configJSONArray.([]interface{})[i]
		confNet := conf.(map[string]interface{})["net"]
		confNetServicePort := confNet.(map[string]interface{})["servicePort"].(string)

		signal := make(chan bool, 1)
		go (*cnfObj[i]).RunWithPublicChanel(publicChanels[confNetServicePort].(map[string]interface{}), signal)
		<-signal
		// logger.Info("已启动：" + strconv.Itoa(i) + "/" + strconv.Itoa(COUNT))
		// go cnfObj[i].Run()
	}

	logger.Debug("节点已经全部启动完成")

	for i := 0; i < COUNT; i++ {
		// 全部都要指针运行，不然偶尔会出现地址取空的问题
		go (*cnfObj[i]).DoRunDiscover()
	}

	// 挂起主协程
	c := make(chan bool)
	d := <-c
	if d {
		return
	}
}
