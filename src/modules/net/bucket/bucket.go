package bucket

import (
	"strconv"

	commonModels "github.com/cnf_core/src/modules/net/common/models"
	config "github.com/cnf_core/src/utils/config"
	error "github.com/cnf_core/src/utils/error"
	"github.com/cnf_core/src/utils/logger"
	"github.com/cnf_core/src/utils/timer"
)

// TODO 这些都加到配置里面8
var NEW_BUCKET_COUNT int = 64
var NEW_BUCKET_LENGTH int = 64
var TRIED_BUCKET_COUNT int = 64
var TRIED_BUCKET_LENGTH int = 16

// Bucket 路由桶
type Bucket struct {
	conf        interface{}
	newBucket   map[int][]*commonModels.Node
	triedBucket map[int][]*commonModels.Node

	seed         []*commonModels.Node
	maxSeedCount int

	kv chan bool // 关于两个bucket和seed的读写锁
}

// Build 初始化路由桶
func (bucket *Bucket) Build(conf interface{}) {
	bucket.conf = conf
	bucket.newBucket = make(map[int][]*commonModels.Node) // 初始化数组对象本身
	for i := 0; i <= NEW_BUCKET_COUNT; i++ {
		bucket.newBucket[i] = make([]*commonModels.Node, NEW_BUCKET_LENGTH+1) // 声明第一个维度每个对象都是一个子数组
	}

	bucket.triedBucket = make(map[int][]*commonModels.Node) // 初始化数组对象本身
	for i := 0; i <= TRIED_BUCKET_COUNT; i++ {
		bucket.triedBucket[i] = make([]*commonModels.Node, TRIED_BUCKET_LENGTH+1) // 声明第一个维度每个对象都是一个子数组
	}

	// 种子功能初始化
	netConf := config.GetNetConf()
	maxSeedCount, maxCountErr := strconv.Atoi(netConf.(map[string]interface{})["maxSeedCount"].(string))
	if maxCountErr != nil {
		logger.Warn("config miss: maxSeedCount. Using default.")
		bucket.maxSeedCount = 100
	}
	bucket.maxSeedCount = maxSeedCount

	bucket.kv = make(chan bool, 1)

	bucket.CollectSeedFromConf()
}

// RunService 负责接收Bucket chanel的消息，并持续处理
// 可能添加new tried节点路由，也可能是添加邻居路由
func (bucket *Bucket) RunService(chanels map[string]chan map[string]interface{}) *error.Error {
	// 接收的
	// go bucket.HandleBucketOperate(chanels["bucketOperateChanel"])

	// 往外输出的
	go bucket.HandleBucketSeed(chanels["bucketSeedChanel"])
	go bucket.HandleBucketNode(chanels["bucketNodeChanel"])

	return nil
}

// ReceiveBucketOperateMsg 消息队列推送Bucket操作事件过来时响应
func (bucket *Bucket) ReceiveBucketOperateMsg(bucketOperate map[string]interface{}) {
	node := bucketOperate["node"].(*commonModels.Node)
	if bucketOperate["bucketEvent"] == "addNew" {
		addNewErr := bucket.AddNewNode(node)
		if addNewErr != nil {
			// 说明节点已经存在，不需要重复添加
		}
	}
}

// HandleBucketSeed 推送数据 这里主要是对外输出可用Seed，给节点发现服务用为主
func (bucket *Bucket) HandleBucketSeed(chanel chan map[string]interface{}) {
	for {
		seed := bucket.GetSeed()
		// 没有种子，就消停会，再从配置里面获取新的种子
		if seed == nil {
			timer.Sleep(10000)
			bucket.CollectSeedFromConf()
			continue
		}

		chanel <- map[string]interface{}{
			"node": seed,
		}
	}
}

// HandleBucketNode 输出可用Node 主要对外输出可用Node，给节点连接服务为主
func (bucket *Bucket) HandleBucketNode(chanel chan map[string]interface{}) {
	for {
		node := bucket.GetRandomNode()

		if node == nil {
			continue
		}

		chanel <- map[string]interface{}{
			"node": node,
		}
	}
}
