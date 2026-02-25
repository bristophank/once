package ui

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/basecamp/once/internal/docker"
)

const (
	backupsPathField = iota
	backupsAutoBackField
)

type SettingsFormBackups struct {
	app        *docker.Application
	settings   docker.ApplicationSettings
	form       *Form
	lastResult *docker.OperationResult
}

func NewSettingsFormBackups(app *docker.Application, lastResult *docker.OperationResult) *SettingsFormBackups {
	pathField := NewTextField("/path/to/backups")
	pathField.SetValue(app.Settings.Backup.Path)

	autoBackField := NewCheckboxField("Automatically create backups", app.Settings.Backup.AutoBack)

	m := &SettingsFormBackups{
		app:      app,
		settings: app.Settings,
		form: NewForm("Done",
			FormItem{Label: "Backup location", Field: pathField},
			FormItem{Label: "Backups", Field: autoBackField},
		),
		lastResult: lastResult,
	}

	m.form.SetActionButton("Run backup now", func() tea.Msg {
		return settingsRunActionMsg{action: func() (string, error) {
			return "Backup complete", runBackup(app, pathField.Value())
		}}
	})
	m.form.OnSubmit(func() tea.Cmd {
		m.settings.Backup.Path = m.form.TextField(backupsPathField).Value()
		m.settings.Backup.AutoBack = m.form.CheckboxField(backupsAutoBackField).Checked()
		return func() tea.Msg { return SettingsSectionSubmitMsg{Settings: m.settings} }
	})
	m.form.OnCancel(func() tea.Cmd {
		return func() tea.Msg { return SettingsSectionCancelMsg{} }
	})

	return m
}

func (m *SettingsFormBackups) Title() string {
	return "Backups"
}

func (m *SettingsFormBackups) Init() tea.Cmd {
	return m.form.Init()
}

func (m *SettingsFormBackups) Update(msg tea.Msg) tea.Cmd {
	return m.form.Update(msg)
}

func (m *SettingsFormBackups) View() string {
	return m.form.View()
}

func (m *SettingsFormBackups) StatusLine() string {
	return formatOperationStatus("backup", m.lastResult)
}

// Helpers

func runBackup(app *docker.Application, dir string) error {
	return app.BackupToFile(context.Background(), dir, app.BackupName())
}
