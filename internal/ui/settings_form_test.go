package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basecamp/amar/internal/docker"
)

func TestSettingsFormApplication_InitialState_NonLocalhost(t *testing.T) {
	settings := docker.ApplicationSettings{
		Image:      "nginx:latest",
		Host:       "app.example.com",
		DisableTLS: false,
	}
	form := NewSettingsFormApplication(settings)

	assert.Equal(t, applicationFieldImage, form.focused)
	assert.Equal(t, "nginx:latest", form.imageInput.Value())
	assert.Equal(t, "app.example.com", form.hostnameInput.Value())
	assert.True(t, form.settings.TLSEnabled())
}

func TestSettingsFormApplication_InitialState_Localhost(t *testing.T) {
	settings := docker.ApplicationSettings{
		Image:      "nginx:latest",
		Host:       "chat.localhost",
		DisableTLS: false,
	}
	form := NewSettingsFormApplication(settings)

	assert.Equal(t, "chat.localhost", form.hostnameInput.Value())
	assert.False(t, form.settings.TLSEnabled(), "TLS should be disabled for localhost even when DisableTLS is false")
}

func TestSettingsFormApplication_TabNavigation(t *testing.T) {
	form := NewSettingsFormApplication(docker.ApplicationSettings{Host: "app.example.com"})
	assert.Equal(t, applicationFieldImage, form.focused)

	form = applicationPressTab(form)
	assert.Equal(t, applicationFieldHostname, form.focused)

	form = applicationPressTab(form)
	assert.Equal(t, applicationFieldTLS, form.focused)

	form = applicationPressTab(form)
	assert.Equal(t, applicationFieldDoneButton, form.focused)

	form = applicationPressTab(form)
	assert.Equal(t, applicationFieldCancelButton, form.focused)

	form = applicationPressTab(form)
	assert.Equal(t, applicationFieldImage, form.focused)
}

func TestSettingsFormApplication_ShiftTabNavigation(t *testing.T) {
	form := NewSettingsFormApplication(docker.ApplicationSettings{Host: "app.example.com"})

	form = applicationPressShiftTab(form)
	assert.Equal(t, applicationFieldCancelButton, form.focused)

	form = applicationPressShiftTab(form)
	assert.Equal(t, applicationFieldDoneButton, form.focused)
}

func TestSettingsFormApplication_SpaceTogglesTLS(t *testing.T) {
	form := NewSettingsFormApplication(docker.ApplicationSettings{Host: "app.example.com"})
	assert.True(t, form.settings.TLSEnabled())

	form = applicationPressTab(form)
	form = applicationPressTab(form)
	assert.Equal(t, applicationFieldTLS, form.focused)

	form = applicationPressSpace(form)
	assert.False(t, form.settings.TLSEnabled())

	form = applicationPressSpace(form)
	assert.True(t, form.settings.TLSEnabled())
}

func TestSettingsFormApplication_SpaceDoesNotToggleTLSForLocalhost(t *testing.T) {
	form := NewSettingsFormApplication(docker.ApplicationSettings{Host: "chat.localhost"})
	assert.False(t, form.settings.TLSEnabled())

	form = applicationPressTab(form)
	form = applicationPressTab(form)
	assert.Equal(t, applicationFieldTLS, form.focused)

	form = applicationPressSpace(form)
	assert.False(t, form.settings.TLSEnabled(), "TLS should remain disabled for localhost")
}

func TestSettingsFormApplication_TLSShowsDisabledForLocalhost(t *testing.T) {
	form := NewSettingsFormApplication(docker.ApplicationSettings{Host: "app.example.com"})
	assert.True(t, form.settings.TLSEnabled())

	form = applicationPressTab(form)
	form = applicationTypeText(form, ".localhost")
	assert.False(t, form.settings.TLSEnabled(), "TLS should show as disabled for localhost")

	form = applicationClearAndType(form, "app.example.com")
	assert.True(t, form.settings.TLSEnabled(), "TLS preference should be preserved")
}

func TestSettingsFormApplication_Submit(t *testing.T) {
	form := NewSettingsFormApplication(docker.ApplicationSettings{
		Name:  "myapp",
		Image: "nginx:latest",
		Host:  "app.example.com",
	})

	form.focused = applicationFieldDoneButton
	section, cmd := form.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	form = section.(SettingsFormApplication)
	require.NotNil(t, cmd)

	msg := cmd()
	submitMsg, ok := msg.(SettingsSectionSubmitMsg)
	require.True(t, ok, "expected SettingsSectionSubmitMsg, got %T", msg)
	assert.Equal(t, "myapp", submitMsg.Settings.Name)
	assert.Equal(t, "nginx:latest", submitMsg.Settings.Image)
	assert.Equal(t, "app.example.com", submitMsg.Settings.Host)
	assert.False(t, submitMsg.Settings.DisableTLS)
}

func TestSettingsFormApplication_Cancel(t *testing.T) {
	form := NewSettingsFormApplication(docker.ApplicationSettings{Host: "app.example.com"})

	form.focused = applicationFieldCancelButton
	_, cmd := form.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(SettingsSectionCancelMsg)
	assert.True(t, ok, "expected SettingsSectionCancelMsg, got %T", msg)
}

func TestSettingsFormEmail_InitialState(t *testing.T) {
	settings := docker.ApplicationSettings{
		SMTP: docker.SMTPSettings{
			Server:   "smtp.example.com",
			Port:     "587",
			Username: "user@example.com",
			Password: "secret",
			From:     "noreply@example.com",
		},
	}
	form := NewSettingsFormEmail(settings)

	assert.Equal(t, emailFieldServer, form.focused)
	assert.Equal(t, "smtp.example.com", form.serverInput.Value())
	assert.Equal(t, "587", form.portInput.Value())
	assert.Equal(t, "user@example.com", form.usernameInput.Value())
	assert.Equal(t, "secret", form.passwordInput.Value())
	assert.Equal(t, "noreply@example.com", form.fromInput.Value())
}

func TestSettingsFormEmail_TabNavigation(t *testing.T) {
	form := NewSettingsFormEmail(docker.ApplicationSettings{})
	assert.Equal(t, emailFieldServer, form.focused)

	form = emailPressTab(form)
	assert.Equal(t, emailFieldPort, form.focused)

	form = emailPressTab(form)
	assert.Equal(t, emailFieldUsername, form.focused)

	form = emailPressTab(form)
	assert.Equal(t, emailFieldPassword, form.focused)

	form = emailPressTab(form)
	assert.Equal(t, emailFieldFrom, form.focused)

	form = emailPressTab(form)
	assert.Equal(t, emailFieldDoneButton, form.focused)

	form = emailPressTab(form)
	assert.Equal(t, emailFieldCancelButton, form.focused)

	form = emailPressTab(form)
	assert.Equal(t, emailFieldServer, form.focused)
}

func TestSettingsFormEmail_Submit(t *testing.T) {
	settings := docker.ApplicationSettings{
		Name: "myapp",
		SMTP: docker.SMTPSettings{
			Server: "smtp.example.com",
			Port:   "587",
		},
	}
	form := NewSettingsFormEmail(settings)

	form.focused = emailFieldDoneButton
	section, cmd := form.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	form = section.(SettingsFormEmail)
	require.NotNil(t, cmd)

	msg := cmd()
	submitMsg, ok := msg.(SettingsSectionSubmitMsg)
	require.True(t, ok, "expected SettingsSectionSubmitMsg, got %T", msg)
	assert.Equal(t, "myapp", submitMsg.Settings.Name)
	assert.Equal(t, "smtp.example.com", submitMsg.Settings.SMTP.Server)
	assert.Equal(t, "587", submitMsg.Settings.SMTP.Port)
}

func TestSettingsFormEmail_Cancel(t *testing.T) {
	form := NewSettingsFormEmail(docker.ApplicationSettings{})

	form.focused = emailFieldCancelButton
	_, cmd := form.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(SettingsSectionCancelMsg)
	assert.True(t, ok, "expected SettingsSectionCancelMsg, got %T", msg)
}

// Helpers

func applicationTypeText(form SettingsFormApplication, text string) SettingsFormApplication {
	for _, r := range text {
		section, _ := form.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		form = section.(SettingsFormApplication)
	}
	return form
}

func applicationClearAndType(form SettingsFormApplication, text string) SettingsFormApplication {
	form.hostnameInput.SetValue("")
	form.settings.Host = ""
	return applicationTypeText(form, text)
}

func applicationPressTab(form SettingsFormApplication) SettingsFormApplication {
	section, _ := form.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	return section.(SettingsFormApplication)
}

func applicationPressShiftTab(form SettingsFormApplication) SettingsFormApplication {
	section, _ := form.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	return section.(SettingsFormApplication)
}

func applicationPressSpace(form SettingsFormApplication) SettingsFormApplication {
	section, _ := form.Update(tea.KeyPressMsg{Code: tea.KeySpace})
	return section.(SettingsFormApplication)
}

func emailPressTab(form SettingsFormEmail) SettingsFormEmail {
	section, _ := form.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	return section.(SettingsFormEmail)
}
