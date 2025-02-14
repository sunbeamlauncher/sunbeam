package cli

import (
	"fmt"

	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/spf13/cobra"
)

func NewCmdEdit() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "edit <extension>",
		Short:             "Edit an extension",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeExtension,
		RunE: func(cmd *cobra.Command, args []string) error {
			entrypoint, err := extensions.FindEntrypoint(utils.ExtensionsDir(), args[0])
			if err != nil {
				return fmt.Errorf("failed to find extension entrypoint: %w", err)
			}

			editCmd, err := utils.EditCmd(entrypoint)
			if err != nil {
				return fmt.Errorf("failed to create editor command: %w", err)
			}

			editCmd.Stdin = cmd.InOrStdin()
			editCmd.Stdout = cmd.OutOrStdout()
			editCmd.Stderr = cmd.ErrOrStderr()

			if err := editCmd.Run(); err != nil {
				return fmt.Errorf("failed to run editor: %w", err)
			}

			return nil
		},
	}

	return cmd
}

func completeExtension(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	extensions, err := LoadExtensions(utils.ExtensionsDir(), false)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var completions []string
	for _, extension := range extensions {
		completions = append(completions, fmt.Sprintf("%s\t%s", extension.Name, extension.Title))
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}
