package services

import (
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
		timer.Sleep(2000)
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

		// 不需要建立相同的tcp连接。
		// logger.Debug(config.ParseNodeID(ncService.conf) + " trying : " + newNode.GetIP() + ":" + newNode.GetServicePort())
		if ncService.CheckBoundAddress(newNode.GetIP(), newNode.GetServicePort()) == true {
			continue
		}

		tryOubountConnErr := ncService.DoTryOutBoundConnect(newNode)
		if tryOubountConnErr != nil {
			logger.Warn(tryOubountConnErr)
		}
	}
}

// DoTryOutBoundConnect 尝试进行主动连接
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
	nodeConn.Build(conn, "outBound")
	nodeConn.SetRemoteAddr(remoteAddr.String())
	// 由于是主动发起连接的，所以要设置nodeID
	nodeConn.SetNodeID(node.GetNodeID())

	// 添加一个未握手的连接到InBoundConn里面去
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

		length, readErr := nodeConn.Socket.Read(tcpSourceDataByte)
		if readErr != nil {
			// 读取数据失败，说明socket已经断掉，所以要结束这个socket
			nodeConn.Socket.Close()
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
