package node

import (
	"context"
	"crypto/tls"
	"github/huzhongqing/gopb/tools/timing"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
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
	totalTimer := timing.NewTimer()
	totalTimer.Reset()

	if strings.ToUpper(req.Scheme) == "HTTPS" {
		if req.Tls != nil {
			cli.SetTLSConfig(req.Tls)
		}
	}
	cli.SetDisableKeepAlives(req.DisableKeepAlive)

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
				timer.Reset()

				// 任务退出，toResponse 会在timeout或者执行完成后退出
				rrCh := make(chan *Response, 1)
				go func() {
					rrCh <- cli.toResponse(timer, req.GenHTTPRequest(), req.ResponseContains != "")
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
	// 计算返回值
	return CalcStats(c, n, totalTimer.Duration(), req.ResponseContains, respChan)
}

func (cli *Client) toResponse(timer *timing.Timer, req *http.Request, readBody bool) *Response {
	resp, err := cli.client.Do(req)
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
	//fmt.Println("response ", *obj)
	return obj
}
