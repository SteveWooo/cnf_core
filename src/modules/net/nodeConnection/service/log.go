package services

// GetStatus 获取节点当前状态
func (ncService *NodeConnectionService) GetStatus() map[string]interface{} {
	serviceStatus := make(map[string]interface{})

	// 处理连接情况
	var outBoundConn []string
	var inBoundConn []string
	for i := 0; i < len(ncService.inBoundConn); i++ {
		nodeConn := ncService.inBoundConn[i]
		if nodeConn == nil {
			continue
		}
		// 握手失败的也不需要
		if nodeConn.IsShaked() == false {
			// continue
		}

		inBoundConn = append(inBoundConn, nodeConn.GetNodeID())
	}

	for i := 0; i < len(ncService.outBoundConn); i++ {
		nodeConn := ncService.outBoundConn[i]
		if nodeConn == nil {
			continue
		}
		// 握手失败的也不需要
		if nodeConn.IsShaked() == false {
			// continue
		}

		outBoundConn = append(outBoundConn, nodeConn.GetNodeID())
	}

	serviceStatus["outBoundConn"] = outBoundConn
	serviceStatus["inBoundConn"] = inBoundConn

	return serviceStatus
}
