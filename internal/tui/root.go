package tui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pomdtr/sunbeam/internal/extensions"
	"github.com/pomdtr/sunbeam/internal/history"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/sunbeam"
)

type RootList struct {
	width, height int
	err           *Detail
	list          *List
	form          *Form

	history   history.History
	generator func() ([]sunbeam.ListItem, error)
}

type ReloadMsg struct{}

func NewRootList(history history.History, generator func() ([]sunbeam.ListItem, error)) *RootList {
	return &RootList{
		history:   history,
		generator: generator,
	}
}

func (c *RootList) Init() tea.Cmd {
	return c.Reload()
}

func (c *RootList) Reload() tea.Cmd {
	rootItems, err := c.generator()
	if err != nil {
		return c.SetError(err)
	}

	c.history.Sort(rootItems)
	if c.list != nil {
		c.list.SetIsLoading(false)
		c.list.SetItems(rootItems...)
		return nil
	} else {
		c.list = NewList(rootItems...)
		c.list.SetEmptyText("No items")
		c.list.SetSize(c.width, c.height)

		return c.list.Init()
	}
}

func (c *RootList) Focus() tea.Cmd {
	return c.list.Focus()
}

func (c *RootList) Blur() tea.Cmd {
	return c.list.SetIsLoading(false)
}

func (c *RootList) SetSize(width, height int) {
	c.width, c.height = width, height
	if c.err != nil {
		c.err.SetSize(width, height)
	}
	if c.form != nil {
		c.form.SetSize(width, height)
	}

	if c.list != nil {
		c.list.SetSize(width, height)
	}
}

func (c *RootList) SetError(err error) tea.Cmd {
	c.err = NewErrorPage(err)
	c.err.SetSize(c.width, c.height)
	return func() tea.Msg {
		return err
	}
}

func (c *RootList) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if c.form != nil {
				c.form = nil
				return c, c.list.Focus()
			}
		case "ctrl+r":
			return c, tea.Batch(c.list.SetIsLoading(true), c.Reload())
		case "ctrl+e":
			return c, func() tea.Msg {
				item, ok := c.list.Selection()
				if !ok {
					return nil
				}

				extensionName := item.Actions[0].Run.Extension
				entrypoint, err := extensions.FindEntrypoint(utils.ExtensionsDir(), extensionName)
				if err != nil {
					return fmt.Errorf("extension %s not found", extensionName)
				}

				return sunbeam.Action{
					Type: sunbeam.ActionTypeEdit,
					Edit: &sunbeam.EditAction{Path: entrypoint},
				}
			}
		}
	case ReloadMsg:
		return c, tea.Batch(c.list.SetIsLoading(true), c.Reload())
	case sunbeam.Action:
		selection, ok := c.list.Selection()
		if !ok {
			return c, nil
		}
		c.history.Update(selection.Id)
		if err := c.history.Save(); err != nil {
			return c, c.SetError(err)
		}

		switch msg.Type {
		case sunbeam.ActionTypeRun:
			entrypoint, err := extensions.FindEntrypoint(utils.ExtensionsDir(), msg.Run.Extension)
			if err != nil {
				return c, c.SetError(err)
			}

			extension, err := extensions.LoadExtension(entrypoint, true)
			if err != nil {
				return c, c.SetError(err)
			}

			command, ok := extension.GetCommand(msg.Run.Command)
			if !ok {
				return c, c.SetError(fmt.Errorf("command %s not found", msg.Run.Command))
			}

			missingParams := FindMissingInputs(command.Params, msg.Run.Params)
			for _, param := range missingParams {
				if param.Optional {
					continue
				}

				c.form = NewForm(func(values map[string]any) tea.Msg {
					params := make(map[string]any)
					for k, v := range msg.Run.Params {
						params[k] = v
					}

					for k, v := range values {
						params[k] = v
					}

					return sunbeam.Action{
						Title: msg.Title,
						Type:  sunbeam.ActionTypeRun,
						Run: &sunbeam.RunAction{
							Extension: msg.Run.Extension,
							Command:   msg.Run.Command,
							Reload:    msg.Run.Reload,
							Params:    params,
						},
					}
				}, missingParams...)

				c.form.SetSize(c.width, c.height)
				return c, c.form.Init()
			}
			c.form = nil

			params := make(map[string]any)

			for k, v := range msg.Run.Params {
				params[k] = v
			}

			switch command.Mode {
			case sunbeam.CommandModeSearch, sunbeam.CommandModeFilter, sunbeam.CommandModeDetail:
				runner := NewRunner(extension, command, params)
				return c, PushPageCmd(runner)
			case sunbeam.CommandModeSilent:
				return c, func() tea.Msg {
					output, err := extension.Output(context.Background(), command, params)
					if err != nil {
						return PushPageMsg{NewErrorPage(err)}
					}

					if len(output) > 0 {
						output = bytes.Trim(output, "\n")
						rows := strings.Split(string(output), "\n")
						return ShowNotificationMsg{rows[len(rows)-1]}
					}

					if msg.Exit {
						return ExitMsg{}
					}

					return nil
				}
			case sunbeam.CommandModeAction:
				return c, func() tea.Msg {
					output, err := extension.Output(context.Background(), command, params)
					if err != nil {
						return PushPageMsg{NewErrorPage(err)}
					}

					var action sunbeam.Action
					if err := json.Unmarshal(output, &action); err != nil {
						return PushPageMsg{NewErrorPage(err)}
					}

					return action
				}
			}
		case sunbeam.ActionTypeCopy:
			return c, func() tea.Msg {
				if err := clipboard.WriteAll(msg.Copy.Text); err != nil {
					return err
				}

				if msg.Exit {
					return ExitMsg{}
				}

				return ShowNotificationMsg{"Copied!"}
			}
		case sunbeam.ActionTypeEdit:
			editCmd := exec.Command(utils.FindEditor(), msg.Edit.Path)
			return c, tea.ExecProcess(editCmd, func(err error) tea.Msg {
				if err != nil {
					return err
				}

				if msg.Edit.Reload {
					return c.Reload()
				}

				if msg.Exit {
					return ExitMsg{}
				}

				return nil
			})
		case sunbeam.ActionTypeOpen:
			return c, func() tea.Msg {
				if err := utils.Open(msg.Open.Url); err != nil {
					return err
				}

				if msg.Exit {
					return ExitMsg{}
				}

				return nil
			}
		case sunbeam.ActionTypeReload:
			return c, tea.Sequence(c.err.SetIsLoading(true), c.Reload())
		default:
			return c, nil
		}
	case error:
		c.err = NewErrorPage(msg)
		c.err.SetSize(c.width, c.height)
		return c, c.err.Init()

	}

	if c.err != nil {
		page, cmd := c.err.Update(msg)
		c.err = page.(*Detail)
		return c, cmd
	}

	if c.form != nil {
		page, cmd := c.form.Update(msg)
		c.form = page.(*Form)
		return c, cmd
	}

	if c.list != nil {
		page, cmd := c.list.Update(msg)
		c.list = page.(*List)
		return c, cmd
	}

	return c, nil
}

func (c *RootList) View() string {
	if c.err != nil {
		return c.err.View()
	}
	if c.form != nil {
		return c.form.View()
	}
	if c.list != nil {
		return c.list.View()
	}

	return ""
}

type History struct {
	entries map[string]int64
	path    string
}

func (h History) Sort(items []sunbeam.ListItem) {
	sort.SliceStable(items, func(i, j int) bool {
		keyI := items[i].Id
		keyJ := items[j].Id

		return h.entries[keyI] > h.entries[keyJ]
	})
}

func LoadHistory(fp string) (History, error) {
	f, err := os.Open(fp)
	if err != nil {
		return History{}, err
	}

	var entries map[string]int64
	if err := json.NewDecoder(f).Decode(&entries); err != nil {
		return History{}, err
	}

	return History{
		entries: entries,
		path:    fp,
	}, nil
}

func (h History) Save() error {
	f, err := os.OpenFile(h.path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(h.path), 0755); err != nil {
			return err
		}

		f, err = os.Create(h.path)
		if err != nil {
			return err
		}
	}

	encoder := json.NewEncoder(f)
	if err := encoder.Encode(h.entries); err != nil {
		return err
	}

	return nil
}
