package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/MadAppGang/httplog"
	"github.com/google/uuid"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/sunbeam"
	"github.com/spf13/cobra"
)

func NewCmdServe() *cobra.Command {
	var flags struct {
		host                   string
		port                   int
		connectionToken        string
		connectionTokenFile    string
		withoutConnectionToken bool
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

				if len(output) == 0 {
					w.WriteHeader(http.StatusNoContent)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(output)
			})

			var connectionToken string
			if flags.connectionToken != "" {
				connectionToken = flags.connectionToken
			} else if flags.connectionTokenFile != "" {
				b, err := os.ReadFile(flags.connectionTokenFile)
				if err != nil {
					return fmt.Errorf("failed to read connection token file: %w", err)
				}

				connectionToken = strings.TrimSpace(string(b))
			} else if !flags.withoutConnectionToken {
				token, err := uuid.NewRandom()
				if err != nil {
					return fmt.Errorf("failed to generate connection token: %w", err)
				}

				connectionToken = token.String()
			}

			var handler http.Handler = http.DefaultServeMux
			handler = cors(handler)

			addr := fmt.Sprintf("%s:%d", flags.host, flags.port)
			if connectionToken != "" {
				handler = withConnectionToken(connectionToken, handler)
				fmt.Fprintf(cmd.ErrOrStderr(), "Server available at http://%s?tkn=%s\n", addr, connectionToken)
			} else {
				fmt.Fprintf(cmd.ErrOrStderr(), "Server available at http://%s\n", addr)
			}

			handler = httplog.Logger(handler)
			return http.ListenAndServe(addr, handler)
		},
	}

	cmd.Flags().StringVar(&flags.host, "host", "127.0.0.1", "host to listen on")
	cmd.Flags().IntVar(&flags.port, "port", 8080, "port to listen on")
	cmd.Flags().StringVar(&flags.connectionToken, "connection-token", "", "connection token")
	cmd.Flags().StringVar(&flags.connectionTokenFile, "connection-token-file", "", "file containing the connection token")
	cmd.Flags().BoolVar(&flags.withoutConnectionToken, "without-connection-token", false, "disable connection token")

	cmd.MarkFlagsMutuallyExclusive("connection-token", "connection-token-file", "without-connection-token")

	return cmd
}

func withConnectionToken(token string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if tokenParam := r.URL.Query().Get("tkn"); tokenParam == token {
			next.ServeHTTP(w, r)
			return
		}

		if auth := r.Header.Get("Authorization"); auth == fmt.Sprintf("Bearer %s", token) {
			next.ServeHTTP(w, r)
			return
		}

		http.Error(w, "unauthorized", http.StatusUnauthorized)
	})
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
