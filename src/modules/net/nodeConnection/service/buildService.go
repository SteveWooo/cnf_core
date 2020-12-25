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
	limitTCPInboundConn  chan bool
	limitTCPOutboundConn chan bool

	// 限制数据处理
	limitProcessTCPData chan bool

	// 自己的公共频道
	myPublicChanel map[string]chan map[string]interface{}

	// 自己的内部频道
	myPrivateChanel map[string]chan map[string]interface{}

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
	ncService.limitTCPOutboundConn = make(chan bool, 8)
	ncService.outBoundConn = make([]*nodeConnectionModels.NodeConn, 8)

	// 设置tcp消息读取协程上限(目前没用上，做成高性能tcp服务器需要用到，参考UDP服务部分实现)
	ncService.limitProcessTCPData = make(chan bool, 5)

	logger.Info("NodeConn TCP 服务创建成功，即将监听 " + address)

	return nil
}

// IsOutBoundFull 检查outbound连接是否满了
func (ncService *NodeConnectionService) IsOutBoundFull() bool {
	for i := 0; i < len(ncService.outBoundConn); i++ {
		if ncService.outBoundConn[i] == nil {
			return false
		}
	}

	return true
}

// IsInBoundFull 检查outbound连接是否满了
func (ncService *NodeConnectionService) IsInBoundFull() bool {
	for i := 0; i < len(ncService.inBoundConn); i++ {
		if ncService.inBoundConn[i] == nil {
			return false
		}
	}

	return true
}

// CheckBoundAddress 检查本地是否已经和这个ip:port连接过。
func (ncService *NodeConnectionService) CheckBoundAddress(ip string, servicePort string) bool {
	for i := 0; i < len(ncService.inBoundConn); i++ {
		if ncService.inBoundConn[i] == nil {
			continue
		}

		if ncService.inBoundConn[i].GetSenderIP() == ip && ncService.inBoundConn[i].GetSenderServicePort() == servicePort {
			return true
		}
	}

	for i := 0; i < len(ncService.outBoundConn); i++ {
		if ncService.outBoundConn[i] == nil {
			continue
		}

		if ncService.outBoundConn[i].GetSenderIP() == ip && ncService.outBoundConn[i].GetSenderServicePort() == servicePort {
			return true
		}
	}

	return false
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

// AddOutBoundConn 添加一个InBound连接
func (ncService *NodeConnectionService) AddOutBoundConn(nodeConn *nodeConnectionModels.NodeConn) *error.Error {
	isFull := true
	for i := 0; i < len(ncService.outBoundConn); i++ {
		if ncService.outBoundConn[i] != nil {
			continue
		}
		ncService.outBoundConn[i] = nodeConn
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

// DeleteOutBoundConn 添加一个InBound连接
func (ncService *NodeConnectionService) DeleteOutBoundConn(nodeConn *nodeConnectionModels.NodeConn) *error.Error {
	for i := 0; i < len(ncService.outBoundConn); i++ {
		if ncService.outBoundConn[i] == nil {
			continue
		}

		if ncService.outBoundConn[i].GetNodeConnID() == nodeConn.GetNodeConnID() {
			ncService.outBoundConn[i] = nil
			return nil
		}
	}

	return error.New(map[string]interface{}{
		"message": "找不到需要删除的连接对象",
	})
}

// DeleteInBoundConn 添加一个InBound连接
func (ncService *NodeConnectionService) DeleteInBoundConn(nodeConn *nodeConnectionModels.NodeConn) *error.Error {
	for i := 0; i < len(ncService.inBoundConn); i++ {
		if ncService.inBoundConn[i] == nil {
			continue
		}

		if ncService.inBoundConn[i].GetNodeConnID() == nodeConn.GetNodeConnID() {
			ncService.inBoundConn[i] = nil
			return nil
		}
	}

	return error.New(map[string]interface{}{
		"message": "找不到需要删除的连接对象",
	})
}

// IsBucketExistShakedNode 检查这个nodeID是否被链接上，只查找已经握手成功的，为了防止重复shakeBack
func (ncService *NodeConnectionService) IsBucketExistShakedNode(nodeID string) bool {
	for i := 0; i < len(ncService.inBoundConn); i++ {
		if ncService.inBoundConn[i] == nil {
			continue
		}
		if ncService.inBoundConn[i].IsShaked() == false {
			continue
		}

		if ncService.inBoundConn[i].GetNodeID() == nodeID {
			return true
		}
	}

	for i := 0; i < len(ncService.outBoundConn); i++ {
		if ncService.outBoundConn[i] == nil {
			continue
		}
		if ncService.outBoundConn[i].IsShaked() == false {
			continue
		}

		if ncService.outBoundConn[i].GetNodeID() == nodeID {
			return true
		}
	}

	return false
}

// IsBucketExistUnShakedNode 检查这个nodeID是否被链接上，只查找还没握手成功的。为了防止重复outBound
func (ncService *NodeConnectionService) IsBucketExistUnShakedNode(nodeID string) bool {
	for i := 0; i < len(ncService.inBoundConn); i++ {
		if ncService.inBoundConn[i] == nil {
			continue
		}

		// 只查找还没握手成功的
		if ncService.inBoundConn[i].IsShaked() == true {
			continue
		}

		if ncService.inBoundConn[i].GetNodeID() == nodeID {
			return true
		}
	}

	for i := 0; i < len(ncService.outBoundConn); i++ {
		if ncService.outBoundConn[i] == nil {
			continue
		}

		// 只查找还没握手成功的
		if ncService.outBoundConn[i].IsShaked() == true {
			continue
		}

		if ncService.outBoundConn[i].GetNodeID() == nodeID {
			return true
		}
	}

	return false
}
