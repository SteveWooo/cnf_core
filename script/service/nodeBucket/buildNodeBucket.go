package nodebucket

import (
	discoverModel "github.com/cnf_core/src/modules/net/models/discover"
	"github.com/cnf_core/src/utils/config"
)

// 管理可用节点的路由桶
var bucket discoverModel.Bucket

// Build 为了初始化全局变量bucket
func Build() {
	bucket.Build(map[string]interface{}{
		"seeds": config.GetNetSeed(),
	})
}
