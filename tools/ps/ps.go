package ps

import (
	"context"
	"github.com/shirou/gopsutil/process"
	"runtime"
	"time"
)

type PS struct {
}

type CPUStat struct {
	UsePercent float64
	Timestamp  int64
}

func (ps *PS) ReadCpuUse(ctx context.Context, pid int32, interval time.Duration) (<-chan *CPUStat, error) {
	if interval.Seconds() <= 0 {
		interval = time.Second
	}
	cpuCh := make(chan *CPUStat, 1)
	p, err := process.NewProcess(pid)
	if err != nil {
		return nil, err
	}
	numCpu := runtime.NumCPU()
	go func() {
		var prev float64
		tInit, err := p.Times()
		if err == nil {
			prev = tInit.Total()
		}
		tick := time.NewTicker(interval)
		for {
			select {
			case <-ctx.Done():
				tick.Stop()
				close(cpuCh)
				return
			case <-tick.C:
				p, err := process.NewProcess(pid)
				if err != nil {
					if err == process.ErrorProcessNotRunning {
						tick.Stop()
						close(cpuCh)
						return
					}
					continue
				}
				t, err := p.Times()
				if err == nil {
					v := t.Total()
					stat := CPUStat{
						UsePercent: 0,
						Timestamp:  time.Now().Unix(),
					}
					stat.UsePercent = (v - prev) / interval.Seconds() / float64(numCpu) * 100
					cpuCh <- &stat

					prev = v
				}
			}
		}
	}()

	return cpuCh, nil
}

type MEMStat struct {
	RSSKb     uint64
	VMSKb     uint64
	Timestamp int64
}

func (ps *PS) ReadMemoryUse(ctx context.Context, pid int32, interval time.Duration) (<-chan *MEMStat, error) {
	if interval.Seconds() <= 0 {
		interval = time.Second
	}
	memCh := make(chan *MEMStat, 1)
	if ok, err := process.PidExists(pid); err != nil {
		return nil, err
	} else {
		if !ok {
			return nil, process.ErrorProcessNotRunning
		}
	}

	go func() {
		tick := time.NewTicker(interval)
		for {
			select {
			case <-ctx.Done():
				tick.Stop()
				close(memCh)
				return
			case <-tick.C:
				p, err := process.NewProcess(pid)
				if err != nil {
					if err == process.ErrorProcessNotRunning {
						tick.Stop()
						close(memCh)
						return
					}
					continue
				}
				stat, err := p.MemoryInfo()
				if err == nil {
					memCh <- &MEMStat{
						RSSKb:     stat.RSS / 1000,
						VMSKb:     stat.VMS / 1000,
						Timestamp: time.Now().Unix(),
					}
				}
			}
		}
	}()

	return memCh, nil
}
