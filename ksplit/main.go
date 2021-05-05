package main

import (
	"fmt"
	"os"
	"strings"

	cmd2 "github.com/go-ksplit/ksplit/ksplit/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	if err := RootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{}

	cmd.PersistentFlags().String("log-level", "off", "Log level")

	cmd.AddCommand(cmd2.CrdSplitCmd())
	cmd.AddCommand(cmd2.AllSplitCmd())
	cmd.AddCommand(cmd2.Version())

	_ = viper.BindPFlags(cmd.Flags())
	_ = viper.BindPFlags(cmd.PersistentFlags())
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	return cmd
}
