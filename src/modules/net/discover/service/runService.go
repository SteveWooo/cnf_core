package service

import (
	"net"
	"strconv"

	"github.com/cnf_core/src/utils/error"
)

// RunService 启动UDP服务器，并持续监听状态。
// @param chanel 与消息队列通信的发现服务专用管道
// @param signal 用于通知上级服务udp服务已就绪
func (discoverService *DiscoverService) RunService(chanels map[string]chan map[string]interface{}, signal chan bool) *error.Error {
	discoverService.myPrivateChanel = make(map[string]chan map[string]interface{})
	discoverService.myPrivateChanel = chanels
	// 处理这个模块的任务队列
	go discoverService.HandleDiscoverEventChanel()

	// 非主节点，不需要监听socket
	confNet := discoverService.conf.(map[string]interface{})["net"]
	if confNet.(map[string]interface{})["masterServer"] != "true" {
		// logger.Debug("非主节点，无需监听")
		signal <- true
		return nil
	}

	udpConn, listenErr := net.ListenUDP("udp", discoverService.socketAddr)

	if listenErr != nil {
		return error.New(map[string]interface{}{
			"message":   "监听UDP端口失败",
			"originErr": listenErr,
		})
	}
	defer udpConn.Close()
	// 也要把socket赋值给shaker
	discoverService.socketConn = udpConn

	signal <- true

	// 暴力读取udp数据
	for {
		// 压入数据，填缓冲池。如果缓冲池满了，就不会有下面的协程创建了，意味着会丢掉过多的UDP包
		discoverService.limitProcessUDPData <- true
		go discoverService.ProcessUDPData(chanels["receiveDiscoverMsgChanel"], udpConn)
	}
}

// ProcessUDPData 协程，负责读取udp数据
func (discoverService *DiscoverService) ProcessUDPData(chanel chan map[string]interface{}, udpConn *net.UDPConn) *error.Error {
	// 源数据
	udpSourceDataByte := make([]byte, 1024)

	length, info, readUDPErr := udpConn.ReadFromUDP(udpSourceDataByte) // 挂起
	if readUDPErr != nil {
		return error.New(map[string]interface{}{
			"message":   "读取UDP数据错误",
			"originErr": readUDPErr,
		})
	}
	message := string(udpSourceDataByte[:length])

	// 提取出信息和报文头内容的数据（但数据未被格式化
	udpSourceData := make(map[string]interface{})
	udpSourceData["message"] = message
	udpSourceData["sourceIP"] = info.IP.String()
	udpSourceData["sourceServicePort"] = strconv.Itoa(info.Port)

	// 🐎这里会卡一下，影响UDP端口读取性能
	udpData, parseUDPError := discoverService.ParseUDPData(udpSourceData)
	if parseUDPError != nil {
		return parseUDPError
	}

	// 把消息推送到消息队列中。只有队列不满的情况下，这条协程才会往下走
	chanel <- udpData.(map[string]interface{})

	// 当数据确实推送到消息队列后，才能释放这条协程，创建新的监听UDP数据的协程。
	<-discoverService.limitProcessUDPData

	return nil
}

// HandleDiscoverEventChanel 处理发现服务的所有事件
func (discoverService *DiscoverService) HandleDiscoverEventChanel() {
	// 变量全体初始化，不要在循环体内创建
	var bucketOperate interface{}
	var receiveErr *error.Error
	eventData := make(map[string]interface{})
	for {
		eventData = <-discoverService.myPrivateChanel["discoverEventChanel"]

		if eventData["event"] == "receiveMsg" {
			// 交给发现服务模块处理消息，把结果透传回来即可
			bucketOperate, receiveErr = discoverService.ReceiveMsg(eventData["udpData"])
			if receiveErr != nil {
				// 不处理
				// logger.Warn(receiveErr.GetMessage())
				return
			}

			// 处理路由Bucket逻辑
			if bucketOperate != nil {
				// 由于Bucket操作有可能在tcp消息中出现，所有需要用一个chanel锁住。
				// logger.Debug(bucketOperate.(map[string]interface{}))
				discoverService.myPrivateChanel["bucketOperateChanel"] <- bucketOperate.(map[string]interface{})
			}
		}

		if eventData["event"] == "processSeed" {
			discoverService.processSeed(discoverService.myPrivateChanel)
		}

		if eventData["event"] == "processDoingPingCache" {
			discoverService.processDoingPingCache(discoverService.myPrivateChanel)
		}

		if eventData["event"] == "doFindNeighbor" {
			discoverService.doFindNeighbor(eventData["findingNodeID"].(string))
		}
	}
}
