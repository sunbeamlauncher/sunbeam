package cli

import (
	"fmt"

	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/spf13/cobra"
)

func NewCmdReload() *cobra.Command {
	return &cobra.Command{
		Use:               "reload <extension>",
		Short:             "Reload an extension",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeExtension,
		RunE: func(cmd *cobra.Command, args []string) error {
			entrypoint, err := extensions.FindEntrypoint(utils.ExtensionsDir(), args[0])
			if err != nil {
				return fmt.Errorf("failed to reload extensions: %w", err)
			}

			ext, err := extensions.LoadExtension(entrypoint, true)
			if err != nil {
				return fmt.Errorf("failed to reload extensions: %w", err)
			}

			cmd.PrintErrln("Extension reloaded:", ext.Name)
			return nil
		},
	}
}
