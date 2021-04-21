package services

import (
	"encoding/base64"
	"encoding/json"
	"strconv"

	commonModels "github.com/cnf_core/src/modules/net/common/models"
	nodeConnectionModels "github.com/cnf_core/src/modules/net/nodeConnection/models"
	"github.com/cnf_core/src/utils/config"
	"github.com/cnf_core/src/utils/sign"
	"github.com/cnf_core/src/utils/timer"
)

// buildCommonBody 统一body构造
func (ncService *NodeConnectionService) buildCommonBody() (map[string]interface{}, string) {
	confNet := ncService.conf.(map[string]interface{})["net"]
	now := strconv.FormatInt(timer.Now(), 10)
	tcpSourceDataMsgFrom := map[string]interface{}{
		"nodeID":    config.ParseNodeID(ncService.conf),
		"ip":        confNet.(map[string]interface{})["ip"],
		"tcpport":   confNet.(map[string]interface{})["servicePort"],
		"udpport":   confNet.(map[string]interface{})["servicePort"],
		"networkid": confNet.(map[string]interface{})["networkid"],
	}

	tcpSourceDataMsg := map[string]interface{}{
		"ts":      now,
		"version": "1",
		"from":    tcpSourceDataMsgFrom,
	}

	return tcpSourceDataMsg, now
}

// GetShakePackString 获取一个握手包
func (ncService *NodeConnectionService) GetShakePackString(shaketype string, targetNodeID string) string {
	tcpSourceDataMsg, now := ncService.buildCommonBody()

	// msg要转换为字符串。以后要用base64
	tcpSourceDataMsgJSONString, msgUncodeErr := json.Marshal(tcpSourceDataMsg)
	if msgUncodeErr != nil {
		return ""
	}

	tcpSourceData := map[string]interface{}{
		"event":        "",
		"msg":          string(tcpSourceDataMsgJSONString),
		"targetNodeID": targetNodeID,
	}

	tcpSourceData["event"] = shaketype

	// var tcpData interface{}
	tcpData, tcpDataJSONUncodeErr := json.Marshal(tcpSourceData)
	if tcpDataJSONUncodeErr != nil {
		return ""
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

	return tcpPackage
}

// GetFindNodePackString 构建一个邻居获取包
// @param findingNodeID 需要查询的NodeID
// @param targetNodeID 目标结点NodeID
func (ncService *NodeConnectionService) GetFindNodePackString(findingNodeID string, targetNodeID string) string {
	tcpSourceDataMsg, now := ncService.buildCommonBody()

	tcpSourceDataMsg["findingNodeID"] = findingNodeID

	// msg要转换为字符串。以后要用base64
	tcpSourceDataMsgJSONString, msgUncodeErr := json.Marshal(tcpSourceDataMsg)
	if msgUncodeErr != nil {
		return ""
	}

	tcpSourceData := map[string]interface{}{
		"event":        "findNode",
		"msg":          string(tcpSourceDataMsgJSONString),
		"targetNodeID": targetNodeID,
	}

	// var tcpData interface{}
	tcpData, tcpDataJSONUncodeErr := json.Marshal(tcpSourceData)
	if tcpDataJSONUncodeErr != nil {
		return ""
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

	return tcpPackage
}

// GetshareNodeNeighborPackString 构造邻居分享包
func (ncService *NodeConnectionService) GetshareNodeNeighborPackString(neighbors []*commonModels.Node, targetNodeID string) string {
	tcpSourceDataMsg, now := ncService.buildCommonBody()

	tcpSourceDataMsg["nodeNeighbor"] = make([]map[string]string, len(neighbors))

	// 把结点信息一个一个加进来
	for i := 0; i < len(neighbors); i++ {
		tcpSourceDataMsg["nodeNeighbor"].([]map[string]string)[i] = map[string]string{
			"ip":          neighbors[i].GetIP(),
			"servicePort": neighbors[i].GetServicePort(),
			"nodeID":      neighbors[i].GetNodeID(),
		}
	}

	// msg要转换为字符串。以后要用base64
	tcpSourceDataMsgJSONString, msgUncodeErr := json.Marshal(tcpSourceDataMsg)
	if msgUncodeErr != nil {
		return ""
	}

	tcpSourceData := map[string]interface{}{
		"event":        "shareNodeNeighbor",
		"msg":          string(tcpSourceDataMsgJSONString),
		"targetNodeID": targetNodeID,
	}

	// var tcpData interface{}
	tcpData, tcpDataJSONUncodeErr := json.Marshal(tcpSourceData)
	if tcpDataJSONUncodeErr != nil {
		return ""
	}

	if tcpDataJSONUncodeErr != nil {
		return ""
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

	return tcpPackage
}

// Send 底层tcp发包
func (ncService *NodeConnectionService) Send(nodeConn *nodeConnectionModels.NodeConn, message string) {
	// logger.Debug("sent:" + message)
	(*nodeConn.Socket).Write([]byte(message))
}
