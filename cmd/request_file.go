package cmd

import (
	"encoding/json"
	"io/ioutil"
)

// 批量请求多个任务

type RequestConfigs []RequestConfig

type RequestConfig struct {
	Duration          string            `json:"duration"`
	Concurrent        int               `json:"concurrent"`
	TotalCalls        int               `json:"total_calls"`
	Method            string            `json:"method"`
	Scheme            string            `json:"scheme"`
	Headers           map[string]string `json:"headers"`
	DisableKeepAlives bool              `json:"disable_keep_alives"`
	Insecure          bool              `json:"insecure"` // 建立不安全连接
	CertFilename      string            `json:"cert_filename"`
	KeyFilename       string            `json:"key_filename"`
	Params            map[string]string `json:"params"`
	Body              string            `json:"body"`
	Contains          string            `json:"contains"`
}

var (
	defaultRequestConfigsJSON = "./request_configs.json"
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
		Scheme:            "",
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
