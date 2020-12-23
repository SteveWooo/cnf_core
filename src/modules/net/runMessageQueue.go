package net

import (
	logger "github.com/cnf_core/src/utils/logger"
)

// RunMessageQueue 处理各个消息队列
func (cnfNet *CnfNet) RunMessageQueue(chanels map[string]chan map[string]interface{}) {
	go cnfNet.HandleDiscoverMsgReceive(chanels)
	go cnfNet.HandleBucketOperate(chanels)
	go cnfNet.HandleNodeConnectionMsg(chanels)
	logger.Info("消息队列启动完成")
}

// HandleDiscoverMsgReceive 关于发现服务的消息处理
func (cnfNet *CnfNet) HandleDiscoverMsgReceive(chanels map[string]chan map[string]interface{}) {
	for {
		udpData := <-chanels["discoverMsgReceiveChanel"] // 发现服务的udp socket如果收不到消息，就会卡死这条协程。

		// TODO: 这里实现端口多路复用。设置一个main节点，用于启动socket
		// 并负责把socket拿到的消息，通过nodeID为唯一标识的chan转发给对应的协程接收

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
		}
	}
}

// HandleBucketOperate 接收数据 主要管理bucket添加节点操作
func (cnfNet *CnfNet) HandleBucketOperate(chanels map[string]chan map[string]interface{}) {
	for {
		bucketOperate := <-chanels["bucketOperateChanel"]
		cnfNet.bucket.ReceiveBucketOperateMsg(bucketOperate)
	}
}

// HandleNodeConnectionMsg 处理TCP数据接收
func (cnfNet *CnfNet) HandleNodeConnectionMsg(chanels map[string]chan map[string]interface{}) {
	for {
		connectionMsg := <-chanels["nodeConnectionMsgChanel"]
		logger.Debug(connectionMsg)
	}
}
