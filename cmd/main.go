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
