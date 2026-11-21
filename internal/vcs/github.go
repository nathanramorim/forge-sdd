package vcs

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type github struct{ run Runner }

func (g *github) Nome() string { return "github" }

// CriarPRDraft cria o PR sempre em draft (RN-7); `--draft` vem por último.
func (g *github) CriarPRDraft(dir string, p NovoPR) (PR, error) {
	corpo := []string{"--body-file", p.CorpoArquivo}
	if p.Corpo != "" || p.CorpoArquivo == "" {
		corpo = []string{"--body", p.Corpo}
	}
	args := append([]string{"pr", "create", "--base", p.Base, "--head", p.Branch, "--title", p.Titulo}, corpo...)
	out, err := g.run(dir, "gh", append(args, "--draft")...)
	if err != nil {
		return PR{}, err
	}
	url := ultimaLinha(out)
	n, err := numeroDaURL(url)
	if err != nil {
		return PR{}, err
	}
	return PR{Numero: n, URL: url, Base: p.Base}, nil
}

func (g *github) ProcurarPRAberto(dir, branch string) (*PR, error) {
	out, err := g.run(dir, "gh", "pr", "list", "--head", branch, "--state", "open",
		"--json", "number,url,baseRefName", "--limit", "1")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(out) == "" {
		return nil, nil
	}
	var list []struct {
		Number      int    `json:"number"`
		URL         string `json:"url"`
		BaseRefName string `json:"baseRefName"`
	}
	if err := json.Unmarshal([]byte(out), &list); err != nil {
		return nil, fmt.Errorf("saída do gh pr list inválida: %w", err)
	}
	if len(list) == 0 {
		return nil, nil
	}
	return &PR{Numero: list[0].Number, URL: list[0].URL, Base: list[0].BaseRefName}, nil
}

// CorrigirBase troca a base de um PR já aberto.
func (g *github) CorrigirBase(dir string, numero int, base string) error {
	_, err := g.run(dir, "gh", "pr", "edit", strconv.Itoa(numero), "--base", base)
	return err
}

func (g *github) MarcarPronto(dir string, numero int) error {
	_, err := g.run(dir, "gh", "pr", "ready", strconv.Itoa(numero))
	return err
}

func (g *github) Mesclar(dir string, m Mescla) error {
	_, err := g.run(dir, "gh", "pr", "merge", strconv.Itoa(m.Numero))
	return err
}

func ultimaLinha(s string) string {
	l := strings.Split(strings.TrimSpace(s), "\n")
	return strings.TrimSpace(l[len(l)-1])
}
