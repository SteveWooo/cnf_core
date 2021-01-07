package discover

import (
	discoverService "github.com/cnf_core/src/modules/net/discover/service"
	"github.com/cnf_core/src/utils/config"
	"github.com/cnf_core/src/utils/error"
)

// Discover 发现服务对象
type Discover struct {
	conf    interface{}
	service discoverService.DiscoverService
	// 与其他节点交互的chanel
	myPublicChanel map[string]chan map[string]interface{}
}

// Build 搭建Discover服务
func (discover *Discover) Build(conf interface{}, myPublicChanel map[string]chan map[string]interface{}) *error.Error {
	discover.conf = conf
	discover.myPublicChanel = myPublicChanel
	discover.service.Build(conf, discover.myPublicChanel)
	return nil
}

// RunService 启动节点通信的TCP相关服务
func (discover *Discover) RunService(chanels map[string]chan map[string]interface{}, signal chan bool) *error.Error {
	discover.service.RunService(chanels, signal)
	return nil
}

// RunDoDiscover 主动寻找种子节点或邻居节点，进行连接
func (discover *Discover) RunDoDiscover(chanels map[string]chan map[string]interface{}) {
	if config.GetArg("logDirname") == "mats_masterAreaKad" {
		discover.service.RunDoDiscoverByMasterArea(chanels)
	} else {
		discover.service.RunDoDiscover(chanels)
	}
}

// ReceiveMsg 接收消息入口
func (discover *Discover) ReceiveMsg(data interface{}) (interface{}, *error.Error) {
	return discover.service.ReceiveMsg(data)
}

// SendMsg 发送消息接口
func (discover *Discover) SendMsg(message string, targetIP string, targetServicePort string) {
	discover.service.Send(message, targetIP, targetServicePort)
}
