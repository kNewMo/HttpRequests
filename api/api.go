package api

import (
	"io"
	"os"
	"math"
	"time"
	"regexp"
	"strconv"
	"strings"
	"net/url"
	"net/http"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"compress/gzip"
	"encoding/json"
	"github.com/kNewMo/HttpRequests/utils/file"
	"github.com/kNewMo/HttpRequests/utils/logger"
	// "fmt"
)

var debug bool
var version string
// 用chan做并发限制，goroutine就不限制了
var sc chan bool
var mc chan bool
var path string
// 校验用key字典
var keys map[string]string
// 初始化一个，避免每次初始消耗
var pathRegExp *regexp.Regexp

type requestData struct {
	Index int `json:"index"`
	Method string `json:"method"`
	URL string `json:"url"`
	Headers []string `json:"headers"`
	Body string `json:"body"`
	Proxy string `json:"proxy"`
	Timeout int `json:"timeout"`
	SkipSSLVerify bool `json:"skipsslverify"`
	JA3 string `json:"ja3"`
	JA4 string `json:"ja4"`
}

type multiRequestData struct {
	Data string `json:"data"`
	Requests []requestData `json:"requests"`
}

type responseData struct {
	Index int
	Result *remoteData
}

type remoteData struct {
	Status int `json:"status"`
	Error string `json:"error"`
	headers map[string]interface{} `json:"-"`
	Headers []string `json:"headers"`
	Content []byte `json:"content"`
}

type multiResponseData struct {
	Data string `json:"data"`
	Result []*remoteData `json:"result"`
}

type defaultResultError struct {
	Code int64 `json:"code"`
	Message string `json:"message"`
}

type defaultResult struct {
	Error *defaultResultError `json:"error"`
	Data interface{} `json:"data"`
}

// 输出文件
func outputFile(w http.ResponseWriter, r *http.Request, name string) {
	f, err := os.Open(name)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer f.Close()
	http.ServeContent(w, r, name, time.Now(), f)
}

// 输出JSON
func outputJson(w http.ResponseWriter, v interface{}, zip bool) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	// 使用gzip输出
	if zip {
		w.Header().Add("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		json.NewEncoder(gz).Encode(&v)
		gz.Close()
	// 直接输出
	} else {
		encoder := json.NewEncoder(w)
		encoder.Encode(&v)
	}
}

// 输出错误
func outputError(w http.ResponseWriter, code int64, message string, zip bool) {
	outputJson(w, &defaultResult{&defaultResultError{code, message}, nil}, zip)
}

// 输出结果
func outputResult(w http.ResponseWriter, data interface{}, zip bool) {
	outputJson(w, &defaultResult{&defaultResultError{0, ""}, data}, zip)
}

// 简单校验签名
func sign(w http.ResponseWriter, r *http.Request) (bool) {
	// 输出版本
	w.Header().Add("Http-Requests", version)
	// 调试时直接返回，不用加密，非调试时注释掉
	if debug {
		return true
	}
	t := r.FormValue("t")
	k := r.FormValue("k")
	s := r.FormValue("s")
	it, _ := strconv.ParseInt(t, 10, 64)
	// 时差在5分钟内
	if math.Abs(float64(it - time.Now().Unix())) < 300 {
		key, ok := keys[k]
		// key 存在
		if ok {
			hash := md5.Sum([]byte(t + key))
			// hash是[16]byte类型，需要切片转换
			if s == hex.EncodeToString(hash[:]) {
				return true
			} else {
				w.Header().Add("HR-Error-Code", "103")
			}
		} else {
			w.Header().Add("HR-Error-Code", "102")
		}
	} else {
		w.Header().Add("HR-Error-Code", "101")
	}
	w.Header().Add("HR-Error-Message", "Sign Error.")
	w.WriteHeader(http.StatusForbidden)
	return false
}

// 发起一个请求
func request(req requestData, c chan bool, dc chan responseData) (*remoteData) {
	var clientReq *http.Request
	var tlsConfig *tls.Config
	var err error
	// 目前只支持post和get模式
	if req.Method == "POST" {
		// 有body
		if len(req.Body) > 0 {
			clientReq, err = http.NewRequest("POST", req.URL, strings.NewReader(req.Body))
		} else {
			clientReq, err = http.NewRequest("POST", req.URL, nil)
		}
	} else {
		clientReq, err = http.NewRequest("GET", req.URL, nil)
	}
	rd := &remoteData{}
	// 没有错误
	if err == nil {
		transport := &http.Transport{}
		// 有配置的ja3，ja4，跳过证书，则初始化
		if (len(req.JA3) > 0) || (len(req.JA4) > 0) || req.SkipSSLVerify {
			tlsConfig = &tls.Config{}
		}
		// 有自定义ja3
		if len(req.JA3) > 0 {
			err = tlsConfigJA3(tlsConfig, req.JA3)
			// 出错了，跳到结束，统一处理返回
			if err != nil {
				rd.Error = err.Error()
				goto end
			}
		}
		// 有自定义ja4
		if len(req.JA4) > 0 {
			err = tlsConfigJA4(tlsConfig, req.JA4)
			if err != nil {
				// 出错了，跳到结束，统一处理返回
				rd.Error = err.Error()
				goto end
			}
		}
		if tlsConfig != nil {
			// 跳过ssl证书验证
			if req.SkipSSLVerify {
				tlsConfig.InsecureSkipVerify = true
			}
			transport.TLSClientConfig = tlsConfig
			// h2协议的，单独设置下
			if (len(tlsConfig.NextProtos) > 0) && (tlsConfig.NextProtos[0] == "h2") {
				transport.ForceAttemptHTTP2 = true
			}
		}
		// 有使用代理，支持http和socks5代理，http://...，socks5://...
		if len(req.Proxy) > 0 {
			proxy, err := url.Parse(req.Proxy)
			if err == nil {
				transport.Proxy = http.ProxyURL(proxy)
			}
		}
		// 处理头部
		for i, _ := range req.Headers {
			before, after, _ := strings.Cut(req.Headers[i], ": ")
			clientReq.Header.Add(before, after)
		}
		client := &http.Client{}
		client.Transport = transport
		// 请求超时，默认55秒
		if req.Timeout == 0 {
			req.Timeout = 55
		}
		client.Timeout = time.Duration(req.Timeout) * time.Second
		resp, err := client.Do(clientReq)
		if err == nil {
			rd.Status = resp.StatusCode
			if len(resp.Header) > 0 {
				// 单请求模式，header是用http.ResponseWriter.Header()输出的，用字典
				if dc == nil {
					// 这里需要初始化下
					rd.headers = map[string]interface{}{}
					for key, _ := range resp.Header {
						if len(resp.Header[key]) == 1 {
							rd.headers[key] = resp.Header[key][0]
						} else {
							rd.headers[key] = resp.Header[key]
						}
					}
				// 多请求模式
				} else {
					for key, _ := range resp.Header {
						if len(resp.Header[key]) == 1 {
							rd.Headers = append(rd.Headers, key + ": " + resp.Header[key][0])
						} else {
							for i, _ := range resp.Header[key] {
								rd.Headers = append(rd.Headers, key + ": " + resp.Header[key][i])
							}
						}
					}
				}
			}
			rd.Content, err = io.ReadAll(resp.Body)
			resp.Body.Close()
			// 读取出错了，也返回异常
			if err != nil {
				rd.Status = 0
				rd.Error = err.Error()
			}
		} else {
			rd.Error = err.Error()
		}
		// 关闭连接，暂时不用连接池
		client.CloseIdleConnections()
	} else {
		rd.Error = err.Error()
	}
	end:
	// 请求完成
	<-c
	if dc == nil {
		return rd
	} else {
		dc <- responseData{req.Index, rd}
		return nil
	}
}

// 发起多个请求
func requests(reqID string, reqs multiRequestData) {
	rl := len(reqs.Requests)
	dc := make(chan responseData, rl)
	for i := 0; i < rl; i++ {
		// 判断是否可以发起新的请求，通道满了就会阻塞
		mc<- true
		reqs.Requests[i].Index = i
		go request(reqs.Requests[i], mc, dc)
	}
	result := make([]*remoteData, rl)
	for i := 0; i < rl; i++ {
		data := <-dc
		result[data.Index] = data.Result
	}
	gf, err := file.Create(path + uuidPath(reqID) + reqID + file.PathSeparator + "data.gz")
	if err != nil {
		logger.WarningLogger("创建多请求任务结果错误:", reqID, err)
		return
	}
	gz := gzip.NewWriter(gf)
	// 如果需要预设的文件名
	// gz.Header.Name = reqID + ".json"
	mrd := multiResponseData{
		Data: reqs.Data,
		Result: result,
	}
	err = json.NewEncoder(gz).Encode(&mrd)
	if err != nil {
		logger.WarningLogger("写入多请求任务结果错误:", reqID, err)
	}
	gz.Close()
	gf.Close()
	// 创建完成标记
	cf, _ := file.Create(path + uuidPath(reqID) + reqID + file.PathSeparator + "completed")
	cf.Close()
}

// 初始化正则
func initRegExp() {
	if pathRegExp == nil {
		pathRegExp = regexp.MustCompile("[^a-z0-9\\-]+")
	}
}

// 设置调试
func SetDebug(d bool) {
	debug = d
}

// 设置版本
func SetVersion(ver string) {
	version = ver
}

// 设置最大并发
func SetLimit(s int32, m int32) {
	sc = make(chan bool, s)
	mc = make(chan bool, m)
}

// 设置多请求数据目录
func SetPath(p string) {
	path = p
	initRegExp()
}

// 设置签名key
func SetKeys(k map[string]string) {
	keys = k
}

// 默认请求
func Default(w http.ResponseWriter, r *http.Request) {
	// 都返回404
	w.WriteHeader(http.StatusNotFound)
}

// 默认请求
func Status(w http.ResponseWriter, r *http.Request) {
	// 验证
	if sign(w, r) {
		outputResult(w, os.Getpid(), false)
	}
}

// 单请求模式
func Single(w http.ResponseWriter, r *http.Request) {
	// 验证
	if sign(w, r) {
		req := requestData{}
		// json数据
		if r.Header.Get("Content-Type") == "application/json" {
			err := json.NewDecoder(r.Body).Decode(&req)
			if err != nil {
				w.Header().Add("HR-Error-Code", "201")
				w.Header().Add("HR-Error-Message", err.Error())
				// 这里用 418 这个正常服务器都不会使用的状态返回，因为 Single 接口基本就是个转发接口，直接将目标返回结果返回给客户端，让客户端的从直接请求迁移到本接口是改造相对来说最小，后面同理
				w.WriteHeader(http.StatusTeapot)
				return
			}
		} else {
			// 要调用一次PostFormValue后PostForm才有值
			_ = r.PostFormValue("headers")
			headers := r.PostForm["headers"]
			timeout, _ := strconv.Atoi(r.PostFormValue("timeout"))
			var skipsslverify bool
			if r.PostFormValue("skipsslverify") == "1" {
				skipsslverify = true
			}
			req.Method = r.PostFormValue("method")
			req.URL = r.PostFormValue("url")
			req.Headers = headers
			req.Body = r.PostFormValue("body")
			req.Proxy = r.PostFormValue("proxy")
			req.Timeout = timeout
			req.SkipSSLVerify = skipsslverify
			req.JA3 = r.PostFormValue("ja3")
			req.JA4 = r.PostFormValue("ja4")
		}
		// 做一些基础检测，不正常的请求就直接返回了
		if req.URL == "" {
			w.Header().Add("HR-Error-Code", "202")
			w.Header().Add("HR-Error-Message", "Bad Request")
			w.WriteHeader(http.StatusTeapot)
			return
		}
		// 可以看看用switch来顺序写入两个队列，达到一个优先队列？保证一定的单次请求数，避免被批量请求占满？
		var result *remoteData
		// 用selected可以判断是否缓冲满了，优先使用single限制
		select {
		case sc<- true:
			result = request(req, sc, nil)
		// single队列满了
		default:
			select {
			case mc<- true:
				result = request(req, mc, nil)
			// 满了直接返回503
			default:
				w.Header().Add("HR-Error-Code", "203")
				w.Header().Add("HR-Error-Message", "Too many Requests")
				w.WriteHeader(http.StatusTeapot)
				return
			}
		}
		// 请求正常
		if result.Status > 0 {
			for key, _ := range result.headers {
				switch result.headers[key].(type) {
				case string:
					w.Header().Add(key, result.headers[key].(string))
				case []string:
					header := result.headers[key].([]string)
					for i, _ := range header {
						w.Header().Add(key, header[i])
					}
				}
			}
			w.WriteHeader(result.Status)
			w.Write(result.Content)
		// 请求异常
		} else {
			w.Header().Add("HR-Error-Code", "204")
			w.Header().Add("HR-Error-Message", result.Error)
			w.WriteHeader(http.StatusTeapot)
		}
	}
}

// 多请求模式，返回一个队列id
func Multi(w http.ResponseWriter, r *http.Request) {
	// 验证
	if sign(w, r) {
		reqs := multiRequestData{}
		var err error
		// json数据
		if r.Header.Get("Content-Type") == "application/json" {
			err = json.NewDecoder(r.Body).Decode(&reqs)
		} else {
			reqs.Data = r.PostFormValue("data")
			err = json.Unmarshal([]byte(r.PostFormValue("requests")), &reqs.Requests)
		}
		// json 解析错误
		if err != nil {
			outputError(w, 418, "", false)
			return
		}
		reqID := uuid()
		// 创建任务目录
		err = os.MkdirAll(path + uuidPath(reqID) + reqID + file.PathSeparator, file.FileMode)
		if err != nil {
			logger.WarningLogger("创建多请求任务目录错误:", reqID, err)
			outputError(w, 500, "", false)
			return
		}
		go requests(reqID, reqs)
		outputResult(w, reqID, false)
	}
}

// 下载多请求的结果
func Download(w http.ResponseWriter, r *http.Request) {
	// 验证
	if sign(w, r) {
		id := r.PostFormValue("id")
		// 替换掉不允许的字符，避免安全问题，不做异常返回了
		id = pathRegExp.ReplaceAllString(id, "")
		fp := path
		if len(id) == 41 {
			fp += uuidPath(id) + id
			// 目录存在
			if file.Exist(fp) {
				// 已经处理完成
				if file.Exist(fp + file.PathSeparator + "completed") {
					outputFile(w, r, fp + file.PathSeparator + "data.gz")
				// 还在处理中
				} else {
					w.WriteHeader(http.StatusAccepted)
				}
			// 已经删除或者请求不存在
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

// 完成下载请求，通知删除结果数据
func Completed(w http.ResponseWriter, r *http.Request) {
	// 验证
	if sign(w, r) {
		id := r.PostFormValue("id")
		// 替换掉不允许的字符，避免安全问题，不做异常返回了
		id = pathRegExp.ReplaceAllString(id, "")
		fp := path
		if len(id) == 41 {
			fp += uuidPath(id) + id
			// 目录存在
			if file.Exist(fp) {
				// 已经处理完成
				if file.Exist(fp + file.PathSeparator + "completed") {
					os.RemoveAll(fp + file.PathSeparator)
					outputResult(w, id, false)
				// 还在处理中，不支持删除
				} else {
					outputError(w, 202, "", false)
				}
			// 已经删除或者请求不存在
			} else {
				outputError(w, 404, "", false)
			}
		} else {
			outputError(w, 404, "", false)
		}
	}
}