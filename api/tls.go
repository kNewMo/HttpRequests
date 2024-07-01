package api

import (
	"errors"
	"strconv"
	"strings"
	"crypto/tls"
)

// ja4 TLS 版本转换
func ja4TLSVersion(v string) (uint16) {
	if v == "s3" {
		return tls.VersionSSL30
	} else if v == "10" {
		return tls.VersionTLS10
	} else if v == "11" {
		return tls.VersionTLS11
	} else if v == "12" {
		return tls.VersionTLS12
	} else if v == "13" {
		return tls.VersionTLS13
	} else {
		return 0
	}
}

// 判断是否是ja3支持的配置
func isSupport(val string, vals []uint16) (uint16, bool) {
	iVal, _ := strconv.Atoi(val)
	uintVal := uint16(iVal)
	for i, _ := range vals {
		if vals[i] == uintVal {
			return uintVal, true
		}
	}
	return 0, false
}

// 判断是否是ja4支持的配置
func isSupport16(val string, vals []uint16) (uint16, bool) {
	iVal, _ := strconv.ParseInt(val, 16, 0)
	uintVal := uint16(iVal)
	for i, _ := range vals {
		if vals[i] == uintVal {
			return uintVal, true
		}
	}
	return 0, false
}

// 创建JA3，JA3 Fullstring 可以从Wireshark抓包获取到
func tlsConfigJA3(tlsConfig *tls.Config, ja3 string) (error) {
	ja := strings.Split(ja3, ",")
	// 数量对，才操作
	if len(ja) == 5 {
		// tls版本
		tempInt, _ := strconv.Atoi(ja[0])
		tlsVersion := uint16(tempInt)
		if (tlsVersion >= tls.VersionSSL30) && (tlsVersion <= tls.VersionTLS13) {
			tlsConfig.MinVersion = tlsVersion
			tlsConfig.MaxVersion = tlsVersion
		} else {
			return errors.New("JA3 unsupport TLS version.")
		}
		// 加密算法顺序
		cipherSuites := strings.Split(ja[1], "-")
		for i, _ := range cipherSuites {
			// 支持的协议，tls/cipher_suites.go，TLS_RSA_WITH_RC4_128_SHA~TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305
			cipherSuite, support := isSupport(cipherSuites[i], tls.SupportCipherSuites)
			// 支持的算法
			if support {
				tlsConfig.CipherSuites = append(tlsConfig.CipherSuites, cipherSuite)
			} else {
				return errors.New("JA3 unsupport cipher suite: " + cipherSuites[i])
			}
		}
		// 扩展顺序
		extensions := strings.Split(ja[2], "-")
		for i, _ := range extensions {
			// 支持的协议，tls/common.go，extensionServerName~extensionRenegotiationInfo
			extension, support := isSupport(extensions[i], tls.SupportExtensions)
			// 支持的算法
			if support {
				// 如果有支持session_ticket，需要初始化，PSKModes是在loadSession中初始化的，所以也要在这里初始化
				if (extension == 35) || (extension == 45) {
					if tlsConfig.ClientSessionCache == nil {
						tlsConfig.ClientSessionCache = tls.NewLRUClientSessionCache(1)
					}
				}
				tlsConfig.Extensions = append(tlsConfig.Extensions, extension)
			} else {
				return errors.New("JA3 unsupport extension: " + extensions[i])
			}
		}
		// Supported Groups顺序
		curveIDs := strings.Split(ja[3], "-")
		for i, _ := range curveIDs {
			// 支持的曲线，tls/common.go，CurveP256~X25519
			curveID, support := isSupport(curveIDs[i], tls.SupportCurveIDs)
			// 支持的算法
			if support {
				tlsConfig.CurvePreferences = append(tlsConfig.CurvePreferences, tls.CurveID(curveID))
			} else {
				return errors.New("JA3 unsupport curve: " + curveIDs[i])
			}
		}
		// Elliptic curves point，似乎正常只等于0，没有不等于0的情况，程序不支持，所以如果不等于0就抛出异常
		if ja[4] != "0" {
			return errors.New("JA3 unsupport TLS elliptic curves point formats.")
		}
		// ja3 只要处理这些即可
		return nil
	} else {
		return errors.New("JA3 string format error.")
	}
}

// 创建JA4，JA4_r 可以从Wireshark抓包获取到，不过部分是sorted的，这里用了不排序的部分，所有不排序的需要用ja4_ro，不过还没看到有简单的工具可以提供这值
func tlsConfigJA4(tlsConfig *tls.Config, ja4_r string) (error) {
	ja := strings.Split(ja4_r, "_")
	// 数量对，才操作
	if len(ja) != 4 {
		return errors.New("JA4 string format error.")
	}
	// 10个字符
	if len(ja[0]) != 10 {
		return errors.New("JA4 string format error.")
	}
	// 不支持的协议，暂时只支持tcp协议
	if ja[0][0:1] != "t" {
		return errors.New("JA4 unsupport protocol.")
	}
	// tls版本，这里要处理，因为和ja3取值不一样
	tlsVersion := ja4TLSVersion(ja[0][1:3])
	if tlsVersion == 0 {
		return errors.New("JA4 unsupport TLS version.")
	}
	// 如果ja3版本大于当前
	if tlsConfig.MinVersion > tlsVersion {
		tlsConfig.MinVersion = tlsVersion
	}
	// 如果ja3版本小于当前
	if tlsConfig.MaxVersion < tlsVersion {
		tlsConfig.MaxVersion = tlsVersion
	}
	/* 不支持的部分
	tlsConfig.Extensions = []uint16{}
	// 是否是SNI，用不上
	if ja[0][3:4] == "d" {
		tlsConfig.Extensions = append(tlsConfig.Extensions, 0)
	}
	// CipherSuites数量，用不上
	ja[0][4:6]
	// Extensions数量，用不上
	ja[0][6:8]
	*/
	// alpn协议
	tlsConfig.NextProtos = []string{}
	// h2的，需要附带h1.1，也有的没有带http1.1，所以可能需要加额外参数？似乎h2时会自动附带上http/1.1，D:\WebSite\go\src\net\http\h2_bundle.go
	if ja[0][8:10] == "h2" {
		tlsConfig.NextProtos = append(tlsConfig.NextProtos, "h2")
		// tlsConfig.NextProtos = append(tlsConfig.NextProtos, "http/1.1")
	// h1.1的
	} else if ja[0][8:10] == "h1" {
		tlsConfig.NextProtos = append(tlsConfig.NextProtos, "http/1.1")
	}
	// 排序后的 CipherSuites，用不上
	// ja[1]
	// 排序后的 Extensions，用不上
	// ja[2]
	// 未排序的 Signature Algorithms
	if len(ja[3]) > 0 {
		signatureAlgorithms := strings.Split(ja[3], ",")
		for i, _ := range signatureAlgorithms {
			signatureAlgorithm, support := isSupport16(signatureAlgorithms[i], tls.SupportSignatureAlgorithms)
			// 支持的算法
			if support {
				tlsConfig.SignatureAlgorithms = append(tlsConfig.SignatureAlgorithms, tls.SignatureScheme(signatureAlgorithm))
			} else {
				return errors.New("JA3 unsupport signature algorithm: " + signatureAlgorithms[i])
			}
		}
	}
	return nil
}