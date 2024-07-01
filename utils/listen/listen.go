package listen

// 一些多进程处理概念，windows.IOCP，linux.epoll，socket sharding
// 参考自 https://github.com/gogf/greuse
import (
	"net"
	"context"
)

// var listenConfig = net.ListenConfig{}

func Listen(network string, address string) (net.Listener, error) {
	return listenConfig.Listen(context.Background(), network, address)
}