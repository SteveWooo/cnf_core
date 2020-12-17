package discover

import (
	error "github.com/cnf_core/src/utils/error"
	json "encoding/json"
)

/**
 * 负责解析接收到的UDP数据包
 * TODO 检查networkid、字段合法性、签名合法性问题
 * @param message 从socket缓冲区中读取到的udp数据包
 */
func ParsePackage(message string) (interface{}, interface{}){
	var jsonMsg interface{}

	// 首先把数据报文Json序列化
	jsonUnMarshalErr := json.Unmarshal([]byte(message), &jsonMsg)
	if jsonUnMarshalErr != nil {
		return nil, error.New(map[string]interface{}{
			"message" : "接收到不合法的UDP数据包",
		})
	}

	return jsonMsg, nil
}