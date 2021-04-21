package services

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	commonModels "github.com/cnf_core/src/modules/net/common/models"
	nodeConnectionModels "github.com/cnf_core/src/modules/net/nodeConnection/models"
	"github.com/cnf_core/src/utils/config"
	"github.com/cnf_core/src/utils/error"
	"github.com/cnf_core/src/utils/logger"
	"github.com/cnf_core/src/utils/router"
)

// ReceiveMsg 处理接收nodeConnetion消息，这里主要做分流
func (ncService *NodeConnectionService) ReceiveMsg(data interface{}) (interface{}, *error.Error) {
	// 收到握手请求时，因为很有可能接收方没有一个连接对象。
	if data.(map[string]interface{})["tcpData"].(map[string]interface{})["event"] == "shakeEvent" {
		// logger.Debug(config.ParseNodeID(ncService.conf) + " get shaked")
		return ncService.HandleShakeEvent(data)
	}

	if data.(map[string]interface{})["tcpData"].(map[string]interface{})["event"] == "shakeDestroyEvent" {
		// logger.Debug("getDestroy" + ncService.myNodeID)
		return ncService.HandleShakeDestroyEvent(data)
	}

	if data.(map[string]interface{})["tcpData"].(map[string]interface{})["event"] == "findNode" {
		return ncService.HandleFindNode(data)
	}

	if data.(map[string]interface{})["tcpData"].(map[string]interface{})["event"] == "shareNodeNeighbor" {
		return ncService.HandleShareNodeNeighbor(data)
	}

	return nil, nil
}

// HandleShareNodeNeighbor 处理邻居节点寻找的请求
func (ncService *NodeConnectionService) HandleShareNodeNeighbor(data interface{}) (interface{}, *error.Error) {
	if data.(map[string]interface{})["tcpData"].(map[string]interface{})["msgJSON"] == nil || data.(map[string]interface{})["tcpData"].(map[string]interface{})["msgJSON"].(map[string]interface{})["nodeNeighbor"] == nil {
		return nil, nil
	}

	if len(ncService.myPrivateChanel["bucketOperateChanel"]) == cap(ncService.myPrivateChanel["bucketOperateChanel"]) {
		return nil, nil
	}

	nodeNeighbors := data.(map[string]interface{})["tcpData"].(map[string]interface{})["msgJSON"].(map[string]interface{})["nodeNeighbor"].([]interface{})
	var seeds []*commonModels.Node = nil
	// 循环创建新的结点变量。这个结点要直接存到bucket里面的，不能搞全局
	for i := 0; i < len(nodeNeighbors); i++ {
		// 不要添加自己到种子里面
		if ncService.myNodeID == nodeNeighbors[i].(map[string]interface{})["nodeID"].(string) {
			continue
		}
		node, _ := commonModels.CreateNode(map[string]interface{}{
			"ip":          nodeNeighbors[i].(map[string]interface{})["ip"].(string),
			"nodeID":      nodeNeighbors[i].(map[string]interface{})["nodeID"].(string),
			"servicePort": nodeNeighbors[i].(map[string]interface{})["servicePort"].(string),
		})
		seeds = append(seeds, node)
	}

	// logger.Debug(ncService.conf.(map[string]interface{})["number"].(string) + " get seeds length : " + strconv.Itoa(len(seeds)))

	if len(seeds) == 0 {
		return nil, nil
	}

	ncService.myPrivateChanel["bucketOperateChanel"] <- map[string]interface{}{
		"event": "addSeedByGroup",
		"seeds": seeds,
	}

	return nil, nil
}

// HandleFindNode 处理邻居节点寻找的请求
func (ncService *NodeConnectionService) HandleFindNode(data interface{}) (interface{}, *error.Error) {
	// 桶里没有数据就退出，不然会卡住这条协程
	if len(ncService.myPrivateChanel["bucketNodeListChanel"]) == 0 {
		return nil, nil
	}

	// 和隔壁doConn共用一个变量，因为处于同一个消息队列中，所以不用担心并发问题
	ncService.doConnectTempNodeListMsg = <-ncService.myPrivateChanel["bucketNodeListChanel"]
	ncService.doConnectTempNodeList = ncService.doConnectTempNodeListMsg["nodeList"].([]*commonModels.Node)
	ncService.doConnectTempNodeListWithDistance = ncService.doConnectTempNodeListMsg["nodeListWithDistance"].([]map[string]interface{})

	// var targetNodeNeighbor []*commonModels.Node
	ncService.receiveMsgTempTargetNodeNeighbor = nil
	ncService.receiveMsgTempTargetNodeNeighbor = make([](*commonModels.Node), 0)

	// 1、先计算出本机节点和需要找的节点之间的距离
	if data.(map[string]interface{})["tcpData"].(map[string]interface{})["msgJSON"] == nil || data.(map[string]interface{})["tcpData"].(map[string]interface{})["msgJSON"].(map[string]interface{})["findingNodeID"] == nil {
		logger.Debug("findNode事件tcp数据包不完整")
		return nil, nil
	}
	ncService.receiveMsgTempDistanceBetweenMeAndFindingNode = router.CalculateDetailDistance(ncService.myNodeID,
		data.(map[string]interface{})["tcpData"].(map[string]interface{})["msgJSON"].(map[string]interface{})["findingNodeID"].(string))

	// 2、然后找到这个结点的距离，在我本地的结点缓存列表中排在哪里，然后取出附近的3个结点
	ncService.receiveMsgTempFoundPosition = -1
	for i := 0; i < len(ncService.doConnectTempNodeListWithDistance); i++ {
		subNodeDistance := ncService.doConnectTempNodeListWithDistance[i]["detailDistance"].([]int64) // hardCode
		for k := 0; k < len(subNodeDistance); k++ {
			if ncService.receiveMsgTempDistanceBetweenMeAndFindingNode[k] == subNodeDistance[k] {
				continue
			}

			if ncService.receiveMsgTempDistanceBetweenMeAndFindingNode[k] > subNodeDistance[k] {
				break
			}

			// 取i附近的几个邻居出来
			ncService.receiveMsgTempFoundPosition = i
			break

		}

		// 如果这个结点就是队里最大的，就直接把foundPosition设置到最后一位
		if i == len(ncService.doConnectTempNodeListWithDistance)-1 && ncService.receiveMsgTempFoundPosition == -1 {
			ncService.receiveMsgTempFoundPosition = i
			break
		}

		if ncService.receiveMsgTempFoundPosition != -1 {
			break
		}
	}

	// 没有邻居
	if ncService.receiveMsgTempFoundPosition == -1 {
		return nil, nil
	}

	// 一个一个加进来
	ncService.receiveMsgTempTargetNodeNeighbor = append(ncService.receiveMsgTempTargetNodeNeighbor, ncService.doConnectTempNodeListWithDistance[ncService.receiveMsgTempFoundPosition]["node"].(*commonModels.Node))
	if ncService.receiveMsgTempFoundPosition-1 >= 0 {
		ncService.receiveMsgTempTargetNodeNeighbor = append(ncService.receiveMsgTempTargetNodeNeighbor, ncService.doConnectTempNodeListWithDistance[ncService.receiveMsgTempFoundPosition-1]["node"].(*commonModels.Node))
	}
	if ncService.receiveMsgTempFoundPosition+1 < len(ncService.doConnectTempNodeListWithDistance) {
		ncService.receiveMsgTempTargetNodeNeighbor = append(ncService.receiveMsgTempTargetNodeNeighbor, ncService.doConnectTempNodeListWithDistance[ncService.receiveMsgTempFoundPosition+1]["node"].(*commonModels.Node))
	}
	// if ncService.receiveMsgTempFoundPosition-2 >= 0 {
	// 	ncService.receiveMsgTempTargetNodeNeighbor = append(ncService.receiveMsgTempTargetNodeNeighbor, ncService.doConnectTempNodeListWithDistance[ncService.receiveMsgTempFoundPosition-2]["node"].(*commonModels.Node))
	// }
	// if ncService.receiveMsgTempFoundPosition+2 < len(ncService.doConnectTempNodeListWithDistance) {
	// 	ncService.receiveMsgTempTargetNodeNeighbor = append(ncService.receiveMsgTempTargetNodeNeighbor, ncService.doConnectTempNodeListWithDistance[ncService.receiveMsgTempFoundPosition+2]["node"].(*commonModels.Node))
	// }

	// logger.Debug(ncService.conf.(map[string]interface{})["number"].(string) + " get data")
	// logger.Debug(data)
	ncService.receiveMsgTempShareNeighborPackString = ncService.GetshareNodeNeighborPackString(ncService.receiveMsgTempTargetNodeNeighbor, data.(map[string]interface{})["tcpData"].(map[string]interface{})["nodeID"].(string))
	// 说明构建失败
	if ncService.receiveMsgTempShareNeighborPackString == "" {
		logger.Error("邻居分享失败")
		return nil, nil
	}

	if data.(map[string]interface{})["nodeConn"] == nil {
		ncService.myPublicChanel["sendNodeConnectionMsgChanel"] <- map[string]interface{}{
			"nodeConn": nil,
			"message":  ncService.receiveMsgTempShareNeighborPackString,
		}
	} else {
		ncService.myPublicChanel["sendNodeConnectionMsgChanel"] <- map[string]interface{}{
			"nodeConn": data.(map[string]interface{})["nodeConn"].(*nodeConnectionModels.NodeConn),
			"message":  ncService.receiveMsgTempShareNeighborPackString,
		}
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
		addInboundErr := ncService.AddInBoundConn(&newNodeConn)
		if addInboundErr != nil {
			// inBound如果满了，也要告诉对方断开连接
			shakeDestroyString := ncService.GetShakePackString("shakeDestroyEvent", remoteNodeID)
			ncService.myPublicChanel["sendNodeConnectionMsgChanel"] <- map[string]interface{}{
				"nodeConn": &newNodeConn,
				"message":  shakeDestroyString,
			}
			return nil, nil
		}

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
