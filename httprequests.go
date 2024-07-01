package main

import (
	"io"
	"os"
	"sync"
	"time"
	"bytes"
	"context"
	"strconv"
	"syscall"
	"os/exec"
	"net/http"
	"os/signal"
	"crypto/md5"
	"encoding/hex"
	"github.com/kNewMo/HttpRequests/api"
	"github.com/kNewMo/HttpRequests/config"
	"github.com/kNewMo/HttpRequests/utils/listen"
	"github.com/kNewMo/HttpRequests/utils/logger"
	// "fmt"
)

// 调试时直接启动http服务，方便调试http服务
const DEBUG = false
const VERSION = "0.91"

var isworker bool
var server *http.Server
var wg sync.WaitGroup

func main() {
	cfg := config.Config()
	// 需要保存日志
	if cfg.Log.Level > 0 {
		logger.Log(cfg.Log.Level, cfg.Log.Path + "log.log")
	}
	// 调试模式，或判断是否是子进程
	if DEBUG || ((len(os.Args) > 1) && (os.Args[1] == "worker")) {
		if DEBUG {
			api.SetDebug(DEBUG)
			logger.InfoLogger(os.Getpid(), "调试模式，将以 Worker 启动。")
		}
		isworker = true
	}
	worker := int(cfg.Worker)
	// 只有一个，直接启动
	if worker == 1 {
		logger.InfoLogger(os.Getpid(), "只配置了一个 Worker，直接以 Worker 启动。")
		isworker = true
	}
	// worker 模式，启动http服务
	if isworker {
		logger.InfoLogger(os.Getpid(), "Woker 启动。")
		api.SetVersion(VERSION)
		api.SetLimit(cfg.Limit.Single, cfg.Limit.Multi)
		api.SetPath(cfg.Path)
		api.SetKeys(cfg.Keys)
		go httpServer(cfg.IP + ":" + strconv.Itoa(int(cfg.Port)))
	// 主进程，启动子进程
	} else {
		logger.InfoLogger(os.Getpid(), "Master 启动。")
		wg = sync.WaitGroup{}
		wg.Add(worker)
		for i := 0; i < worker; i++ {
			go startWorker()
		}
		// 5秒后测试是否启动成功
		time.Sleep(5 * time.Second)
		client := &http.Client{}
		if cfg.IP == "" {
			cfg.IP = "0.0.0.0"
		}
		// 测试多次，基本确保能随机到
		worker += worker
		wsb := []byte{}
		// multicast
		workerIsStart := false
		multicastWorker := false
		// 正常测试肯定能在时差内完成，不重复计算了
		t := strconv.FormatInt(time.Now().Unix(), 10)
		k := ""
		key := ""
		for k, _ = range cfg.Keys {
			key = cfg.Keys[k]
			break
		}
		hash := md5.Sum([]byte(t + key))
		for i := 0; i < worker; i++ {
			resp, err := client.Get("http://" + cfg.IP + ":" + strconv.Itoa(int(cfg.Port)) + "/status?t=" + t + "&k=" + k + "&s=" + hex.EncodeToString(hash[:]))
			// 失败了下一次再试
			if err != nil {
				continue
			}
			rb, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			// 关闭连接，要不然会在一个已经打开的长连接中重复请求
			client.CloseIdleConnections()
			// 失败了下一次再试
			if err != nil {
				continue
			}
			// 第一次赋值
			if len(wsb) == 0 {
				workerIsStart = true
				wsb = rb
			// 如果与第一次值不同,则说明支持同端口监听
			} else if bytes.Compare(wsb, rb) != 0 {
				logger.InfoLogger("系统支持端口复用，目前多个 Woker 有效工作。")
				multicastWorker = true
				break
			}
		}
		if workerIsStart {
			if !multicastWorker {
				logger.WarningLogger("系统似乎不支持端口复用，目前只有一个 Woker 有效工作。")
			}
		} else {
			logger.ErrorLogger("Woker 似乎全启动失败。")
		}
	}
	//创建监听退出chan
	c := make(chan os.Signal)
	//监听指定信号 ctrl+c kill
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	for s := range c {
		switch s {
		case syscall.SIGINT, syscall.SIGTERM:
			// logger.InfoLogger(os.Getpid(), "关闭信号。")
			// 子进程，停止http服务
			if isworker {
				err := server.Shutdown(context.TODO())
				if err != nil {
					logger.WarningLogger("HTTP Shutdown:", err)
				}
				logger.InfoLogger(os.Getpid(), "Woker 关闭。")
			// 主进程
			} else {
				// 等待子进程退出
				wg.Wait()
				logger.InfoLogger(os.Getpid(), "Master 关闭。")
			}
			os.Exit(0)
		default:
			logger.InfoLogger(os.Getpid(), "未知信号：", s)
		}
	}
}

// 启动子进程
func startWorker() {
	path, err := os.Executable()
	if err != nil {
		logger.InfoLogger("启动 Worker 失败，无法获取程序路径", err)
	}
	cmd := exec.Command(path, "worker")
	err = cmd.Start()
	if err != nil {
		logger.InfoLogger("启动 Worker 失败：", err)
	}
	logger.InfoLogger(cmd.Process.Pid, "启动 Worker。")
	cmd.Wait()
	wg.Done()
	logger.InfoLogger(cmd.Process.Pid, "Worker 已关闭。")
}

// 启动http服务，如果需要限制并发，可以用golang.org/x/net/netutil，LimitListener
func httpServer(address string) {
	http.HandleFunc("/", api.Default)
	http.HandleFunc("/status", api.Status)
	http.HandleFunc("/single", api.Single)
	http.HandleFunc("/multi", api.Multi)
	http.HandleFunc("/download", api.Download)
	http.HandleFunc("/completed", api.Completed)
	server = &http.Server{}
	listener, err := listen.Listen("tcp", address)
	if (err != nil) {
		logger.ErrorLogger("Listen Error:", err)
	}
	defer listener.Close()
	err = server.Serve(listener)
	// 上面已经被阻塞了，在Shutdown之后也会进入执行，或者是在监听失败时执行
	if (err != nil) && (err != http.ErrServerClosed) {
		// 启动失败，置为未启动
		logger.ErrorLogger("HTTP Serve Error:", err)
	}
}