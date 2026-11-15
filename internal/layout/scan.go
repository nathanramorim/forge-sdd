package layout

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// WorkTreeLabel é o valor de InitiativeInfo.Branch quando o estado vem da árvore de trabalho.
const WorkTreeLabel = "(árvore de trabalho)"

// InitiativeInfo descreve uma iniciativa encontrada em sdd/discovery/<id>/.
type InitiativeInfo struct {
	ID      string         // <id>-<nome>
	Branch  string         // branch (ou "(árvore de trabalho)") de onde o estado foi lido
	Status  string         // resumo do bloco "Status" de status.md
	Next    string         // primeira linha de "Próximo passo" de status.md
	Tasks   map[string]int // contagem de tasks da fila por situação
	Current bool           // true quando é a iniciativa da branch atual
}

// ListInitiatives devolve as iniciativas de dir. Sem allBranches lê apenas a árvore de
// trabalho; com allBranches une as de todas as branches locais e remotas, sem checkout.
func ListInitiatives(dir string, allBranches bool) ([]InitiativeInfo, error) {
	current, _, curErr := Current(dir)
	if curErr != nil && !errors.Is(curErr, errNotRepo) {
		return nil, curErr
	}
	workTree := map[string]InitiativeInfo{}

	entries, err := os.ReadDir(Discovery(dir))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("layout: ler %s: %w", Discovery(dir), err)
	}
	for _, e := range entries {
		if !e.IsDir() || !ValidInitiative(e.Name()) {
			continue
		}
		base := filepath.Join(Discovery(dir), e.Name())
		status, _ := os.ReadFile(filepath.Join(base, StatusFileName))
		queue, _ := os.ReadFile(filepath.Join(base, QueueFileName))
		workTree[e.Name()] = buildInitiative(e.Name(), WorkTreeLabel, status, queue)
	}

	chosen := map[string]InitiativeInfo{}
	for id, it := range workTree {
		chosen[id] = it
	}

	if allBranches {
		refs, err := listRefs(dir)
		if err != nil {
			return nil, err
		}
		owners := map[string]InitiativeInfo{}
		others := map[string]InitiativeInfo{}
		for _, ref := range refs {
			ids, err := idsInRef(dir, ref)
			if err != nil {
				continue // ref sem sdd/discovery
			}
			for _, id := range ids {
				bucket := others
				if refOwnsInitiative(ref, id) {
					bucket = owners
				}
				if _, done := bucket[id]; done {
					continue // refs locais vêm antes das remotas
				}
				base := SddDirName + "/" + DiscoveryDirName + "/" + id + "/"
				status, _ := gitShow(dir, ref, base+StatusFileName)
				queue, _ := gitShow(dir, ref, base+QueueFileName)
				bucket[id] = buildInitiative(id, ref, []byte(status), []byte(queue))
			}
		}
		for id, it := range owners {
			if id == current && hasKey(workTree, id) {
				continue // a árvore de trabalho é a fonte mais recente da branch atual
			}
			chosen[id] = it
		}
		for id, it := range others {
			if _, ok := chosen[id]; !ok {
				chosen[id] = it
			}
		}
	}

	out := make([]InitiativeInfo, 0, len(chosen))
	for id, it := range chosen {
		it.Current = id == current
		out = append(out, it)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func hasKey(m map[string]InitiativeInfo, k string) bool { _, ok := m[k]; return ok }

// statusKeyRe limita as chaves de situação lidas de queue.json (dado não confiável, SEG-01).
var statusKeyRe = regexp.MustCompile(`^[a-z][a-z_]{0,31}$`)

// maxStateFileBytes é o limite de leitura de status.md e queue.json vindos de refs (SEG-01).
const maxStateFileBytes = 64 * 1024

// sanitizeText troca caracteres de controle (inclusive ESC e C1) por espaço, para que texto lido
// de refs não confiáveis não injete sequências de terminal nem quebre a saída (SEG-01).
func sanitizeText(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f):
			return ' '
		case (r >= 0x202a && r <= 0x202e) || (r >= 0x2066 && r <= 0x2069) || r == 0x200e || r == 0x200f:
			return ' ' // controles de direção bidi (CVE-2021-42574)
		}
		return r
	}, s)
}

func buildInitiative(id, branch string, status, queue []byte) InitiativeInfo {
	it := InitiativeInfo{ID: id, Branch: sanitizeText(branch), Tasks: map[string]int{}}
	it.Status, it.Next = summarizeStatus(string(status))
	var q struct {
		Items []struct {
			Status string `json:"status"`
		} `json:"items"`
	}
	if json.Unmarshal(queue, &q) == nil {
		for _, item := range q.Items {
			key := item.Status
			if !statusKeyRe.MatchString(key) {
				key = "outro"
			}
			it.Tasks[key]++
		}
	}
	return it
}

// summarizeStatus extrai o resumo do bloco "## Status" e a primeira linha de "## Proximo passo".
func summarizeStatus(text string) (status, next string) {
	lines := strings.Split(text, "\n")
	section := ""
	inFence := false
	var parts []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			section = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(trimmed, "## ")))
			section = strings.NewReplacer("ó", "o", "é", "e").Replace(section)
			inFence = false
			continue
		}
		switch {
		case section == "status":
			if strings.HasPrefix(trimmed, "```") {
				inFence = !inFence
				continue
			}
			if trimmed != "" && (inFence || len(parts) == 0) {
				parts = append(parts, trimmed)
			}
		case strings.HasPrefix(section, "proximo passo") && next == "" && trimmed != "":
			next = trimmed
		}
	}
	return truncate(sanitizeText(strings.Join(parts, " | ")), 100), truncate(sanitizeText(next), 100)
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

// ---- git (somente leitura, argv explícito: SEG-03) ----

// errNotRepo identifica a falha "não é um repositório git", que o chamador pode tratar como
// ausência de refs; qualquer outra falha do git deve ser propagada.
var errNotRepo = errors.New("não é um repositório git")

// Git executa o git em dir com argv explícito, sem shell (SEG-03), e devolve a saída padrão.
// "Não é um repositório" é reconhecível por errors.Is(err, ErrNotRepo).
func Git(dir string, args ...string) (string, error) { return gitOut(dir, args...) }

// ErrNotRepo é o erro devolvido por Git fora de um repositório.
var ErrNotRepo = errNotRepo

func gitOut(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	// Mensagens em inglês: a detecção de "não é um repositório" não pode depender do idioma do git.
	cmd.Env = append(os.Environ(), "LC_ALL=C", "LANGUAGE=C")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if strings.Contains(msg, "not a git repository") {
			return "", fmt.Errorf("layout: git %s: %w (%s)", strings.Join(args, " "), errNotRepo, msg)
		}
		return "", fmt.Errorf("layout: git %s: %w (%s)", strings.Join(args, " "), err, msg)
	}
	return string(out), nil
}

// safeRef recusa nomes de ref que poderiam ser lidos como opção do git ou conter controle (SEG-01).
func safeRef(ref string) bool {
	if ref == "" || strings.HasPrefix(ref, "-") {
		return false
	}
	for _, r := range ref {
		if r < 0x20 || r == 0x7f || r == ' ' || r == ':' || r == '\\' {
			return false
		}
	}
	return true
}

func listRefs(dir string) ([]string, error) {
	out, err := gitOut(dir, "for-each-ref", "--format=%(refname:short)", "refs/heads", "refs/remotes")
	if err != nil {
		return nil, err
	}
	var refs []string
	for _, line := range strings.Split(out, "\n") {
		ref := strings.TrimSpace(line)
		if ref == "" || strings.HasSuffix(ref, "/HEAD") || !safeRef(ref) {
			continue
		}
		refs = append(refs, ref)
	}
	return refs, nil
}

// idsInRef lista as pastas <id>-<nome> de sdd/discovery em ref.
func idsInRef(dir, ref string) ([]string, error) {
	names, err := namesInRef(dir, ref, path.Join(SddDirName, DiscoveryDirName)+"/")
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, n := range names {
		if ValidInitiative(n) {
			ids = append(ids, n)
		}
	}
	return ids, nil
}

// namesInRef devolve os nomes (sem o prefixo) das entradas diretas de path em ref.
func namesInRef(dir, ref, path string) ([]string, error) {
	if !safeRef(ref) {
		return nil, fmt.Errorf("layout: ref inválida %q", ref)
	}
	out, err := gitOut(dir, "ls-tree", "--name-only", ref, path)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		names = append(names, strings.TrimPrefix(line, path))
	}
	return names, nil
}

// treeNamesInRef devolve o nome base de todas as entradas (arquivos e pastas, recursivo) de paths em ref.
func treeNamesInRef(dir, ref string, paths ...string) ([]string, error) {
	if !safeRef(ref) {
		return nil, fmt.Errorf("layout: ref inválida %q", ref)
	}
	args := append([]string{"ls-tree", "-r", "-t", "--name-only", ref, "--"}, paths...)
	out, err := gitOut(dir, args...)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, line := range strings.Split(out, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			names = append(names, path.Base(line))
		}
	}
	return names, nil
}

// gitShow lê um arquivo de ref, recusando blobs acima de maxStateFileBytes (SEG-01).
func gitShow(dir, ref, path string) (string, error) {
	if !safeRef(ref) {
		return "", fmt.Errorf("layout: ref inválida %q", ref)
	}
	sizeOut, err := gitOut(dir, "cat-file", "-s", ref+":"+path)
	if err != nil {
		return "", err
	}
	size, convErr := strconv.ParseInt(strings.TrimSpace(sizeOut), 10, 64)
	if convErr != nil {
		return "", fmt.Errorf("layout: tamanho inválido de %s:%s: %w", ref, path, convErr)
	}
	if size > maxStateFileBytes {
		return "", fmt.Errorf("layout: %s:%s tem %d bytes (limite %d)", ref, path, size, maxStateFileBytes)
	}
	return gitOut(dir, "show", ref+":"+path)
}

// idOfRef devolve a id de iniciativa que a ref nomeia, com ou sem prefixo de remoto.
func idOfRef(ref string) (string, bool) {
	if id, ok := FromBranch(ref); ok {
		return id, true
	}
	if _, rest, found := strings.Cut(ref, "/"); found {
		return FromBranch(rest)
	}
	return "", false
}

// refOwnsInitiative informa se a ref é a branch da iniciativa (<prefixo>/<id>), com ou sem remoto.
func refOwnsInitiative(ref, id string) bool {
	got, ok := idOfRef(ref)
	return ok && got == id
}

// ---- próximo ID ----

var (
	leadingNumber = regexp.MustCompile(`^(?:(?:discovery|criteria|plan|feat|fix)-)?([0-9]{1,6})(?:[^0-9]|$)`)
	workitemID    = regexp.MustCompile(`^[a-z0-9]{1,32}$`)
)

// NextID calcula o próximo identificador conforme a convenção de nomes ("sequencial", "hash"
// ou "workitem"), considerando a árvore de trabalho e todas as branches para não colidir.
// Em "workitem", o id vem de workitem (obrigatório).
func NextID(dir, convention, workitem string) (string, error) {
	known, err := knownNames(dir)
	if err != nil {
		return "", err
	}

	switch convention {
	case "workitem":
		id := strings.ToLower(workitem)
		if !workitemID.MatchString(id) {
			return "", fmt.Errorf("layout: convenção workitem exige --workitem alfanumérico (até 32 caracteres), recebido %q", workitem)
		}
		for _, n := range known {
			if tok, _, ok := strings.Cut(stripKind(n), "-"); ok && tok == id {
				return "", fmt.Errorf("layout: workitem %q já está em uso por %q", id, n)
			}
		}
		return id, nil
	case "hash":
		taken := map[string]bool{}
		for _, n := range known {
			if tok, _, ok := strings.Cut(stripKind(n), "-"); ok {
				taken[tok] = true
			}
		}
		for i := 0; i < 64; i++ {
			buf := make([]byte, 2)
			if _, err := rand.Read(buf); err != nil {
				return "", fmt.Errorf("layout: gerar hash: %w", err)
			}
			h := hex.EncodeToString(buf)
			if !taken[h] {
				return h, nil
			}
		}
		return "", errors.New("layout: não foi possível gerar um hash livre")
	case "", "sequencial":
		max := 0
		for _, n := range known {
			if m := leadingNumber.FindStringSubmatch(n); m != nil {
				if v, err := strconv.Atoi(m[1]); err == nil && v > max {
					max = v
				}
			}
		}
		return fmt.Sprintf("%02d", max+1), nil
	default:
		return "", fmt.Errorf("layout: convenção de nomes desconhecida %q (use sequencial, hash ou workitem)", convention)
	}
}

func stripKind(n string) string {
	for _, k := range []string{"discovery-", "criteria-", "plan-", "feat-", "fix-"} {
		if strings.HasPrefix(n, k) {
			return strings.TrimPrefix(n, k)
		}
	}
	return n
}

// knownNames reúne nomes de pastas e arquivos de sdd/discovery, sdd/features e sdd/archive na
// árvore de trabalho e em todas as refs, mais os nomes das próprias branches.
func knownNames(dir string) ([]string, error) {
	seen := map[string]bool{}
	add := func(n string) {
		if n != "" {
			seen[n] = true
		}
	}
	// Percorre recursivamente: fixes e features aninhados em pastas de iniciativa também contam.
	for _, d := range []string{Discovery(dir), Features(dir), Archive(dir)} {
		if _, err := os.Stat(d); err != nil {
			continue
		}
		_ = filepath.WalkDir(d, func(p string, e os.DirEntry, err error) error {
			if err == nil && p != d {
				add(e.Name())
			}
			return nil
		})
	}
	refs, err := listRefs(dir)
	if err != nil {
		if !errors.Is(err, errNotRepo) {
			return nil, err // falha real do git: não arriscar um ID colidente
		}
		// Fora de um repositório git só a árvore de trabalho vale.
		var names []string
		for n := range seen {
			names = append(names, n)
		}
		return names, nil
	}
	for _, ref := range refs {
		if id, ok := idOfRef(ref); ok {
			add(id)
		}
		names, err := treeNamesInRef(dir, ref,
			path.Join(SddDirName, DiscoveryDirName), path.Join(SddDirName, FeaturesDirName), path.Join(SddDirName, ArchiveDirName))
		if err != nil {
			return nil, err
		}
		for _, n := range names {
			add(n)
		}
	}
	out := make([]string, 0, len(seen))
	for n := range seen {
		out = append(out, n)
	}
	sort.Strings(out)
	return out, nil
}
