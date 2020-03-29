package ps

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"
)

func cpuRun() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	var i int
	for {
		select {
		case <-ctx.Done():
			return
		default:
			i++
			strconv.Itoa(i)
		}
	}
}

func memRun() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	var i string
	for {
		select {
		case <-ctx.Done():
			return
		default:
			i += "-"
		}
	}
}

func TestCPU(t *testing.T) {
	pid := int32(12256)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	statCh := IntervalReadCpuUsePercent(ctx, pid, 5*time.Second)
	for {
		select {
		case s := <-statCh:
			if s == nil {
				return
			}
			fmt.Println(s.String())
		}
	}
}

func TestMem(t *testing.T) {
	pid := int32(17656)
	//ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	//defer cancel()
	stat := IntervalReadMemoryUse(context.Background(), pid, time.Second)
	for {
		select {
		case s := <-stat:
			if s == nil {
				return
			}
			fmt.Println(s.String())
		}
	}
}
