package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/internal/fzf"
	"github.com/pomdtr/sunbeam/pkg/sunbeam"
)

func separator(n int) string {
	separator := strings.Repeat("─", n)
	return lipgloss.NewStyle().Bold(true).Render(separator)
}

type StatusBar struct {
	Width int

	notification string

	cursor   int
	actions  []sunbeam.Action
	filtered []sunbeam.Action
	expanded bool
}

type ShowNotificationMsg struct {
	Title string
}

type HideNotificationMsg struct{}

func NewStatusBar(actions ...sunbeam.Action) StatusBar {
	return StatusBar{
		actions:  actions,
		filtered: actions,
	}
}

func (c *StatusBar) SetActionsNoSelection(actions ...sunbeam.Action) {
	// this is called all the time to set the action for no-selection
	// we want to keep the current expanded and cursor status rather than updating them
	c.actions = actions
	c.filtered = actions
}

func (c *StatusBar) SetActions(actions ...sunbeam.Action) {
	c.expanded = false
	c.cursor = 0
	c.actions = actions
	c.filtered = actions
}

func (c *StatusBar) FilterActions(query string) {
	if query == "" {
		c.filtered = c.actions
		return
	}

	c.filtered = make([]sunbeam.Action, 0)
	for i := 0; i < len(c.actions); i++ {
		if fzf.Score(c.actions[i].Title, query) > 0 {
			c.filtered = append(c.filtered, c.actions[i])
		}
	}

	sort.SliceStable(c.filtered, func(i, j int) bool {
		return fzf.Score(c.filtered[i].Title, query) > fzf.Score(c.filtered[j].Title, query)
	})

	c.cursor = 0
}

func (p StatusBar) Update(msg tea.Msg) (StatusBar, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "right":
			if !p.expanded {
				break
			}

			if msg.String() == "tab" && len(p.filtered) == 0 {
				return p, nil
			}

			if p.cursor < len(p.filtered)-1 {
				p.cursor++
			} else {
				p.cursor = 0
			}

			return p, nil
		case "shift+tab", "left":
			if !p.expanded {
				break
			}

			if p.cursor > 0 {
				p.cursor--
			} else {
				p.cursor = len(p.filtered) - 1
			}
		case "ctrl+d":
			if p.expanded {
				break
			}

			return p, PopPageCmd
		}
	case ShowNotificationMsg:
		p.Reset()
		if msg.Title == "" {
			return p, nil
		}

		p.notification = msg.Title
		return p, tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
			return HideNotificationMsg{}
		})
	case HideNotificationMsg:
		p.notification = ""
		return p, nil
	}

	return p, nil
}

func (c *StatusBar) Reset() {
	c.expanded = false
	c.cursor = 0
	c.filtered = c.actions
}

func ActionTitle(action sunbeam.Action) string {
	if action.Title != "" {
		return action.Title
	}

	switch action.Type {
	case sunbeam.ActionTypeRun:
		return "Run Command"
	case sunbeam.ActionTypeCopy:
		return "Copy to Clipboard"
	case sunbeam.ActionTypeOpen:
		return "Open URL"
	default:
		return string(action.Type)
	}
}

func (c StatusBar) View() string {
	var accessory string
	if len(c.actions) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, separator(c.Width), strings.Repeat(" ", c.Width))
	}
	if c.expanded {
		accessories := make([]string, len(c.filtered))
		for i, action := range c.filtered {
			accessories[i] = renderAction(ActionTitle(action), "", i == c.cursor)
		}

		availableWidth := c.Width
		availableWidth -= 3 * (len(c.filtered) - 2) // 3 spaces between each action
		availableWidth -= 2                         // 2 spaces for the margins
		startIdx := 0
		endIdx := len(accessories)

		for lipgloss.Width(strings.Join(accessories[startIdx:endIdx], " · ")) > availableWidth {
			if endIdx-1 > c.cursor {
				endIdx--
			} else if startIdx < c.cursor {
				startIdx++
			} else {
				break
			}
		}

		accessory = strings.Join(accessories[startIdx:endIdx], " · ")
		if startIdx > 0 {
			accessory = fmt.Sprintf("… · %s", accessory)
		}
		if endIdx < len(accessories) {
			accessory = fmt.Sprintf("%s · …", accessory)
		}

	} else {
		if len(c.filtered) > 1 {
			accessory = fmt.Sprintf("%s · Actions %s", renderAction(ActionTitle(c.filtered[0]), "enter", false), lipgloss.NewStyle().Faint(true).Render("tab"))
		} else {
			accessory = renderAction(ActionTitle(c.filtered[0]), "enter", false)
		}
	}

	var statusbar string
	if c.expanded {
		statusbar = fmt.Sprintf("   %s ", accessory)
	} else {

		blanks := strings.Repeat(" ", max(c.Width-lipgloss.Width(accessory)-lipgloss.Width(c.notification)-4, 0))
		statusbar = fmt.Sprintf("   %s%s%s ", lipgloss.NewStyle().Faint(true).Render(c.notification), blanks, accessory)
	}

	return lipgloss.JoinVertical(lipgloss.Left, separator(c.Width), statusbar)
}

func renderAction(title string, subtitle string, selected bool) string {
	var view string
	if subtitle != "" {
		view = fmt.Sprintf("%s %s", title, lipgloss.NewStyle().Faint(true).Render(subtitle))
	} else {
		view = title
	}

	if selected {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true).Render(view)
	}

	return view
}
