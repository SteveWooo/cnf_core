package services

import (
	"encoding/base64"
	"encoding/json"
	"strconv"

	nodeConnectionModels "github.com/cnf_core/src/modules/net/nodeConnection/models"
	"github.com/cnf_core/src/utils/config"
	"github.com/cnf_core/src/utils/error"
	"github.com/cnf_core/src/utils/sign"
	"github.com/cnf_core/src/utils/timer"
)

// SendShake 发送握手包
func (ncService *NodeConnectionService) SendShake(nodeConn *nodeConnectionModels.NodeConn, targetNodeID string) *error.Error {
	confNet := ncService.conf.(map[string]interface{})["net"]
	now := strconv.FormatInt(timer.Now(), 10)
	tcpSourceDataMsgFrom := map[string]interface{}{
		"nodeID":  config.ParseNodeID(ncService.conf),
		"ip":      confNet.(map[string]interface{})["ip"],
		"tcpport": confNet.(map[string]interface{})["servicePort"],
		"udpport": confNet.(map[string]interface{})["servicePort"],
	}

	tcpSourceDataMsg := map[string]interface{}{
		"ts":      now,
		"version": "1",
		"from":    tcpSourceDataMsgFrom,
	}

	// msg要转换为字符串。以后要用base64
	tcpSourceDataMsgJSONString, msgUncodeErr := json.Marshal(tcpSourceDataMsg)
	if msgUncodeErr != nil {
		return error.New(map[string]interface{}{
			"message": "shaker包创建失败，消息部分转码失败",
		})
	}

	// tcpSourceDataMsgBase64String := base64.StdEncoding.EncodeToString(tcpSourceDataMsgJSONString)

	tcpSourceData := map[string]interface{}{
		"event":        "shakeEvent",
		"msg":          string(tcpSourceDataMsgJSONString),
		"targetNodeID": targetNodeID,
	}

	// var tcpData interface{}
	tcpData, tcpDataJSONUncodeErr := json.Marshal(tcpSourceData)
	if tcpDataJSONUncodeErr != nil {
		return error.New(map[string]interface{}{
			"message": "shaker包创建失败",
		})
	}

	// 消息主体需要先做base64编码，防止乱读。这个其实就是数据包中的content部分
	tcpDataBase64String := base64.StdEncoding.EncodeToString(tcpData)
	content := "ts:" + now + ";nodeid:" + config.ParseNodeID(ncService.conf) + ";content:" + tcpDataBase64String

	// 内容长度获取。内容长度是从ts:开始计算的。不仅仅是content部分。因为这里要处理粘包断包问题
	contentLength := len(content)

	// 数据包头hash字段获取
	contentHash := sign.Hash(tcpDataBase64String)

	// 封装整个数据包
	tcpPackage := "hash:" + contentHash + ";content-length:" + strconv.Itoa(contentLength) + ";" + content

	// 发送
	ncService.Send(nodeConn, tcpPackage)

	return nil
}

// Send 底层tcp发包
func (ncService *NodeConnectionService) Send(nodeConn *nodeConnectionModels.NodeConn, message string) {
	// logger.Debug("sent:" + message)
	nodeConn.Socket.Write([]byte(message))
}
