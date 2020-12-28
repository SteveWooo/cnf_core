package services

import (
	"encoding/base64"
	"encoding/json"
	"net"
	"strings"

	"github.com/cnf_core/src/utils/error"
	"github.com/cnf_core/src/utils/logger"

	nodeConnectionModels "github.com/cnf_core/src/modules/net/nodeConnection/models"
)

// RunService 启动节点通信的TCP相关服务
func (ncService *NodeConnectionService) RunService(chanels map[string]chan map[string]interface{}, signal chan bool) *error.Error {
	ncService.myPrivateChanel = chanels
	// 非主节点，不需要监听socket
	confNet := ncService.conf.(map[string]interface{})["net"]
	if confNet.(map[string]interface{})["masterServer"] != "true" {
		// logger.Debug("非主节点，无需监听")
		signal <- true
		return nil
	}

	tcpListener, listenErr := net.Listen("tcp", ncService.socketAddr)
	// logger.Debug("listen: " + ncService.socketAddr)
	if listenErr != nil {
		return error.New(map[string]interface{}{
			"message":   "监听UDP端口失败",
			"originErr": listenErr,
		})
	}

	ncService.tcpListener = tcpListener
	signal <- true

	for {
		ncService.limitTCPInboundConn <- true

		conn, acceptErr := ncService.tcpListener.Accept()
		if acceptErr != nil {
			logger.Error("被动连接发生错误")
			<-ncService.limitTCPInboundConn
			continue
		}

		// 先处理好这个conn，再去处理他收到的消息
		remoteAddr := conn.RemoteAddr()
		var nodeConn nodeConnectionModels.NodeConn
		nodeConn.Build(&conn, "inBound")
		nodeConn.SetRemoteAddr(remoteAddr.String())

		// 添加一个未握手的连接到InBoundConn里面去
		addConnErr := ncService.AddInBoundSocket(&nodeConn)
		if addConnErr != nil {
			// 连接失败就要关闭掉这条socket
			(*nodeConn.Socket).Close()
			continue
		}

		// logger.Debug("create a new inBound")
		go ncService.ProcessInboundTCPData(&nodeConn)
	}
}

// SalveHandleNodeInBoundConnectionCreateEvent 子节点处理节点创建成功事件。（但是不知道应该给哪个子节点）
func (ncService *NodeConnectionService) SalveHandleNodeInBoundConnectionCreateEvent(nodeConnectionCreateResp map[string]interface{}) {
	nodeConn := nodeConnectionCreateResp["nodeConn"].(*nodeConnectionModels.NodeConn)
	addConnErr := ncService.AddInBoundConn(nodeConn)
	if addConnErr != nil {
		logger.Error("子节点中，添加节点到inbound失败")
		return
	}
}

// ProcessInboundTCPData 狂读TCP socket
func (ncService *NodeConnectionService) ProcessInboundTCPData(nodeConn *nodeConnectionModels.NodeConn) {
	chanel := ncService.myPrivateChanel["receiveNodeConnectionMsgChanel"]
	for {
		tcpSourceDataByte := make([]byte, 2048)

		// 断包粘包处理方式（目前出现得不多，先不管了）
		// length := 0
		// for {
		// 	var readErr interface{}
		// 	length, readErr = (*nodeConn.Socket).Read(tcpSourceDataByte)
		// 	if readErr != nil {
		// 		// TODO 连接失败要全部子节点都断掉

		// 		// 释放一个inBound限制
		// 		<-ncService.limitTCPInboundConn
		// 		// logger.Debug(config.ParseNodeID(ncService.conf) + " inBound relive")
		// 		// logger.Debug(readErr)
		// 		return
		// 	}
		// 	if length == tcpSourceDataByte
		// }

		length, readErr := (*nodeConn.Socket).Read(tcpSourceDataByte)
		if readErr != nil {
			// TODO 连接失败要全部子节点都断掉

			// 释放一个inBound限制
			<-ncService.limitTCPInboundConn
			// logger.Debug(config.ParseNodeID(ncService.conf) + " inBound relive")
			// logger.Debug(readErr)
			return
		}

		tcpSourceData := string(tcpSourceDataByte[:length])

		tcpData, parseErr := ncService.ParseTCPData(tcpSourceData)
		if parseErr != nil {
			logger.Error(tcpSourceData)
			logger.Error(parseErr.GetMessage())
			continue
		}

		chanel <- map[string]interface{}{
			"nodeConn":    nodeConn,
			"tcpData":     tcpData.(map[string]interface{}),
			"messageFrom": "inBound",
		}
	}
}

// ParseTCPData 解析TCP数据包
func (ncService *NodeConnectionService) ParseTCPData(tcpSourceData string) (interface{}, *error.Error) {
	// 取出content部分直接处理
	tcpSourceData = tcpSourceData[strings.Index(tcpSourceData, "content:")+len("content:"):]
	contentBase64 := tcpSourceData[:]

	contentByte, decodeErr := base64.StdEncoding.DecodeString(contentBase64)
	if decodeErr != nil {
		return nil, error.New(map[string]interface{}{
			"message": "tcp数据包解析失败: enCodeBase64",
		})
	}
	content := string(contentByte)
	var contentJSON interface{}
	jSONDecodeErr := json.Unmarshal([]byte(content), &contentJSON)
	if jSONDecodeErr != nil {
		return nil, error.New(map[string]interface{}{
			"message": "tcp数据包解析失败:content json unMarshal",
		})
	}

	var msgJSON interface{}
	// logger.Debug(contentJSON)
	contentJSONMsg := contentJSON.(map[string]interface{})["msg"].(string)
	msgJSONDecodeErr := json.Unmarshal([]byte(contentJSONMsg), &msgJSON)
	if msgJSONDecodeErr != nil {
		return nil, error.New(map[string]interface{}{
			"message": "tcp数据包解析失败:message content json umMarshal",
		})
	}

	contentJSON.(map[string]interface{})["msgJSON"] = msgJSON
	contentJSON.(map[string]interface{})["nodeID"] = msgJSON.(map[string]interface{})["from"].(map[string]interface{})["nodeID"]

	return contentJSON, nil
}
