package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Negativo: a ajuda do raiz e do attach não voltam a vender "5 estações" como fluxo.
func TestLongSemCincoEstacoes(t *testing.T) {
	assert.NotContains(t, rootCmd.Short+rootCmd.Long, "5 estações")
	assert.Contains(t, rootCmd.Long, "3 sessões")
	assert.NotContains(t, rootCmd.Long, "pronta para uso com")
	assert.NotContains(t, attachCmd.Long, "5 estações canônicas do")
	assert.Contains(t, attachCmd.Long, "legadas")
	assert.Contains(t, attachCmd.Long, "3 sessões")
}

func TestRootCommandUse(t *testing.T) {
	assert.Equal(t, "forge", rootCmd.Use, "rootCmd.Use deve ser 'forge'")
	usage := rootCmd.UsageString()
	assert.Contains(t, usage, "forge [command]", "Usage deve informar 'forge [command]'")
}
