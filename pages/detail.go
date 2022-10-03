package pages

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/bubbles"
	commands "github.com/pomdtr/sunbeam/commands"
)

var infoStyle = func() lipgloss.Style {
	b := lipgloss.RoundedBorder()
	b.Left = "┤"
	return titleStyle.Copy().BorderStyle(b)
}()

type ActionRunner func(commands.ScriptAction) tea.Cmd

type DetailContainer struct {
	response  commands.DetailResponse
	runAction ActionRunner
	width     int
	height    int
	viewport  *viewport.Model
}

func NewDetailContainer(response *commands.DetailResponse, runAction ActionRunner) DetailContainer {
	viewport := viewport.New(0, 0)
	var content string
	if lipgloss.HasDarkBackground() {
		content, _ = glamour.Render(response.Text, "dark")
	} else {
		content, _ = glamour.Render(response.Text, "light")
	}
	viewport.SetContent(content)

	return DetailContainer{
		response:  *response,
		runAction: runAction,
		viewport:  &viewport,
	}
}

func (c DetailContainer) SetSize(width, height int) {
	c.viewport.Width = width
	c.viewport.Height = height - lipgloss.Height(c.headerView()) - lipgloss.Height(c.footerView())
}

func (c DetailContainer) Init() tea.Cmd {
	return nil
}

func (c DetailContainer) headerView() string {
	return bubbles.SunbeamHeader(c.viewport.Width)
}

func (c DetailContainer) footerView() string {
	return bubbles.SunbeamFooter(c.viewport.Width, c.response.Title)
}

func (c DetailContainer) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes:
			for _, action := range c.response.Actions {
				if action.Keybind == string(msg.Runes) {
					return c, c.runAction(action)
				}
			}
		case tea.KeyEscape:
			return c, PopCmd
		}
	}
	var cmd tea.Cmd
	model, cmd := c.viewport.Update(msg)
	c.viewport = &model
	return c, cmd
}

func (c DetailContainer) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, c.headerView(), c.viewport.View(), c.footerView())
}
