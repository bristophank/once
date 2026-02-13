package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUniqueName(t *testing.T) {
	ns := &Namespace{name: "test"}

	assert.Equal(t, "myapp", ns.UniqueName("myapp"))

	ns.AddApplication(ApplicationSettings{Name: "myapp"})
	assert.Equal(t, "myapp.1", ns.UniqueName("myapp"))

	ns.AddApplication(ApplicationSettings{Name: "myapp.1"})
	assert.Equal(t, "myapp.2", ns.UniqueName("myapp"))

	// Unrelated app doesn't affect the name
	assert.Equal(t, "other", ns.UniqueName("other"))
}

func TestContainerAppName(t *testing.T) {
	ns := &Namespace{name: "once"}

	t.Run("standard app", func(t *testing.T) {
		assert.Equal(t, "campfire", ns.containerAppName("once-app-campfire-a1b2c3"))
	})

	t.Run("dotted unique name", func(t *testing.T) {
		assert.Equal(t, "campfire.1", ns.containerAppName("once-app-campfire.1-d4e5f6"))
	})

	t.Run("dashed app name", func(t *testing.T) {
		assert.Equal(t, "my-app", ns.containerAppName("once-app-my-app-abcdef"))
	})

	t.Run("wrong namespace", func(t *testing.T) {
		assert.Equal(t, "", ns.containerAppName("other-app-campfire-a1b2c3"))
	})

	t.Run("not a container name", func(t *testing.T) {
		assert.Equal(t, "", ns.containerAppName("something-else"))
	})

	t.Run("no ID suffix", func(t *testing.T) {
		assert.Equal(t, "", ns.containerAppName("once-app-campfire"))
	})
}
