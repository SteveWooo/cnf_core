package service

import (
	"encoding/json"
	"strconv"

	commonModels "github.com/cnf_core/src/modules/net/common/models"
	discoverModel "github.com/cnf_core/src/modules/net/discover/models"
	"github.com/cnf_core/src/utils/error"
	"github.com/cnf_core/src/utils/sign"
)

// ReceiveMsg 作为发现服务消息接收的入口
func (discoverService *DiscoverService) ReceiveMsg(data interface{}) (interface{}, *error.Error) {
	body := data.(map[string]interface{})["body"]
	msg := body.(map[string]interface{})["msgJSON"]

	nodeID := body.(map[string]interface{})["senderNodeID"].(string)
	sourceIP := data.(map[string]interface{})["sourceIP"].(string)
	sourceServicePort := data.(map[string]interface{})["sourceServicePort"].(string)
	// logger.Debug(discoverService.conf.(map[string]interface{})["number"].(string) + " receive : " + msg.(map[string]interface{})["type"].(string))
	// Ping case
	if msg.(map[string]interface{})["type"] == "1" {
		pingErr := discoverService.ReceivePing(data)
		if pingErr != nil {
			return nil, pingErr
		}

		// 收到Ping 就要回pong
		discoverService.DoPong(sourceIP, sourceServicePort, nodeID)
	}

	// Pong case
	if msg.(map[string]interface{})["type"] == "2" {
		pongErr := discoverService.ReceivePong(data)
		if pongErr != nil {
			return nil, pongErr
		}
	}

	// 处理寻找邻居数据包
	if msg.(map[string]interface{})["type"] == "3" {
		return discoverService.HandleFindNode(data)
	}

	// 处理返回邻居数据包
	if msg.(map[string]interface{})["type"] == "4" {
		return discoverService.HandleShareNodeNeighbor(data)
	}

	cache, existCache := discoverService.pingPongCache[nodeID]
	if existCache == false {
		// 无缘无故收到Pong就会这样。不用管他
		// logger.Debug(config.ParseNodeID(discoverService.conf) + " receive : " + msg.(map[string]interface{})["type"].(string) + " from: " + nodeID)
		// logger.Debug(discoverService.conf.(map[string]interface{})["number"].(string) + " receive : " + msg.(map[string]interface{})["type"].(string))
		// logger.Debug(discoverService.conf.(map[string]interface{})["number"].(string) + " do release")
		// <-discoverService.pingPongCacheLock
		// logger.Debug(discoverService.conf.(map[string]interface{})["number"].(string) + " release")
		return nil, error.New(map[string]interface{}{
			"message": "非法数据包：收到数据包，但是没有给这个节点做缓存。",
		})
	}

	// 完成握手，加入Bucket
	if (cache.GetDoingPing() == false || cache.GetDoingPing() == true) && cache.GetPing() == true && cache.GetPong() == true {
		delete(discoverService.pingPongCache, nodeID)
		// logger.Debug(discoverService.conf.(map[string]interface{})["number"].(string) + " do release")
		// <-discoverService.pingPongCacheLock
		// logger.Debug(discoverService.conf.(map[string]interface{})["number"].(string) + " release")
		newNode, newNodeError := commonModels.CreateNode(map[string]interface{}{
			"ip":          sourceIP,
			"servicePort": sourceServicePort,
			"nodeID":      nodeID,
		})
		if newNodeError != nil {
			return nil, newNodeError
		}

		// if msg.(map[string]interface{})["type"] == "1" {
		// 	discoverService.DoPong(sourceIP, sourceServicePort, nodeID)
		// }

		return map[string]interface{}{
			"event": "addNew",
			"node":  newNode,
		}, nil
	}

	// 非主动doingPong，并收到对方ping的
	// 这时候需要先回复Pong，再主动发一个Ping
	if cache.GetDoingPing() == false && cache.GetPing() == true && cache.GetPong() == false {
		// logger.Debug("retu")
		// discoverService.DoPong(sourceIP, sourceServicePort, nodeID)
		// logger.Debug(discoverService.conf.(map[string]interface{})["number"].(string) + " do release")
		// <-discoverService.pingPongCacheLock
		// logger.Debug(discoverService.conf.(map[string]interface{})["number"].(string) + " release")
		discoverService.DoPing(sourceIP, sourceServicePort, nodeID)
		return nil, nil
	}

	// 主动发起Ping TODO：检查进程时，定期重试ping，或者删除掉这个缓存
	if cache.GetDoingPing() == true && cache.GetPing() == true && cache.GetPong() == false {
		// discoverService.DoPong(sourceIP, sourceServicePort, nodeID)
		// discoverService.DoPing(sourceIP, sourceServicePort, nodeID)
		// logger.Debug(discoverService.conf.(map[string]interface{})["number"].(string) + " do release")
		// <-discoverService.pingPongCacheLock
		// logger.Debug(discoverService.conf.(map[string]interface{})["number"].(string) + " release")
		return nil, nil
	}

	// 没有doingPing，也没有收到过ping，直接收到pong的
	// 无视即可
	if cache.GetDoingPing() == false && cache.GetPing() == false && cache.GetPong() == true {
		// logger.Debug("2")
		// logger.Debug(discoverService.conf.(map[string]interface{})["number"].(string) + " do release")
		// <-discoverService.pingPongCacheLock
		// logger.Debug(discoverService.conf.(map[string]interface{})["number"].(string) + " release")
		return nil, nil
	}

	// 主动doingPing，并且收到pong的情况
	// 这时候不需要干啥，干等获得对方的Ping即可
	if cache.GetDoingPing() == true && cache.GetPing() == false && cache.GetPong() == true {
		// logger.Debug("ret3")
		// logger.Debug(discoverService.conf.(map[string]interface{})["number"].(string) + " do release")
		// <-discoverService.pingPongCacheLock
		// logger.Debug(discoverService.conf.(map[string]interface{})["number"].(string) + " release")
		return nil, nil
	}

	// <-discoverService.pingPongCacheLock
	return nil, nil
}

// ReceivePing 接收Ping包的处理逻辑
func (discoverService *DiscoverService) ReceivePing(data interface{}) *error.Error {
	// 凡是收到Ping的，都要检查是否有Cache，没有的话就要创建。
	// 有Cache，说明是主动doPing过的
	setPingCacheErr := discoverService.SetPingCache(data)
	if setPingCacheErr != nil {
		return error.New(map[string]interface{}{
			"message": "设置Ping缓存时失败",
		})
	}

	// shaker.DoPong(sourceIP, sourceServicePort)
	return nil
}

// ReceivePong 接收Pong包的所有处理逻辑
func (discoverService *DiscoverService) ReceivePong(data interface{}) *error.Error {
	setPongCacheErr := discoverService.SetPongCache(data)
	if setPongCacheErr != nil {
		return error.New(map[string]interface{}{
			"message": "设置Pong缓存时失败",
		})
	}
	return nil
}

// SetPingCache 收到Ping包，设置对应的缓存
// @param data udp数据包Json
func (discoverService *DiscoverService) SetPingCache(data interface{}) *error.Error {
	body := data.(map[string]interface{})["body"]
	nodeID := body.(map[string]interface{})["senderNodeID"].(string)
	if nodeID == "" {
		return error.New(map[string]interface{}{
			"message": "数据包中解析不出NodeID",
		})
	}

	cache, existCache := discoverService.pingPongCache[nodeID]

	// cache存在，说明已经收到这个Node的握手包，
	// 或者本节点已经向这个节点主动发起过握手了
	if existCache {
		cache.SetReceivePing()
	}

	// 空则创建新的
	if !existCache {
		newCache := discoverModel.CreatePingPongCache(nodeID)
		newCache.SetReceivePing()
		discoverService.pingPongCache[nodeID] = newCache
	}

	return nil
}

// SetPongCache 收到Pong包，设置对应的缓存
// @param data udp数据包Json
func (discoverService *DiscoverService) SetPongCache(data interface{}) *error.Error {
	body := data.(map[string]interface{})["body"]
	nodeID := body.(map[string]interface{})["senderNodeID"].(string)
	if nodeID == "" {
		return error.New(map[string]interface{}{
			"message": "数据包中解析不出NodeID",
		})
	}

	cache, existCache := discoverService.pingPongCache[nodeID]
	if !existCache {
		// 未发出过Ping或者未收到过Ping时，收到的Pong不需要理会
		return nil
	}

	cache.SetReceivePong()

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
	// logger.Debug(body)
	// 提取消息
	msg := body.(map[string]interface{})["msg"].(string)
	messageHash := sign.Hash(msg)
	signature := body.(map[string]interface{})["signature"]
	rcid, _ := strconv.Atoi(body.(map[string]interface{})["recid"].(string))

	// 1、通过恢复公钥的方式获取nodeId
	recoverPublicKey, recoverErr := sign.Recover(signature.(string), messageHash, uint64(rcid))
	body.(map[string]interface{})["senderPublicKey"] = recoverPublicKey
	// body.(map[string]interface{})["senderNodeID"] = recoverPublicKey

	if recoverErr != nil {
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

	// 从body中获得目标nodeid，用于接收方的端口多路复用
	data["targetNodeID"] = body.(map[string]interface{})["targetNodeID"]
	delete(body.(map[string]interface{}), "targetNodeID")

	if data["targetNodeID"] == "" || len(data["targetNodeID"].(string)) != 34 {
		return nil, error.New(map[string]interface{}{
			"message":   "接收到不合法的msg内容，targetNodeID不合法",
			"originErr": msgJSONUnMarshalErr,
		})
	}

	return data, nil
}
