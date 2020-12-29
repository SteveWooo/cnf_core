package services

import (
	"math/rand"
	"net"

	commonModels "github.com/cnf_core/src/modules/net/common/models"
	nodeConnectionModels "github.com/cnf_core/src/modules/net/nodeConnection/models"
	"github.com/cnf_core/src/utils/config"
	"github.com/cnf_core/src/utils/error"
	"github.com/cnf_core/src/utils/logger"
	"github.com/cnf_core/src/utils/timer"
	// commonModels "github.com/cnf_core/src/modules/net/common/models"
)

// RunFindConnection 主动找可用节点
func (ncService *NodeConnectionService) RunFindConnection(chanels map[string]chan map[string]interface{}) {
	go ncService.DoFindConnection(chanels)
}

// DoFindConnection 主动找可用节点
func (ncService *NodeConnectionService) DoFindConnection(chanels map[string]chan map[string]interface{}) {
	for {
		timer.Sleep(3000 + rand.Intn(3000))
		// continue
		// timer.Sleep(1000)
		// 如果outbound无空位，则不需要进行尝试连接
		if ncService.IsOutBoundFull() == true {
			continue
		}

		nodeChanelMsg := <-chanels["bucketNodeChanel"]
		newNode := nodeChanelMsg["node"].(*commonModels.Node)
		// logger.Debug(config.ParseNodeID(ncService.conf) + " bucket : ")
		// logger.Debug(ncService.inBoundConn)
		// logger.Debug(ncService.outBoundConn)

		// 不要和自己连接
		if newNode.GetNodeID() == config.ParseNodeID(ncService.conf) {
			continue
		}

		// 已经握手成功的节点，不需要重复建立连接
		if ncService.IsBucketExistUnShakedNode(newNode.GetNodeID()) || ncService.IsBucketExistShakedNode(newNode.GetNodeID()) {
			continue
		}

		// 端口多路复用逻辑 Master节点会调用以下函数 MasterDoTryOutBoundConnect
		ncService.myPublicChanel["submitNodeConnectionCreateChanel"] <- map[string]interface{}{
			"newNode":         newNode, // 包含了ip:port:nodeID
			"targetNodeID":    config.ParseNodeID(ncService.conf),
			"shakePackString": ncService.GetShakePackString("outBound", newNode.GetNodeID()),
		}

		// 非端口多路复用逻辑：
		// tryOubountConnErr := ncService.DoTryOutBoundConnect(newNode)
		// if tryOubountConnErr != nil {
		// 	logger.Warn(tryOubountConnErr)
		// }
	}
}

// MasterDoTryOutBoundConnect master节点尝试建立连接，并创建一个NodeConn返回（端口多路复用函数
// @param newNode 目标连接节点IP
// @param targetNodeID 发起连接的本地节点ip
func (ncService *NodeConnectionService) MasterDoTryOutBoundConnect(newNode *commonModels.Node, targetNodeID string) (map[string]interface{}, *error.Error) {
	// 如果是本地连本地，那就返回一个空socket的nodeConn回去即可
	cnfNet := ncService.conf.(map[string]interface{})["net"]
	if newNode.GetIP() == cnfNet.(map[string]interface{})["ip"] && newNode.GetServicePort() == cnfNet.(map[string]interface{})["servicePort"] {
		// logger.Debug("本地连接创立！to: " + newNode.GetIP() + ":" + newNode.GetServicePort())
		var nodeConn nodeConnectionModels.NodeConn
		nodeConn.Build(nil, "outBound")
		nodeConn.SetRemoteAddr(newNode.GetIP() + ":" + newNode.GetServicePort())
		// 由于是主动发起连接的，所以要设置nodeID
		nodeConn.SetNodeID(newNode.GetNodeID())
		// 同时设置子节点的targetNodeID
		nodeConn.SetTargetNodeID(targetNodeID)

		return map[string]interface{}{
			"nodeConn": &nodeConn,
			"connType": "outBound",
			"newNode":  newNode,
		}, nil
	}

	// 如果不是本地节点，就需要创建一个socket
	if newNode.GetIP() != cnfNet.(map[string]interface{})["ip"] || newNode.GetServicePort() != cnfNet.(map[string]interface{})["servicePort"] {
		// outBoundSocket是一个nodeConn对象，但是只为了获取socekt几
		outBoundSocket, isNew, getSocketErr := ncService.GetOutBoundSocket(newNode, targetNodeID)
		if getSocketErr != nil {
			return nil, getSocketErr
		}

		var nodeConn nodeConnectionModels.NodeConn
		nodeConn.Build(outBoundSocket.GetSocket(), "outBound")
		nodeConn.SetRemoteAddr((*outBoundSocket.GetSocket()).RemoteAddr().String())
		// 由于是主动发起连接的，所以要设置nodeID
		nodeConn.SetNodeID(newNode.GetNodeID())
		// 同时设置子节点的targetNodeID
		nodeConn.SetTargetNodeID(targetNodeID)

		if isNew == true {
			// 处理这个conn
			// logger.Debug("is new socket")
			go ncService.MasterProcessOutboundTCPData(&nodeConn)
		} else {
			// logger.Debug("not new socket")
		}

		// 然后把这个连接添加到Master桶里。注意，这里是根据targetNodeID做区分的。所以targetNodeID很重要
		// ⭐targetNodeID其实就是本地子节点的ID，标识为targetNodeID，是为了别人发消息的时候，能锁定节点位置
		// addConnErr := ncService.AddMasterOutBoundConn(&nodeConn)
		// if addConnErr != nil {
		// 	return nil, addConnErr
		// }
		// 不需要发送握手包

		return map[string]interface{}{
			"nodeConn": &nodeConn,
			"connType": "outBound",
			"newNode":  newNode,
		}, nil
	}

	return nil, error.New(map[string]interface{}{
		"message":  "逻辑黑洞",
		"connType": "outBound",
		"newNode":  newNode,
	})
}

// MasterProcessOutboundTCPData 狂读TCP socket（端口多路复用函数
func (ncService *NodeConnectionService) MasterProcessOutboundTCPData(nodeConn *nodeConnectionModels.NodeConn) {
	chanel := ncService.myPrivateChanel["receiveNodeConnectionMsgChanel"] // 主节点的私有频道
	for {
		tcpSourceDataByte := make([]byte, 1024)

		length, readErr := (*nodeConn.Socket).Read(tcpSourceDataByte)
		if readErr != nil {
			// 读取数据失败，说明socket已经断掉，所以要结束这个socket
			(*nodeConn.Socket).Close()
			ncService.DeleteMasterOutBoundConn(nodeConn)

			// TODO 通知子节点删除这个outBound连接

			// logger.Debug(config.ParseNodeID(ncService.conf) + " outBound relive")
			// logger.Debug(readErr)
			return
		}

		tcpSourceData := string(tcpSourceDataByte[:length])

		tcpData, parseErr := ncService.ParseTCPData(tcpSourceData)
		if parseErr != nil {
			logger.Error(parseErr.GetMessage())
			continue
		}

		chanel <- map[string]interface{}{
			"nodeConn":    nodeConn,
			"tcpData":     tcpData.(map[string]interface{}),
			"messageFrom": "outBound",
		}
	}
}

// SalveHandleNodeOutBoundConnectionCreateEvent 子节点处理节点创建成功事件。
func (ncService *NodeConnectionService) SalveHandleNodeOutBoundConnectionCreateEvent(nodeConnectionCreateResp map[string]interface{}) {
	// logger.Debug(nodeConnectionCreateResp)
	nodeConn := nodeConnectionCreateResp["nodeConn"].(*nodeConnectionModels.NodeConn)
	addConnErr := ncService.AddOutBoundConn(nodeConn)

	if addConnErr != nil {
		// logger.Error("子节点中，添加节点到outbound失败")
		return
	}

	// ⭐停顿一个随机时间再发送shakeEvent，尽可能防止双向连接（当然真的建立了双向链接，目前也没什么好的解决方法
	timer.Sleep(rand.Intn(1000))

	// 发送握手请求
	shakePackageString := ncService.GetShakePackString("shakeEvent", nodeConn.GetNodeID())
	ncService.myPublicChanel["sendNodeConnectionMsgChanel"] <- map[string]interface{}{
		"nodeConn": nodeConn,
		"message":  shakePackageString,
	}
	nodeConn.SetShaker(nil)
}

// DoTryOutBoundConnect 尝试进行主动连接（非端口多路复用函数
func (ncService *NodeConnectionService) DoTryOutBoundConnect(node *commonModels.Node) *error.Error {
	nodeAddress := node.GetIP() + ":" + node.GetServicePort()
	conn, connErr := net.Dial("tcp", nodeAddress)

	if connErr != nil {
		return error.New(map[string]interface{}{
			"message": "主动创建tcp连接失败",
		})
	}

	// 先处理好这个conn，再去处理他收到的消息
	remoteAddr := conn.RemoteAddr()
	var nodeConn nodeConnectionModels.NodeConn
	nodeConn.Build(&conn, "outBound")
	nodeConn.SetRemoteAddr(remoteAddr.String())
	// 由于是主动发起连接的，所以要设置nodeID
	nodeConn.SetNodeID(node.GetNodeID())

	// 添加一个未握手的连接到OutBoundConn里面去
	addConnErr := ncService.AddOutBoundConn(&nodeConn)
	if addConnErr != nil {
		return addConnErr
	}

	// 由于是outbound，所以连接成功就马上发送一个握手包
	ncService.SendShake(&nodeConn, node.GetNodeID())

	go ncService.ProcessOutboundTCPData(&nodeConn)
	return nil
}

// ProcessOutboundTCPData 狂读TCP socket
func (ncService *NodeConnectionService) ProcessOutboundTCPData(nodeConn *nodeConnectionModels.NodeConn) {
	chanel := ncService.myPrivateChanel["receiveNodeConnectionMsgChanel"]
	for {
		tcpSourceDataByte := make([]byte, 1024)

		length, readErr := (*nodeConn.Socket).Read(tcpSourceDataByte)
		if readErr != nil {
			// 读取数据失败，说明socket已经断掉，所以要结束这个socket
			(*nodeConn.Socket).Close()
			ncService.DeleteOutBoundConn(nodeConn)

			// 释放一个outBound限制
			<-ncService.limitTCPOutboundConn
			// logger.Debug(config.ParseNodeID(ncService.conf) + " outBound relive")
			// logger.Debug(readErr)
			return
		}

		tcpSourceData := string(tcpSourceDataByte[:length])

		tcpData, parseErr := ncService.ParseTCPData(tcpSourceData)
		if parseErr != nil {
			logger.Error(parseErr.GetMessage())
			continue
		}

		chanel <- map[string]interface{}{
			"nodeConn":    nodeConn,
			"tcpData":     tcpData.(map[string]interface{}),
			"messageFrom": "outBound",
		}
	}
}
