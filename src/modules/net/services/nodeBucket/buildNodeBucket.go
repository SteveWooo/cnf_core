package nodeBucket

import (
	models "github.com/cnf_core/src/modules/net/models"
	// error "github.com/cnf_core/src/utils/error"
	// logger "github.com/cnf_core/src/utils/logger"
)

// 配置项
var NEW_BUCKET_COUNT int = 16
var NEW_BUCKET_LENGTH int = 64
var TRIED_BUCKET_COUNT int = 16
var TRIED_BUCKET_LENGTH int = 16

/**
 * 重要的全局变量, newBucket负责存放未建立过连接的节点信息. triedBucket存放建立过连接的节点信息
 */
var newBucket map[int]interface{}
var triedBucket map[int]interface{}

func Build(){
	newBucket = make(map[int]interface{}) // 初始化数组对象本身
	for i:=0; i<NEW_BUCKET_COUNT;i++ {
		newBucket[i] = make(map[int]interface{}) // 声明第一个维度每个对象都是一个子数组
	}

	triedBucket = make(map[int]interface{}) // 初始化数组对象本身
	for i:=0; i<TRIED_BUCKET_COUNT;i++ {
		triedBucket[i] = make(map[int]interface{}) // 声明第一个维度每个对象都是一个子数组
	}
}

/**
 * 添加一个新节点到new桶里, 需要用kad算法分配具体位置
 */
func AddNodeToNewBucket (node models.Node) interface{} {
	// TODO: 按照kad算法把节点加入到桶中. 这里先默认加入到第一个桶的第一个位置里
	newBucket[0].(map[int]interface{})[0] = node
	return nil
}