package service

import (
	"math/rand"

	"github.com/cnf_core/src/utils/router"
	"github.com/cnf_core/src/utils/timer"
)

// RunDoDiscoverByMasterArea 主动寻找种子节点或邻居节点，进行连接
func (discoverService *DiscoverService) RunDoDiscoverByMasterArea(chanels map[string]chan map[string]interface{}) {
	// 应该用一个chanel来进行这个任务
	pingPongCacheloop := 0
	processSeedEventMap := make(map[string]interface{})
	processSeedEventMap["event"] = "processSeed"

	processDoingPingCacheEventMap := make(map[string]interface{})
	processDoingPingCacheEventMap["event"] = "processDoingPingCache"

	// 主动发起邻居查询请求
	doFindNeighborLoop := 0
	doFindNeighborTarget := ""     // 当前需要找的邻居
	doFindNeighborMasterIndex := 0 // 区域Master结点需要用上的参数，代表这一轮找到第几个邻居

	doFindNeighborEventMap := make(map[string]interface{})
	doFindNeighborEventMap["event"] = "doFindNeighbor"

	for {
		timer.Sleep(100 + rand.Intn(100))
		discoverService.myPrivateChanel["discoverEventChanel"] <- processSeedEventMap

		pingPongCacheloop++
		if pingPongCacheloop >= 100 {
			pingPongCacheloop = 0
			discoverService.myPrivateChanel["discoverEventChanel"] <- processDoingPingCacheEventMap
		}

		doFindNeighborLoop++
		if doFindNeighborLoop >= 20 {
			doFindNeighborLoop = 0

			// 区域master结点需要循环找别的区域的master结点
			if discoverService.masterAreaIsMaster == true {
				doFindNeighborTarget = router.MasterAreas[doFindNeighborMasterIndex]
				doFindNeighborMasterIndex++
				// 循环完了，要重置一遍
				if doFindNeighborMasterIndex >= len(router.MasterAreas) {
					doFindNeighborMasterIndex = 0
				}
			}

			// 如果不是区域Master结点，就要尽量找本区域的Master结点
			if discoverService.masterAreaIsMaster == false {
				doFindNeighborTarget = router.MasterAreas[discoverService.masterAreaLocateArea]
			}

			doFindNeighborEventMap["findingNodeID"] = doFindNeighborTarget

			discoverService.myPrivateChanel["discoverEventChanel"] <- doFindNeighborEventMap
		}
	}
}
