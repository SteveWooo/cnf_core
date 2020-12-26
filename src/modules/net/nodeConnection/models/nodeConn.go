package models

import (
	"net"
	"strings"

	"github.com/cnf_core/src/utils/error"
	"github.com/cnf_core/src/utils/sign"
	"github.com/cnf_core/src/utils/timer"
)

// NodeConn 客制化连接对象
type NodeConn struct {
	// 连接对象唯一标识
	connID string

	// 对方节点ID
	nodeID string

	// 连接发起节点的4层地址
	senderIP          string
	senderServicePort string

	// 标识这个节点是否完成握手
	shaked bool

	// 收到来自这个nodeConn的destroy请求
	needDestroy bool

	// 本地子节点的NodeID
	targetNodeID string

	// 连接类型 inBound outBound
	connType string

	// socket handle
	Socket *net.Conn

	// 建立的时间戳
	ts int64
}

// GetTs 反射ts
func (nodeConn *NodeConn) GetTs() int64 {
	return nodeConn.ts
}

// GetNodeConnID 获取自己的唯一标识
func (nodeConn *NodeConn) GetNodeConnID() string {
	return nodeConn.connID
}

// GetNodeID 获取自己的唯一标识
func (nodeConn *NodeConn) GetNodeID() string {
	return nodeConn.nodeID
}

// GetSenderIP 反射
func (nodeConn *NodeConn) GetSenderIP() string {
	return nodeConn.senderIP
}

// GetSenderServicePort 反射
func (nodeConn *NodeConn) GetSenderServicePort() string {
	return nodeConn.senderServicePort
}

// Build 构建自己
func (nodeConn *NodeConn) Build(socket *net.Conn, connType string) {
	// 先创建唯一标识
	keys := sign.GenKeys()
	nodeConn.connID = keys.(map[string]string)["publicKey"]

	// 然后设置socket
	nodeConn.Socket = socket

	nodeConn.connType = connType

	// 默认必须是未握手的
	nodeConn.shaked = false

	// 默认不需要毁灭
	nodeConn.needDestroy = false

	nodeConn.ts = timer.Now()
}

// SetDestroy 节点收到需要关闭这个连接的请求
func (nodeConn *NodeConn) SetDestroy() {
	nodeConn.needDestroy = true
}

// GetDestroy 反射
func (nodeConn *NodeConn) GetDestroy() bool {
	return nodeConn.needDestroy
}

// SetTargetNodeID 设置本地子节点的NodeID
func (nodeConn *NodeConn) SetTargetNodeID(targetNodeID string) {
	nodeConn.targetNodeID = targetNodeID
}

// GetTargetNodeID 设置本地子节点的NodeID
func (nodeConn *NodeConn) GetTargetNodeID() string {
	return nodeConn.targetNodeID
}

// GetConnType 获取自己的唯一标识
func (nodeConn *NodeConn) GetConnType() string {
	return nodeConn.connType
}

// GetSocket 反射
func (nodeConn *NodeConn) GetSocket() *net.Conn {
	return nodeConn.Socket
}

// SetRemoteAddr 设置这个连接的发起地址信息。
// @param remoteAddr ip:port
func (nodeConn *NodeConn) SetRemoteAddr(remoteAddr string) {
	ip := remoteAddr[0:strings.Index(remoteAddr, ":")]
	servicePort := remoteAddr[strings.Index(remoteAddr, ":")+1:]
	nodeConn.senderIP = ip
	nodeConn.senderServicePort = servicePort
}

// SetNodeID 设置这个连接的发起地址信息。
// @param nodeID string
func (nodeConn *NodeConn) SetNodeID(nodeID string) {
	nodeConn.nodeID = nodeID
}

// SetShaker 握手包来了，找到对应的conn，设置shaker
func (nodeConn *NodeConn) SetShaker(data interface{}) *error.Error {
	if data != nil {
		tcpData := data.(map[string]interface{})["tcpData"]
		nodeConn.targetNodeID = tcpData.(map[string]interface{})["targetNodeID"].(string)
		// logger.Debug(tcpData)
		tcpDataMsg := tcpData.(map[string]interface{})["msgJSON"]
		tcpDataMsgFrom := tcpDataMsg.(map[string]interface{})["from"]
		nodeConn.nodeID = tcpDataMsgFrom.(map[string]interface{})["nodeID"].(string)
	}

	nodeConn.shaked = true

	// 更新时间戳
	nodeConn.ts = timer.Now()

	return nil
}

// IsShaked 检查这个连接是否已经握手完成
func (nodeConn *NodeConn) IsShaked() bool {
	return nodeConn.shaked
}
