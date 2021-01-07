package services

// RunMonitor 监控InBound和outBound的情况。nodeID长期为空的
func (ncService *NodeConnectionService) RunMonitor() {
	// 重复检查是否存在需要destroy的连接。有就从桶里删除
	for i := 0; i < len(ncService.inBoundConn); i++ {
		if ncService.inBoundConn[i] == nil {
			continue
		}

		if ncService.inBoundConn[i].GetDestroy() == true {
			// logger.Debug("clean inBound")
			if len(ncService.myPrivateChanel["bucketOperateChanel"]) != cap(ncService.myPrivateChanel["bucketOperateChanel"]) {
				ncService.myPrivateChanel["bucketOperateChanel"] <- map[string]interface{}{
					"event":  "deleteNode",
					"nodeID": ncService.outBoundConn[i].GetNodeID(),
				}
			}
			ncService.DeleteInBoundConn(ncService.inBoundConn[i])
		}
	}

	for i := 0; i < len(ncService.outBoundConn); i++ {
		if ncService.outBoundConn[i] == nil {
			continue
		}

		if ncService.outBoundConn[i].GetDestroy() == true {
			// logger.Debug("clean outBound")
			if len(ncService.myPrivateChanel["bucketOperateChanel"]) != cap(ncService.myPrivateChanel["bucketOperateChanel"]) {
				ncService.myPrivateChanel["bucketOperateChanel"] <- map[string]interface{}{
					"event":  "deleteNode",
					"nodeID": ncService.outBoundConn[i].GetNodeID(),
				}
			}
			ncService.DeleteOutBoundConn(ncService.outBoundConn[i])
		}
	}
}
