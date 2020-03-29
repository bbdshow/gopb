package main

import (
	"github.com/spf13/cobra"
	"github/huzhongqing/gopb/cmd/cli"
	"log"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "gopb",
		Short: "gopb quickly test performance benchmark http server",
	}

	rootCmd.AddCommand(
		cli.GetStartCmd())

	log.Printf("%s \n", rootCmd.Execute())
}
