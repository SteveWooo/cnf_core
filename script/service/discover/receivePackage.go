package discover

import (
	commonModels "github.com/cnf_core/src/modules/net/models/common"
	discoverModel "github.com/cnf_core/src/modules/net/models/discover"
	error "github.com/cnf_core/src/utils/error"
)

// ReceiveMsg 作为发现服务消息接收的入口
func ReceiveMsg(data interface{}) (interface{}, *error.Error) {
	body := data.(map[string]interface{})["body"]
	msg := body.(map[string]interface{})["msgJSON"]

	nodeID := body.(map[string]interface{})["nodeID"].(string)
	sourceIP := data.(map[string]interface{})["sourceIP"].(string)
	sourceServicePort := data.(map[string]interface{})["sourceServicePort"].(string)
	// logger.Debug(body)
	// Ping case
	if msg.(map[string]interface{})["type"] == "1" {
		pingErr := ReceivePing(data)
		if pingErr != nil {
			return nil, pingErr
		}
	}

	// Pong case
	if msg.(map[string]interface{})["type"] == "2" {
		pongErr := ReceivePong(data)
		if pongErr != nil {
			return nil, pongErr
		}
	}

	cache, existCache := pingPongCache[nodeID]
	if existCache == false {
		return nil, error.New(map[string]interface{}{
			"message": "非法数据包：收到数据包，但是没有给这个节点做缓存。",
		})
	}

	// 完成握手，加入Bucket
	if (cache.GetDoingPing() == false || cache.GetDoingPing() == true) && cache.GetPing() == true && cache.GetPong() == true {
		delete(pingPongCache, nodeID)
		// logger.Debug("000")
		newNode, newNodeError := commonModels.CreateNode(map[string]interface{}{
			"ip":          sourceIP,
			"servicePort": sourceServicePort,
			"nodeID":      nodeID,
		})
		if newNodeError != nil {
			return nil, newNodeError
		}

		if msg.(map[string]interface{})["type"] == "1" {
			shaker.DoPong(sourceIP, sourceServicePort)
		}

		return map[string]interface{}{
			"bucketEvent": "addNew",
			"node":        newNode,
		}, nil
	}

	// 非主动doingPong，并收到对方ping的
	// 这时候需要先回复Pong，再主动发一个Ping
	if cache.GetDoingPing() == false && cache.GetPing() == true && cache.GetPong() == false {
		// logger.Debug("retu")
		shaker.DoPong(sourceIP, sourceServicePort)
		shaker.DoPing(sourceIP, sourceServicePort)
		return nil, nil
	}

	// 主动发起Ping，但是也收到了对方的Ping，这时候就再发一个Ping给对方即可
	if cache.GetDoingPing() == true && cache.GetPing() == true && cache.GetPong() == false {
		shaker.DoPong(sourceIP, sourceServicePort)
		shaker.DoPing(sourceIP, sourceServicePort)
		return nil, nil
	}

	// 没有doingPing，也没有收到过ping，直接收到pong的
	// 无视即可
	if cache.GetDoingPing() == false && cache.GetPing() == false && cache.GetPong() == true {
		// logger.Debug("2")
		return nil, nil
	}

	// 主动doingPing，并且收到pong的情况
	// 这时候不需要干啥，干等获得对方的Ping即可
	if cache.GetDoingPing() == true && cache.GetPing() == false && cache.GetPong() == true {
		// logger.Debug("ret3")
		return nil, nil
	}

	return nil, nil
}

// ReceivePing 接收Ping包的处理逻辑
func ReceivePing(data interface{}) *error.Error {
	// 凡是收到Ping的，都要检查是否有Cache，没有的话就要创建。
	// 有Cache，说明是主动doPing过的
	setPingCacheErr := setPingCache(data)
	if setPingCacheErr != nil {
		return error.New(map[string]interface{}{
			"message": "设置Ping缓存时失败",
		})
	}

	// shaker.DoPong(sourceIP, sourceServicePort)
	return nil
}

// ReceivePong 接收Pong包的所有处理逻辑
func ReceivePong(data interface{}) *error.Error {
	setPongCacheErr := setPongCache(data)
	if setPongCacheErr != nil {
		return error.New(map[string]interface{}{
			"message": "设置Pong缓存时失败",
		})
	}
	return nil
}

// setPingCache 收到Ping包，设置对应的缓存
// @param data udp数据包Json
func setPingCache(data interface{}) *error.Error {
	body := data.(map[string]interface{})["body"]
	nodeID := body.(map[string]interface{})["nodeID"].(string)
	if nodeID == "" {
		return error.New(map[string]interface{}{
			"message": "数据包中解析不出NodeID",
		})
	}

	cache, existCache := pingPongCache[nodeID]

	// cache存在，说明已经收到这个Node的握手包，
	// 或者本节点已经向这个节点主动发起过握手了
	if existCache {
		cache.SetReceivePing()
	}

	// 空则创建新的
	if !existCache {
		newCache := discoverModel.CreatePingPongCache(nodeID)
		newCache.SetReceivePing()
		pingPongCache[nodeID] = newCache
	}

	return nil
}

// setPongCache 收到Pong包，设置对应的缓存
// @param data udp数据包Json
func setPongCache(data interface{}) *error.Error {
	body := data.(map[string]interface{})["body"]
	nodeID := body.(map[string]interface{})["nodeID"].(string)
	if nodeID == "" {
		return error.New(map[string]interface{}{
			"message": "数据包中解析不出NodeID",
		})
	}

	cache, existCache := pingPongCache[nodeID]
	if !existCache {
		// 未发出过Ping或者未收到过Ping时，收到的Pong不需要理会
		return nil
	}

	cache.SetReceivePong()

	return nil
}