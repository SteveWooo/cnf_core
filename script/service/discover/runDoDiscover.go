package discover

import (
	commonModels "github.com/cnf_core/src/modules/net/models/common"
	discoverModel "github.com/cnf_core/src/modules/net/models/discover"
	logger "github.com/cnf_core/src/utils/logger"
	"github.com/cnf_core/src/utils/timer"
)

// RunDiscover 主动寻找种子节点或邻居节点，进行连接
func RunDiscover(chanels map[string]chan map[string]interface{}) {
	go processSeed(chanels)
	go processDoingPingCache(chanels)
}

// processSeed 持续处理种子，把种子塞进DoingPing缓存里面
func processSeed(chanels map[string]chan map[string]interface{}) {
	// 不断获取seed，然后使用shaker发起握手。
	for {

		seed := <-chanels["bucketSeedChanel"]
		if seed == nil {
			logger.Warn("主动发现节点模块中，获取到空的种子节点")
			continue
		}
		seedNode := seed["node"].(*commonModels.Node)

		nodeID := seedNode.GetNodeID()
		newCache := discoverModel.CreatePingPongCache(nodeID)
		// 设置标识为主动发起的缓存
		newCache.SetDoingPing()
		pingPongCache[nodeID] = newCache

		shaker.DoPing(seedNode.GetIP(), seedNode.GetServicePort())
	}
}

// processDoingPingCache 持续处理缓存队列的数据
// 用于检查
func processDoingPingCache(chanels map[string]chan map[string]interface{}) {
	for {
		// for _, cache := range pingPongCache {
		// 检查
		// }
		timer.Sleep(1000) // 持续检查缓存队列的情况
	}
}
