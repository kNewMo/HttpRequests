package main

import (
	"io"
	"os"
	"log"
	"bytes"
	"github.com/kNewMo/HttpRequests/utils/file"
)

var root string

// 修改common.go文件
func common() {
	name := "common.go"
	log.Println("开始修改 ", name, " 文件。")
	f, err := os.OpenFile(root + name, os.O_RDWR, 0644)
	if err != nil {
		log.Fatalln("打开 ", name, " 文件失败：", err)
	}
	content, err := io.ReadAll(f)
	if err != nil {
		log.Fatalln("读取 ", name, " 文件失败：", err)
	}
	newContent := []byte{}
	// 增加简单的扩展
	posA := bytes.Index(content, []byte("extensionServerName              uint16 = 0"))
	if posA == -1 {
		log.Fatalln(name, " 文件中修改位置 extensionServerName              uint16 = 0 没有找到。")
	}
	// 先填充原数据
	newContent = append(newContent, content[0:posA]...)
	// 填充新的扩展
	newContent = append(newContent, []byte(`extensionPadding                 uint16 = 21
	extensionCompressCertificate     uint16 = 27
	extensionApplicationSettings     uint16 = 0x4469
	// extensionEncryptedClientHello    uint16 = 0xfe0d
	`)...)
	// extensionEncryptedClientHello未支持，等golang支持，chrome117版本之后才支持，所以可以用117版本之前的ua，或者禁用ech
	// 添加Curve
	posB := bytes.Index(content, []byte("X25519    CurveID = 29"))
	if posB == -1 {
		log.Fatalln(name, " 文件中修改位置 X25519    CurveID = 29 没有找到。")
	}
	// 是在位置之后修改
	posB += 22
	// 填充原数据
	newContent = append(newContent, content[posA:posB]...)
	// 填充新的Curve，这个是只要加配置就好，似乎不需要真的支持，因为大多数服务器不支持
	// 先不添加了，避免有bug
	newContent = append(newContent, []byte(`
	// X25519Kyber768Draft00    CurveID = 0x6399`)...)
	// 增加config配置项可以往下传输
	posA = bytes.Index(content, []byte("type Config struct {"))
	if posA == -1 {
		log.Fatalln(name, " 文件中修改位置 type Config struct { 没有找到。")
	}
	// 是在位置之后修改
	posA += 20
	// 填充原数据
	newContent = append(newContent, content[posB:posA]...)
	// 添加TLS 扩展，用ja3的的排序
	newContent = append(newContent, []byte(`
	Extensions []uint16
	SignatureAlgorithms []SignatureScheme`)...)
	// 修改config对象的clone方法
	posB = bytes.Index(content, []byte("return &Config{"))
	if posB == -1 {
		log.Fatalln(name, " 文件中修改位置 return &Config{ 没有找到。")
	}
	// 是在位置之后修改
	posB += 15
	// 填充原数据
	newContent = append(newContent, content[posA:posB]...)
	// 增加clone的增项
	newContent = append(newContent, []byte(`
		Extensions:                  c.Extensions,
		SignatureAlgorithms:         c.SignatureAlgorithms,`)...)
	// 增加一个方法，用于默认的扩展排序
	posA = bytes.Index(content, []byte("var supportedVersions = []uint16{"))
	if posA == -1 {
		log.Fatalln(name, " 文件中修改位置 var supportedVersions = []uint16{ 没有找到。")
	}
	// 填充原数据
	newContent = append(newContent, content[posB:posA]...)
	// 增加extensions的取值
	newContent = append(newContent, []byte(`func (c *Config) extensions() []uint16 {
	if c == nil || len(c.Extensions) == 0 {
		return SupportExtensions
	}
	return c.Extensions
}

`)...)
	// 填充原数据
	newContent = append(newContent, content[posA:]...)
	// 从头开始写
	f.Seek(0, os.SEEK_SET)
	_, err = f.Write(newContent)
	if err != nil {
		log.Fatalln("写入 ", name, " 文件失败：", err)
	}
	// 因为都是新增，超出了原始大小了，不需要截断
	// err = f.Truncate(int64(len(newContent)))
	// if err != nil {
	// 	log.Fatalln("截断 ", name, " 文件失败：", err)
	// }
	f.Close()
	log.Println("写入 ", name, " 文件成功。")
}

// 修改handshake_client.go文件
func handshake_client() {
	name := "handshake_client.go"
	log.Println("开始修改 ", name, " 文件。")
	f, err := os.OpenFile(root + name, os.O_RDWR, 0644)
	if err != nil {
		log.Fatalln("打开 ", name, " 文件失败：", err)
	}
	content, err := io.ReadAll(f)
	if err != nil {
		log.Fatalln("读取 ", name, " 文件失败：", err)
	}
	newContent := []byte{}
	// 传输hellomsg属性，以便排序
	posA := bytes.Index(content, []byte("hello := &clientHelloMsg{"))
	if posA == -1 {
		log.Fatalln(name, " 文件中修改位置 hello := &clientHelloMsg{ 没有找到。")
	}
	// 是在位置之后修改
	posA += 25
	// 填充原数据
	newContent = append(newContent, content[0:posA]...)
	// 增加TLS 扩展，用ja3的的排序
	newContent = append(newContent, []byte(`
		extensions:                   config.extensions(),`)...)
	// 增加TLS13的算法支持
	posB := bytes.Index(content, []byte("configCipherSuites := config.cipherSuites()"))
	if posB == -1 {
		log.Fatalln(name, " 文件中修改位置 configCipherSuites := config.cipherSuites() 没有找到。")
	}
	// 填充原数据
	newContent = append(newContent, content[posA:posB]...)
	// 增加TLS13算法
	newContent = append(newContent, []byte(`preferenceOrderTLS13 := []uint16{}
	if hello.supportedVersions[0] == VersionTLS13 {
		if hasAESGCMHardwareSupport {
			preferenceOrderTLS13 = defaultCipherSuitesTLS13
		} else {
			preferenceOrderTLS13 = defaultCipherSuitesTLS13NoAES
		}
	}
	`)...)
	// 让CipherSuites可以按指定的顺序排序
	posA = bytes.Index(content, []byte("for _, suiteId := range preferenceOrder {"))
	if posA == -1 {
		log.Fatalln(name, " 文件中修改位置 for _, suiteId := range preferenceOrder { 没有找到。")
	}
	// 填充原数据
	newContent = append(newContent, content[posB:posA]...)
	// 修改为按传输的值顺序
	newContent = append(newContent, []byte("for _, suiteId := range configCipherSuites {")...)
	// 不需要这段原数据
	posA += 41
	// 继续修改以便排序
	posB = bytes.Index(content, []byte("suite := mutualCipherSuite(configCipherSuites, suiteId)"))
	if posB == -1 {
		log.Fatalln(name, " 文件中修改位置 suite := mutualCipherSuite(configCipherSuites, suiteId) 没有找到。")
	}
	// 填充原数据
	newContent = append(newContent, content[posA:posB]...)
	// 修改为按传输的值顺序
	newContent = append(newContent, []byte("suite := mutualCipherSuite(preferenceOrder, suiteId)")...)
	// 不需要这段原数据
	posB += 55
	// 增加tls13算法的支持
	posA = bytes.Index(content, []byte(`if suite == nil {
			continue`))
	if posA == -1 {
		log.Fatalln(name, " 文件中修改位置 if suite == nil {\n			continue 没有找到。")
	}
	// 是在位置之后修改
	posA += 17
	// 填充原数据
	newContent = append(newContent, content[posB:posA]...)
	// tls13算法的支持
	newContent = append(newContent, []byte(`
			suite := mutualCipherSuiteTLS13(preferenceOrderTLS13, suiteId)
			if suite == nil {
				continue
			}`)...)
	// 不需要这段原数据
	posA += 12
	// 因为增加了tls13的，所以要修改下判断逻辑
	posB = bytes.Index(content, []byte("if hello.vers < VersionTLS12 && suite.flags&suiteTLS12 != 0 {"))
	if posB == -1 {
		log.Fatalln(name, " 文件中修改位置 if hello.vers < VersionTLS12 && suite.flags&suiteTLS12 != 0 { 没有找到。")
	}
	// 填充原数据
	newContent = append(newContent, content[posA:posB]...)
	// 修改为支持tls13
	newContent = append(newContent, []byte("if hello.supportedVersions[0] != VersionTLS13 && hello.vers < VersionTLS12 && suite.flags&suiteTLS12 != 0 {")...)
	// 不需要这段原数据
	posB += 61
	// 增加支持自定义SignatureAlgorithms
	posA = bytes.Index(content, []byte("hello.supportedSignatureAlgorithms = supportedSignatureAlgorithms()"))
	if posA == -1 {
		log.Fatalln(name, " 文件中修改位置 hello.supportedSignatureAlgorithms = supportedSignatureAlgorithms() 没有找到。")
	}
	// 填充原数据
	newContent = append(newContent, content[posB:posA]...)
	// 增加支持
	newContent = append(newContent, []byte(`if len(config.SignatureAlgorithms) > 0 {
			hello.supportedSignatureAlgorithms = config.SignatureAlgorithms
		} else {
			hello.supportedSignatureAlgorithms = supportedSignatureAlgorithms()
		}`)...)
	// 不需要这段原数据
	posA += 67
	// 删掉tls13，因为已经在上面实现了
	posB = bytes.Index(content, []byte(`if hasAESGCMHardwareSupport {
			hello.cipherSuites = append(hello.cipherSuites, defaultCipherSuitesTLS13...)
		} else {
			hello.cipherSuites = append(hello.cipherSuites, defaultCipherSuitesTLS13NoAES...)
		}`))
	if posB == -1 {
		log.Fatalln(name, " 文件中修改位置 if hasAESGCMHardwareSupport {\n			hello.cipherSuites = append(hello.cipherSuites, defaultCipherSuitesTLS13...)\n		} else {\n			hello.cipherSuites = append(hello.cipherSuites, defaultCipherSuitesTLS13NoAES...)\n		} 没有找到。")
	}
	// 填充原数据
	newContent = append(newContent, content[posA:posB]...)
	// 不需要这段原数据
	posB += 213
	// 填充原数据
	newContent = append(newContent, content[posB:]...)
	// 从头开始写
	f.Seek(0, os.SEEK_SET)
	_, err = f.Write(newContent)
	if err != nil {
		log.Fatalln("写入 ", name, " 文件失败：", err)
	}
	// 因为有新增，超出了原始大小了，不需要截断
	// err = f.Truncate(int64(len(newContent)))
	// if err != nil {
	// 	log.Fatalln("截断 ", name, " 文件失败：", err)
	// }
	f.Close()
	log.Println("写入 ", name, " 文件成功。")
}

// handshake_messages.go文件
func handshake_messages() {
	name := "handshake_messages.go"
	log.Println("开始修改 ", name, " 文件。")
	f, err := os.OpenFile(root + name, os.O_RDWR, 0644)
	if err != nil {
		log.Fatalln("打开 ", name, " 文件失败：", err)
	}
	content, err := io.ReadAll(f)
	if err != nil {
		log.Fatalln("读取 ", name, " 文件失败：", err)
	}
	newContent := []byte{}
	// 增加TLS 扩展，用ja3的的排序
	posA := bytes.Index(content, []byte("type clientHelloMsg struct {"))
	if posA == -1 {
		log.Fatalln(name, " 文件中修改位置 type clientHelloMsg struct { 没有找到。")
	}
	// 是在位置之后修改
	posA += 28
	// 填充原数据
	newContent = append(newContent, content[0:posA]...)
	// TLS 扩展，用ja3的的排序
	newContent = append(newContent, []byte(`
	extensions                       []uint16`)...)
	// 将扩展的改为排序的
	posB := bytes.Index(content, []byte("if len(m.serverName) > 0 {"))
	if posB == -1 {
		log.Fatalln(name, " 文件中修改位置 if len(m.serverName) > 0 { 没有找到。")
	}
	// 填充原数据
	newContent = append(newContent, content[posA:posB]...)
	// 增加排序支持
	newContent = append(newContent, []byte(`for i, _ := range m.extensions {
		if m.extensions[i] == extensionServerName {
	`)...)
	// 查找限位，因为后面的关键词可能存在多处，限制在限位之前才修改
	posC := bytes.Index(content, []byte("b.AddUint8(typeClientHello)"))
	if posC == -1 {
		log.Fatalln(name, " 文件中修改位置 b.AddUint8(typeClientHello) 没有找到。")
	}
	// 这个关键词有多处
	posA = bytes.Index(content, []byte("if m.ocspStapling {"))
	if (posA == -1) && (posA < posC) {
		log.Fatalln(name, " 文件中修改位置 if m.ocspStapling { 没有找到。")
	}
	// 填充原数据
	newContent = append(newContent, content[posB:posA]...)
	// 增加排序支持
	newContent = append(newContent, []byte(`	} else if m.extensions[i] == extensionStatusRequest {
	`)...)
	// 继续修改
	posB = bytes.Index(content, []byte("if len(m.supportedCurves) > 0 {"))
	if (posB == -1) && (posA < posC) {
		log.Fatalln(name, " 文件中修改位置 if len(m.supportedCurves) > 0 { 没有找到。")
	}
	// 填充原数据
	newContent = append(newContent, content[posA:posB]...)
	// 增加排序支持
	newContent = append(newContent, []byte(`	} else if m.extensions[i] == extensionSupportedCurves {
	`)...)
	// 这个关键词有多处
	posA = bytes.Index(content, []byte("if len(m.supportedPoints) > 0 {"))
	if (posA == -1) && (posA < posC) {
		log.Fatalln(name, " 文件中修改位置 if len(m.supportedPoints) > 0 { 没有找到。")
	}
	// 填充原数据
	newContent = append(newContent, content[posB:posA]...)
	// 增加排序支持
	newContent = append(newContent, []byte(`	} else if m.extensions[i] == extensionSupportedPoints {
	`)...)
	// 这个关键词有多处
	posB = bytes.Index(content, []byte("if m.ticketSupported {"))
	if (posB == -1) && (posA < posC) {
		log.Fatalln(name, " 文件中修改位置 if m.ticketSupported { 没有找到。")
	}
	// 填充原数据
	newContent = append(newContent, content[posA:posB]...)
	// 增加排序支持
	newContent = append(newContent, []byte(`	} else if m.extensions[i] == extensionSessionTicket {
	`)...)
	// 这个关键词有多处
	posA = bytes.Index(content, []byte("if len(m.supportedSignatureAlgorithms) > 0 {"))
	if (posA == -1) && (posA < posC) {
		log.Fatalln(name, " 文件中修改位置 if len(m.supportedSignatureAlgorithms) > 0 { 没有找到。")
	}
	// 填充原数据
	newContent = append(newContent, content[posB:posA]...)
	// 增加排序支持
	newContent = append(newContent, []byte(`	} else if m.extensions[i] == extensionSignatureAlgorithms {
	`)...)
	// 这个关键词有多处
	posB = bytes.Index(content, []byte("if len(m.supportedSignatureAlgorithmsCert) > 0 {"))
	if (posB == -1) && (posA < posC) {
		log.Fatalln(name, " 文件中修改位置 if len(m.supportedSignatureAlgorithmsCert) > 0 { 没有找到。")
	}
	// 填充原数据
	newContent = append(newContent, content[posA:posB]...)
	// 增加排序支持
	newContent = append(newContent, []byte(`	} else if m.extensions[i] == extensionSignatureAlgorithmsCert {
	`)...)
	// 这个关键词有多处
	posA = bytes.Index(content, []byte("if m.secureRenegotiationSupported {"))
	if (posA == -1) && (posA < posC) {
		log.Fatalln(name, " 文件中修改位置 if m.secureRenegotiationSupported { 没有找到。")
	}
	// 填充原数据
	newContent = append(newContent, content[posB:posA]...)
	// 增加排序支持
	newContent = append(newContent, []byte(`	} else if m.extensions[i] == extensionRenegotiationInfo {
	`)...)
	// 这个关键词有多处
	posB = bytes.Index(content, []byte("if m.extendedMasterSecret {"))
	if (posB == -1) && (posA < posC) {
		log.Fatalln(name, " 文件中修改位置 if m.extendedMasterSecret { 没有找到。")
	}
	// 填充原数据
	newContent = append(newContent, content[posA:posB]...)
	// 增加排序支持
	newContent = append(newContent, []byte(`	} else if m.extensions[i] == extensionExtendedMasterSecret {
	`)...)
	// 继续修改
	posA = bytes.Index(content, []byte("if len(m.alpnProtocols) > 0 {"))
	if (posA == -1) && (posA < posC) {
		log.Fatalln(name, " 文件中修改位置 if len(m.alpnProtocols) > 0 { 没有找到。")
	}
	// 填充原数据
	newContent = append(newContent, content[posB:posA]...)
	// 增加排序支持
	newContent = append(newContent, []byte(`	} else if m.extensions[i] == extensionALPN {
	`)...)
	// 这个关键词有多处
	posB = bytes.Index(content, []byte("if m.scts {"))
	if (posB == -1) && (posA < posC) {
		log.Fatalln(name, " 文件中修改位置 if m.scts { 没有找到。")
	}
	// 填充原数据
	newContent = append(newContent, content[posA:posB]...)
	// 增加排序支持
	newContent = append(newContent, []byte(`	} else if m.extensions[i] == extensionSCT {
	`)...)
	// 继续修改
	posA = bytes.Index(content, []byte("if len(m.supportedVersions) > 0 {"))
	if (posA == -1) && (posA < posC) {
		log.Fatalln(name, " 文件中修改位置 if len(m.supportedVersions) > 0 { 没有找到。")
	}
	// 填充原数据
	newContent = append(newContent, content[posB:posA]...)
	// 增加排序支持
	newContent = append(newContent, []byte(`	} else if m.extensions[i] == extensionSupportedVersions {
	`)...)
	// 这个关键词有多处
	posB = bytes.Index(content, []byte("if len(m.cookie) > 0 {"))
	if (posB == -1) && (posA < posC) {
		log.Fatalln(name, " 文件中修改位置 if len(m.cookie) > 0 { 没有找到。")
	}
	// 填充原数据
	newContent = append(newContent, content[posA:posB]...)
	// 增加排序支持
	newContent = append(newContent, []byte(`	} else if m.extensions[i] == extensionCookie {
	`)...)
	// 继续修改
	posA = bytes.Index(content, []byte("if len(m.keyShares) > 0 {"))
	if (posA == -1) && (posA < posC) {
		log.Fatalln(name, " 文件中修改位置 if len(m.keyShares) > 0 { 没有找到。")
	}
	// 填充原数据
	newContent = append(newContent, content[posB:posA]...)
	// 增加排序支持
	newContent = append(newContent, []byte(`	} else if m.extensions[i] == extensionKeyShare {
	`)...)
	// 这个关键词有多处
	posB = bytes.Index(content, []byte("if m.earlyData {"))
	if (posB == -1) && (posA < posC) {
		log.Fatalln(name, " 文件中修改位置 if m.earlyData { 没有找到。")
	}
	// 填充原数据
	newContent = append(newContent, content[posA:posB]...)
	// 增加排序支持
	newContent = append(newContent, []byte(`	} else if m.extensions[i] == extensionEarlyData {
	`)...)
	// 继续修改
	posA = bytes.Index(content, []byte("if len(m.pskModes) > 0 {"))
	if (posA == -1) && (posA < posC) {
		log.Fatalln(name, " 文件中修改位置 if len(m.pskModes) > 0 { 没有找到。")
	}
	// 填充原数据
	newContent = append(newContent, content[posB:posA]...)
	// 增加排序支持
	newContent = append(newContent, []byte(`	} else if m.extensions[i] == extensionPSKModes {
	`)...)
	// 这个关键词有多处
	posB = bytes.Index(content, []byte("if m.quicTransportParameters != nil {"))
	if (posB == -1) && (posA < posC) {
		log.Fatalln(name, " 文件中修改位置 if m.quicTransportParameters != nil { 没有找到。")
	}
	// 填充原数据
	newContent = append(newContent, content[posA:posB]...)
	// 增加排序支持
	newContent = append(newContent, []byte(`	} else if m.extensions[i] == extensionQUICTransportParameters {
	`)...)
	// 继续修改
	posA = bytes.Index(content, []byte("if len(m.pskIdentities) > 0 {"))
	if (posA == -1) && (posA < posC) {
		log.Fatalln(name, " 文件中修改位置 if len(m.pskIdentities) > 0 { 没有找到。")
	}
	// 填充原数据
	newContent = append(newContent, content[posB:posA]...)
	// 增加排序支持
	newContent = append(newContent, []byte(`	} else if m.extensions[i] == extensionPreSharedKey {
	`)...)
	// 这个关键词有多处，因为是修改第一处，所以简单处理
	posB = bytes.Index(content, []byte("extBytes, err := exts.Bytes()"))
	if (posB == -1) && (posA < posC) {
		log.Fatalln(name, " 文件中修改位置 extBytes, err := exts.Bytes() 没有找到。")
	}
	// 填充原数据
	newContent = append(newContent, content[posA:posB]...)
	// 增加一个基础扩展，填充，避免一些bug上冲突，似乎实在扩展长度在505 到 507之间时需要填充到512字节以上，https://datatracker.ietf.org/doc/html/rfc7685#section-4，这里先填充一个字节，以后可以具体再完善
	newContent = append(newContent, []byte(`	} else if m.extensions[i] == extensionPadding {
			exts.AddUint16(extensionPadding)
			exts.AddUint16LengthPrefixed(func(exts *cryptobyte.Builder) {
				exts.AddUint8(0)
			})
`)...)
	// 增加一个扩展，证书压缩，https://datatracker.ietf.org/doc/html/rfc8879
	newContent = append(newContent, []byte(`		} else if m.extensions[i] == extensionCompressCertificate {
			exts.AddUint16(extensionCompressCertificate)
			exts.AddUint16LengthPrefixed(func(exts *cryptobyte.Builder) {
				exts.AddUint8(2)
				exts.AddUint16(0)
			})
`)...)
	// 增加一个扩展，还不清楚作用，没找到文档
	newContent = append(newContent, []byte(`		} else if m.extensions[i] == extensionApplicationSettings {
			exts.AddUint16(extensionApplicationSettings)
			exts.AddUint16LengthPrefixed(func(exts *cryptobyte.Builder) {
				exts.AddUint16LengthPrefixed(func(exts *cryptobyte.Builder) {
					exts.AddUint8LengthPrefixed(func(exts *cryptobyte.Builder) {
						exts.AddBytes([]byte("h2"))
					})
				})
			})
`)...)
	// 增加}收尾
	newContent = append(newContent, []byte(`		}
	}
	`)...)
	// 填充原数据
	newContent = append(newContent, content[posB:]...)
	// 从头开始写
	f.Seek(0, os.SEEK_SET)
	_, err = f.Write(newContent)
	if err != nil {
		log.Fatalln("写入 ", name, " 文件失败：", err)
	}
	// 因为都是新增，超出了原始大小了，不需要截断
	// err = f.Truncate(int64(len(newContent)))
	// if err != nil {
	// 	log.Fatalln("截断 ", name, " 文件失败：", err)
	// }
	f.Close()
	log.Println("写入 ", name, " 文件成功。")
}

// key_schedule.go文件
func key_schedule() {
	name := "key_schedule.go"
	log.Println("开始修改 ", name, " 文件。")
	f, err := os.OpenFile(root + name, os.O_RDWR, 0644)
	if err != nil {
		log.Fatalln("打开 ", name, " 文件失败：", err)
	}
	content, err := io.ReadAll(f)
	if err != nil {
		log.Fatalln("读取 ", name, " 文件失败：", err)
	}
	newContent := []byte{}
	// 传输hellomsg属性，以便排序
	posA := bytes.Index(content, []byte("return ecdh.P521(), true"))
	if posA == -1 {
		log.Fatalln(name, " 文件中修改位置 return ecdh.P521(), true 没有找到。")
	}
	// 是在位置之后修改
	posA += 24
	// 填充原数据
	newContent = append(newContent, content[0:posA]...)
	// 增加TLS 扩展，用ja3的的排序
	newContent = append(newContent, []byte(`
	// case X25519Kyber768Draft00:
	// 	return ecdh.P521(), true`)...)
	// 填充原数据
	newContent = append(newContent, content[posA:]...)
	// 从头开始写
	f.Seek(0, os.SEEK_SET)
	_, err = f.Write(newContent)
	if err != nil {
		log.Fatalln("写入 ", name, " 文件失败：", err)
	}
	// 因为有新增，超出了原始大小了，不需要截断
	// err = f.Truncate(int64(len(newContent)))
	// if err != nil {
	// 	log.Fatalln("截断 ", name, " 文件失败：", err)
	// }
	f.Close()
	log.Println("写入 ", name, " 文件成功。")
}

// 增加support.go文件
func support() {
	name := "support.go"
	log.Println("开始修改 ", name, " 文件。")
	f, err := os.OpenFile(root + name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalln("打开 ", name, " 文件失败：", err)
	}
	_, err = f.Write([]byte(`package tls

// 定义一些ja3要用的检测项
// 当前版本的 cipherSuitesPreferenceOrder 与 cipherSuitesPreferenceOrderNoAES 是相等的，所以这里不做判断，直接用 cipherSuitesPreferenceOrder，还需要合并tls1.3的defaultCipherSuitesTLS13
var SupportCipherSuites = append(cipherSuitesPreferenceOrder, defaultCipherSuitesTLS13...)

var SupportCurveIDs = []uint16{
	uint16(CurveP256),
	uint16(CurveP384),
	uint16(CurveP521),
	uint16(X25519),
	// uint16(X25519Kyber768Draft00),
}

// 按handshake_messages.go文件中marshal()的排序
var SupportExtensions = []uint16{
	extensionServerName,
	extensionStatusRequest,
	extensionSupportedCurves,
	extensionSupportedPoints,
	extensionSessionTicket,
	extensionSignatureAlgorithms,
	extensionSignatureAlgorithmsCert,
	extensionRenegotiationInfo,
	extensionExtendedMasterSecret,
	extensionALPN,
	extensionSCT,
	extensionSupportedVersions,
	extensionCookie,
	extensionKeyShare,
	extensionEarlyData,
	extensionPSKModes,
	extensionQUICTransportParameters,
	extensionPreSharedKey,
	extensionCertificateAuthorities,
	extensionPadding,
	extensionCompressCertificate,
	extensionApplicationSettings,
	// extensionEncryptedClientHello,
}

var SupportSignatureAlgorithms = []uint16{
	uint16(PSSWithSHA256),
	uint16(ECDSAWithP256AndSHA256),
	uint16(Ed25519),
	uint16(PSSWithSHA384),
	uint16(PSSWithSHA512),
	uint16(PKCS1WithSHA256),
	uint16(PKCS1WithSHA384),
	uint16(PKCS1WithSHA512),
	uint16(ECDSAWithP384AndSHA384),
	uint16(ECDSAWithP521AndSHA512),
	uint16(PKCS1WithSHA1),
	uint16(ECDSAWithSHA1),
}`))
	if err != nil {
		log.Fatalln("写入 ", name, " 文件失败：", err)
	}
	f.Close()
	log.Println("写入 ", name, " 文件成功。")
}

func main() {
	var err error
	root, err = file.Root()
	if err != nil {
		log.Fatalln("无法获取运行目录：", err)
	}
	common()
	handshake_client()
	handshake_messages()
	// key_schedule()
	support()
}