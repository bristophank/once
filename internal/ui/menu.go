package ui

import (
	"strconv"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/basecamp/once/internal/mouse"
)

var menuKeys = struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
}{
	Up:     NewKeyBinding("up", "k"),
	Down:   NewKeyBinding("down", "j"),
	Select: NewKeyBinding("enter"),
}

type MenuItem struct {
	Label    string
	Key      int
	Shortcut key.Binding
}

type MenuSelectMsg struct{ Key int }

type Menu struct {
	items    []MenuItem
	selected int
	padWidth int
}

func NewMenu(items ...MenuItem) Menu {
	m := Menu{
		items: items,
	}
	m.measureItems()
	return m
}

func (m *Menu) Update(msg tea.Msg) tea.Cmd {
	count := len(m.items)
	if count == 0 {
		return nil
	}

	switch msg := msg.(type) {
	case MouseEvent:
		if msg.IsClick {
			for i, item := range m.items {
				if msg.Target == menuItemTarget(i) {
					m.selected = i
					return m.selectItem(item.Key)
				}
			}
		}

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, menuKeys.Up):
			m.selected = (m.selected - 1 + count) % count
		case key.Matches(msg, menuKeys.Down):
			m.selected = (m.selected + 1) % count
		case key.Matches(msg, menuKeys.Select):
			return m.selectItem(m.items[m.selected].Key)
		default:
			for i, item := range m.items {
				if key.Matches(msg, item.Shortcut) {
					m.selected = i
					return m.selectItem(item.Key)
				}
			}
		}
	}

	return nil
}

func (m *Menu) View() string {
	itemStyle := lipgloss.NewStyle()
	selectedStyle := lipgloss.NewStyle().Reverse(true)
	keyStyle := lipgloss.NewStyle().Foreground(Colors.Border)

	lines := make([]string, len(m.items))
	for i, item := range m.items {
		padding := strings.Repeat(" ", m.padWidth-len(item.Label))
		shortcutStr := item.Shortcut.Help().Key
		styledKey := keyStyle.Render(shortcutStr)

		var line string
		if m.selected == i {
			line = selectedStyle.Render(item.Label) + padding + styledKey
		} else {
			line = itemStyle.Render(item.Label) + padding + styledKey
		}
		lines[i] = mouse.Mark(menuItemTarget(i), line)
	}

	return strings.Join(lines, "\n")
}

// Private

func (m *Menu) measureItems() {
	maxLen := 0
	for _, item := range m.items {
		if len(item.Label) > maxLen {
			maxLen = len(item.Label)
		}
	}
	m.padWidth = maxLen + 2
}

func (m *Menu) selectItem(key int) tea.Cmd {
	return func() tea.Msg { return MenuSelectMsg{Key: key} }
}

// Helpers

func menuItemTarget(i int) string {
	return "menu-item:" + strconv.Itoa(i)
}
