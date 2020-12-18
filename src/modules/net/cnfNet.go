package net

import (
	// error "github.com/cnf_core/src/utils/error"

	discoverService "github.com/cnf_core/src/modules/net/services/discover"
	nodeBucket "github.com/cnf_core/src/modules/net/services/nodeBucket"
	"github.com/cnf_core/src/utils/config"
	logger "github.com/cnf_core/src/utils/logger"
	"github.com/cnf_core/src/utils/router"
	"github.com/cnf_core/src/utils/sign"

	messageQueue "github.com/cnf_core/src/modules/net/services/messageQueue"
)

func Build() interface{} {
	// 路由桶的初始化
	nodeBucket.Build()

	// 发现服务初始化
	discoverService.Build()

	// 消息队列初始化
	messageQueue.Build()

	return nil
}

/**
 * 运行cnf网络
 */
func Run() interface{} {
	// 初始化所有管道
	discoverChanel := make(chan map[string]string, 5)

	logger.Info("正在启动Cnf网络组件...")

	// 启动发现服务
	go discoverService.Run(discoverChanel)

	// 启动消息队列
	go messageQueue.Run(map[string]chan map[string]string{
		"discoverChanel": discoverChanel,
	})

	privateKeyStr := "fe8ae933d351191288dfcfdd1fd032e384e587a1868e568224974cccd92f0221"

	pubKey := sign.GetPublicKey(privateKeyStr)
	// logger.Debug(pubKey)

	myPublicKey := config.GetNodeId()
	logger.Debug(myPublicKey)

	distance := router.CalculateDistance(myPublicKey, pubKey)
	// distance := router.CalculateDistance("11223455", "22223456")
	logger.Debug(distance)

	// msg := "helloorld11as1asd23sdasdaasda"
	// msgSha256Hash := sha256.Sum256([]byte(msg))
	// msgHash := hex.EncodeToString(msgSha256Hash[:])
	// // // logger.Debug(msgHash)

	// signature, _ := sign.Sign(msgHash, privateKeyStr)
	// logger.Debug(signature)
	// logger.Debug(len(signature))
	// // logger.Debug("===============")
	// // // verified := sign.Verify(signature, msgHash, pubKey)
	// // // logger.Debug(verified)
	// // // logger.Debug("===============")
	// // recoverPublicKey, _ := sign.Recover(signature, msgHash, 0)
	// // logger.Debug(recoverPublicKey)

	return nil
}
