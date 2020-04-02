package node

import (
	"context"
	"crypto/tls"
	"github/huzhongqing/gopb/tools/timing"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Client struct {
	client *http.Client
}

func NewClient() *Client {
	return &Client{client: NewDefaultClient()}
}

func (cli *Client) SetClient(c *http.Client) {
	cli.client = c
}

func (cli *Client) SetDisableKeepAlives(disable bool) {
	cli.client.Transport.(*http.Transport).DisableKeepAlives = disable
}

func (cli *Client) SetTLSConfig(tlsCfg *tls.Config) {
	cli.client.Transport.(*http.Transport).TLSClientConfig = tlsCfg
}

func (cli Client) Do(ctx context.Context, c, n int, req Request) *StatResult {
	if n == -1 {
		_, ok := ctx.Deadline()
		if !ok {
			// 默认时间
			ctx, _ = context.WithTimeout(ctx, 30*time.Second)
		}
	}
	totalTimer := timing.NewTimer()
	totalTimer.Reset()
	if strings.ToUpper(req.Scheme) == "HTTPS" {
		if req.Tls != nil && req.Insecure {
			cli.SetTLSConfig(req.Tls)
		}
	}
	cli.SetDisableKeepAlives(req.DisableKeepAlive)

	request := req.GenHTTPRequest()

	cChan := make(chan struct{}, c)
	respChan := make(chan *Response, c+1)
	statChan := make(chan *StatResult, 1)
	go func() {
		// 计算返回值
		statChan <- ConstantlyCalcStats(request.URL.String(), c, req.ResponseContains, respChan)
	}()
	wg := sync.WaitGroup{}
	for i := 0; i < n || n == -1; i++ {
		select {
		case <-ctx.Done():
			// 未分配的任务，不再分配
			goto exit
		default:
			wg.Add(1)
			cChan <- struct{}{}
			go func() {
				timer := timing.NewTimer()
				timer.Reset()

				// 任务退出，toResponse 会在timeout或者执行完成后退出
				rrCh := make(chan *Response, 1)
				go func() {
					rrCh <- cli.toResponse(timer, request.Clone(context.TODO()), req.ResponseContains != "")
				}()

				select {
				case r := <-rrCh:
					respChan <- r
				case <-ctx.Done():
					// 正在运行的任务，立即返回
				}

				wg.Done()
				<-cChan
			}()
		}
	}
exit:
	wg.Wait()
	respChan <- nil

	stat := <-statChan
	stat.Duration = totalTimer.Duration() / 1e3

	return stat
}

func (cli *Client) toResponse(timer *timing.Timer, req *http.Request, readBody bool) *Response {
	resp, err := cli.client.Do(req)
	obj := &Response{
		RequestSize:  req.ContentLength,
		ResponseSize: 0,
		StatusCode:   0,
		Duration:     timer.Duration(),
		Error:        err,
		Body:         "",
	}
	if err == nil {
		obj.StatusCode = resp.StatusCode
		obj.ResponseSize = resp.ContentLength // 如果 resp.ContentLength = -1 也没关系
		if readBody {
			b, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				obj.Body = string(b)
				obj.ResponseSize = int64(len(b))
			}
		}
		// close
		_ = resp.Body.Close()
	}
	//fmt.Println("response ", *obj)
	return obj
}
