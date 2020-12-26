package net

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/cnf_core/src/utils/config"
	"github.com/cnf_core/src/utils/timer"
)

// GetLog 通过http协议把节点状态Log到前端去
func (cnfNet *CnfNet) GetLog() map[string]interface{} {
	nodeConnectionStatus := cnfNet.nodeConnection.GetStatus()

	netStatus := make(map[string]interface{})
	netStatus["nodeConnectionStatus"] = nodeConnectionStatus

	return netStatus
}

// DoLogHTTP 不断发送日志
func (cnfNet *CnfNet) DoLogHTTP() {
	// return
	for {
		timer.Sleep(5000)
		client := &http.Client{}
		netStatus := cnfNet.GetLog()

		httpBody := make(map[string]interface{})
		httpBody["nodeID"] = config.ParseNodeID(cnfNet.conf)
		httpBody["netStatus"] = netStatus

		httpBodyJSON, _ := json.Marshal(httpBody)
		req, _ := http.NewRequest("POST", "http://localhost:8081/api/update_node_status", bytes.NewReader(httpBodyJSON))
		resp, doErr := client.Do(req)
		if doErr != nil {
			// logger.Debug(doErr)
			continue
		}
		body, _ := ioutil.ReadAll(resp.Body)
		bodyStr := string(body)
		if bodyStr != "" {
		}

	}
}
