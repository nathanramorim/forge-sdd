package main

import (
	"context"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/acp"
	"github.com/forge-sdd/cli/internal/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartACPBackground_SanitizaOrfasEEncerraSemTravar(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, queue.RecordActivityDetailed(dir, queue.ActivityItem{
		Timestamp: time.Now().UTC().Add(-time.Minute).Format(time.RFC3339),
		Station:   queue.StationBuilder,
		Tool:      "run_command",
		Action:    "go test",
		Status:    "running",
	}))

	dispatcher := acp.NewDispatcherWithVersion("test")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := dispatcher.Pool()
	require.NotNil(t, pool)
	stop := startACPBackground(ctx, dir, pool)
	require.True(t, pool.HeartbeatMonitor().IsRunning(), "heartbeat deve estar ativo após o boot")

	items, err := queue.LoadActivities(dir)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "success", items[0].Status, "atividade órfã deve ser encerrada no boot do servidor")

	done := make(chan struct{})
	go func() {
		stop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("encerramento do heartbeat travou")
	}
	assert.False(t, pool.HeartbeatMonitor().IsRunning(), "heartbeat deve parar após o encerramento")
}

func TestStartACPBackground_PoolNuloNaoQuebra(t *testing.T) {
	stop := startACPBackground(context.Background(), t.TempDir(), nil)
	require.NotNil(t, stop)
	stop()
}
