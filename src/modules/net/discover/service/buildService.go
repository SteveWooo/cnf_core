package service

import (
	"net"

	discoverModel "github.com/cnf_core/src/modules/net/discover/models"
	"github.com/cnf_core/src/utils/error"
	"github.com/cnf_core/src/utils/logger"
)

// DiscoverService 发现服务
type DiscoverService struct {
	conf       interface{}
	socketAddr *net.UDPAddr
	socketConn *net.UDPConn

	// 处理握手缓存
	pingPongCache map[string]*discoverModel.PingPongCachePackage

	// 缓存最大数量
	maxPingPong int

	// 并发监听socket的协程数
	limitProcessUDPData chan bool

	// 本届节点的公共频道
	myPublicChanel map[string]chan map[string]interface{}

	// 内部消息转发的chanel
	myPrivateChanel map[string]chan map[string]interface{}
}

// Build 构建发现服务
func (discoverService *DiscoverService) Build(conf interface{}, myPublicChanel map[string]chan map[string]interface{}) *error.Error {
	// 初始化配置
	discoverService.conf = conf
	discoverService.myPublicChanel = myPublicChanel

	// 初始化缓存变量
	discoverService.pingPongCache = make(map[string]*discoverModel.PingPongCachePackage)
	discoverService.maxPingPong = 1024

	// 并发监听socket量
	discoverService.limitProcessUDPData = make(chan bool, 5)

	confNet := discoverService.conf.(map[string]interface{})["net"]
	// ⭐JSON读取配置文件的数字时，默认会读取为float64，所以要先抓换成uint64，再换成字符串。
	address := confNet.(map[string]interface{})["ip"].(string) + ":" + confNet.(map[string]interface{})["servicePort"].(string)

	udpAddr, resolveErr := net.ResolveUDPAddr("udp", address)
	if resolveErr != nil {
		return error.New(map[string]interface{}{
			"message":   "创建UDP socket时失败",
			"originErr": resolveErr,
		})
	}
	discoverService.socketAddr = udpAddr

	logger.Info("Discover UDP 服务创建成功，即将监听 " + address)
	return nil
}
