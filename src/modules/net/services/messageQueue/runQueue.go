package messageQueue

import (
	logger "github.com/cnf_core/src/utils/logger"
	discoverService "github.com/cnf_core/src/modules/net/services/discover"
)

func Build (){

}

/**
 * 从多个管道中获取数据，
 */
func Run(chanels map[string]chan string) {
	go handleDiscoverMsg(chanels["discoverChanel"])
	logger.Info("消息队列启动完成")
}

/**
 * Discover 服务的消息处理
 */
func handleDiscoverMsg (chanel chan string) {
	for {
		msg := <- chanel // 收不到消息，就会卡死这条协程。
		logger.Debug("receive:")
		logger.Debug(msg)

		// 这里作为处理发现服务消息的主入口
		jsonMsg, parseErr := discoverService.ParsePackage(msg)
		if parseErr != nil {
			// 不处理
			logger.Warn(parseErr.(map[string]interface{})["message"])
			continue 
		}

		logger.Debug(jsonMsg)
	}
}