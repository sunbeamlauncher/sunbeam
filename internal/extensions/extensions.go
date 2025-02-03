package extensions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/acarl005/stripansi"
	"github.com/pomdtr/sunbeam/internal/schemas"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/sunbeam"
)

type Extension struct {
	Name string `json:"name"`
	sunbeam.Manifest
	Entrypoint string `json:"-"`
}

type Preferences map[string]any

type Metadata struct {
	Type       ExtensionType `json:"type"`
	Origin     string        `json:"origin"`
	Entrypoint string        `json:"entrypoint"`
}

type ExtensionType string

const (
	ExtensionTypeLocal ExtensionType = "local"
	ExtensionTypeHttp  ExtensionType = "http"
)

func (e Extension) GetCommand(name string) (sunbeam.Command, bool) {
	for _, command := range e.Manifest.Commands {
		if command.Name == name {
			return command, true
		}
	}
	return sunbeam.Command{}, false
}

func (e Extension) RootItems() []sunbeam.ListItem {
	var items []sunbeam.ListItem

	for i, action := range e.Manifest.Root {
		title := action.Title
		action.Title = "Run"

		if action.Type == sunbeam.ActionTypeRun {
			action.Run.Extension = e.Name
		}

		items = append(items, sunbeam.ListItem{
			Title:    title,
			Subtitle: e.Manifest.Title,
			Id:       fmt.Sprintf("%s-%d", e.Name, i),
			Accessories: []string{
				e.Name,
			},
			Actions: []sunbeam.Action{action},
		})
	}

	return items
}

func (ext Extension) Output(ctx context.Context, command sunbeam.Command, params sunbeam.Params) ([]byte, error) {
	cmd, err := ext.CmdContext(context.Background(), command, params)
	if err != nil {
		return nil, err
	}

	var exitErr *exec.ExitError
	if output, err := cmd.Output(); err == nil {
		return output, nil
	} else if errors.As(err, &exitErr) {
		return nil, fmt.Errorf("command failed: %s", stripansi.Strip(string(exitErr.Stderr)))
	} else {
		return nil, err
	}
}

func (e Extension) CmdContext(ctx context.Context, command sunbeam.Command, params sunbeam.Params) (*exec.Cmd, error) {
	if params == nil {
		params = make(map[string]any)
	}

	for _, spec := range command.Params {
		if _, ok := params[spec.Name]; ok {
			continue
		}

		if !spec.Optional {
			return nil, fmt.Errorf("missing required parameter %s", spec.Name)
		}

		params[spec.Name] = spec.Default
	}

	cmd := exec.CommandContext(ctx, e.Entrypoint, command.Name)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	defer stdin.Close()

	encoder := json.NewEncoder(stdin)
	if err := encoder.Encode(params); err != nil {
		return nil, err
	}

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "SUNBEAM=1")
	return cmd, nil
}

func FindEntrypoint(extensionDir string, name string) (string, error) {
	entries, err := os.ReadDir(extensionDir)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if name == strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())) {
			return filepath.Join(extensionDir, entry.Name()), nil
		}

	}

	return "", fmt.Errorf("entrypoint not found")
}

func LoadExtension(entrypoint string, reload bool) (Extension, error) {
	name := strings.TrimSuffix(filepath.Base(entrypoint), filepath.Ext(entrypoint))

	entrypointInfo, err := os.Stat(entrypoint)
	if err != nil {
		return Extension{}, err
	}

	// if the entrypoint is a symlink, resolve it
	if entrypointInfo.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(entrypoint)
		if err != nil {
			return Extension{}, err
		}

		info, err := os.Stat(target)
		if err != nil {
			return Extension{}, err
		}

		entrypointInfo = info
	}

	manifestPath := filepath.Join(utils.CacheDir(), "extensions", name+".json")
	if manifestInfo, err := os.Stat(manifestPath); err != nil || entrypointInfo.ModTime().After(manifestInfo.ModTime()) || reload {
		if os.Getenv("SUNBEAM") == "1" {
			return Extension{}, errors.New("nested extension detected")
		}

		manifest, err := cacheManifest(entrypoint, manifestPath)
		if err != nil {
			return Extension{}, err
		}

		return Extension{
			Name:       name,
			Manifest:   manifest,
			Entrypoint: entrypoint,
		}, nil
	}

	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return Extension{}, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest sunbeam.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return Extension{}, fmt.Errorf("failed to decode manifest: %w", err)
	}

	return Extension{
		Name:       name,
		Manifest:   manifest,
		Entrypoint: entrypoint,
	}, nil
}

func cacheManifest(entrypoint string, manifestPath string) (sunbeam.Manifest, error) {
	manifest, err := ExtractManifest(entrypoint)
	if err != nil {
		return sunbeam.Manifest{}, fmt.Errorf("failed to extract manifest: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
		return sunbeam.Manifest{}, fmt.Errorf("failed to create directory: %w", err)
	}

	f, err := os.Create(manifestPath)
	if err != nil {
		return sunbeam.Manifest{}, fmt.Errorf("failed to create manifest: %w", err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(manifest); err != nil {
		return sunbeam.Manifest{}, fmt.Errorf("failed to write manifest: %w", err)
	}

	return manifest, nil
}

func ExtractManifest(entrypoint string) (sunbeam.Manifest, error) {
	entrypoint, err := filepath.Abs(entrypoint)
	if err != nil {
		return sunbeam.Manifest{}, err
	}

	if err := os.Chmod(entrypoint, 0755); err != nil {
		return sunbeam.Manifest{}, err
	}

	cmd := exec.Command(entrypoint)
	cmd.Dir = filepath.Dir(entrypoint)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "SUNBEAM=1")

	manifestBytes, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return sunbeam.Manifest{}, fmt.Errorf("command failed: %s", stripansi.Strip(string(exitErr.Stderr)))
		}

		return sunbeam.Manifest{}, err
	}

	if err := schemas.ValidateManifest(manifestBytes); err != nil {
		return sunbeam.Manifest{}, err
	}

	var manifest sunbeam.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return sunbeam.Manifest{}, err
	}

	return manifest, nil
}
