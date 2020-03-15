package node

import (
	"context"
	"github.com/huzhongqing/httplib"
	"github/huzhongqing/gopb/tools/timing"
	"io/ioutil"
	"net/http"
	"sync"
)

type Client struct {
	clients *httplib.ClientPool
}

func NewClient() Client {
	return Client{clients: httplib.NewClientPool()}
}

func (cli Client) Do(ctx context.Context, c, n int, req Request) *StatResult {
	var tr *http.Transport
	if req.Scheme == "https" {
		tr = &http.Transport{TLSClientConfig: req.Tls, DisableKeepAlives: req.DisableKeepAlive}
	} else {
		tr = &http.Transport{DisableKeepAlives: req.DisableKeepAlive}
	}
	totalTimer := timing.NewTimer()
	totalTimer.Reset()

	cChan := make(chan struct{}, c)
	respChan := make(chan *Response, n+1)
	wg := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		select {
		case <-ctx.Done():
			// 未分配的任务，不再分配
			goto exit
		default:
			wg.Add(1)
			cChan <- struct{}{}
			go func() {
				timer := timing.NewTimer()
				httpClient := cli.clients.PullClient()
				defer func() {
					cli.clients.PushClient(httpClient)
					wg.Done()
					<-cChan
				}()
				request := httpClient.Request(req.URL, req.Method)
				if tr != nil {
					request.SetTransport(*tr)
				}
				for k, v := range req.Headers {
					request.SetHeader(k, v)
				}
				if req.Body != nil {
					request.SetBody(req.Body)
				}
				timer.Reset()

				// 任务退出，toResponse 会在timeout或者执行完成后退出
				rrCh := make(chan *Response, 1)
				go func() {
					rrCh <- toResponse(timer, request, req.ResponseContains != "")
				}()

				select {
				case r := <-rrCh:
					respChan <- r
				case <-ctx.Done():
					// 正在运行的任务，立即返回
				}
			}()
		}
	}
exit:
	wg.Wait()
	// 计算返回值
	return CalcStats(c, n, totalTimer.Duration(), req.ResponseContains, respChan)
}

func toResponse(timer *timing.Timer, request *httplib.HTTPRequest, readBody bool) *Response {
	resp, err := request.Response()
	obj := &Response{
		Size:       0,
		StatusCode: 0,
		Duration:   timer.Duration(),
		Error:      err != nil,
		Body:       "",
	}
	if err == nil {
		obj.StatusCode = resp.StatusCode
		if resp.ContentLength < 0 { // 可能是未知长度
			b, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				obj.Size = int64(len(b))
			}
		} else {
			obj.Size = resp.ContentLength
			if readBody {
				b, err := ioutil.ReadAll(resp.Body)
				if err == nil {
					obj.Body = string(b)
				}
			}
		}
		// close
		_ = resp.Body.Close()
	}
	return obj
}
