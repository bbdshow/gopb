package ps

import (
	"context"
	"fmt"
	"testing"
	"time"
)

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
	pid := int32(3504)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	stat := IntervalReadMemoryUse(ctx, pid, time.Second)
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
