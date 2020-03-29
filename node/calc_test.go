package node

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

func TestCalcStats(t *testing.T) {
	cli := NewClient()
	req := Request{
		Method: "GET",
		Scheme: "http",
		//URL:    "http://127.0.0.1:20001/mock",
		URL:              "http://www.baidu.com",
		Headers:          nil,
		Body:             "",
		DisableKeepAlive: false,
		Insecure:         false,
		Tls:              nil,
		ResponseContains: "o",
	}
	n := -1
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	//ctx := context.Background()
	stat := cli.Do(ctx, 2, n, req)
	fmt.Println(stat.FormatString())

	//time.Sleep(time.Minute)
}

func TestMockServer(t *testing.T) {
	mockServe("20001")
}

func mockServe(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/mock", mockHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), mux))
}

func mockHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Duration(rand.Int31n(10)) * time.Millisecond)
	w.WriteHeader(200)
	w.Write([]byte("ok mock server"))
}
