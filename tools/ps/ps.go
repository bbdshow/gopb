package ps

import (
	"context"
	"encoding/json"
	"github.com/shirou/gopsutil/process"
	"runtime"
	"time"
)

type CPUStat struct {
	UsePercent float64 `json:"use_percent"`
	Timestamp  int64   `json:"timestamp"`
}

func (s CPUStat) String() string {
	v, _ := json.Marshal(s)
	return string(v)
}
func IntervalReadCpuUsePercent(ctx context.Context, pid int32, interval time.Duration) <-chan *CPUStat {
	if interval.Seconds() <= 0 {
		interval = time.Second
	}
	numCpu := runtime.NumCPU()
	cpuCh := make(chan *CPUStat, 1)
	var preUse float64
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(cpuCh)
				return
			default:
				stat := CPUStat{
					UsePercent: 0,
					Timestamp:  time.Now().Unix(),
				}
				use, err := cpuUse(pid)
				if err != nil && err == process.ErrorProcessNotRunning {
					close(cpuCh)
					return
				}

				if preUse <= 0 {
					preUse = use
					break
				}

				stat.UsePercent = (use - preUse) / interval.Seconds() / float64(numCpu) * 100
				cpuCh <- &stat

				preUse = use
			}
			time.Sleep(interval)
		}
	}()

	return cpuCh
}

func cpuUse(pid int32) (float64, error) {
	p, err := process.NewProcess(pid)
	if err != nil {
		return 0, err
	}
	t, err := p.Times()
	if err != nil {
		return 0, err
	}
	return t.Total(), nil
}

type MEMStat struct {
	RSS       uint64 `json:"rss"` // bytes
	VMS       uint64 `json:"vms"`
	Stack     uint64 `json:"stack"`
	Swap      uint64 `json:"swap"`
	Timestamp int64  `json:"timestamp"`
}

func (s MEMStat) String() string {
	v, _ := json.Marshal(s)
	return string(v)
}

func IntervalReadMemoryUse(ctx context.Context, pid int32, interval time.Duration) <-chan *MEMStat {
	if interval.Seconds() <= 0 {
		interval = time.Second
	}
	memCh := make(chan *MEMStat, 1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(memCh)
				return
			default:
				p, err := process.NewProcess(pid)
				if err != nil && err == process.ErrorProcessNotRunning {
					close(memCh)
					return
				}
				if p == nil {
					break
				}

				stat, err := p.MemoryInfo()
				if err == nil {
					memCh <- &MEMStat{
						RSS:       stat.RSS,
						VMS:       stat.VMS,
						Stack:     stat.Stack,
						Swap:      stat.Swap,
						Timestamp: time.Now().Unix(),
					}
				}
			}
			time.Sleep(interval)
		}
	}()

	return memCh
}
