package ui

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/basecamp/amar/internal/docker"
)

type emailFormField int

const (
	emailFieldServer emailFormField = iota
	emailFieldPort
	emailFieldUsername
	emailFieldPassword
	emailFieldFrom
	emailFieldDoneButton
	emailFieldCancelButton
	emailFieldCount
)

type SettingsFormEmail struct {
	width, height int
	focused       emailFormField
	settings      docker.ApplicationSettings
	serverInput   textinput.Model
	portInput     textinput.Model
	usernameInput textinput.Model
	passwordInput textinput.Model
	fromInput     textinput.Model
}

func NewSettingsFormEmail(settings docker.ApplicationSettings) SettingsFormEmail {
	server := textinput.New()
	server.Placeholder = "smtp.example.com"
	server.Prompt = ""
	server.CharLimit = 256
	server.SetValue(settings.SMTP.Server)
	server.Focus()

	port := textinput.New()
	port.Placeholder = "587"
	port.Prompt = ""
	port.CharLimit = 5
	port.SetValue(settings.SMTP.Port)

	username := textinput.New()
	username.Placeholder = "user@example.com"
	username.Prompt = ""
	username.CharLimit = 256
	username.SetValue(settings.SMTP.Username)

	password := textinput.New()
	password.Placeholder = "password"
	password.Prompt = ""
	password.CharLimit = 256
	password.EchoMode = textinput.EchoPassword
	password.SetValue(settings.SMTP.Password)

	from := textinput.New()
	from.Placeholder = "noreply@example.com"
	from.Prompt = ""
	from.CharLimit = 256
	from.SetValue(settings.SMTP.From)

	return SettingsFormEmail{
		focused:       emailFieldServer,
		settings:      settings,
		serverInput:   server,
		portInput:     port,
		usernameInput: username,
		passwordInput: password,
		fromInput:     from,
	}
}

func (m SettingsFormEmail) Title() string {
	return "Email"
}

func (m SettingsFormEmail) Init() tea.Cmd {
	return nil
}

func (m SettingsFormEmail) Update(msg tea.Msg) (SettingsSection, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		inputWidth := min(m.width-4, 60)
		m.serverInput.SetWidth(inputWidth)
		m.portInput.SetWidth(inputWidth)
		m.usernameInput.SetWidth(inputWidth)
		m.passwordInput.SetWidth(inputWidth)
		m.fromInput.SetWidth(inputWidth)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
			return m.focusNext()
		case key.Matches(msg, key.NewBinding(key.WithKeys("shift+tab"))):
			return m.focusPrev()
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			return m.handleEnter()
		}
	}

	switch m.focused {
	case emailFieldServer:
		var cmd tea.Cmd
		m.serverInput, cmd = m.serverInput.Update(msg)
		m.settings.SMTP.Server = m.serverInput.Value()
		cmds = append(cmds, cmd)
	case emailFieldPort:
		var cmd tea.Cmd
		m.portInput, cmd = m.portInput.Update(msg)
		m.settings.SMTP.Port = m.portInput.Value()
		cmds = append(cmds, cmd)
	case emailFieldUsername:
		var cmd tea.Cmd
		m.usernameInput, cmd = m.usernameInput.Update(msg)
		m.settings.SMTP.Username = m.usernameInput.Value()
		cmds = append(cmds, cmd)
	case emailFieldPassword:
		var cmd tea.Cmd
		m.passwordInput, cmd = m.passwordInput.Update(msg)
		m.settings.SMTP.Password = m.passwordInput.Value()
		cmds = append(cmds, cmd)
	case emailFieldFrom:
		var cmd tea.Cmd
		m.fromInput, cmd = m.fromInput.Update(msg)
		m.settings.SMTP.From = m.fromInput.Value()
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m SettingsFormEmail) View() string {
	serverLabel := Styles.Label.Render("SMTP Server")
	serverField := Styles.Focus(Styles.Input, m.focused == emailFieldServer).
		Render(m.serverInput.View())

	portLabel := Styles.Label.Render("SMTP Port")
	portField := Styles.Focus(Styles.Input, m.focused == emailFieldPort).
		Render(m.portInput.View())

	usernameLabel := Styles.Label.Render("SMTP Username")
	usernameField := Styles.Focus(Styles.Input, m.focused == emailFieldUsername).
		Render(m.usernameInput.View())

	passwordLabel := Styles.Label.Render("SMTP Password")
	passwordField := Styles.Focus(Styles.Input, m.focused == emailFieldPassword).
		Render(m.passwordInput.View())

	fromLabel := Styles.Label.Render("SMTP From")
	fromField := Styles.Focus(Styles.Input, m.focused == emailFieldFrom).
		Render(m.fromInput.View())

	doneButton := Styles.Focus(Styles.ButtonPrimary, m.focused == emailFieldDoneButton).
		Render("Done")
	cancelButton := Styles.Focus(Styles.Button, m.focused == emailFieldCancelButton).
		Render("Cancel")

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, doneButton, cancelButton)

	form := lipgloss.JoinVertical(lipgloss.Left,
		serverLabel,
		serverField,
		portLabel,
		portField,
		usernameLabel,
		usernameField,
		passwordLabel,
		passwordField,
		fromLabel,
		fromField,
		"",
		buttons,
	)

	return form
}

// Private

func (m SettingsFormEmail) focusNext() (SettingsSection, tea.Cmd) {
	m.blurCurrent()
	m.focused = (m.focused + 1) % emailFieldCount
	return m.focusCurrent()
}

func (m SettingsFormEmail) focusPrev() (SettingsSection, tea.Cmd) {
	m.blurCurrent()
	m.focused = (m.focused - 1 + emailFieldCount) % emailFieldCount
	return m.focusCurrent()
}

func (m *SettingsFormEmail) blurCurrent() {
	switch m.focused {
	case emailFieldServer:
		m.serverInput.Blur()
	case emailFieldPort:
		m.portInput.Blur()
	case emailFieldUsername:
		m.usernameInput.Blur()
	case emailFieldPassword:
		m.passwordInput.Blur()
	case emailFieldFrom:
		m.fromInput.Blur()
	}
}

func (m SettingsFormEmail) focusCurrent() (SettingsSection, tea.Cmd) {
	var cmd tea.Cmd
	switch m.focused {
	case emailFieldServer:
		cmd = m.serverInput.Focus()
	case emailFieldPort:
		cmd = m.portInput.Focus()
	case emailFieldUsername:
		cmd = m.usernameInput.Focus()
	case emailFieldPassword:
		cmd = m.passwordInput.Focus()
	case emailFieldFrom:
		cmd = m.fromInput.Focus()
	}
	return m, cmd
}

func (m SettingsFormEmail) handleEnter() (SettingsSection, tea.Cmd) {
	switch m.focused {
	case emailFieldServer, emailFieldPort, emailFieldUsername, emailFieldPassword, emailFieldFrom:
		return m.focusNext()
	case emailFieldDoneButton:
		return m, func() tea.Msg { return SettingsSectionSubmitMsg{Settings: m.settings} }
	case emailFieldCancelButton:
		return m, func() tea.Msg { return SettingsSectionCancelMsg{} }
	}
	return m, nil
}
