package ui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Direction for stack layout
type Direction int

const (
	Vertical Direction = iota
	Horizontal
)

// Size specifications for stack children
type Size interface {
	isSize()
}

type fixedSize int
type fillSize struct{}
type fitSize struct{}
type percentSize int

func (fixedSize) isSize()   {}
func (fillSize) isSize()    {}
func (fitSize) isSize()     {}
func (percentSize) isSize() {}

func Fixed(n int) Size   { return fixedSize(n) }
func Fill() Size         { return fillSize{} }
func Fit() Size          { return fitSize{} }
func Percent(n int) Size { return percentSize(n) }

// ComponentSizeMsg is sent to child components to inform them of their allocated size.
// This is distinct from tea.WindowSizeMsg which represents the terminal window size.
type ComponentSizeMsg struct {
	Width  int
	Height int
}

// Content wraps a render function as a Component for use in layouts.
// The render function receives the allocated width and height.
type Content struct {
	render        func(width, height int) string
	width, height int
}

func NewContent(render func(width, height int) string) Content {
	return Content{render: render}
}

func (c Content) Init() tea.Cmd { return nil }

func (c Content) Update(msg tea.Msg) (Component, tea.Cmd) {
	if msg, ok := msg.(ComponentSizeMsg); ok {
		c.width, c.height = msg.Width, msg.Height
	}
	return c, nil
}

func (c Content) View() string {
	return c.render(c.width, c.height)
}

// StackChild combines a component with its size rule
type StackChild struct {
	Component Component
	Size      Size
}

func WithFixed(size int, c Component) StackChild {
	return StackChild{Component: c, Size: Fixed(size)}
}

func WithFill(c Component) StackChild {
	return StackChild{Component: c, Size: Fill()}
}

func WithPercent(pct int, c Component) StackChild {
	return StackChild{Component: c, Size: Percent(pct)}
}

// StackLayout manages child component sizing and layout
type StackLayout struct {
	direction Direction
	children  []StackChild
	width     int
	height    int
}

func NewStackLayout(direction Direction, children ...StackChild) StackLayout {
	return StackLayout{
		direction: direction,
		children:  children,
	}
}

func (s StackLayout) Init() tea.Cmd {
	cmds := make([]tea.Cmd, len(s.children))
	for i, child := range s.children {
		cmds[i] = child.Component.Init()
	}
	return tea.Batch(cmds...)
}

func (s StackLayout) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width, s.height = msg.Width, msg.Height
		s.resizeChildren(&cmds)
	case ComponentSizeMsg:
		s.width, s.height = msg.Width, msg.Height
		s.resizeChildren(&cmds)
	default:
		for i, child := range s.children {
			var cmd tea.Cmd
			s.children[i].Component, cmd = child.Component.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	return s, tea.Batch(cmds...)
}

func (s StackLayout) View() string {
	views := make([]string, len(s.children))
	for i, child := range s.children {
		views[i] = child.Component.View()
	}
	if s.direction == Vertical {
		return lipgloss.JoinVertical(lipgloss.Left, views...)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, views...)
}

// Private

func (s *StackLayout) resizeChildren(cmds *[]tea.Cmd) {
	sizes := s.calculateSizes()
	for i, child := range s.children {
		var childWidth, childHeight int
		if s.direction == Vertical {
			childWidth, childHeight = s.width, sizes[i]
		} else {
			childWidth, childHeight = sizes[i], s.height
		}
		var cmd tea.Cmd
		s.children[i].Component, cmd = child.Component.Update(ComponentSizeMsg{
			Width:  childWidth,
			Height: childHeight,
		})
		if cmd != nil {
			*cmds = append(*cmds, cmd)
		}
	}
}

func (s *StackLayout) calculateSizes() []int {
	total := s.height
	if s.direction == Horizontal {
		total = s.width
	}

	sizes := make([]int, len(s.children))
	remaining := total
	fillCount := 0

	// First pass: fixed, percent, and fit sizes
	for i, child := range s.children {
		switch sz := child.Size.(type) {
		case fixedSize:
			sizes[i] = int(sz)
			remaining -= sizes[i]
		case percentSize:
			sizes[i] = total * int(sz) / 100
			remaining -= sizes[i]
		case fitSize:
			sizes[i] = s.measureFit(i)
			remaining -= sizes[i]
		case fillSize:
			fillCount++
		}
	}

	// Second pass: distribute remaining space to fills
	if fillCount > 0 && remaining > 0 {
		perFill := remaining / fillCount
		for i, child := range s.children {
			if _, ok := child.Size.(fillSize); ok {
				sizes[i] = perFill
			}
		}
	}

	return sizes
}

func (s *StackLayout) measureFit(i int) int {
	child := s.children[i]

	// Give the component the cross-axis dimension to render with
	var sizeMsg ComponentSizeMsg
	if s.direction == Vertical {
		sizeMsg = ComponentSizeMsg{Width: s.width, Height: 0}
	} else {
		sizeMsg = ComponentSizeMsg{Width: 0, Height: s.height}
	}

	updated, _ := child.Component.Update(sizeMsg)
	s.children[i].Component = updated

	// Render and measure
	view := updated.View()
	if s.direction == Vertical {
		return lipgloss.Height(view)
	}
	return lipgloss.Width(view)
}
