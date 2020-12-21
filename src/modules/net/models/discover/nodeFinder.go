package discover

import (
	"strconv"

	commonModels "github.com/cnf_core/src/modules/net/models/common"
	"github.com/cnf_core/src/utils/config"
	error "github.com/cnf_core/src/utils/error"
	"github.com/cnf_core/src/utils/logger"
)

// NodeFinder 节点路由寻找器
type NodeFinder struct {
	seed         []*commonModels.Node
	maxSeedCount int
}

// Build 初始化节点寻找器
func (nodeFinder *NodeFinder) Build(prop map[string]interface{}) *error.Error {
	netConf := config.GetNetConf()
	maxSeedCount, maxCountErr := strconv.Atoi(netConf.(map[string]interface{})["maxSeedCount"].(string))
	if maxCountErr != nil {
		logger.Warn("config miss: maxSeedCount. Using default.")
		nodeFinder.maxSeedCount = 100
	}
	nodeFinder.maxSeedCount = maxSeedCount

	nodeFinder.CollectSeedFromConf()

	return nil
}

//CollectSeedFromConf 从配置中获取种子
func (nodeFinder *NodeFinder) CollectSeedFromConf() {
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
		nodeFinder.AddSeed(node)
	}
}

// AddSeed 添加一个节点到种子
// 这里有可能是在初始化阶段就添加，也有可能在收到邻居信息的时候添加
func (nodeFinder *NodeFinder) AddSeed(node *commonModels.Node) *error.Error {
	// 如果超出限制的长度，就先取一个走
	if len(nodeFinder.seed) >= nodeFinder.maxSeedCount {
		nodeFinder.GetSeed()
	}
	// 把配置文件里面的种子列表加进来
	nodeFinder.seed = append(nodeFinder.seed, node)
	return nil
}

// GetSeed 取出一个种子节点，并删除
func (nodeFinder *NodeFinder) GetSeed() *commonModels.Node {
	if len(nodeFinder.seed) == 0 {
		return nil
	}
	node := nodeFinder.seed[0:1]
	nodeFinder.seed = nodeFinder.seed[1:]

	return node[0]
}
