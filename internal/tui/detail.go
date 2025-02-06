package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"
	"github.com/pomdtr/sunbeam/internal/utils"
	"github.com/pomdtr/sunbeam/pkg/sunbeam"
)

type Detail struct {
	actionsFocused bool

	isLoading bool
	spinner   spinner.Model
	viewport  viewport.Model
	statusBar StatusBar
	input     textinput.Model

	text          string
	width, height int

	Style    lipgloss.Style
	Markdown bool
}

func AnsiStyle() ansi.StyleConfig {
	style := styles.PinkStyleConfig
	style.Document.BlockPrefix = ""

	return style
}

func NewDetail(text string, actions ...sunbeam.Action) *Detail {
	viewport := viewport.New(0, 0)
	viewport.Style = lipgloss.NewStyle()

	statusBar := NewStatusBar(actions...)
	items := make([]FilterItem, 0)
	for _, action := range actions {
		items = append(items, ListItem{
			Title:    action.Title,
			Subtitle: action.Key,
			Actions:  []sunbeam.Action{action},
		})
	}

	input := textinput.New()
	input.PlaceholderStyle = lipgloss.NewStyle().Faint(true)
	input.Prompt = ""
	input.Placeholder = "Search Actions..."

	d := Detail{
		spinner:   spinner.New(),
		input:     input,
		viewport:  viewport,
		statusBar: statusBar,
		text:      text,
	}

	_ = d.RefreshContent()
	return &d
}

func (d *Detail) Init() tea.Cmd {
	return nil
}

func (d *Detail) Focus() tea.Cmd {
	d.input.SetValue("")
	d.input.Blur()
	d.statusBar.Reset()
	return nil
}

func (d *Detail) Blur() tea.Cmd {
	return nil
}

type DetailMsg string

func (d *Detail) SetIsLoading(isLoading bool) tea.Cmd {
	d.isLoading = isLoading
	if isLoading {
		return d.spinner.Tick
	}

	return nil
}

func (c *Detail) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			if c.statusBar.expanded {
				break
			}

			if len(c.statusBar.actions) == 0 {
				break
			}

			c.statusBar.expanded = true
			return c, c.input.Focus()
		case "q":
			if c.actionsFocused {
				break
			}

			return c, PopPageCmd
		case "esc":
			if c.statusBar.expanded {
				c.statusBar.Reset()
				c.input.Blur()
				return c, nil
			}
			return c, func() tea.Msg {
				return PopPageMsg{}
			}
		}
	}
	var cmds []tea.Cmd
	var cmd tea.Cmd

	if c.isLoading {
		c.spinner, cmd = c.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	if c.input.Focused() {
		c.input, cmd = c.input.Update(msg)
		c.statusBar.FilterActions(c.input.Value())
		cmds = append(cmds, cmd)
	} else {
		c.viewport, cmd = c.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	c.statusBar, cmd = c.statusBar.Update(msg)
	cmds = append(cmds, cmd)

	return c, tea.Batch(cmds...)
}

func (c *Detail) RefreshContent() error {
	var content string
	if c.Markdown {
		render, err := glamour.NewTermRenderer(
			glamour.WithStyles(AnsiStyle()),
			glamour.WithWordWrap(c.width),
		)
		if err != nil {
			return err
		}

		content, err = render.Render(c.text)
		if err != nil {
			return err
		}
	} else {
		content = wrap.String(wordwrap.String(utils.StripAnsi(c.text), c.width-4), c.width-4)
		content = lipgloss.NewStyle().Padding(0, 2).Render(content)
	}

	c.viewport.SetContent(content)
	return nil
}

func (c *Detail) SetSize(width, height int) {
	c.width, c.height = width, height

	c.viewport.Height = height - 4
	c.viewport.Width = width

	c.statusBar.Width = width
	_ = c.RefreshContent()
}

func (c *Detail) View() string {
	var headerRow string
	if c.isLoading {
		headerRow = fmt.Sprintf(" %s", c.spinner.View())
	} else {
		headerRow = "  "
	}

	if c.input.Focused() {
		headerRow = fmt.Sprintf("%s %s", headerRow, c.input.View())
	}

	return lipgloss.JoinVertical(lipgloss.Left, headerRow, separator(c.width), c.viewport.View(), c.statusBar.View())
}
