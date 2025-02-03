package cli

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/cobra"
)

func NewCmdFetch() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch a remote extension",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				resp, err := http.Get(args[0])
				if err != nil {
					return fmt.Errorf("failed to fetch extension: %w", err)
				}
				defer resp.Body.Close()

				_, err = io.Copy(cmd.OutOrStdout(), resp.Body)
				if err != nil {
					return fmt.Errorf("failed to copy response body: %w", err)
				}

				return nil
			}

			target, err := url.JoinPath(args[0], args[1])
			if err != nil {
				return fmt.Errorf("failed to join URL path: %w", err)
			}

			resp, err := http.Post(target, "application/json", os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to fetch extension: %w", err)
			}
			defer resp.Body.Close()

			_, err = io.Copy(cmd.OutOrStdout(), resp.Body)
			if err != nil {
				return fmt.Errorf("failed to copy response body: %w", err)
			}

			return nil
		},
	}

	return cmd
}
