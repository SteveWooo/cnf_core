package models

import "github.com/cnf_core/src/utils/timer"

// PingPongCachePackage 处理PingPong的缓存数据
type PingPongCachePackage struct {
	nodeID      string
	doingPing   bool
	receivePing bool
	receivePong bool
	ts          int64

	kv chan bool
}

// CreatePingPongCache 初始化一个缓存对象
func CreatePingPongCache(nodeID string) *PingPongCachePackage {
	newCache := PingPongCachePackage{
		nodeID:      nodeID,
		doingPing:   false,
		receivePing: false,
		receivePong: false,
		ts:          timer.Now(),
		kv:          make(chan bool, 1),
	}

	return &newCache
}

// SetDoingPing 反射
func (c *PingPongCachePackage) SetDoingPing() {
	c.kv <- true
	c.doingPing = true
	<-c.kv
}

// SetReceivePing 反射
func (c *PingPongCachePackage) SetReceivePing() {
	c.kv <- true
	c.receivePing = true
	<-c.kv
}

// SetReceivePong 反射
func (c *PingPongCachePackage) SetReceivePong() {
	c.kv <- true
	c.receivePong = true
	<-c.kv
}

// GetPing 反射
func (c *PingPongCachePackage) GetPing() bool {
	c.kv <- true
	status := c.receivePing
	<-c.kv
	return status
}

// GetPong 反射
func (c *PingPongCachePackage) GetPong() bool {
	c.kv <- true
	status := c.receivePong
	<-c.kv
	return status
}

// GetDoingPing 反射
func (c *PingPongCachePackage) GetDoingPing() bool {
	c.kv <- true
	status := c.doingPing
	<-c.kv
	return status
}
