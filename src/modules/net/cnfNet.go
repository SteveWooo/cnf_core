package net

import (
	// error "github.com/cnf_core/src/utils/error"
	// config "github.com/cnf_core/src/utils/config"

	"github.com/cnf_core/src/utils/config"
	"github.com/cnf_core/src/utils/error"

	bucketModel "github.com/cnf_core/src/modules/net/bucket"
	discoverModel "github.com/cnf_core/src/modules/net/discover"
	nodeConnectionModel "github.com/cnf_core/src/modules/net/nodeConnection"
)

// CnfNet 网络层对象
type CnfNet struct {
	conf           interface{}
	nodeConnection nodeConnectionModel.NodeConnection
	discover       discoverModel.Discover
	bucket         bucketModel.Bucket

	// 本节点的公共频道
	myPublicChanel map[string]chan map[string]interface{}

	// 本结点的私有频道
	myPrivateChanel map[string]chan map[string]interface{}

	// 所有节点的公共频道
	publicChanels map[string]interface{}
}

// Build 初始化网络层的各个服务
func (cnfNet *CnfNet) Build(conf interface{}, myPublicChanel map[string]chan map[string]interface{}) *error.Error {
	cnfNet.conf = conf

	// 把自己的公共频道写入构建中
	cnfNet.myPublicChanel = myPublicChanel

	// 路由桶的初始化
	cnfNet.bucket.Build(conf)

	// 发现服务初始化
	cnfNet.discover.Build(conf, cnfNet.myPublicChanel)

	// 节点连接通讯服务的初始化
	cnfNet.nodeConnection.Build(conf, cnfNet.myPublicChanel)

	return nil
}

// Run 运行cnf网络
func (cnfNet *CnfNet) Run(signal chan bool) interface{} {
	// 把公共频道赋值给cnfnet对象
	cnfNet.publicChanels = map[string]interface{}{
		config.ParseNodeID(cnfNet.conf): cnfNet.myPublicChanel,
	}

	// logger.Debug(cnfNet.publicChanels)

	cnfNet.doRun()
	signal <- true

	return nil
}

// RunWithPublicChanel 使用公共频道启动对象, 实现多路复用
func (cnfNet *CnfNet) RunWithPublicChanel(publicChanels map[string]interface{}, signal chan bool) interface{} {
	// 把公共频道赋值给cnfnet对象
	cnfNet.publicChanels = publicChanels

	// logger.Debug(cnfNet.publicChanels)

	cnfNet.doRun()

	signal <- true

	return nil
}

// DoRun 运行的主流程
func (cnfNet *CnfNet) doRun() interface{} {
	// 初始化所有管道
	cnfNet.myPrivateChanel = make(map[string]chan map[string]interface{})
	// 这只有Master节点才用到
	cnfNet.myPrivateChanel["receiveDiscoverMsgChanel"] = make(chan map[string]interface{}, 100)       // 管理udp socket中获取到消息的chanel
	cnfNet.myPrivateChanel["receiveNodeConnectionMsgChanel"] = make(chan map[string]interface{}, 100) // 管理tcp socket中获取到消息的chanel
	// 非master节点都能用到
	cnfNet.myPrivateChanel["bucketOperateChanel"] = make(chan map[string]interface{}, 2) // 一般用于添加bucket节点，或seed
	cnfNet.myPrivateChanel["bucketSeedChanel"] = make(chan map[string]interface{}, 2)    // bucket服务往这个通道输送邻居节点、种子，给doDiscover服务用
	cnfNet.myPrivateChanel["bucketNodeChanel"] = make(chan map[string]interface{}, 2)    // bucket服务往这个通道输送可用节点，给tcp服务尝试连接。

	// 用于多路复用的chanel
	cnfNet.myPrivateChanel["discoverEventChanel"] = make(chan map[string]interface{}, 10) // 发现服务的消息队列

	// logger.Info(config.ParseNodeID(cnfNet.conf) + "正在启动Cnf网络组件...")

	// 启动发现服务
	signal := make(chan bool, 1)
	go cnfNet.discover.RunService(cnfNet.myPrivateChanel, signal)
	<-signal
	// logger.Info("Discover UDP 服务监听成功")

	// 启动tcp数据接收服务
	go cnfNet.nodeConnection.RunService(cnfNet.myPrivateChanel, signal)
	<-signal
	// logger.Info("NodeConn TCP 服务监听成功")

	// 启动路由桶服务，
	// 获取、推送种子节点
	// 处理接收节点信息通道
	go cnfNet.bucket.RunService(cnfNet.myPrivateChanel)

	// 启动消息队列，其实用于各个chanel之间的消息转发
	go cnfNet.RunMessageQueue(cnfNet.myPrivateChanel)

	return nil
}

// DoRunDiscover 开始进行路由
func (cnfNet *CnfNet) DoRunDiscover() interface{} {
	// fmt.Println(cnfNet.conf.(map[string]interface{})["number"].(string)+" (inCnfNet) bucket chanel Addr: ", cnfNet.myPrivateChanel["bucketSeedChanel"])
	// 开始主动寻找新路由节点
	go cnfNet.discover.RunDoDiscover(cnfNet.myPrivateChanel)

	// 开始主动尝试建立连接
	go cnfNet.nodeConnection.RunFindConnection(cnfNet.myPrivateChanel)
	return nil
}
