package node

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Request struct {
	Method           string
	Scheme           string
	URL              string
	Params           map[string]string
	Headers          map[string]string
	Body             string
	DisableKeepAlive bool
	Insecure         bool // 建立不安全连接
	Tls              *tls.Config
	ResponseContains string
}

func (r *Request) GenHTTPRequest() *http.Request {
	formatURL := r.URL
	if len(r.Params) > 0 {
		if !strings.HasSuffix(r.URL, "?") {
			formatURL = r.URL + "?"
		}
		param := ""
		for k, v := range r.Params {
			param += fmt.Sprintf("%s=%s&", k, url.QueryEscape(v))
		}

		formatURL = formatURL + strings.TrimRight(param, "&")
	}
	_url, err := url.Parse(formatURL)
	if err != nil {
		panic("url parse " + err.Error())
	}
	r.Scheme = _url.Scheme
	req := &http.Request{
		Method:        r.Method,
		URL:           _url,
		Body:          nil,
		GetBody:       nil,
		ContentLength: 0,
		Header:        http.Header{},
	}
	for k, v := range r.Headers {
		req.Header.Set(k, v)
	}

	if r.Body != "" {
		buf := bytes.NewBuffer([]byte(r.Body))
		req.Body = ioutil.NopCloser(buf)
		req.ContentLength = int64(buf.Len())
	}

	return req
}

type Response struct {
	Size       int64  `json:"size"`
	StatusCode int    `json:"status_code"`
	Duration   int64  `json:"duration"` // micro
	Error      error  `json:"error"`
	Body       string `json:"body"`
}
