package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"golang.org/x/term"

	"github.com/atotto/clipboard"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/history"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/sunbeam"
	"github.com/spf13/cobra"
)

func NewCmdExtension(alias string, extension extensions.Extension) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:     alias,
		Short:   extension.Manifest.Title,
		Long:    extension.Manifest.Description,
		Args:    cobra.NoArgs,
		GroupID: "extension",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !term.IsTerminal(int(os.Stdout.Fd())) {
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				encoder.SetEscapeHTML(false)

				return encoder.Encode(extension.Manifest)
			}

			if len(extension.Manifest.Root) == 0 {
				return cmd.Usage()
			}

			history, err := history.Load(history.Path)
			if err != nil {
				return err
			}

			rootList := tui.NewRootList(history, func() ([]sunbeam.ListItem, error) {
				return extension.RootItems(), nil
			})

			return tui.Draw(rootList)
		},
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	commands := extension.Manifest.Commands
	sort.Slice(extension.Manifest.Commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})

	for _, command := range commands {
		cmd := NewSubCmdCustom(alias, extension, command)
		rootCmd.AddCommand(cmd)
	}

	return rootCmd, nil
}

func NewSubCmdCustom(alias string, extension extensions.Extension, command sunbeam.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   command.Name,
		Short: command.Description,
		RunE: func(cmd *cobra.Command, args []string) error {
			params := make(map[string]any)
			if !term.IsTerminal(int(os.Stdin.Fd())) {
				bts, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}

				if len(bts) > 0 {
					err = json.Unmarshal(bts, &params)
					if err != nil {
						return err
					}
				}
			}

			for _, param := range command.Params {
				if !cmd.Flags().Changed(param.Name) {
					if _, ok := params[param.Name]; ok {
						continue
					}

					if param.Optional {
						continue
					}

					return fmt.Errorf("missing required input: %s", param.Name)
				}

				switch param.Type {
				case sunbeam.InputString:
					value, err := cmd.Flags().GetString(param.Name)
					if err != nil {
						return err
					}
					params[param.Name] = value
				case sunbeam.InputBoolean:
					value, err := cmd.Flags().GetBool(param.Name)
					if err != nil {
						return err
					}
					params[param.Name] = value
				case sunbeam.InputNumber:
					value, err := cmd.Flags().GetInt(param.Name)
					if err != nil {
						return err
					}
					params[param.Name] = value
				}
			}

			return runExtension(extension, command, params)
		},
	}

	for _, input := range command.Params {
		switch input.Type {
		case sunbeam.InputString:
			cmd.Flags().String(input.Name, "", input.Description)
		case sunbeam.InputBoolean:
			cmd.Flags().Bool(input.Name, false, input.Description)
		case sunbeam.InputNumber:
			cmd.Flags().Int(input.Name, 0, input.Description)
		}

		if !input.Optional && term.IsTerminal(int(os.Stdin.Fd())) {
			_ = cmd.MarkFlagRequired(input.Name)
		}
	}

	return cmd
}

func runExtension(extension extensions.Extension, command sunbeam.Command, params sunbeam.Params) error {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		cmd, err := extension.CmdContext(context.Background(), command, params)
		if err != nil {
			return err
		}

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		return cmd.Run()
	}

	switch command.Mode {
	case sunbeam.CommandModeSearch, sunbeam.CommandModeFilter, sunbeam.CommandModeDetail:
		runner := tui.NewRunner(extension, command, params)
		return tui.Draw(runner)
	case sunbeam.CommandModeSilent:
		cmd, err := extension.CmdContext(context.Background(), command, params)
		if err != nil {
			return err
		}

		return cmd.Run()
	case sunbeam.CommandModeAction:
		output, err := extension.Output(context.Background(), command, params)
		if err != nil {
			return fmt.Errorf("failed to run command: %w", err)
		}

		var action sunbeam.Action
		if err = json.Unmarshal(output, &action); err != nil {
			return fmt.Errorf("failed to unmarshal action: %w", err)
		}

		switch action.Type {
		case sunbeam.ActionTypeRun:
			command, ok := extension.GetCommand(action.Run.Command)
			if !ok {
				return fmt.Errorf("command not found: %s", action.Run.Command)
			}

			return runExtension(extension, command, action.Run.Params)
		case sunbeam.ActionTypeReload:
			return nil
		case sunbeam.ActionTypeOpen:
			return utils.Open(action.Open.Url)
		case sunbeam.ActionTypeCopy:
			return clipboard.WriteAll(action.Copy.Text)
		default:
			return fmt.Errorf("unknown action type: %s", action.Type)
		}
	default:
		return fmt.Errorf("unknown command mode: %s", command.Mode)
	}

}
