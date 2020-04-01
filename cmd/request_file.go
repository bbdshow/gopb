package cmd

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github/huzhongqing/gopb/node"
	"io/ioutil"
	"strings"
	"time"
)

// 批量请求多个任务

type RequestConfigs []RequestConfig

type RequestConfig struct {
	Duration          string            `json:"duration"`
	Concurrent        int               `json:"concurrent"`
	TotalCalls        int               `json:"total_calls"`
	Method            string            `json:"method"`
	URL               string            `json:"url"`
	Headers           map[string]string `json:"headers"`
	DisableKeepAlives bool              `json:"disable_keep_alives"`
	Insecure          bool              `json:"insecure"` // 建立不安全连接
	CertFilename      string            `json:"cert_filename"`
	KeyFilename       string            `json:"key_filename"`
	Params            map[string]string `json:"params"`
	Body              string            `json:"body"`
	Contains          string            `json:"contains"`
}

func (v RequestConfig) ToRequest() (node.Request, error) {
	req := node.Request{
		Method:           v.Method,
		URL:              v.URL,
		Params:           v.Params,
		Headers:          v.Headers,
		Body:             v.Body,
		DisableKeepAlive: v.DisableKeepAlives,
		Insecure:         v.Insecure,
		Tls:              nil,
		ResponseContains: v.Contains,
	}
	if v.CertFilename != "" && v.KeyFilename != "" {
		cert, err := tls.LoadX509KeyPair(v.CertFilename, v.KeyFilename)
		if err != nil {
			return node.Request{}, err
		}
		req.Tls = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}
	return req, nil
}

func (v RequestConfig) GetDuration() time.Duration {
	d, err := time.ParseDuration(v.Duration)
	if err != nil {
		panic(err)
	}
	return d
}

var (
	defaultRequestConfigsJSON = "./req_cfg.json"
)

func GenEmptyFile(filename string) error {
	if filename == "" {
		filename = defaultRequestConfigsJSON
	}

	cfgs := RequestConfigs{}
	cfgs = append(cfgs, RequestConfig{
		Duration:          "0s",
		Concurrent:        1,
		TotalCalls:        -1,
		Method:            "",
		URL:               "",
		Headers:           map[string]string{},
		DisableKeepAlives: false,
		Insecure:          false,
		CertFilename:      "",
		KeyFilename:       "",
		Params:            map[string]string{},
		Body:              "",
		Contains:          "",
	})

	b, err := json.MarshalIndent(cfgs, "", "    ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, b, 0666)
}

func ReadRequestConfigsFile(filename string) (RequestConfigs, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	cfgs := RequestConfigs{}
	if err := json.Unmarshal(b, &cfgs); err != nil {
		return nil, err
	}
	return cfgs, nil
}

func ToSaveFilename(filename string) string {
	suffix := fmt.Sprintf("_%d_stat.json", time.Now().Unix())
	if strings.HasSuffix(filename, ".json") {
		return strings.Replace(filename, ".json", suffix, 1)
	}
	return filename + suffix
}

func StatsResultToFile(stats []*node.StatResult, filename string) error {
	b, _ := json.MarshalIndent(stats, "", "\t")
	return ioutil.WriteFile(filename, b, 0666)
}
