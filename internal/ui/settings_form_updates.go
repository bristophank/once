package ui

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/basecamp/once/internal/docker"
)

const updatesAutoUpdateField = 0

type SettingsFormUpdates struct {
	app        *docker.Application
	settings   docker.ApplicationSettings
	form       *Form
	lastResult *docker.OperationResult
}

func NewSettingsFormUpdates(app *docker.Application, lastResult *docker.OperationResult) *SettingsFormUpdates {
	autoUpdateField := NewCheckboxField("Automatically apply updates", app.Settings.AutoUpdate)

	m := &SettingsFormUpdates{
		app:      app,
		settings: app.Settings,
		form: NewForm("Done",
			FormItem{Label: "Updates", Field: autoUpdateField},
		),
		lastResult: lastResult,
	}

	m.form.SetActionButton("Check for updates", func() tea.Msg {
		return settingsRunActionMsg{action: func() (string, error) {
			changed, err := app.Update(context.Background(), nil)
			if err != nil {
				return "", err
			}
			if !changed {
				return "Already running the latest version", nil
			}
			return "Update complete", nil
		}}
	})
	m.form.OnSubmit(func() tea.Cmd {
		m.settings.AutoUpdate = m.form.CheckboxField(updatesAutoUpdateField).Checked()
		return func() tea.Msg { return SettingsSectionSubmitMsg{Settings: m.settings} }
	})
	m.form.OnCancel(func() tea.Cmd {
		return func() tea.Msg { return SettingsSectionCancelMsg{} }
	})

	return m
}

func (m *SettingsFormUpdates) Title() string {
	return "Updates"
}

func (m *SettingsFormUpdates) Init() tea.Cmd {
	return m.form.Init()
}

func (m *SettingsFormUpdates) Update(msg tea.Msg) tea.Cmd {
	return m.form.Update(msg)
}

func (m *SettingsFormUpdates) View() string {
	return m.form.View()
}

func (m *SettingsFormUpdates) StatusLine() string {
	return formatOperationStatus("checked", m.lastResult)
}
