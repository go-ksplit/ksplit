package cmd

import (
	"fmt"
	"time"

	"github.com/go-ksplit/ksplit/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Version() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "version",
		Short:         "ksplit version information",
		Long:          `Prints the current version of ksplit`,
		SilenceErrors: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			v := viper.GetViper()
			if v.GetBool("verbose") {
				fmt.Printf("ksplit %s built at %s with sha %s\n", version.Version(), version.BuildTime().Format(time.RFC3339), version.GitSHA())
			} else {
				fmt.Printf("ksplit %s\n", version.Version())
			}

			return nil
		},
	}

	cmd.Flags().Bool("verbose", false, "when set, also print build info")

	return cmd
}
