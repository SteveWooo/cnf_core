package config

import (
	json "encoding/json"
	ioutil "io/ioutil"
	os "os"
	filepath "path/filepath"

	error "github.com/cnf_core/src/utils/error"
	"github.com/cnf_core/src/utils/sign"
)

var config interface{}

// 载入全局配置
func loadConfig() (interface{}, interface{}) {
	configFilePath := GetArg("configure")
	if configFilePath == "" {
		// logger.Error("missing console argv: configure")
		return nil, error.New(map[string]interface{}{
			"message": "缺乏控制台参数：configure",
		})
	}

	// 判断输入参数是绝对路径还是相对路径, 相对路径的起点是可执行文件的当前目录
	if filepath.IsAbs(configFilePath) {
		configFilePath = filepath.Clean(configFilePath)
	} else {
		// 获取当前目录
		executablePath, exError := os.Executable()
		if exError != nil {
			// logger.Error(exError)
			return nil, error.New(map[string]interface{}{
				"message":   "获取当前执行文件目录失败",
				"originErr": exError,
			})
		}
		executableDir := filepath.Dir(executablePath)

		// 获得当前目录下, 取得规整化配置文件路径
		configFilePath = filepath.Clean(executableDir) + filepath.Clean("\\") + filepath.Clean(configFilePath)
	}

	// 然后利用配置文件路径, 读取配置文件出来
	configFile, _ := ioutil.ReadFile(configFilePath)

	// 配置文件map存放地儿
	deCodeError := json.Unmarshal(configFile, &config)

	if deCodeError != nil {
		// logger.Error(deCodeError)
		return nil, error.New(map[string]interface{}{
			"message":   "JSON解析失败",
			"originErr": deCodeError,
		})
	}

	return config, nil
}

// GetConfig 整体获取配置map
func GetConfig() interface{} {
	return config
}

// GetNetSeed 获取配置文件中的种子
func GetNetSeed() []interface{} {
	conf := GetConfig()
	confNet := conf.(map[string]interface{})["net"]
	seed := confNet.(map[string]interface{})["seed"].([]interface{})
	return seed
}

// GetNetConf 获取网络配置
func GetNetConf() interface{} {
	conf := GetConfig()
	confNet := conf.(map[string]interface{})["net"]
	return confNet
}

// GetNodeID 解析出ID
func GetNodeID() string {
	conf := GetConfig()
	confNet := conf.(map[string]interface{})["net"]
	nodeID := sign.GetPublicKey(confNet.(map[string]interface{})["localPrivateKey"].(string))

	return nodeID
}

// SetConfig 设置配置项
func SetConfig(conf interface{}) {
	config = conf
}
