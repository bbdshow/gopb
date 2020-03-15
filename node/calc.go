package node

import (
	"sort"
	"strings"
)

type StatResult struct {
	Concurrent     int
	DoTime         int64
	TotalCalls     int
	HasCalled      int
	Contains       int
	Transferred    int64
	Succeed        int
	Errors         int
	Resp200        int
	Resp300        int
	Resp400        int
	Resp500        int
	Times          []int
	Line99Time     int
	Line95Time     int
	LineMedianTime int
	MaxTime        int
}

func (r StatResult) String() string {
	return ""
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
	}

	// 升序，然后计算时间分布
	sort.Ints(r.Times)
	timeNum := len(r.Times)
	r.Line99Time = r.Times[(timeNum/100*99)] / 1000
	r.Line95Time = r.Times[(timeNum/100*95)] / 1000 // ms
	r.LineMedianTime = r.Times[(timeNum-1)/2] / 1000
	r.MaxTime = r.Times[timeNum-1] / 1000

	return nil
}
