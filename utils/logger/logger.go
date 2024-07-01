package logger

import (
	"os"
	"log"
	"github.com/kNewMo/HttpRequests/utils/file"
)

var logLevel int8
var lf *os.File

// 初始化日志文件
func Log(ll int8, logFile string) {
	logLevel = ll
	// 之前有打开过了，关闭
	if lf != nil {
		lf.Close()
		lf = nil
	}
	if logLevel > 0 {
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, file.FileMode)
		if err == nil {
			log.SetOutput(f)
			lf = f
			// Lshortfile没用，下面都是统一调用，都是本文件
			// log.SetFlags(log.Ldate|log.Ltime|log.Lshortfile)
		}
	}
}

// 普通信息日志
func InfoLogger(v ...interface{}) {
	if (logLevel > 0) && (logLevel <= 1) {
		v = append([]interface{}{"[Info]"}, v...)
		log.Println(v...)
	}
}

// 警告信息日志
func WarningLogger(v ...interface{}) {
	if (logLevel > 0) && (logLevel <= 2) {
		v = append([]interface{}{"[Warning]"}, v...)
		log.Println(v...)
	}
}

// 错误信息日志
func ErrorLogger(v ...interface{}) {
	// 始终输出
	if logLevel <= 3 {
		v = append([]interface{}{"[Error]"}, v...)
		// 主动退出程序
		log.Fatalln(v...)
	}
}