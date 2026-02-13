package ui

import (
	"context"
	"fmt"
	"image/color"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/basecamp/once/internal/docker"
	"github.com/basecamp/once/internal/metrics"
)

const (
	PanelHeight = 10
	PanelGap    = 1
)

type dashboardKeyMap struct {
	Up        key.Binding
	Down      key.Binding
	Settings  key.Binding
	StartStop key.Binding
	NewApp    key.Binding
	Logs      key.Binding
	Quit      key.Binding
}

func (k dashboardKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Settings, k.Logs, k.NewApp, k.StartStop, k.Quit}
}

func (k dashboardKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Up, k.Down, k.Settings, k.Logs, k.NewApp, k.StartStop, k.Quit}}
}

var dashboardKeys = dashboardKeyMap{
	Up:        key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:      key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Settings:  key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "settings")),
	StartStop: key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "start/stop")),
	NewApp:    key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new app")),
	Logs:      key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "logs")),
	Quit:      key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "quit")),
}

type Dashboard struct {
	namespace     *docker.Namespace
	scraper       *metrics.MetricsScraper
	dockerScraper *docker.Scraper
	apps          []*docker.Application
	selectedIndex int
	width, height int
	viewport      viewport.Model
	toggling      bool
	togglingApp   string
	progress      ProgressBusy
	help          Help
	showingMenu   bool
	settingsMenu  SettingsMenu
}

type dashboardTickMsg struct{}

type startStopFinishedMsg struct {
	err error
}

func NewDashboard(ns *docker.Namespace, apps []*docker.Application, selectedIndex int,
	scraper *metrics.MetricsScraper, dockerScraper *docker.Scraper) Dashboard {

	vp := viewport.New()
	vp.MouseWheelEnabled = true
	vp.KeyMap.Up.SetEnabled(false)
	vp.KeyMap.Down.SetEnabled(false)
	vp.KeyMap.PageUp.SetEnabled(false)
	vp.KeyMap.PageDown.SetEnabled(false)
	vp.KeyMap.HalfPageUp.SetEnabled(false)
	vp.KeyMap.HalfPageDown.SetEnabled(false)
	vp.KeyMap.Left.SetEnabled(false)
	vp.KeyMap.Right.SetEnabled(false)

	return Dashboard{
		namespace:     ns,
		scraper:       scraper,
		dockerScraper: dockerScraper,
		apps:          apps,
		selectedIndex: selectedIndex,
		viewport:      vp,
		help:          NewHelp(),
	}
}

func (m Dashboard) Init() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return dashboardTickMsg{} })
}

func (m Dashboard) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.progress = NewProgressBusy(m.width, Colors.Border)
		m.help.SetWidth(m.width)
		m.updateViewportSize()
		m.rebuildViewportContent()

		if m.showingMenu {
			m.settingsMenu, _ = m.settingsMenu.Update(msg)
		}

	case ComponentSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.progress = NewProgressBusy(m.width, Colors.Border)
		m.help.SetWidth(m.width)
		m.updateViewportSize()
		m.rebuildViewportContent()

	case tea.MouseClickMsg:
		if m.showingMenu {
			var cmd tea.Cmd
			m.settingsMenu, cmd = m.settingsMenu.Update(msg)
			return m, cmd
		}
		if cmd := m.help.Update(msg, dashboardKeys); cmd != nil {
			return m, cmd
		}

	case tea.MouseWheelMsg:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		if m.showingMenu {
			var cmd tea.Cmd
			m.settingsMenu, cmd = m.settingsMenu.Update(msg)
			return m, cmd
		}

		if key.Matches(msg, dashboardKeys.Quit) {
			return m, func() tea.Msg { return quitMsg{} }
		}
		if key.Matches(msg, dashboardKeys.Up) {
			if m.selectedIndex > 0 {
				m.selectedIndex--
				m.rebuildViewportContent()
				m.scrollToSelection()
			}
			return m, nil
		}
		if key.Matches(msg, dashboardKeys.Down) {
			if m.selectedIndex < len(m.apps)-1 {
				m.selectedIndex++
				m.rebuildViewportContent()
				m.scrollToSelection()
			}
			return m, nil
		}
		if key.Matches(msg, dashboardKeys.NewApp) {
			return m, func() tea.Msg { return navigateToInstallMsg{} }
		}
		if key.Matches(msg, dashboardKeys.Settings) && len(m.apps) > 0 {
			app := m.apps[m.selectedIndex]
			m.showingMenu = true
			m.settingsMenu = NewSettingsMenu(app)
			m.settingsMenu, _ = m.settingsMenu.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			return m, nil
		}
		if key.Matches(msg, dashboardKeys.StartStop) && len(m.apps) > 0 && !m.toggling {
			app := m.apps[m.selectedIndex]
			m.toggling = true
			m.togglingApp = app.Settings.Name
			m.progress = NewProgressBusy(m.width, Colors.Border)
			m.updateViewportSize()
			m.rebuildViewportContent()
			return m, tea.Batch(m.progress.Init(), m.runStartStop(app))
		}
		if key.Matches(msg, dashboardKeys.Logs) && len(m.apps) > 0 {
			return m, func() tea.Msg { return navigateToLogsMsg{app: m.apps[m.selectedIndex]} }
		}

	case SettingsMenuCloseMsg:
		m.showingMenu = false

	case SettingsMenuSelectMsg:
		m.showingMenu = false
		return m, func() tea.Msg {
			return navigateToSettingsSectionMsg(msg)
		}

	case startStopFinishedMsg:
		m.toggling = false
		m.togglingApp = ""
		m.updateViewportSize()
		m.rebuildViewportContent()

	case dashboardTickMsg:
		m.rebuildViewportContent()
		cmds = append(cmds, tea.Tick(time.Second, func(time.Time) tea.Msg { return dashboardTickMsg{} }))

	case progressBusyTickMsg:
		if m.toggling {
			var cmd tea.Cmd
			m.progress, cmd = m.progress.Update(msg)
			cmds = append(cmds, cmd)
		}

	case namespaceChangedMsg:
		previousName := ""
		if m.selectedIndex < len(m.apps) {
			previousName = m.apps[m.selectedIndex].Settings.Name
		}
		m.apps = m.namespace.Applications()
		m.selectedIndex = 0
		for i, app := range m.apps {
			if app.Settings.Name == previousName {
				m.selectedIndex = i
				break
			}
		}
		if m.selectedIndex >= len(m.apps) && len(m.apps) > 0 {
			m.selectedIndex = len(m.apps) - 1
		}
		m.rebuildViewportContent()
		m.scrollToSelection()
	}

	if m.showingMenu {
		var cmd tea.Cmd
		m.settingsMenu, cmd = m.settingsMenu.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Dashboard) View() string {
	helpView := m.help.View(dashboardKeys)
	helpLine := Styles.HelpLine(m.width, helpView)

	var content string
	if m.toggling {
		content = m.viewport.View() + "\n" + m.progress.View() + "\n" + helpLine
	} else {
		content = m.viewport.View() + "\n" + helpLine
	}

	if m.showingMenu {
		contentLayer := newZoneLayer(content)
		menuLayer := centeredZoneLayer(m.settingsMenu.View(), m.width, m.height)
		return renderPreservingZones(contentLayer, menuLayer)
	}

	return content
}

// Private

func (m Dashboard) runStartStop(app *docker.Application) tea.Cmd {
	return func() tea.Msg {
		var err error
		if app.Running {
			err = app.Stop(context.Background())
		} else {
			err = app.Start(context.Background())
		}
		return startStopFinishedMsg{err: err}
	}
}

func (m *Dashboard) updateViewportSize() {
	helpHeight := 1
	progressHeight := 0
	if m.toggling {
		progressHeight = 1
	}
	vpHeight := m.height - helpHeight - progressHeight
	if vpHeight < 0 {
		vpHeight = 0
	}
	m.viewport.SetHeight(vpHeight)
	m.viewport.SetWidth(m.width)
}

func (m *Dashboard) rebuildViewportContent() {
	panels := make([]string, len(m.apps))
	for i, app := range m.apps {
		panels[i] = m.renderPanel(app, i == m.selectedIndex)
	}
	joined := strings.Join(panels, strings.Repeat("\n", PanelGap+1))
	m.viewport.SetContent(joined)
}

func (m *Dashboard) scrollToSelection() {
	panelTop := m.selectedIndex * (PanelHeight + PanelGap)
	panelBottom := panelTop + PanelHeight
	if panelTop < m.viewport.YOffset() {
		m.viewport.SetYOffset(panelTop)
	} else if panelBottom > m.viewport.YOffset()+m.viewport.Height() {
		m.viewport.SetYOffset(panelBottom - m.viewport.Height())
	}
}

func (m Dashboard) renderPanel(app *docker.Application, selected bool) string {
	borderColor := Colors.Border
	if selected {
		borderColor = Colors.Focused
	}

	title := app.Settings.URL()
	if title == "" {
		title = app.Settings.Name
	}

	isToggling := m.toggling && m.togglingApp == app.Settings.Name
	stateLine := renderStateLine(app, isToggling)

	innerWidth := m.width - 4
	if innerWidth < 0 {
		innerWidth = 0
	}
	titleLine := lipgloss.Place(innerWidth, 1, lipgloss.Center, lipgloss.Center,
		Styles.Title.Render(title))

	content := lipgloss.JoinVertical(lipgloss.Left, titleLine, "", stateLine)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(m.width - 2).
		Height(PanelHeight - 2).
		Render(content)
}

// Helpers

func renderStateLine(app *docker.Application, toggling bool) string {
	var status string
	var statusColor color.Color
	if toggling && app.Running {
		status = "stopping..."
		statusColor = Colors.Warning
	} else if toggling {
		status = "starting..."
		statusColor = Colors.Warning
	} else if app.Running {
		status = "running"
		statusColor = Colors.Success
	} else {
		status = "stopped"
		statusColor = Colors.Error
	}

	stateStyle := lipgloss.NewStyle().Foreground(statusColor)
	stateDisplay := fmt.Sprintf("State: %s", stateStyle.Render(status))

	if app.Running && !app.RunningSince.IsZero() {
		stateDisplay += fmt.Sprintf(" (up %s)", formatDuration(time.Since(app.RunningSince)))
	}

	return stateDisplay
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		mins := int(d.Minutes()) % 60
		if mins == 0 {
			return fmt.Sprintf("%dh", hours)
		}
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	if hours == 0 {
		return fmt.Sprintf("%dd", days)
	}
	return fmt.Sprintf("%dd %dh", days, hours)
}
