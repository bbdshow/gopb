package node

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

func TestCalcStats(t *testing.T) {
	go func() {
		mockServe("20001")
	}()
	time.Sleep(2 * time.Second)
	cli := NewClient()
	req := Request{
		Method: "GET",
		Scheme: "http",
		URL:    "http://127.0.0.1:20001/mock",
		//URL:              "http://www.baidu.com",
		Headers:          nil,
		Body:             "",
		DisableKeepAlive: false,
		Insecure:         false,
		Tls:              nil,
		ResponseContains: "o",
	}
	n := 100
	stat := cli.Do(context.Background(), 1, n, req)
	fmt.Println(stat.FormatString())
	assert.Equal(t, stat.HasCalled, n, fmt.Sprintf("hasCalled must equal %d", n))

	//time.Sleep(time.Minute)
}

func mockServe(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/mock", mockHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), mux))
}

func mockHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Duration(rand.Int31n(100)) * time.Millisecond)
	w.WriteHeader(200)
	w.Write([]byte("ok"))
}
