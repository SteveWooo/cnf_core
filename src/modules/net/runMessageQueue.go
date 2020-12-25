package net

import (
	"github.com/cnf_core/src/utils/config"
	logger "github.com/cnf_core/src/utils/logger"
)

// RunMessageQueue 处理各个消息队列
func (cnfNet *CnfNet) RunMessageQueue(chanels map[string]chan map[string]interface{}) {
	// 路由bucket 的队列
	go cnfNet.HandleBucketOperate(chanels)

	// discover 接收消息的队列
	go cnfNet.HandleDiscoverMsgReceive(chanels)
	go cnfNet.HandleSubNodeDiscoverMsgReceive(chanels)

	// discover 消息发送队列处理
	go cnfNet.HandleDiscoverMsgSend(chanels)

	// nodeConn 消息接收队列
	go cnfNet.HandleNodeConnectionMsgReceive(chanels)
	go cnfNet.HandleSubNodeConnectionMsgReceive(chanels)

	logger.Info("消息队列启动完成")
}

// HandleDiscoverMsgReceive 关于发现服务的消息处理
func (cnfNet *CnfNet) HandleDiscoverMsgReceive(chanels map[string]chan map[string]interface{}) {
	confNet := cnfNet.conf.(map[string]interface{})["net"]
	if confNet.(map[string]interface{})["masterServer"] == "true" {
		for {
			udpData := <-chanels["receiveDiscoverMsgChanel"] // 发现服务的udp socket如果收不到消息，就会卡死这条协程。
			chanel, exist := cnfNet.publicChanels[udpData["targetNodeID"].(string)]
			// logger.Debug(udpData["targetNodeID"].(string) + " receive")
			if exist {
				chanel.(map[string]chan map[string]interface{})["receiveDiscoverMsgChanel"] <- udpData
			} else {
				logger.Error("接收到未知Udp数据包，本服务器无该节点")
			}
		}
	}
}

// HandleSubNodeDiscoverMsgReceive 公共节点接收master节点的消息推送服务
func (cnfNet *CnfNet) HandleSubNodeDiscoverMsgReceive(chanels map[string]chan map[string]interface{}) {
	for {
		myNodeID := config.ParseNodeID(cnfNet.conf)
		udpData := <-cnfNet.publicChanels[myNodeID].(map[string]chan map[string]interface{})["receiveDiscoverMsgChanel"]

		// 交给发现服务模块处理消息，把结果透传回来即可
		bucketOperate, receiveErr := cnfNet.discover.ReceiveMsg(udpData)
		if receiveErr != nil {
			// 不处理
			logger.Warn(receiveErr.GetMessage())
			continue
		}

		// 处理路由Bucket逻辑
		if bucketOperate != nil {
			// 由于Bucket操作有可能在tcp消息中出现，所有需要用一个chanel锁住。
			chanels["bucketOperateChanel"] <- bucketOperate.(map[string]interface{})
			// logger.Debug(bucketOperate.(map[string]interface{}))
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
				for {
					sendingData := <-c["sendDiscoverMsgChanel"]
					cnfNet.discover.SendMsg(sendingData["message"].(string), sendingData["targetIP"].(string), sendingData["targetServicePort"].(string))
				}
			}
			go getMessageFromSubChanel(nodeID, chanel.(map[string]chan map[string]interface{}))
		}
	}
}

// HandleBucketOperate 接收数据 主要管理bucket添加节点操作
func (cnfNet *CnfNet) HandleBucketOperate(chanels map[string]chan map[string]interface{}) {
	for {
		bucketOperate := <-chanels["bucketOperateChanel"]
		// logger.Debug(bucketOperate)
		cnfNet.bucket.ReceiveBucketOperateMsg(bucketOperate)
	}
}

// HandleNodeConnectionMsgReceive 处理TCP数据接收
func (cnfNet *CnfNet) HandleNodeConnectionMsgReceive(chanels map[string]chan map[string]interface{}) {
	confNet := cnfNet.conf.(map[string]interface{})["net"]
	if confNet.(map[string]interface{})["masterServer"] == "true" {
		for {
			connectionMsg := <-chanels["receiveNodeConnectionMsgChanel"]
			tcpData := connectionMsg["tcpData"]
			chanel, exist := cnfNet.publicChanels[tcpData.(map[string]interface{})["targetNodeID"].(string)]
			if exist {
				chanel.(map[string]chan map[string]interface{})["receiveNodeConnectionMsgChanel"] <- connectionMsg
			} else {
				logger.Error("接收到未知NodeID tcp数据包，本服务器无该节点")
			}
		}
	}
}

// HandleSubNodeConnectionMsgReceive 公共节点接收master节点的消息推送服务
func (cnfNet *CnfNet) HandleSubNodeConnectionMsgReceive(chanels map[string]chan map[string]interface{}) {
	for {
		myNodeID := config.ParseNodeID(cnfNet.conf)
		connectionMsg := <-cnfNet.publicChanels[myNodeID].(map[string]chan map[string]interface{})["receiveNodeConnectionMsgChanel"]
		// 交给发现服务模块处理消息，把结果透传回来即可
		connectionMsgData, nodeConnReceiveErr := cnfNet.nodeConnection.ReceiveMsg(connectionMsg)
		if nodeConnReceiveErr != nil {
			continue
		}

		// 检查桶里有没有这个节点，如果还没有在桶里出现的话，就断开拒绝连接

		cnfNet.nodeConnection.HandleMsg(connectionMsgData)

		// 接收消息经过简单的处理后，在这里判断一下这个消息能不能用。

		// if receiveErr != nil {
		// 	// 不处理
		// 	logger.Warn(receiveErr.GetMessage())
		// 	continue
		// }

		// // 处理路由Bucket逻辑
		// if bucketOperate != nil {
		// 	// 由于Bucket操作有可能在tcp消息中出现，所有需要用一个chanel锁住。
		// 	chanels["bucketOperateChanel"] <- bucketOperate.(map[string]interface{})
		// }
	}
}
