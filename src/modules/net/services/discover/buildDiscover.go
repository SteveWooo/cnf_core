package discover

import (
	"net"

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
