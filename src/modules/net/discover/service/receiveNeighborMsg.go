package service

import (
	commonModels "github.com/cnf_core/src/modules/net/common/models"
	"github.com/cnf_core/src/utils/error"
	"github.com/cnf_core/src/utils/logger"
	"github.com/cnf_core/src/utils/router"
)

// HandleFindNode 处理邻居节点寻找的请求
func (discoverService *DiscoverService) HandleFindNode(data interface{}) (interface{}, *error.Error) {
	// 桶里没有数据就退出，不然会卡住这条协程
	if len(discoverService.myPrivateChanel["bucketNodeListChanel"]) == 0 {
		return nil, nil
	}

	// 和隔壁doConn共用一个变量，因为处于同一个消息队列中，所以不用担心并发问题
	discoverService.receiveMsgTempNodeListMsg = <-discoverService.myPrivateChanel["bucketNodeListChanel"]
	discoverService.receiveMsgTempNodeList = discoverService.receiveMsgTempNodeListMsg["nodeList"].([]*commonModels.Node)
	discoverService.receiveMsgTempNodeListWithDistance = discoverService.receiveMsgTempNodeListMsg["nodeListWithDistance"].([]map[string]interface{})

	// var targetNodeNeighbor []*commonModels.Node
	discoverService.receiveMsgTempTargetNodeNeighbor = nil
	discoverService.receiveMsgTempTargetNodeNeighbor = make([](*commonModels.Node), 0)

	// 1、先计算出本机节点和需要找的节点之间的距离
	if data.(map[string]interface{})["body"].(map[string]interface{})["msgJSON"] == nil || data.(map[string]interface{})["body"].(map[string]interface{})["msgJSON"].(map[string]interface{})["findingNodeID"] == nil {
		logger.Debug("findNode事件tcp数据包不完整")
		return nil, nil
	}
	discoverService.receiveMsgTempDistanceBetweenMeAndFindingNode = router.CalculateDetailDistance(discoverService.myNodeID,
		data.(map[string]interface{})["body"].(map[string]interface{})["msgJSON"].(map[string]interface{})["findingNodeID"].(string))

	// 2、然后找到这个结点的距离，在我本地的结点缓存列表中排在哪里，然后取出附近的3个结点
	discoverService.receiveMsgTempFoundPosition = -1
	for i := 0; i < len(discoverService.receiveMsgTempNodeListWithDistance); i++ {
		discoverService.receiveNeighborMsgSubNodeDistance = discoverService.receiveMsgTempNodeListWithDistance[i]["detailDistance"].([]int64) // hardCode
		for k := 0; k < len(discoverService.receiveNeighborMsgSubNodeDistance); k++ {
			if discoverService.receiveMsgTempDistanceBetweenMeAndFindingNode[k] == discoverService.receiveNeighborMsgSubNodeDistance[k] {
				continue
			}

			if discoverService.receiveMsgTempDistanceBetweenMeAndFindingNode[k] > discoverService.receiveNeighborMsgSubNodeDistance[k] {
				break
			}

			// 取i附近的几个邻居出来
			discoverService.receiveMsgTempFoundPosition = i
			break

		}

		// 如果这个结点就是队里最大的，就直接把foundPosition设置到最后一位
		if i == len(discoverService.receiveMsgTempNodeListWithDistance)-1 && discoverService.receiveMsgTempFoundPosition == -1 {
			discoverService.receiveMsgTempFoundPosition = i
			break
		}

		if discoverService.receiveMsgTempFoundPosition != -1 {
			break
		}
	}

	// 没有邻居
	if discoverService.receiveMsgTempFoundPosition == -1 {
		return nil, nil
	}

	// 一个一个加进来
	discoverService.receiveMsgTempTargetNodeNeighbor = append(discoverService.receiveMsgTempTargetNodeNeighbor, discoverService.receiveMsgTempNodeListWithDistance[discoverService.receiveMsgTempFoundPosition]["node"].(*commonModels.Node))
	if discoverService.receiveMsgTempFoundPosition+1 < len(discoverService.receiveMsgTempNodeListWithDistance) {
		discoverService.receiveMsgTempTargetNodeNeighbor = append(discoverService.receiveMsgTempTargetNodeNeighbor, discoverService.receiveMsgTempNodeListWithDistance[discoverService.receiveMsgTempFoundPosition+1]["node"].(*commonModels.Node))
	}
	// if discoverService.receiveMsgTempFoundPosition-1 >= 0 {
	// 	discoverService.receiveMsgTempTargetNodeNeighbor = append(discoverService.receiveMsgTempTargetNodeNeighbor, discoverService.receiveMsgTempNodeListWithDistance[discoverService.receiveMsgTempFoundPosition-1]["node"].(*commonModels.Node))
	// }
	// if discoverService.receiveMsgTempFoundPosition+2 < len(discoverService.receiveMsgTempNodeListWithDistance) {
	// 	discoverService.receiveMsgTempTargetNodeNeighbor = append(discoverService.receiveMsgTempTargetNodeNeighbor, discoverService.receiveMsgTempNodeListWithDistance[discoverService.receiveMsgTempFoundPosition+2]["node"].(*commonModels.Node))
	// }
	// if discoverService.receiveMsgTempFoundPosition-2 >= 0 {
	// 	discoverService.receiveMsgTempTargetNodeNeighbor = append(discoverService.receiveMsgTempTargetNodeNeighbor, discoverService.receiveMsgTempNodeListWithDistance[discoverService.receiveMsgTempFoundPosition-2]["node"].(*commonModels.Node))
	// }

	// logger.Debug(discoverService.conf.(map[string]interface{})["number"].(string) + " get data")
	// logger.Debug(data)
	discoverService.receiveMsgTempShareNeighborPackString = discoverService.GetshareNodeNeighborPackString(discoverService.receiveMsgTempTargetNodeNeighbor, data.(map[string]interface{})["body"].(map[string]interface{})["senderNodeID"].(string))
	// 说明构建失败
	if discoverService.receiveMsgTempShareNeighborPackString == "" {
		logger.Error("邻居分享失败")
		return nil, nil
	}

	discoverService.DoSend(discoverService.receiveMsgTempShareNeighborPackString, data.(map[string]interface{})["sourceIP"].(string), data.(map[string]interface{})["sourceServicePort"].(string))

	return nil, nil
}

// HandleShareNodeNeighbor 处理邻居节点分享
func (discoverService *DiscoverService) HandleShareNodeNeighbor(data interface{}) (interface{}, *error.Error) {
	if data.(map[string]interface{})["body"].(map[string]interface{})["msgJSON"] == nil || data.(map[string]interface{})["body"].(map[string]interface{})["msgJSON"].(map[string]interface{})["nodeNeighbor"] == nil {
		return nil, nil
	}

	if len(discoverService.myPrivateChanel["bucketOperateChanel"]) == cap(discoverService.myPrivateChanel["bucketOperateChanel"]) {
		return nil, nil
	}

	discoverService.receiveNeighborMsgNodeNeighbors = data.(map[string]interface{})["body"].(map[string]interface{})["msgJSON"].(map[string]interface{})["nodeNeighbor"].([]interface{})

	discoverService.receiveNeighborMsgSeed = make([]*commonModels.Node, 0)
	// 循环创建新的结点变量。这个结点要直接存到bucket里面的，不能搞全局
	for i := 0; i < len(discoverService.receiveNeighborMsgNodeNeighbors); i++ {
		// 不要添加自己到种子里面
		if discoverService.myNodeID == discoverService.receiveNeighborMsgNodeNeighbors[i].(map[string]interface{})["nodeID"].(string) {
			continue
		}

		// 正在doingPong的不要重复做
		_, discoverService.receiveNeighborMsgIsDoingPing = discoverService.pingPongCache[discoverService.receiveNeighborMsgNodeNeighbors[i].(map[string]interface{})["nodeID"].(string)]
		if discoverService.receiveNeighborMsgIsDoingPing == true {
			// logger.Debug("is doing ping")
			continue
		}

		node, _ := commonModels.CreateNode(map[string]interface{}{
			"ip":          discoverService.receiveNeighborMsgNodeNeighbors[i].(map[string]interface{})["ip"].(string),
			"nodeID":      discoverService.receiveNeighborMsgNodeNeighbors[i].(map[string]interface{})["nodeID"].(string),
			"servicePort": discoverService.receiveNeighborMsgNodeNeighbors[i].(map[string]interface{})["servicePort"].(string),
		})
		discoverService.receiveNeighborMsgSeed = append(discoverService.receiveNeighborMsgSeed, node)
	}

	// logger.Debug(ncService.conf.(map[string]interface{})["number"].(string) + " get seeds length : " + strconv.Itoa(len(discoverService.receiveNeighborMsgSeed)))

	if len(discoverService.receiveNeighborMsgSeed) == 0 {
		return nil, nil
	}

	discoverService.myPrivateChanel["bucketOperateChanel"] <- map[string]interface{}{
		"event": "addSeedByGroup",
		"seeds": discoverService.receiveNeighborMsgSeed,
	}

	return nil, nil
}
