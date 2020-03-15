package node

import (
	"crypto/tls"
)

type Config struct {
	Concurrent int
	TotalCalls int
	Port       string
	Nodes      []string // 每个node设置并发和总请求数
}

type Request struct {
	Method           string
	Scheme           string
	URL              string
	Headers          map[string]string
	Body             []byte
	DisableKeepAlive bool
	Insecure         bool // 建立不安全连接
	Tls              *tls.Config
	ResponseContains string
}

type Response struct {
	Size       int64  `json:"size"`
	StatusCode int    `json:"status_code"`
	Duration   int64  `json:"duration"`
	Error      bool   `json:"error"`
	Body       string `json:"body"`
}
