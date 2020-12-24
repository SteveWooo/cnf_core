package config

// Load config 包的入口
func Load() (interface{}, interface{}) {
	// 载入控制台参数
	loadArgs()

	// 载入配置文件
	conf, loadErr := loadConfig(GetArg("configure"))
	return conf, loadErr
}

// LoadByPath 通过路径载入配置
func LoadByPath(path string) (interface{}, interface{}) {
	// 载入控制台参数
	loadArgs()

	// 载入配置文件
	conf, loadErr := loadConfig(path)

	return conf, loadErr
}
