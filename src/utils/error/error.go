package error

import (
	"fmt"

	logger "github.com/cnf_core/src/utils/logger"
)

type Error struct {
	code      int
	message   string
	originErr interface{}
}

// GetMessage 获取错误信息
func (e *Error) GetMessage() string {
	return e.message
}

func typeof(obj interface{}) string {
	return fmt.Sprintf("%T", obj)
}

// New 创建一个错误对象
func New(info map[string]interface{}) *Error {
	var error Error
	// logger.Error(info["message"].(string))
	if info["message"] != nil {
		error.message = info["message"].(string)
	}

	if info["originErr"] != nil {
		// error.originErr = info["originErr"].( typeof(info["originErr"]) )
		logger.Error(info["originErr"])
	}

	return &error
}
