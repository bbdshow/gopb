package node

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
)

type StatResult struct {
	URL               string  `json:"url"`
	Concurrent        int     `json:"concurrent"` // 并发
	Duration          int64   `json:"duration"`
	SumTime           float64 `json:"sum_time"`
	CallsNumber       int     `json:"calls_number"`
	Contains          int     `json:"contains"`
	RequestSize       int64   `json:"request_size"`
	ResponseSize      int64   `json:"response_size"`
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
	return fmt.Sprintf(`========== Benchmark ==========
URL: %s
Concurrent: %d
Total calls number: %d
Succeed: %d
Error: %d
Request body size: %s
Response body size: %s
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
Match response: %d`,
		r.URL, r.Concurrent, r.CallsNumber, r.Succeed, r.Errors, byteSizeToString(r.RequestSize), byteSizeToString(r.ResponseSize),
		timeMillToString(int(r.Duration)), r.RequestsPerSecond, timeMillToString(r.AvgTime), timeMillToString(r.LineMedianTime),
		timeMillToString(r.Line95Time), timeMillToString(r.Line99Time), timeMillToString(r.MaxTime),
		r.Resp200, r.Resp300, r.Resp400, r.Resp500, r.Contains)
}

func timeMillToString(t int) string {
	return (time.Duration(t) * time.Millisecond).String()
}

func byteSizeToString(s int64) string {
	size := float64(s)
	switch {
	case size >= 1e4 && size < 1e7:
		return fmt.Sprintf("%.2f KB", size/1024)
	case size >= 1e7 && size < 1e10:
		return fmt.Sprintf("%.2f MB", size/(1024*1024))
	case size >= 1e10:
		return fmt.Sprintf("%.2f GB", size/(1024*1024*1024))
	}
	return fmt.Sprintf("%d b", s)
}

func ConstantlyCalcStats(url string, c int, contains string, stats chan *Response) *StatResult {
	var errCount int64
	r := &StatResult{
		URL:        url,
		Concurrent: c,
		Times:      make([]int, 0, c),
	}

	for {
		select {
		case stat := <-stats:
			if stat == nil {
				goto exitFor
			}
			r.CallsNumber++
			if stat.Error != nil {
				if errCount < 10 || errCount%1000 == 0 {
					log.Printf("error count %d: %s", errCount+1, stat.Error.Error())
				}
				errCount++
				r.Errors++
			}
			r.RequestSize += stat.RequestSize
			r.ResponseSize += stat.ResponseSize
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
			r.SumTime += float64(stat.Duration)
			r.Times = append(r.Times, int(stat.Duration))

			continue
		}
	exitFor:
		break
	}

	if len(r.Times) == 0 {
		return r
	}
	// 升序，然后计算时间分布
	sort.Ints(r.Times)

	timeNum := len(r.Times)
	r.SumTime = r.SumTime / 1e3
	r.RequestsPerSecond = float64(timeNum) / (r.SumTime / 1e3)
	r.AvgTime = int(r.SumTime / float64(timeNum))
	r.LineMedianTime = r.Times[(timeNum-1)/2] / 1000
	r.Line95Time = r.Times[(timeNum/100*95)] / 1000 // ms
	r.Line99Time = r.Times[(timeNum/100*99)] / 1000
	r.MaxTime = r.Times[timeNum-1] / 1000

	r.Times = nil

	return r
}
