package src

import (
	"strconv"

	cnfNet "github.com/cnf_core/src/modules/net"
	config "github.com/cnf_core/src/utils/config"
	"github.com/cnf_core/src/utils/logger"
	"github.com/cnf_core/src/utils/sign"
	// logger "github.com/cnf_core/src/utils/logger"
)

// Cnf 整体对象
type Cnf struct {
	conf   interface{}
	cnfNet cnfNet.CnfNet

	// 给其他协程使用, 配合配置中net.masterServer参数, 实现端口多路复用
	myPublicChanel map[string]chan map[string]interface{}
}

// Build 主程序配置入口
func (cnf *Cnf) Build(conf interface{}) {
	// 检查是否有publicKey，没有的话，就调用ecc算法，根据privatekey生成
	confNet := conf.(map[string]interface{})["net"]
	if confNet.(map[string]interface{})["publicKey"] == nil {
		confNet.(map[string]interface{})["publicKey"] = sign.GetPublicKey(confNet.(map[string]interface{})["localPrivateKey"].(string))
		conf.(map[string]interface{})["net"] = confNet
	}

	// 首先把配置文件, 全局对象之类的初始化好
	config.SetConfig(conf)
	cnf.conf = conf

	// 初始化公共频道，公共频道的缓存长度取决于有多少个节点需要共用一个端口

	cnf.myPublicChanel = make(map[string]chan map[string]interface{})
	if confNet.(map[string]interface{})["publicChanelLength"] != nil {
		chanelLength, convErr := strconv.Atoi(confNet.(map[string]interface{})["publicChanelLength"].(string))
		if convErr != nil {
			logger.Error("配置文件中，publicChanelLength必须是整数。已使用默认参数")
			cnf.initPublicChanel(1000) // 默认1000
		} else {
			cnf.initPublicChanel(chanelLength)
		}
	} else {
		cnf.initPublicChanel(1000) // 默认1000
	}

	// 网络层入口构建
	cnf.cnfNet.Build(conf, cnf.myPublicChanel)
}

// GetPublicChanel 获取公共消息chanel
// @return nodeID 节点唯一标识
// @return publicChanel 节点的公共频道
func (cnf *Cnf) GetPublicChanel() (string, map[string]chan map[string]interface{}) {
	return config.ParseNodeID(cnf.conf), cnf.myPublicChanel
}

// initPublicChanel 初始化本子节点的公共管道
// 一般根据publicChanel（子节点数）来决定管道数量
func (cnf *Cnf) initPublicChanel(chanelLength int) {
	// 初始化多路复用的公共频道
	/// ⭐应该只有主节点需要那么多条通道，普通节点不用那么多也行

	// 关于udp数据包的
	cnf.myPublicChanel["receiveDiscoverMsgChanel"] = make(chan map[string]interface{}, chanelLength) // 接收Udp消息，主节点把消息扔这里
	cnf.myPublicChanel["sendDiscoverMsgChanel"] = make(chan map[string]interface{}, chanelLength)    // 要发送udp数据，扔这里

	// 关于tcp连接创建的
	cnf.myPublicChanel["submitNodeConnectionCreateChanel"] = make(chan map[string]interface{}, chanelLength)  // 子节点需要创建TCP连接的话，子节点就往这里扔一个请求
	cnf.myPublicChanel["receiveNodeConnectionCreateChanel"] = make(chan map[string]interface{}, chanelLength) // 主节点创建连接成功，并握手成功的，就把conn对象扔回这个chanel里面

	// 这个缓存一旦变小，就会卡死
	cnf.myPublicChanel["receiveNodeConnectionMsgChanel"] = make(chan map[string]interface{}, chanelLength) // 接收到tcp消息，主节点把消息扔这里
	cnf.myPublicChanel["sendNodeConnectionMsgChanel"] = make(chan map[string]interface{}, chanelLength)    // 子节点发送tcp消息的

	// 给主进程用的日志通道
	cnf.myPublicChanel["logChanel"] = make(chan map[string]interface{}, 10)
}

// Run 主程序入口, 无公共Chanel的实现
func (cnf *Cnf) Run(signal chan bool) {
	sign := make(chan bool, 1)
	go cnf.cnfNet.Run(sign)
	<-sign

	// 常驻，监听各个协程的状态
	// go cnf.cnfNet.DoLogHTTP()
	// go cnf.cnfNet.DoLogUDP()
	go cnf.cnfNet.DoLogChanel()

	if signal != nil {
		signal <- true
	}

	// 挂起主协程
	c := make(chan bool)
	d := <-c
	if d {
		return
	}
}

// RunWithPublicChanel 主程序入口, 有公共chanel实现的, 用于端口多路复用
func (cnf *Cnf) RunWithPublicChanel(nodeChanels map[string]interface{}, signal chan bool) {
	// 防止chanel全局变量未写入就启动了程序
	sign := make(chan bool, 1)
	go cnf.cnfNet.RunWithPublicChanel(nodeChanels, sign)
	<-sign

	// 常驻，监听各个协程的状态
	// go cnf.cnfNet.DoLogHTTP()
	// go cnf.cnfNet.DoLogUDP()
	go cnf.cnfNet.DoLogChanel()

	if signal != nil {
		signal <- true
	}

	// 挂起主协程
	c := make(chan bool)
	d := <-c
	if d {
		return
	}
}

// DoRunDiscover 反射
func (cnf *Cnf) DoRunDiscover() {
	cnf.cnfNet.DoRunDiscover()
}
