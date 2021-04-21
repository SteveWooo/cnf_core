package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

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

	// 载入一些初始配置
	config.Load()
	myIP := GetIP()
	for {
		if myIP == "" {
			myIP = GetIP()
			continue
		}
		break
	}
	// myIP := "192.168.10.200"

	COUNT, _ := strconv.Atoi(config.GetArg("nodeCount"))

	// 用一个大JSON来存配置
	configJSONArray, _ := config.LoadByPath("../config/conf." + myIP + "-" + strconv.Itoa(COUNT) + ".json")

	// 少结点测试用
	// configJSONArray, _ := config.LoadByPath("../config/conf." + myIP + ".json")

	// 同一个端口，才用同一套公共频道
	publicChanels := make(map[string]interface{})
	cnfObj := make([]*cnf.Cnf, COUNT)

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

	// 监听实验信号
	go HandleExamSignal()

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

// HandleExamSignal 监听实验进程发来的信号，统一使用8082http端口
func HandleExamSignal() {
	onExitsCallback := func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()         //解析参数，默认是不会解析的
		fmt.Fprintf(w, "200") //这个写入到w的是输出到客户端的
		os.Exit(0)
	}

	http.HandleFunc("/exit", onExitsCallback) //设置访问的路由
	err := http.ListenAndServe(":8082", nil)  //设置监听的端口
	if err != nil {
		logger.Debug(err)
	}
}

func getCPUSample() (idle, total uint64) {
	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return
	}
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if fields[0] == "cpu" {
			numFields := len(fields)
			for i := 1; i < numFields; i++ {
				val, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					fmt.Println("Error: ", i, fields[i], err)
				}
				total += val // tally up all the numbers to get total ticks
				if i == 4 {  // idle is the 5th field in the cpu line
					idle = val
				}
			}
			return
		}
	}
	return
}

func getCPUUsage() float64 {
	idle0, total0 := getCPUSample()
	time.Sleep(3 * time.Second)
	idle1, total1 := getCPUSample()

	idleTicks := float64(idle1 - idle0)
	totalTicks := float64(total1 - total0)
	cpuUsage := 100 * (totalTicks - idleTicks) / totalTicks

	// fmt.Printf("CPU usage is %f%% [busy: %f, total: %f]\n", cpuUsage, totalTicks-idleTicks, totalTicks)
	return cpuUsage
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
	// 日志要带点系统信息
	var m runtime.MemStats
	var httpBodySource map[string]interface{}
	for {
		timer.Sleep(1000)
		logDataLock <- true
		runtime.ReadMemStats(&m)
		httpBodySource = map[string]interface{}{
			"osMem":      m.Sys / 1024 / 1024,
			"osCPUUsage": getCPUUsage(),
			"nodes":      logData,
		}

		httpBodyJSON, _ := json.Marshal(httpBodySource)
		<-logDataLock
		req, _ := http.NewRequest("POST", "http://192.168.10.201:8081/api/update_node_status", bytes.NewReader(httpBodyJSON))
		// req, _ := http.NewRequest("POST", "http://192.168.31.164:8081/api/update_node_status", bytes.NewReader(httpBodyJSON))
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
