package messageQueue

import (
	discoverService "github.com/cnf_core/src/modules/net/services/discover"
	logger "github.com/cnf_core/src/utils/logger"
)

func Build() {

}

/**
 * 从多个管道中获取数据，
 */
func Run(chanels map[string]chan map[string]string) {
	go handleDiscoverMsg(chanels["discoverChanel"])
	logger.Info("消息队列启动完成")
}

/**
 * Discover 服务的消息处理
 */
func handleDiscoverMsg(chanel chan map[string]string) {
	for {
		udpData := <-chanel // 收不到消息，就会卡死这条协程。

		// 这里作为处理发现服务消息的主入口
		data, parseErr := discoverService.ParsePackage(udpData)
		if parseErr != nil {
			// 不处理
			logger.Warn(parseErr.(map[string]interface{})["message"])
			continue
		}

		// 交给发现服务模块处理消息，把结果透传回来即可
		handleErr := discoverService.ReceiveMsg(data)
		if handleErr != nil {
			// 不处理
			logger.Warn(handleErr.(map[string]interface{})["message"])
			continue
		}
	}
}
