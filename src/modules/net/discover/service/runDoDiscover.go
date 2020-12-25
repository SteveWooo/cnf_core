package service

import (
	commonModels "github.com/cnf_core/src/modules/net/common/models"
	discoverModel "github.com/cnf_core/src/modules/net/discover/models"
	"github.com/cnf_core/src/utils/logger"
	"github.com/cnf_core/src/utils/timer"
)

// RunDoDiscover 主动寻找种子节点或邻居节点，进行连接
func (discoverService *DiscoverService) RunDoDiscover(chanels map[string]chan map[string]interface{}) {
	go discoverService.ProcessSeed(chanels)
	go discoverService.ProcessDoingPingCache(chanels)
}

// ProcessSeed 持续处理种子，把种子塞进DoingPing缓存里面
func (discoverService *DiscoverService) ProcessSeed(chanels map[string]chan map[string]interface{}) {
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
		// logger.Debug(config.ParseNodeID(discoverService.conf) + " get seed: " + nodeID)
		// 设置标识为主动发起的缓存
		newCache.SetDoingPing(seedNode.GetIP(), seedNode.GetServicePort())
		discoverService.pingPongCache[nodeID] = newCache

		discoverService.DoPing(seedNode.GetIP(), seedNode.GetServicePort(), nodeID)
	}
}

// ProcessDoingPingCache 持续处理缓存队列的数据
// 用于检查
func (discoverService *DiscoverService) ProcessDoingPingCache(chanels map[string]chan map[string]interface{}) {
	for {
		now := timer.Now()
		for _, cache := range discoverService.pingPongCache {
			// 不能删除太快，不然网络卡一卡就卡没了
			if now-cache.GetTs() >= 120000 && cache.GetDoingPing() == true {
				// 重发Ping
				discoverService.DoPing(cache.GetIP(), cache.GetServicePort(), cache.GetNodeID())
			}

		}
		timer.Sleep(1000) // 持续检查缓存队列的情况
	}
}
