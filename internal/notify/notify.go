// Package notify entrega as notificações do pipeline: o inbox da sessão Discovery (canal padrão
// obrigatório), toast do sistema e os canais opcionais de e-mail e webhook. Segredos vêm só de
// variáveis de ambiente e nunca aparecem em arquivos, mensagens de erro ou logs (SEG-04, SEG-05).
package notify

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/queue"
)

// Tipos de notificação.
const (
	TypePRApproved = "pr_approved"
	TypeEscalation = "escalation"
	TypeInfo       = "info"

	maxMessageBytes = 2048
	maxTitleBytes   = 200
	maxFieldBytes   = 512
	httpTimeout     = 10 * time.Second
)

var (
	idRe        = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,80}$`)
	ansiRe      = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]|\x1b\][^\x07\x1b]*(\x07|\x1b\\)|\x1b[@-_]`)
	errInvalid  = errors.New("notificação inválida")
	validTypes  = map[string]bool{TypePRApproved: true, TypeEscalation: true, TypeInfo: true}
	loopbackIPs = map[string]bool{"localhost": true, "127.0.0.1": true, "::1": true}
)

// Notification é o conteúdo entregue em todos os canais e gravado no inbox.
type Notification struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	PRURL     string    `json:"pr_url,omitempty"`
	CardURL   string    `json:"card_url,omitempty"`
	Feature   string    `json:"feature,omitempty"`
	JobID     string    `json:"job_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	Read      bool      `json:"read"`
	// PMComment é a instrução ao agente (que alcança o MCP) para comentar no card; o Go não executa.
	PMComment string `json:"pm_comment_request,omitempty"`
}

// Deps injeta os efeitos externos para os testes (TST-04).
type Deps struct {
	Getenv    func(string) string
	GOOS      string
	RunToast  func(name string, args ...string) error
	HTTP      *http.Client
	SendMail  func(addr string, a smtp.Auth, from string, to []string, msg []byte) error
	Now       func() time.Time
	AllowHTTP bool // só testes: permite http fora de loopback
}

// DefaultDeps usa os efeitos reais.
func DefaultDeps() Deps {
	return Deps{
		Getenv: os.Getenv,
		GOOS:   runtime.GOOS,
		RunToast: func(name string, args ...string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return exec.CommandContext(ctx, name, args...).Run()
		},
		HTTP:     &http.Client{Timeout: httpTimeout},
		SendMail: smtp.SendMail,
		Now:      time.Now,
	}
}

func (d Deps) withDefaults() Deps {
	def := DefaultDeps()
	if d.Getenv == nil {
		d.Getenv = def.Getenv
	}
	if d.GOOS == "" {
		d.GOOS = def.GOOS
	}
	if d.RunToast == nil {
		d.RunToast = def.RunToast
	}
	if d.HTTP == nil {
		d.HTTP = def.HTTP
	}
	if d.SendMail == nil {
		d.SendMail = def.SendMail
	}
	if d.Now == nil {
		d.Now = def.Now
	}
	return d
}

// Sanitize remove sequências ANSI e caracteres de controle e limita o tamanho (SEG-01).
func Sanitize(s string, max int, keepNewlines bool) string {
	s = ansiRe.ReplaceAllString(s, "")
	s = strings.Map(func(r rune) rune {
		switch {
		case r == '\n' || r == '\t':
			if keepNewlines {
				return r
			}
			return ' '
		case r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f):
			return ' '
		case (r >= 0x202a && r <= 0x202e) || (r >= 0x2066 && r <= 0x2069) || r == 0x200e || r == 0x200f:
			return ' '
		}
		return r
	}, s)
	s = strings.TrimSpace(s)
	if len(s) > max {
		s = strings.ToValidUTF8(s[:max], "")
	}
	return s
}

func (n *Notification) normalize(now time.Time) error {
	if !validTypes[n.Type] {
		return fmt.Errorf("%w: tipo %q", errInvalid, n.Type)
	}
	n.Title = Sanitize(n.Title, maxTitleBytes, false)
	n.Message = Sanitize(n.Message, maxMessageBytes, true)
	n.PRURL = Sanitize(n.PRURL, maxFieldBytes, false)
	n.CardURL = Sanitize(n.CardURL, maxFieldBytes, false)
	n.Feature = Sanitize(n.Feature, 100, false)
	n.JobID = Sanitize(n.JobID, 100, false)
	n.PMComment = Sanitize(n.PMComment, maxMessageBytes, true)
	if n.Title == "" || n.Message == "" {
		return fmt.Errorf("%w: título e mensagem são obrigatórios", errInvalid)
	}
	if n.CreatedAt.IsZero() {
		n.CreatedAt = now.UTC()
	}
	if n.ID == "" {
		sum := sha256.Sum256([]byte(n.Type + "|" + n.JobID + "|" + n.Title + "|" + strconv.FormatInt(n.CreatedAt.UnixNano(), 10)))
		n.ID = n.CreatedAt.Format("20060102T150405") + "-" + hex.EncodeToString(sum[:4])
	}
	if !idRe.MatchString(strings.ToLower(n.ID)) {
		return fmt.Errorf("%w: id %q", errInvalid, n.ID)
	}
	n.ID = strings.ToLower(n.ID)
	return nil
}

// Result descreve o que cada canal fez, sem expor segredos.
type Result struct {
	InboxPath string
	Channels  map[string]string // canal -> "ok", "desligado" ou "falhou: <motivo>"
}

// Dispatch grava no inbox (sempre) e aciona os canais habilitados. Falha de canal opcional nunca
// impede os demais nem a gravação no inbox. Só a falha do inbox é erro.
func Dispatch(dir string, cfg config.NotificationsConfig, n Notification, deps Deps) (Result, error) {
	deps = deps.withDefaults()
	res := Result{Channels: map[string]string{}}
	if err := n.normalize(deps.Now()); err != nil {
		return res, err
	}
	if cfg.Channels.PMComment.Enabled {
		n.PMComment = pmCommentText(n)
	}

	path, err := writeInbox(dir, n)
	if err != nil {
		return res, err
	}
	res.InboxPath = path
	res.Channels["discovery_session"] = "ok"

	if cfg.Channels.DiscoverySession.DesktopToast {
		res.Channels["desktop_toast"] = status(toast(n, deps))
	} else {
		res.Channels["desktop_toast"] = "desligado"
	}
	if cfg.Channels.Email.Enabled {
		res.Channels["email"] = status(sendEmail(cfg.Channels.Email, n, deps))
	} else {
		res.Channels["email"] = "desligado"
	}
	if cfg.Channels.Webhook.Enabled {
		res.Channels["webhook"] = status(sendWebhook(cfg.Channels.Webhook, n, deps))
	} else {
		res.Channels["webhook"] = "desligado"
	}
	if cfg.Channels.PMComment.Enabled {
		res.Channels["pm_comment"] = "ok (instrução gravada para o agente)"
	} else {
		res.Channels["pm_comment"] = "desligado"
	}
	return res, nil
}

func status(err error) string {
	if err == nil {
		return "ok"
	}
	return "falhou: " + Sanitize(err.Error(), 200, false)
}

func pmCommentText(n Notification) string {
	switch n.Type {
	case TypePRApproved:
		return fmt.Sprintf("Comente no card %s: PR revisado, sem inconformidades, pronto para aprovação final: %s", n.CardURL, n.PRURL)
	case TypeEscalation:
		return fmt.Sprintf("Comente no card %s: o ciclo automático de correção terminou e exige intervenção humana (PR: %s)", n.CardURL, n.PRURL)
	}
	return fmt.Sprintf("Comente no card %s: %s", n.CardURL, n.Title)
}

// ---- inbox ----

func writeInbox(dir string, n Notification) (string, error) {
	inbox := layout.Inbox(dir)
	path := filepath.Join(inbox, n.ID+".json")
	if _, err := os.Lstat(path); err == nil {
		return path, nil // idempotente
	}
	if err := queue.WriteJSONAtomic(path, n, 0o644); err != nil {
		return "", fmt.Errorf("gravar inbox: %w", err)
	}
	return path, nil
}

// InboxItem é uma notificação lida do inbox.
type InboxItem struct {
	Notification
	Path     string `json:"-"`
	Archived bool   `json:"-"`
}

// ListInbox devolve as notificações do inbox, da mais antiga para a mais nova. unreadOnly filtra
// as não lidas. Arquivos corrompidos ou acima de 64 KB são ignorados.
func ListInbox(dir string, unreadOnly bool) ([]InboxItem, error) {
	entries, err := os.ReadDir(layout.Inbox(dir))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var items []InboxItem
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || strings.HasPrefix(name, ".") || !strings.HasSuffix(name, ".json") {
			continue
		}
		p := filepath.Join(layout.Inbox(dir), name)
		it, ok := readInboxFile(p)
		if !ok || (unreadOnly && it.Read) {
			continue
		}
		items = append(items, it)
	}
	sort.Slice(items, func(i, j int) bool {
		if !items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].CreatedAt.Before(items[j].CreatedAt)
		}
		return items[i].ID < items[j].ID
	})
	return items, nil
}

func readInboxFile(p string) (InboxItem, bool) {
	info, err := os.Lstat(p)
	if err != nil || !info.Mode().IsRegular() || info.Size() > 64*1024 {
		return InboxItem{}, false
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return InboxItem{}, false
	}
	var n Notification
	if err := json.Unmarshal(data, &n); err != nil || !idRe.MatchString(n.ID) {
		return InboxItem{}, false
	}
	// Texto vindo do disco também é dado não confiável (FLX-08).
	n.Title = Sanitize(n.Title, maxTitleBytes, false)
	n.Message = Sanitize(n.Message, maxMessageBytes, true)
	n.PRURL = Sanitize(n.PRURL, maxFieldBytes, false)
	n.PMComment = Sanitize(n.PMComment, maxMessageBytes, true)
	n.CardURL = Sanitize(n.CardURL, maxFieldBytes, false)
	n.Feature = Sanitize(n.Feature, 100, false)
	n.JobID = Sanitize(n.JobID, 100, false)
	return InboxItem{Notification: n, Path: p}, true
}

// MarkRead marca a notificação como lida; id inexistente é erro.
func MarkRead(dir, id string) error {
	if !idRe.MatchString(id) {
		return fmt.Errorf("%w: id %q", errInvalid, id)
	}
	p := filepath.Join(layout.Inbox(dir), id+".json")
	it, ok := readInboxFile(p)
	if !ok {
		return fmt.Errorf("notificação %q não encontrada", id)
	}
	it.Read = true
	return queue.WriteJSONAtomic(p, it.Notification, 0o644)
}

// MarkAllRead marca todas como lidas e devolve quantas.
func MarkAllRead(dir string) (int, error) {
	items, err := ListInbox(dir, true)
	if err != nil {
		return 0, err
	}
	for _, it := range items {
		if err := MarkRead(dir, it.ID); err != nil {
			return 0, err
		}
	}
	return len(items), nil
}

// ArchiveRead move as notificações lidas para discovery/archive.
func ArchiveRead(dir string) (int, error) {
	items, err := ListInbox(dir, false)
	if err != nil {
		return 0, err
	}
	moved := 0
	for _, it := range items {
		if !it.Read {
			continue
		}
		dst := filepath.Join(layout.InboxArchive(dir), filepath.Base(it.Path))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return moved, err
		}
		if err := os.Rename(it.Path, dst); err != nil {
			return moved, err
		}
		moved++
	}
	return moved, nil
}

// FormatInbox renderiza as notificações para o terminal ou para o hook.
func FormatInbox(items []InboxItem, compact bool) string {
	if len(items) == 0 {
		return ""
	}
	var b strings.Builder
	if compact {
		fmt.Fprintf(&b, "[forge] %d notificação(ões) não lida(s) do pipeline:\n", len(items))
	}
	for _, it := range items {
		switch it.Type {
		case TypePRApproved:
			fmt.Fprintf(&b, "- PR revisado, sem inconformidades e pronto para aprovação final: %s", orDash(it.PRURL))
		case TypeEscalation:
			fmt.Fprintf(&b, "- ATENÇÃO, intervenção humana necessária: %s", it.Title)
			if it.PRURL != "" {
				fmt.Fprintf(&b, " (PR: %s)", it.PRURL)
			}
		default:
			fmt.Fprintf(&b, "- %s", it.Title)
		}
		if it.Feature != "" {
			fmt.Fprintf(&b, " [%s]", it.Feature)
		}
		fmt.Fprintf(&b, " (id %s)\n", it.ID)
		if !compact && it.Message != "" {
			for _, line := range strings.Split(it.Message, "\n") {
				fmt.Fprintf(&b, "    %s\n", line)
			}
		}
		if it.PMComment != "" && !compact {
			fmt.Fprintf(&b, "    Instrução ao agente: %s\n", it.PMComment)
		}
	}
	return b.String()
}

func orDash(s string) string {
	if s == "" {
		return "(sem link)"
	}
	return s
}

// ---- toast ----

func toast(n Notification, d Deps) error {
	title := Sanitize(n.Title, 100, false)
	msg := Sanitize(firstLine(n.Message), 200, false)
	switch d.GOOS {
	case "darwin":
		script := fmt.Sprintf("display notification %s with title %s", appleScriptString(msg), appleScriptString(title))
		return d.RunToast("osascript", "-e", script)
	case "linux":
		return d.RunToast("notify-send", "--", title, msg)
	case "windows":
		script := "$t='" + psSingle(title) + "';$m='" + psSingle(msg) + "';" +
			"[void][Windows.UI.Notifications.ToastNotificationManager,Windows.UI.Notifications,ContentType=WindowsRuntime];" +
			"$x=[Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02);" +
			"$n=$x.GetElementsByTagName('text');[void]$n.Item(0).AppendChild($x.CreateTextNode($t));[void]$n.Item(1).AppendChild($x.CreateTextNode($m));" +
			"[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('Forge SDD').Show([Windows.UI.Notifications.ToastNotification]::new($x))"
		return d.RunToast("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	}
	return fmt.Errorf("toast não suportado em %s", d.GOOS)
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func appleScriptString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}

// psSingle duplica as aspas simples do PowerShell, inclusive as tipográficas (U+2018 a U+201B), que
// ele também trata como delimitador.
func psSingle(s string) string {
	return strings.NewReplacer("'", "''", "\u2018", "''", "\u2019", "''", "\u201a", "''", "\u201b", "''").Replace(s)
}

// ---- webhook ----

func sendWebhook(c config.WebhookChannel, n Notification, d Deps) error {
	raw := d.Getenv(c.URLEnv)
	if raw == "" {
		return fmt.Errorf("variável de ambiente %s não definida", c.URLEnv)
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return errors.New("URL do webhook inválida")
	}
	host := u.Hostname()
	if u.Scheme != "https" && !(u.Scheme == "http" && (loopbackIPs[host] || d.AllowHTTP)) {
		return errors.New("URL do webhook deve usar https (http só em loopback)")
	}
	text := fmt.Sprintf("%s\n%s", n.Title, firstLine(n.Message))
	if n.PRURL != "" {
		text += "\n" + n.PRURL
	}
	var payload any
	switch c.Provider {
	case "slack":
		payload = map[string]string{"text": text}
	default:
		payload = map[string]any{"type": n.Type, "title": n.Title, "message": n.Message, "pr_url": n.PRURL, "card_url": n.CardURL, "feature": n.Feature, "id": n.ID}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, raw, bytes.NewReader(body))
	if err != nil {
		return errors.New("falha ao montar a requisição do webhook")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.HTTP.Do(req)
	if err != nil {
		return errors.New("falha ao chamar o webhook") // sem repetir a URL (segredo)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("webhook respondeu HTTP %d", resp.StatusCode)
	}
	return nil
}

// ---- e-mail ----

func sendEmail(c config.EmailChannel, n Notification, d Deps) error {
	from, err := mail.ParseAddress(c.From)
	if err != nil {
		return errors.New("remetente inválido")
	}
	var rcpts []string
	for _, to := range c.To {
		a, err := mail.ParseAddress(to)
		if err != nil {
			return errors.New("destinatário inválido")
		}
		rcpts = append(rcpts, a.Address)
	}
	subject := "[Forge] " + n.Title
	body := n.Message
	if n.PRURL != "" {
		body += "\n\nPR: " + n.PRURL
	}
	if n.CardURL != "" {
		body += "\nCard: " + n.CardURL
	}
	// Cabeçalhos sem quebra de linha: título e endereços já foram sanitizados e validados.
	msg := []byte("From: " + from.Address + "\r\nTo: " + strings.Join(rcpts, ", ") + "\r\nSubject: " + subject +
		"\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n" + strings.ReplaceAll(body, "\n", "\r\n") + "\r\n")
	var auth smtp.Auth
	if c.UserEnv != "" {
		user, pass := d.Getenv(c.UserEnv), d.Getenv(c.PasswordEnv)
		if user == "" || pass == "" {
			return fmt.Errorf("variáveis %s e %s precisam estar definidas", c.UserEnv, c.PasswordEnv)
		}
		auth = smtp.PlainAuth("", user, pass, c.Host)
	}
	addr := net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
	if err := d.SendMail(addr, auth, from.Address, rcpts, msg); err != nil {
		return errors.New("falha ao enviar o e-mail") // o erro do SMTP pode conter usuário ou servidor
	}
	return nil
}
