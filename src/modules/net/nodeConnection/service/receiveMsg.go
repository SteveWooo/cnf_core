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
	// nodeConn := data.(map[string]interface{})["nodeConn"].(*nodeConnectionModels.NodeConn)
	// logger.Debug(tcpData)
	// logger.Debug(data.(map[string]interface{})["nodeConn"])
	if tcpData.(map[string]interface{})["event"] == "shakeEvent" {
		return ncService.HandleShakeEvent(data)
	}

	if tcpData.(map[string]interface{})["event"] == "shakeBackEvent" {
		return ncService.HandleShakeBackEvent(data)
	}

	// 先检查这个conn是否已经shake过

	return nil, nil
}

// HandleShakeEvent 处理接收nodeConnetion消息
// 只有inBound连接，才有可能收到shakerEvent，所以收到shake就必然回复一个shakeBack
func (ncService *NodeConnectionService) HandleShakeEvent(data interface{}) (interface{}, *error.Error) {
	nodeConn := data.(map[string]interface{})["nodeConn"].(*nodeConnectionModels.NodeConn)
	nodeConnID := nodeConn.GetNodeConnID()

	if ncService.CheckShakeEvent(data) == false {
		return nil, nil
	}

	updated := false
	// 更新连接池的内容
	for i := 0; i < len(ncService.inBoundConn); i++ {
		if ncService.inBoundConn[i] == nil {
			continue
		}

		if ncService.inBoundConn[i].GetNodeConnID() == nodeConnID {
			// 还没握手，应答一个握手包。这个握手包必然是shakeBack的
			ncService.inBoundConn[i].SetShaker(data)

			ncService.SendShake(nodeConn, nodeConn.GetNodeID())
			updated = true
		}
	}

	// Todo 找不到nodeConn，而且是salve服务器，那就创建一个NodeConn
	if updated == false {
		return nil, error.New(map[string]interface{}{
			"message": "InBound池中找不到对应的nodeConn，握手失败",
		})
	}

	return nil, nil
}

// CheckShakeEvent 检查握手包，返回false则不在进行下去
func (ncService *NodeConnectionService) CheckShakeEvent(data interface{}) bool {
	// 如果已经在连接池内，就不要再接收这个握手了，直接删除掉
	tcpData := data.(map[string]interface{})["tcpData"]
	nodeID := tcpData.(map[string]interface{})["nodeID"].(string)
	nodeConn := data.(map[string]interface{})["nodeConn"].(*nodeConnectionModels.NodeConn)

	if ncService.IsBucketExistShakedNode(nodeID) == true {
		nodeConn.Socket.Close()
		if nodeConn.GetConnType() == "inBound" {
			ncService.DeleteInBoundConn(nodeConn)
		}

		if nodeConn.GetConnType() == "outBound" {
			ncService.DeleteOutBoundConn(nodeConn)
		}

		return false
	}

	return true
}

// HandleShakeBackEvent 只有outBound的连接才会给你发shakeBack
func (ncService *NodeConnectionService) HandleShakeBackEvent(data interface{}) (interface{}, *error.Error) {
	nodeConn := data.(map[string]interface{})["nodeConn"].(*nodeConnectionModels.NodeConn)
	nodeConnID := nodeConn.GetNodeConnID()

	if ncService.CheckShakeBackEvent(data) == false {
		return nil, nil
	}

	updated := false
	// 更新连接池的内容
	for i := 0; i < len(ncService.outBoundConn); i++ {
		if ncService.outBoundConn[i] == nil {
			continue
		}

		if ncService.outBoundConn[i].GetNodeConnID() == nodeConnID {
			ncService.outBoundConn[i].SetShaker(data)
			updated = true
		}
	}

	// Todo 找不到nodeConn，而且是salve服务器，那就创建一个NodeConn
	if updated == false {
		return nil, error.New(map[string]interface{}{
			"message": "InBound池中找不到对应的nodeConn，握手失败",
		})
	}

	return nil, nil
}

// CheckShakeBackEvent 检查握手包，返回false则不在进行下去
func (ncService *NodeConnectionService) CheckShakeBackEvent(data interface{}) bool {
	// 如果已经在连接池内，就不要再接收这个握手了，直接删除掉
	tcpData := data.(map[string]interface{})["tcpData"]
	nodeID := tcpData.(map[string]interface{})["nodeID"].(string)
	nodeConn := data.(map[string]interface{})["nodeConn"].(*nodeConnectionModels.NodeConn)

	// logger.Debug(tcpData)

	if ncService.IsBucketExistShakedNode(nodeID) == true {
		nodeConn.Socket.Close()
		if nodeConn.GetConnType() == "inBound" {
			ncService.DeleteInBoundConn(nodeConn)
		}

		if nodeConn.GetConnType() == "outBound" {
			ncService.DeleteOutBoundConn(nodeConn)
		}

		return false
	}

	return true
}
