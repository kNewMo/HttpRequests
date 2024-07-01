package listen

import (
	"net"
	"syscall"
	"github.com/kNewMo/HttpRequests/utils/logger"
)

var listenConfig = net.ListenConfig{
	Control: Control,
}

const (
	SO_REUSEPORT = 0xf
)

func Control(network string, address string, c syscall.RawConn) error {
	return c.Control(func(fd uintptr) {
		// linux下测试可行，可以多个进程监听一个，自动分配请求量，只需要SO_REUSEPORT就好了，不需要SO_REUSEADDR，有可能不同cpu的linux不一样？未测试
		err := syscall.SetsockoptInt(handle(fd), syscall.SOL_SOCKET, SO_REUSEPORT, 1)
		if err != nil {
			logger.WarningLogger("无法启用端口复用。错误：", err)
		}
	})
}

func handle(fd uintptr) int {
	return int(fd)
}