package ps

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestCPU(t *testing.T) {
	ps := &PS{}
	pid := int32(23384)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	state, err := ps.ReadCpuUse(ctx, pid, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	for {
		select {
		case s := <-state:
			if s == nil {
				return
			}
			fmt.Println("use", s.UsePercent, "Timestamp", s.Timestamp)
		}
	}
}

func TestMem(t *testing.T) {
	pid := int32(2928)
	ps := &PS{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	stat, err := ps.ReadMemoryUse(ctx, pid, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	for {
		select {
		case s := <-stat:
			if s == nil {
				return
			}
			fmt.Println("rss", s.RSSKb, "vm", s.VMSKb, "Timestamp", s.Timestamp)
		}
	}
}
