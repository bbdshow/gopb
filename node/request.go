package node

import (
	"net"
	"net/http"
	"time"
)

var (
	DefaultTransport = NewDefaultTransport()

	DefaultClient = &http.Client{
		Transport:     DefaultTransport,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       30 * time.Second,
	}
)

func NewDefaultClient() *http.Client {
	return &http.Client{
		Transport:     NewDefaultTransport(),
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       30 * time.Second,
	}
}

func NewDefaultTransport() *http.Transport {
	return &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          1024,
		MaxIdleConnsPerHost:   64,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableKeepAlives:     false,
	}
}

func RequestDo(client *http.Client, req *http.Request) (*http.Response, error) {
	return client.Do(req)
}
