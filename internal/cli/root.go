package cli

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"

	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/history"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/sunbeam"
	"github.com/spf13/cobra"
)

var (
	Version = "dev"
)

func NewRootCmd() (*cobra.Command, error) {
	var flags struct {
		reload bool
	}
	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:          "sunbeam",
		Short:        "Command Line Launcher",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

See https://pomdtr.github.io/sunbeam for more information.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if flags.reload {
				exts, err := LoadExtensions(utils.ExtensionsDir(), true)
				if err != nil {
					return fmt.Errorf("failed to reload extensions: %w", err)
				}

				fmt.Fprintf(os.Stderr, "Reloaded %d extensions\n", len(exts))
			}

			if !term.IsTerminal(int(os.Stdout.Fd())) {
				exts, err := LoadExtensions(utils.ExtensionsDir(), false)
				if err != nil {
					return nil
				}

				encoder := json.NewEncoder(os.Stdout)
				encoder.SetEscapeHTML(false)
				return encoder.Encode(exts)
			}

			history, err := history.Load(history.Path)
			if err != nil {
				return err
			}

			rootList := tui.NewRootList(history, func() ([]sunbeam.ListItem, error) {
				exts, err := LoadExtensions(utils.ExtensionsDir(), false)
				if err != nil {
					return nil, err
				}

				var items []sunbeam.ListItem
				for _, extension := range exts {
					items = append(items, extension.RootItems()...)
				}

				return items, nil
			})
			return tui.Draw(rootList)

		},
	}

	rootCmd.Flags().BoolVar(&flags.reload, "reload", false, "Reload extension manifests")
	rootCmd.AddCommand(NewCmdValidate())
	rootCmd.AddCommand(NewCmdFetch())
	rootCmd.AddCommand(NewCmdServe())
	rootCmd.AddCommand(NewCmdEdit())

	rootCmd.AddGroup(&cobra.Group{
		ID:    "extension",
		Title: "Extensions Commands:",
	})

	exts, err := LoadExtensions(utils.ExtensionsDir(), false)
	if errors.Is(err, os.ErrNotExist) {
		return rootCmd, nil
	} else if err != nil {
		return nil, err
	}

	for _, extension := range exts {
		command, err := NewCmdExtension(extension.Name, extension)
		if err != nil {
			return nil, err
		}
		rootCmd.AddCommand(command)
	}

	return rootCmd, nil
}

func LoadExtensions(extensionDir string, reload bool) ([]extensions.Extension, error) {
	extensionMap := make(map[string]struct{})
	exts := make([]extensions.Extension, 0)
	entries, err := os.ReadDir(extensionDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		extension, err := extensions.LoadExtension(filepath.Join(extensionDir, entry.Name()), reload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load extension %s: %v", entry.Name(), err)
			continue
		}

		if _, ok := extensionMap[extension.Name]; ok {
			fmt.Fprintf(os.Stderr, "duplicate extension alias: %s", extension.Name)
			continue
		}

		extensionMap[extension.Name] = struct{}{}

		exts = append(exts, extension)
	}

	return exts, nil
}
