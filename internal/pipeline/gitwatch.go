package pipeline

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	gitWatchTimeout = 3 * time.Second
	maxWatchLines   = 5
)

// GitFunc roda um comando git no diretório e devolve a saída; injetável nos testes.
type GitFunc func(ctx context.Context, dir string, args ...string) (string, error)

// ErrGitSlow indica que o git passou do teto de tempo (repositório enorme ou travado).
var ErrGitSlow = errors.New("git passou do teto de tempo")

const maxGitOutput = 512 * 1024

// defaultGit roda o git sem pegar o lock opcional do index (não atrapalha o git do agente), com
// teto de tempo e de saída.
func defaultGit(ctx context.Context, dir string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, gitWatchTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", append([]string{"--no-optional-locks", "-C", dir}, args...)...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	if err := cmd.Start(); err != nil {
		return "", err
	}
	data, _ := io.ReadAll(io.LimitReader(stdout, maxGitOutput))
	_, _ = io.Copy(io.Discard, stdout) // esvazia o resto para o git terminar
	werr := cmd.Wait()
	if ctx.Err() == context.DeadlineExceeded {
		return "", ErrGitSlow
	}
	return string(data), werr
}

// GitWatcher acompanha o `git status` do diretório do job e descreve o que mudou, para qualquer
// CLI de agente (inclusive as que não imprimem nada). Nunca falha o job: sem git, não diz nada.
type GitWatcher struct {
	Dir      string
	Git      GitFunc
	Now      func() time.Time
	Activity func() time.Time // último evento do agente; nil = ignora
	Beat     time.Duration    // intervalo do batimento (padrão 30 s)

	mu                            sync.Mutex
	started, lastChange, lastBeat time.Time
	init, warned                  bool
	prefix                        string // caminho de Dir dentro do repositório (git rev-parse --show-prefix)
	base, prev                    map[string]string
}

// NewGitWatcher cria o acompanhamento; o início do job é o instante da criação.
func NewGitWatcher(dir string, now func() time.Time, activity func() time.Time) *GitWatcher {
	if now == nil {
		now = time.Now
	}
	n := now()
	return &GitWatcher{Dir: dir, Git: defaultGit, Now: now, Activity: activity, Beat: 30 * time.Second, started: n, lastChange: n, lastBeat: n}
}

var watchIgnore = []string{".forge/", "sdd/.runtime/", "sdd/.metrics/"}

func parsePorcelain(out, prefix string) map[string]string {
	m := map[string]string{}
	for _, l := range strings.Split(out, "\n") {
		if len(l) < 4 {
			continue
		}
		path := strings.TrimSpace(l[3:])
		if i := strings.Index(path, " -> "); i >= 0 {
			path = path[i+4:]
		}
		path = strings.Trim(path, `"`)
		if prefix != "" {
			if !strings.HasPrefix(path, prefix) {
				continue // fora do diretório do job
			}
			path = strings.TrimPrefix(path, prefix)
		}
		skip := false
		for _, p := range watchIgnore {
			if strings.HasPrefix(path, p) {
				skip = true
			}
		}
		if !skip && path != "" {
			m[path] = l[:2]
		}
	}
	return m
}

// Count diz quantos arquivos diferem do estado do início do job.
func (w *GitWatcher) Count() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.count()
}

func (w *GitWatcher) count() int {
	n := 0
	for p, st := range w.prev {
		if w.base[p] != st {
			n++
		}
	}
	return n
}

// Poll faz uma consulta e devolve as linhas a mostrar (vazio quando nada mudou).
func (w *GitWatcher) Poll(ctx context.Context) []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.init {
		if p, err := w.Git(ctx, w.Dir, "rev-parse", "--show-prefix"); err == nil {
			w.prefix = strings.TrimSpace(p)
		}
	}
	out, err := w.Git(ctx, w.Dir, "status", "--porcelain", "-uall")
	if err != nil {
		if errors.Is(err, ErrGitSlow) && !w.warned {
			w.warned = true
			return []string{"  (acompanhamento por git lento neste repositório: sigo só com os eventos do agente)"}
		}
		return nil
	}
	cur := parsePorcelain(out, w.prefix)
	now := w.Now()
	if !w.init {
		w.init, w.base, w.prev = true, cur, cur
		return nil
	}
	var changed []string
	for p, st := range cur {
		if w.prev[p] != st {
			changed = append(changed, p)
		}
	}
	sort.Strings(changed)
	w.prev = cur

	var lines []string
	for i, p := range changed {
		if i >= maxWatchLines {
			lines = append(lines, fmt.Sprintf("  ✎ (+%d arquivos)", len(changed)-maxWatchLines))
			break
		}
		lines = append(lines, "  ✎ "+p)
	}
	if len(changed) > 0 {
		w.lastChange, w.lastBeat = now, now
		return lines
	}
	last := w.lastChange
	if w.Activity != nil {
		if a := w.Activity(); a.After(last) {
			last = a
		}
	}
	if now.Sub(last) >= w.Beat && now.Sub(w.lastBeat) >= w.Beat {
		w.lastBeat = now
		return []string{fmt.Sprintf("  … trabalhando há %s · %d arquivos alterados", FormatElapsed(now.Sub(w.started)), w.count())}
	}
	return nil
}

// Run consulta a cada intervalo até o contexto acabar, entregando as linhas a emit.
func (w *GitWatcher) Run(ctx context.Context, every time.Duration, emit func(string)) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			for _, l := range w.Poll(ctx) {
				emit(l)
			}
		}
	}
}

// FormatElapsed escreve uma duração curta: 45s, 2m10s, 1h05m.
func FormatElapsed(d time.Duration) string {
	d = d.Round(time.Second)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm%02ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%02dm", int(d.Hours()), int(d.Minutes())%60)
}
