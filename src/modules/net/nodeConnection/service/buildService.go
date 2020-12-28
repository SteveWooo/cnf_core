package services

import (
	"net"

	commonModels "github.com/cnf_core/src/modules/net/common/models"
	nodeConnectionModels "github.com/cnf_core/src/modules/net/nodeConnection/models"
	"github.com/cnf_core/src/utils/error"
)

var INBOUND_CONN_MAX int = 117
var OUTBOUND_CONN_MAX int = 8

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

	// 子节点的内外连接（不带socket
	inBoundConn  []*nodeConnectionModels.NodeConn
	outBoundConn []*nodeConnectionModels.NodeConn

	// 统一socket存放处。需要就在这里拿
	masterOutBoundSocket []*nodeConnectionModels.NodeConn
	masterInBoundSocket  []*nodeConnectionModels.NodeConn

	// 主节点的内外连接（带socket），用目标nodeID索引
	masterInBoundConn  map[string][]*nodeConnectionModels.NodeConn
	masterOutBoundConn map[string][]*nodeConnectionModels.NodeConn
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
	ncService.inBoundConn = make([]*nodeConnectionModels.NodeConn, INBOUND_CONN_MAX)
	ncService.outBoundConn = make([]*nodeConnectionModels.NodeConn, OUTBOUND_CONN_MAX)

	// 创建统一socket
	ncService.limitTCPInboundConn = make(chan bool, 10240)
	ncService.limitTCPOutboundConn = make(chan bool, 10240)
	ncService.masterOutBoundSocket = make([]*nodeConnectionModels.NodeConn, 10240)
	ncService.masterInBoundSocket = make([]*nodeConnectionModels.NodeConn, 10240)

	// 设置tcp消息读取协程上限(目前没用上，做成高性能tcp服务器需要用到，参考UDP服务部分实现)
	ncService.limitProcessTCPData = make(chan bool, 5)

	// logger.Info("NodeConn TCP 服务创建成功，即将监听 " + address)

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

// GetOutBoundSocket 根据目标节点，创建一个可用的socket
// @param newNode 目标节点，主要为了获取IP port
// @param targetNodeID 子节点ID
// @return nodeConn 带socket的nodeConn对象
// @return isNew 是否新对象，如果是新socket，要读数据
// @return error 错误
func (ncService *NodeConnectionService) GetOutBoundSocket(newNode *commonModels.Node, targetNodeID string) (*nodeConnectionModels.NodeConn, bool, *error.Error) {
	for i := 0; i < len(ncService.masterOutBoundSocket); i++ {
		if ncService.masterOutBoundSocket[i] == nil {
			continue
		}

		if newNode.GetIP() == ncService.masterOutBoundSocket[i].GetSenderIP() && newNode.GetServicePort() == ncService.masterOutBoundSocket[i].GetSenderServicePort() {
			var nodeConn nodeConnectionModels.NodeConn
			// 复用socket
			nodeConn.Build(ncService.masterOutBoundSocket[i].GetSocket(), "outBound")
			nodeConn.SetRemoteAddr(ncService.masterOutBoundSocket[i].GetSenderIP() + ":" + ncService.masterOutBoundSocket[i].GetSenderServicePort())
			// 设置新的目标节点nodeID
			nodeConn.SetNodeID(newNode.GetNodeID())
			// 设置子节点的ID，作为自己的唯一标识
			nodeConn.SetTargetNodeID(targetNodeID)

			return &nodeConn, false, nil
		}
	}

	// 没有可以复用的，就创建一个。创建之前要先找个空位
	for i := 0; i < len(ncService.masterOutBoundSocket); i++ {
		if ncService.masterOutBoundSocket[i] != nil {
			continue
		}
		nodeAddress := newNode.GetIP() + ":" + newNode.GetServicePort()
		conn, connErr := net.Dial("tcp", nodeAddress)
		if connErr != nil {
			return nil, false, error.New(map[string]interface{}{
				"message": "主动创建tcp连接失败",
			})
		}
		remoteAddr := conn.RemoteAddr()
		var nodeConn nodeConnectionModels.NodeConn
		nodeConn.Build(&conn, "outBound")
		nodeConn.SetRemoteAddr(remoteAddr.String())
		// 由于是主动发起连接的，所以要设置nodeID
		nodeConn.SetNodeID(newNode.GetNodeID())
		// 同时设置子节点的targetNodeID
		nodeConn.SetTargetNodeID(targetNodeID)

		// 放入到主socket池中
		ncService.masterOutBoundSocket[i] = &nodeConn

		return &nodeConn, true, nil
	}

	return nil, false, error.New(map[string]interface{}{
		"message": "socket已满，不可创建新的outBoundSocket",
	})
}

// AddInBoundSocket 添加一个被链接的socket, 被连接的socket还是多多益善吧
func (ncService *NodeConnectionService) AddInBoundSocket(nodeConn *nodeConnectionModels.NodeConn) *error.Error {
	for i := 0; i < len(ncService.masterInBoundSocket); i++ {
		if ncService.masterInBoundSocket[i] != nil {
			continue
		}
		ncService.masterInBoundSocket[i] = nodeConn
		return nil
	}
	return error.New(map[string]interface{}{
		"message": "对内连接socket已经用光了",
	})
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
			"message": "outbound 连接数已满",
		})
	}
	return nil
}

// AddMasterOutBoundConn 添加一个InBound连接
func (ncService *NodeConnectionService) AddMasterOutBoundConn(nodeConn *nodeConnectionModels.NodeConn) *error.Error {
	// 先检查这个本地子节点是否已经创建了
	targetNodeID := nodeConn.GetTargetNodeID()
	if ncService.masterOutBoundConn[targetNodeID] == nil {
		ncService.masterOutBoundConn[targetNodeID] = make([]*nodeConnectionModels.NodeConn, OUTBOUND_CONN_MAX)
	}

	isFull := true
	for i := 0; i < len(ncService.masterOutBoundConn[targetNodeID]); i++ {
		if ncService.masterOutBoundConn[targetNodeID][i] != nil {
			continue
		}
		ncService.masterOutBoundConn[targetNodeID][i] = nodeConn
		isFull = false
		break
	}
	if isFull {
		return error.New(map[string]interface{}{
			"message": "outbound 连接数已满",
		})
	}
	return nil
}

// AddMasterInBoundConn 添加一个InBound连接
func (ncService *NodeConnectionService) AddMasterInBoundConn(nodeConn *nodeConnectionModels.NodeConn) *error.Error {
	// 先检查这个本地子节点是否已经创建了
	targetNodeID := nodeConn.GetTargetNodeID()
	if ncService.masterInBoundConn[targetNodeID] == nil {
		ncService.masterInBoundConn[targetNodeID] = make([]*nodeConnectionModels.NodeConn, INBOUND_CONN_MAX)
	}

	isFull := true
	for i := 0; i < len(ncService.masterInBoundConn[targetNodeID]); i++ {
		if ncService.masterInBoundConn[targetNodeID][i] != nil {
			continue
		}
		ncService.masterInBoundConn[targetNodeID][i] = nodeConn
		isFull = false
		break
	}
	if isFull {
		return error.New(map[string]interface{}{
			"message": "outbound 连接数已满",
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

// DeleteMasterInBoundConn 添加一个InBound连接
func (ncService *NodeConnectionService) DeleteMasterInBoundConn(nodeConn *nodeConnectionModels.NodeConn) *error.Error {
	targetNodeID := nodeConn.GetTargetNodeID()
	if ncService.masterInBoundConn[targetNodeID] == nil {
		return error.New(map[string]interface{}{
			"message": "需要删除的conn的sub节点targetNodeID不存在",
		})
	}

	for i := 0; i < len(ncService.masterInBoundConn[targetNodeID]); i++ {
		if ncService.masterInBoundConn[targetNodeID][i] == nil {
			continue
		}

		if ncService.masterInBoundConn[targetNodeID][i].GetNodeConnID() == nodeConn.GetNodeConnID() {
			ncService.masterInBoundConn[targetNodeID][i] = nil
			return nil
		}
	}

	return error.New(map[string]interface{}{
		"message": "找不到需要删除的连接对象",
	})
}

// DeleteMasterOutBoundConn 添加一个InBound连接
func (ncService *NodeConnectionService) DeleteMasterOutBoundConn(nodeConn *nodeConnectionModels.NodeConn) *error.Error {
	targetNodeID := nodeConn.GetTargetNodeID()
	if ncService.masterOutBoundConn[targetNodeID] == nil {
		return error.New(map[string]interface{}{
			"message": "需要删除的outbound conn的sub节点targetNodeID不存在",
		})
	}

	for i := 0; i < len(ncService.masterOutBoundConn[targetNodeID]); i++ {
		if ncService.masterOutBoundConn[targetNodeID][i] == nil {
			continue
		}

		if ncService.masterOutBoundConn[targetNodeID][i].GetNodeConnID() == nodeConn.GetNodeConnID() {
			ncService.masterOutBoundConn[targetNodeID][i] = nil
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
