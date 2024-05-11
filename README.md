# HttpRequests

一个支持模拟JA3的http请求工具，支持单次和批量请求，支持对每个请求单独使用代理。

## 说明

- 只支持TCP协议的http请求，还不支持UDP下的QUIC协议。
- 只支持不是最新版的http组件的JA3的模拟，无法模拟较新的Chome浏览器等。
- windows下基于go1.22.0版本编译，linux下基于go.1.21.0编译，目前只编译了amd64的版本。
- 编译需要替换go的源码，go/src/crypto/tls，请先备份go/src/crypto/tls目录，再用tls目录下对应源码覆盖

## 文档

**single** 单次请求

**multi** 批量请求

**download** 批量请求结果下载
