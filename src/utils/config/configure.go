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

func GetConfig() interface{} {
	return config
}

func GetNodeId() string {
	conf := GetConfig()
	confNet := conf.(map[string]interface{})["net"]
	nodeId := sign.GetPublicKey(confNet.(map[string]interface{})["localPrivateKey"].(string))

	return nodeId
}

func SetConfig(conf interface{}) {
	config = conf
}
