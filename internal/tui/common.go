package tui

import (
	"errors"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	ErrValidation = errors.New("validation error")
)

type errMsg struct {
	err  string
	code int
}

func (e errMsg) String() string {
	return e.err
}

type doneMsg struct{}

type requireAuthMsg struct{}

func requireAuthCmd() tea.Msg {
	return requireAuthMsg{}
}

type spinMsg struct{}

func spinnerSpinCmd() tea.Msg { return spinMsg{} }

type resetSpinnerMsg struct{}

func spinnerResetCmd() tea.Msg { return resetSpinnerMsg{} }

type selDiscUserMsg struct { // selected Discovered User Msg
	id, name, email string
}
