package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func PopPageCmd() tea.Msg {
	return PopPageMsg{}
}

type PopPageMsg struct{}

func PushPageCmd(page Page) tea.Cmd {
	return func() tea.Msg {
		return PushPageMsg{
			Page: page,
		}
	}
}

type PushPageMsg struct {
	Page Page
}

type Page interface {
	Init() tea.Cmd
	Update(tea.Msg) (Page, tea.Cmd)
	View() string

	Focus() tea.Cmd
	Blur() tea.Cmd

	SetSize(width, height int)
}

type ExitMsg struct {
}

func ExitCmd() tea.Msg {
	return ExitMsg{}
}

type Paginator struct {
	width, height int
	pages         []Page
}

func NewPaginator(root Page) *Paginator {
	return &Paginator{
		pages: []Page{root},
	}
}

func (m *Paginator) Init() tea.Cmd {
	if len(m.pages) == 0 {
		return nil
	}

	return m.pages[0].Init()
}

func (m *Paginator) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		if msg.Height%2 == 0 {
			m.SetSize(msg.Width, msg.Height-1)
		} else {
			m.SetSize(msg.Width, msg.Height)
		}
		return m, nil
	case PushPageMsg:
		cmd := m.Push(msg.Page)
		return m, cmd
	case PopPageMsg:
		if len(m.pages) > 1 {
			return m, m.Pop()
		}

		return m, tea.Quit
	case ExitMsg:
		return m, tea.Quit
	}

	// Update the current page
	var cmd tea.Cmd

	if len(m.pages) > 0 {
		currentPageIdx := len(m.pages) - 1
		m.pages[currentPageIdx], cmd = m.pages[currentPageIdx].Update(msg)
	} else {
		return m, nil
	}

	return m, cmd
}

func (m *Paginator) View() string {
	if len(m.pages) > 0 {
		currentPage := m.pages[len(m.pages)-1]
		return currentPage.View()
	}

	return ""
}

func (m *Paginator) SetSize(width, height int) {
	m.width = width
	m.height = height

	for _, page := range m.pages {
		page.SetSize(m.width, m.height)
	}
}

func (m *Paginator) Push(page Page) tea.Cmd {
	var cmd tea.Cmd
	if len(m.pages) > 0 {
		cmd = m.pages[len(m.pages)-1].Blur()
	}
	page.SetSize(m.width, m.height)
	m.pages = append(m.pages, page)
	return tea.Sequence(cmd, page.Init())
}

func (m *Paginator) Pop() tea.Cmd {
	var cmds []tea.Cmd
	if len(m.pages) > 0 {
		cmds = append(cmds, m.pages[len(m.pages)-1].Blur())
		m.pages = m.pages[:len(m.pages)-1]
	}

	if len(m.pages) > 0 {
		cmds = append(cmds, m.pages[len(m.pages)-1].Focus())
	}

	return tea.Sequence(cmds...)
}

func Draw(page Page) error {
	paginator := NewPaginator(page)
	p := tea.NewProgram(paginator, tea.WithAltScreen())

	_, err := p.Run()
	return err
}
