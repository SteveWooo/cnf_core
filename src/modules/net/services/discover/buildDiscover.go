package discover

import (
	"net"
	"strconv"

	discoverModel "github.com/cnf_core/src/modules/net/models/discover"
	config "github.com/cnf_core/src/utils/config"
	error "github.com/cnf_core/src/utils/error"
	logger "github.com/cnf_core/src/utils/logger"
)

// 本地UDP服务socket
var localSocketAddr *net.UDPAddr

// 握手包发送、接收对象
var shaker discoverModel.Shaker

// 握手用的缓存
var pingPongCache map[string]*discoverModel.PingPongCachePackage = make(map[string]*discoverModel.PingPongCachePackage)

// MaxPingPong 缓存最大值
var MaxPingPong int = 1024

// Build 节点发现服务的构建入口, 主要初始化配置项
func Build() *error.Error {
	err := createUDPServer()
	if err != nil {
		return err
	}

	// 给握手节点赋值初始化
	shaker.SetNodeID(config.GetNodeID())

	return nil
}

// createUdpServer 为了初始化udp socket
func createUDPServer() *error.Error {
	conf := config.GetConfig()
	confNet := conf.(map[string]interface{})["net"]
	// ⭐JSON读取配置文件的数字时，默认会读取为float64，所以要先抓换成uint64，再换成字符串。
	address := confNet.(map[string]interface{})["ip"].(string) + ":" + confNet.(map[string]interface{})["servicePort"].(string)

	udpAddr, resolveErr := net.ResolveUDPAddr("udp", address)
	if resolveErr != nil {
		return error.New(map[string]interface{}{
			"message":   "创建UDP socket时失败",
			"originErr": resolveErr,
		})
	}
	localSocketAddr = udpAddr

	logger.Info("Discover Udp 服务创建成功，即将监听 " + address)

	return nil
}

// 用一个缓冲池，限制协程数
var limitProcessUDPData chan bool

// Run 启动UDP服务器，并持续监听状态。
// @param chanel 与消息队列通信的发现服务专用管道
// @param signal 用于通知上级服务udp服务已就绪
func RunService(chanels map[string]chan map[string]interface{}, signal chan bool) *error.Error {
	limitProcessUDPData = make(chan bool, 5)
	udpConn, listenErr := net.ListenUDP("udp", localSocketAddr)
	defer udpConn.Close()

	if listenErr != nil {
		return error.New(map[string]interface{}{
			"message":   "监听UDP端口失败",
			"originErr": listenErr,
		})
	}
	// 也要把socket赋值给shaker
	shaker.SetSocketConn(udpConn)

	signal <- true

	// 暴力读取udp数据
	for {
		// 压入数据，填缓冲池。如果缓冲池满了，就不会有下面的协程创建了，意味着会丢掉过多的UDP包
		limitProcessUDPData <- true
		go processUDPData(chanels["discoverMsgChanel"], udpConn)
	}
}

// processUdpData 协程，负责读取udp数据
func processUDPData(chanel chan map[string]interface{}, udpConn *net.UDPConn) *error.Error {
	// 源数据
	udpSourceData := make([]byte, 1024)

	length, info, readUDPErr := udpConn.ReadFromUDP(udpSourceData) // 挂起
	if readUDPErr != nil {
		return error.New(map[string]interface{}{
			"message":   "读取UDP数据错误",
			"originErr": readUDPErr,
		})
	}
	message := string(udpSourceData[:length])

	// 提取出信息和报文头内容的数据（但数据未被格式化
	udpData := make(map[string]interface{})
	udpData["message"] = message
	udpData["sourceIP"] = info.IP.String()
	udpData["sourceServicePort"] = strconv.Itoa(info.Port)

	// 把消息推送到消息队列中。只有队列不满的情况下，这条协程才会往下走
	chanel <- udpData

	// 当数据确实推送到消息队列后，才能释放这条协程，创建新的监听UDP数据的协程。
	<-limitProcessUDPData

	return nil
}
