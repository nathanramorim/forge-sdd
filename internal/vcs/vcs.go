// Package vcs isola o sistema de versionamento (PR, draft, merge) atrás de um adaptador lido da
// Constituição (`VCS / Work Item System`). Só `github` e `nenhum` existem; `azure-devops` está
// adiado e é recusado (ADR-0004: nenhum adaptador aprova; o merge continua sendo decisão humana).
package vcs

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Runner roda `name args...` em dir e devolve stdout; injetável nos testes (TST-04).
type Runner func(dir, name string, args ...string) (string, error)

// ErrSemPR indica que o VCS `nenhum` não tem PR: a branch fica pronta (RN-25).
var ErrSemPR = errors.New("VCS nenhum: não há PR")

// PR descreve um PR aberto.
type PR struct {
	Numero int
	URL    string
	Base   string
}

// NovoPR reúne os dados para criar um PR; CorpoArquivo é caminho relativo a dir e vale quando
// Corpo (texto) está vazio.
type NovoPR struct {
	Base, Branch, Titulo, Corpo, CorpoArquivo string
}

// Mescla identifica o que mesclar: Numero no github; Branch e Base no merge local do `nenhum`.
type Mescla struct {
	Numero       int
	Branch, Base string
}

// Adaptador é a fronteira do VCS. Nenhum método aprova nada: mesclar exige a decisão já tomada
// pelo chamador (veredito conforme do SHA atual).
type Adaptador interface {
	Nome() string
	CriarPRDraft(dir string, p NovoPR) (PR, error)
	ProcurarPRAberto(dir, branch string) (*PR, error)
	CorrigirBase(dir string, numero int, base string) error
	MarcarPronto(dir string, numero int) error
	Mesclar(dir string, m Mescla) error
}

// Novo devolve o adaptador do VCS; valor desconhecido ou adiado é recusado.
func Novo(nome string, run Runner) (Adaptador, error) {
	if run == nil {
		run = RunGitOrGh
	}
	switch strings.TrimSpace(nome) {
	case "github":
		return &github{run: run}, nil
	case "nenhum":
		return &nenhum{run: run}, nil
	case "azure-devops":
		return nil, errors.New(`VCS "azure-devops" não verificado: adaptador adiado nesta entrega`)
	}
	return nil, fmt.Errorf("VCS %q desconhecido; use github ou nenhum (Constituição, VCS / Work Item System)", nome)
}

// DaConstituicao lê o VCS da Constituição de dir e devolve o adaptador.
func DaConstituicao(dir string, run Runner) (Adaptador, error) {
	nome, err := LeVCS(dir)
	if err != nil {
		return nil, err
	}
	return Novo(nome, run)
}

var vcsLinhaRe = regexp.MustCompile(`(?m)^\|\s*VCS / Work Item System\s*\|\s*([^|]*?)\s*\|`)

// LeVCS lê o valor de `VCS / Work Item System` em sdd/memory/constitution.md. Sem arquivo ou sem
// a linha, vale `github` (projeto legado, comportamento anterior).
func LeVCS(dir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(dir, "sdd", "memory", "constitution.md"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "github", nil
		}
		return "", fmt.Errorf("ler a Constituição: %w", err)
	}
	m := vcsLinhaRe.FindSubmatch(data)
	if m == nil {
		return "github", nil
	}
	return strings.TrimSpace(string(m[1])), nil
}

// RunGitOrGh é o Runner real: argv explícito e Dir fixo, sem shell; só `git` e `gh` passam (SEG-03).
func RunGitOrGh(dir, name string, args ...string) (string, error) {
	if name != "git" && name != "gh" {
		return "", fmt.Errorf("comando %q não permitido", name)
	}
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		sub := ""
		if len(args) > 0 {
			sub = " " + args[0]
		}
		return strings.TrimSpace(stdout.String()), fmt.Errorf("%s%s: %w: %s", name, sub, err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

var prNumeroRe = regexp.MustCompile(`/pull/([0-9]+)\s*$`)

func numeroDaURL(s string) (int, error) {
	m := prNumeroRe.FindStringSubmatch(s)
	if m == nil {
		return 0, fmt.Errorf("não consegui ler o número do PR na saída do gh: %q", truncar(s, 80))
	}
	return strconv.Atoi(m[1])
}

func truncar(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}
