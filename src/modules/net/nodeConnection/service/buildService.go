package services

import (
	"net"

	"github.com/cnf_core/src/utils/error"
	"github.com/cnf_core/src/utils/logger"
)

// NodeConnectionService 节点连接服务的tcp handle
type NodeConnectionService struct {
	conf        interface{}
	socketAddr  string
	tcpListener net.Listener

	// 限制连接创立事件
	limitTCPInboundConn chan bool

	// 限制数据处理
	limitProcessTCPData chan bool
}

// Build 节点通讯服务的初始化
func (ncService *NodeConnectionService) Build(conf interface{}) *error.Error {
	ncService.conf = conf

	confNet := conf.(map[string]interface{})["net"]

	// 设置socket地址
	address := confNet.(map[string]interface{})["ip"].(string) + ":" + confNet.(map[string]interface{})["servicePort"].(string)
	ncService.socketAddr = address

	// 设置处理tcp连接创建的协程上限
	ncService.limitTCPInboundConn = make(chan bool, 117)

	// 设置tcp消息读取协程上限
	ncService.limitProcessTCPData = make(chan bool, 5)

	logger.Info("NodeConn TCP 服务创建成功，即将监听 " + address)

	return nil
}
