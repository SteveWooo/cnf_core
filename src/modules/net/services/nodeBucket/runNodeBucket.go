package nodebucket

import (
	commonModels "github.com/cnf_core/src/modules/net/models/common"
	error "github.com/cnf_core/src/utils/error"
	"github.com/cnf_core/src/utils/timer"
)

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

// HandleBucketOperate 接收数据 主要管理bucket添加节点操作
func HandleBucketOperate(chanel chan map[string]interface{}) {
	for {
		bucketOperate := <-chanel
		node := bucketOperate["node"].(*commonModels.Node)
		if bucketOperate["bucketEvent"] == "addNew" {
			addNewErr := bucket.AddNewNode(node)
			if addNewErr != nil {
				// 说明节点已经存在，不需要重复添加
			}
		}
	}
}

// HandleBucketSeed 推送数据 这里主要是对外输出可用Seed，给节点发现服务用为主
func HandleBucketSeed(chanel chan map[string]interface{}) {
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
func HandleBucketNode(chanel chan map[string]interface{}) {
	for {

	}
}
