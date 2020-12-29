package net

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"

	"github.com/cnf_core/src/utils/config"
	logger "github.com/cnf_core/src/utils/logger"
	"github.com/cnf_core/src/utils/timer"
)

// GetLog 通过http协议把节点状态Log到前端去
func (cnfNet *CnfNet) GetLog() map[string]interface{} {
	nodeConnectionStatus := cnfNet.nodeConnection.GetStatus()

	netStatus := make(map[string]interface{})
	netStatus["nodeConnectionStatus"] = nodeConnectionStatus
	netStatus["bucketStatus"] = cnfNet.bucket.GetStatus()

	return netStatus
}

// DoLogHTTP 不断发送日志
func (cnfNet *CnfNet) DoLogHTTP() {
	// return
	for {
		timer.Sleep(5000 + rand.Intn(5000))
		client := &http.Client{}
		netStatus := cnfNet.GetLog()

		httpBody := make(map[string]interface{})
		httpBody["nodeID"] = config.ParseNodeID(cnfNet.conf)
		httpBody["netStatus"] = netStatus
		httpBody["number"] = cnfNet.conf.(map[string]interface{})["number"]

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

// DoLogUDP UDP的方式发日志
func (cnfNet *CnfNet) DoLogUDP() {
	return
	for {
		timer.Sleep(10000 + rand.Intn(20000))
		udpAddr, _ := net.ResolveUDPAddr("udp4", "127.0.0.1:8081")
		udpConn, udpConnErr := net.DialUDP("udp", nil, udpAddr)
		if udpConnErr != nil {
			logger.Error(udpConnErr)
		}
		netStatus := cnfNet.GetLog()
		logBody := make(map[string]interface{})
		logBody["nodeID"] = config.ParseNodeID(cnfNet.conf)
		logBody["netStatus"] = netStatus
		logBody["number"] = cnfNet.conf.(map[string]interface{})["number"]

		logBodyJSONByte, _ := json.Marshal(logBody)
		logBodyJSON := string(logBodyJSONByte)
		logBodyJSON = logBodyJSON + "\r\n"

		udpConn.Write([]byte(logBodyJSON))

		udpConn.Close()
	}
}

// DoLogChanel 通过进程内通道发送日志到主线程上，由主线程把日志发到日志中心去
func (cnfNet *CnfNet) DoLogChanel() {
	for {
		timer.Sleep(2000 + rand.Intn(2000))

		netStatus := cnfNet.GetLog()
		logBody := make(map[string]interface{})
		logBody["nodeID"] = config.ParseNodeID(cnfNet.conf)
		logBody["netStatus"] = netStatus
		logBody["number"] = cnfNet.conf.(map[string]interface{})["number"]
		cnfNet.myPublicChanel["logChanel"] <- logBody
	}
}
