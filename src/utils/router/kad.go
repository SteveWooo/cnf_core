package router

import (
	"strconv"
)

// CalculateDistance 计算两个节点之间的距离
func CalculateDistance(nodeID1 string, nodeID2 string) int {
	var sourceID [64]int64

	nodeID1 = nodeID1[2:len(nodeID1)]
	nodeID2 = nodeID2[2:len(nodeID2)]
	// 比较两个nodeId的异或距离
	for i := 0; i < len(nodeID1); i += 2 {
		base1, _ := strconv.ParseInt(nodeID1[i:i+2], 16, 10)
		base2, _ := strconv.ParseInt(nodeID2[i:i+2], 16, 10)
		sourceID[i/2] = base1 ^ base2
	}

	// 计算最远距离，决定存放在哪个桶里
	bucketCount := 0
	for i := 0; i < len(sourceID); i++ {
		if sourceID[i] == 0 {
			bucketCount++
		} else {
			break
		}
	}

	// logger.Debug(sourceID)

	return 64 - bucketCount
}
