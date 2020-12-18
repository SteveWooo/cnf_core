package discover

var NEW_BUCKET_COUNT int = 16
var NEW_BUCKET_LENGTH int = 64
var TRIED_BUCKET_COUNT int = 16
var TRIED_BUCKET_LENGTH int = 16

/**
 * 路由桶
 */
type Bucket struct {
	newBucket   map[int]interface{}
	triedBucket map[int]interface{}
}

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
