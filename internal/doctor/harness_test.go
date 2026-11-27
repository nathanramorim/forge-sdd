package doctor

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const veredito = ".agents/skills/sdd-governanca/references/veredito.md"

func escreve(t *testing.T, dir, rel, conteudo string) {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(conteudo), 0o644); err != nil {
		t.Fatal(err)
	}
}

// projeto monta um projeto mínimo e saudável, com a skill de Governança real do repositório.
func projeto(t *testing.T, versao string) string {
	t.Helper()
	dir := t.TempDir()
	real, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(veredito)))
	if err != nil {
		t.Fatal(err)
	}
	escreve(t, dir, "sdd/.sddrc", `{"project":"p","version":"`+versao+`"}`)
	escreve(t, dir, "sdd/.sdd-version", versao+"\n")
	escreve(t, dir, veredito, string(real))
	for _, e := range []string{"sdd-concepcao", "sdd-planejamento", "sdd-build", "sdd-governanca", "sdd-operacao"} {
		escreve(t, dir, ".agents/skills/"+e+"/SKILL.md", "# "+e)
	}
	escreve(t, dir, ".agents/rules/sdd-seguranca.md", "# regra")
	return dir
}

func deps(versao string) Deps {
	return Deps{
		Versao:  versao,
		Usuario: "eu",
		Runner:  func(string, ...string) ([]byte, error) { return nil, nil },
	}
}

func achar(t *testing.T, r Relatorio, id string) Checagem {
	t.Helper()
	for _, c := range r.Checagens {
		if c.ID == id {
			return c
		}
	}
	t.Fatalf("checagem %q ausente", id)
	return Checagem{}
}

func TestDiagnoseJSON(t *testing.T) {
	dir := projeto(t, "5.0.0")
	r := Diagnose(dir, deps("5.0.0"))
	if err := ValidateRelatorio(r); err != nil {
		t.Fatal(err)
	}
	if r.Falhas != 0 || r.Avisos != 0 || r.ExitCode() != 0 {
		t.Fatalf("projeto saudável deveria estar todo ok: %+v", r)
	}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) > 2048 {
		t.Errorf("saída com %d bytes, limite 2048", len(data))
	}
	var volta Relatorio
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&volta); err != nil {
		t.Fatalf("JSON não decodifica de volta: %v", err)
	}
	if volta.Versao != 1 || len(volta.Checagens) != len(r.Checagens) {
		t.Errorf("JSON instável: %+v", volta)
	}
}

func TestDiagnoseLimiteDe2KBComMuitosProblemas(t *testing.T) {
	dir := projeto(t, "5.0.0")
	for i := 0; i < 30; i++ {
		escreve(t, dir, "sdd/discovery/97-x/verdicts/verdict-feat-97-"+string(rune('a'+i%26))+string(rune('a'+i/26))+".json", `{"task_id":"x"}`)
		escreve(t, dir, ".claude/worktrees/agent-"+strings.Repeat("a", 20)+string(rune('a'+i%26))+string(rune('a'+i/26))+"/x", "")
	}
	data, _ := json.Marshal(Diagnose(dir, deps("5.0.0")))
	if len(data) > 2048 {
		t.Fatalf("saída com %d bytes, limite 2048", len(data))
	}
}

func TestDiagnoseSemEfeitoColateral(t *testing.T) {
	dir := projeto(t, "5.0.0")
	escreve(t, dir, ".forge/runtime/worktrees/job-1/feat-1/x", "x")
	escreve(t, dir, ".claude/worktrees/agent-1/x", "x")
	escreve(t, dir, "sdd/discovery/97-x/verdicts/v.json", `{"task_id":"x"}`)
	antes := arvore(t, dir)
	Diagnose(dir, deps("5.0.0"))
	depois := arvore(t, dir)
	if antes != depois {
		t.Errorf("o diagnóstico alterou a árvore:\nantes:\n%s\ndepois:\n%s", antes, depois)
	}
}

func arvore(t *testing.T, dir string) string {
	t.Helper()
	var b strings.Builder
	err := filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		b.WriteString(p + " " + info.ModTime().String() + "\n")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return b.String()
}

func TestDiagnoseDetectaWorktreeOrfao(t *testing.T) {
	dir := projeto(t, "5.0.0")
	escreve(t, dir, ".forge/runtime/worktrees/job-1/feat-1/x", "x")
	escreve(t, dir, ".claude/worktrees/agent-1/x", "x")
	c := achar(t, Diagnose(dir, deps("5.0.0")), "worktrees")
	if c.Status != StatusAviso {
		t.Fatalf("status = %s, esperado aviso", c.Status)
	}
	for _, quer := range []string{"job-1/feat-1", ".claude/worktrees/agent-1"} {
		if !strings.Contains(c.Detalhe, quer) {
			t.Errorf("detalhe %q não cita %q", c.Detalhe, quer)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, ".claude", "worktrees", "agent-1", "x")); err != nil {
		t.Errorf("o doctor não pode remover worktrees: %v", err)
	}
}

func TestDiagnoseContratoDivergenteFalha(t *testing.T) {
	dir := projeto(t, "5.0.0")
	escreve(t, dir, veredito, "# Veredito\n```json\n{\"veredito\":\"conforme\",\"task_id\":\"x\"}\n```\n")
	r := Diagnose(dir, deps("5.0.0"))
	if c := achar(t, r, "contrato-skill-binario"); c.Status != StatusFalha || c.Sugestao == "" {
		t.Fatalf("contrato divergente deveria falhar com sugestão: %+v", c)
	}
	if r.ExitCode() != ExitFalha {
		t.Errorf("exit = %d, esperado %d", r.ExitCode(), ExitFalha)
	}
}

func TestDiagnoseSemSkillDeGovernancaFalha(t *testing.T) {
	dir := projeto(t, "5.0.0")
	if err := os.Remove(filepath.Join(dir, filepath.FromSlash(veredito))); err != nil {
		t.Fatal(err)
	}
	if c := achar(t, Diagnose(dir, deps("5.0.0")), "contrato-skill-binario"); c.Status != StatusFalha {
		t.Fatalf("status = %s, esperado falha", c.Status)
	}
}

func TestDiagnoseVersaoDoBinario(t *testing.T) {
	casos := []struct {
		nome, binario, projeto, quer string
	}{
		{"binário antigo", "4.9.1-beta", "5.0.0", StatusFalha},
		{"binário beta", "5.0.0-beta", "5.0.0", StatusFalha},
		{"igual", "5.0.0", "5.0.0", StatusOK},
		{"projeto antigo", "5.0.0", "4.5.0", StatusAviso},
		{"dev", "dev", "5.0.0", StatusOK},
	}
	for _, tc := range casos {
		t.Run(tc.nome, func(t *testing.T) {
			dir := projeto(t, tc.projeto)
			c := achar(t, Diagnose(dir, deps(tc.binario)), "versao")
			if c.Status != tc.quer {
				t.Fatalf("status = %s (%s), esperado %s", c.Status, c.Detalhe, tc.quer)
			}
			if tc.quer == StatusFalha && !strings.Contains(c.Sugestao, "npm install -g @nathanramorim/forge-sdd") {
				t.Errorf("sugestão sem o comando de atualização: %q", c.Sugestao)
			}
		})
	}
}

func TestDiagnoseVersaoSemSddrc(t *testing.T) {
	if c := achar(t, Diagnose(t.TempDir(), deps("5.0.0")), "versao"); c.Status != StatusFalha {
		t.Fatalf("status = %s, esperado falha", c.Status)
	}
}

func TestDiagnosePareceresIlegiveis(t *testing.T) {
	dir := projeto(t, "5.0.0")
	escreve(t, dir, "sdd/discovery/97-x/verdicts/verdict-feat-97-01.json", `{"task_id":"x","veredito":"conforme"}`)
	c := achar(t, Diagnose(dir, deps("5.0.0")), "pareceres-gravados")
	if c.Status != StatusAviso || !strings.Contains(c.Detalhe, "verdict-feat-97-01.json") {
		t.Fatalf("esperado aviso citando o arquivo: %+v", c)
	}
}

func workersCom(linhas string) Deps {
	d := deps("5.0.0")
	d.Runner = func(string, ...string) ([]byte, error) { return []byte(linhas), nil }
	return d
}

func TestDiagnoseWorkersDeOutraVersao(t *testing.T) {
	d := workersCom("eu 11 /h/.cache/forge-sdd/4.9.1-beta/forge-sdd worker --role builder\n" +
		"eu 12 /h/.cache/forge-sdd/4.9.1-beta/forge-sdd status\n" +
		"eu 13 /h/.cache/forge-sdd/5.0.0/forge-sdd worker --role reviewer\n" +
		"eu 14 /h/.cache/forge-sdd/5.0.0-beta/forge-sdd worker\n")
	c := achar(t, Diagnose(projeto(t, "5.0.0"), d), "workers")
	if c.Status != StatusAviso || !strings.Contains(c.Detalhe, "pid 11 (4.9.1-beta)") || !strings.Contains(c.Detalhe, "pid 14 (5.0.0-beta)") || strings.Contains(c.Detalhe, "pid 12") || strings.Contains(c.Detalhe, "pid 13") {
		t.Fatalf("detalhe inesperado: %+v", c)
	}
}

func TestDiagnoseWorkersNaoExecutaArgv0Forjado(t *testing.T) {
	tmp := t.TempDir()
	sentinela := filepath.Join(tmp, "PWNED")
	falso := filepath.Join(tmp, "forgeX")
	escreve(t, tmp, "forgeX", "#!/bin/sh\ntouch "+sentinela+"\n")
	if err := os.Chmod(falso, 0o755); err != nil {
		t.Fatal(err)
	}
	d := workersCom("eu 99 " + falso + " worker --role builder\n")
	c := achar(t, Diagnose(projeto(t, "5.0.0"), d), "workers")
	if _, err := os.Stat(sentinela); err == nil {
		t.Fatal("o doctor executou um caminho vindo do ps")
	}
	if c.Status != StatusAviso || !strings.Contains(c.Detalhe, "versão desconhecida") {
		t.Fatalf("esperado aviso com versão desconhecida e o caminho: %+v", c)
	}
}

func TestDiagnoseWorkersIgnoraOutroUsuarioERelativo(t *testing.T) {
	d := workersCom("outro 21 /h/.cache/forge-sdd/4.0.0/forge-sdd worker\neu 22 forge worker\neu 23 ./forge-sdd worker\n")
	if c := achar(t, Diagnose(projeto(t, "5.0.0"), d), "workers"); c.Status != StatusOK {
		t.Fatalf("esperado ok: %+v", c)
	}
}

func TestDiagnoseWorkersEscapaControle(t *testing.T) {
	d := workersCom("eu 31 /h/forgeX\x1b[31m/y worker\n")
	c := achar(t, Diagnose(projeto(t, "5.0.0"), d), "workers")
	if strings.ContainsRune(c.Detalhe, 0x1b) {
		t.Fatalf("caractere de controle impresso: %q", c.Detalhe)
	}
}

func TestVersaoMenor(t *testing.T) {
	casos := []struct {
		a, b string
		quer bool
	}{
		{"4.9.1-beta", "5.0.0", true},
		{"5.0.0-beta", "5.0.0", true},
		{"5.0.0", "5.0.0", false},
		{"5.0.0", "5.0.0-beta", false},
		{"5.0.0", "5.0.1", true},
		{"5.0.1", "5.0.0", false},
	}
	for _, c := range casos {
		if got := versaoMenor(c.a, c.b); got != c.quer {
			t.Errorf("versaoMenor(%s, %s) = %v", c.a, c.b, got)
		}
	}
}

func TestDiagnoseWorkersFalhaDoPs(t *testing.T) {
	d := deps("5.0.0")
	d.Runner = func(string, ...string) ([]byte, error) { return nil, errors.New("sem ps") }
	if c := achar(t, Diagnose(projeto(t, "5.0.0"), d), "workers"); c.Status != StatusAviso {
		t.Fatalf("status = %s, esperado aviso", c.Status)
	}
}

func TestDiagnoseForgeJSONInvalidoFalha(t *testing.T) {
	dir := projeto(t, "5.0.0")
	escreve(t, dir, ".forge/config/forge.json", `{"campo_inexistente":1}`)
	if c := achar(t, Diagnose(dir, deps("5.0.0")), "forge-json"); c.Status != StatusFalha {
		t.Fatalf("status = %s, esperado falha", c.Status)
	}
}

func TestDiagnoseSkillsAusentesAvisa(t *testing.T) {
	dir := projeto(t, "5.0.0")
	if err := os.Remove(filepath.Join(dir, ".agents", "skills", "sdd-build", "SKILL.md")); err != nil {
		t.Fatal(err)
	}
	if c := achar(t, Diagnose(dir, deps("5.0.0")), "skills-rules"); c.Status != StatusAviso || !strings.Contains(c.Detalhe, "sdd-build") {
		t.Fatalf("esperado aviso citando sdd-build: %+v", c)
	}
}

func TestValidateRelatorioRecusaInvalidos(t *testing.T) {
	invalidos := map[string]Relatorio{
		"status inválido":    {Checagens: []Checagem{{ID: "a", Status: "x"}}},
		"id repetido":        {Checagens: []Checagem{{ID: "a", Status: StatusOK}, {ID: "a", Status: StatusOK}}},
		"falha sem sugestão": {Falhas: 1, Checagens: []Checagem{{ID: "a", Status: StatusFalha}}},
		"contador errado":    {Checagens: []Checagem{{ID: "a", Status: StatusAviso, Sugestao: "s"}}},
	}
	for nome, r := range invalidos {
		if err := ValidateRelatorio(r); err == nil {
			t.Errorf("%s: esperava erro", nome)
		}
	}
}

func TestDiagnoseSaidaSemCaracteresDeControle(t *testing.T) {
	sujo := "\x1b[31mX\x1b[2J\r\x00\u202e\u2066\u200b\u2028Y"
	dir := projeto(t, "1.0.0"+sujo)
	nome := strings.ReplaceAll(sujo, "\x00", "")
	escreve(t, dir, ".claude/worktrees/x"+nome+"/f", "x")
	escreve(t, dir, "sdd/discovery/97-x/verdicts/v"+nome+".json", `{"task_id":"x"}`)
	d := workersCom("eu 7 /h/forge" + sujo + "/y worker\n")
	r := Diagnose(dir, d)
	if err := ValidateRelatorio(r); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var texto strings.Builder
	for _, c := range r.Checagens {
		texto.WriteString(c.Detalhe + "\n" + c.Sugestao + "\n")
	}
	for nome, saida := range map[string]string{"texto": texto.String(), "json": string(data)} {
		for _, r := range strings.ReplaceAll(saida, "\n", "") {
			if r < 0x20 || r == 0x7f || r == 0x202e || r == 0x2066 || r == 0x200b || r == 0x2028 {
				t.Errorf("%s contém caractere proibido %U: %q", nome, r, saida)
				break
			}
		}
	}
	for _, id := range []string{"versao", "worktrees", "pareceres-gravados", "workers"} {
		achar(t, r, id)
	}
}

// fix-118-04: checkout no Windows grava o veredito.md com CRLF; o contrato continua válido.
func TestDiagnoseContratoComCRLFPassa(t *testing.T) {
	dir := projeto(t, "5.0.0")
	p := filepath.Join(dir, filepath.FromSlash(veredito))
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	crlf := strings.ReplaceAll(string(b), "\n", "\r\n")
	escreve(t, dir, veredito, crlf)
	if c := achar(t, Diagnose(dir, deps("5.0.0")), "contrato-skill-binario"); c.Status != StatusOK {
		t.Fatalf("veredito.md com CRLF deveria passar: %+v", c)
	}
}
