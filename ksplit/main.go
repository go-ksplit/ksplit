package ksplit

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/go-ksplit/ksplit/pkg"
)

func main() {
	if err := RootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "ksplit directory",
		Short:         "split kubernetes yaml within a directory",
		Long:          `ksplit reformats generated kubernetes yaml into a more easily readable file format`,
		Example:       "ksplit myk8syamldir/",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.Help()
				return errors.New("Please supply a directory")
			}

			err := pkg.MaybeSplitMultidocYamlFs(args[0])
			if err != nil {
				return errors.Wrap(err, "root cmd")
			}
			return nil
		},
	}

	cmd.PersistentFlags().String("log-level", "off", "Log level")

	_ = viper.BindPFlags(cmd.Flags())
	_ = viper.BindPFlags(cmd.PersistentFlags())
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	return cmd
}
