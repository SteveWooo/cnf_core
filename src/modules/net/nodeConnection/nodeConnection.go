package connection

import (
	nodeConnectionService "github.com/cnf_core/src/modules/net/nodeConnection/service"
	"github.com/cnf_core/src/utils/error"
)

// NodeConnection 节点连接服务的主对象
type NodeConnection struct {
	conf    interface{}
	service nodeConnectionService.NodeConnectionService

	myPublicChanel map[string]chan map[string]interface{}
}

// Build 节点通讯服务的初始化
func (nc *NodeConnection) Build(conf interface{}, myPublicChanel map[string]chan map[string]interface{}) *error.Error {
	nc.conf = conf
	nc.myPublicChanel = myPublicChanel
	nc.service.Build(conf, nc.myPublicChanel)
	return nil
}

// RunService 启动节点通信的TCP相关服务
func (nc *NodeConnection) RunService(chanels map[string]chan map[string]interface{}, signal chan bool) *error.Error {
	nc.service.RunService(chanels, signal)
	return nil
}

// ReceiveMsg 接收消息入口
func (nc *NodeConnection) ReceiveMsg(data interface{}) (interface{}, *error.Error) {
	return nc.service.ReceiveMsg(data)
}

// HandleMsg 上面的接收消息结束后，再来判断是否要处理这条消息
func (nc *NodeConnection) HandleMsg(data interface{}) (interface{}, *error.Error) {
	return nc.service.HandleMsg(data)
}
