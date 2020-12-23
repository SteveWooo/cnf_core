package services

import (
	"encoding/base64"
	"encoding/json"
	"net"
	"strings"

	"github.com/cnf_core/src/utils/error"
	"github.com/cnf_core/src/utils/logger"
)

// RunService 启动节点通信的TCP相关服务
func (ncService *NodeConnectionService) RunService(chanels map[string]chan map[string]interface{}, signal chan bool) *error.Error {
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

		go ncService.ProcessTCPData(chanels["nodeConnectionMsgChanel"], conn)
	}
}

// ProcessTCPData 狂读TCP socket
func (ncService *NodeConnectionService) ProcessTCPData(chanel chan map[string]interface{}, conn net.Conn) {
	for {
		tcpSourceDataByte := make([]byte, 1024)

		length, readErr := conn.Read(tcpSourceDataByte)
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

		chanel <- tcpData.(map[string]interface{})
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

	return contentJSON, nil
}
