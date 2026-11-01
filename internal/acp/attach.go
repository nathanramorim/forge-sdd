package acp

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/forge-sdd/cli/internal/queue"
)

const (
	// DefaultMaxHistoryLogEntries define o tamanho máximo do buffer de histórico em memória por estação.
	DefaultMaxHistoryLogEntries = 100
	// DefaultSubscriberBufferSize define a capacidade do canal pub/sub de cada cliente anexado.
	DefaultSubscriberBufferSize = 128
)

// AttachBridge gerencia as anexações interativas e a multiplexação de streams pub/sub
// de logs e comandos das 5 estações canônicas de forma thread-safe (ARQ-03, COD-01).
type AttachBridge struct {
	mu          sync.RWMutex
	pool        *SessionPool
	subscribers map[string]map[uint64]chan SessionLogEntry // Chave: station -> subID -> chan
	history     map[string][]SessionLogEntry               // Chave: station -> entradas ordenadas
	maxHistory  int
	nextSubID   uint64
}

// NewAttachBridge instancia um multiplexador de anexação com histórico configurável.
func NewAttachBridge(pool *SessionPool) *AttachBridge {
	return &AttachBridge{
		pool:        pool,
		subscribers: make(map[string]map[uint64]chan SessionLogEntry),
		history:     make(map[string][]SessionLogEntry),
		maxHistory:  DefaultMaxHistoryLogEntries,
	}
}

// Attach anexa um cliente aos logs em tempo real da estação especificada.
// Retorna um canal de leitura unidirecional, uma função idempotente de limpeza (detach)
// e um erro se a estação não for canônica.
// A desconexão através da função detach() NUNCA encerra ou altera a sessão no pool.
func (b *AttachBridge) Attach(station string) (<-chan SessionLogEntry, func(), error) {
	if !queue.IsValidStation(station) {
		return nil, nil, fmt.Errorf("estação inválida: %q. Estações canônicas válidas: %s", station, strings.Join(queue.CanonicalStations(), ", "))
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.nextSubID++
	subID := b.nextSubID
	ch := make(chan SessionLogEntry, DefaultSubscriberBufferSize)

	if b.subscribers[station] == nil {
		b.subscribers[station] = make(map[uint64]chan SessionLogEntry)
	}
	b.subscribers[station][subID] = ch

	if b.pool != nil {
		b.pool.IncrementAttachedClients(station)
	}

	var once sync.Once
	detach := func() {
		once.Do(func() {
			b.removeSubscriber(station, subID)
		})
	}

	return ch, detach, nil
}

// removeSubscriber remove a inscrição e fecha o canal de forma atômica e segura.
func (b *AttachBridge) removeSubscriber(station string, subID uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if subs, ok := b.subscribers[station]; ok {
		if ch, exists := subs[subID]; exists {
			delete(subs, subID)
			close(ch)
			if b.pool != nil {
				b.pool.DecrementAttachedClients(station)
			}
		}
	}
}

// BroadcastLog envia uma entrada de log para o histórico da estação e para todos os ouvintes conectados.
func (b *AttachBridge) BroadcastLog(station string, entry SessionLogEntry) {
	if !queue.IsValidStation(station) {
		return
	}
	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	if entry.Type == "" {
		entry.Type = "info"
	}

	b.mu.Lock()
	// 1. Armazena no histórico delimitado
	h := b.history[station]
	if len(h) >= b.maxHistory {
		h = append(h[1:], entry)
	} else {
		h = append(h, entry)
	}
	b.history[station] = h

	// 2. Envia para todos os inscritos da estação de forma não-bloqueante
	if subs, ok := b.subscribers[station]; ok {
		for _, ch := range subs {
			select {
			case ch <- entry:
			default:
				// Buffer cheio: descarta para este cliente lento sem travar os demais
			}
		}
	}
	b.mu.Unlock()
}

// GetHistory retorna uma cópia defensiva do histórico de logs recente da estação.
func (b *AttachBridge) GetHistory(station string) []SessionLogEntry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	h, ok := b.history[station]
	if !ok || len(h) == 0 {
		return nil
	}
	cp := make([]SessionLogEntry, len(h))
	copy(cp, h)
	return cp
}

// AttachedClients retorna o número atual de clientes conectados à estação.
func (b *AttachBridge) AttachedClients(station string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if subs, ok := b.subscribers[station]; ok {
		return len(subs)
	}
	return 0
}

// Run conecta o terminal do usuário aos streams de entrada e saída, exibindo banner e logs em tempo real.
func (b *AttachBridge) Run(ctx context.Context, station string, in io.Reader, out io.Writer) error {
	if out == nil {
		out = io.Discard
	}

	ch, detach, err := b.Attach(station)
	if err != nil {
		return err
	}
	defer detach()

	// Obtém metadados da sessão no pool
	sessionID := fmt.Sprintf("sess-%s", station)
	state := SessionStateStandby
	activeItemID := ""

	if b.pool != nil {
		if sess, ok := b.pool.GetSessionByStation(station); ok {
			sessionID = sess.SessionID
			state = sess.State
			activeItemID = sess.ActiveItemID
		}
	} else {
		for _, def := range CanonicalPoolSessions {
			if def.Station == station {
				sessionID = def.ID
				break
			}
		}
	}

	// 1. Exibe banner formatado com estado
	fmt.Fprint(out, formatAttachBanner(station, sessionID, state, activeItemID))

	// 2. Exibe histórico existente
	for _, entry := range b.GetHistory(station) {
		fmt.Fprint(out, formatLogLine(entry))
	}

	// 3. Monitoramento de entrada interativa do terminal
	inputDone := make(chan struct{})
	if in != nil {
		go func() {
			scanner := bufio.NewScanner(in)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if strings.EqualFold(line, "exit") || strings.EqualFold(line, "quit") {
					close(inputDone)
					return
				}
				if line != "" {
					b.BroadcastLog(station, SessionLogEntry{
						Timestamp: time.Now().UTC().Format(time.RFC3339),
						Type:      "user_input",
						Message:   line,
					})
				}
			}
			close(inputDone)
		}()
	}

	// 4. Multiplexação de streams e saída graciosa
	for {
		select {
		case <-ctx.Done():
			fmt.Fprintf(out, "\n[detach] Desconectado da estação %s (sinal de interrupção). A sessão no pool permanece ativa.\n", station)
			return nil
		case <-inputDone:
			fmt.Fprintf(out, "\n[detach] Desconectado da estação %s (comando exit). A sessão no pool permanece ativa.\n", station)
			return nil
		case entry, ok := <-ch:
			if !ok {
				return nil
			}
			fmt.Fprint(out, formatLogLine(entry))
		}
	}
}

func formatAttachBanner(station, sessionID string, state SessionState, activeItemID string) string {
	var badge string
	switch state {
	case SessionStateStandby:
		badge = "🟢 Standby"
	case SessionStateProcessing:
		badge = "⚡ Processing"
	case SessionStateApprovalPending:
		badge = "⏸️ Approval Pending"
	case SessionStateError:
		badge = "❌ Error"
	case SessionStateOffline:
		badge = "⚪ Offline"
	default:
		if state == "" {
			badge = "🟢 Standby"
		} else {
			badge = fmt.Sprintf("🔵 %s", state)
		}
	}

	banner := "================================================================================\n"
	banner += fmt.Sprintf("🔌 FORGE-SDD ATTACH — Estação: %s | Sessão: %s | Estado: %s\n", station, sessionID, badge)
	if activeItemID != "" {
		banner += fmt.Sprintf("📋 Item ativo: %s\n", activeItemID)
	}
	banner += "Desconexão graciosa: pressione Ctrl+C ou digite 'exit' para sair sem afetar a sessão.\n"
	banner += "================================================================================\n"
	return banner
}

func formatLogLine(e SessionLogEntry) string {
	ts := e.Timestamp
	if ts == "" {
		ts = time.Now().UTC().Format(time.RFC3339)
	}
	tipo := e.Type
	if tipo == "" {
		tipo = "info"
	}
	return fmt.Sprintf("[%s] [%s] %s\n", ts, tipo, e.Message)
}
