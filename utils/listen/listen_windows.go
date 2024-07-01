package listen

import (
	"net"
	"syscall"
	"github.com/kNewMo/HttpRequests/utils/logger"
)

var listenConfig = net.ListenConfig{
	Control: Control,
}

func Control(network string, address string, c syscall.RawConn) error {
	return c.Control(func(fd uintptr) {
		// windows下可以同时监听，但是请求只会转到第一个启动的，如果第一个退出后则第二个自动接管，目前等于无法用，找了很多文档都没找到相关的方法，有人知道怎么实现可以麻烦告知我一下
		err := syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
		if err != nil {
			logger.WarningLogger("无法启用端口复用。错误：", err)
		}
	})
}