package api

import (
	"time"
	"strconv"
	"math/rand"
	"crypto/md5"
	"encoding/hex"
	"github.com/kNewMo/HttpRequests/utils/file"
)

// 创建一个月日前缀的当前时间随机的md5的UUID
func uuid() (string) {
	hash := md5.Sum([]byte(strconv.FormatInt(time.Now().UnixNano() + rand.Int63(), 10)))
	// hash是[16]byte类型，需要切片转换
	uuid := hex.EncodeToString(hash[:])
	return time.Now().Format("0102") + "-" + uuid[0:8] + "-" + uuid[8:12] + "-" + uuid[12:16] + "-" + uuid[16:20] + "-" + uuid[20:32]
}

// 把文件按uuid分片到3级目录，先按月日前缀拆分了，这样方便以后按日期清除批量请求未正常下载的数据
func uuidPath(uuid string) (string) {
	// 正常已经可以支持4294967296的文件数了，个人是基本达不到这个量的
	return uuid[0:4] + file.PathSeparator + uuid[5:7] + file.PathSeparator + uuid[7:9] + file.PathSeparator + uuid[9:11] + file.PathSeparator
}