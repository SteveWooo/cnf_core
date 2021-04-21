package bucket

import (
	"math/rand"
	"strconv"

	commonModels "github.com/cnf_core/src/modules/net/common/models"
	config "github.com/cnf_core/src/utils/config"
	error "github.com/cnf_core/src/utils/error"
	"github.com/cnf_core/src/utils/logger"
	"github.com/cnf_core/src/utils/router"
)

// TODO 这些都加到配置里面8
var NEW_BUCKET_COUNT int = 64
var NEW_BUCKET_LENGTH int = 64
var TRIED_BUCKET_COUNT int = 64
var TRIED_BUCKET_LENGTH int = 16

// Bucket 路由桶
type Bucket struct {
	conf            interface{}
	myNodeID        string
	newBucketLock   chan bool
	newBucket       map[int][]*commonModels.Node
	triedBucketLock chan bool
	triedBucket     map[int][]*commonModels.Node

	// 用来存桶里所有节点，之后方便拿出来用
	nodeCacheLock         chan bool
	nodeCache             []*commonModels.Node
	nodeCacheWithDistance []map[string]interface{}
	maxNodeCacheCount     int
	// 缓存桶里的节点，用hashmap来存，高性能查询
	nodeCacheHashMap map[string]bool

	seedLock     chan bool
	seed         []*commonModels.Node
	maxSeedCount int

	// 临时变量
	bucketTempSeed                 *commonModels.Node
	bucketTempNodeList             []*commonModels.Node
	bucketTempNodeListWithDistance []map[string]interface{}
	bucketTempIsNodeExist          bool

	// 自己的内部频道
	myPrivateChanel map[string]chan map[string]interface{}
}

// Build 初始化路由桶
func (bucket *Bucket) Build(conf interface{}) {
	bucket.conf = conf
	bucket.myNodeID = config.ParseNodeID(conf)
	// bucket.newBucket = make(map[int][]*commonModels.Node) // 初始化数组对象本身
	// for i := 0; i <= NEW_BUCKET_COUNT; i++ {
	// 	bucket.newBucket[i] = make([]*commonModels.Node, NEW_BUCKET_LENGTH+1) // 声明第一个维度每个对象都是一个子数组
	// }

	// bucket.triedBucket = make(map[int][]*commonModels.Node) // 初始化数组对象本身
	// for i := 0; i <= TRIED_BUCKET_COUNT; i++ {
	// 	bucket.triedBucket[i] = make([]*commonModels.Node, TRIED_BUCKET_LENGTH+1) // 声明第一个维度每个对象都是一个子数组
	// }

	// 缓存最新加进来的节点，随机就找他们了
	bucket.nodeCache = make([]*commonModels.Node, 0)
	bucket.maxNodeCacheCount = NEW_BUCKET_COUNT*(NEW_BUCKET_LENGTH+1) + TRIED_BUCKET_COUNT*(TRIED_BUCKET_LENGTH+1)
	bucket.nodeCacheHashMap = make(map[string]bool)

	bucket.newBucketLock = make(chan bool, 1)
	bucket.triedBucketLock = make(chan bool, 1)
	bucket.seedLock = make(chan bool, 1)
	bucket.nodeCacheLock = make(chan bool, 1)

	// 种子功能初始化
	netConf := config.GetNetConf()
	maxSeedCount, maxCountErr := strconv.Atoi(netConf.(map[string]interface{})["maxSeedCount"].(string))
	if maxCountErr != nil {
		logger.Warn("config miss: maxSeedCount. Using default.")
		bucket.maxSeedCount = 100
	}
	bucket.maxSeedCount = maxSeedCount

	// bucket.CollectSeedFromConf()
}

//CollectSeedFromConf 从配置中获取种子
func (bucket *Bucket) CollectSeedFromConf() {
	// netConf := config.GetNetConf()
	netConf := bucket.conf.(map[string]interface{})["net"]
	// 把配置文件里面的种子列表加进来
	confSeeds := netConf.(map[string]interface{})["seed"].([]interface{})
	for i := 0; i < len(confSeeds); i++ {
		node, nodeCreateErr := commonModels.CreateNode(map[string]interface{}{
			"nodeID":      confSeeds[i].(map[string]interface{})["nodeID"],
			"ip":          confSeeds[i].(map[string]interface{})["ip"],
			"servicePort": confSeeds[i].(map[string]interface{})["servicePort"],
		})
		if nodeCreateErr != nil {
			continue
		}
		bucket.AddSeed(node)
	}
}

// AddSeedByGroup 批量添加种子结点
func (bucket *Bucket) AddSeedByGroup(seeds []*commonModels.Node) *error.Error {
	// logger.Debug(bucket.conf.(map[string]interface{})["number"].(string) + " adding seed length: " + strconv.Itoa(len(seeds)))
	for i := 0; i < len(seeds); i++ {
		bucket.AddSeed(seeds[i])
	}
	// logger.Debug(bucket.conf.(map[string]interface{})["number"].(string) + " seed length: " + strconv.Itoa(len(bucket.seed)))
	// 清空一下，防止其他骚操作导致没被GC
	seeds = nil
	return nil
}

// AddSeed 添加一个节点到种子
// 这里有可能是在初始化阶段就添加，也有可能在收到邻居信息的时候添加
func (bucket *Bucket) AddSeed(node *commonModels.Node) *error.Error {
	// 先检查桶里有没有这个节点
	exist := bucket.IsNodeExist(node)
	if exist {
		// logger.Debug(bucket.conf.(map[string]interface{})["number"].(string) + " node exiss")
		return nil
	}

	// 再查找seed桶里是否有重复
	bucket.seedLock <- true
	for i := 0; i < len(bucket.seed); i++ {
		if bucket.seed[i].GetNodeID() == node.GetNodeID() {
			return nil
		}
	}

	// 如果超出限制的长度，就先取一个走
	if len(bucket.seed) >= bucket.maxSeedCount {
		bucket.GetSeed()
	}

	// 把配置文件里面的种子列表加进来
	bucket.seed = append(bucket.seed, node)
	<-bucket.seedLock
	return nil
}

// GetSeed 取出一个种子节点，并删除
func (bucket *Bucket) GetSeed() *commonModels.Node {
	if len(bucket.seed) == 0 {
		return nil
	}
	bucket.seedLock <- true
	node := bucket.seed[0:1]
	bucket.seed = bucket.seed[1:]

	n := node[0]
	<-bucket.seedLock

	return n
}

// AddNewNode 添加节点到新桶里
func (bucket *Bucket) AddNewNode(n *commonModels.Node) *error.Error {
	if bucket.IsNodeExist(n) {
		return error.New(map[string]interface{}{
			"message": "新节点已存在桶里，无需重复添加",
		})
	}

	// bucket.newBucketLock <- true
	// myNodeID := config.ParseNodeID(bucket.conf)
	// bucketNum := router.CalculateDistance(myNodeID, n.GetNodeID()) // 获得bucket编号
	// // 超长的先丢掉，就算不超长，都先把第一个元素丢掉，然后再新增一个新元素进来
	// bucket.newBucket[bucketNum] = bucket.newBucket[bucketNum][1:]
	// bucket.newBucket[bucketNum] = append(bucket.newBucket[bucketNum], n)
	// <-bucket.newBucketLock

	bucket.AddNodeCache(n)
	return nil
}

// AddNodeCache 把节点放入缓存中，按照距离远近放入
func (bucket *Bucket) AddNodeCache(n *commonModels.Node) *error.Error {
	bucket.nodeCacheLock <- true
	// 超长的话先删掉一个
	if len(bucket.nodeCache) >= bucket.maxNodeCacheCount {
		bucket.nodeCache = bucket.nodeCache[0 : len(bucket.nodeCache)-1]
		bucket.nodeCacheWithDistance = bucket.nodeCacheWithDistance[0 : len(bucket.nodeCacheWithDistance)-1]
	}

	// 放入缓存就按照距离排序好
	detailDistance := router.CalculateDetailDistance(bucket.myNodeID, n.GetNodeID())
	// 为空直接放入
	if len(bucket.nodeCache) == 0 {
		bucket.nodeCache = append(bucket.nodeCache, n)
		bucket.nodeCacheWithDistance = append(bucket.nodeCacheWithDistance, map[string]interface{}{
			"node":           n,
			"detailDistance": detailDistance,
		})
		bucket.nodeCacheHashMap[n.GetNodeID()] = true
		<-bucket.nodeCacheLock
		return nil
	}

	for i := 0; i < len(bucket.nodeCache); i++ {
		nodeDistance := router.CalculateDetailDistance(bucket.myNodeID, bucket.nodeCache[i].GetNodeID())
		isInsert := false
		for k := 0; k < len(detailDistance); k++ {
			// 距离远就放后面
			if detailDistance[k] > nodeDistance[k] {
				break
			}

			if detailDistance[k] == nodeDistance[k] {
				continue
			}

			// 插入排序
			bigNodeIndex := i
			newCache := []*commonModels.Node{n}
			var newCacheWithDistance []map[string]interface{}
			newCacheWithDistance = append(newCacheWithDistance, map[string]interface{}{
				"node":           n,
				"detailDistance": detailDistance,
			})

			if bigNodeIndex == 0 {
				bucket.nodeCache = append(newCache, bucket.nodeCache...)
				bucket.nodeCacheWithDistance = append(newCacheWithDistance, bucket.nodeCacheWithDistance...)
			} else {
				frontCache := bucket.nodeCache[0:bigNodeIndex]
				afterCache := bucket.nodeCache[bigNodeIndex:]
				bucket.nodeCache = append(frontCache, append(newCache, afterCache...)...)

				frontCacheWithDistance := bucket.nodeCacheWithDistance[0:bigNodeIndex]
				afterCacheWithDistance := bucket.nodeCacheWithDistance[bigNodeIndex:]
				bucket.nodeCacheWithDistance = append(frontCacheWithDistance, append(newCacheWithDistance, afterCacheWithDistance...)...)
			}

			isInsert = true
			break
		}

		// 如果比最后一位都要大，那么就直接放最后
		if i == len(bucket.nodeCache)-1 && isInsert == false {
			bucket.nodeCache = append(bucket.nodeCache, n)
			bucket.nodeCacheWithDistance = append(bucket.nodeCacheWithDistance, map[string]interface{}{
				"node":           n,
				"detailDistance": detailDistance,
			})
			isInsert = true
		}

		if isInsert {
			break
		}
	}

	bucket.nodeCacheHashMap[n.GetNodeID()] = true
	<-bucket.nodeCacheLock
	return nil
}

// GetNodeList 获取当前桶里的所有节点，按照距离排序
func (bucket *Bucket) GetNodeList() []*commonModels.Node {
	// var list []*commonModels.Node
	// listCount := 0
	// bucket.nodeCacheLock <- true
	// for i := 0; i < len(bucket.nodeCache); i++ {
	// 	if bucket.nodeCache[i] != nil {
	// 		listCount++
	// 		list = append(list, bucket.nodeCache[i])
	// 	}
	// }
	// <-bucket.nodeCacheLock
	// return list

	return bucket.nodeCache
}

// GetNodeListWithDistance 获取带距离的结点列表
func (bucket *Bucket) GetNodeListWithDistance() []map[string]interface{} {
	return bucket.nodeCacheWithDistance
}

// GetRandomNode 获得一个随机节点，给tcp服务进行连接尝试
func (bucket *Bucket) GetRandomNode() *commonModels.Node {
	// 全量获取
	// var b map[int][]*commonModels.Node
	// var bucketCount int
	// bucket.kv <- true
	// // 先随机出一个桶
	// bucketRand := rand.Intn(99)
	// if bucketRand < 50 {
	// 	b = bucket.newBucket
	// 	bucketCount = NEW_BUCKET_COUNT + 1
	// } else {
	// 	b = bucket.triedBucket
	// 	bucketCount = TRIED_BUCKET_COUNT + 1
	// }
	// nodes := make([]*commonModels.Node, 0)
	// for i := 0; i < bucketCount; i++ {
	// 	for k := 0; k < len(b[i]); k++ {
	// 		if b[i][k] != nil {
	// 			nodes = append(nodes, b[i][k])
	// 		}
	// 	}
	// }

	// if len(nodes) == 0 {
	// 	<-bucket.kv
	// 	return nil
	// }

	// ranIndex := rand.Intn(len(nodes))
	// node := nodes[ranIndex]
	// <-bucket.kv

	// 从缓存中获取，高性能
	bucket.nodeCacheLock <- true
	if len(bucket.nodeCache) == 0 {
		<-bucket.nodeCacheLock
		return nil
	}
	randIndex := rand.Intn(len(bucket.nodeCache))
	<-bucket.nodeCacheLock
	return bucket.nodeCache[randIndex]
}

// IsNodeExist 检查节点是否存在路由桶里
func (bucket *Bucket) IsNodeExist(n *commonModels.Node) bool {
	// 全量查找
	// bucket.kv <- true
	// for i := 0; i < len(bucket.newBucket); i++ {
	// 	for k := 0; k < len(bucket.newBucket[i]); k++ {
	// 		b := bucket.newBucket[i][k]
	// 		if b == nil {
	// 			continue
	// 		}
	// 		if b.GetNodeID() == n.GetNodeID() {
	// 			<-bucket.kv
	// 			return true
	// 		}
	// 	}
	// }

	// for i := 0; i < len(bucket.triedBucket); i++ {
	// 	for k := 0; k < len(bucket.triedBucket[i]); k++ {
	// 		b := bucket.triedBucket[i][k]
	// 		if b == nil {
	// 			continue
	// 		}
	// 		if b.GetNodeID() == n.GetNodeID() {
	// 			<-bucket.kv
	// 			return true
	// 		}
	// 	}
	// }

	// <-bucket.kv
	// return false

	// 高性能查找
	bucket.nodeCacheLock <- true
	_, bucket.bucketTempIsNodeExist = bucket.nodeCacheHashMap[n.GetNodeID()]
	<-bucket.nodeCacheLock

	if bucket.bucketTempIsNodeExist == true {
		return true
	}

	return false
}
