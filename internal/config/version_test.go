package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersaoDoBinario(t *testing.T) {
	t.Cleanup(func() { SetBinaryVersion("") })

	SetBinaryVersion("")
	assert.Equal(t, "dev", BinaryVersion())

	SetBinaryVersion("9.9.9-teste")
	assert.Equal(t, "9.9.9-teste", BinaryVersion())
	assert.Equal(t, "9.9.9-teste", Defaults().SddVersion)
}
