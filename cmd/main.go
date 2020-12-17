package main

import (
	cnf "github.com/cnf_core/src"
);

func main(){
	cnf.Build() // 首先把cnf对象构建好, 里面包含了配置文件的引入

	cnf.Run() // 启动对象的所有服务
}