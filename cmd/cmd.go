package cmd

import (
	"context"
	"fmt"
	"github/huzhongqing/gopb/node"
	"log"
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

	filename   string
	mode       string
	resultSave bool
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
	startCmd.Flags().BoolVar(&disableKeepAlives, "disable-keep-alives", false, "disableKeepAlives")
	startCmd.Flags().BoolVarP(&resultSave, "result-save", "s", false, "requests the result stats to save the file")
	return startCmd
}

func GetStartWithFileCmd() *cobra.Command {
	startCmd := &cobra.Command{
		Use:     "start-with-file",
		Short:   "start with request config file",
		Example: "start-with-file -m parallel -f ./request_config.json",
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgs, err := ReadRequestConfigsFile(filename)
			if err != nil {
				return err
			}
			statChan := make(chan *node.StatResult, len(cfgs))
			go func() {
				switch strings.ToLower(mode) {
				case "parallel":
					ParallelDo(cfgs, statChan)
				default:
					SerialDo(cfgs, statChan)
				}
			}()

			stats := make([]*node.StatResult, 0, len(cfgs))
			for {
				select {
				case stat := <-statChan:
					if stat == nil {
						if resultSave {
							return StatsResultToFile(stats, ToSaveFilename(filename))
						}
						return nil
					}
					if resultSave {
						stats = append(stats, stat)
					} else {
						fmt.Println(stat.FormatString())
					}
				}
			}
		},
	}

	startCmd.Flags().StringVarP(&filename, "filename", "f", defaultRequestConfigsJSON, "request configs filename")
	startCmd.Flags().StringVarP(&mode, "mode", "m", "serial", "file requests mode support serial | parallel")
	startCmd.Flags().BoolVarP(&resultSave, "result-save", "s", false, "requests the result stats to save the file")
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
		request, err := cfg.ToRequest()
		if err != nil {
			log.Printf("tls config %s \n", err.Error())
			break
		}
		statChan <- cli.Do(ctx, cfg.Concurrent, cfg.TotalCalls, request)
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
			request, err := cfg.ToRequest()
			if err != nil {
				log.Printf("tls config %s \n", err.Error())
			} else {
				stat := cli.Do(ctx, cfg.Concurrent, cfg.TotalCalls, request)
				statChan <- stat
			}

			wg.Done()
		}(cfg)
	}
	wg.Wait()
	close(statChan)
}
