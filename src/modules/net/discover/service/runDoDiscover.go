package service

import (
	"math/rand"

	commonModels "github.com/cnf_core/src/modules/net/common/models"
	discoverModel "github.com/cnf_core/src/modules/net/discover/models"

	// commonModels "github.com/cnf_core/src/modules/net/common/models"

	"github.com/cnf_core/src/utils/config"
	"github.com/cnf_core/src/utils/logger"
	"github.com/cnf_core/src/utils/timer"
)

// RunDoDiscover 主动寻找种子节点或邻居节点，进行连接
func (discoverService *DiscoverService) RunDoDiscover(chanels map[string]chan map[string]interface{}) {
	// 应该用一个chanel来进行这个任务
	// go discoverService.processSeed(chanels)
	// go discoverService.processDoingPingCache(chanels)
	loop := 0
	processSeedEventMap := make(map[string]interface{})
	processSeedEventMap["event"] = "processSeed"

	processDoingPingCacheEventMap := make(map[string]interface{})
	processDoingPingCacheEventMap["event"] = "processDoingPingCache"
	for {
		timer.Sleep(100 + rand.Intn(100))
		discoverService.myPrivateChanel["discoverEventChanel"] <- processSeedEventMap

		loop++
		if loop >= 5 {
			loop = 0
			// discoverService.myPrivateChanel["processDoingPingCache"] <- map[string]interface{}{
			// 	"event": "processDoingPingCache",
			// }
		}
	}
}

// ProcessSeed 持续处理种子，把种子塞进DoingPing缓存里面
func (discoverService *DiscoverService) processSeed(chanels map[string]chan map[string]interface{}) {
	// 没有种子可以处理，就不管他了，不要被卡死
	if len(chanels["bucketSeedChanel"]) == 0 {
		return
	}
	discoverService.runDoDiscoverTemp["seed"] = <-chanels["bucketSeedChanel"]
	if discoverService.runDoDiscoverTemp["seed"] == nil {
		logger.Warn("主动发现节点模块中，获取到空的种子节点")
		return
	}

	discoverService.runDoDiscoverTemp["seedNode"] = (discoverService.runDoDiscoverTemp["seed"].(map[string]interface{})["node"]).(*commonModels.Node)

	// 不要发现自己
	if discoverService.runDoDiscoverTemp["seedNode"].(*commonModels.Node).GetNodeID() == config.ParseNodeID(discoverService.conf) {
		logger.Debug(discoverService.conf.(map[string]interface{})["number"].(string) + " get myself")
		return
	}

	// 正在doingPong的不要重复做
	_, discoverService.runDoDiscoverTemp["isDoingPingPong"] = discoverService.pingPongCache[discoverService.runDoDiscoverTemp["seedNodeNodeID"].(string)]
	if discoverService.runDoDiscoverTemp["isDoingPingPong"] == true {
		// logger.Debug("is doing ping")
		return
	}

	discoverService.runDoDiscoverTemp["seedNodeNodeID"] = discoverService.runDoDiscoverTemp["seedNode"].(*commonModels.Node).GetNodeID()

	newCache := discoverModel.CreatePingPongCache(discoverService.runDoDiscoverTemp["seedNodeNodeID"].(string))
	// 设置标识为主动发起的缓存
	newCache.SetDoingPing(discoverService.runDoDiscoverTemp["seedNode"].(*commonModels.Node).GetIP(),
		discoverService.runDoDiscoverTemp["seedNode"].(*commonModels.Node).GetServicePort())

	discoverService.pingPongCache[discoverService.runDoDiscoverTemp["seedNodeNodeID"].(string)] = nil
	discoverService.pingPongCache[discoverService.runDoDiscoverTemp["seedNodeNodeID"].(string)] = newCache

	discoverService.DoPing(discoverService.runDoDiscoverTemp["seedNode"].(*commonModels.Node).GetIP(),
		discoverService.runDoDiscoverTemp["seedNode"].(*commonModels.Node).GetServicePort(),
		discoverService.runDoDiscoverTemp["seedNodeNodeID"].(string))
}

// ProcessDoingPingCache 持续处理缓存队列的数据
// 用于检查
func (discoverService *DiscoverService) processDoingPingCache(chanels map[string]chan map[string]interface{}) {
	discoverService.runDoDiscoverTempNow = timer.Now()
	for discoverService.runDoDiscoverTempProcessCacheNodeID, discoverService.runDoDiscoverTempProcessCacheCache = range discoverService.pingPongCache {
		// 不能太快，不然网络卡一卡就卡没了
		if discoverService.runDoDiscoverTempNow-discoverService.runDoDiscoverTempProcessCacheCache.GetTs() >= 30000 && discoverService.runDoDiscoverTempProcessCacheCache.GetDoingPing() == true {
			// 重发Ping
			// logger.Debug(discoverService.conf.(map[string]interface{})["number"].(string) + " 11")
			discoverService.DoPing(discoverService.runDoDiscoverTempProcessCacheCache.GetIP(),
				discoverService.runDoDiscoverTempProcessCacheCache.GetServicePort(),
				discoverService.runDoDiscoverTempProcessCacheCache.GetNodeID())
			// discoverService.pingPongCache[discoverService.runDoDiscoverTempProcessCacheNodeID] = nil
		}

	}
}
