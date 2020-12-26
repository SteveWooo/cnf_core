package services

import (
	"math/rand"
	"strconv"

	"github.com/cnf_core/src/utils/logger"
	"github.com/cnf_core/src/utils/timer"
)

// RunMonitor 监控InBound和outBound的情况。nodeID长期为空的
func (ncService *NodeConnectionService) RunMonitor() {
	for {
		// timer.Sleep(1000)
		timer.Sleep(1000 + rand.Intn(1000))
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
		// localNODE1: 044eb742421d2e24342983c2f6248fbe5400d855d952cdf6bd6f64117696d729a3e0fbaff48863237f510ce4645197a216b5cc2cedff17e3cef7e0d4366d7291e8
		// socketNode1 : 043eb06ad52d601c563950c454445b0345e5fcfe5f0036b8b86aef4eb3e2825f15ddcc5ed27bd9645d710505dcb819c0fbca9dae7f612867f59364583e5ec7fbb9
		if false {

			// logger.Debug("masterInBoundConn: " + strconv.Itoa(len(ncService.masterInBoundConn)))
			// logger.Debug("masterOutBoundConn: " + strconv.Itoa(len(ncService.masterOutBoundConn)))

			inBoundCount := 0
			for i := 0; i < len(ncService.inBoundConn); i++ {
				if ncService.inBoundConn[i] != nil {
					inBoundCount++
				}
			}

			outBoundCount := 0
			for i := 0; i < len(ncService.outBoundConn); i++ {
				if ncService.outBoundConn[i] != nil {
					outBoundCount++
				}
			}
			logger.Debug("InBoundConn: " + strconv.Itoa(inBoundCount) + "; " + "OutBoundConn: " + strconv.Itoa(outBoundCount))
			// logger.Debug()
		}
	}
}
