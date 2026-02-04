package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockComponent struct {
	width, height int
	viewText      string
}

func (m mockComponent) Init() tea.Cmd { return nil }

func (m mockComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	if msg, ok := msg.(ComponentSizeMsg); ok {
		m.width, m.height = msg.Width, msg.Height
	}
	return m, nil
}

func (m mockComponent) View() string {
	if m.viewText != "" {
		return m.viewText
	}
	return "mock"
}

func TestStackLayoutVerticalSizing(t *testing.T) {
	child1 := mockComponent{viewText: "a"}
	child2 := mockComponent{viewText: "b"}
	child3 := mockComponent{viewText: "c"}

	layout := NewStackLayout(Vertical,
		WithFixed(10, child1),
		WithFill(child2),
		WithPercent(20, child3),
	)

	updated, _ := layout.Update(tea.WindowSizeMsg{Width: 80, Height: 100})
	layout = updated.(StackLayout)

	sizes := layout.calculateSizes()

	assert.Equal(t, 10, sizes[0])  // fixed 10
	assert.Equal(t, 70, sizes[1])  // fill gets remaining: 100 - 10 - 20 = 70
	assert.Equal(t, 20, sizes[2])  // 20% of 100
}

func TestStackLayoutHorizontalSizing(t *testing.T) {
	child1 := mockComponent{}
	child2 := mockComponent{}

	layout := NewStackLayout(Horizontal,
		WithPercent(30, child1),
		WithFill(child2),
	)

	updated, _ := layout.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	layout = updated.(StackLayout)

	sizes := layout.calculateSizes()

	assert.Equal(t, 30, sizes[0]) // 30% of 100
	assert.Equal(t, 70, sizes[1]) // fill gets remaining: 100 - 30 = 70
}

func TestStackLayoutChildrenReceiveSize(t *testing.T) {
	child1 := mockComponent{}
	child2 := mockComponent{}

	layout := NewStackLayout(Vertical,
		WithFixed(20, child1),
		WithFill(child2),
	)

	updated, _ := layout.Update(tea.WindowSizeMsg{Width: 80, Height: 100})
	layout = updated.(StackLayout)

	// Children should have received ComponentSizeMsg
	c1 := layout.children[0].Component.(mockComponent)
	c2 := layout.children[1].Component.(mockComponent)

	assert.Equal(t, 80, c1.width)
	assert.Equal(t, 20, c1.height)

	assert.Equal(t, 80, c2.width)
	assert.Equal(t, 80, c2.height) // 100 - 20 = 80
}

func TestStackLayoutComponentSizeMsg(t *testing.T) {
	child := mockComponent{}
	layout := NewStackLayout(Vertical, WithFill(child))

	updated, _ := layout.Update(ComponentSizeMsg{Width: 60, Height: 40})
	layout = updated.(StackLayout)

	assert.Equal(t, 60, layout.width)
	assert.Equal(t, 40, layout.height)

	c := layout.children[0].Component.(mockComponent)
	assert.Equal(t, 60, c.width)
	assert.Equal(t, 40, c.height)
}

func TestStackLayoutMultipleFills(t *testing.T) {
	child1 := mockComponent{}
	child2 := mockComponent{}
	child3 := mockComponent{}

	layout := NewStackLayout(Vertical,
		WithFixed(10, child1),
		WithFill(child2),
		WithFill(child3),
	)

	updated, _ := layout.Update(tea.WindowSizeMsg{Width: 80, Height: 100})
	layout = updated.(StackLayout)

	sizes := layout.calculateSizes()

	assert.Equal(t, 10, sizes[0]) // fixed
	assert.Equal(t, 45, sizes[1]) // (100-10)/2 = 45
	assert.Equal(t, 45, sizes[2]) // (100-10)/2 = 45
}

func TestStackLayoutView(t *testing.T) {
	child1 := mockComponent{viewText: "top"}
	child2 := mockComponent{viewText: "bottom"}

	layout := NewStackLayout(Vertical,
		WithFixed(1, child1),
		WithFill(child2),
	)

	view := layout.View()
	require.Contains(t, view, "top")
	require.Contains(t, view, "bottom")
}

func TestStackLayoutForwardsOtherMessages(t *testing.T) {
	type customMsg struct{ value int }

	receivedMsg := false
	child := &msgTrackingComponent{onMsg: func(msg tea.Msg) {
		if m, ok := msg.(customMsg); ok && m.value == 42 {
			receivedMsg = true
		}
	}}

	layout := NewStackLayout(Vertical, WithFill(child))
	layout.Update(customMsg{value: 42})

	assert.True(t, receivedMsg)
}

type msgTrackingComponent struct {
	onMsg func(tea.Msg)
}

func (m *msgTrackingComponent) Init() tea.Cmd { return nil }

func (m *msgTrackingComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	if m.onMsg != nil {
		m.onMsg(msg)
	}
	return m, nil
}

func (m *msgTrackingComponent) View() string { return "" }

func TestStackLayoutFitSize(t *testing.T) {
	// Content that renders to 3 lines
	content := NewContent(func(w, h int) string {
		return "line1\nline2\nline3"
	})

	layout := NewStackLayout(Vertical,
		StackChild{Component: content, Size: Fit()},
		WithFill(mockComponent{viewText: "fill"}),
	)

	updated, _ := layout.Update(ComponentSizeMsg{Width: 80, Height: 100})
	layout = updated.(StackLayout)

	sizes := layout.calculateSizes()

	assert.Equal(t, 3, sizes[0])  // fit: 3 lines
	assert.Equal(t, 97, sizes[1]) // fill: 100 - 3 = 97
}

func TestContentComponent(t *testing.T) {
	renderCalled := false
	content := NewContent(func(w, h int) string {
		renderCalled = true
		return "hello"
	})

	// Update with size
	updated, _ := content.Update(ComponentSizeMsg{Width: 40, Height: 10})
	content = updated.(Content)

	assert.Equal(t, 40, content.width)
	assert.Equal(t, 10, content.height)

	// View should call render
	view := content.View()
	assert.True(t, renderCalled)
	assert.Equal(t, "hello", view)
}
