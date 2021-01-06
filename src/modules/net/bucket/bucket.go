package bucket

import (
	"math/rand"

	commonModels "github.com/cnf_core/src/modules/net/common/models"
	error "github.com/cnf_core/src/utils/error"
	"github.com/cnf_core/src/utils/timer"
)

// RunService 负责接收Bucket chanel的消息，并持续处理
// 可能添加new tried节点路由，也可能是添加邻居路由
func (bucket *Bucket) RunService(chanels map[string]chan map[string]interface{}) *error.Error {
	bucket.myPrivateChanel = chanels

	// 管理整个bucket的消息队列
	go bucket.HandleBucketOperate()

	seedLoop := 0
	pushBucketSeedEventMap := make(map[string]interface{})
	pushBucketSeedEventMap["event"] = "pushBucketSeed"

	pushNodeListChanelEventMap := make(map[string]interface{})
	pushNodeListChanelEventMap["event"] = "pushNodeListChanel"

	collectSeedFromConfEventMap := make(map[string]interface{})
	collectSeedFromConfEventMap["event"] = "collectSeedFromConf"
	for {
		timer.Sleep(100 + rand.Intn(100))

		if seedLoop == 0 {
			bucket.myPrivateChanel["bucketOperateChanel"] <- collectSeedFromConfEventMap
		}
		seedLoop++
		if seedLoop >= 100 {
			seedLoop = 0
		}

		bucket.myPrivateChanel["bucketOperateChanel"] <- pushBucketSeedEventMap

		bucket.myPrivateChanel["bucketOperateChanel"] <- pushNodeListChanelEventMap
	}
}

// HandleBucketOperate 接收数据 主要管理bucket添加节点操作
func (bucket *Bucket) HandleBucketOperate() {
	chanel := bucket.myPrivateChanel["bucketOperateChanel"]
	var bucketOperate map[string]interface{}
	for {
		bucketOperate = <-chanel
		if bucketOperate["event"] == "addNew" {
			bucket.AddNewNode(bucketOperate["node"].(*commonModels.Node))
		}

		// 批量添加种子结点
		if bucketOperate["event"] == "addSeedByGroup" {
			bucket.AddSeedByGroup(bucketOperate["seeds"].([]*commonModels.Node))
		}

		if bucketOperate["event"] == "pushBucketSeed" {
			bucket.HandleBucketSeed()
		}

		if bucketOperate["event"] == "pushNodeListChanel" {
			bucket.HandleBucketNodeList()
		}

		if bucketOperate["event"] == "collectSeedFromConf" {
			// 桶里如果没有东西，就要找种子源了
			// if len(bucket.nodeCache) != 0 {
			// 	continue
			// }
			bucket.CollectSeedFromConf()
		}
	}
}

// HandleBucketSeed 推送数据 这里主要是对外输出可用Seed，给节点发现服务用为主
func (bucket *Bucket) HandleBucketSeed() {
	// 满了就不要阻塞队列了
	if len(bucket.myPrivateChanel["bucketSeedChanel"]) == cap(bucket.myPrivateChanel["bucketSeedChanel"]) {
		return
	}

	bucket.bucketTempSeed = bucket.GetSeed()
	if bucket.bucketTempSeed == nil {
		// logger.Debug(bucket.conf.(map[string]interface{})["number"].(string) + " no seed")
		return
	}

	// 检查桶里是否已经有这个Node了，有的话就不推了
	if bucket.IsNodeExist(bucket.bucketTempSeed) == true {
		// logger.Debug(bucket.conf.(map[string]interface{})["number"].(string) + " newNode is exists")
		return
	}

	// logger.Debug(bucket.conf.(map[string]interface{})["number"].(string) + " push seed")
	bucket.myPrivateChanel["bucketSeedChanel"] <- map[string]interface{}{
		"node": bucket.bucketTempSeed,
	}
}

// HandleBucketNodeList 输出可用Node列表，外部进行筛选
func (bucket *Bucket) HandleBucketNodeList() {
	// 满了就不要阻塞队列了
	if len(bucket.myPrivateChanel["bucketNodeListChanel"]) == cap(bucket.myPrivateChanel["bucketNodeListChanel"]) {
		return
	}

	bucket.bucketTempNodeList = bucket.GetNodeList()
	bucket.bucketTempNodeListWithDistance = bucket.GetNodeListWithDistance()
	if len(bucket.bucketTempNodeListWithDistance) == 0 {
		return
	}
	bucket.myPrivateChanel["bucketNodeListChanel"] <- map[string]interface{}{
		"nodeList":             bucket.bucketTempNodeList,
		"nodeListWithDistance": bucket.bucketTempNodeListWithDistance,
	}
}
