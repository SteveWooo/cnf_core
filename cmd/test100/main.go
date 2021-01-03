package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net"
	_ "net/http/pprof"
	"runtime"
	"strings"

	"log"
	"net/http"

	cnf "github.com/cnf_core/src"
	config "github.com/cnf_core/src/utils/config"
	"github.com/cnf_core/src/utils/logger"
	"github.com/cnf_core/src/utils/timer"
)

func main() {
	// 性能监控
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	myIP := GetIP()
	// myIP := "192.168.10.200"

	COUNT := 10000
	// 同一个端口，才用同一套公共频道
	publicChanels := make(map[string]interface{})
	cnfObj := make([]*cnf.Cnf, COUNT)

	// 用一个大JSON来存配置
	configJSONArray, _ := config.LoadByPath("../config/conf." + myIP + ".json")

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

	go HandleChanelLog(publicChanels)

	// 挂起主协程
	c := make(chan bool)
	d := <-c
	if d {
		return
	}
}

// HandleChanelLog 管理子节点的所有日志
func HandleChanelLog(publicChanels map[string]interface{}) {
	logDataLock := make(chan bool, 1)
	logData := make(map[string]interface{})
	for _, publicChanel := range publicChanels {
		publicChanelMap := publicChanel.(map[string]interface{})
		for nodeID, chanel := range publicChanelMap {
			// 初始化map变量
			logData[nodeID] = make(map[string]interface{})

			readLog := func(nodeID string, chanel interface{}) {
				for {
					data := <-chanel.(map[string]chan map[string]interface{})["logChanel"]
					logDataLock <- true
					logData[nodeID] = data
					<-logDataLock
				}
			}
			go readLog(nodeID, chanel)
		}
	}

	// 定时发送日志
	client := &http.Client{}
	for {
		timer.Sleep(1000)
		logDataLock <- true
		httpBodyJSON, _ := json.Marshal(logData)
		<-logDataLock
		// req, _ := http.NewRequest("POST", "http://192.168.10.200:8081/api/update_node_status", bytes.NewReader(httpBodyJSON))
		req, _ := http.NewRequest("POST", "http://192.168.31.136:8081/api/update_node_status", bytes.NewReader(httpBodyJSON))
		resp, doErr := client.Do(req)
		if doErr != nil {
			// logger.Debug(doErr)
			continue
		}
		body, _ := ioutil.ReadAll(resp.Body)
		bodyStr := string(body)
		if bodyStr != "" {
		}
	}
}

// GetIP 获取本地IP
func GetIP() string {
	interfaceName := ""
	if runtime.GOOS == "windows" {
		interfaceName = "WLAN"
	}

	if runtime.GOOS == "linux" {
		interfaceName = "eth0"
	}
	ifi, _ := net.InterfaceByName(interfaceName)
	addrs, _ := ifi.Addrs()
	for _, a := range addrs {
		ip := a.String()
		if ip[0:7] == "192.168" {
			return ip[0:strings.Index(ip, "/")]
		}
	}

	return ""
}
