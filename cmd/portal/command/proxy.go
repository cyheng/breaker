package command

import (
	"breaker/feature"
	"breaker/pkg/errwrap"
	"breaker/plugin"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	cmdRoot.AddCommand(proxyCmd)
}

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "http proxy",
	Run: func(cmd *cobra.Command, args []string) {
		conf := &feature.PluginHttpProxy{}
		err := feature.LoadFromFile(cfgFile, conf)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = errwrap.PanicToError(func() {
			conf.OnInit()
		})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		pxy := &plugin.HttpProxy{}
		addr := net.JoinHostPort("0.0.0.0", fmt.Sprint(conf.ProxyPort))
		fmt.Printf("proxy listen at :%v", addr)
		if err := http.ListenAndServe(addr, pxy); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}
