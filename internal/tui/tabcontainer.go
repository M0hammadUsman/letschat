package tui

import (
	"github.com/M0hammadUsman/letschat/internal/client"
	"github.com/M0hammadUsman/letschat/internal/domain"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/stopwatch"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	zone "github.com/lrstanley/bubblezone"
	"net/http"
	"strconv"
	"time"
)

// once this is != "" TabContainerModel.spinner will searching
var ioStatus string

type TabContainerModel struct {
	discover  DiscoverModel
	tabs      []string
	activeTab int
	errMsg    *errMsg
	user      *domain.User
	timer     timer.Model
	stopwatch stopwatch.Model
	spinner   *spinner.Model
	client    *client.Client
}

func InitialTabContainerModel() TabContainerModel {
	t := []string{
		"🔎 Discover",
		"💭 Letschat",
		"⚙️ Settings",
	}
	c := client.Get()

	s := spinner.New(spinner.WithStyle(spinnerStyle))
	s.Spinner = spinner.Points

	return TabContainerModel{
		discover:  InitialDiscoverModel(c, &s),
		tabs:      t,
		activeTab: 1,
		timer:     timer.New(0),
		stopwatch: stopwatch.New(),
		spinner:   &s,

		client: c,
	}

}

func (m TabContainerModel) Init() tea.Cmd {
	return tea.Batch(m.discover.Init(), m.getCurrentActiveUser(), m.stopwatch.Init())
}

func (m TabContainerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		terminalHeight = msg.Height
		terminalWidth = msg.Width

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "shift+tab":
			if m.activeTab == len(m.tabs)-1 {
				m.activeTab = 0
			} else {
				m.activeTab++
			}
		case "enter":
			if !m.timer.Timedout() {
				return m, nil
			}
		}

	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonLeft:
			for i, t := range m.tabs {
				if zone.Get(t).InBounds(msg) {
					m.activeTab = i
				}
			}
		default:
		}

	case requireAuthMsg:
		loginModel := InitialLoginModel()
		return loginModel, loginModel.Init()

	case *errMsg:
		m.resetSpinner()
		m.errMsg = msg
		if m.timer.Timedout() {
			m.timer = timer.New(5 * time.Second)
			return m, m.timer.Init()
		}

	case timer.TickMsg:
		return m, m.handleTimerUpdate(msg)

	case timer.TimeoutMsg:
		m.errMsg = nil

	case spinMsg:
		return m, m.spinner.Tick

	case spinner.TickMsg:
		return m, m.handleSpinnerUpdate(msg)

	case resetSpinnerMsg:
		m.resetSpinner()

	case *domain.User:
		m.user = msg
	}

	return m, tea.Batch(m.handleDiscoverModelUpdate(msg), m.handleStopwatchUpdate(msg))
}

func (m TabContainerModel) View() string {

	if m.errMsg != nil {
		return renderErrContainer(m.errMsg.err, m.errMsg.code, m.timer.View())
	}

	tabs := make([]string, len(m.tabs))

	for i, t := range m.tabs {
		if i == m.activeTab {
			t = zone.Mark(t, activeTab.Render(t))
			tabs = append(tabs, t)
		} else {
			t = zone.Mark(t, tab.Render(t))
			tabs = append(tabs, t)
		}
	}

	t := lipgloss.JoinHorizontal(
		lipgloss.Center,
		tabs...,
	)
	s := "Session Uptime: " + m.stopwatch.View()
	if ioStatus != "" {
		s = ioStatus + " " + m.spinner.View()
	}
	if m.user != nil {
		t = renderTabsWithGapsAndText(t, m.user.Name, s)
	} else {
		t = renderTabsWithGapsAndText(t, "", s)
	}
	content := m.populateActiveTabContent()
	c := renderContainerWithTabs(t, content)

	return zone.Scan(c)
}

// Helpers & Stuff -----------------------------------------------------------------------------------------------------

func renderTabsWithGapsAndText(tabs, textL, textR string) string {
	w := (terminalWidth - lipgloss.Width(tabs) - 4) / 2
	gapL := tabGapLeft.Width(w).Render(statusText.Render("Letschat"))
	gapR := tabGapRight.Width(w).Render(statusText.Render(textR))
	if textL != "" {
		gapL = tabGapLeft.Width(w).Render(statusText.Render(textL))
	}
	return lipgloss.JoinHorizontal(lipgloss.Bottom, gapL, tabs, gapR)
}

func renderContainerWithTabs(tabs string, content string) string {
	w := lipgloss.Width(tabs) - 2
	h := terminalHeight - lipgloss.Height(tabs) - 1
	c := tabContainer.Width(max(0, w)).Height(max(0, h)).Render(content)
	return lipgloss.JoinVertical(lipgloss.Right, tabs, c)
}

func renderErrContainer(err string, code int, timer string) string {
	h := errHeaderStyle.Render(strconv.Itoa(code), "-", http.StatusText(code))
	margin := errContainerStyle.GetWidth() - (lipgloss.Width(h) + 6)
	t := lipgloss.NewStyle().Foreground(dangerColor).MarginLeft(margin).Render(timer)
	h = lipgloss.JoinHorizontal(lipgloss.Left, h, t)
	d := errDescStyle.Render(ansi.Wordwrap(err, 58, " ")) // 58 -> sweet spot
	e := lipgloss.JoinVertical(lipgloss.Left, h, d)
	e = errContainerStyle.Render(e)
	return lipgloss.Place(terminalWidth, terminalHeight,
		lipgloss.Center, lipgloss.Center,
		e,
		lipgloss.WithWhitespaceChars("↯"),
		lipgloss.WithWhitespaceForeground(darkGreyColor))
}

func (m *TabContainerModel) handleDiscoverModelUpdate(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, 0)
	var cmd tea.Cmd
	if m.activeTab == 0 {
		m.discover, cmd = m.discover.Update(msg)
		cmds = append(cmds, cmd)
	}
	return tea.Batch(cmds...)
}

func (m *TabContainerModel) handleTimerUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.timer, cmd = m.timer.Update(msg)
	return cmd
}

func (m *TabContainerModel) handleStopwatchUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.stopwatch, cmd = m.stopwatch.Update(msg)
	return cmd
}

func (m *TabContainerModel) handleSpinnerUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	*m.spinner, cmd = m.spinner.Update(msg)
	return cmd
}

func (m *TabContainerModel) resetSpinner() {
	s := spinner.New(
		spinner.WithStyle(spinnerStyle),
		spinner.WithSpinner(spinner.Points),
	)
	m.spinner = &s
	ioStatus = ""
}

func (m *TabContainerModel) populateActiveTabContent() string {
	if m.activeTab == 0 {
		return m.discover.View()
	}
	return ""
}

func (m *TabContainerModel) getCurrentActiveUser() tea.Cmd {
	return func() tea.Msg {
		u, _, code := m.client.GetCurrentActiveUser()
		if code == http.StatusUnauthorized {
			return requireAuthMsg{}
		}
		return u
	}
}
