package bucket

import (
	"math/rand"

	commonModels "github.com/cnf_core/src/modules/net/common/models"
	config "github.com/cnf_core/src/utils/config"
	error "github.com/cnf_core/src/utils/error"
	"github.com/cnf_core/src/utils/router"
)

//CollectSeedFromConf 从配置中获取种子
func (bucket *Bucket) CollectSeedFromConf() {
	netConf := config.GetNetConf()
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

// AddSeed 添加一个节点到种子
// 这里有可能是在初始化阶段就添加，也有可能在收到邻居信息的时候添加
func (bucket *Bucket) AddSeed(node *commonModels.Node) *error.Error {
	// 先检查桶里有没有这个节点
	exist := bucket.IsNodeExist(node)
	if exist {
		return nil
	}
	// 如果超出限制的长度，就先取一个走
	if len(bucket.seed) >= bucket.maxSeedCount {
		bucket.GetSeed()
	}
	bucket.kv <- true
	// 把配置文件里面的种子列表加进来
	bucket.seed = append(bucket.seed, node)
	<-bucket.kv
	return nil
}

// GetSeed 取出一个种子节点，并删除
func (bucket *Bucket) GetSeed() *commonModels.Node {
	if len(bucket.seed) == 0 {
		return nil
	}
	bucket.kv <- true
	node := bucket.seed[0:1]
	bucket.seed = bucket.seed[1:]

	n := node[0]
	<-bucket.kv

	return n
}

// AddNewNode 添加节点到新桶里
func (bucket *Bucket) AddNewNode(n *commonModels.Node) *error.Error {
	if bucket.IsNodeExist(n) {
		return error.New(map[string]interface{}{
			"message": "新节点已存在桶里，无需重复添加",
		})
	}

	bucket.kv <- true
	myNodeID := config.GetNodeID()
	bucketNum := router.CalculateDistance(myNodeID, n.GetNodeID()) // 获得bucket编号
	// 超长的先丢掉，就算不超长，都先把第一个元素丢掉，然后再新增一个新元素进来
	bucket.newBucket[bucketNum] = bucket.newBucket[bucketNum][1:]
	bucket.newBucket[bucketNum] = append(bucket.newBucket[bucketNum], n)
	<-bucket.kv
	return nil
}

// GetRandomNode 获得一个随机节点，给tcp服务进行连接尝试
func (bucket *Bucket) GetRandomNode() *commonModels.Node {
	var b map[int][]*commonModels.Node
	var bucketCount int

	bucket.kv <- true
	// 先随机出一个桶
	bucketRand := rand.Intn(99)
	if bucketRand < 50 {
		b = bucket.newBucket
		bucketCount = NEW_BUCKET_COUNT
	} else {
		b = bucket.triedBucket
		bucketCount = TRIED_BUCKET_COUNT
	}
	nodes := make([]*commonModels.Node, 0)
	for i := 0; i < bucketCount; i++ {
		for k := 0; k < len(b[i]); k++ {
			if b[i][k] != nil {
				nodes = append(nodes, b[i][k])
			}
		}
	}

	if len(nodes) == 0 {
		<-bucket.kv
		return nil
	}

	ranIndex := rand.Intn(len(nodes))
	node := nodes[ranIndex]
	<-bucket.kv
	return node
}

// IsNodeExist 检查节点是否存在路由桶里
func (bucket *Bucket) IsNodeExist(n *commonModels.Node) bool {
	bucket.kv <- true
	for i := 0; i < len(bucket.newBucket); i++ {
		for k := 0; k < len(bucket.newBucket[i]); k++ {
			b := bucket.newBucket[i][k]
			if b == nil {
				continue
			}
			if b.GetNodeID() == n.GetNodeID() {
				<-bucket.kv
				return true
			}
		}
	}

	<-bucket.kv

	return false
}
