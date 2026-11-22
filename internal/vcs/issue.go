package vcs

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var issueNumeroRe = regexp.MustCompile(`/issues/([0-9]+)\s*$`)

// CriarIssue cria a Issue via `gh issue create` (argv explícito, sem shell; SEG-03).
func (g *github) CriarIssue(dir string, n NovaIssue) (Issue, error) {
	titulo := strings.TrimSpace(n.Titulo)
	if titulo == "" {
		return Issue{}, errors.New("título da issue vazio")
	}
	args := []string{"issue", "create", "--title", titulo, "--body", n.Corpo}
	for _, e := range n.Etiquetas {
		if e = strings.TrimSpace(e); e != "" {
			args = append(args, "--label", e)
		}
	}
	out, err := g.run(dir, "gh", args...)
	if err != nil {
		return Issue{}, err
	}
	url := ultimaLinha(out)
	m := issueNumeroRe.FindStringSubmatch(url)
	if m == nil {
		return Issue{}, fmt.Errorf("não consegui ler o número da issue na saída do gh: %q", truncar(out, 80))
	}
	num, err := strconv.Atoi(m[1])
	if err != nil {
		return Issue{}, err
	}
	return Issue{Numero: num, URL: url, Titulo: titulo}, nil
}

// ProcurarIssueAberta lista as abertas por busca e filtra no cliente (a busca do gh é aproximada):
// vale título igual ao base ou terminado em "] <base>". Saída fora do esquema é recusada (SEG-01).
func (g *github) ProcurarIssueAberta(dir, tituloBase string) (*Issue, error) {
	base := strings.TrimSpace(tituloBase)
	if base == "" {
		return nil, errors.New("título-base da issue vazio")
	}
	out, err := g.run(dir, "gh", "issue", "list", "--state", "open", "--search", base,
		"--json", "number,url,title")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(out) == "" {
		return nil, nil
	}
	var list []struct {
		Number *int    `json:"number"`
		URL    *string `json:"url"`
		Title  *string `json:"title"`
	}
	dec := json.NewDecoder(strings.NewReader(out))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&list); err != nil {
		return nil, fmt.Errorf("saída do gh issue list inválida: %w", err)
	}
	for _, it := range list {
		if it.Number == nil || it.URL == nil || it.Title == nil || *it.Number <= 0 {
			return nil, errors.New("saída do gh issue list inválida: campo ausente")
		}
		t := strings.TrimSpace(*it.Title)
		if t == base || strings.HasSuffix(t, "] "+base) {
			return &Issue{Numero: *it.Number, URL: *it.URL, Titulo: t}, nil
		}
	}
	return nil, nil
}
