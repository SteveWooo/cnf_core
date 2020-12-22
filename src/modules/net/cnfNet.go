package net

import (
	// error "github.com/cnf_core/src/utils/error"
	// config "github.com/cnf_core/src/utils/config"
	discoverService "github.com/cnf_core/src/modules/net/services/discover"
	nodeBucketService "github.com/cnf_core/src/modules/net/services/nodeBucket"
	logger "github.com/cnf_core/src/utils/logger"

	messageQueue "github.com/cnf_core/src/modules/net/services/messageQueue"
)

// Build 初始化网络层的各个服务
func Build() interface{} {
	// 路由桶的初始化
	nodeBucketService.Build()

	// 发现服务初始化
	discoverService.Build()

	// 消息队列初始化
	messageQueue.Build()

	return nil
}

// Run 运行cnf网络
func Run() interface{} {
	// 初始化所有管道
	chanels := map[string]chan map[string]interface{}{
		"discoverMsgChanel": make(chan map[string]interface{}, 5), // 管理udp socket中获取到消息的chanel

		"bucketOperateChanel": make(chan map[string]interface{}, 5), // 一般用于添加bucket节点，或seed
		"bucketSeedChanel":    make(chan map[string]interface{}, 5), // bucket服务往这个通道输送邻居节点，给doDiscover服务用
		"bucketNodeChanel":    make(chan map[string]interface{}, 5), // bucket服务往这个通道输送可用节点，给tcp服务尝试连接。
	}

	logger.Info("正在启动Cnf网络组件...")

	// 启动发现服务
	signal := make(chan bool, 1)
	go discoverService.RunService(chanels, signal)
	<-signal
	logger.Info("Discover Udp 服务监听成功")

	// 启动tcp数据接收服务

	// 启动路由桶服务，
	// 获取、推送种子节点
	// 处理接收节点信息通道
	go nodeBucketService.RunService(chanels)

	// 启动消息队列，其实用于各个chanel之间的消息转发
	go messageQueue.Run(chanels)

	// 开始主动寻找新路由节点
	go discoverService.RunDiscover(chanels)

	// 开始主动尝试建立连接

	return nil
}
