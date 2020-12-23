package discover

import (
	discoverService "github.com/cnf_core/src/modules/net/discover/service"
	"github.com/cnf_core/src/utils/error"
)

// Discover 发现服务对象
type Discover struct {
	conf    interface{}
	service discoverService.DiscoverService
}

// Build 搭建Discover服务
func (discover *Discover) Build(conf interface{}) *error.Error {
	discover.conf = conf
	discover.service.Build(conf)
	return nil
}

// RunService 启动节点通信的TCP相关服务
func (discover *Discover) RunService(chanels map[string]chan map[string]interface{}, signal chan bool) *error.Error {
	discover.service.RunService(chanels, signal)
	return nil
}

// RunDoDiscover 主动寻找种子节点或邻居节点，进行连接
func (discover *Discover) RunDoDiscover(chanels map[string]chan map[string]interface{}) {
	discover.service.RunDoDiscover(chanels)
}

// ReceiveMsg 接收消息入口
func (discover *Discover) ReceiveMsg(data interface{}) (interface{}, *error.Error) {
	return discover.service.ReceiveMsg(data)
}
