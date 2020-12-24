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
	// 非主节点，不需要监听socket
	confNet := ncService.conf.(map[string]interface{})["net"]
	if confNet.(map[string]interface{})["masterServer"] != "true" {
		signal <- true
		return nil
	}

	tcpListener, listenErr := net.Listen("tcp", ncService.socketAddr)
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
		nodeConn.Build(conn)
		nodeConn.SetRemoteAddr(remoteAddr.String())

		// 添加一个未握手的连接到InBoundConn里面去
		ncService.AddInBoundConn(&nodeConn)

		go ncService.ProcessTCPData(chanels["receiveNodeConnectionMsgChanel"], &nodeConn)
	}
}

// ProcessTCPData 狂读TCP socket
func (ncService *NodeConnectionService) ProcessTCPData(chanel chan map[string]interface{}, nodeConn *nodeConnectionModels.NodeConn) {
	for {
		tcpSourceDataByte := make([]byte, 1024)

		length, readErr := nodeConn.Socket.Read(tcpSourceDataByte)
		if readErr != nil {
			// 读取数据失败，说明socket已经断掉，所以要结束这个socket

			// 释放一个inBound限制
			<-ncService.limitTCPInboundConn
			logger.Debug("relive")
			return
		}

		tcpSourceData := string(tcpSourceDataByte[:length])

		tcpData, parseErr := ncService.ParseTCPData(tcpSourceData)
		if parseErr != nil {
			logger.Error(parseErr.GetMessage())
			continue
		}

		chanel <- map[string]interface{}{
			"nodeConn": nodeConn,
			"tcpData":  tcpData.(map[string]interface{}),
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
			"message": "tcp数据包解析失败",
		})
	}
	content := string(contentByte)
	var contentJSON interface{}
	jSONDecodeErr := json.Unmarshal([]byte(content), &contentJSON)
	if jSONDecodeErr != nil {
		return nil, error.New(map[string]interface{}{
			"message": "tcp数据包解析失败",
		})
	}

	var msgJSON interface{}
	// logger.Debug(contentJSON)
	contentJSONMsg := contentJSON.(map[string]interface{})["msg"].(string)
	msgJSONDecodeErr := json.Unmarshal([]byte(contentJSONMsg), &msgJSON)
	if msgJSONDecodeErr != nil {
		return nil, error.New(map[string]interface{}{
			"message": "tcp数据包解析失败",
		})
	}

	contentJSON.(map[string]interface{})["msgJSON"] = msgJSON
	contentJSON.(map[string]interface{})["senderNodeID"] = msgJSON.(map[string]interface{})["nodeID"]

	return contentJSON, nil
}
