package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/bus"
	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
	"github.com/forge-sdd/cli/internal/queue"
	"github.com/forge-sdd/cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type safeClock struct {
	mu  sync.Mutex
	cur time.Time
}

func newSafeClock(t time.Time) *safeClock {
	return &safeClock{cur: t}
}

func (c *safeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cur
}

func (c *safeClock) Add(d time.Duration) time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cur = c.cur.Add(d)
	return c.cur
}

type fakeTmuxDeliverer struct {
	mu          sync.Mutex
	paneTexts   []string
	paneIdx     int
	currentText string
	buffers     map[string][]byte
	pastes      []struct{ name, target string }
	keys        []struct {
		target string
		keys   []string
	}
	errOn map[string]error
}

func (f *fakeTmuxDeliverer) CapturePane(target string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.errOn != nil {
		if err, ok := f.errOn["CapturePane"]; ok {
			return "", err
		}
	}
	if len(f.paneTexts) > 0 {
		if f.paneIdx >= len(f.paneTexts) {
			return f.paneTexts[len(f.paneTexts)-1], nil
		}
		txt := f.paneTexts[f.paneIdx]
		f.paneIdx++
		return txt, nil
	}
	if f.currentText == "" {
		f.currentText = "prompt> "
	}
	return f.currentText, nil
}

func (f *fakeTmuxDeliverer) LoadBuffer(name string, data []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.errOn != nil {
		if err, ok := f.errOn["LoadBuffer"]; ok {
			return err
		}
	}
	if f.buffers == nil {
		f.buffers = make(map[string][]byte)
	}
	f.buffers[name] = data
	return nil
}

func (f *fakeTmuxDeliverer) PasteBuffer(name, target string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.errOn != nil {
		if err, ok := f.errOn["PasteBuffer"]; ok {
			return err
		}
	}
	f.pastes = append(f.pastes, struct{ name, target string }{name, target})
	if len(f.paneTexts) == 0 {
		f.currentText = fmt.Sprintf("prompt> [pasted #%d: %s]", len(f.pastes), name)
	}
	return nil
}

func (f *fakeTmuxDeliverer) SendKeys(target string, keys ...string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.errOn != nil {
		if err, ok := f.errOn["SendKeys"]; ok {
			return err
		}
	}
	f.keys = append(f.keys, struct {
		target string
		keys   []string
	}{target, keys})
	if len(f.paneTexts) == 0 {
		f.currentText = fmt.Sprintf("prompt> [keys #%d: %s]", len(f.keys), strings.Join(keys, " "))
	}
	return nil
}

func (f *fakeTmuxDeliverer) PastesLen() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.pastes)
}

func (f *fakeTmuxDeliverer) Buffers() map[string][]byte {
	f.mu.Lock()
	defer f.mu.Unlock()
	cpy := make(map[string][]byte, len(f.buffers))
	for k, v := range f.buffers {
		cpy[k] = v
	}
	return cpy
}

func (f *fakeTmuxDeliverer) LastPaste() (struct{ name, target string }, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.pastes) == 0 {
		return struct{ name, target string }{}, false
	}
	return f.pastes[len(f.pastes)-1], true
}

func (f *fakeTmuxDeliverer) LastKey() ([]string, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.keys) == 0 {
		return nil, false
	}
	return f.keys[len(f.keys)-1].keys, true
}

func (f *fakeTmuxDeliverer) KeyCount(name string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	c := 0
	for _, k := range f.keys {
		if len(k.keys) > 0 && k.keys[0] == name {
			c++
		}
	}
	return c
}

func setupProjectWithJob(t *testing.T, role string) (string, bus.Job) {
	t.Helper()
	dir := project(t, func(d string) {
		cfg := config.DefaultForgeConfig()
		cfg.SessionMode = config.SessionModeLive
		cfg.Context = config.ContextClear
		cfg.Live.StartupWaitSeconds = 0
		cfg.Live.ReadyQuietSeconds = 1
		data, _ := json.Marshal(cfg)
		testutil.WriteFile(t, d, ".forge/config/forge.json", string(data))
	})
	prNum := 10
	prURL := "https://github.com/pr/10"
	job := bus.Job{
		ID:                "job-test-1",
		Feature:           "auth-oauth",
		Title:             "Auth",
		SpecPath:          "sdd/discovery/77-auth-oauth/criteria.md",
		Type:              bus.TypeFeatureBuild,
		Priority:          bus.PriorityNormal,
		PR:                bus.PR{Number: &prNum, URL: &prURL, Branch: "feat/77-auth-oauth", BaseBranch: "main"},
		Attempt:           1,
		MaxReworkAttempts: 3,
		CreatedAt:         now,
	}
	if role == layout.RoleReviewer {
		job.Type = bus.TypePRReview
	}
	_, _, err := bus.Enqueue(dir, role, job, now)
	require.NoError(t, err)
	return dir, job
}

func TestDeliver(t *testing.T) {
	// 1. Colagem por load-buffer + paste-buffer -p + Enter, prompt intacto
	t.Run("colagem intacta", func(t *testing.T) {
		dir, _ := setupProjectWithJob(t, layout.RoleBuilder)
		target := "=test:forge.1"
		d := &fakeTmuxDeliverer{}

		clk := newSafeClock(now)
		var out bytes.Buffer

		// Simula conclusão rápida logo após a entrega
		go func() {
			time.Sleep(50 * time.Millisecond)
			_, _ = bus.UpdatePR(dir, layout.RoleBuilder, "job-test-1", "forge-builder", 10, "https://github.com/pr/10", "", clk.Now())
		}()

		err := RunWorker(context.Background(), WorkerOptions{
			Dir:           dir,
			Role:          layout.RoleBuilder,
			DeliverTarget: target,
			Deliverer:     d,
			Once:          true,
			Now:           clk.Now,
			Out:           &out,
			Sleep: func(_ context.Context, dt time.Duration) bool {
				clk.Add(dt)
				time.Sleep(10 * time.Millisecond)
				return true
			},
		})
		require.NoError(t, err)

		// Confere buffers
		bufs := d.Buffers()
		require.NotEmpty(t, bufs)
		var promptContent string
		for name, buf := range bufs {
			if strings.HasPrefix(name, "forge-job-test-1") {
				promptContent = string(buf)
				break
			}
		}
		require.NotEmpty(t, promptContent, "deve ter carregado o buffer do job")
		assert.Contains(t, promptContent, "auth-oauth")
		// Confere colagem e teclas
		lastP, okP := d.LastPaste()
		require.True(t, okP)
		assert.Equal(t, target, lastP.target)
		lastK, okK := d.LastKey()
		require.True(t, okK)
		assert.Equal(t, []string{"Enter"}, lastK)
		assert.Contains(t, out.String(), "entregue ao painel =test:forge.1")
	})

	// 2. Nunca entrega em painel ocupado: se o painel não parar, devolve à fila (requeue)
	t.Run("painel ocupado", func(t *testing.T) {
		dir, _ := setupProjectWithJob(t, layout.RoleBuilder)
		target := "=test:forge.1"
		// Painel muda constantemente
		var busyTexts []string
		for i := 0; i < 200; i++ {
			busyTexts = append(busyTexts, fmt.Sprintf("busy output line %d", i))
		}
		d := &fakeTmuxDeliverer{paneTexts: busyTexts}

		clk := newSafeClock(now)
		var out bytes.Buffer

		err := RunWorker(context.Background(), WorkerOptions{
			Dir:           dir,
			Role:          layout.RoleBuilder,
			DeliverTarget: target,
			Deliverer:     d,
			Once:          true,
			Now:           clk.Now,
			Out:           &out,
			Sleep: func(_ context.Context, dt time.Duration) bool {
				clk.Add(dt)
				return true
			},
		})
		require.NoError(t, err)

		// Job deve ter sido devolvido (requeued)
		pend, err := bus.List(dir, layout.RoleBuilder, layout.StatePending)
		require.NoError(t, err)
		require.Len(t, pend, 1)
		assert.Contains(t, out.String(), "painel ocupado ou não respondeu")
	})

	// 3. Reenvio único do Enter e falha 'não aceitou' vai a Requeue
	t.Run("reenvio enter e requeue", func(t *testing.T) {
		dir, _ := setupProjectWithJob(t, layout.RoleBuilder)
		target := "=test:forge.1"
		// Painel fica estático antes e DEPOIS da colagem (nunca muda)
		d := &fakeTmuxDeliverer{
			paneTexts: []string{"prompt> "},
		}

		clk := newSafeClock(now)
		var out bytes.Buffer

		err := RunWorker(context.Background(), WorkerOptions{
			Dir:           dir,
			Role:          layout.RoleBuilder,
			DeliverTarget: target,
			Deliverer:     d,
			Once:          true,
			Now:           clk.Now,
			Out:           &out,
			Sleep: func(_ context.Context, dt time.Duration) bool {
				clk.Add(dt)
				return true
			},
		})
		require.NoError(t, err)

		// Deve ter tentado Enter novamente (pelo menos 2 Enters: 1 da entrega + 1 do retry)
		enterCount := d.KeyCount("Enter")
		assert.GreaterOrEqual(t, enterCount, 2, "deve ter reenviado Enter")
		assert.Contains(t, out.String(), "o agente não aceitou o prompt")

		// Job foi requeued
		pend, err := bus.List(dir, layout.RoleBuilder, layout.StatePending)
		require.NoError(t, err)
		require.Len(t, pend, 1)
	})
}

func TestDeliverEstado(t *testing.T) {
	// 1. Builder conclui por estado (PR aberto)
	t.Run("builder pr registrado", func(t *testing.T) {
		dir, _ := setupProjectWithJob(t, layout.RoleBuilder)
		target := "=test:forge.1"
		d := &fakeTmuxDeliverer{}

		clk := newSafeClock(now)
		var out bytes.Buffer

		go func() {
			time.Sleep(50 * time.Millisecond)
			_, _ = bus.UpdatePR(dir, layout.RoleBuilder, "job-test-1", "forge-builder", 42, "https://github.com/pr/42", "", clk.Now())
		}()

		err := RunWorker(context.Background(), WorkerOptions{
			Dir:           dir,
			Role:          layout.RoleBuilder,
			DeliverTarget: target,
			Deliverer:     d,
			Once:          true,
			Now:           clk.Now,
			Out:           &out,
			Sleep: func(_ context.Context, dt time.Duration) bool {
				clk.Add(dt)
				time.Sleep(10 * time.Millisecond)
				return true
			},
		})
		require.NoError(t, err)

		done, err := bus.List(dir, layout.RoleBuilder, layout.StateDone)
		require.NoError(t, err)
		require.Len(t, done, 1)
		assert.Contains(t, out.String(), "job job-test-1 concluído")

		// Revisor deve ter recebido a revisão
		revJobs, err := bus.List(dir, layout.RoleReviewer, layout.StatePending)
		require.NoError(t, err)
		require.Len(t, revJobs, 1)
	})

	// 2. Reviewer conclui por estado (veredito aplicado)
	t.Run("reviewer veredito aplicado", func(t *testing.T) {
		dir, _ := setupProjectWithJob(t, layout.RoleReviewer)
		target := "=test:forge.2"
		d := &fakeTmuxDeliverer{}

		clk := newSafeClock(now)
		var out bytes.Buffer
		deps, _ := notifyDeps(nil)

		go func() {
			time.Sleep(50 * time.Millisecond)
			_, _ = ApplyVerdict(dir, "job-test-1", "forge-reviewer", queue.Verdict{Veredito: "conforme", Resumo: "Aprovado"}, deps, clk.Now())
		}()

		err := RunWorker(context.Background(), WorkerOptions{
			Dir:           dir,
			Role:          layout.RoleReviewer,
			DeliverTarget: target,
			Deliverer:     d,
			Once:          true,
			Now:           clk.Now,
			Out:           &out,
			Sleep: func(_ context.Context, dt time.Duration) bool {
				clk.Add(dt)
				time.Sleep(10 * time.Millisecond)
				return true
			},
		})
		require.NoError(t, err)

		done, err := bus.List(dir, layout.RoleReviewer, layout.StateDone)
		require.NoError(t, err)
		require.Len(t, done, 1)
		assert.Contains(t, out.String(), "job job-test-1 concluído")
	})

	// 3. Timeout do job e renovação de lease
	t.Run("timeout e renovacao lease", func(t *testing.T) {
		dir := project(t, func(d string) {
			cfg := config.DefaultForgeConfig()
			cfg.SessionMode = config.SessionModeLive
			cfg.Builder.JobTimeoutMinutes = 1
			cfg.Live.StartupWaitSeconds = 0
			cfg.Live.ReadyQuietSeconds = 1
			data, _ := json.Marshal(cfg)
			testutil.WriteFile(t, d, ".forge/config/forge.json", string(data))
		})
		job := bus.Job{
			ID:                "job-timeout",
			Feature:           "auth-oauth",
			Title:             "Auth",
			SpecPath:          "sdd/discovery/77-auth-oauth/criteria.md",
			Type:              bus.TypeFeatureBuild,
			Priority:          bus.PriorityNormal,
			PR:                bus.PR{Branch: "feat/77-auth-oauth", BaseBranch: "main"},
			Attempt:           1,
			MaxReworkAttempts: 3,
			CreatedAt:         now,
		}
		_, _, err := bus.Enqueue(dir, layout.RoleBuilder, job, now)
		require.NoError(t, err)

		target := "=test:forge.1"
		d := &fakeTmuxDeliverer{}

		clk := newSafeClock(now)
		var out bytes.Buffer

		err = RunWorker(context.Background(), WorkerOptions{
			Dir:           dir,
			Role:          layout.RoleBuilder,
			DeliverTarget: target,
			Deliverer:     d,
			Once:          true,
			Now:           clk.Now,
			Out:           &out,
			RenewEvery:    20 * time.Second,
			Sleep: func(_ context.Context, dt time.Duration) bool {
				if d.PastesLen() >= 2 {
					clk.Add(15 * time.Minute) // avança o tempo rapidamente após a entrega
				} else {
					clk.Add(dt)
				}
				return true
			},
		})
		require.NoError(t, err)

		failed, err := bus.List(dir, layout.RoleBuilder, layout.StateFailed)
		require.NoError(t, err)
		require.Len(t, failed, 1)
		require.NotEmpty(t, failed[0].Job.Events)
		assert.Contains(t, failed[0].Job.Events[len(failed[0].Job.Events)-1].Note, "tempo limite excedido")
	})
}

func TestDeliverContexto(t *testing.T) {
	// 1. clear envia reset_command antes de cada job e espera ocioso
	t.Run("clear envia reset_command", func(t *testing.T) {
		dir, _ := setupProjectWithJob(t, layout.RoleBuilder)
		target := "=test:forge.1"
		d := &fakeTmuxDeliverer{}

		clk := newSafeClock(now)
		var out bytes.Buffer

		go func() {
			time.Sleep(50 * time.Millisecond)
			_, _ = bus.UpdatePR(dir, layout.RoleBuilder, "job-test-1", "forge-builder", 10, "https://github.com/pr/10", "", clk.Now())
		}()

		err := RunWorker(context.Background(), WorkerOptions{
			Dir:           dir,
			Role:          layout.RoleBuilder,
			DeliverTarget: target,
			Deliverer:     d,
			Once:          true,
			Now:           clk.Now,
			Out:           &out,
			Sleep: func(_ context.Context, dt time.Duration) bool {
				clk.Add(dt)
				time.Sleep(10 * time.Millisecond)
				return true
			},
		})
		require.NoError(t, err)

		// Deve ter enviado /clear (comando padrão do claude)
		foundClear := false
		for _, buf := range d.Buffers() {
			if string(buf) == "/clear" {
				foundClear = true
				break
			}
		}
		assert.True(t, foundClear, "deve ter enviado /clear antes do job")
	})

	// 2. continuous não envia reset_command
	t.Run("continuous nao envia reset", func(t *testing.T) {
		dir := project(t, func(d string) {
			cfg := config.DefaultForgeConfig()
			cfg.SessionMode = config.SessionModeLive
			cfg.Context = config.ContextContinuous
			cfg.Live.StartupWaitSeconds = 0
			cfg.Live.ReadyQuietSeconds = 1
			data, _ := json.Marshal(cfg)
			testutil.WriteFile(t, d, ".forge/config/forge.json", string(data))
		})
		job := bus.Job{
			ID:                "job-cont",
			Feature:           "auth-oauth",
			Title:             "Auth",
			SpecPath:          "sdd/discovery/77-auth-oauth/criteria.md",
			Type:              bus.TypeFeatureBuild,
			Priority:          bus.PriorityNormal,
			PR:                bus.PR{Branch: "feat/77-auth-oauth", BaseBranch: "main"},
			Attempt:           1,
			MaxReworkAttempts: 3,
			CreatedAt:         now,
		}
		_, _, err := bus.Enqueue(dir, layout.RoleBuilder, job, now)
		require.NoError(t, err)

		target := "=test:forge.1"
		d := &fakeTmuxDeliverer{}

		clk := newSafeClock(now)
		var out bytes.Buffer

		go func() {
			time.Sleep(50 * time.Millisecond)
			_, _ = bus.UpdatePR(dir, layout.RoleBuilder, "job-cont", "forge-builder", 10, "https://github.com/pr/10", "", clk.Now())
		}()

		err = RunWorker(context.Background(), WorkerOptions{
			Dir:           dir,
			Role:          layout.RoleBuilder,
			DeliverTarget: target,
			Deliverer:     d,
			Once:          true,
			Now:           clk.Now,
			Out:           &out,
			Sleep: func(_ context.Context, dt time.Duration) bool {
				clk.Add(dt)
				time.Sleep(10 * time.Millisecond)
				return true
			},
		})
		require.NoError(t, err)

		foundClear := false
		for _, buf := range d.Buffers() {
			if string(buf) == "/clear" {
				foundClear = true
				break
			}
		}
		assert.False(t, foundClear, "continuous não deve enviar /clear")
	})

	// 3. reset_command vazio avisa uma vez
	t.Run("reset vazio avisa", func(t *testing.T) {
		dir := project(t, func(d string) {
			cfg := config.DefaultForgeConfig()
			cfg.SessionMode = config.SessionModeLive
			cfg.Context = config.ContextClear
			cfg.Live.StartupWaitSeconds = 0
			cfg.Live.ReadyQuietSeconds = 1
			p := cfg.CLIProfiles["claude"]
			p.ResetCommand = ""
			cfg.CLIProfiles["claude"] = p
			data, _ := json.Marshal(cfg)
			testutil.WriteFile(t, d, ".forge/config/forge.json", string(data))
		})
		job := bus.Job{
			ID:                "job-no-reset",
			Feature:           "auth-oauth",
			Title:             "Auth",
			SpecPath:          "sdd/discovery/77-auth-oauth/criteria.md",
			Type:              bus.TypeFeatureBuild,
			Priority:          bus.PriorityNormal,
			PR:                bus.PR{Branch: "feat/77-auth-oauth", BaseBranch: "main"},
			Attempt:           1,
			MaxReworkAttempts: 3,
			CreatedAt:         now,
		}
		_, _, err := bus.Enqueue(dir, layout.RoleBuilder, job, now)
		require.NoError(t, err)

		target := "=test:forge.1"
		d := &fakeTmuxDeliverer{}

		clk := newSafeClock(now)
		var out bytes.Buffer

		go func() {
			time.Sleep(50 * time.Millisecond)
			_, _ = bus.UpdatePR(dir, layout.RoleBuilder, "job-no-reset", "forge-builder", 10, "https://github.com/pr/10", "", clk.Now())
		}()

		err = RunWorker(context.Background(), WorkerOptions{
			Dir:           dir,
			Role:          layout.RoleBuilder,
			DeliverTarget: target,
			Deliverer:     d,
			Once:          true,
			Now:           clk.Now,
			Out:           &out,
			Sleep: func(_ context.Context, dt time.Duration) bool {
				clk.Add(dt)
				time.Sleep(10 * time.Millisecond)
				return true
			},
		})
		require.NoError(t, err)

		assert.Contains(t, out.String(), "sem reset_command: mantendo contexto contínuo")
	})
}
