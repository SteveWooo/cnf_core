package services

import (
	"github.com/cnf_core/src/utils/timer"
)

// RunMonitor 监控InBound和outBound的情况。nodeID长期为空的
func (ncService *NodeConnectionService) RunMonitor() {
	for {
		timer.Sleep(1000)
		// 重复检查是否存在需要destroy的连接。有就从桶里删除
		for i := 0; i < len(ncService.inBoundConn); i++ {
			if ncService.inBoundConn[i] == nil {
				continue
			}

			if ncService.inBoundConn[i].GetDestroy() == true {
				// logger.Debug("clear inBound")
				ncService.DeleteInBoundConn(ncService.inBoundConn[i])
			}
		}

		for i := 0; i < len(ncService.outBoundConn); i++ {
			if ncService.outBoundConn[i] == nil {
				continue
			}

			if ncService.outBoundConn[i].GetDestroy() == true {
				// logger.Debug("clear outBound")
				ncService.DeleteOutBoundConn(ncService.outBoundConn[i])
			}
		}
	}
}
