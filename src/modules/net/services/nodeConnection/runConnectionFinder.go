package nodeconnection

import (
	"github.com/cnf_core/src/utils/timer"
)

// RunConnectionFinder 寻找节点并尝试连接
func RunConnectionFinder(chanels map[string]chan map[string]interface{}) {
	go handleNodeProvider(chanels["bucketNodeChanel"])
}

func handleNodeProvider(chanel chan map[string]interface{}) {
	for {
		node := <-chanel
		if node != nil {
		}

		// logger.Debug(node)

		// 检查这个节点是否已经连接成功

		// 尝试与该节点进行连接

		timer.Sleep(1000)
	}
}
