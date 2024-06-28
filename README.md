# HttpRequests

一个支持模拟 JA3 的 http 请求工具，支持单次和批量请求，支持对每个请求单独使用代理。

## 说明

- 只支持 TCP 协议的 http 请求，还不支持 UDP 下的 QUIC 协议。
- 只支持不是最新版的 http 组件的 JA3 的模拟，无法模拟较新的 Chrome 浏览器等。
- windows 下基于 go1.22.0 版本编译，linux 下基于 go.1.21.0 编译，目前只编译了 amd64 的版本。
- 编译需要替换 go 的源码，go/src/crypto/tls，请先备份 go/src/crypto/tls 目录，再将 change/change.exe 复制到 tls 目录运行即可。
- 目前暂时只支持了 GET 和 POST 的请求，其他模式的请求还不支持。

## 文档

### 简单验证
- 目前先使用简单的验证，避免接口被无授权使用，目前使用简单验证，不验证传输的数据，也就是可能存在被中间人抓取 url 后可以 5 分钟内发起请求。建议前端使用 nginx 等通过 https 转发请求。
- 验证算法：`http://[ip]:[port]/[apiname]?t=[时间戳]&k=[配置的key索引]&s=[md5(时间戳+key)]`
- 验证错误时会在 http header 里返回 `HR-Error-Code: 10*` 和 `HR-Error-Message: Sign Error.`
- 例子：`http://127.0.0.1:180/status?t=1136185445&k=default&s=215d4fe4f0295cdba24d83bf967122a8`
- HR-Error-Code：
  - 101：时间超过了验证期，±5 分钟。
  - 102：密钥配置不存在。
  - 103：签名验证不匹配。

### 请求数据结构
```
{
	// 模式，可选，目前只支持 GET 和 POST，默认 GET
	"method": "GET/POST",
	// 地址，必填
	"url": "https://...",
	// http header，可选
	"headers": [
		"User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) HttpRequests/Test"
	],
	// 正文，可选
	"body": "",
	// 代理，可选，目前支持 http 和 socks5，http 代理为 http://...
	"proxy": "socks5://user:pass@127.0.0.1:1080",
	// 超时，可选，默认 55 秒超时
	"timeout": "30",
	// 跳过证书验证，可选，默认为不跳过，为 1 则忽略证书错误
	"skipsslverify": "1",
	// JA3 Fullstring，可选，从 Wireshark 抓包获取到，默认使用 golang 自己的，不同版本的 golang 编译出来的不一样
	"ja3": "771,49195-49196-52393-49199-49200-52392-49161-49162-49171-49172-156-157-47-53,65281-0-23-35-13-5-16-11-10,29-23-24,0",
	// JA4_r，可选，从 Wireshark 抓包获取到，默认使用 golang 自己的，不同版本的 golang 编译出来的不一样
	"ja4": "t12d1409h1_002f,0035,009c,009d,c009,c00a,c013,c014,c02b,c02c,c02f,c030,cca8,cca9_0005,000a,000b,000d,0017,0023,ff01_0403,0804,0401,0503,0805,0501,0806,0601,0201"
}
```
### 返回数据结构
- 返回数据分两种，status，multi，completed 接口返回 JSON 数据。
  ```
	{
		// 错误信息
		"error": {
			// 错误代码
			"code": 0,
			// 错误描述
			"message": ""
		},
		// 数据
		"data": [data]
	}
  ```
- single，download 接口直接返回 http 数据，如果有异常错误则在 http header 的 `Http Status` 以及 `HR-Error-Code`，`HR-Error-Message` 中返回。

### 接口
**status** 查看状态
- 请求模式：GET
- 请求参数：无
- 返回：
  ```
	{
		"error": {
			"code": 0,
			"message": ""
		},
		"data": 进程 pid
	}
  ```

**single** 单次请求
- 请求模式：POST
- 请求参数：`请求数据结构`
  - Content-Type 为 `application/x-www-form-urlencoded` 时，将 `请求数据结构` 每一项作为 Form 项传输。
  - Content-Type 为 `application/json` 时，直接将 `请求数据结构` 用 JSON 传输。
- 返回：本接口是转发请求，会将目标请求原始数据直接返回，等同于直接请求目标获取的结果，如果有异常错误则返回 418 的 http staus，具体错误看 http header 里的 `HR-Error-Code` 和 `HR-Error-Message`。
- `HR-Error-Code`：
  - 201：请求数据 JSON 解析错误。
  - 202：Bad Request，目前只有 url 为空的时候才返回。
  - 203：Too many Requests，已经达到服务的并发限制，因为 single 是即时的接口，达到限制就无法再发起请求。
  - 204：其他错误，可能是解析JA3，JA4出错，也可能是请求时网络问题等等，具体看 `HR-Error-Message`。
- 例子：
  ```
	fetch("http://127.0.0.1:180/single?t=1136185445&k=default&s=215d4fe4f0295cdba24d83bf967122a8", {
		"headers": {
			"content-type": "application/x-www-form-urlencoded",
		},
		"body": "url=https://tls.browserleaks.com/json&ja3=771,49195-49196-52393-49199-49200-52392-49161-49162-49171-49172-156-157-47-53,65281-0-23-35-13-5-16-11-10,29-23-24,0&ja4=t12d1409h1_002f,0035,009c,009d,c009,c00a,c013,c014,c02b,c02c,c02f,c030,cca8,cca9_0005,000a,000b,000d,0017,0023,ff01_0403,0804,0401,0503,0805,0501,0806,0601,0201",
		"method": "POST"
	});
  ```

**multi** 批量请求
- 请求模式：POST
- 请求参数：
  - 数据结构
    ```
	{
		// 返回数据时原文回传的数据，可选，可以用来存储本次批量请求的发起参数，这样即时部分请求失败，可以在不做额外的存储情况下也可以重新发起请求
		"data": "回传数据",
		// 多个请求的数组，必填
		"requests": [`请求数据结构`,`请求数据结构`]
	}
    ```
  - Content-Type 为 `application/x-www-form-urlencoded` 时，将 `请求数据结构` 每一项作为 Form 项传输，如果项的值非字符串则转为 JSON 后作为该项的值。
  - Content-Type 为 `application/json` 时，直接将 `请求参数` 用 JSON 传输。
- 返回：
  ```
	{
		"error": {
			"code": 0,
			"message": ""
		},
		"data": 任务 pid
	}
  ```
- error.code：
  - 418：请求数据 JSON 解析错误。
  - 500：无法创建结果目录。
- 例子：
  ```
	fetch("http://127.0.0.1:180/multi?t=1136185445&k=default&s=215d4fe4f0295cdba24d83bf967122a8", {
		"headers": {
			"content-type": "application/json",
		},
		"body": "{\"requests\":[{\"method\":\"GET\",\"url\":\"https://tls.browserleaks.com/json\",\"headers\":[\"User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) HttpRequests/Test\"],\"ja3\":\"771,49195-49196-52393-49199-49200-52392-49161-49162-49171-49172-156-157-47-53,65281-0-23-35-13-5-16-11-10,29-23-24,0\",\"ja4\":\"t12d1409h1_002f,0035,009c,009d,c009,c00a,c013,c014,c02b,c02c,c02f,c030,cca8,cca9_0005,000a,000b,000d,0017,0023,ff01_0403,0804,0401,0503,0805,0501,0806,0601,0201\"},{\"method\":\"GET\",\"url\":\"https://browserleaks.com/img/logo.png\",\"headers\":[\"User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) HttpRequests/Test\"],\"ja3\":\"771,49195-49196-52393-49199-49200-52392-49161-49162-49171-49172-156-157-47-53,65281-0-23-35-13-5-16-11-10,29-23-24,0\",\"ja4\":\"t12d1409h1_002f,0035,009c,009d,c009,c00a,c013,c014,c02b,c02c,c02f,c030,cca8,cca9_0005,000a,000b,000d,0017,0023,ff01_0403,0804,0401,0503,0805,0501,0806,0601,0201\"}]}",
		"method": "POST"
	});
  ```

**download** 批量请求结果下载
- 请求模式：POST
- 请求参数：
  - id：任务 id
- 返回：
  - 任务数据，JSON 后 gzip 格式。
  - 数据结构
    ```
	{
		// 返回请求时传的 data 原文
		"data": "请求时的 data",
		// 多个请求结果的数组
		"result": [
			{
				// 请求返回的 Http Status
				"status": 200,
				// 如果 status 为 0 的话，这里会有相应的错误信息
				"error": "",
				// 请求的 http header
				"headers": [
					"Access-Control-Allow-Headers: *",
					"Content-Type: application/json",
					"..."
				],
				// 请求的正文，为了避免转换可能带来的不确定性，直接传输二进制内容，因为 json 本身无法传输二进制，所以是 base64 encode 后的，base64 decode 后即可
				"content": "ewogICJ1c2VyX2FnZW50IjogIk1vemlsbGEvNS4wIChXaW5kb3dzIE5UIDEwLjA7IFdpbjY0OyB4NjQpIEh0dHBSZXF1ZXN0cy9UZXN0IiwKICAiamEzX2hhc2giOiAiZDhjODdiOWJmZGUzODg5Nzk3OWU0MTI0MjYyNmMyZjMiLAogICJqYTNfdGV4dCI6ICI3NzEsNDkxOTUtNDkxOTYtNTIzOTMtNDkxOTktNDkyMDAtNTIzOTItNDkxNjEtNDkxNjItNDkxNzEtNDkxNzItMTU2LTE1Ny00Ny01Myw2NTI4MS0wLTIzLTM1LTEzLTUtMTYtMTEtMTAsMjktMjMtMjQsMCIsCiAgImphM25faGFzaCI6ICJlNjVlNTNmNmQ5YTdhMGRmN2U5N2NmMWJkNWJhNjA4MiIsCiAgImphM25fdGV4dCI6ICI3NzEsNDkxOTUtNDkxOTYtNTIzOTMtNDkxOTktNDkyMDAtNTIzOTItNDkxNjEtNDkxNjItNDkxNzEtNDkxNzItMTU2LTE1Ny00Ny01MywwLTUtMTAtMTEtMTMtMTYtMjMtMzUtNjUyODEsMjktMjMtMjQsMCIsCiAgImFrYW1haV9oYXNoIjogIiIsCiAgImFrYW1haV90ZXh0IjogIiIKfQo="
			},
			// 更多的数据
			{
				// ...
			}
		]
	}
    ```
- 错误：`Http Status`
  - 202：任务还在进行中，稍后再试。完成时间未知，建议发起 multi 请求后，每一分钟尝试一次。最好能根据实际的请求预估完成时间之后请求。
  - 404：任务不存在或者任务数据已删除。
- 例子：
  ```
	fetch("http://127.0.0.1:180/download?t=1136185445&k=default&s=215d4fe4f0295cdba24d83bf967122a8", {
		"headers": {
			"content-type": "application/x-www-form-urlencoded",
		},
		"body": "id=0102-a83b988d-4e95-67cd-5e50-dbdf9290d656",
		"method": "POST"
	});
  ```
  
**completed** 批量请求结果下载完成，通知删除结果
- 请求模式：POST
- 请求参数：
  - id：任务 id
- 返回：
  ```
	{
		"error": {
			"code": 0,
			"message": ""
		},
		"data": 任务 pid
	}
  ```
- error.code：
  - 202：任务还在进行中，稍后再试。
  - 404：任务不存在或者任务数据已删除。
- 例子：
  ```
	fetch("http://127.0.0.1:180/completed?t=1136185445&k=default&s=215d4fe4f0295cdba24d83bf967122a8", {
		"headers": {
			"content-type": "application/x-www-form-urlencoded",
		},
		"body": "id=0102-a83b988d-4e95-67cd-5e50-dbdf9290d656",
		"method": "POST"
	});
  ```