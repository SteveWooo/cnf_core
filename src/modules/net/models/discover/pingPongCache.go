package discover

import "github.com/cnf_core/src/utils/timer"

// PingPongCachePackage 处理PingPong的缓存数据
type PingPongCachePackage struct {
	nodeID      string
	doingPing   bool
	receivePing bool
	receivePong bool
	ts          int64
}

// CreatePingPongCache 初始化一个缓存对象
func CreatePingPongCache(nodeID string) *PingPongCachePackage {
	newCache := PingPongCachePackage{
		nodeID:      nodeID,
		doingPing:   false,
		receivePing: false,
		receivePong: false,
		ts:          timer.Now(),
	}

	return &newCache
}

// SetDoingPing 反射
func (c *PingPongCachePackage) SetDoingPing() {
	c.doingPing = true
}

// SetReceivePing 反射
func (c *PingPongCachePackage) SetReceivePing() {
	c.receivePing = true
}

// SetReceivePong 反射
func (c *PingPongCachePackage) SetReceivePong() {
	c.receivePong = true
}

// GetPing 反射
func (c *PingPongCachePackage) GetPing() bool {
	return c.receivePing
}

// GetPong 反射
func (c *PingPongCachePackage) GetPong() bool {
	return c.receivePong
}

// GetdoingPing 反射
func (c *PingPongCachePackage) GetDoingPing() bool {
	return c.doingPing
}
