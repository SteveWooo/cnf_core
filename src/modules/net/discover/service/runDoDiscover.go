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
	pingPongCacheloop := 0
	processSeedEventMap := make(map[string]interface{})
	processSeedEventMap["event"] = "processSeed"

	processDoingPingCacheEventMap := make(map[string]interface{})
	processDoingPingCacheEventMap["event"] = "processDoingPingCache"

	// 主动发起邻居查询请求
	doFindNeighborLoop := 0
	doFindNeighborEventMap := make(map[string]interface{})
	doFindNeighborEventMap["event"] = "doFindNeighbor"
	doFindNeighborEventMap["findingNodeID"] = discoverService.myNodeID

	for {
		// timer.Sleep(100 + rand.Intn(100))
		timer.Sleep(2000 + rand.Intn(2000))
		discoverService.myPrivateChanel["discoverEventChanel"] <- processSeedEventMap

		pingPongCacheloop++
		if pingPongCacheloop >= 100 {
			pingPongCacheloop = 0
			// discoverService.myPrivateChanel["discoverEventChanel"] <- processDoingPingCacheEventMap
		}

		doFindNeighborLoop++
		if doFindNeighborLoop >= 50 {
			doFindNeighborLoop = 0
			// discoverService.myPrivateChanel["discoverEventChanel"] <- doFindNeighborEventMap
		}
	}
}

// doFindNeighbor 发起找邻居请求
func (discoverService *DiscoverService) doFindNeighbor(findingNodeID string) {
	// 拿结点桶里的nodeCache来发
	if len(discoverService.myPrivateChanel["bucketNodeListChanel"]) == 0 {
		// logger.Debug(discoverService.conf.(map[string]interface{})["number"].(string) + "桶里无结点可用：" + strconv.Itoa(len(discoverService.myPrivateChanel["bucketNodeListChanel"])))
		return
	}
	discoverService.doFindNeighborNodeCacheListMsg = <-discoverService.myPrivateChanel["bucketNodeListChanel"]
	discoverService.doFindNeighborNodeCacheList = discoverService.doFindNeighborNodeCacheListMsg["nodeList"].([]*commonModels.Node)

	var findNeighborPackString string
	for i := 0; i < len(discoverService.doFindNeighborNodeCacheList); i++ {
		if len(discoverService.myPublicChanel["sendDiscoverMsgChanel"]) == cap(discoverService.myPublicChanel["sendDiscoverMsgChanel"]) {
			continue
		}

		// 找上面配置好的结点
		findNeighborPackString = discoverService.GetFindNodePackString(findingNodeID, discoverService.doFindNeighborNodeCacheList[i].GetNodeID())

		discoverService.DoSend(findNeighborPackString, discoverService.doFindNeighborNodeCacheList[i].GetIP(), discoverService.doFindNeighborNodeCacheList[i].GetServicePort())
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
		if discoverService.runDoDiscoverTempProcessCacheCache == nil {
			continue
		}
		if discoverService.runDoDiscoverTempNow-discoverService.runDoDiscoverTempProcessCacheCache.GetTs() >= 30000 && discoverService.runDoDiscoverTempProcessCacheCache.GetDoingPing() == true {
			// 重发Ping，或者删掉它不管了
			// discoverService.DoPing(discoverService.runDoDiscoverTempProcessCacheCache.GetIP(),
			// 	discoverService.runDoDiscoverTempProcessCacheCache.GetServicePort(),
			// 	discoverService.runDoDiscoverTempProcessCacheCache.GetNodeID())
			discoverService.pingPongCache[discoverService.runDoDiscoverTempProcessCacheNodeID] = nil
		}

	}
}
