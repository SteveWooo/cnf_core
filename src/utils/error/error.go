package error

import (
	"fmt"
	logger "github.com/cnf_core/src/utils/logger"
)

type Error struct {
	code int
	message string
	originErr interface{}
}

func typeof(obj interface{}) string{
	return fmt.Sprintf("%T", obj)
}

func New(info map[string]interface{}) Error{
	var error Error
	if info["message"] != nil {
		error.message = info["message"].(string)
	}

	if info["originErr"] != nil {
		// error.originErr = info["originErr"].( typeof(info["originErr"]) )
		logger.Error(info["originErr"])
	}

	return error
}