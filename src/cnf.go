package src

import (
	cnfNet "github.com/cnf_core/src/modules/net"
	config "github.com/cnf_core/src/utils/config"
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
	// 首先把配置文件, 全局对象之类的初始化好
	config.SetConfig(conf)
	cnf.conf = conf

	cnf.myPublicChanel = make(map[string]chan map[string]interface{})
	// 初始化多路复用的公共频道
	cnf.myPublicChanel["receiveDiscoverMsgChanel"] = make(chan map[string]interface{}, 5)       // 接收Udp消息，扔这里
	cnf.myPublicChanel["receiveNodeConnectionMsgChanel"] = make(chan map[string]interface{}, 5) // 接收tcp消息，扔这里

	cnf.myPublicChanel["sendDiscoverMsgChanel"] = make(chan map[string]interface{}, 5)       // 要发送udp数据，扔这里
	cnf.myPublicChanel["sendNodeConnectionMsgChanel"] = make(chan map[string]interface{}, 5) // 要发送tcp数据，扔这里

	// 网络层入口构建
	cnf.cnfNet.Build(conf, cnf.myPublicChanel)
}

// GetPublicChanel 获取公共消息chanel
// @return nodeID 节点唯一标识
// @return publicChanel 节点的公共频道
func (cnf *Cnf) GetPublicChanel() (string, map[string]chan map[string]interface{}) {
	return config.ParseNodeID(cnf.conf), cnf.myPublicChanel
}

// Run 主程序入口, 无公共Chanel的实现
func (cnf *Cnf) Run() {
	go cnf.cnfNet.Run()

	// 常驻，监听各个协程的状态
	go cnf.cnfNet.DoLogHTTP()

	// 挂起主协程
	c := make(chan bool)
	d := <-c
	if d {
		return
	}
}

// RunWithPublicChanel 主程序入口, 有公共chanel实现的, 用于端口多路复用
func (cnf *Cnf) RunWithPublicChanel(nodeChanels map[string]interface{}) {
	go cnf.cnfNet.RunWithPublicChanel(nodeChanels)

	// 常驻，监听各个协程的状态
	go cnf.cnfNet.DoLogHTTP()

	// 挂起主协程
	c := make(chan bool)
	d := <-c
	if d {
		return
	}
}
