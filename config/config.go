package config

import (
	sl "log"
	"runtime"
	"strings"
	"github.com/kNewMo/HttpRequests/utils/file"
)

type limit struct {
	// single队列用来保证单个请求
	Single int32 `json:"single"`
	// multi用来处理批量请求，批量请求可能容易满
	Multi int32 `json:"multi"`
}

type log struct {
	Level int8 `json:"level"`
	Path string `json:"path"`
}

type config struct {
	IP string `json:"ip"`
	Port uint16 `json:"port"`
	Worker uint16 `json:"worker"`
	Limit *limit `json:"limit"`
	Path string `json:"path"`
	// 推荐算法，md5(ixspy+md5(ae))，md5(ixspy+md5(et))，md5(ixspy+md5(tt))
	Keys map[string]string `json:"keys"`
	Log *log `json:"log"`
}

func Config() (*config) {
	root, err := file.Root()
	if err != nil {
		sl.Fatalln("无法获取运行目录：", err)
	}
	// 没有传config参数，则从当前目录读取
	path := root + "config.json"
	var c *config
	err = file.LoadJsonToStruct(path, &c)
	if err != nil {
		sl.Fatalln("读取配置文件 ", path, " 失败：", err)
	}
	// 没有设置端口，默认180端口
	if c.Port == 0 {
		c.Port = 180
	}
	if c.Worker == 0 {
		c.Worker = uint16(runtime.NumCPU())
	}
	if c.Limit.Single == 0 {
		c.Limit.Single = 256
	}
	if c.Limit.Multi == 0 {
		c.Limit.Multi = 256
	}
	// 没有设置多请求数据目录，默认当前目录下的result目录
	if c.Path == "" {
		c.Path = root + "result" + file.PathSeparator
	} else {
		// 如果不是linux绝对路径也不是windows绝对路径，就是是相对路径
		if (c.Path[0:1] != "/") && (strings.Index(c.Path, ":") == -1 ) {
			c.Path = root + c.Path
		}
	}
	// 不是目录，补全
	if c.Path[len(c.Path) - 1:] != file.PathSeparator {
		c.Path += file.PathSeparator
	}
	// 如果是linux，才启用多进程，目前先只判断linux，还有其他的freebsd，openbsd，netbsd，darwin等似乎也支持，目前我们也不太需要
	// 不是linux，先强制只用1个woker，因为没有测试过，避免意外
	// if runtime.GOOS != "linux" {
	// 	c.Worker = 1
	// }
	// 检测日志配置
	if c.Log.Path == "" {
		c.Log.Path = root
	}
	return c
}