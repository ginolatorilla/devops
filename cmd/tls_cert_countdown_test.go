package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTlsCertCountdown(t *testing.T) {
	t.Run("Expired", func(t *testing.T) {
		t.Parallel()
		cmd := newTlsCertCountdownCmd()
		cmd.SetIn(strings.NewReader(`notAfter=2024-10-22 00:00:00Z`))

		err := cmd.Execute()

		assert.NoError(t, err)
	})

	t.Run("Valid", func(t *testing.T) {
		t.Parallel()
		cmd := newTlsCertCountdownCmd()
		cmd.SetIn(strings.NewReader(`notAfter=2034-10-22 00:00:00Z`))

		err := cmd.Execute()

		assert.NoError(t, err)
	})
}

func TestTlsCertCountdown_PanicIfExpDateMissing(t *testing.T) {
	t.Parallel()
	cmd := newTlsCertCountdownCmd()
	cmd.SetIn(strings.NewReader(``))

	assert.Panics(t, func() { cmd.Execute() })
}
