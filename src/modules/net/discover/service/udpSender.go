package service

import (
	"encoding/json"
	"net"

	"github.com/cnf_core/src/utils/error"
	"github.com/cnf_core/src/utils/sign"
	"github.com/cnf_core/src/utils/timer"
)

// DoPong 处理Pong数据包
// @param targetIP string 目标IP
// @param targetServicePort string 目标服务端口
func (discoverService *DiscoverService) DoPong(targetIP string, targetServicePort string) interface{} {
	body, buildBodyErr := discoverService.BuildPackageBody("2")
	if buildBodyErr != nil {
		return buildBodyErr
	}

	discoverService.Send(body, targetIP, targetServicePort)
	return nil
}

// DoPing 处理Pong数据包
// @param targetIP string 目标IP
// @param targetServicePort string 目标服务端口
func (discoverService *DiscoverService) DoPing(targetIP string, targetServicePort string) interface{} {
	body, buildBodyErr := discoverService.BuildPackageBody("1")
	if buildBodyErr != nil {
		return buildBodyErr
	}

	discoverService.Send(body, targetIP, targetServicePort)
	return nil
}

// Send 把数据发送到目标地址
// @param message string 数据内容
// @param targetIP string 目标IP
// @param targetServicePort string 目标服务端口
func (discoverService *DiscoverService) Send(message string, targetIP string, targetServicePort string) interface{} {
	ipPort := targetIP + ":" + targetServicePort

	targetAddress, resolveErr := net.ResolveUDPAddr("udp", ipPort)
	if resolveErr != nil {
		return error.New(map[string]interface{}{
			"message": "创建udp地址对象失败",
		})
	}
	// logger.Debug(s.SocketConn)
	discoverService.socketConn.WriteToUDP([]byte(message), targetAddress)

	return nil
}

// BuildPackageBody 构建一个规范的UDP包
func (discoverService *DiscoverService) BuildPackageBody(packType string) (string, interface{}) {
	conf := discoverService.conf
	confNet := conf.(map[string]interface{})["net"]
	now := timer.Now() // 获取毫秒时间戳

	// 发现数据包的主体内容
	msgSource := map[string]interface{}{
		"ts":      now,
		"type":    packType,
		"version": "1",
		"from": map[string]interface{}{
			"ip":        confNet.(map[string]interface{})["ip"],
			"tcpport":   confNet.(map[string]interface{})["servicePort"],
			"udpport":   confNet.(map[string]interface{})["servicePort"],
			"networkid": confNet.(map[string]interface{})["networkid"],
		},
	}

	msgJSON, msgJSONMarshalErr := json.Marshal(msgSource)
	if msgJSONMarshalErr != nil {
		return "", error.New(map[string]interface{}{
			"message": "系统错误。shaker.go DoPong msgJSONMarshalErr",
		})
	}

	msg := string(msgJSON)
	msgHash := sign.Hash(msg)

	// 之后还要签名
	signature, signErr := sign.Sign(msgHash, confNet.(map[string]interface{})["localPrivateKey"].(string))
	if signErr != nil {
		return "", error.New(map[string]interface{}{
			"message": "系统错误。shaker.go DoPong signErr",
		})
	}

	// 取出recid，单独作为字段
	recid := ""
	if signature[0:2] == "1b" {
		recid = "0"
	}
	if signature[0:2] == "1c" {
		recid = "1"
	}

	signature = signature[2:]

	bodySource := map[string]interface{}{
		"msg":       msg,
		"recid":     recid,
		"signature": signature,
	}

	bodyJSON, bodyJSONMarshalErr := json.Marshal(bodySource)
	if bodyJSONMarshalErr != nil {
		return "", error.New(map[string]interface{}{
			"message": "系统错误。shaker.go DoPong bodyJSONMarshalErr",
		})
	}
	body := string(bodyJSON)
	return body, nil
}
