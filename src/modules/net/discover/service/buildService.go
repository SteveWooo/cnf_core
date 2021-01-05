package service

import (
	"net"

	commonModels "github.com/cnf_core/src/modules/net/common/models"
	discoverModel "github.com/cnf_core/src/modules/net/discover/models"
	"github.com/cnf_core/src/utils/error"
)

// DiscoverService 发现服务
type DiscoverService struct {
	conf       interface{}
	socketAddr *net.UDPAddr
	socketConn *net.UDPConn

	// 处理握手缓存
	pingPongCacheLock chan bool
	pingPongCache     map[string]*discoverModel.PingPongCachePackage

	// 缓存最大数量
	maxPingPong int

	// 并发监听socket的协程数
	limitProcessUDPData chan bool

	// 本届节点的公共频道
	myPublicChanel map[string]chan map[string]interface{}

	// 内部消息转发的chanel
	myPrivateChanel map[string]chan map[string]interface{}

	// 主动发现服务中循环使用的变量
	runDoDiscoverTemp    map[string]interface{}
	runDoDiscoverTempNow int64

	runDoDiscoverTempProcessCacheNodeID string
	runDoDiscoverTempProcessCacheCache  *discoverModel.PingPongCachePackage
}

// Build 构建发现服务
func (discoverService *DiscoverService) Build(conf interface{}, myPublicChanel map[string]chan map[string]interface{}) *error.Error {
	// 初始化配置
	discoverService.conf = conf
	discoverService.myPublicChanel = myPublicChanel

	// 初始化缓存变量
	discoverService.pingPongCache = make(map[string]*discoverModel.PingPongCachePackage)
	discoverService.pingPongCacheLock = make(chan bool, 1)
	discoverService.maxPingPong = 1024

	// 初始化一些循环使用的内存变量
	discoverService.runDoDiscoverTemp = map[string]interface{}{
		"seed":            make(map[string]interface{}),
		"seedNode":        &commonModels.Node{},
		"seedNodeNodeID":  "",
		"isDoingPingPong": true,
	}

	// 并发监听socket量（只有master需要）
	discoverService.limitProcessUDPData = make(chan bool, 100)

	confNet := discoverService.conf.(map[string]interface{})["net"]
	// ⭐无论如何都要监听0.0.0.0就可以了
	address := "0.0.0.0:" + confNet.(map[string]interface{})["servicePort"].(string)

	udpAddr, resolveErr := net.ResolveUDPAddr("udp", address)
	if resolveErr != nil {
		return error.New(map[string]interface{}{
			"message":   "创建UDP socket时失败",
			"originErr": resolveErr,
		})
	}
	discoverService.socketAddr = udpAddr

	// logger.Info("Discover UDP 服务创建成功，即将监听 " + address)
	return nil
}
