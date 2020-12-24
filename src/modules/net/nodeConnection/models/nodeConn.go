package models

import (
	"net"
	"strings"

	"github.com/cnf_core/src/utils/error"
	"github.com/cnf_core/src/utils/sign"
)

// NodeConn 客制化连接对象
type NodeConn struct {
	// 连接对象唯一标识
	connID string

	// 连接发起方节点ID
	senderNodeID string

	// 连接发起节点的4层地址
	senderIP          string
	senderServicePort string

	// 被链接的NodeId
	targetNodeID string

	// socket handle
	Socket net.Conn
}

// GetNodeConnID 获取自己的唯一标识
func (nodeConn *NodeConn) GetNodeConnID() string {
	return nodeConn.connID
}

// GetSenderNodeID 获取自己的唯一标识
func (nodeConn *NodeConn) GetSenderNodeID() string {
	return nodeConn.senderNodeID
}

// Build 构建自己
func (nodeConn *NodeConn) Build(socket net.Conn) {
	// 先创建唯一标识
	keys := sign.GenKeys()
	nodeConn.connID = keys.(map[string]string)["publicKey"]

	// 然后设置socket
	nodeConn.Socket = socket
}

// SetRemoteAddr 设置这个连接的发起地址信息。
// @param remoteAddr ip:port
func (nodeConn *NodeConn) SetRemoteAddr(remoteAddr string) {
	ip := remoteAddr[strings.Index(remoteAddr, ":")+1:]
	servicePort := remoteAddr[:strings.Index(remoteAddr, ":")]
	nodeConn.senderIP = ip
	nodeConn.senderServicePort = servicePort
}

// SetShaker 握手包来了，找到对应的conn，设置shaker
func (nodeConn *NodeConn) SetShaker(data interface{}) *error.Error {
	tcpData := data.(map[string]interface{})["tcpData"]
	nodeConn.targetNodeID = tcpData.(map[string]interface{})["targetNodeID"].(string)
	// logger.Debug(tcpData)
	tcpDataMsg := tcpData.(map[string]interface{})["msgJSON"]
	tcpDataMsgFrom := tcpDataMsg.(map[string]interface{})["from"]
	nodeConn.senderNodeID = tcpDataMsgFrom.(map[string]interface{})["nodeID"].(string)

	return nil
}

// IsShaked 检查这个连接是否已经握手完成
func (nodeConn *NodeConn) IsShaked() bool {
	if nodeConn.senderNodeID != "" && nodeConn.targetNodeID != "" {
		return true
	}
	return false
}
