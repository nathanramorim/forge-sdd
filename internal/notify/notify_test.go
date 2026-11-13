package notify

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fixedNow = time.Date(2026, 11, 13, 15, 0, 0, 0, time.UTC)

func baseDeps(env map[string]string) (Deps, *[][]string) {
	var calls [][]string
	var mu sync.Mutex
	return Deps{
		Getenv: func(k string) string { return env[k] },
		GOOS:   "linux",
		RunToast: func(name string, args ...string) error {
			mu.Lock()
			defer mu.Unlock()
			calls = append(calls, append([]string{name}, args...))
			return nil
		},
		Now: func() time.Time { return fixedNow },
	}, &calls
}

func approved() Notification {
	return Notification{
		Type: TypePRApproved, Title: "PR pronto: auth-oauth", Feature: "auth-oauth", JobID: "job-1",
		Message: "PR revisado, sem inconformidades e pronto para aprovação final.",
		PRURL:   "https://github.com/acme/app/pull/12", CardURL: "https://app.clickup.com/t/8421",
	}
}

func TestDispatch_GravaInboxETostEApenasCanaisHabilitados(t *testing.T) {
	dir := t.TempDir()
	deps, toasts := baseDeps(nil)
	cfg := config.DefaultNotificationsConfig()

	res, err := Dispatch(dir, cfg, approved(), deps)
	require.NoError(t, err)
	assert.FileExists(t, res.InboxPath)
	assert.Equal(t, layout.Inbox(dir), filepath.Dir(res.InboxPath))
	assert.Equal(t, "ok", res.Channels["discovery_session"])
	assert.Equal(t, "ok", res.Channels["desktop_toast"])
	assert.Equal(t, "desligado", res.Channels["email"])
	assert.Equal(t, "desligado", res.Channels["webhook"])
	require.Len(t, *toasts, 1)
	assert.Equal(t, "notify-send", (*toasts)[0][0])

	items, err := ListInbox(dir, true)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "https://github.com/acme/app/pull/12", items[0].PRURL)
	assert.False(t, items[0].Read)
}

func TestDispatch_ToastDesligadoNaoDispara(t *testing.T) {
	dir := t.TempDir()
	deps, toasts := baseDeps(nil)
	cfg := config.DefaultNotificationsConfig()
	cfg.Channels.DiscoverySession.DesktopToast = false
	res, err := Dispatch(dir, cfg, approved(), deps)
	require.NoError(t, err)
	assert.Empty(t, *toasts)
	assert.Equal(t, "desligado", res.Channels["desktop_toast"])
}

func TestToast_ComandoPorSistemaOperacional(t *testing.T) {
	for goos, wantCmd := range map[string]string{"darwin": "osascript", "linux": "notify-send", "windows": "powershell"} {
		dir := t.TempDir()
		deps, toasts := baseDeps(nil)
		deps.GOOS = goos
		_, err := Dispatch(dir, config.DefaultNotificationsConfig(), approved(), deps)
		require.NoError(t, err, goos)
		require.Len(t, *toasts, 1, goos)
		assert.Equal(t, wantCmd, (*toasts)[0][0], goos)
	}
	dir := t.TempDir()
	deps, _ := baseDeps(nil)
	deps.GOOS = "plan9"
	res, err := Dispatch(dir, config.DefaultNotificationsConfig(), approved(), deps)
	require.NoError(t, err, "SO sem toast não bloqueia o inbox")
	assert.Contains(t, res.Channels["desktop_toast"], "falhou")
}

func TestToast_TextoComAspasENaoViraComando(t *testing.T) {
	dir := t.TempDir()
	deps, toasts := baseDeps(nil)
	deps.GOOS = "darwin"
	n := approved()
	n.Title = `PR "x"; do shell script "rm -rf ~"`
	_, err := Dispatch(dir, config.DefaultNotificationsConfig(), n, deps)
	require.NoError(t, err)
	script := (*toasts)[0][2]
	assert.Contains(t, script, `\"x\"`, "aspas escapadas no AppleScript")
	assert.Equal(t, "-e", (*toasts)[0][1], "o texto vai como um único argumento")
}

func TestWebhook_SlackEGenericoPorVariavelDeAmbiente(t *testing.T) {
	var mu sync.Mutex
	var bodies []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var m map[string]any
		_ = json.NewDecoder(r.Body).Decode(&m)
		mu.Lock()
		bodies = append(bodies, m)
		mu.Unlock()
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(200)
	}))
	defer srv.Close()

	for _, provider := range []string{"slack", "generic"} {
		dir := t.TempDir()
		deps, _ := baseDeps(map[string]string{"FORGE_WEBHOOK_URL": srv.URL + "/segredo-do-hook"})
		cfg := config.DefaultNotificationsConfig()
		cfg.Channels.Webhook = config.WebhookChannel{Enabled: true, Provider: provider, URLEnv: "FORGE_WEBHOOK_URL"}
		res, err := Dispatch(dir, cfg, approved(), deps)
		require.NoError(t, err)
		assert.Equal(t, "ok", res.Channels["webhook"], provider)
	}
	mu.Lock()
	defer mu.Unlock()
	require.Len(t, bodies, 2)
	assert.Contains(t, bodies[0]["text"], "PR pronto: auth-oauth")
	assert.Contains(t, bodies[0]["text"], "pull/12")
	assert.Equal(t, "pr_approved", bodies[1]["type"])
}

func TestWebhook_Violacoes(t *testing.T) {
	srvErr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) }))
	defer srvErr.Close()

	cases := map[string]string{
		"variável ausente":  "",
		"http fora de loop": "http://exemplo.com/hook",
		"url inválida":      "::::",
		"esquema ftp":       "ftp://x/y",
		"resposta 500":      srvErr.URL + "/token-secreto",
	}
	for name, url := range cases {
		dir := t.TempDir()
		deps, _ := baseDeps(map[string]string{"FORGE_WEBHOOK_URL": url})
		cfg := config.DefaultNotificationsConfig()
		cfg.Channels.Webhook = config.WebhookChannel{Enabled: true, Provider: "slack", URLEnv: "FORGE_WEBHOOK_URL"}
		res, err := Dispatch(dir, cfg, approved(), deps)
		require.NoError(t, err, "%s: falha de canal opcional não é erro do Dispatch", name)
		assert.Contains(t, res.Channels["webhook"], "falhou", name)
		assert.NotContains(t, res.Channels["webhook"], "token-secreto", "%s: a URL não vaza na mensagem", name)
		assert.FileExists(t, res.InboxPath, "%s: o inbox é gravado mesmo assim", name)
	}
}

// smtpTestServer é um servidor SMTP mínimo em processo.
func smtpTestServer(t *testing.T) (addr string, received func() string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })
	var mu sync.Mutex
	var data strings.Builder
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				rd := bufio.NewReader(c)
				w := func(s string) { _, _ = io.WriteString(c, s+"\r\n") }
				w("220 teste ESMTP")
				inData := false
				for {
					line, err := rd.ReadString('\n')
					if err != nil {
						return
					}
					line = strings.TrimRight(line, "\r\n")
					if inData {
						if line == "." {
							inData = false
							w("250 ok")
							continue
						}
						mu.Lock()
						data.WriteString(line + "\n")
						mu.Unlock()
						continue
					}
					switch up := strings.ToUpper(line); {
					case strings.HasPrefix(up, "EHLO"), strings.HasPrefix(up, "HELO"):
						w("250 teste")
					case strings.HasPrefix(up, "MAIL"), strings.HasPrefix(up, "RCPT"):
						w("250 ok")
					case up == "DATA":
						inData = true
						w("354 fim com .")
					case up == "QUIT":
						w("221 tchau")
						return
					default:
						w("250 ok")
					}
				}
			}(c)
		}
	}()
	return ln.Addr().String(), func() string { mu.Lock(); defer mu.Unlock(); return data.String() }
}

func TestEmail_EnviaPorSMTPDeTeste(t *testing.T) {
	addr, received := smtpTestServer(t)
	host, port, _ := net.SplitHostPort(addr)
	var p int
	_, _ = fmt.Sscanf(port, "%d", &p)

	dir := t.TempDir()
	deps, _ := baseDeps(nil)
	cfg := config.DefaultNotificationsConfig()
	cfg.Channels.Email = config.EmailChannel{Enabled: true, Provider: "smtp", Host: host, Port: p, To: []string{"time@exemplo.com"}, From: "forge@exemplo.com"}
	res, err := Dispatch(dir, cfg, approved(), deps)
	require.NoError(t, err)
	assert.Equal(t, "ok", res.Channels["email"])
	assert.Eventually(t, func() bool { return strings.Contains(received(), "Subject: [Forge] PR pronto: auth-oauth") }, 2*time.Second, 20*time.Millisecond)
	assert.Contains(t, received(), "PR: https://github.com/acme/app/pull/12")
}

func TestEmail_FalhaNaoBloqueiaInboxNemWebhook(t *testing.T) {
	var hits int
	var mu sync.Mutex
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { mu.Lock(); hits++; mu.Unlock(); w.WriteHeader(200) }))
	defer srv.Close()

	dir := t.TempDir()
	deps, _ := baseDeps(map[string]string{"FORGE_WEBHOOK_URL": srv.URL, "SMTP_USER": "joao", "SMTP_PASS": "senha-super-secreta"})
	deps.SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		return errors.New("535 credenciais joao/senha-super-secreta recusadas por smtp.interno")
	}
	cfg := config.DefaultNotificationsConfig()
	cfg.Channels.Email = config.EmailChannel{Enabled: true, Provider: "smtp", Host: "smtp.interno", Port: 587, UserEnv: "SMTP_USER", PasswordEnv: "SMTP_PASS", To: []string{"a@b.com"}, From: "x@y.com"}
	cfg.Channels.Webhook = config.WebhookChannel{Enabled: true, Provider: "slack", URLEnv: "FORGE_WEBHOOK_URL"}

	res, err := Dispatch(dir, cfg, approved(), deps)
	require.NoError(t, err)
	assert.Contains(t, res.Channels["email"], "falhou")
	assert.NotContains(t, res.Channels["email"], "senha-super-secreta")
	assert.NotContains(t, res.Channels["email"], "smtp.interno")
	assert.Equal(t, "ok", res.Channels["webhook"])
	mu.Lock()
	assert.Equal(t, 1, hits)
	mu.Unlock()
	assert.FileExists(t, res.InboxPath)
}

func TestEmail_CredencialAusenteEhFalhaDoCanal(t *testing.T) {
	dir := t.TempDir()
	deps, _ := baseDeps(nil)
	deps.SendMail = func(string, smtp.Auth, string, []string, []byte) error {
		t.Fatal("não deve enviar sem credencial")
		return nil
	}
	cfg := config.DefaultNotificationsConfig()
	cfg.Channels.Email = config.EmailChannel{Enabled: true, Provider: "smtp", Host: "h", Port: 25, UserEnv: "SMTP_USER", PasswordEnv: "SMTP_PASS", To: []string{"a@b.com"}, From: "x@y.com"}
	res, err := Dispatch(dir, cfg, approved(), deps)
	require.NoError(t, err)
	assert.Contains(t, res.Channels["email"], "SMTP_USER")
}

func TestNenhumSegredoEmArquivosNemSaidas(t *testing.T) {
	const secret = "s3cr3t-token-xyz"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
	defer srv.Close()
	dir := t.TempDir()
	deps, _ := baseDeps(map[string]string{"FORGE_WEBHOOK_URL": srv.URL + "/" + secret, "SMTP_PASS": secret, "SMTP_USER": "u"})
	deps.SendMail = func(string, smtp.Auth, string, []string, []byte) error { return errors.New(secret) }
	cfg := config.DefaultNotificationsConfig()
	cfg.Channels.Email = config.EmailChannel{Enabled: true, Provider: "smtp", Host: "h", Port: 25, UserEnv: "SMTP_USER", PasswordEnv: "SMTP_PASS", To: []string{"a@b.com"}, From: "x@y.com"}
	cfg.Channels.Webhook = config.WebhookChannel{Enabled: true, Provider: "slack", URLEnv: "FORGE_WEBHOOK_URL"}
	res, err := Dispatch(dir, cfg, approved(), deps)
	require.NoError(t, err)
	assert.NotContains(t, fmt.Sprint(res.Channels), secret)
	require.NoError(t, filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		data, _ := os.ReadFile(p)
		assert.NotContains(t, string(data), secret, p)
		return nil
	}))
}

func TestSanitize_RemoveControleAnsiEBidi(t *testing.T) {
	in := "ok\x1b[31mvermelho\x1b[0m\x1b]0;titulo\x07fim‮\r\nlinha2\u009b"
	out := Sanitize(in, 500, false)
	for _, r := range out {
		assert.False(t, r < 0x20 || (r >= 0x80 && r <= 0x9f) || r == 0x202e, "%q", r)
	}
	assert.Contains(t, out, "vermelho")
	assert.NotContains(t, out, "titulo")
	assert.Equal(t, 10, len(Sanitize(strings.Repeat("ab", 20), 10, false)))
	assert.Contains(t, Sanitize("a\nb", 10, true), "\n")
}

func TestDispatch_Violacoes(t *testing.T) {
	deps, _ := baseDeps(nil)
	cfg := config.DefaultNotificationsConfig()
	for name, n := range map[string]Notification{
		"tipo inválido":      {Type: "spam", Title: "x", Message: "y"},
		"sem título":         {Type: TypeInfo, Message: "y"},
		"sem mensagem":       {Type: TypeInfo, Title: "x"},
		"id com ..":          {Type: TypeInfo, Title: "x", Message: "y", ID: "../fora"},
		"id com barra":       {Type: TypeInfo, Title: "x", Message: "y", ID: "a/b"},
		"mensagem só de ESC": {Type: TypeInfo, Title: "x", Message: "\x1b[2J\x1b[H"},
	} {
		dir := t.TempDir()
		_, err := Dispatch(dir, cfg, n, deps)
		require.Error(t, err, name)
		assert.NoDirExists(t, layout.Inbox(dir), name)
	}

	// Mensagem acima do limite é truncada, não rejeitada.
	dir := t.TempDir()
	n := approved()
	n.Message = strings.Repeat("a", 5000)
	res, err := Dispatch(dir, cfg, n, deps)
	require.NoError(t, err)
	items, _ := ListInbox(dir, false)
	require.Len(t, items, 1)
	assert.LessOrEqual(t, len(items[0].Message), maxMessageBytes)
	_ = res
}

func TestInbox_ListaMarcaLidaEArquiva(t *testing.T) {
	dir := t.TempDir()
	deps, _ := baseDeps(nil)
	cfg := config.DefaultNotificationsConfig()
	cfg.Channels.DiscoverySession.DesktopToast = false

	first := approved()
	first.ID = "n-001"
	first.CreatedAt = fixedNow
	second := Notification{Type: TypeEscalation, Title: "Intervenção necessária: auth-oauth", Message: "3 tentativas esgotadas", PRURL: "https://x/pull/1", ID: "n-002", CreatedAt: fixedNow.Add(time.Minute)}
	for _, n := range []Notification{second, first} {
		_, err := Dispatch(dir, cfg, n, deps)
		require.NoError(t, err)
	}

	items, err := ListInbox(dir, true)
	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, "n-001", items[0].ID, "ordem cronológica")
	text := FormatInbox(items, false)
	assert.Contains(t, text, "PR revisado, sem inconformidades e pronto para aprovação final: https://github.com/acme/app/pull/12")
	assert.Contains(t, text, "intervenção humana necessária")

	require.NoError(t, MarkRead(dir, "n-001"))
	unread, _ := ListInbox(dir, true)
	require.Len(t, unread, 1)
	assert.Equal(t, "n-002", unread[0].ID)
	all, _ := ListInbox(dir, false)
	assert.Len(t, all, 2)

	n, err := ArchiveRead(dir)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.FileExists(t, filepath.Join(layout.InboxArchive(dir), "n-001.json"))
	left, _ := ListInbox(dir, false)
	assert.Len(t, left, 1)

	cnt, err := MarkAllRead(dir)
	require.NoError(t, err)
	assert.Equal(t, 1, cnt)
	assert.Error(t, MarkRead(dir, "nao-existe"))
	assert.Error(t, MarkRead(dir, "../x"))
}

func TestInbox_ArquivoAdulteradoNaoInjetaTerminal(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(layout.Inbox(dir), 0o755))
	evil := `{"id":"evil-1","type":"info","title":"\u001b[2J limpo","message":"linha\u001b]0;pwned\u0007","pr_url":"https://x/\u001b[31m","created_at":"2026-11-13T10:00:00Z","read":false}`
	require.NoError(t, os.WriteFile(filepath.Join(layout.Inbox(dir), "evil-1.json"), []byte(evil), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(layout.Inbox(dir), "quebrado.json"), []byte("{nao"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(layout.Inbox(dir), "grande.json"), []byte(strings.Repeat("a", 70*1024)), 0o644))

	items, err := ListInbox(dir, true)
	require.NoError(t, err)
	require.Len(t, items, 1, "arquivo corrompido e arquivo grande são ignorados")
	out := FormatInbox(items, false)
	assert.NotContains(t, out, "\x1b")
	assert.NotContains(t, out, "pwned")
}

func TestPMComment_GravaInstrucaoParaOAgente(t *testing.T) {
	dir := t.TempDir()
	deps, _ := baseDeps(nil)
	cfg := config.DefaultNotificationsConfig()
	cfg.Channels.DiscoverySession.DesktopToast = false
	cfg.Channels.PMComment.Enabled = true
	res, err := Dispatch(dir, cfg, approved(), deps)
	require.NoError(t, err)
	assert.Contains(t, res.Channels["pm_comment"], "instrução gravada")
	items, _ := ListInbox(dir, true)
	require.Len(t, items, 1)
	assert.Contains(t, items[0].PMComment, "https://app.clickup.com/t/8421")
	assert.Contains(t, FormatInbox(items, false), "Instrução ao agente")
}

func TestHookFormato_CompactoESemNada(t *testing.T) {
	assert.Equal(t, "", FormatInbox(nil, true))
	items := []InboxItem{{Notification: Notification{ID: "a", Type: TypePRApproved, Title: "t", PRURL: "https://x/pull/3", Feature: "f"}}}
	out := FormatInbox(items, true)
	assert.Contains(t, out, "1 notificação(ões) não lida(s)")
	assert.Contains(t, out, "https://x/pull/3")
}

func TestPsSingle_DuplicaAspasTipograficas(t *testing.T) {
	assert.Equal(t, "a''b''c''d''e", psSingle("a'b\u2018c\u2019d\u201ae"))
	assert.Equal(t, "x''", psSingle("x\u201b"))
}
