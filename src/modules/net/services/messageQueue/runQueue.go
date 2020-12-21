package messagequeue

import (
	discoverService "github.com/cnf_core/src/modules/net/services/discover"
	logger "github.com/cnf_core/src/utils/logger"
)

func Build() {

}

// Run 从多个管道中获取数据，
func Run(chanels map[string]chan map[string]interface{}) {
	go handleDiscoverMsg(chanels)
	logger.Info("消息队列启动完成")
}

/**
 * Discover 服务的消息处理
 */
func handleDiscoverMsg(chanels map[string]chan map[string]interface{}) {
	for {
		udpData := <-chanels["discoverMsgChanel"] // 发现服务的udp socket如果收不到消息，就会卡死这条协程。
		// 这里作为处理发现服务消息的主入口
		data, parseErr := discoverService.ParsePackage(udpData)
		if parseErr != nil {
			// 不处理
			// logger.Warn(parseErr.(error.Error).GetMessage())
			logger.Warn(parseErr.GetMessage())
			continue
		}

		// 交给发现服务模块处理消息，把结果透传回来即可
		bucketOperate, receiveErr := discoverService.ReceiveMsg(data)
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
