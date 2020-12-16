package net

import (
	// error "github.com/cnf_core/src/utils/error"
	// logger "github.com/cnf_core/src/utils/logger"
	nodeBucket "github.com/cnf_core/src/modules/net/services/nodeBucket"
)

/**
 * 定义netData全局变量的
 */
func init(){
	
}

func Build(){
	// 路由桶的初始化
	nodeBucket.Build()
}