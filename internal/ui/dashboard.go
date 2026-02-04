package ui

import (
	"context"
	"fmt"
	"image/color"
	"slices"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/basecamp/amar/internal/docker"
	"github.com/basecamp/amar/internal/metrics"
)

var chartColors = struct {
	Green  lipgloss.Style
	Red    lipgloss.Style
	Blue   lipgloss.Style
	Purple lipgloss.Style
}{
	Green:  lipgloss.NewStyle().Foreground(lipgloss.Color("#50fa7b")),
	Red:    lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5555")),
	Blue:   lipgloss.NewStyle().Foreground(lipgloss.Color("#8be9fd")),
	Purple: lipgloss.NewStyle().Foreground(lipgloss.Color("#bd93f9")),
}

type dashboardKeyMap struct {
	Settings key.Binding
	Upgrade  key.Binding
	NewApp   key.Binding
	Logs     key.Binding
	PrevApp  key.Binding
	NextApp  key.Binding
	Quit     key.Binding
}

func (k dashboardKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.PrevApp, k.NextApp, k.Settings, k.Logs, k.NewApp, k.Upgrade, k.Quit}
}

func (k dashboardKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.PrevApp, k.NextApp, k.Settings, k.Logs, k.NewApp, k.Upgrade, k.Quit}}
}

var dashboardKeys = dashboardKeyMap{
	Settings: key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "settings")),
	Upgrade:  key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "upgrade")),
	NewApp:   key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new app")),
	Logs:     key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "logs")),
	PrevApp:  key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "prev app")),
	NextApp:  key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "next app")),
	Quit:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "quit")),
}

// dashboardState holds state that Content render functions need access to.
// Using a pointer allows closures to see current values.
type dashboardState struct {
	app       *docker.Application
	upgrading bool
	progress  ProgressBusy
	help      help.Model
}

type Dashboard struct {
	namespace     *docker.Namespace
	scraper       *metrics.MetricsScraper
	dockerScraper *docker.Scraper
	width, height int
	showingMenu   bool
	settingsMenu  SettingsMenu
	state         *dashboardState
	layout        StackLayout
}

type dashboardTickMsg struct{}

type upgradeFinishedMsg struct {
	err error
}

func NewDashboard(ns *docker.Namespace, app *docker.Application, scraper *metrics.MetricsScraper, dockerScraper *docker.Scraper) Dashboard {
	service := app.Settings.Name

	state := &dashboardState{
		app:  app,
		help: help.New(),
	}

	header := NewContent(func(width, height int) string {
		return renderInfoBox(width, state.app, state.upgrading)
	})

	footer := NewContent(func(width, height int) string {
		helpView := state.help.View(dashboardKeys)
		helpLine := lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(helpView)
		if state.upgrading {
			return state.progress.View() + "\n" + helpLine
		}
		return helpLine
	})

	allReqChart := NewChart("Requests/min", chartColors.Green, UnitCount, func() []float64 {
		samples := scraper.Fetch(service, ChartHistoryLength)
		data := make([]float64, len(samples))
		for i, s := range samples {
			data[i] = float64(s.Success + s.ClientErrors + s.ServerErrors)
		}
		slices.Reverse(data)
		return SlidingSum(data, ChartSlidingWindow)
	})

	errorChart := NewChart("Errors/min", chartColors.Red, UnitCount, func() []float64 {
		samples := scraper.Fetch(service, ChartHistoryLength)
		data := make([]float64, len(samples))
		for i, s := range samples {
			data[i] = float64(s.ServerErrors)
		}
		slices.Reverse(data)
		return SlidingSum(data, ChartSlidingWindow)
	})

	cpuChart := NewChart("CPU", chartColors.Blue, UnitPercent, func() []float64 {
		samples := dockerScraper.Fetch(service, ChartHistoryLength)
		data := make([]float64, len(samples))
		for i, s := range samples {
			data[i] = s.CPUPercent
		}
		slices.Reverse(data)
		return data
	})

	memoryChart := NewChart("Memory", chartColors.Purple, UnitBytes, func() []float64 {
		samples := dockerScraper.Fetch(service, ChartHistoryLength)
		data := make([]float64, len(samples))
		for i, s := range samples {
			data[i] = float64(s.MemoryBytes)
		}
		slices.Reverse(data)
		return data
	})

	allReqChart.refreshData()
	errorChart.refreshData()
	cpuChart.refreshData()
	memoryChart.refreshData()

	chartRow1 := NewStackLayout(Horizontal,
		WithPercent(50, allReqChart),
		WithFill(errorChart),
	)
	chartRow2 := NewStackLayout(Horizontal,
		WithPercent(50, cpuChart),
		WithFill(memoryChart),
	)
	chartsLayout := NewStackLayout(Vertical,
		WithPercent(50, chartRow1),
		WithFill(chartRow2),
	)

	layout := NewStackLayout(Vertical,
		StackChild{Component: header, Size: Fit()},
		StackChild{Component: chartsLayout, Size: Fill()},
		StackChild{Component: footer, Size: Fit()},
	)

	return Dashboard{
		namespace:     ns,
		scraper:       scraper,
		dockerScraper: dockerScraper,
		state:         state,
		layout:        layout,
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
		m.state.progress = NewProgressBusy(m.width, lipgloss.Color("#6272a4"))
		m.state.help.SetWidth(m.width)

		updated, _ := m.layout.Update(ComponentSizeMsg{Width: m.width, Height: m.height})
		m.layout = updated.(StackLayout)

		if m.state.upgrading {
			cmds = append(cmds, m.state.progress.Init())
		}
		if m.showingMenu {
			m.settingsMenu, _ = m.settingsMenu.Update(msg)
		}

	case ComponentSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.state.progress = NewProgressBusy(m.width, lipgloss.Color("#6272a4"))
		m.state.help.SetWidth(m.width)

		updated, _ := m.layout.Update(msg)
		m.layout = updated.(StackLayout)

	case tea.KeyMsg:
		if m.showingMenu {
			var cmd tea.Cmd
			m.settingsMenu, cmd = m.settingsMenu.Update(msg)
			return m, cmd
		}

		if key.Matches(msg, dashboardKeys.Quit) {
			return m, func() tea.Msg { return quitMsg{} }
		}
		if key.Matches(msg, dashboardKeys.PrevApp) {
			return m, func() tea.Msg { return switchAppMsg{delta: -1} }
		}
		if key.Matches(msg, dashboardKeys.NextApp) {
			return m, func() tea.Msg { return switchAppMsg{delta: 1} }
		}
		if key.Matches(msg, dashboardKeys.NewApp) {
			return m, func() tea.Msg { return navigateToInstallMsg{} }
		}
		if key.Matches(msg, dashboardKeys.Settings) {
			m.showingMenu = true
			m.settingsMenu = NewSettingsMenu(m.state.app)
			m.settingsMenu, _ = m.settingsMenu.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			return m, nil
		}
		if key.Matches(msg, dashboardKeys.Upgrade) && !m.state.upgrading {
			m.state.upgrading = true
			m.state.progress = NewProgressBusy(m.width, lipgloss.Color("#6272a4"))
			return m, tea.Batch(m.state.progress.Init(), m.runUpgrade())
		}
		if key.Matches(msg, dashboardKeys.Logs) {
			return m, func() tea.Msg { return navigateToLogsMsg{app: m.state.app} }
		}

	case SettingsMenuCloseMsg:
		m.showingMenu = false

	case SettingsMenuSelectMsg:
		m.showingMenu = false
		return m, func() tea.Msg {
			return navigateToSettingsSectionMsg(msg)
		}

	case upgradeFinishedMsg:
		m.state.upgrading = false

	case dashboardTickMsg:
		cmds = append(cmds, tea.Tick(time.Second, func(time.Time) tea.Msg { return dashboardTickMsg{} }))

	case scrapeDoneMsg:
		updated, _ := m.layout.Update(ChartRefreshMsg{})
		m.layout = updated.(StackLayout)

	case progressBusyTickMsg:
		if m.state.upgrading {
			var cmd tea.Cmd
			m.state.progress, cmd = m.state.progress.Update(msg)
			cmds = append(cmds, cmd)
		}

	case namespaceChangedMsg:
		if app := m.namespace.Application(m.state.app.Settings.Name); app != nil {
			m.state.app = app
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Dashboard) View() string {
	content := m.layout.View()

	if m.showingMenu {
		contentLayer := lipgloss.NewLayer(content)
		menuLayer := CenteredLayer(m.settingsMenu.View(), m.width, m.height)
		return lipgloss.NewCanvas(contentLayer, menuLayer).Render()
	}

	return content
}

// Private

func (m Dashboard) runUpgrade() tea.Cmd {
	return func() tea.Msg {
		err := m.state.app.Update(context.Background(), nil)
		return upgradeFinishedMsg{err: err}
	}
}

// Helpers

func renderInfoBox(width int, app *docker.Application, upgrading bool) string {
	var status string
	var statusColor color.Color
	if upgrading {
		status = "upgrading..."
		statusColor = lipgloss.Color("#f1fa8c")
	} else if app.Running {
		status = "running"
		statusColor = lipgloss.Color("#50fa7b")
	} else {
		status = "stopped"
		statusColor = lipgloss.Color("#ff5555")
	}

	stateStyle := lipgloss.NewStyle().Foreground(statusColor)
	stateDisplay := fmt.Sprintf("State: %s", stateStyle.Render(status))

	if app.Running && !app.RunningSince.IsZero() && !upgrading {
		stateDisplay += fmt.Sprintf(" (up %s)", formatDuration(time.Since(app.RunningSince)))
	}

	var extraLines []string
	extraLines = append(extraLines, stateDisplay)
	if url := app.Settings.URL(); url != "" {
		extraLines = append(extraLines, fmt.Sprintf("URL: %s", url))
	}
	return Styles.TitleBox(width, app.Settings.Name, extraLines...)
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
