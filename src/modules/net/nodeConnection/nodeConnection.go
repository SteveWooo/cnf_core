package connection

import (
	nodeConnectionService "github.com/cnf_core/src/modules/net/nodeConnection/service"
	"github.com/cnf_core/src/utils/error"
)

// NodeConnection 节点连接服务的主对象
type NodeConnection struct {
	conf    interface{}
	service nodeConnectionService.NodeConnectionService
}

// Build 节点通讯服务的初始化
func (nc *NodeConnection) Build(conf interface{}) *error.Error {
	nc.conf = conf
	nc.service.Build(conf)
	return nil
}

// RunService 启动节点通信的TCP相关服务
func (nc *NodeConnection) RunService(chanels map[string]chan map[string]interface{}, signal chan bool) *error.Error {
	nc.service.RunService(chanels, signal)
	return nil
}
