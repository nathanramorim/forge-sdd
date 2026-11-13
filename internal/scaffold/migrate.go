package scaffold

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/layout"
)

// Move descreve a movimentação de um artefato legado, com caminhos relativos a targetDir.
type Move struct {
	From string
	To   string
}

// Skip descreve um artefato que a migração não tocou, com o motivo.
type Skip struct {
	Path   string
	Reason string
}

// LayoutPlan é o plano da migração do layout v3 (discoveries soltos e sdd/features/)
// para o layout por iniciativa. Produzido sem escrever nada em disco.
type LayoutPlan struct {
	Moves   []Move
	Skipped []Skip
}

// Empty informa se não há nada a mover.
func (p LayoutPlan) Empty() bool { return len(p.Moves) == 0 }

var legacyDiscoveryPrefixes = []string{"discovery-", "criteria-", "plan-"}

// PlanLayoutMigration calcula o que a migração moveria para sdd/archive/ (ARQ-05: não escreve).
// Pastas de iniciativa (sdd/discovery/<id>-<nome>/) permanecem onde estão. Links simbólicos,
// destinos já existentes e arquivos fora do padrão são preservados e reportados (SEG-02).
func PlanLayoutMigration(targetDir string) (LayoutPlan, error) {
	var plan LayoutPlan

	// 1. Discoveries soltos: sdd/discovery/{discovery,criteria,plan}-*.md.
	discoveryDir := layout.Discovery(targetDir)
	entries, err := os.ReadDir(discoveryDir)
	if err != nil && !os.IsNotExist(err) {
		return plan, fmt.Errorf("migração: ler %s: %w", discoveryDir, err)
	}
	for _, e := range entries {
		name := e.Name()
		rel := path.Join(layout.SddDirName, layout.DiscoveryDirName, name)
		if e.IsDir() && layout.ValidInitiative(name) {
			continue // iniciativa no layout novo
		}
		if name == ".gitkeep" {
			continue
		}
		if e.IsDir() || !isLegacyDiscoveryFile(name) {
			plan.Skipped = append(plan.Skipped, Skip{Path: rel, Reason: "fora do padrão legado; preservado"})
			continue
		}
		plan.addMove(targetDir, rel, path.Join(layout.SddDirName, layout.ArchiveDirName, layout.DiscoveryDirName, name))
	}

	// 2. sdd/features/: cada entrada de primeiro nível vai para sdd/archive/features/.
	featuresDir := layout.Features(targetDir)
	entries, err = os.ReadDir(featuresDir)
	if err != nil && !os.IsNotExist(err) {
		return plan, fmt.Errorf("migração: ler %s: %w", featuresDir, err)
	}
	for _, e := range entries {
		name := e.Name()
		if name == ".gitkeep" {
			continue
		}
		plan.addMove(targetDir, path.Join(layout.SddDirName, layout.FeaturesDirName, name),
			path.Join(layout.SddDirName, layout.ArchiveDirName, layout.FeaturesDirName, name))
	}

	sort.Slice(plan.Moves, func(i, j int) bool { return plan.Moves[i].From < plan.Moves[j].From })
	return plan, nil
}

func isLegacyDiscoveryFile(name string) bool {
	if !strings.HasSuffix(name, ".md") {
		return false
	}
	for _, p := range legacyDiscoveryPrefixes {
		if strings.HasPrefix(name, p) {
			return true
		}
	}
	return false
}

// addMove valida a origem e o destino antes de registrar a movimentação.
func (p *LayoutPlan) addMove(targetDir, from, to string) {
	src := filepath.Join(targetDir, filepath.FromSlash(from))
	if reason := unsafeEntry(src); reason != "" {
		p.Skipped = append(p.Skipped, Skip{Path: from, Reason: reason})
		return
	}
	if _, err := os.Lstat(filepath.Join(targetDir, filepath.FromSlash(to))); err == nil {
		p.Skipped = append(p.Skipped, Skip{Path: from, Reason: "destino já existe em " + to + "; não sobrescrito"})
		return
	}
	p.Moves = append(p.Moves, Move{From: from, To: to})
}

// unsafeEntry recusa links simbólicos (na entrada ou dentro dela) e tipos que não sejam
// arquivo regular ou diretório (SEG-02). Devolve o motivo, ou vazio se for seguro.
func unsafeEntry(src string) string {
	info, err := os.Lstat(src)
	if err != nil {
		return "inacessível: " + err.Error()
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "link simbólico recusado (SEG-02)"
	}
	if !info.IsDir() {
		if !info.Mode().IsRegular() {
			return "tipo de arquivo não suportado"
		}
		return ""
	}
	reason := ""
	_ = filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			reason = "inacessível: " + err.Error()
			return fs.SkipAll
		}
		if d.Type()&fs.ModeSymlink != 0 {
			reason = "contém link simbólico (SEG-02)"
			return fs.SkipAll
		}
		if !d.IsDir() && !d.Type().IsRegular() {
			reason = "contém tipo de arquivo não suportado"
			return fs.SkipAll
		}
		return nil
	})
	return reason
}

// ApplyLayoutMigration copia as origens para sdd/backups/<timestamp>/ e então move cada
// artefato para sdd/archive/. Sem movimentos, não cria backup nem altera nada (idempotente).
// Devolve o diretório de backup criado (vazio se não houve movimentos).
func ApplyLayoutMigration(targetDir string, plan LayoutPlan) (string, error) {
	if plan.Empty() {
		return "", nil
	}
	backupRoot := filepath.Join(layout.Root(targetDir), "backups", time.Now().UTC().Format("20060102T150405Z"))

	for _, m := range plan.Moves {
		src := filepath.Join(targetDir, filepath.FromSlash(m.From))
		if err := copyTree(src, filepath.Join(backupRoot, filepath.FromSlash(m.From))); err != nil {
			return backupRoot, fmt.Errorf("migração: backup de %s: %w", m.From, err)
		}
	}
	for _, m := range plan.Moves {
		src := filepath.Join(targetDir, filepath.FromSlash(m.From))
		dst := filepath.Join(targetDir, filepath.FromSlash(m.To))
		if _, err := os.Lstat(dst); err == nil {
			return backupRoot, fmt.Errorf("migração: destino %s já existe; nada foi sobrescrito", m.To)
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return backupRoot, fmt.Errorf("migração: criar %s: %w", filepath.Dir(dst), err)
		}
		if err := os.Rename(src, dst); err != nil {
			return backupRoot, fmt.Errorf("migração: mover %s para %s: %w", m.From, m.To, err)
		}
	}

	// sdd/features/ vazio deixa de existir; sdd/discovery/ nunca é removido.
	removeIfEmpty(layout.Features(targetDir))
	return backupRoot, nil
}

func removeIfEmpty(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.Name() != ".gitkeep" {
			return
		}
	}
	_ = os.RemoveAll(dir)
}

// copyTree copia arquivos regulares e diretórios, sem seguir links.
func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("tipo de arquivo não suportado: %s", p)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		in, err := os.Open(p)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, in); err != nil {
			_ = out.Close()
			return err
		}
		return out.Close()
	})
}

// ---- artefatos de agente obsoletos (feat 69-06) ----

//go:embed all:legacy
var legacyFS embed.FS

// ObsoletePlan separa os artefatos obsoletos em removíveis (idênticos ao template anterior)
// e preservados (editados pelo usuário).
type ObsoletePlan struct {
	Remove   []string
	Preserve []string
}

// legacyAgentArtifacts mapeia cada template anterior ao caminho que ele gerava no projeto.
func legacyAgentArtifacts() (map[string]string, error) {
	out := map[string]string{}
	for _, kind := range []string{"commands", "roles"} {
		entries, err := legacyFS.ReadDir(path.Join("legacy", kind))
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			name := strings.TrimSuffix(e.Name(), ".tmpl")
			out[path.Join("legacy", kind, e.Name())] = path.Join(".agents", kind, name)
		}
	}
	return out, nil
}

// PlanObsoleteAgentArtifacts decide o que fazer com .agents/commands/*.md e .agents/roles/*.md:
// só é removível o arquivo cujo conteúdo é idêntico ao do template anterior renderizado com cfg;
// qualquer diferença significa edição do usuário e o arquivo é preservado (SEG-02).
func PlanObsoleteAgentArtifacts(targetDir string, cfg config.Config) (ObsoletePlan, error) {
	var plan ObsoletePlan
	artifacts, err := legacyAgentArtifacts()
	if err != nil {
		return plan, err
	}
	keys := make([]string, 0, len(artifacts))
	for k := range artifacts {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, tmplPath := range keys {
		rel := artifacts[tmplPath]
		abs := filepath.Join(targetDir, filepath.FromSlash(rel))
		info, err := os.Lstat(abs)
		if err != nil {
			continue // não existe no projeto
		}
		if !info.Mode().IsRegular() {
			plan.Preserve = append(plan.Preserve, rel)
			continue
		}
		onDisk, err := os.ReadFile(abs)
		if err != nil {
			plan.Preserve = append(plan.Preserve, rel)
			continue
		}
		raw, err := legacyFS.ReadFile(tmplPath)
		if err != nil {
			return plan, err
		}
		tmpl, err := template.New(tmplPath).Parse(string(raw))
		if err != nil {
			return plan, fmt.Errorf("template legado %s: %w", tmplPath, err)
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, cfg); err != nil {
			return plan, fmt.Errorf("renderizar template legado %s: %w", tmplPath, err)
		}
		if bytes.Equal(onDisk, buf.Bytes()) {
			plan.Remove = append(plan.Remove, rel)
		} else {
			plan.Preserve = append(plan.Preserve, rel)
		}
	}
	return plan, nil
}

// RemoveObsoleteAgentArtifacts remove os artefatos idênticos ao template anterior e, se as
// pastas ficarem vazias, remove também .agents/commands e .agents/roles.
func RemoveObsoleteAgentArtifacts(targetDir string, cfg config.Config) (ObsoletePlan, error) {
	plan, err := PlanObsoleteAgentArtifacts(targetDir, cfg)
	if err != nil {
		return plan, err
	}
	for _, rel := range plan.Remove {
		if err := os.Remove(filepath.Join(targetDir, filepath.FromSlash(rel))); err != nil && !os.IsNotExist(err) {
			return plan, fmt.Errorf("remover obsoleto %s: %w", rel, err)
		}
	}
	for _, d := range []string{"commands", "roles"} {
		dir := filepath.Join(targetDir, ".agents", d)
		if entries, err := os.ReadDir(dir); err == nil && len(entries) == 0 {
			_ = os.Remove(dir)
		}
	}
	return plan, nil
}
