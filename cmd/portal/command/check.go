package command

import (
	"breaker/feature"
	"breaker/pkg/errwrap"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	cmdRoot.AddCommand(checkCmd)
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "check configuration",
	Run: func(cmd *cobra.Command, args []string) {
		conf := &feature.PortalConfig{}
		err := feature.LoadFromFile(cfgFile, conf)
		if err != nil {

		}
		err = errwrap.PanicToError(func() {
			conf.OnInit()
		})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("portal: the configuration file %s syntax is ok\n", cfgFile)
	},
}
