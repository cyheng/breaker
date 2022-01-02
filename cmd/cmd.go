package main

import (
	"breaker/pkg/config"
	"breaker/services"
	"fmt"
	log "github.com/sirupsen/logrus"
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
		config, err := config.LoadFromFile(cfgFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)

		}
		err = services.Run(config.ServiceName(), config)
		if err != nil {
			log.Error(err)
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
