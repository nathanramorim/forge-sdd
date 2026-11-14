package pipeline

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// MaxUsageLimitPause é o teto da pausa: passou disso o job falha com o motivo.
	MaxUsageLimitPause = 6 * time.Hour
	// UnknownUsageLimitPause é a espera quando a mensagem não diz quando o limite volta.
	UnknownUsageLimitPause = 30 * time.Minute
)

var (
	usageLimitRe = regexp.MustCompile(`(?i)(hit your [^.\n]*limit|usage limit|rate[ _-]?limit|quota (?:exceeded|exhausted)|too many requests|resource[ _]exhausted)`)
	// "resets 10:40pm (America/Sao_Paulo)", "resets at 22:40", "reset at 10pm".
	resetClockRe = regexp.MustCompile(`(?i)resets?\s+(?:at\s+)?(\d{1,2})(?::(\d{2}))?\s*(am|pm)?(?:\s*\(([A-Za-z_]+(?:/[A-Za-z_+-]+)+|UTC)\))?`)
	// "try again in 30 minutes", "in 2 hours", "after 45 seconds".
	resetInRe = regexp.MustCompile(`(?i)\b(?:in|after)\s+(\d+)\s*(second|minute|hour)s?\b`)
)

// UsageLimit descreve um limite de uso detectado na saída de um agente.
type UsageLimit struct {
	Message string    // a linha da mensagem, já saneada
	ResetAt time.Time // quando o worker retoma
	Known   bool      // o horário veio da mensagem
	Capped  bool      // passa do teto: o job falha em vez de pausar
}

// DetectUsageLimit procura uma mensagem de limite de uso nas últimas linhas úteis da saída. Só vale
// com saída diferente de zero (quem chama confere), para um texto do código ou da conversa que cite
// "rate limit" não pausar um job que terminou bem.
func DetectUsageLimit(tail []string, now time.Time) (UsageLimit, bool) {
	for i := len(tail) - 1; i >= 0; i-- {
		line := tail[i]
		if !usageLimitRe.MatchString(line) {
			continue
		}
		ul := UsageLimit{Message: strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "✗")), ResetAt: now.Add(UnknownUsageLimitPause)}
		if at, ok := parseResetClock(line, now); ok {
			ul.ResetAt, ul.Known = at, true
		} else if d, ok := parseResetIn(line); ok {
			ul.ResetAt, ul.Known = now.Add(d), true
		}
		ul.Capped = ul.ResetAt.Sub(now) > MaxUsageLimitPause
		return ul, true
	}
	return UsageLimit{}, false
}

func parseResetClock(line string, now time.Time) (time.Time, bool) {
	m := resetClockRe.FindStringSubmatch(line)
	if m == nil {
		return time.Time{}, false
	}
	h, _ := strconv.Atoi(m[1])
	min := 0
	if m[2] != "" {
		min, _ = strconv.Atoi(m[2])
	}
	switch strings.ToLower(m[3]) {
	case "pm":
		if h < 12 {
			h += 12
		}
	case "am":
		if h == 12 {
			h = 0
		}
	}
	if h > 23 || min > 59 {
		return time.Time{}, false
	}
	loc := now.Location()
	if m[4] != "" {
		l, err := time.LoadLocation(m[4])
		if err != nil {
			return time.Time{}, false
		}
		loc = l
	}
	n := now.In(loc)
	at := time.Date(n.Year(), n.Month(), n.Day(), h, min, 0, 0, loc)
	if !at.After(n) {
		at = at.Add(24 * time.Hour)
	}
	return at, true
}

func parseResetIn(line string) (time.Duration, bool) {
	m := resetInRe.FindStringSubmatch(line)
	if m == nil {
		return 0, false
	}
	n, _ := strconv.Atoi(m[1])
	switch strings.ToLower(m[2]) {
	case "second":
		return time.Duration(n) * time.Second, true
	case "minute":
		return time.Duration(n) * time.Minute, true
	}
	return time.Duration(n) * time.Hour, true
}

// pauseUntil espera até `until`, mostrando o tempo restante a cada minuto; devolve false se o
// contexto acabar antes (Ctrl+C). sleep e now são injetáveis nos testes.
func pauseUntil(ctx context.Context, until time.Time, agent string, now func() time.Time, sleep func(context.Context, time.Duration) bool, say func(string)) bool {
	for {
		left := until.Sub(now())
		if left <= 0 {
			return true
		}
		say(fmt.Sprintf("⏸ limite de uso do %s: retomo às %s (faltam %s)", agent, until.Format("15:04"), FormatElapsed(left)))
		step := time.Minute
		if left < step {
			step = left
		}
		if !sleep(ctx, step) {
			return false
		}
	}
}

func realSleep(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
