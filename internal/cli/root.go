package cli

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/pomdtr/sunbeam/internal/config"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/tui"
	"github.com/pomdtr/sunbeam/internal/types"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var (
	Version = "dev"
)

const (
	CommandGroupCore      = "core"
	CommandGroupExtension = "extension"
)

func IsSunbeamRunning() bool {
	return len(os.Getenv("SUNBEAM")) > 0
}

type NonInteractiveOutput struct {
	Extensions []extensions.Extension `json:"extensions"`
	Items      []types.ListItem       `json:"items"`
}

//go:embed embed/sunbeam.json
var configBytes []byte

func NewRootCmd() (*cobra.Command, error) {
	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:                "sunbeam",
		Short:              "Command Line Launcher",
		SilenceUsage:       true,
		DisableFlagParsing: true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return nil, cobra.ShellCompDirectiveDefault
			}

			entrypoint, err := filepath.Abs(args[0])
			if err != nil {
				return nil, cobra.ShellCompDirectiveDefault
			}

			extension, err := extensions.ExtractManifest(entrypoint)
			if err != nil {
				return nil, cobra.ShellCompDirectiveDefault
			}

			var completions []string
			for _, command := range extension.Commands {
				completions = append(completions, fmt.Sprintf("%s\t%s", command.Name, command.Title))
			}

			return completions, cobra.ShellCompDirectiveNoFileComp
		},
		Args: cobra.ArbitraryArgs,
		Long: `Sunbeam is a command line launcher for your terminal, inspired by fzf and raycast.

See https://pomdtr.github.io/sunbeam for more information.`,
	}

	rootCmd.AddGroup(&cobra.Group{
		ID:    CommandGroupCore,
		Title: "Core Commands:",
	})
	rootCmd.AddCommand(NewCmdQuery())
	rootCmd.AddCommand(NewValidateCmd())
	rootCmd.AddCommand(NewCmdFetch())
	rootCmd.AddCommand(NewCmdEdit())
	rootCmd.AddCommand(NewCmdCopy())
	rootCmd.AddCommand(NewCmdPaste())
	rootCmd.AddCommand(NewCmdOpen())

	docCmd := &cobra.Command{
		Use:    "docs",
		Short:  "Generate documentation for sunbeam",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, err := buildDoc(rootCmd)
			if err != nil {
				return err
			}

			fmt.Print(heredoc.Docf(`---
			outline: 2
			---

			# Cli

			%s
			`, doc))
			return nil
		},
	}
	rootCmd.AddCommand(docCmd)

	manCmd := &cobra.Command{
		Use:    "generate-man-pages [path]",
		Short:  "Generate Man Pages for sunbeam",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			header := &doc.GenManHeader{
				Title:   "MINE",
				Section: "3",
			}
			err := doc.GenManTree(rootCmd, header, args[0])
			if err != nil {
				return err
			}

			return nil
		},
	}
	rootCmd.AddCommand(manCmd)

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of sunbeam",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(Version)
		},
	}
	rootCmd.AddCommand(versionCmd)

	if IsSunbeamRunning() {
		return rootCmd, nil
	}

	rootCmd.AddGroup(&cobra.Group{
		ID:    CommandGroupExtension,
		Title: "Extension Commands:",
	})

	if _, err := os.Stat(config.Path); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(config.Path), 0755); err != nil {
			return nil, err
		}

		if err := os.WriteFile(config.Path, []byte(configBytes), 0644); err != nil {
			return nil, err
		}
	}

	cfg, err := config.Load(config.Path)
	if err != nil {
		return nil, err
	}
	rootCmd.AddCommand(NewCmdExtension(cfg))

	extensionMap := make(map[string]extensions.Extension)
	for alias, extensionConfig := range cfg.Extensions {
		extension, err := extensions.LoadExtension(extensionConfig.Origin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading extension %s: %s\n", alias, err)
			continue
		}
		extensionMap[alias] = extension

		command, err := NewCmdCustom(alias, extension, extensionConfig)
		if err != nil {
			return nil, err
		}
		rootCmd.AddCommand(command)
	}

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
			return cmd.Help()
		}

		if len(args) == 0 {
			if len(extractListItems(cfg, extensionMap)) == 0 {
				return cmd.Usage()
			}

			rootList := tui.NewRootList("Sunbeam", func() (config.Config, []types.ListItem, error) {
				cfg, err := config.Load(config.Path)
				if err != nil {
					return config.Config{}, nil, err
				}

				extensionMap := make(map[string]extensions.Extension)
				for alias, extensionConfig := range cfg.Extensions {
					extension, err := extensions.LoadExtension(extensionConfig.Origin)
					if err != nil {
						continue
					}
					extensionMap[alias] = extension
				}

				return cfg, extractListItems(cfg, extensionMap), nil
			})
			return tui.Draw(rootList)
		}

		var entrypoint string
		if args[0] == "-" {
			tempfile, err := os.CreateTemp("", "entrypoint-*%s")
			if err != nil {
				return err
			}
			defer os.Remove(tempfile.Name())

			if _, err := io.Copy(tempfile, os.Stdin); err != nil {
				return err
			}

			if err := tempfile.Close(); err != nil {
				return err
			}

			entrypoint = tempfile.Name()
		} else if extensions.IsRemote(args[0]) {
			tempfile, err := os.CreateTemp("", "entrypoint-*%s")
			if err != nil {
				return err
			}
			defer os.Remove(tempfile.Name())

			if err := extensions.DownloadEntrypoint(args[0], tempfile.Name()); err != nil {
				return err
			}

			entrypoint = tempfile.Name()
		} else {
			e, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}

			if _, err := os.Stat(e); err != nil {
				return fmt.Errorf("error loading extension: %w", err)
			}

			entrypoint = e
		}

		if err := os.Chmod(entrypoint, 0755); err != nil {
			return err
		}

		manifest, err := extensions.ExtractManifest(entrypoint)
		if err != nil {
			return fmt.Errorf("error loading extension: %w", err)
		}

		rootCmd, err := NewCmdCustom(filepath.Base(entrypoint), extensions.Extension{
			Manifest:   manifest,
			Entrypoint: entrypoint,
		}, extensions.Config{
			Origin: entrypoint,
		})
		if err != nil {
			return fmt.Errorf("error loading extension: %w", err)
		}

		rootCmd.Use = "extension"
		rootCmd.SetArgs(args[1:])
		return rootCmd.Execute()
	}

	return rootCmd, nil
}

func buildDoc(command *cobra.Command) (string, error) {
	var page strings.Builder
	err := doc.GenMarkdown(command, &page)
	if err != nil {
		return "", err
	}

	out := strings.Builder{}
	for _, line := range strings.Split(page.String(), "\n") {
		if strings.Contains(line, "SEE ALSO") {
			break
		}

		out.WriteString(line + "\n")
	}

	for _, child := range command.Commands() {
		if child.GroupID == CommandGroupExtension {
			continue
		}

		if child.Hidden {
			continue
		}

		childPage, err := buildDoc(child)
		if err != nil {
			return "", err
		}
		out.WriteString(childPage)
	}

	return out.String(), nil
}

func extractListItems(cfg config.Config, extensionMap map[string]extensions.Extension) []types.ListItem {
	var items []types.ListItem
	for _, oneliner := range cfg.Oneliners {
		item := types.ListItem{
			Id:          fmt.Sprintf("oneliner - %s", oneliner.Command),
			Title:       oneliner.Title,
			Accessories: []string{"Oneliner"},
			Actions: []types.Action{
				{
					Title:   "Run",
					Type:    types.ActionTypeExec,
					Command: oneliner.Command,
					Dir:     oneliner.Dir,
					Exit:    oneliner.Exit,
				},
				{
					Title: "Copy Command",
					Key:   "c",
					Type:  types.ActionTypeCopy,
					Text:  oneliner.Command,
					Exit:  true,
				},
			},
		}

		items = append(items, item)
	}

	for alias, extensionConfig := range cfg.Extensions {
		extension, ok := extensionMap[alias]
		if !ok {
			continue
		}

		var rootItems []types.RootItem
		rootItems = append(rootItems, extension.Root()...)
		rootItems = append(rootItems, extensionConfig.Items...)

		for _, rootItem := range rootItems {
			items = append(items, types.ListItem{
				Id:          fmt.Sprintf("%s - %s", alias, rootItem.Title),
				Title:       rootItem.Title,
				Subtitle:    extension.Manifest.Title,
				Accessories: []string{"Command"},
				Actions: []types.Action{
					{
						Title:     "Run",
						Type:      types.ActionTypeRun,
						Extension: alias,
						Command:   rootItem.Command,
						Params:    rootItem.Params,
						Exit:      true,
					},
				},
			})
		}
	}

	return items
}