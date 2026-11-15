package pipeline

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Task é uma task de uma iniciativa e suas dependências, lidas do cabeçalho da spec.
type Task struct {
	ID        string   // feat-70-04
	File      string   // caminho do arquivo
	DependsOn []string // ids de outras tasks da mesma pasta
	Level     int      // onda topológica (0 = sem dependência)
}

var (
	taskFileRe = regexp.MustCompile(`^(feat-[a-z0-9]+-[0-9]{2}[a-z]?)-`)
	taskRefRe  = regexp.MustCompile(`feat-[a-z0-9]+-[0-9]{2}[a-z]?`)
	dependsRe  = regexp.MustCompile(`^\*\*Depende de:\*\*\s*(.*)$`)
)

// PlanWaves ordena as tasks de featuresDir em lotes de execução: tasks sem dependência pendente
// formam uma onda, cada onda é dividida em lotes de até maxParallel tasks, e os lotes rodam em
// sequência. Ciclo ou dependência para task inexistente é erro.
func PlanWaves(featuresDir string, maxParallel int) ([][]Task, error) {
	if maxParallel < 1 {
		maxParallel = 1
	}
	entries, err := os.ReadDir(featuresDir)
	if err != nil {
		return nil, fmt.Errorf("ler %s: %w", featuresDir, err)
	}
	tasks := map[string]*Task{}
	for _, e := range entries {
		n := e.Name()
		m := taskFileRe.FindStringSubmatch(n)
		if e.IsDir() || m == nil || !strings.HasSuffix(n, ".md") {
			continue
		}
		t := &Task{ID: m[1], File: filepath.Join(featuresDir, n)}
		if _, dup := tasks[t.ID]; dup {
			return nil, fmt.Errorf("task %s duplicada em %s", t.ID, featuresDir)
		}
		deps, err := readDependsOn(t.File)
		if err != nil {
			return nil, err
		}
		t.DependsOn = deps
		tasks[t.ID] = t
	}
	if len(tasks) == 0 {
		return nil, fmt.Errorf("nenhuma task feat-*.md em %s", featuresDir)
	}
	for _, t := range tasks {
		for _, d := range t.DependsOn {
			if _, ok := tasks[d]; !ok {
				return nil, fmt.Errorf("%s depende de %s, que não existe em %s", t.ID, d, featuresDir)
			}
			if d == t.ID {
				return nil, fmt.Errorf("%s depende de si mesma", t.ID)
			}
		}
	}

	// Kahn por níveis.
	remaining := map[string]*Task{}
	for id, t := range tasks {
		remaining[id] = t
	}
	done := map[string]bool{}
	var batches [][]Task
	for level := 0; len(remaining) > 0; level++ {
		var ready []Task
		for id, t := range remaining {
			ok := true
			for _, d := range t.DependsOn {
				if !done[d] {
					ok = false
					break
				}
			}
			if ok {
				t.Level = level
				ready = append(ready, *t)
				_ = id
			}
		}
		if len(ready) == 0 {
			var stuck []string
			for id := range remaining {
				stuck = append(stuck, id)
			}
			sort.Strings(stuck)
			return nil, fmt.Errorf("ciclo de dependências entre: %s", strings.Join(stuck, ", "))
		}
		sort.Slice(ready, func(i, j int) bool { return ready[i].ID < ready[j].ID })
		for _, t := range ready {
			done[t.ID] = true
			delete(remaining, t.ID)
		}
		for i := 0; i < len(ready); i += maxParallel {
			end := i + maxParallel
			if end > len(ready) {
				end = len(ready)
			}
			batches = append(batches, ready[i:end])
		}
	}
	return batches, nil
}

func readDependsOn(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 256*1024)
	for i := 0; sc.Scan() && i < 40; i++ { // o cabeçalho fica nas primeiras linhas
		if m := dependsRe.FindStringSubmatch(strings.TrimSpace(sc.Text())); m != nil {
			seen := map[string]bool{}
			var deps []string
			for _, ref := range taskRefRe.FindAllString(m[1], -1) {
				if !seen[ref] {
					seen[ref] = true
					deps = append(deps, ref)
				}
			}
			return deps, nil
		}
	}
	return nil, sc.Err()
}

// FormatWaves descreve os lotes para o agente.
func FormatWaves(batches [][]Task) string {
	var b strings.Builder
	for i, batch := range batches {
		ids := make([]string, len(batch))
		for j, t := range batch {
			ids[j] = t.ID
		}
		mode := "em paralelo"
		if len(batch) == 1 {
			mode = "sozinha"
		}
		fmt.Fprintf(&b, "lote %d (onda %d, %s): %s\n", i+1, batch[0].Level, mode, strings.Join(ids, ", "))
	}
	return b.String()
}
