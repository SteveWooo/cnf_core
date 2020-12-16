package config

// config 包的入口
func Load(){
	// 载入控制台参数
	loadArgs()

	// 载入配置文件
	loadConfig()
}