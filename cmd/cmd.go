package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github/huzhongqing/gopb/node"
	"net/url"
	"strings"
	"time"
)

var (
	rootCmd = &cobra.Command{
		Use:     "gopb",
		Short:   "gopb quickly test performance benchmark http server",
		Version: Version(),
	}
)

var (
	// 并发
	concurrent int
	// 请求数
	totalCalls int
	// 持续时间
	duration time.Duration

	headers string

	body string
	// 返回body包含检查
	responseBodyContains string

	disableKeepAlives bool
)

func init() {

	rootCmd.AddCommand(GetStartCmd())
}

func Execute() error {
	return rootCmd.Execute()
}

func Version() string {
	return "0.0.1"
}

func GetStartCmd() *cobra.Command {
	startCmd := &cobra.Command{
		Use:     "start [method] [url]",
		Short:   "setting http method url, start benchmark",
		Example: "start GET http://127.0.0.1:8080",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("args invalid")
			}
			method := args[0]
			switch strings.ToUpper(method) {
			case "GET", "POST", "PUT", "DELETE":
			default:
				return fmt.Errorf("method suppert GET POST PUT DELETE")
			}

			_url, err := url.Parse(args[1])
			if err != nil {
				return err
			}

			req := node.Request{
				Method:           strings.ToUpper(method),
				Scheme:           _url.Scheme,
				URL:              _url.String(),
				Params:           nil,
				Headers:          stringToHeaders(headers),
				Body:             body,
				DisableKeepAlive: disableKeepAlives,
				Insecure:         false,
				Tls:              nil,
				ResponseContains: responseBodyContains,
			}
			ctx := context.Background()

			if duration.Milliseconds() > 0 {
				ctx, _ = context.WithTimeout(ctx, duration)
				totalCalls = -1 // 不限制总数
			}
			stat := node.NewClient().Do(ctx, concurrent, totalCalls, req)
			fmt.Println(stat.FormatString())
			return nil
		},
	}
	startCmd.Flags().IntVarP(&concurrent, "concurrent", "c", 1, "concurrent requests")
	startCmd.Flags().IntVarP(&totalCalls, "totalCalls", "t", -1, "total calls number, -1 no limit")
	startCmd.Flags().DurationVarP(&duration, "duration", "d", 0, "benchmark duration time, total calls is invalid when set")
	startCmd.Flags().StringVar(&headers, "headers", "User-Agent:gopb_benchmark\nContent-Type:text/html;", "headers use '\\n' as the separator ")
	startCmd.Flags().StringVarP(&body, "body", "b", "", "request body")
	startCmd.Flags().StringVar(&responseBodyContains, "contains", "", "response body contains")
	startCmd.Flags().BoolVar(&disableKeepAlives, "disableKeepAlives", false, "disableKeepAlives")
	return startCmd
}

func GenRequestFile() *cobra.Command {
	return &cobra.Command{}
}

func stringToHeaders(v string) map[string]string {
	headers := make(map[string]string)
	kvs := strings.Split(v, "\n")
	for _, kv := range kvs {
		strs := strings.Split(kv, ":")
		if len(strs) == 2 {
			headers[strs[0]] = strs[1]
		}
	}
	return headers
}
