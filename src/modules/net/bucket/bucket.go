package bucket

import (
	commonModels "github.com/cnf_core/src/modules/net/common/models"
	error "github.com/cnf_core/src/utils/error"
	"github.com/cnf_core/src/utils/timer"
)

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
			// logger.Debug(addNewErr)
			// 说明节点已经存在，不需要重复添加
		}
	}
}

// HandleBucketSeed 推送数据 这里主要是对外输出可用Seed，给节点发现服务用为主
func (bucket *Bucket) HandleBucketSeed(chanel chan map[string]interface{}) {
	for {
		seed := bucket.GetSeed()
		// 没有种子，就从配置里面获取新的种子
		if seed == nil {
			timer.Sleep(1000)
			bucket.CollectSeedFromConf()
			continue
		}

		// 检查桶里是否已经有这个Node了，有的话就不推了
		if bucket.IsNodeExist(seed) == true {
			// logger.Debug(bucket.conf.(map[string]interface{})["number"].(string) + " newNode is exists")
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
			timer.Sleep(2000)
			continue
		}
		chanel <- map[string]interface{}{
			"node": node,
		}
	}
}
