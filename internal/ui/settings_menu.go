package ui

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/basecamp/amar/internal/docker"
)

type settingsMenuItem int

const (
	menuItemApplication settingsMenuItem = iota
	menuItemEmail
	menuItemEnvironment
	menuItemCount
)

type SettingsMenuCloseMsg struct{}

type SettingsMenuSelectMsg struct {
	app     *docker.Application
	section SettingsSectionType
}

type SettingsMenu struct {
	app           *docker.Application
	selected      settingsMenuItem
	width, height int
}

func NewSettingsMenu(app *docker.Application) SettingsMenu {
	return SettingsMenu{
		app:      app,
		selected: menuItemApplication,
	}
}

func (m SettingsMenu) Init() tea.Cmd {
	return nil
}

func (m SettingsMenu) Update(msg tea.Msg) (SettingsMenu, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			m.selected = (m.selected - 1 + menuItemCount) % menuItemCount
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			m.selected = (m.selected + 1) % menuItemCount
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			return m, m.selectCurrent()
		case key.Matches(msg, key.NewBinding(key.WithKeys("a"))):
			return m, m.selectSection(SettingsSectionApplication)
		case key.Matches(msg, key.NewBinding(key.WithKeys("e"))):
			return m, m.selectSection(SettingsSectionEmail)
		case key.Matches(msg, key.NewBinding(key.WithKeys("v"))):
			return m, m.selectSection(SettingsSectionEnvironment)
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			return m, func() tea.Msg { return SettingsMenuCloseMsg{} }
		}
	}

	return m, nil
}

func (m SettingsMenu) View() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Border).
		Padding(1, 2)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(Colors.Primary).
		MarginBottom(1)

	itemStyle := lipgloss.NewStyle().
		PaddingLeft(2)

	selectedStyle := itemStyle.
		Foreground(Colors.Focused).
		Bold(true)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6272a4")).
		MarginTop(1).
		Align(lipgloss.Center)

	title := titleStyle.Render("Settings")

	items := []string{
		m.renderItem("[A]pplication", menuItemApplication, itemStyle, selectedStyle),
		m.renderItem("[E]mail", menuItemEmail, itemStyle, selectedStyle),
		m.renderItem("En[v]ironment", menuItemEnvironment, itemStyle, selectedStyle),
	}

	help := helpStyle.Render("esc to close")

	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		strings.Join(items, "\n"),
		help,
	)

	return boxStyle.Render(content)
}

// Private

func (m SettingsMenu) renderItem(label string, item settingsMenuItem, normal, selected lipgloss.Style) string {
	if m.selected == item {
		return selected.Render("> " + label)
	}
	return normal.Render("  " + label)
}

func (m SettingsMenu) selectCurrent() tea.Cmd {
	var section SettingsSectionType
	switch m.selected {
	case menuItemApplication:
		section = SettingsSectionApplication
	case menuItemEmail:
		section = SettingsSectionEmail
	case menuItemEnvironment:
		section = SettingsSectionEnvironment
	}
	return m.selectSection(section)
}

func (m SettingsMenu) selectSection(section SettingsSectionType) tea.Cmd {
	return func() tea.Msg {
		return SettingsMenuSelectMsg{app: m.app, section: section}
	}
}
