package node

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/url"
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

func (r *Request) GenHTTPRequest() *http.Request {
	_url, _ := url.Parse(r.URL)
	req := &http.Request{
		Method:        r.Method,
		URL:           _url,
		Body:          nil,
		GetBody:       nil,
		ContentLength: 0,
	}
	for k, v := range r.Headers {
		req.Header.Set(k, v)
	}
	if r.Body != nil {
		buf := bytes.NewBuffer(r.Body)
		req.Body = ioutil.NopCloser(buf)
		req.ContentLength = int64(buf.Len())
	}

	return req
}

type Response struct {
	Size       int64  `json:"size"`
	StatusCode int    `json:"status_code"`
	Duration   int64  `json:"duration"`
	Error      bool   `json:"error"`
	Body       string `json:"body"`
}
