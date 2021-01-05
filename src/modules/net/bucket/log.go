package bucket

// GetStatus 获取节点当前状态
func (bucket *Bucket) GetStatus() map[string]interface{} {
	serviceStatus := make(map[string]interface{})

	// 处理连接情况
	var newBucket []string
	var triedBucket []string
	bucket.newBucketLock <- true
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
	<-bucket.newBucketLock

	bucket.triedBucketLock <- true
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
	<-bucket.triedBucketLock

	bucket.nodeCacheLock <- true
	for i := 0; i < len(bucket.nodeCache); i++ {
		newBucket = append(newBucket, bucket.nodeCache[i].GetNodeID())
	}

	// serviceStatus["newBucket"] = newBucket
	// serviceStatus["triedBucket"] = triedBucket

	serviceStatus["newBucket"] = newBucket

	<-bucket.nodeCacheLock
	return serviceStatus
}
