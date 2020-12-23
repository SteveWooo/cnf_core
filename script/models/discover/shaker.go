package discover

import (
	"encoding/json"
	"net"
	"time"

	"github.com/cnf_core/src/utils/config"
	error "github.com/cnf_core/src/utils/error"
	"github.com/cnf_core/src/utils/sign"
)

// Shaker 握手对象handle
type Shaker struct {
	NodeID     string
	SocketConn *net.UDPConn
}

// SetNodeID 必须使用指针，不然改不了数值
func (s *Shaker) SetNodeID(nodeID string) {
	s.NodeID = nodeID
}

// SetSocketConn 把socket设置进来
func (s *Shaker) SetSocketConn(socketConn *net.UDPConn) {
	s.SocketConn = socketConn
}

func buildBody(packType string) (string, interface{}) {
	conf := config.GetConfig()
	confNet := conf.(map[string]interface{})["net"]
	now := time.Now().UnixNano() / 1e6 // 获取毫秒时间戳

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

/**
 * 处理Pong数据包
 * @param targetIP string 目标IP
 * @param targetServicePort string 目标服务端口
 */
func (s Shaker) DoPong(targetIP string, targetServicePort string) interface{} {
	body, buildBodyErr := buildBody("2")
	if buildBodyErr != nil {
		return buildBodyErr
	}

	s.Send(body, targetIP, targetServicePort)
	return nil
}

/**
 * 处理Pong数据包
 * @param targetIP string 目标IP
 * @param targetServicePort string 目标服务端口
 */
func (s Shaker) DoPing(targetIP string, targetServicePort string) interface{} {
	body, buildBodyErr := buildBody("1")
	if buildBodyErr != nil {
		return buildBodyErr
	}

	s.Send(body, targetIP, targetServicePort)
	return nil
}

/**
 * 把数据发送到目标地址
 * @param message string 数据内容
 * @param targetIP string 目标IP
 * @param targetServicePort string 目标服务端口
 */
func (s Shaker) Send(message string, targetIP string, targetServicePort string) interface{} {
	ipPort := targetIP + ":" + targetServicePort

	targetAddress, resolveErr := net.ResolveUDPAddr("udp", ipPort)
	if resolveErr != nil {
		return error.New(map[string]interface{}{
			"message": "创建udp地址对象失败",
		})
	}
	// logger.Debug(s.SocketConn)
	s.SocketConn.WriteToUDP([]byte(message), targetAddress)

	return nil
}
