package net

import (
	// error "github.com/cnf_core/src/utils/error"
	logger "github.com/cnf_core/src/utils/logger"
	nodeBucket "github.com/cnf_core/src/modules/net/services/nodeBucket"
	discoverService "github.com/cnf_core/src/modules/net/services/discover"

	messageQueue "github.com/cnf_core/src/modules/net/services/messageQueue"
)

func Build() interface{}{
	// 路由桶的初始化
	nodeBucket.Build()

	// 发现服务初始化
	discoverService.Build()

	// 消息队列初始化
	messageQueue.Build()
	
	return nil
}

/**
 * 运行cnf网络
 */
func Run() interface{}{
	// 初始化所有管道
	discoverChanel := make(chan string, 5)

	logger.Info("正在启动Cnf网络组件...")
	
	// 启动发现服务
	go discoverService.Run(discoverChanel)

	// 启动消息队列
	go messageQueue.Run(map[string]chan string {
		"discoverChanel" : discoverChanel,
	})

	return nil
}