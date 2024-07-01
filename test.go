package main

import (
	"io"
	"os"
	"log"
	"time"
	"strconv"
	"runtime"
	"net/url"
	"os/exec"
	"net/http"
	"crypto/md5"
	"encoding/hex"
	"compress/gzip"
	"encoding/json"
	"github.com/kNewMo/HttpRequests/config"
	"github.com/kNewMo/HttpRequests/utils/file"
	// "fmt"
)

type defaultResultError struct {
	Code int64 `json:"code"`
	Message string `json:"message"`
}

type defaultResult struct {
	Error *defaultResultError `json:"error"`
	Data interface{} `json:"data"`
}

type remoteData struct {
	Status int `json:"status"`
	Error string `json:"error"`
	Headers []string `json:"headers"`
	Content []byte `json:"content"`
}

type multiData struct {
	Data string `json:"data"`
	Result []*remoteData `json:"result"`
}

func main() {
	// 载入配置
	cfg := config.Config()
	client := &http.Client{}
	if cfg.IP == "" {
		cfg.IP = "0.0.0.0"
	}
	result := defaultResult{}
	// 生成hash
	t := strconv.FormatInt(time.Now().Unix(), 10)
	k := ""
	key := ""
	for k, _ = range cfg.Keys {
		key = cfg.Keys[k]
		break
	}
	hash := md5.Sum([]byte(t + key))
	// 先验证状态
	log.Println("http://" + cfg.IP + ":" + strconv.Itoa(int(cfg.Port)) + "/status?t=" + t + "&k=" + k + "&s=" + hex.EncodeToString(hash[:]))
	resp, err := client.Get("http://" + cfg.IP + ":" + strconv.Itoa(int(cfg.Port)) + "/status?t=" + t + "&k=" + k + "&s=" + hex.EncodeToString(hash[:]))
	// 失败了退出
	if err != nil {
		log.Println("似乎HttpRequests没有启动起来：", err)
		return
	}
	rb, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	// 关闭连接，要不然会在一个已经打开的长连接中重复请求
	client.CloseIdleConnections()
	// 失败了退出
	if err != nil {
		log.Println("接收数据失败：", err)
		return
	}
	err = json.Unmarshal(rb, &result)
	if err != nil {
		log.Println("似乎没有返回正确的数据类型：", err, string(rb))
		return
	}
	if result.Error.Code == 0 {
		log.Println("HttpRequests服务正常，进程ID：", result.Data)
	}
	// 单次请求
	log.Println("开始单请求测试。")
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) HttpRequests/Test"
	ja3Hash := "d8c87b9bfde38897979e41242626c2f3"
	d := url.Values{}
	d.Set("method", "GET")
	d.Set("url", "https://tls.browserleaks.com/json")
	d.Add("headers", "User-Agent: " + userAgent)
	// d.Set("body", "")
	// d.Set("proxy", "socks5://user:pass@127.0.0.1:1080")
	// d.Set("timeout", "1")
	// d.Set("skipsslverify", "1")
	d.Set("ja3", "771,49195-49196-52393-49199-49200-52392-49161-49162-49171-49172-156-157-47-53,65281-0-23-35-13-5-16-11-10,29-23-24,0")
	d.Set("ja4", "t12d1409h1_002f,0035,009c,009d,c009,c00a,c013,c014,c02b,c02c,c02f,c030,cca8,cca9_0005,000a,000b,000d,0017,0023,ff01_0403,0804,0401,0503,0805,0501,0806,0601,0201")
	// d.Set("ja3", "772,4865-4866-4867-49195-49196-52393-49199-49200-52392-49161-49162-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43-21,29-23-24,0")
	// d.Set("ja4", "t13d1713h1_002f,0035,009c,009d,1301,1302,1303,c009,c00a,c013,c014,c02b,c02c,c02f,c030,cca8,cca9_0005,000a,000b,000d,0015,0017,0023,002b,002d,0033,ff01_0403,0804,0401,0503,0805,0501,0806,0601,0201")
	resp, err = client.PostForm("http://" + cfg.IP + ":" + strconv.Itoa(int(cfg.Port)) + "/single?t=" + t + "&k=" + k + "&s=" + hex.EncodeToString(hash[:]), d)
	// 失败了退出
	if err != nil {
		log.Println("似乎HttpRequests异常退出了：", err)
		return
	}
	rb, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	// 关闭连接，要不然会在一个已经打开的长连接中重复请求
	client.CloseIdleConnections()
	// 失败了退出
	if err != nil {
		log.Println("接收数据失败：", err)
		return
	}
	sResult := map[string]string{}
	err = json.Unmarshal(rb, &sResult)
	if err != nil {
		log.Println("似乎没有返回正确的数据类型：", err, string(rb))
		log.Println("返回Header：", resp.Header)
		return
	}
	_, ok := sResult["user_agent"]
	if ok {
		if (sResult["user_agent"] == userAgent) {
			log.Println("UserAgent测试正常：", sResult["user_agent"])
		} else {
			log.Println("UserAgent测试异常，期望的UserAgent：", userAgent, "，返回的UserAgent：", sResult["user_agent"])
		}
	} else {
		log.Println("似乎没有返回正确的数据：", sResult)
	}
	_, ok = sResult["ja3_hash"]
	if ok {
		if (sResult["ja3_hash"] == ja3Hash) {
			log.Println("JA3测试正常：", sResult["ja3_hash"])
		} else {
			log.Println("JA3测试异常，期望的JA3：", ja3Hash, "，返回的JA3：", sResult["ja3_hash"])
		}
	} else {
		log.Println("似乎没有返回正确的数据：", sResult)
	}
	// 多次请求
	log.Println("开始多请求测试。")
	d = url.Values{}
	d.Set("data", "data")
	d.Set("requests", "[{\"method\":\"GET\",\"url\":\"https://tls.browserleaks.com/json\",\"headers\":[\"User-Agent: " + userAgent + "\"],\"ja3\":\"771,49195-49196-52393-49199-49200-52392-49161-49162-49171-49172-156-157-47-53,65281-0-23-35-13-5-16-11-10,29-23-24,0\",\"ja4\":\"t12d1409h1_002f,0035,009c,009d,c009,c00a,c013,c014,c02b,c02c,c02f,c030,cca8,cca9_0005,000a,000b,000d,0017,0023,ff01_0403,0804,0401,0503,0805,0501,0806,0601,0201\"},{\"method\":\"GET\",\"url\":\"https://browserleaks.com/img/logo.png\",\"headers\":[\"User-Agent: " + userAgent + "\"],\"ja3\":\"771,49195-49196-52393-49199-49200-52392-49161-49162-49171-49172-156-157-47-53,65281-0-23-35-13-5-16-11-10,29-23-24,0\",\"ja4\":\"t12d1409h1_002f,0035,009c,009d,c009,c00a,c013,c014,c02b,c02c,c02f,c030,cca8,cca9_0005,000a,000b,000d,0017,0023,ff01_0403,0804,0401,0503,0805,0501,0806,0601,0201\"}]")
	resp, err = client.PostForm("http://" + cfg.IP + ":" + strconv.Itoa(int(cfg.Port)) + "/multi?t=" + t + "&k=" + k + "&s=" + hex.EncodeToString(hash[:]), d)
	// 失败了退出
	if err != nil {
		log.Println("似乎HttpRequests异常退出了：", err)
		return
	}
	rb, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	// 关闭连接，要不然会在一个已经打开的长连接中重复请求
	client.CloseIdleConnections()
	// 失败了退出
	if err != nil {
		log.Println("接收数据失败：", err)
		return
	}
	result = defaultResult{}
	err = json.Unmarshal(rb, &result)
	if err != nil {
		log.Println("似乎没有返回正确的数据类型：", err, string(rb))
		return
	}
	if result.Error.Code == 0 {
		log.Println("初始多请求成功，请求ID：", result.Data)
	} else {
		log.Println("初始化多请求错误：", result.Error.Code)
	}
	// 下载结果
	log.Println("开始下载多请求结果。")
	d = url.Values{}
	d.Set("id", result.Data.(string))
	for {
		resp, err = client.PostForm("http://" + cfg.IP + ":" + strconv.Itoa(int(cfg.Port)) + "/download?t=" + t + "&k=" + k + "&s=" + hex.EncodeToString(hash[:]), d)
		// 失败了退出
		if err != nil {
			log.Println("似乎HttpRequests异常退出了：", err)
			return
		}
		// 已经下载完成
		if resp.StatusCode == 200 {
			// 读取，只输出长度
			// rb, err = io.ReadAll(resp.Body)
			// resp.Body.Close()
			// // 关闭连接，要不然会在一个已经打开的长连接中重复请求
			// client.CloseIdleConnections()
			// log.Println("多请求任务下载完成，数据包长度：", len(rb))
			// 读取解压，确认结果
			md := multiData{}
			gz, err := gzip.NewReader(resp.Body)
			err = json.NewDecoder(gz).Decode(&md)
			gz.Close()
			resp.Body.Close()
			// 关闭连接，要不然会在一个已经打开的长连接中重复请求
			client.CloseIdleConnections()
			// 失败了退出
			if err != nil {
				log.Println("接收数据失败：", err)
				return
			}
			log.Println("多请求任务结果：", len(md.Result), "条，原始data：", md.Data)
			for i, _ := range md.Result {
				mr := md.Result[i]
				log.Println("多请求任务 ", i, " 结果状态：", mr.Status, "，Header：", len(mr.Headers), " 条，长度：", len(mr.Content), "。")
				// 这里是图片，保存，看看是不是可行
				if i == 0 {
					err = json.Unmarshal(mr.Content, &sResult)
					if err != nil {
						log.Println("似乎没有返回正确的数据类型：", err, string(mr.Content))
						return
					}
					_, ok = sResult["user_agent"]
					if ok {
						if (sResult["user_agent"] == userAgent) {
							log.Println("UserAgent测试正常：", sResult["user_agent"])
						} else {
							log.Println("UserAgent测试异常，期望的UserAgent：", userAgent, "，返回的UserAgent：", sResult["user_agent"])
						}
					} else {
						log.Println("似乎没有返回正确的数据：", sResult)
					}
					_, ok = sResult["ja3_hash"]
					if ok {
						if (sResult["ja3_hash"] == ja3Hash) {
							log.Println("JA3测试正常：", sResult["ja3_hash"])
						} else {
							log.Println("JA3测试异常，期望的JA3：", ja3Hash, "，返回的JA3：", sResult["ja3_hash"])
						}
					} else {
						log.Println("似乎没有返回正确的数据：", sResult)
					}
				} else if i == 1 {
					root, err := file.Root()
					if err != nil {
						log.Println("无法获取目录保存临时图片：", err)
						return
					}
					tfn := root + "test." + strconv.FormatInt(time.Now().Unix(), 10) + ".png"
					tf, err := file.Create(tfn)
					if err != nil {
						log.Println("创建临时图片错误：", err)
						return
					}
					tf.Write(mr.Content)
					tf.Close()
					// 是windows，尝试打开图片
					if runtime.GOOS == "windows" {
						log.Println("临时图片保存正常，正在尝试自动打开：", tfn)
						cmd := exec.Command("cmd", "/c", "start", tfn)
						err = cmd.Start()
						if err != nil {
							log.Println("临时图片打开失败：", err)
						}
						cmd.Wait()
						time.Sleep(5 * time.Second)
					} else {
						log.Println("临时图片保存正常，请手动确认是否正常：", tfn, "，30秒后自动删除。")
						time.Sleep(30 * time.Second)
					}
					// 删除
					os.RemoveAll(tfn)
					log.Println("临时图片已删除。")
				}
			}
			// 下载完成，删除数据
			resp, err = client.PostForm("http://" + cfg.IP + ":" + strconv.Itoa(int(cfg.Port)) + "/completed?t=" + t + "&k=" + k + "&s=" + hex.EncodeToString(hash[:]), d)
			// 失败了退出
			if err != nil {
				log.Println("似乎HttpRequests异常退出了：", err)
				return
			}
			rb, err = io.ReadAll(resp.Body)
			resp.Body.Close()
			// 关闭连接，要不然会在一个已经打开的长连接中重复请求
			client.CloseIdleConnections()
			// 失败了退出
			if err != nil {
				log.Println("接收数据失败：", err)
				return
			}
			result = defaultResult{}
			err = json.Unmarshal(rb, &result)
			if err != nil {
				log.Println("似乎没有返回正确的数据类型：", err, string(rb))
				return
			}
			if result.Error.Code == 0 {
				log.Println("多请求任务数据删除成功，请求ID：", result.Data)
			// 目前不返回删除失败
			// } else {
			// 	log.Println("多请求任务删除错误：", result.Error.Code)
			}
			break
		} else if resp.StatusCode == 202 {
			log.Println("多请求任务还在进行中，等待 1 秒后自动重试。")
			time.Sleep(1 * time.Second)
		} else if resp.StatusCode == 404 {
			log.Println("多请求任务不存在或者已经完成。")
			break
		}
	}
}