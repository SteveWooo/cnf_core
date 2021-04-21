package connection

// GetStatus 获取nodeConnection的所有状态
func (nodeConnection *NodeConnection) GetStatus() map[string]interface{} {
	nodeConnectionStatus := make(map[string]interface{})
	nodeConnectionStatus["serviceStatus"] = nodeConnection.service.GetStatus()
	return nodeConnectionStatus
}
