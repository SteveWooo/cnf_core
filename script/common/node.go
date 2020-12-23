package common

import (
	error "github.com/cnf_core/src/utils/error"
	"github.com/cnf_core/src/utils/timer"
)

// Node 全局Node的统一数据结构
type Node struct {
	nodeID      string
	ip          string
	servicePort string
	ts          int64
}

// GetNodeID 反射
func (n *Node) GetNodeID() string {
	return n.nodeID
}

// CreateNode 创建一个Node实例化对象
func CreateNode(info map[string]interface{}) (*Node, *error.Error) {
	var node Node
	// 检查必备参数
	if info["nodeID"] == nil ||
		info["ip"] == nil ||
		info["servicePort"] == nil {
		return nil, error.New(map[string]interface{}{
			"message": "创建Node时参数不足",
		})
	}
	node.nodeID = info["nodeID"].(string)
	node.ip = info["ip"].(string)
	node.servicePort = info["servicePort"].(string)
	node.ts = timer.Now()

	return &node, nil
}

// GetIP 反射
func (n *Node) GetIP() string {
	return n.ip
}

// GetServicePort 反射
func (n *Node) GetServicePort() string {
	return n.servicePort
}
