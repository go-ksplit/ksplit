package cmd

import (
	"github.com/go-ksplit/ksplit/pkg"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func AllSplitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "all",
		Short:         "split kubernetes yaml within a directory",
		Long:          `ksplit reformats generated kubernetes yaml into a more easily readable file format`,
		Example:       "ksplit all myk8syamldir/",
		SilenceErrors: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.Help()
				return errors.New("Please supply a directory")
			}

			err := pkg.MaybeSplitMultidocYamlFs(args[0])
			if err != nil {
				return errors.Wrap(err, "allsplit cmd")
			}
			return nil
		},
	}

	return cmd
}
