package node

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"testing"
	"time"
)

func init() {
	go func() {
		mockServe("10001")
	}()
}

func TestCalcStats(t *testing.T) {
	time.Sleep(time.Second)
	cli := NewClient()
	req := Request{
		Method:           "GET",
		Scheme:           "http",
		URL:              "http://127.0.0.1:10001",
		Headers:          nil,
		Body:             nil,
		DisableKeepAlive: false,
		Insecure:         false,
		Tls:              nil,
		ResponseContains: "GET",
	}
	n := 10
	stat := cli.Do(context.Background(), 1, n, req)
	fmt.Println(stat.String())
	assert.Equal(t, stat.HasCalled, n, fmt.Sprintf("hasCalled must equal %d", n))
}

func mockServe(port string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.WriteHeader(200)
			w.Write([]byte("GET METHOD"))
		case "POST":
			w.WriteHeader(200)
			w.Write([]byte("POST METHOD"))
		}
		return
	})
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatal(err)
	}
}
