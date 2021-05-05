package cmd

import (
	"github.com/go-ksplit/ksplit/pkg/splitter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func CrdSplitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "crdsplit",
		Short:         "split CRDs from non-CRD files",
		Long:          `...`,
		Example:       "ksplit crdsplit myk8syamldir/",
		SilenceErrors: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.Help()
				return errors.New("Please supply a directory")
			}

			err := splitter.MaybeSplitCRDsFs(args[0])
			if err != nil {
				return errors.Wrap(err, "crdsplit cmd")
			}
			return nil
		},
	}

	return cmd
}
