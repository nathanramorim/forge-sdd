package layout

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Branch por feature (feat/<N>-<slug>) acha a pasta da iniciativa de mesmo número (task 0.7b).
func TestCurrentResolveIniciativaPorPrefixo(t *testing.T) {
	dir := repoComIniciativas(t, "feat/106-07b-estado-por-feature-layout", "106-pr-draft-revisao-e-ci", "10-outra", "107-x")
	id, ok, err := Current(dir)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "106-pr-draft-revisao-e-ci", id)

	// Branch da própria iniciativa e sem pasta correspondente seguem como antes.
	exact := repoComIniciativas(t, "feat/69-teste-de-layout")
	id, ok, err = Current(exact)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "69-teste-de-layout", id)

	// Prefixo ambíguo é erro, não escolha silenciosa.
	amb := repoComIniciativas(t, "feat/106-07b-x", "106-a", "106-b")
	_, _, err = Current(amb)
	assert.ErrorIs(t, err, ErrAmbiguousInitiative)

	// Branch fora do padrão não aponta para iniciativa.
	main := repoComIniciativas(t, "main", "106-a")
	_, ok, err = Current(main)
	require.NoError(t, err)
	assert.False(t, ok)
}

// Na branch da feature a fila e o status são os da iniciativa (a spec veio da principal), nunca a fila legada.
func TestBranchDeFeatureLeAFilaDaPrincipal(t *testing.T) {
	dir := repoComIniciativas(t, "feat/106-07b-estado-por-feature-layout", "106-pr-draft-revisao-e-ci")
	base := filepath.Join(Discovery(dir), "106-pr-draft-revisao-e-ci")
	assert.Equal(t, filepath.Join(base, QueueFileName), QueuePath(dir))
	assert.NotEqual(t, LegacyQueueFile(dir), QueuePath(dir))
	sp, ok := StatusPath(dir)
	require.True(t, ok)
	assert.Equal(t, filepath.Join(base, StatusFileName), sp)
}
