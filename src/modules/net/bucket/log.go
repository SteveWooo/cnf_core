package bucket

// GetStatus 获取节点当前状态
func (bucket *Bucket) GetStatus() map[string]interface{} {
	serviceStatus := make(map[string]interface{})

	// 处理连接情况
	var newBucket []string
	var triedBucket []string
	for i := 0; i < len(bucket.newBucket); i++ {
		nodeGroup := bucket.newBucket[i]
		for k := 0; k < len(nodeGroup); k++ {
			node := nodeGroup[k]
			if node == nil {
				continue
			}
			newBucket = append(newBucket, node.GetNodeID())
		}
	}

	for i := 0; i < len(bucket.triedBucket); i++ {
		nodeGroup := bucket.triedBucket[i]
		for k := 0; k < len(nodeGroup); k++ {
			node := nodeGroup[k]
			if node == nil {
				continue
			}
			triedBucket = append(triedBucket, node.GetNodeID())
		}
	}

	serviceStatus["newBucket"] = newBucket
	serviceStatus["triedBucket"] = triedBucket

	return serviceStatus
}
