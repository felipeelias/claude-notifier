package notifier_test

import (
	"testing"

	"github.com/felipeelias/claude-notifier/internal/notifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterAndCreate(t *testing.T) {
	reg := notifier.NewRegistry()
	assert.Empty(t, reg.All())

	err := reg.Register("mock", func() notifier.Notifier {
		return &mockNotifier{name: "mock"}
	})
	require.NoError(t, err)

	factories := reg.All()
	require.Len(t, factories, 1)

	n := factories["mock"]()
	assert.Equal(t, "mock", n.Name())
}

func TestRegisterMultiple(t *testing.T) {
	reg := notifier.NewRegistry()
	require.NoError(t, reg.Register("a", func() notifier.Notifier { return &mockNotifier{name: "a"} }))
	require.NoError(t, reg.Register("b", func() notifier.Notifier { return &mockNotifier{name: "b"} }))
	assert.Len(t, reg.All(), 2)
}

func TestRegisterDuplicate(t *testing.T) {
	reg := notifier.NewRegistry()
	require.NoError(t, reg.Register("a", func() notifier.Notifier { return &mockNotifier{name: "a"} }))
	err := reg.Register("a", func() notifier.Notifier { return &mockNotifier{name: "a"} })
	assert.EqualError(t, err, `notifier "a" already registered`)
}
