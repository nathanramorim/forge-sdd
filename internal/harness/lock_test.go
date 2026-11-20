package harness

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var t0 = time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC)

func TestLock_AdquireRenovaELibera(t *testing.T) {
	dir := t.TempDir()
	l, err := AdquirirLock(dir, "104-x", "a", time.Minute, t0)
	require.NoError(t, err)
	require.NoError(t, l.Renovar(t0.Add(30*time.Second)))
	info, err := lerLock(l.Caminho())
	require.NoError(t, err)
	assert.Equal(t, t0.Add(90*time.Second), info.Expira)
	require.NoError(t, l.Liberar())
	require.NoError(t, l.Liberar(), "liberar duas vezes não é erro")
	_, err = AdquirirLock(dir, "104-x", "b", time.Minute, t0)
	assert.NoError(t, err)
}

func TestLock_SegundoDonoRecusado(t *testing.T) {
	dir := t.TempDir()
	_, err := AdquirirLock(dir, "104-x", "a", time.Minute, t0)
	require.NoError(t, err)
	_, err = AdquirirLock(dir, "104-x", "b", time.Minute, t0.Add(30*time.Second))
	assert.ErrorIs(t, err, ErrLockOcupado)
	_, err = AdquirirLock(dir, "104-outra", "b", time.Minute, t0)
	assert.NoError(t, err, "outra iniciativa não é afetada")
}

func TestLock_OrfaoExpiradoERecuperado(t *testing.T) {
	dir := t.TempDir()
	_, err := AdquirirLock(dir, "104-x", "morto", time.Minute, t0) // processo "morre" sem liberar
	require.NoError(t, err)
	l, err := AdquirirLock(dir, "104-x", "vivo", time.Minute, t0.Add(2*time.Minute))
	require.NoError(t, err)
	info, err := lerLock(l.Caminho())
	require.NoError(t, err)
	assert.Equal(t, "vivo", info.Dono)
	orfaos, _ := filepath.Glob(filepath.Join(dir, "*.orfao.*"))
	assert.Empty(t, orfaos, "o arquivo órfão não pode sobrar")
}

func TestLock_ArquivoCorrompidoETratadoComoOrfao(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "104-x.lock"), []byte("{lixo"), 0o644))
	_, err := AdquirirLock(dir, "104-x", "a", time.Minute, t0)
	assert.NoError(t, err)
}

func TestLock_DonoAntigoNaoLiberaNemRenovaLockTomado(t *testing.T) {
	dir := t.TempDir()
	antigo, err := AdquirirLock(dir, "104-x", "antigo", time.Minute, t0)
	require.NoError(t, err)
	_, err = AdquirirLock(dir, "104-x", "novo", time.Minute, t0.Add(5*time.Minute))
	require.NoError(t, err)
	assert.ErrorIs(t, antigo.Renovar(t0.Add(6*time.Minute)), ErrLockPerdido)
	assert.ErrorIs(t, antigo.Liberar(), ErrLockPerdido)
}

func TestLock_ParametrosInvalidos(t *testing.T) {
	dir := t.TempDir()
	for _, c := range []struct {
		ini, dono string
		lease     time.Duration
	}{
		{"", "a", time.Minute}, {"x", "", time.Minute}, {"x", "a", 0},
	} {
		_, err := AdquirirLock(dir, c.ini, c.dono, c.lease, t0)
		assert.Error(t, err)
	}
}

// Concorrência de verdade: com o lock livre, só uma das goroutines pode vencer; com o lock órfão,
// também só uma recupera.
func TestLock_ConcorrenciaUmSoVence(t *testing.T) {
	for _, orfao := range []bool{false, true} {
		dir := t.TempDir()
		agora := t0
		if orfao {
			_, err := AdquirirLock(dir, "104-x", "morto", time.Minute, t0)
			require.NoError(t, err)
			agora = t0.Add(time.Hour)
		}
		var vencedores, ocupados atomic.Int32
		var wg sync.WaitGroup
		inicio := make(chan struct{})
		for i := 0; i < 32; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-inicio
				_, err := AdquirirLock(dir, "104-x", "g", time.Minute, agora)
				switch {
				case err == nil:
					vencedores.Add(1)
				case errors.Is(err, ErrLockOcupado):
					ocupados.Add(1)
				default:
					t.Errorf("erro inesperado: %v", err)
				}
			}()
		}
		close(inicio)
		wg.Wait()
		assert.EqualValues(t, 1, vencedores.Load(), "órfão=%v: exatamente um vence", orfao)
		assert.EqualValues(t, 31, ocupados.Load())
	}
}
