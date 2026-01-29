package ui

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/basecamp/amar/internal/docker"
)

type applicationFormField int

const (
	applicationFieldImage applicationFormField = iota
	applicationFieldHostname
	applicationFieldTLS
	applicationFieldDoneButton
	applicationFieldCancelButton
	applicationFieldCount
)

type SettingsFormApplication struct {
	width, height int
	focused       applicationFormField
	settings      docker.ApplicationSettings
	imageInput    textinput.Model
	hostnameInput textinput.Model
}

func NewSettingsFormApplication(settings docker.ApplicationSettings) SettingsFormApplication {
	image := textinput.New()
	image.Placeholder = "user/repo:tag"
	image.Prompt = ""
	image.CharLimit = 256
	image.SetValue(settings.Image)
	image.Focus()

	hostname := textinput.New()
	hostname.Placeholder = "app.example.com"
	hostname.Prompt = ""
	hostname.CharLimit = 256
	hostname.SetValue(settings.Host)

	return SettingsFormApplication{
		focused:       applicationFieldImage,
		settings:      settings,
		imageInput:    image,
		hostnameInput: hostname,
	}
}

func (m SettingsFormApplication) Title() string {
	return "Application"
}

func (m SettingsFormApplication) Init() tea.Cmd {
	return nil
}

func (m SettingsFormApplication) Update(msg tea.Msg) (SettingsSection, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		inputWidth := min(m.width-4, 60)
		m.imageInput.SetWidth(inputWidth)
		m.hostnameInput.SetWidth(inputWidth)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
			return m.focusNext()
		case key.Matches(msg, key.NewBinding(key.WithKeys("shift+tab"))):
			return m.focusPrev()
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			return m.handleEnter()
		case key.Matches(msg, key.NewBinding(key.WithKeys("space"))) && m.focused == applicationFieldTLS:
			if !docker.IsLocalhost(m.settings.Host) {
				m.settings.DisableTLS = !m.settings.DisableTLS
			}
			return m, nil
		}
	}

	switch m.focused {
	case applicationFieldImage:
		var cmd tea.Cmd
		m.imageInput, cmd = m.imageInput.Update(msg)
		m.settings.Image = m.imageInput.Value()
		cmds = append(cmds, cmd)
	case applicationFieldHostname:
		var cmd tea.Cmd
		m.hostnameInput, cmd = m.hostnameInput.Update(msg)
		m.settings.Host = m.hostnameInput.Value()
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m SettingsFormApplication) View() string {
	imageLabel := Styles.Label.Render("Image")
	imageField := Styles.Focus(Styles.Input, m.focused == applicationFieldImage).
		Render(m.imageInput.View())

	hostnameLabel := Styles.Label.Render("Hostname")
	hostnameField := Styles.Focus(Styles.Input, m.focused == applicationFieldHostname).
		Render(m.hostnameInput.View())

	tlsLabel := Styles.Label.Render("TLS")
	var tlsText string
	if docker.IsLocalhost(m.settings.Host) {
		tlsText = "Not available for localhost"
	} else if m.settings.TLSEnabled() {
		tlsText = "[x] Enabled"
	} else {
		tlsText = "[ ] Enabled"
	}
	tlsField := Styles.Focus(Styles.Input, m.focused == applicationFieldTLS).
		Render(tlsText)

	doneButton := Styles.Focus(Styles.ButtonPrimary, m.focused == applicationFieldDoneButton).
		Render("Done")
	cancelButton := Styles.Focus(Styles.Button, m.focused == applicationFieldCancelButton).
		Render("Cancel")

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, doneButton, cancelButton)

	form := lipgloss.JoinVertical(lipgloss.Left,
		imageLabel,
		imageField,
		hostnameLabel,
		hostnameField,
		tlsLabel,
		tlsField,
		"",
		buttons,
	)

	return form
}

// Private

func (m SettingsFormApplication) focusNext() (SettingsSection, tea.Cmd) {
	m.blurCurrent()
	m.focused = (m.focused + 1) % applicationFieldCount
	return m.focusCurrent()
}

func (m SettingsFormApplication) focusPrev() (SettingsSection, tea.Cmd) {
	m.blurCurrent()
	m.focused = (m.focused - 1 + applicationFieldCount) % applicationFieldCount
	return m.focusCurrent()
}

func (m *SettingsFormApplication) blurCurrent() {
	switch m.focused {
	case applicationFieldImage:
		m.imageInput.Blur()
	case applicationFieldHostname:
		m.hostnameInput.Blur()
	}
}

func (m SettingsFormApplication) focusCurrent() (SettingsSection, tea.Cmd) {
	var cmd tea.Cmd
	switch m.focused {
	case applicationFieldImage:
		cmd = m.imageInput.Focus()
	case applicationFieldHostname:
		cmd = m.hostnameInput.Focus()
	}
	return m, cmd
}

func (m SettingsFormApplication) handleEnter() (SettingsSection, tea.Cmd) {
	switch m.focused {
	case applicationFieldImage, applicationFieldHostname, applicationFieldTLS:
		return m.focusNext()
	case applicationFieldDoneButton:
		return m, func() tea.Msg { return SettingsSectionSubmitMsg{Settings: m.settings} }
	case applicationFieldCancelButton:
		return m, func() tea.Msg { return SettingsSectionCancelMsg{} }
	}
	return m, nil
}
