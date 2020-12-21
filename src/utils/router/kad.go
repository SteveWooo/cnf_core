package router

import (
	"strconv"
)

func CalculateDistance(nodeId_1 string, nodeId_2 string) int {
	var sourceId [64]int64

	nodeId_1 = nodeId_1[2:len(nodeId_1)]
	nodeId_2 = nodeId_2[2:len(nodeId_2)]
	// 比较两个nodeId的异或距离
	for i := 0; i < len(nodeId_1); i += 2 {
		base_1, _ := strconv.ParseInt(nodeId_1[i:i+2], 16, 10)
		base_2, _ := strconv.ParseInt(nodeId_2[i:i+2], 16, 10)
		sourceId[i/2] = base_1 ^ base_2
	}

	// 计算最远距离，决定存放在哪个桶里
	bucketCount := 0
	for i := 0; i < len(sourceId); i++ {
		if sourceId[i] == 0 {
			bucketCount++
		} else {
			break
		}
	}

	return bucketCount
}
