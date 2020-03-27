package node

import (
	"encoding/json"
	"sort"
	"strings"
)

type StatResult struct {
	Concurrent     int   `json:"concurrent"`
	DoTime         int64 `json:"do_time"`
	TotalCalls     int   `json:"total_calls"`
	HasCalled      int   `json:"has_called"`
	Contains       int   `json:"contains"`
	Transferred    int64 `json:"transferred"`
	Succeed        int   `json:"succeed"`
	Errors         int   `json:"errors"`
	Resp200        int   `json:"resp_2xx"`
	Resp300        int   `json:"resp_3xx"`
	Resp400        int   `json:"resp_4xx"`
	Resp500        int   `json:"resp_5xx"`
	Times          []int `json:"-"`
	Line99Time     int   `json:"line_99_time"`
	Line95Time     int   `json:"line_95_time"`
	LineMedianTime int   `json:"line_median_time"`
	MaxTime        int   `json:"max_time"`
}

func (r StatResult) String() string {
	v, _ := json.Marshal(r)
	return string(v)
}

func CalcStats(c, n int, doTime int64, contains string, stats chan *Response) *StatResult {
	r := &StatResult{
		Concurrent:     c,
		DoTime:         doTime / 1000, //ms
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
		r.Transferred += stat.Size
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

		r.Times[i] = int(stat.Duration)
		i++

		if len(stats) == 0 {
			break
		}
	}

	// 升序，然后计算时间分布
	sort.Ints(r.Times)

	timeNum := len(r.Times)
	r.Line99Time = r.Times[(timeNum/100*99)] / 1000
	r.Line95Time = r.Times[(timeNum/100*95)] / 1000 // ms
	r.LineMedianTime = r.Times[(timeNum-1)/2] / 1000
	r.MaxTime = r.Times[timeNum-1] / 1000

	r.Times = nil

	return r
}
