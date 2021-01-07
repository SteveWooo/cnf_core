package net

import (
	nodeConnectionModels "github.com/cnf_core/src/modules/net/nodeConnection/models"
	"github.com/cnf_core/src/utils/config"
	"github.com/cnf_core/src/utils/error"
	logger "github.com/cnf_core/src/utils/logger"
)

// RunMessageQueue 处理各个消息队列
func (cnfNet *CnfNet) RunMessageQueue(chanels map[string]chan map[string]interface{}) {
	// discover 接收消息的队列
	go cnfNet.HandleDiscoverMsgReceive(chanels)
	go cnfNet.HandleSubNodeDiscoverMsgReceive(chanels)

	// discover 消息发送队列处理
	go cnfNet.HandleDiscoverMsgSend(chanels)

	// nodeConn 消息接收队列
	go cnfNet.HandleNodeConnectionMsgReceive(chanels)
	go cnfNet.HandleSubNodeConnectionMsgReceive(chanels)

	// nodeConn 创建处理
	go cnfNet.HandleSubNodeSubmitNodeConnectionCreate(chanels)
	// nodeConn 创建成功处理
	go cnfNet.HandleSubNodeReceiveNodeConnectionCreate(chanels)

	// nodeConn 子节点消息发送队列处理
	go cnfNet.HandleSubNodeSendNodeConnectionMsg(chanels)

	// logger.Info("消息队列启动完成")
}

// HandleDiscoverMsgReceive 关于发现服务的消息处理
func (cnfNet *CnfNet) HandleDiscoverMsgReceive(chanels map[string]chan map[string]interface{}) {
	confNet := cnfNet.conf.(map[string]interface{})["net"]
	if confNet.(map[string]interface{})["masterServer"] == "true" {
		var udpData map[string]interface{}
		var targetChanel interface{}
		var exist bool
		for {
			udpData = <-chanels["receiveDiscoverMsgChanel"] // 发现服务的udp socket如果收不到消息，就会卡死这条协程。
			targetChanel, exist = cnfNet.publicChanels[udpData["targetNodeID"].(string)]
			// logger.Debug(udpData)
			// logger.Debug(udpData["targetNodeID"].(string) + " receive")
			if exist {
				targetChanel.(map[string]chan map[string]interface{})["receiveDiscoverMsgChanel"] <- udpData
			} else {
				// logger.Error("接收到未知Udp数据包，本服务器无该节点 - 1:" + udpData["targetNodeID"].(string))
			}
		}
	}
}

// HandleSubNodeDiscoverMsgReceive 公共节点接收master节点的消息推送服务
func (cnfNet *CnfNet) HandleSubNodeDiscoverMsgReceive(chanels map[string]chan map[string]interface{}) {
	myNodeID := config.ParseNodeID(cnfNet.conf)
	var udpData map[string]interface{}
	for {
		udpData = <-cnfNet.publicChanels[myNodeID].(map[string]chan map[string]interface{})["receiveDiscoverMsgChanel"]
		if udpData != nil {

		}
		// logger.Debug(udpData)
		cnfNet.myPrivateChanel["discoverEventChanel"] <- map[string]interface{}{
			"event":   "receiveMsg",
			"udpData": udpData,
		}
	}
}

// HandleDiscoverMsgSend Master节点处理子节点的消息发送问题
func (cnfNet *CnfNet) HandleDiscoverMsgSend(chanels map[string]chan map[string]interface{}) {
	confNet := cnfNet.conf.(map[string]interface{})["net"]

	// 主节点的处理
	if confNet.(map[string]interface{})["masterServer"] == "true" {
		for nodeID, chanel := range cnfNet.publicChanels {
			// 匿名函数，循环获取子节点的消息发送队列请求，然后替子节点发送消息
			getMessageFromSubChanel := func(nid string, c map[string]chan map[string]interface{}) {
				var sendingData map[string]interface{}
				for {
					sendingData = <-c["sendDiscoverMsgChanel"]
					cnfNet.discover.SendMsg(sendingData["message"].(string), sendingData["targetIP"].(string), sendingData["targetServicePort"].(string))
				}
			}
			go getMessageFromSubChanel(nodeID, chanel.(map[string]chan map[string]interface{}))
		}
	}
}

// HandleSubNodeSubmitNodeConnectionCreate 主节点监听，处理子节点提交创建连接对象
func (cnfNet *CnfNet) HandleSubNodeSubmitNodeConnectionCreate(chanels map[string]chan map[string]interface{}) {
	confNet := cnfNet.conf.(map[string]interface{})["net"]
	if confNet.(map[string]interface{})["masterServer"] == "true" {
		for nodeID, subNodeChanel := range cnfNet.publicChanels {
			// 匿名函数，循环获取子节点的创建连接请求
			getNodeConnCreateReq := func(nid string, c map[string]chan map[string]interface{}) {
				var nodeConnCreateReq map[string]interface{}
				var nodeConnCreateMsg interface{}
				var nodeConnCreateErr *error.Error
				for {
					nodeConnCreateReq = <-c["submitNodeConnectionCreateChanel"]
					nodeConnCreateMsg, nodeConnCreateErr = cnfNet.nodeConnection.MasterDoTryOutBoundConnect(nodeConnCreateReq)
					if nodeConnCreateErr != nil {
						logger.Error("连接创建失败")
						continue
					}
					// logger.Debug(nodeConn)
					// 再传给子节点的连接创立成功通道，也就是 HandleSubNodeReceiveNodeConnectionCreate 函数
					c["receiveNodeConnectionCreateChanel"] <- nodeConnCreateMsg.(map[string]interface{})
				}
			}
			go getNodeConnCreateReq(nodeID, subNodeChanel.(map[string]chan map[string]interface{}))
		}
	}
}

// HandleSubNodeReceiveNodeConnectionCreate 子节点成功获取新创建的nodeConn入口
func (cnfNet *CnfNet) HandleSubNodeReceiveNodeConnectionCreate(chanels map[string]chan map[string]interface{}) {
	myNodeID := config.ParseNodeID(cnfNet.conf)
	var nodeConnectionCreateResp map[string]interface{}
	for {
		nodeConnectionCreateResp = <-cnfNet.publicChanels[myNodeID].(map[string]chan map[string]interface{})["receiveNodeConnectionCreateChanel"]
		// 子节点收到创建nodeConn成功消息后，走子节点的创建成功逻辑即可。
		if nodeConnectionCreateResp["connType"] == "outBound" {
			go cnfNet.nodeConnection.SalveHandleNodeOutBoundConnectionCreateEvent(nodeConnectionCreateResp)
		}

		if nodeConnectionCreateResp["connType"] == "inBound" {
			go cnfNet.nodeConnection.SalveHandleNodeInBoundConnectionCreateEvent(nodeConnectionCreateResp)
		}
	}
}

// HandleSubNodeSendNodeConnectionMsg 主节点监听，处理子节点提交nodeConnection消息发送功能
func (cnfNet *CnfNet) HandleSubNodeSendNodeConnectionMsg(chanels map[string]chan map[string]interface{}) {
	confNet := cnfNet.conf.(map[string]interface{})["net"]
	if confNet.(map[string]interface{})["masterServer"] == "true" {
		for nodeID, subNodeChanel := range cnfNet.publicChanels {
			// 匿名函数，循环获取子节点的创建连接请求
			getSubNodeSendNodeConnectionMsgReq := func(nid string, c map[string]chan map[string]interface{}) {
				var sendNodeConnectionMsgReq map[string]interface{}
				var tcpMessage string
				var tcpData interface{}
				var tempNodeConn *nodeConnectionModels.NodeConn
				for {
					sendNodeConnectionMsgReq = <-c["sendNodeConnectionMsgChanel"]
					tcpMessage = sendNodeConnectionMsgReq["message"].(string)
					// 本地节点消息，直接发送即可
					if sendNodeConnectionMsgReq["nodeConn"] == nil || sendNodeConnectionMsgReq["nodeConn"].(*nodeConnectionModels.NodeConn).GetSocket() == nil {
						// 直接转发给本地节点，不需要走socket
						tcpData, _ = cnfNet.nodeConnection.ParseTCPData(tcpMessage)
						chanels["receiveNodeConnectionMsgChanel"] <- map[string]interface{}{
							"nodeConn":     nil,
							"tcpData":      tcpData,
							"localMessage": true,
						}

						continue
					}

					// 非本地节点，使用socket发送数据报文
					tempNodeConn = sendNodeConnectionMsgReq["nodeConn"].(*nodeConnectionModels.NodeConn)
					if tempNodeConn.GetSocket() != nil {
						cnfNet.nodeConnection.SendMsg(tempNodeConn, tcpMessage)
					}
				}
			}
			go getSubNodeSendNodeConnectionMsgReq(nodeID, subNodeChanel.(map[string]chan map[string]interface{}))
		}
	}
}

// HandleNodeConnectionMsgReceive 处理TCP数据接收
func (cnfNet *CnfNet) HandleNodeConnectionMsgReceive(chanels map[string]chan map[string]interface{}) {
	confNet := cnfNet.conf.(map[string]interface{})["net"]
	if confNet.(map[string]interface{})["masterServer"] == "true" {
		var connectionMsg map[string]interface{}
		var tcpData interface{}
		var localChanel interface{}
		var isChanelExist bool
		for {
			connectionMsg = <-chanels["receiveNodeConnectionMsgChanel"]
			tcpData = connectionMsg["tcpData"]

			localChanel, isChanelExist = cnfNet.publicChanels[tcpData.(map[string]interface{})["targetNodeID"].(string)]
			if isChanelExist {
				localChanel.(map[string]chan map[string]interface{})["receiveNodeConnectionMsgChanel"] <- connectionMsg
			} else {
				logger.Error("接收到未知NodeID tcp数据包，本服务器无该节点")
			}
		}
	}
}

// HandleSubNodeConnectionMsgReceive 公共节点接收master节点的消息推送服务
func (cnfNet *CnfNet) HandleSubNodeConnectionMsgReceive(chanels map[string]chan map[string]interface{}) {
	myNodeID := config.ParseNodeID(cnfNet.conf)
	var connectionMsg map[string]interface{}
	for {
		connectionMsg = <-cnfNet.publicChanels[myNodeID].(map[string]chan map[string]interface{})["receiveNodeConnectionMsgChanel"]

		cnfNet.myPrivateChanel["nodeConnectionEventChanel"] <- map[string]interface{}{
			"event":   "receiveMsg",
			"tcpData": connectionMsg,
		}
	}
}
