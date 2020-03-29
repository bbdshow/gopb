package node

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

type StatResult struct {
	URL               string  `json:"url"`
	Concurrent        int     `json:"concurrent"` // 并发
	TotalTime         float64 `json:"total_time"`
	TotalCalls        int     `json:"total_calls"`
	HasCalled         int     `json:"has_called"`
	Contains          int     `json:"contains"`
	ResponseBodySize  int64   `json:"response_body_size"`
	Succeed           int     `json:"succeed"`
	Errors            int     `json:"errors"`
	Resp200           int     `json:"resp_2xx"`
	Resp300           int     `json:"resp_3xx"`
	Resp400           int     `json:"resp_4xx"`
	Resp500           int     `json:"resp_5xx"`
	Times             []int   `json:"-"`
	RequestsPerSecond float64 `json:"requests_per_second"`
	AvgTime           int     `json:"avg_time"`
	LineMedianTime    int     `json:"line_median_time"`
	Line95Time        int     `json:"line_95_time"`
	Line99Time        int     `json:"line_99_time"`
	MaxTime           int     `json:"max_time"`
}

func (r StatResult) String() string {
	v, _ := json.Marshal(r)
	return string(v)
}

func (r StatResult) FormatString() string {
	return fmt.Sprintf(`========== Performance benchmark ==========
URL: %s
Concurrent: %d
Total calls: %d
Has called: %d
Succeed: %d
Error: %d
Response body size: %d(byte)
========== Times ==========
Total time: %s
Requests per second: %.2f
Avg time per request: %s
Median time per request: %s
95th percentile time: %s
99th percentile time: %s
Slowest time for request: %s
========== Status ==========
Status code 2xx: %d
Status code 3xx: %d
Status code 4xx: %d
Status code 5xx: %d
Match Response: %d`,
		r.URL, r.Concurrent, r.TotalCalls, r.HasCalled, r.Succeed, r.Errors, r.ResponseBodySize,
		timeMillToString(int(r.TotalTime)), r.RequestsPerSecond, timeMillToString(r.AvgTime), timeMillToString(r.LineMedianTime),
		timeMillToString(r.Line95Time), timeMillToString(r.Line99Time), timeMillToString(r.MaxTime),
		r.Resp200, r.Resp300, r.Resp400, r.Resp500, r.Contains)
}

func timeMillToString(t int) string {
	return (time.Duration(t) * time.Millisecond).String()
}

func CalcStats(url string, c, n int, contains string, stats chan *Response) *StatResult {
	r := &StatResult{
		URL:            url,
		Concurrent:     c,
		TotalCalls:     n,
		HasCalled:      len(stats),
		Contains:       0,
		Succeed:        0,
		Resp200:        0,
		Resp300:        0,
		Resp400:        0,
		Resp500:        0,
		Times:          make([]int, len(stats)),
		Line99Time:     0,
		Line95Time:     0,
		LineMedianTime: 0,
	}

	if r.HasCalled == 0 {
		return r
	}
	i := 0
	for stat := range stats {
		if stat == nil {
			break
		}
		if stat.Error {
			r.Errors++
		}
		r.ResponseBodySize += stat.Size
		if len(contains) > 0 && len(stat.Body) > 0 {
			if strings.Contains(stat.Body, contains) {
				r.Contains++
			}
		}
		switch {
		case stat.StatusCode < 200:
		case stat.StatusCode < 300:
			r.Resp200++
			r.Succeed++
		case stat.StatusCode < 400:
			r.Resp300++
		case stat.StatusCode < 500:
			r.Resp400++
		case stat.StatusCode < 600:
			r.Resp500++
		}
		r.TotalTime += float64(stat.Duration)
		r.Times[i] = int(stat.Duration)
		i++

		if len(stats) == 0 {
			break
		}
	}

	// 升序，然后计算时间分布
	sort.Ints(r.Times)

	timeNum := len(r.Times)
	r.TotalTime = r.TotalTime / 1e3
	r.RequestsPerSecond = float64(timeNum) / (r.TotalTime / 1e3)
	r.AvgTime = int(r.TotalTime / float64(timeNum))
	r.LineMedianTime = r.Times[(timeNum-1)/2] / 1000
	r.Line95Time = r.Times[(timeNum/100*95)] / 1000 // ms
	r.Line99Time = r.Times[(timeNum/100*99)] / 1000
	r.MaxTime = r.Times[timeNum-1] / 1000

	r.Times = nil

	return r
}
