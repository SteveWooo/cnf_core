package discover

// 作为发现服务消息接收的入口
func ReceiveMsg(data interface{}) interface{} {
	body := data.(map[string]interface{})["body"]
	msg := body.(map[string]interface{})["msgJson"]

	// Ping case
	if msg.(map[string]interface{})["type"] == "1" {
		pingErr := ReceivePing(data)
		if pingErr != nil {
			return pingErr
		}
	}

	// Pong case
	if msg.(map[string]interface{})["type"] == "2" {
		pongErr := ReceivePong(data)
		if pongErr != nil {
			return pongErr
		}
	}

	return nil
}

// 接收Ping包的处理逻辑
func ReceivePing(data interface{}) interface{} {
	sourceIP := data.(map[string]interface{})["sourceIP"].(string)
	sourceServicePort := data.(map[string]interface{})["sourceServicePort"].(string)
	shaker.DoPong(sourceIP, sourceServicePort)
	return nil
}

// 接收Pong包的所有处理逻辑
func ReceivePong(data interface{}) interface{} {

	return nil
}
