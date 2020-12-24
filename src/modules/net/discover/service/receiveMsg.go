package service

import (
	commonModels "github.com/cnf_core/src/modules/net/common/models"
	discoverModel "github.com/cnf_core/src/modules/net/discover/models"
	"github.com/cnf_core/src/utils/error"
)

// ReceiveMsg 作为发现服务消息接收的入口
func (discoverService *DiscoverService) ReceiveMsg(data interface{}) (interface{}, *error.Error) {
	body := data.(map[string]interface{})["body"]
	msg := body.(map[string]interface{})["msgJSON"]

	nodeID := body.(map[string]interface{})["senderNodeID"].(string)
	sourceIP := data.(map[string]interface{})["sourceIP"].(string)
	sourceServicePort := data.(map[string]interface{})["sourceServicePort"].(string)
	// logger.Debug(config.ParseNodeID(discoverService.conf) + " receive : " + msg.(map[string]interface{})["type"].(string) + " from: " + nodeID)
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

	cache, existCache := discoverService.pingPongCache[nodeID]
	if existCache == false {
		// logger.Debug("senderNodeID: " + nodeID + "\nmyNodeID :" + config.ParseNodeID(discoverService.conf))
		return nil, error.New(map[string]interface{}{
			"message": "非法数据包：收到数据包，但是没有给这个节点做缓存。",
		})
	}

	// 完成握手，加入Bucket
	if (cache.GetDoingPing() == false || cache.GetDoingPing() == true) && cache.GetPing() == true && cache.GetPong() == true {
		delete(discoverService.pingPongCache, nodeID)
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
			"bucketEvent": "addNew",
			"node":        newNode,
		}, nil
	}

	// 非主动doingPong，并收到对方ping的
	// 这时候需要先回复Pong，再主动发一个Ping
	if cache.GetDoingPing() == false && cache.GetPing() == true && cache.GetPong() == false {
		// logger.Debug("retu")
		// discoverService.DoPong(sourceIP, sourceServicePort, nodeID)
		discoverService.DoPing(sourceIP, sourceServicePort, nodeID)
		return nil, nil
	}

	// 主动发起Ping TODO：检查进程时，定期重试ping，或者删除掉这个缓存
	if cache.GetDoingPing() == true && cache.GetPing() == true && cache.GetPong() == false {
		// discoverService.DoPong(sourceIP, sourceServicePort, nodeID)
		// discoverService.DoPing(sourceIP, sourceServicePort, nodeID)
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
