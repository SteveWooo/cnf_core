package timer

import "time"

// Now 返回当前时间戳 毫秒
func Now() int64 {
	return time.Now().UnixNano() / 1e6
}

// Sleep 睡眠封装，输入毫秒
func Sleep(t int) {
	time.Sleep(time.Duration(t) * time.Millisecond)
}
