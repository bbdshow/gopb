package main

import (
	"context"
	"strconv"
	"time"
)

func main() {
	//cpuRun()
	memRun()
}

func cpuRun() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	var int int
	for {
		select {
		case <-ctx.Done():
			return
		default:
			int++
			strconv.Itoa(int)
		}
	}
}

func memRun() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	var int string
	for {
		select {
		case <-ctx.Done():
			return
		default:
			int += "-"
		}
	}
}
