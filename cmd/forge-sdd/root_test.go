package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCommandUse(t *testing.T) {
	assert.Equal(t, "forge", rootCmd.Use, "rootCmd.Use deve ser 'forge'")
	usage := rootCmd.UsageString()
	assert.Contains(t, usage, "forge [command]", "Usage deve informar 'forge [command]'")
}
