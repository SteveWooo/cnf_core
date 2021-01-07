package services

import (
	"math/rand"

	commonModels "github.com/cnf_core/src/modules/net/common/models"
	"github.com/cnf_core/src/utils/config"
	"github.com/cnf_core/src/utils/router"
	"github.com/cnf_core/src/utils/timer"
	// commonModels "github.com/cnf_core/src/modules/net/common/models"
)

// RunFindConnectionByMasterArea 主动找可用节点（使用master结点算法的查找）
func (ncService *NodeConnectionService) RunFindConnectionByMasterArea(chanels map[string]chan map[string]interface{}) {
	DoFindConnectionByNodeListEventMap := make(map[string]interface{})
	DoFindConnectionByNodeListEventMap["event"] = "doFindConnectionByNodeList"

	DoFindNeighborEventMap := make(map[string]interface{})
	DoFindNeighborEventMap["event"] = "doFindNeighborByMasterArea"

	DoMonitorEventMap := make(map[string]interface{})
	DoMonitorEventMap["event"] = "doMonitor"

	// 连接发起频率控制
	doFindConnLoop := 0

	// 邻居寻找频率控制
	doFindNeighborLoop := 0
	doFindNeighborTarget := ""     // 当前需要找的邻居
	doFindNeighborMasterIndex := 0 // 区域Master结点需要用上的参数，代表这一轮找到第几个邻居
	for {
		timer.Sleep(100 + rand.Intn(100))

		if doFindConnLoop == 0 {
			ncService.myPrivateChanel["nodeConnectionEventChanel"] <- DoFindConnectionByNodeListEventMap
			ncService.myPrivateChanel["nodeConnectionEventChanel"] <- DoMonitorEventMap
		}
		doFindConnLoop++
		if doFindConnLoop == 5 {
			doFindConnLoop = 0
		}

		if doFindNeighborLoop == 0 {
			// 区域master结点需要循环找别的区域的master连接
			if ncService.masterAreaIsMaster == true {
				doFindNeighborTarget = router.MasterAreas[doFindNeighborMasterIndex]
				doFindNeighborMasterIndex++
				// 循环完了，要重置一遍
				if doFindNeighborMasterIndex >= len(router.MasterAreas) {
					doFindNeighborMasterIndex = 0
				}
			}

			// 如果不是区域Master结点，就要尽量找本区域的Master结点进行连接
			if ncService.masterAreaIsMaster == false {
				doFindNeighborTarget = router.MasterAreas[ncService.masterAreaLocateArea]
			}

			DoFindNeighborEventMap["findingNodeID"] = doFindNeighborTarget
			// ncService.myPrivateChanel["nodeConnectionEventChanel"] <- DoFindNeighborEventMap // 改为在udp发现协议中传播邻居消息
		}
		doFindNeighborLoop++
		if doFindNeighborLoop == 20 {
			doFindNeighborLoop = 0
		}
	}
}

// DoFindConnectionByNodeListAndMasterArea 主动找可用节点
// Master算法中，区域master结点优先与其他区域的Master结点发起连接，没有的话就与NodeList第一位发起连接。
// 而普通结点优先与本区域的Master结点发起连接，没有Master就与区域的其他结点发起连接，本区域其他结点都没有了，就与NodeList第一位发起连接
func (ncService *NodeConnectionService) DoFindConnectionByNodeListAndMasterArea(chanels map[string]chan map[string]interface{}) {
	// 如果outbound无空位，则不需要进行尝试连接
	if ncService.IsOutBoundFull() == true {
		return
	}
	// 队列为空就要退出，防止卡死协程
	if len(chanels["bucketNodeListChanel"]) == 0 {
		return
	}
	ncService.doConnectTempNodeListMsg = <-chanels["bucketNodeListChanel"]
	ncService.doConnectTempNodeList = ncService.doConnectTempNodeListMsg["nodeList"].([]*commonModels.Node)
	ncService.doConnectTempNewNode = nil

	// 区域Master和区域Normal结点，遍历这个NodeList3次后，得出一个优先级连接列表，再进入下一步找节点连接环节
	ncService.doConnectTempnodeListQueue = make([]*commonModels.Node, 0)
	// Master的优先级：
	// 1 其他区域的Master
	// 2 本区域的Master
	// 3 本区域的normal
	// 4 其他区域的normal
	if ncService.masterAreaIsMaster == true {
		// 1 找其他区域的master出来
		for i := 0; i < len(ncService.doConnectTempNodeList); i++ {
			if ncService.doConnectTempNodeList[i].IsAreaMaster() == true && ncService.doConnectTempNodeList[i].GetMasterAreaLocation() != ncService.masterAreaLocateArea {
				ncService.doConnectTempnodeListQueue = append(ncService.doConnectTempnodeListQueue, ncService.doConnectTempNodeList[i])
			}
		}

		// 2 找本区域的Master出来
		for i := 0; i < len(ncService.doConnectTempNodeList); i++ {
			if ncService.doConnectTempNodeList[i].IsAreaMaster() == true && ncService.doConnectTempNodeList[i].GetMasterAreaLocation() == ncService.masterAreaLocateArea {
				ncService.doConnectTempnodeListQueue = append(ncService.doConnectTempnodeListQueue, ncService.doConnectTempNodeList[i])
			}
		}

		// 3 找本区域的Normal出来
		for i := 0; i < len(ncService.doConnectTempNodeList); i++ {
			if ncService.doConnectTempNodeList[i].IsAreaMaster() == false && ncService.doConnectTempNodeList[i].GetMasterAreaLocation() == ncService.masterAreaLocateArea {
				ncService.doConnectTempnodeListQueue = append(ncService.doConnectTempnodeListQueue, ncService.doConnectTempNodeList[i])
			}
		}

		// 4 找其他区域的Normal出来
		for i := 0; i < len(ncService.doConnectTempNodeList); i++ {
			if ncService.doConnectTempNodeList[i].IsAreaMaster() == false && ncService.doConnectTempNodeList[i].GetMasterAreaLocation() != ncService.masterAreaLocateArea {
				ncService.doConnectTempnodeListQueue = append(ncService.doConnectTempnodeListQueue, ncService.doConnectTempNodeList[i])
			}
		}
	}

	// Normal的优先级
	// 1 本区域的Master
	// 2 本区域的normal
	// 3 其他区域的Master
	// 4 其他区域的normal
	if ncService.masterAreaIsMaster == false {
		// 1 找本区域的Master出来
		for i := 0; i < len(ncService.doConnectTempNodeList); i++ {
			if ncService.doConnectTempNodeList[i].IsAreaMaster() == true && ncService.doConnectTempNodeList[i].GetMasterAreaLocation() == ncService.masterAreaLocateArea {
				ncService.doConnectTempnodeListQueue = append(ncService.doConnectTempnodeListQueue, ncService.doConnectTempNodeList[i])
			}
		}

		// 2 找本区域的Normal出来
		for i := 0; i < len(ncService.doConnectTempNodeList); i++ {
			if ncService.doConnectTempNodeList[i].IsAreaMaster() == false && ncService.doConnectTempNodeList[i].GetMasterAreaLocation() == ncService.masterAreaLocateArea {
				ncService.doConnectTempnodeListQueue = append(ncService.doConnectTempnodeListQueue, ncService.doConnectTempNodeList[i])
			}
		}

		// 3 找其他区域的master出来
		for i := 0; i < len(ncService.doConnectTempNodeList); i++ {
			if ncService.doConnectTempNodeList[i].IsAreaMaster() == true && ncService.doConnectTempNodeList[i].GetMasterAreaLocation() != ncService.masterAreaLocateArea {
				ncService.doConnectTempnodeListQueue = append(ncService.doConnectTempnodeListQueue, ncService.doConnectTempNodeList[i])
			}
		}

		// 4 找其他区域的Normal出来
		for i := 0; i < len(ncService.doConnectTempNodeList); i++ {
			if ncService.doConnectTempNodeList[i].IsAreaMaster() == false && ncService.doConnectTempNodeList[i].GetMasterAreaLocation() != ncService.masterAreaLocateArea {
				ncService.doConnectTempnodeListQueue = append(ncService.doConnectTempnodeListQueue, ncService.doConnectTempNodeList[i])
			}
		}
	}

	// 找到和自己距离最近的节点进行连接，NodeList已经排好了顺序
	for i := 0; i < len(ncService.doConnectTempnodeListQueue); i++ {
		// 不要和自己连接
		if ncService.doConnectTempnodeListQueue[i].GetNodeID() == config.ParseNodeID(ncService.conf) {
			continue
		}
		if ncService.IsBucketExistUnShakedNode(ncService.doConnectTempnodeListQueue[i].GetNodeID()) || ncService.IsBucketExistShakedNode(ncService.doConnectTempnodeListQueue[i].GetNodeID()) {
			continue
		}
		ncService.doConnectTempNewNode = ncService.doConnectTempnodeListQueue[i]
	}

	// 无可用节点连接
	if ncService.doConnectTempNewNode == nil {
		return
	}

	// 端口多路复用逻辑 Master节点会调用以下函数 MasterDoTryOutBoundConnect
	ncService.myPublicChanel["submitNodeConnectionCreateChanel"] <- map[string]interface{}{
		"newNode":         ncService.doConnectTempNewNode, // 包含了ip:port:nodeID
		"targetNodeID":    config.ParseNodeID(ncService.conf),
		"shakePackString": ncService.GetShakePackString("outBound", ncService.doConnectTempNewNode.GetNodeID()),
	}
}
