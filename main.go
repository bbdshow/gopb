package main

import (
	"github/huzhongqing/gopb/cmd"
	"log"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
