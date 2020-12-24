package services

import (
	"net"

	nodeConnectionModels "github.com/cnf_core/src/modules/net/nodeConnection/models"
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

	// 自己的公共频道
	myPublicChanel map[string]chan map[string]interface{}

	inBoundConn  []*nodeConnectionModels.NodeConn
	outBoundConn []*nodeConnectionModels.NodeConn
}

// Build 节点通讯服务的初始化
func (ncService *NodeConnectionService) Build(conf interface{}, myPublicChanel map[string]chan map[string]interface{}) *error.Error {
	ncService.conf = conf
	ncService.myPublicChanel = myPublicChanel

	confNet := conf.(map[string]interface{})["net"]

	// 设置socket地址
	address := confNet.(map[string]interface{})["ip"].(string) + ":" + confNet.(map[string]interface{})["servicePort"].(string)
	ncService.socketAddr = address

	// 设置处理tcp连接创建的协程上限
	ncService.limitTCPInboundConn = make(chan bool, 117)
	ncService.inBoundConn = make([]*nodeConnectionModels.NodeConn, 117)

	// 设置tcp消息读取协程上限
	ncService.limitProcessTCPData = make(chan bool, 5)
	ncService.outBoundConn = make([]*nodeConnectionModels.NodeConn, 5)

	logger.Info("NodeConn TCP 服务创建成功，即将监听 " + address)

	return nil
}

// AddInBoundConn 添加一个InBound连接
func (ncService *NodeConnectionService) AddInBoundConn(nodeConn *nodeConnectionModels.NodeConn) *error.Error {
	isFull := true
	for i := 0; i < len(ncService.inBoundConn); i++ {
		if ncService.inBoundConn[i] != nil {
			continue
		}
		ncService.inBoundConn[i] = nodeConn
		isFull = false
		break
	}
	if isFull {
		return error.New(map[string]interface{}{
			"message": "Inbound 连接数已满",
		})
	}
	return nil
}
