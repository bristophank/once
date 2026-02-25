package ui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/basecamp/once/internal/docker"
)

type SettingsFormEnvironment struct {
	settings docker.ApplicationSettings
	form     *Form
}

func NewSettingsFormEnvironment(settings docker.ApplicationSettings) *SettingsFormEnvironment {
	m := &SettingsFormEnvironment{
		settings: settings,
		form:     NewForm("Done"),
	}

	m.form.OnSubmit(func() tea.Cmd {
		return func() tea.Msg { return SettingsSectionCancelMsg{} }
	})
	m.form.OnCancel(func() tea.Cmd {
		return func() tea.Msg { return SettingsSectionCancelMsg{} }
	})

	return m
}

func (m *SettingsFormEnvironment) Title() string {
	return "Environment"
}

func (m *SettingsFormEnvironment) Init() tea.Cmd {
	return m.form.Init()
}

func (m *SettingsFormEnvironment) Update(msg tea.Msg) tea.Cmd {
	return m.form.Update(msg)
}

func (m *SettingsFormEnvironment) StatusLine() string { return "" }

func (m *SettingsFormEnvironment) View() string {
	placeholder := lipgloss.NewStyle().
		Foreground(Colors.Border).
		Italic(true).
		Render("(Environment variable editing coming soon)")

	return lipgloss.JoinVertical(lipgloss.Left,
		placeholder,
		"",
		m.form.View(),
	)
}
