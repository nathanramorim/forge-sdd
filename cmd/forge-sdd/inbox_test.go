package main

import (
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/notify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetInboxFlags() {
	for _, f := range []string{"hook", "include-read", "mark-all", "archive"} {
		_ = inboxCmd.Flags().Set(f, "false")
	}
	_ = inboxCmd.Flags().Set("mark-read", "")
}

func TestInboxCLI_ListaHookEMarcaLida(t *testing.T) {
	dir := t.TempDir()
	resetInboxFlags()
	cfg := config.DefaultNotificationsConfig()
	cfg.Channels.DiscoverySession.DesktopToast = false
	_, err := notify.Dispatch(dir, cfg, notify.Notification{
		Type: notify.TypePRApproved, Title: "PR pronto", Message: "ok", PRURL: "https://github.com/acme/app/pull/7", ID: "n-1", CreatedAt: time.Now(),
	}, notify.Deps{})
	require.NoError(t, err)

	out, err := runRoot(t, "inbox", "--dir", dir)
	require.NoError(t, err)
	assert.Contains(t, out, "https://github.com/acme/app/pull/7")
	assert.Contains(t, out, "pronto para aprovação final")

	resetInboxFlags()
	out, err = runRoot(t, "inbox", "--dir", dir, "--hook")
	require.NoError(t, err)
	assert.Contains(t, out, "1 notificação(ões) não lida(s)")

	resetInboxFlags()
	out, err = runRoot(t, "inbox", "--dir", dir, "--hook")
	require.NoError(t, err)
	assert.Equal(t, "", out, "o hook não repete o que já foi mostrado")

	resetInboxFlags()
	out, err = runRoot(t, "inbox", "--dir", dir)
	require.NoError(t, err)
	assert.Contains(t, out, "Nenhuma notificação não lida")

	resetInboxFlags()
	out, err = runRoot(t, "inbox", "--dir", dir, "--include-read")
	require.NoError(t, err)
	assert.Contains(t, out, "pull/7")

	resetInboxFlags()
	out, err = runRoot(t, "inbox", "--dir", dir, "--archive")
	require.NoError(t, err)
	assert.Contains(t, out, "arquivadas: 1")

	resetInboxFlags()
	_, err = runRoot(t, "inbox", "--dir", dir, "--mark-read", "../x")
	require.Error(t, err)
}
