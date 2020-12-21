package config

import (
	os "os"
)

var args map[string]string

// loadArgs 此函数负责遍历载入命令行参数, 规定使用:
// - 名称缩写
// -- 全称
func loadArgs() {
	args = make(map[string]string)

	for index, value := range os.Args {
		// 启动文件命令不需要理会
		if index == 0 {
			continue
		}

		if len(value) > 2 && value[0:2] == "--" && os.Args[index+1] != "" {
			key := value[2:]
			val := os.Args[index+1]
			setArg(key, val)
			continue
		}

		if len(value) > 1 && value[0:1] == "-" && os.Args[index+1] != "" {
			key := value[1:]
			val := os.Args[index+1]
			setArg(key, val)
			continue
		}
	}
}

// setArg 写入args变量中
func setArg(key string, value string) {
	args[key] = value
}

// GetArg 获取对应的控制台变量
func GetArg(key string) string {
	return args[key]
}
