package cli

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/MadAppGang/httplog"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/sunbeam"
	"github.com/spf13/cobra"
)

func NewCmdServe() *cobra.Command {
	var flags struct {
		addr string
	}

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve the extension API",
		RunE: func(cmd *cobra.Command, args []string) error {
			http.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
				exts, err := LoadExtensions(utils.ExtensionsDir(), true)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)

				encoder := json.NewEncoder(w)
				encoder.SetEscapeHTML(false)
				_ = encoder.Encode(exts)
			})

			http.HandleFunc("GET /{extension}", func(w http.ResponseWriter, r *http.Request) {
				entrypoint, err := extensions.FindEntrypoint(utils.ExtensionsDir(), r.PathValue("extension"))
				if err != nil {
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}

				extension, err := extensions.LoadExtension(entrypoint, true)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)

				encoder := json.NewEncoder(w)
				encoder.SetEscapeHTML(false)
				_ = encoder.Encode(extension)
			})

			http.HandleFunc("POST /{extension}/{command}", func(w http.ResponseWriter, r *http.Request) {
				entrypoint, err := extensions.FindEntrypoint(utils.ExtensionsDir(), r.PathValue("extension"))
				if err != nil {
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}

				extension, err := extensions.LoadExtension(entrypoint, true)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				command, ok := extension.GetCommand(r.PathValue("command"))
				if !ok {
					http.Error(w, "command not found", http.StatusNotFound)
					return
				}

				var params sunbeam.Params
				if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				cmd, err := extension.CmdContext(r.Context(), command, params)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				output, err := cmd.Output()
				if err != nil {
					if exitErr, ok := err.(*exec.ExitError); ok {
						http.Error(w, fmt.Sprintf("command failed: %s", string(exitErr.Stderr)), http.StatusInternalServerError)
						return
					}
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(output)
			})

			if strings.HasPrefix(flags.addr, "unix/") {
				socketPath := strings.TrimPrefix(flags.addr, "unix/")
				if _, err := os.Stat(socketPath); err == nil {
					if err := os.Remove(socketPath); err != nil {
						return fmt.Errorf("failed to remove existing socket: %w", err)
					}
				}

				listener, err := net.Listen("unix", socketPath)
				if err != nil {
					return fmt.Errorf("failed to listen on unix socket: %w", err)
				}

				fmt.Fprintln(cmd.OutOrStdout(), "Listening on", socketPath)
				return http.Serve(listener, httplog.Logger(http.DefaultServeMux))
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Listening on", flags.addr)
			return http.ListenAndServe(flags.addr, httplog.Logger(http.DefaultServeMux))
		},
	}

	cmd.Flags().StringVar(&flags.addr, "addr", ":8080", "address to listen on")

	return cmd
}
