package discover

// TODO 这些都加到配置里面8
var NEW_BUCKET_COUNT int = 64
var NEW_BUCKET_LENGTH int = 64
var TRIED_BUCKET_COUNT int = 64
var TRIED_BUCKET_LENGTH int = 16

// Bucket 路由桶
type Bucket struct {
	newBucket   map[int]interface{}
	triedBucket map[int]interface{}
}

// Build 初始化路由桶
func (bucket *Bucket) Build() {
	bucket.newBucket = make(map[int]interface{}) // 初始化数组对象本身
	for i := 0; i < NEW_BUCKET_COUNT; i++ {
		bucket.newBucket[i] = make(map[int]interface{}) // 声明第一个维度每个对象都是一个子数组
	}

	bucket.triedBucket = make(map[int]interface{}) // 初始化数组对象本身
	for i := 0; i < TRIED_BUCKET_COUNT; i++ {
		bucket.triedBucket[i] = make(map[int]interface{}) // 声明第一个维度每个对象都是一个子数组
	}
}
