package timing

import (
	"sync"
	"time"
)

type Timing struct {
	Enable  bool
	mutex   sync.Mutex
	methods map[string]*Waiting
}

func NewTiming(enable bool) *Timing {
	t := &Timing{
		Enable:  enable,
		mutex:   sync.Mutex{},
		methods: make(map[string]*Waiting),
	}
	return t
}

type Waiting struct {
	Count int64
	Total int64
	Max   int64
	Min   int64
}

func (t *Timing) Do(method string, f func()) {
	if !t.enable() {
		f()
		return
	}
	start := time.Now()
	f()
	t.waiting(method, time.Now().Sub(start).Nanoseconds())
}

func (t *Timing) waiting(method string, nano int64) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	w, ok := t.methods[method]
	if !ok {
		t.methods[method] = &Waiting{
			Count: 1,
			Total: nano,
			Max:   nano,
			Min:   nano,
		}
		return
	}
	w.Total += nano
	w.Count++

	if nano > w.Max {
		w.Max = nano
	}
	if nano < w.Min {
		w.Min = nano
	}
}

type MethodData struct {
	Method        string
	DoCount       int64
	TimeNanoTotal int64
	MaxTimeNano   int64
	MinTimeNano   int64
	AvgTimeNano   int64
}

func (t *Timing) GetMethodData() []*MethodData {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	data := make([]*MethodData, 0, len(t.methods))
	for k, v := range t.methods {
		d := &MethodData{
			Method:        k,
			DoCount:       v.Count,
			TimeNanoTotal: v.Total,
			MaxTimeNano:   v.Max,
			MinTimeNano:   v.Min,
			AvgTimeNano:   v.Total / v.Count,
		}
		data = append(data, d)
	}

	return data
}

func (t *Timing) SetEnable(enable bool) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.Enable = enable
}

func (t *Timing) Clear() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.methods = make(map[string]*Waiting)
	return
}

func (t *Timing) enable() bool {
	return t.Enable
}
