package config

// Load config 包的入口
func Load() (interface{}, interface{}) {
	// 载入控制台参数
	loadArgs()

	// 载入配置文件
	conf, loadErr := loadConfig()
	return conf, loadErr
}
