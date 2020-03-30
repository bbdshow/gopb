package cmd

import (
	"context"
	"fmt"
	"github/huzhongqing/gopb/node"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
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

	filename string
)

func init() {

	rootCmd.AddCommand(
		GetStartCmd(),
		GetGenerateCmd(),
		GetStartWithFileCmd())
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

func GetStartWithFileCmd() *cobra.Command {
	startCmd := &cobra.Command{
		Use:     "startWithFile [mode] ",
		Short:   "start with request config file, mode support serial | parallel",
		Example: "startWithFile serial -f ./request_config.json",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgs, err := ReadRequestConfigsFile(filename)
			if err != nil {
				return err
			}
			statChan := make(chan *node.StatResult, len(cfgs))
			go func() {
				switch strings.ToLower(args[0]) {
				case "parallel":
					ParallelDo(cfgs, statChan)
				default:
					SerialDo(cfgs, statChan)
				}
			}()

			for {
				select {
				case stat := <-statChan:
					if stat == nil {
						return nil
					}
					fmt.Println(stat.FormatString())
				}
			}
		},
	}

	startCmd.Flags().StringVarP(&filename, "filename", "f", defaultRequestConfigsJSON, "request configs filename")

	return startCmd
}

func GetGenerateCmd() *cobra.Command {
	generate := &cobra.Command{
		Use:     "generate",
		Short:   "generate empty requests config, support json",
		Example: "generate json",
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GenEmptyFile(filename)
		},
	}

	generate.Flags().StringVarP(&filename, "filename", "f", defaultRequestConfigsJSON, "request configs filename")

	return generate
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

func SerialDo(cfgs RequestConfigs, statChan chan *node.StatResult) {
	cli := node.NewClient()
	for _, cfg := range cfgs {
		ctx := context.TODO()
		if cfg.GetDuration() > 0 {
			ctx, _ = context.WithTimeout(ctx, cfg.GetDuration())
		}
		statChan <- cli.Do(ctx, cfg.Concurrent, cfg.TotalCalls, cfg.ToRequest())
	}
	close(statChan)
}

func ParallelDo(cfgs RequestConfigs, statChan chan *node.StatResult) {
	cli := node.NewClient()
	wg := sync.WaitGroup{}
	for _, cfg := range cfgs {
		wg.Add(1)
		go func(cfg RequestConfig) {
			ctx := context.TODO()
			if cfg.GetDuration() > 0 {
				ctx, _ = context.WithTimeout(ctx, cfg.GetDuration())
			}
			stat := cli.Do(ctx, cfg.Concurrent, cfg.TotalCalls, cfg.ToRequest())
			statChan <- stat

			wg.Done()
		}(cfg)
	}
	wg.Wait()
	close(statChan)
}
