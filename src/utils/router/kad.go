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
	var base1, base2 int64
	for i := 0; i < len(nodeID1); i += 2 {
		base1, _ = strconv.ParseInt(nodeID1[i:i+2], 16, 10)
		base2, _ = strconv.ParseInt(nodeID2[i:i+2], 16, 10)
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

// CalculateDetailDistance 计算两个节点之间的距离
func CalculateDetailDistance(nodeID1 string, nodeID2 string) []int64 {
	var sourceID []int64

	nodeID1 = nodeID1[2:len(nodeID1)]
	nodeID2 = nodeID2[2:len(nodeID2)]
	// 初始化 sourceID 的长度
	// 比较两个nodeId的异或距离
	var base1, base2 int64
	for i := 0; i < len(nodeID1); i += 2 {
		base1, _ = strconv.ParseInt(nodeID1[i:i+2], 16, 10)
		base2, _ = strconv.ParseInt(nodeID2[i:i+2], 16, 10)
		sourceID = append(sourceID, 1)
		sourceID[i/2] = base1 ^ base2
	}

	return sourceID
}

// MasterAreas 8个特区的参数
var MasterAreas []string = []string{
	"0400000000000000000000000000000000",
	"0420000000000000000000000000000000",
	"0440000000000000000000000000000000",
	"0460000000000000000000000000000000",
	"0480000000000000000000000000000000",
	"04A0000000000000000000000000000000",
	"04C0000000000000000000000000000000",
	"04E0000000000000000000000000000000",
}

// LocateNode 计算输入的结点属于哪个分区，并返回这个结点是否这个区域内的master
// @return area int 分区编号
// @return isMaster bool 是否在这个分区的master区域内
func LocateNode(nodeID string) (int, bool) {
	var distanceGroup [][]int64
	var i, k int
	// 计算与各个分区的距离
	for i = 0; i < len(MasterAreas); i++ {
		distanceGroup = append(distanceGroup, CalculateDetailDistance(nodeID, MasterAreas[i]))
	}
	// 选出最小的距离
	minDistance := distanceGroup[0]
	areaNum := 0
	for i = 1; i < len(distanceGroup); i++ {
		for k = 0; k < len(distanceGroup[i]); k++ {
			if distanceGroup[i][k] == minDistance[k] {
				continue
			}
			if distanceGroup[i][k] > minDistance[k] {
				break
			}
			// 如果比minDistance小，就放这个进minDistance
			minDistance = distanceGroup[i]
			areaNum = i
		}
	}

	if distanceGroup[areaNum][0] <= 1 {
		return areaNum, true
	}

	return areaNum, false
}
