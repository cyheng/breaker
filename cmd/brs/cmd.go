package main

import (
	"breaker/app"
	"breaker/pkg/config"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

const (
	version = "1.0.0"
)

var (
	cfgFile     string
	showVersion bool
)

var cmdRoot = &cobra.Command{
	Use: "breaker",
	Run: func(cmd *cobra.Command, args []string) {
		if showVersion {
			fmt.Println(version)
		}
		features, err := config.LoadFromFile(cfgFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if len(features) == 0 {
			fmt.Println("at lease one feature required")
			os.Exit(1)
		}
		breaker := app.New(features...)

		if breaker.Run() != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// 后台运行
		{
			osSignals := make(chan os.Signal, 1)
			signal.Notify(osSignals, os.Interrupt, os.Kill, syscall.SIGTERM)
			<-osSignals
		}
	},
}

func init() {
	cmdRoot.PersistentFlags().StringVarP(&cfgFile, "config", "c", "config.ini", "config file of breaker")
	cmdRoot.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "version of breaker")

}
