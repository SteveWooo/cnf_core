package discover

import (
	json "encoding/json"

	error "github.com/cnf_core/src/utils/error"
	"github.com/cnf_core/src/utils/sign"
)

// ParsePackage 负责解析接收到的UDP数据包
// @param data 从socket缓冲区中读取到的udp数据包
// TODO 检查networkid、字段合法性、签名合法性问题
func ParsePackage(udpData map[string]interface{}) (interface{}, *error.Error) {
	// 包含头部的数据报文
	data := make(map[string]interface{})
	// 数据报文的主题内容
	var body interface{}

	// 首先把数据报文Json序列化
	udpDataMessage, _ := udpData["message"].(string)
	jsonUnMarshalErr := json.Unmarshal([]byte(udpDataMessage), &body)
	if jsonUnMarshalErr != nil {
		return nil, error.New(map[string]interface{}{
			"message":   "接收到不合法的UDP数据包",
			"originErr": jsonUnMarshalErr,
		})
	}

	// 提取消息
	msg := body.(map[string]interface{})["msg"].(string)
	messageHash := sign.Hash(msg)
	signature := body.(map[string]interface{})["signature"]
	rcid := body.(map[string]interface{})["recid"].(float64)

	// 获取nodeId
	recoverPublicKey, _ := sign.Recover(signature.(string), messageHash, uint64(rcid))
	body.(map[string]interface{})["nodeID"] = recoverPublicKey
	if recoverPublicKey == "" {
		return nil, error.New(map[string]interface{}{
			"message": "接收到不合法的msg内容，无法解析NodeID",
		})
	}

	// 然后把Msg部分Json反序列化
	var msgJSON interface{}
	msgJSONUnMarshalErr := json.Unmarshal([]byte(msg), &msgJSON)
	if msgJSONUnMarshalErr != nil {
		return nil, error.New(map[string]interface{}{
			"message":   "接收到不合法的msg内容",
			"originErr": msgJSONUnMarshalErr,
		})
	}
	body.(map[string]interface{})["msgJSON"] = msgJSON

	data["body"] = body
	data["sourceIP"] = udpData["sourceIP"]
	data["sourceServicePort"] = udpData["sourceServicePort"]

	return data, nil
}
