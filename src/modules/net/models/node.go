package models

import (
	error "github.com/cnf_core/src/utils/error"
)

type Node struct {
	nodeId string
	ip string
	servicePort int
}

func CreateNode(info map[string]interface{}) (Node, interface{}) {
	var node Node
	// 检查必备参数
	if (info["nodeId"] == nil ||
		info["ip"] == nil ||
		info["servicePort"] == nil) {
		return node, error.New(map[string]interface{}{
			"message" : "创建Node时参数不足",
		})
	}
	node.nodeId = info["nodeId"].(string)
	node.ip = info["ip"].(string)
	node.servicePort = info["servicePort"].(int)

	return node, nil
}