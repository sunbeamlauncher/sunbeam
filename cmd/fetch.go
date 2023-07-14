package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

func NewFetchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fetch <url>",
		GroupID: coreGroupID,
		Args:    cobra.ExactArgs(1),
		Short:   "Fetch http using a curl-like syntax",
		RunE: func(cmd *cobra.Command, args []string) error {
			headersFlag, _ := cmd.Flags().GetStringArray("header")

			var input []byte
			if !isatty.IsTerminal(os.Stdin.Fd()) {
				b, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				input = b
			}

			var method string
			if cmd.Flags().Changed("method") {
				method, _ = cmd.Flags().GetString("method")
			} else if len(input) > 0 {
				method = http.MethodPost
			} else {
				method = http.MethodGet
			}

			if method == "GET" && len(input) > 0 {
				return fmt.Errorf("cannot specify request body for GET request")
			}

			req, err := http.NewRequest(method, args[0], bytes.NewReader(input))
			if err != nil {
				return err
			}

			for _, header := range headersFlag {
				parts := strings.SplitN(header, ":", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid header: %s", header)
				}

				req.Header.Set(parts[0], parts[1])
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().String("method", "", "http method")
	cmd.Flags().StringArrayP("header", "H", []string{}, "http header")

	return cmd
}
