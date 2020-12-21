package nodebucket

import (
	discoverModel "github.com/cnf_core/src/modules/net/models/discover"
	"github.com/cnf_core/src/utils/config"
	"github.com/cnf_core/src/utils/error"
	"github.com/cnf_core/src/utils/logger"
	"github.com/cnf_core/src/utils/timer"
)

// 管理可用节点的路由桶
var bucket discoverModel.Bucket

// 管理种子节点的路由桶（即未握手节点）
var nodeFinder discoverModel.NodeFinder

// Build 为了初始化全局变量bucket
func Build() {
	bucket.Build()

	nodeFinder.Build(map[string]interface{}{
		"seeds": config.GetNetSeed(),
	})
}

// RunService 负责接收Bucket chanel的消息，并持续处理
// 可能添加new tried节点路由，也可能是添加邻居路由
func RunService(chanels map[string]chan map[string]interface{}) *error.Error {
	// 接收的
	go HandleBucketOperate(chanels["bucketOperateChanel"])
	go HandleBucketNode(chanels["bucketNodeChanel"])

	// 往外输出的
	go HandleBucketSeed(chanels["bucketSeedChanel"])

	return nil
}

// HandleBucketOperate 主要管理bucket添加节点操作
func HandleBucketOperate(chanel chan map[string]interface{}) {
	for {
		bucketOperate := <-chanel
		logger.Debug(bucketOperate)
	}
}

// HandleBucketSeed 这里主要是对外输出可用Seed，给节点发现服务用为主
func HandleBucketSeed(chanel chan map[string]interface{}) {
	for {
		seed := nodeFinder.GetSeed()
		// 没有种子，就消停会，再从配置里面获取新的种子
		if seed == nil {
			timer.Sleep(1000)
			nodeFinder.CollectSeedFromConf()
			continue
		}

		chanel <- map[string]interface{}{
			"node": seed,
		}
	}
}

// HandleBucketNode 主要对外输出可用Node，给节点连接服务为主
func HandleBucketNode(chanel chan map[string]interface{}) {
	for {

	}
}
