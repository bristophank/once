package ui

import (
	"fmt"
	"image/color"
	"math/rand/v2"
	"time"

	tea "charm.land/bubbletea/v2"
)

type ProgressBusy struct {
	Width int
	Color color.Color

	pattern []rune
}

type ProgressBusyTickMsg struct{}

func NewProgressBusy(width int, clr color.Color) *ProgressBusy {
	return &ProgressBusy{
		Width:   width,
		Color:   clr,
		pattern: generateBraillePattern(width),
	}
}

func (p *ProgressBusy) Init() tea.Cmd {
	if p == nil {
		return nil
	}
	return p.tick()
}

func (p *ProgressBusy) Update(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case ProgressBusyTickMsg:
		p.pattern = generateBraillePattern(p.Width)
		return p.tick()
	}
	return nil
}

func (p *ProgressBusy) View() string {
	if p == nil || p.Width <= 0 {
		return ""
	}

	return colorToANSI(p.Color) + string(p.pattern) + "\x1b[0m"
}

// Private

func (p *ProgressBusy) tick() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(time.Time) tea.Msg {
		return ProgressBusyTickMsg{}
	})
}

// Helpers

func generateBraillePattern(width int) []rune {
	pattern := make([]rune, width)
	for i := range pattern {
		// Braille patterns: U+2800 to U+28FF (256 patterns)
		pattern[i] = rune(0x2800 + rand.IntN(256))
	}
	return pattern
}

func colorToANSI(c color.Color) string {
	if c == nil {
		return ""
	}
	r, g, b, _ := c.RGBA()
	r8, g8, b8 := r>>8, g>>8, b>>8
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r8, g8, b8)
}
