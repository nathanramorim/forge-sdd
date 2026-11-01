package acp

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const (
	// DefaultHeartbeatInterval é o intervalo padrão entre ticks do monitor de batimento.
	DefaultHeartbeatInterval = 1 * time.Second

	// DefaultStallTimeout é o tempo máximo sem telemetria antes de considerar uma sessão em processamento como stalled.
	DefaultStallTimeout = 30 * time.Second
)

// HeartbeatMonitor executa ciclos periódicos não-bloqueantes para monitorar a integridade
// das sessões no pool, atualizando LastSeenAt e detectando stalls (ARQ-03, COD-01).
type HeartbeatMonitor struct {
	pool         *SessionPool
	dispatcher   *RPCDispatcher
	interval     time.Duration
	stallTimeout time.Duration
	ticker       *time.Ticker
	stopCh       chan struct{}
	doneCh       chan struct{}
	running      bool
	mu           sync.Mutex
	onTick       func()
	onStalled    func(*PooledSession)
}

// NewHeartbeatMonitor instancia um monitor com intervalos padrão.
func NewHeartbeatMonitor(pool *SessionPool, dispatcher *RPCDispatcher) *HeartbeatMonitor {
	return NewHeartbeatMonitorConfig(pool, dispatcher, DefaultHeartbeatInterval, DefaultStallTimeout)
}

// NewHeartbeatMonitorConfig instancia um monitor com intervalos customizáveis.
func NewHeartbeatMonitorConfig(pool *SessionPool, dispatcher *RPCDispatcher, interval, stallTimeout time.Duration) *HeartbeatMonitor {
	if interval <= 0 {
		interval = DefaultHeartbeatInterval
	}
	if stallTimeout <= 0 {
		stallTimeout = DefaultStallTimeout
	}
	return &HeartbeatMonitor{
		pool:         pool,
		dispatcher:   dispatcher,
		interval:     interval,
		stallTimeout: stallTimeout,
	}
}

// SetInterval altera o intervalo do ticker.
func (hm *HeartbeatMonitor) SetInterval(interval time.Duration) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	if interval > 0 {
		hm.interval = interval
		if hm.running && hm.ticker != nil {
			hm.ticker.Reset(interval)
		}
	}
}

// SetStallTimeout altera o limite de timeout para detecção de stall.
func (hm *HeartbeatMonitor) SetStallTimeout(timeout time.Duration) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	if timeout > 0 {
		hm.stallTimeout = timeout
	}
}

// SetOnTick configura um hook executado ao final de cada tick.
func (hm *HeartbeatMonitor) SetOnTick(fn func()) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.onTick = fn
}

// SetOnStalled configura um hook executado quando uma sessão stalled é detectada.
func (hm *HeartbeatMonitor) SetOnStalled(fn func(*PooledSession)) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.onStalled = fn
}

// IsRunning indica se o monitor de heartbeat está ativo em background.
func (hm *HeartbeatMonitor) IsRunning() bool {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	return hm.running
}

// Start inicia o loop periódico do monitor em uma goroutine dedicada.
func (hm *HeartbeatMonitor) Start(ctx context.Context) {
	hm.mu.Lock()
	if hm.running {
		hm.mu.Unlock()
		return
	}
	hm.running = true
	stopCh := make(chan struct{})
	doneCh := make(chan struct{})
	hm.stopCh = stopCh
	hm.doneCh = doneCh
	ticker := time.NewTicker(hm.interval)
	hm.ticker = ticker
	hm.mu.Unlock()

	go func() {
		defer func() {
			ticker.Stop()
			close(doneCh)
			hm.mu.Lock()
			hm.running = false
			hm.mu.Unlock()
		}()

		for {
			select {
			case <-ticker.C:
				hm.Tick()
			case <-stopCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop finaliza a goroutine do monitor de forma graciosa e determinística (COD-01).
func (hm *HeartbeatMonitor) Stop() {
	hm.mu.Lock()
	if !hm.running {
		hm.mu.Unlock()
		return
	}
	hm.running = false
	if hm.ticker != nil {
		hm.ticker.Stop()
	}
	close(hm.stopCh)
	doneCh := hm.doneCh
	hm.mu.Unlock()

	if doneCh != nil {
		<-doneCh
	}
}

// Tick executa uma única iteração de verificação de heartbeat de forma não-bloqueante.
func (hm *HeartbeatMonitor) Tick() {
	if hm.pool == nil {
		return
	}

	now := time.Now().UTC()
	hm.mu.Lock()
	stallTimeout := hm.stallTimeout
	onTick := hm.onTick
	onStalled := hm.onStalled
	dispatcher := hm.dispatcher
	hm.mu.Unlock()

	stalled := hm.pool.updateHeartbeat(now, stallTimeout)

	for _, s := range stalled {
		if onStalled != nil {
			onStalled(s)
		}
		if dispatcher != nil {
			dispatcher.BroadcastNotification(NewNotification("session/stalled", s))
			dispatcher.BroadcastSessionNotification(SessionNotification{
				SessionID: s.SessionID,
				Type:      "session_stalled",
				Station:   s.Station,
				Message:   fmt.Sprintf("Sessão %s na estação %s excedeu timeout de processamento", s.SessionID, s.Station),
				Payload:   s,
				Timestamp: now.Format(time.RFC3339),
			})
		}
	}

	if onTick != nil {
		onTick()
	}
}
