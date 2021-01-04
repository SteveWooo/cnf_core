package services

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	nodeConnectionModels "github.com/cnf_core/src/modules/net/nodeConnection/models"
	"github.com/cnf_core/src/utils/config"
	"github.com/cnf_core/src/utils/error"
)

// ReceiveMsg 处理接收nodeConnetion消息，这里主要做分流
func (ncService *NodeConnectionService) ReceiveMsg(data interface{}) (interface{}, *error.Error) {
	tcpData := data.(map[string]interface{})["tcpData"]
	// logger.Debug(tcpData)

	// 收到握手请求时，因为很有可能接收方没有一个连接对象。
	if tcpData.(map[string]interface{})["event"] == "shakeEvent" {
		// logger.Debug(config.ParseNodeID(ncService.conf) + " get shaked")
		return ncService.HandleShakeEvent(data)
	}

	if tcpData.(map[string]interface{})["event"] == "shakeDestroyEvent" {
		// logger.Debug("getDestroy")
		return ncService.HandleShakeDestroyEvent(data)
	}

	return nil, nil
}

// HandleShakeEvent 处理接收nodeConnetion消息
// 只有inBound连接，才有可能收到shakerEvent，所以收到shake就必然回复一个shakeBack
func (ncService *NodeConnectionService) HandleShakeEvent(data interface{}) (interface{}, *error.Error) {
	// 这是主节点的conn。主节点的conn收到这个子节点的shake请求
	var nodeConn *nodeConnectionModels.NodeConn = nil
	if data.(map[string]interface{})["nodeConn"] != nil {
		nodeConn = data.(map[string]interface{})["nodeConn"].(*nodeConnectionModels.NodeConn)
	}
	tcpData := data.(map[string]interface{})["tcpData"]
	remoteNodeID := tcpData.(map[string]interface{})["nodeID"].(string)

	// 首先看看我有没有inBound和outBound有没有连接了你
	alreadyShake := false
	if alreadyShake == false {
		for i := 0; i < len(ncService.inBoundConn); i++ {
			if ncService.inBoundConn[i] == nil {
				continue
			}
			if ncService.inBoundConn[i].GetNodeID() == remoteNodeID {
				if ncService.inBoundConn[i].IsShaked() == true {
					alreadyShake = true
					break
				}
				continue
			}
		}
	}
	if alreadyShake == false {
		for i := 0; i < len(ncService.outBoundConn); i++ {
			if ncService.outBoundConn[i] == nil {
				continue
			}
			if ncService.outBoundConn[i].GetNodeID() == remoteNodeID {
				if ncService.outBoundConn[i].IsShaked() == true {
					alreadyShake = true
					break
				}
				continue
			}
		}
	}

	// 因为我在tryOutBound中就已经建立了shake，而你比我晚到，所以我要无视你的shake请求，你就是我的inbound。
	if alreadyShake == true {
		shakeDestroyString := ncService.GetShakePackString("shakeDestroyEvent", remoteNodeID)
		ncService.myPublicChanel["sendNodeConnectionMsgChanel"] <- map[string]interface{}{
			"nodeConn": nodeConn,
			"message":  shakeDestroyString,
		}

		return nil, nil
	}

	cnfNet := ncService.conf.(map[string]interface{})["net"]
	// 如果我inBound和outbound之中，有或者无连接了你，但还没shake成功，那我就shakeBack，然后把你加入到我的inBound桶中，同时setShake
	if alreadyShake == false {
		// 创建一个inBound nodeConn，复制主节点conn的内容过来

		var newNodeConn nodeConnectionModels.NodeConn
		if nodeConn == nil {
			newNodeConn.Build(nil, "inBound")
			newNodeConn.SetRemoteAddr(cnfNet.(map[string]interface{})["ip"].(string) + ":" + cnfNet.(map[string]interface{})["servicePort"].(string))
		} else {
			newNodeConn.Build(nodeConn.GetSocket(), "inBound")
			newNodeConn.SetRemoteAddr((*nodeConn.GetSocket()).RemoteAddr().String())
		}

		newNodeConn.SetNodeID(remoteNodeID)
		newNodeConn.SetTargetNodeID(config.ParseNodeID(ncService.conf))

		// 然后把这个新的nodeConn加入到自己的inBound里面
		ncService.AddInBoundConn(&newNodeConn)

		// 单点shaker。收到shake就直接set，而另一方发起doTryOutbound的时候，就已经当作shaked了
		newNodeConn.SetShaker(data)
	}
	return nil, nil
}

// HandleShakeBackEvent 只有outBound的连接才会给你发shakeBack
func (ncService *NodeConnectionService) HandleShakeBackEvent(data interface{}) (interface{}, *error.Error) {
	nodeConn := data.(map[string]interface{})["nodeConn"].(*nodeConnectionModels.NodeConn)
	tcpData := data.(map[string]interface{})["tcpData"]
	remoteNodeID := tcpData.(map[string]interface{})["nodeID"].(string)

	// 首先看看我有没有inBound和outBound有没有连接了你
	alreadyShake := false
	if alreadyShake == false {
		for i := 0; i < len(ncService.inBoundConn); i++ {
			if ncService.inBoundConn[i] == nil {
				continue
			}
			if ncService.inBoundConn[i].GetNodeID() == remoteNodeID {
				if ncService.inBoundConn[i].IsShaked() == true {
					alreadyShake = true
					break
				}
				continue
			}
		}
	}
	if alreadyShake == false {
		for i := 0; i < len(ncService.outBoundConn); i++ {
			if ncService.outBoundConn[i] == nil {
				continue
			}
			if ncService.outBoundConn[i].GetNodeID() == remoteNodeID {
				if ncService.outBoundConn[i].IsShaked() == true {
					alreadyShake = true
					break
				}
				continue
			}
		}
	}

	// 如果我inBound和outBound之中，有连接了你，也已经shake成功了，那我就不回复Shake，甚至不需要搭理这个shakeBack事件
	if alreadyShake == true {
		// ncService.DeleteUnShakedNodeConn(remoteNodeID)
	}

	// 如果我都没和你建立shake，那么我就要在outBound中，找到你，然后setShake。代表建立了shaked
	if alreadyShake == false {
		for i := 0; i < len(ncService.outBoundConn); i++ {
			if ncService.outBoundConn[i] == nil {
				continue
			}
			if ncService.outBoundConn[i].GetNodeID() == remoteNodeID {

				shakeBackAgainPackageString := ncService.GetShakePackString("shakeBackAgainEvent", remoteNodeID)
				ncService.myPublicChanel["sendNodeConnectionMsgChanel"] <- map[string]interface{}{
					"nodeConn": nodeConn,
					"message":  shakeBackAgainPackageString,
				}

				ncService.outBoundConn[i].SetShaker(data)
				return nil, nil
			}
		}
	}

	return nil, error.New(map[string]interface{}{
		"message": "outBound中找不到remoteNodeID",
	})
}

// HandleShakeBackAgainEvent 只有inBound的连接才会给你发shakeBackAgain
func (ncService *NodeConnectionService) HandleShakeBackAgainEvent(data interface{}) (interface{}, *error.Error) {
	tcpData := data.(map[string]interface{})["tcpData"]
	remoteNodeID := tcpData.(map[string]interface{})["nodeID"].(string)

	// 首先看看我有没有inBound和outBound有没有连接了你
	alreadyShake := false
	if alreadyShake == false {
		for i := 0; i < len(ncService.inBoundConn); i++ {
			if ncService.inBoundConn[i] == nil {
				continue
			}
			if ncService.inBoundConn[i].GetNodeID() == remoteNodeID {
				if ncService.inBoundConn[i].IsShaked() == true {
					alreadyShake = true
					break
				}
				continue
			}
		}
	}
	if alreadyShake == false {
		for i := 0; i < len(ncService.outBoundConn); i++ {
			if ncService.outBoundConn[i] == nil {
				continue
			}
			if ncService.outBoundConn[i].GetNodeID() == remoteNodeID {
				if ncService.outBoundConn[i].IsShaked() == true {
					alreadyShake = true
					break
				}
				continue
			}
		}
	}

	// 如果我inBound和outBound之中，有连接了你，也已经shake成功了，那我就不回复Shake，甚至不需要搭理这个shakeBack事件
	if alreadyShake == true {
		// ncService.DeleteUnShakedNodeConn(remoteNodeID)
	}

	// 如果我都没和你建立shake，那么我就要在outBound中，找到你，然后setShake。代表建立了shaked
	if alreadyShake == false {
		for i := 0; i < len(ncService.inBoundConn); i++ {
			if ncService.inBoundConn[i] == nil {
				continue
			}
			if ncService.inBoundConn[i].GetNodeID() == remoteNodeID {
				ncService.inBoundConn[i].SetShaker(data)
				break
			}
		}
	}

	return nil, nil
}

// DeleteUnShakedNodeConn 由于这个remoteNodeID已经握手成功，所以要把同ID，但未握手成功的节点清理掉
func (ncService *NodeConnectionService) DeleteUnShakedNodeConn(remoteNodeID string) {
	for i := 0; i < len(ncService.inBoundConn); i++ {
		if ncService.inBoundConn[i] == nil {
			continue
		}
		if ncService.inBoundConn[i].GetNodeID() == remoteNodeID {
			if ncService.inBoundConn[i].IsShaked() == false {
				ncService.inBoundConn[i] = nil
			}
		}
	}

	for i := 0; i < len(ncService.outBoundConn); i++ {
		if ncService.outBoundConn[i] == nil {
			continue
		}
		if ncService.outBoundConn[i].GetNodeID() == remoteNodeID {
			if ncService.outBoundConn[i].IsShaked() == false {
				ncService.outBoundConn[i] = nil
			}
		}
	}
}

// HandleShakeDestroyEvent 对方节点要求断开主动连接的事件请求
func (ncService *NodeConnectionService) HandleShakeDestroyEvent(data interface{}) (interface{}, *error.Error) {
	// 这是主节点的conn。主节点的conn收到这个子节点的shake请求
	tcpData := data.(map[string]interface{})["tcpData"]
	remoteNodeID := tcpData.(map[string]interface{})["nodeID"].(string)

	for i := 0; i < len(ncService.outBoundConn); i++ {
		if ncService.outBoundConn[i] == nil {
			continue
		}
		if ncService.outBoundConn[i].GetNodeID() == remoteNodeID {
			ncService.outBoundConn[i].SetDestroy()
			// ncService.outBoundConn[i] = nil
			return nil, nil
		}
	}

	return nil, nil
}

// ParseTCPData 解析TCP数据包
func (ncService *NodeConnectionService) ParseTCPData(tcpSourceData string) (interface{}, *error.Error) {
	// 取出content部分直接处理
	tcpSourceData = tcpSourceData[strings.Index(tcpSourceData, "content:")+len("content:"):]
	contentBase64 := tcpSourceData[:]

	contentByte, decodeErr := base64.StdEncoding.DecodeString(contentBase64)
	if decodeErr != nil {
		return nil, error.New(map[string]interface{}{
			"message": "tcp数据包解析失败: enCodeBase64",
		})
	}
	content := string(contentByte)
	var contentJSON interface{}
	jSONDecodeErr := json.Unmarshal([]byte(content), &contentJSON)
	if jSONDecodeErr != nil {
		return nil, error.New(map[string]interface{}{
			"message": "tcp数据包解析失败:content json unMarshal",
		})
	}

	var msgJSON interface{}
	// logger.Debug(contentJSON)
	contentJSONMsg := contentJSON.(map[string]interface{})["msg"].(string)
	msgJSONDecodeErr := json.Unmarshal([]byte(contentJSONMsg), &msgJSON)
	if msgJSONDecodeErr != nil {
		return nil, error.New(map[string]interface{}{
			"message": "tcp数据包解析失败:message content json umMarshal",
		})
	}

	contentJSON.(map[string]interface{})["msgJSON"] = msgJSON
	contentJSON.(map[string]interface{})["nodeID"] = msgJSON.(map[string]interface{})["from"].(map[string]interface{})["nodeID"]

	return contentJSON, nil
}
