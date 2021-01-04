package connection

import (
	commonModels "github.com/cnf_core/src/modules/net/common/models"
	nodeConnectionModels "github.com/cnf_core/src/modules/net/nodeConnection/models"
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

// RunFindConnection 启动节点通信的TCP相关服务
func (nc *NodeConnection) RunFindConnection(chanels map[string]chan map[string]interface{}) *error.Error {
	nc.service.RunFindConnection(chanels)
	// 启动监控，主要防止双向链接
	go nc.service.RunMonitor()
	return nil
}

// ReceiveMsg 接收消息入口 1
func (nc *NodeConnection) ReceiveMsg(data interface{}) (interface{}, *error.Error) {
	return nc.service.ReceiveMsg(data)
}

// MasterDoTryOutBoundConnect 反射service函数
func (nc *NodeConnection) MasterDoTryOutBoundConnect(data interface{}) (interface{}, *error.Error) {
	newNode := data.(map[string]interface{})["newNode"].(*commonModels.Node)
	targetNodeID := data.(map[string]interface{})["targetNodeID"].(string)
	return nc.service.MasterDoTryOutBoundConnect(newNode, targetNodeID)
}

// SalveHandleNodeOutBoundConnectionCreateEvent 反射service函数
func (nc *NodeConnection) SalveHandleNodeOutBoundConnectionCreateEvent(nodeConnectionCreateResp map[string]interface{}) {
	nc.service.SalveHandleNodeOutBoundConnectionCreateEvent(nodeConnectionCreateResp)
}

// SalveHandleNodeInBoundConnectionCreateEvent 反射service函数
func (nc *NodeConnection) SalveHandleNodeInBoundConnectionCreateEvent(nodeConnectionCreateResp map[string]interface{}) {
	nc.service.SalveHandleNodeInBoundConnectionCreateEvent(nodeConnectionCreateResp)
}

// SendMsg 反射service函数
func (nc *NodeConnection) SendMsg(nodeConn *nodeConnectionModels.NodeConn, message string) {
	nc.service.Send(nodeConn, message)
}

// ParseTCPData 反射函数
func (nc *NodeConnection) ParseTCPData(tcpSourceData string) (interface{}, *error.Error) {
	return nc.service.ParseTCPData(tcpSourceData)
}
