package services

import (
	nodeConnectionModels "github.com/cnf_core/src/modules/net/nodeConnection/models"
	"github.com/cnf_core/src/utils/error"
)

// ReceiveMsg 处理接收nodeConnetion消息，这里主要做分流
func (ncService *NodeConnectionService) ReceiveMsg(data interface{}) (interface{}, *error.Error) {
	return data, nil
}

// HandleMsg 从消息队列走一圈回来，判断是否需要处理这条消息的事件
func (ncService *NodeConnectionService) HandleMsg(data interface{}) (interface{}, *error.Error) {
	tcpData := data.(map[string]interface{})["tcpData"]
	// logger.Debug(data.(map[string]interface{})["nodeConn"])
	if tcpData.(map[string]interface{})["event"] == "shakeEvent" {
		return ncService.HandleShakeEvent(data)
	}

	// 先检查这个conn是否已经shake过

	return nil, nil
}

// HandleShakeEvent 处理接收nodeConnetion消息
func (ncService *NodeConnectionService) HandleShakeEvent(data interface{}) (interface{}, *error.Error) {
	nodeConn := data.(map[string]interface{})["nodeConn"].(*nodeConnectionModels.NodeConn)
	nodeConnID := nodeConn.GetNodeConnID()

	updated := false
	// 更新连接池的内容
	for i := 0; i < len(ncService.inBoundConn); i++ {
		if ncService.inBoundConn[i] == nil {
			continue
		}

		if ncService.inBoundConn[i].GetNodeConnID() == nodeConnID {
			ncService.inBoundConn[i].SetShaker(data)
			updated = true
		}
	}

	// Todo 找不到nodeConn，而且是salve服务器，那就创建一个NodeConn
	if updated == false {
		return nil, error.New(map[string]interface{}{
			"message": "InBound池中找不到对应的nodeConn，握手失败",
		})
	}

	// 应答一个握手包
	ncService.SendShake(nodeConn, nodeConn.GetSenderNodeID())

	return nil, nil
}
