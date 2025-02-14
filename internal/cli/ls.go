package cli

import (
	"fmt"
	"os"

	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/spf13/cobra"
)

func NewCmdLs() *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List all extensions",
		Run: func(cmd *cobra.Command, args []string) {
			exts, err := LoadExtensions(utils.ExtensionsDir(), false)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %s\n", err)
				os.Exit(1)
			}

			for _, ext := range exts {
				cmd.Println(ext.Name)
			}
		},
	}
}
