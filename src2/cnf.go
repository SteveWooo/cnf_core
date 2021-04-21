package src2

import (
	logger "github.com/cnf_core/src2/utils/logger"
)

type Cnf struct {
	conf interface{}
}

func (cnf *Cnf) Build() {
	logger.Debug("hello")
}