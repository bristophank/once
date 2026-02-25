package ui

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/basecamp/once/internal/docker"
)

var settingsMenuCloseKey = WithHelp(NewKeyBinding("esc"), "esc", "close")

type SettingsMenuCloseMsg struct{}

type SettingsMenuSelectMsg struct {
	app     *docker.Application
	section SettingsSectionType
}

type SettingsMenu struct {
	app  *docker.Application
	menu Menu
	help Help
}

func NewSettingsMenu(app *docker.Application) SettingsMenu {
	return SettingsMenu{
		app: app,
		menu: NewMenu(
			MenuItem{Label: "Application", Key: int(SettingsSectionApplication), Shortcut: WithHelp(NewKeyBinding("a"), "a", "")},
			MenuItem{Label: "Email", Key: int(SettingsSectionEmail), Shortcut: WithHelp(NewKeyBinding("e"), "e", "")},
			MenuItem{Label: "Environment", Key: int(SettingsSectionEnvironment), Shortcut: WithHelp(NewKeyBinding("v"), "v", "")},
			MenuItem{Label: "Resources", Key: int(SettingsSectionResources), Shortcut: WithHelp(NewKeyBinding("r"), "r", "")},
			MenuItem{Label: "Updates", Key: int(SettingsSectionUpdates), Shortcut: WithHelp(NewKeyBinding("u"), "u", "")},
			MenuItem{Label: "Backups", Key: int(SettingsSectionBackups), Shortcut: WithHelp(NewKeyBinding("b"), "b", "")},
		),
		help: NewHelp(),
	}
}

func (m *SettingsMenu) Init() tea.Cmd {
	return nil
}

func (m *SettingsMenu) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case MouseEvent:
		if cmd := m.help.Update(msg); cmd != nil {
			return cmd
		}

	case tea.KeyPressMsg:
		if key.Matches(msg, settingsMenuCloseKey) {
			return func() tea.Msg { return SettingsMenuCloseMsg{} }
		}

	case MenuSelectMsg:
		return m.selectSection(SettingsSectionType(msg.Key))
	}

	return m.menu.Update(msg)
}

func (m *SettingsMenu) View() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Border).
		Padding(1, 4)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(Colors.Primary).
		MarginBottom(1)

	title := titleStyle.Render("Settings")

	menuView := m.menu.View()

	helpView := m.help.View([]key.Binding{settingsMenuCloseKey})
	menuWidth := lipgloss.Width(menuView)
	helpLine := lipgloss.NewStyle().MarginTop(1).Width(menuWidth).Align(lipgloss.Center).Render(helpView)

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		menuView,
		helpLine,
	)

	return boxStyle.Render(content)
}

// Private

func (m *SettingsMenu) selectSection(section SettingsSectionType) tea.Cmd {
	return func() tea.Msg {
		return SettingsMenuSelectMsg{app: m.app, section: section}
	}
}
