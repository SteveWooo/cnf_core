package service

import (
	"math/rand"

	commonModels "github.com/cnf_core/src/modules/net/common/models"
	discoverModel "github.com/cnf_core/src/modules/net/discover/models"
	"github.com/cnf_core/src/utils/config"
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
		timer.Sleep(1000 + rand.Intn(1000))
		// fmt.Println(discoverService.conf.(map[string]interface{})["number"].(string)+" (inRundiscover) bucket chanel Addr: ", chanels["bucketSeedChanel"])
		// logger.Debug(discoverService.conf.(map[string]interface{})["number"].(string) + " todo get seed")
		seed := <-chanels["bucketSeedChanel"]
		// logger.Debug(discoverService.conf.(map[string]interface{})["number"].(string) + " done get seed")
		if seed == nil {
			logger.Warn("主动发现节点模块中，获取到空的种子节点")
			continue
		}
		// logger.Debug("Do discover")
		seedNode := seed["node"].(*commonModels.Node)

		// 不要发现自己
		if seedNode.GetNodeID() == config.ParseNodeID(discoverService.conf) {
			logger.Debug(discoverService.conf.(map[string]interface{})["number"].(string) + " get myself")
			continue
		}

		nodeID := seedNode.GetNodeID()
		newCache := discoverModel.CreatePingPongCache(nodeID)
		// logger.Debug(discoverService.conf.(map[string]interface{})["number"].(string) + " get seed: " + nodeID)
		// 设置标识为主动发起的缓存
		newCache.SetDoingPing(seedNode.GetIP(), seedNode.GetServicePort())
		// discoverService.pingPongCacheLock <- true
		discoverService.pingPongCache[nodeID] = newCache
		// <-discoverService.pingPongCacheLock
		discoverService.DoPing(seedNode.GetIP(), seedNode.GetServicePort(), nodeID)
	}
}

// ProcessDoingPingCache 持续处理缓存队列的数据
// 用于检查
func (discoverService *DiscoverService) ProcessDoingPingCache(chanels map[string]chan map[string]interface{}) {
	for {
		now := timer.Now()
		// discoverService.pingPongCacheLock <- true
		for _, cache := range discoverService.pingPongCache {
			// 不能太快，不然网络卡一卡就卡没了
			if now-cache.GetTs() >= 30000 && cache.GetDoingPing() == true {
				// 重发Ping
				// logger.Debug(discoverService.conf.(map[string]interface{})["number"].(string) + " 11")
				discoverService.DoPing(cache.GetIP(), cache.GetServicePort(), cache.GetNodeID())
			}

		}
		// <-discoverService.pingPongCacheLock
		timer.Sleep(1000) // 持续检查缓存队列的情况
	}
}
