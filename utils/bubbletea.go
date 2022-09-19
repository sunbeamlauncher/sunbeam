package utils

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

var NewErrorCmd = func(format string, values any) func() tea.Msg {
	return SendMsg(fmt.Errorf(format, values))
}

func SendMsg[T any](msg T) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}