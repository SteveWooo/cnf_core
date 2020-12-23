package service

import (
	"encoding/json"
	"net"
	"strconv"

	"github.com/cnf_core/src/utils/error"
	"github.com/cnf_core/src/utils/sign"
)

// RunService 启动UDP服务器，并持续监听状态。
// @param chanel 与消息队列通信的发现服务专用管道
// @param signal 用于通知上级服务udp服务已就绪
func (discoverService *DiscoverService) RunService(chanels map[string]chan map[string]interface{}, signal chan bool) *error.Error {
	udpConn, listenErr := net.ListenUDP("udp", discoverService.socketAddr)

	if listenErr != nil {
		return error.New(map[string]interface{}{
			"message":   "监听UDP端口失败",
			"originErr": listenErr,
		})
	}
	defer udpConn.Close()
	// 也要把socket赋值给shaker
	discoverService.socketConn = udpConn

	signal <- true

	// 暴力读取udp数据
	for {
		// 压入数据，填缓冲池。如果缓冲池满了，就不会有下面的协程创建了，意味着会丢掉过多的UDP包
		discoverService.limitProcessUDPData <- true
		go discoverService.ProcessUDPData(chanels["discoverMsgReceiveChanel"], udpConn)
	}
}

// ProcessUDPData 协程，负责读取udp数据
func (discoverService *DiscoverService) ProcessUDPData(chanel chan map[string]interface{}, udpConn *net.UDPConn) *error.Error {
	// 源数据
	udpSourceDataByte := make([]byte, 1024)

	length, info, readUDPErr := udpConn.ReadFromUDP(udpSourceDataByte) // 挂起
	if readUDPErr != nil {
		return error.New(map[string]interface{}{
			"message":   "读取UDP数据错误",
			"originErr": readUDPErr,
		})
	}
	message := string(udpSourceDataByte[:length])

	// 提取出信息和报文头内容的数据（但数据未被格式化
	udpSourceData := make(map[string]interface{})
	udpSourceData["message"] = message
	udpSourceData["sourceIP"] = info.IP.String()
	udpSourceData["sourceServicePort"] = strconv.Itoa(info.Port)

	udpData, parseUDPError := discoverService.ParseUDPData(udpSourceData)
	if parseUDPError != nil {
		return parseUDPError
	}

	// 把消息推送到消息队列中。只有队列不满的情况下，这条协程才会往下走
	chanel <- udpData.(map[string]interface{})

	// 当数据确实推送到消息队列后，才能释放这条协程，创建新的监听UDP数据的协程。
	<-discoverService.limitProcessUDPData

	return nil
}

// ParseUDPData 负责解析接收到的UDP数据包
// @param data 从socket缓冲区中读取到的udp数据包
// TODO 检查networkid、字段合法性、签名合法性问题
func (discoverService *DiscoverService) ParseUDPData(udpData map[string]interface{}) (interface{}, *error.Error) {
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
