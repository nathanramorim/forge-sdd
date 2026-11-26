package config

import (
	"fmt"
	"regexp"
)

// Modos de instalar o forge numa sessão do Claude Code na nuvem (nuvem.instalar).
const (
	NuvemInstalarNPM = "npm" // npm install -g do pacote publicado (padrão)
	NuvemInstalarGit = "git" // clona o repositório do forge e compila com Go (versões não publicadas)
)

// NuvemConfig liga o forge em sessões do Claude Code na nuvem: um hook de início de sessão,
// versionado no repositório, instala o forge e sobe o Builder e o Reviewer em segundo plano.
// Fora da nuvem o hook não faz nada.
type NuvemConfig struct {
	Ativo bool `json:"ativo"`
	// Repos são os repositórios que a sessão precisa ter clonados ao lado deste (modo hub). Vazio
	// quando o forge cuida só deste repositório.
	Repos []string `json:"repos,omitempty"`
	// Instalar é npm (padrão) ou git.
	Instalar string `json:"instalar,omitempty"`
	// Ref é a versão do pacote npm ou a branch/tag do repositório no modo git. Vazio: a versão do
	// forge que ativou (npm) ou main (git).
	Ref string `json:"ref,omitempty"`
}

var (
	nuvemRepoRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,99}(/[A-Za-z0-9][A-Za-z0-9._-]{0,99})?$`)
	nuvemRefRe  = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/+-]{0,99}$`)
)

// InstalarEfetivo devolve o modo de instalação, npm quando omitido.
func (n NuvemConfig) InstalarEfetivo() string {
	if n.Instalar == "" {
		return NuvemInstalarNPM
	}
	return n.Instalar
}

// Validate confere os campos da nuvem: entram num script de shell, então só aceitam nomes simples.
func (n NuvemConfig) Validate() error {
	if i := n.Instalar; i != "" && i != NuvemInstalarNPM && i != NuvemInstalarGit {
		return fmt.Errorf("nuvem.instalar %q inválido (use npm ou git)", i)
	}
	if n.Ref != "" && !nuvemRefRe.MatchString(n.Ref) {
		return fmt.Errorf("nuvem.ref %q inválido (só letras, números, ., _, -, + e /)", n.Ref)
	}
	if len(n.Repos) > 32 {
		return fmt.Errorf("nuvem.repos aceita no máximo 32 repositórios")
	}
	for _, r := range n.Repos {
		if !nuvemRepoRe.MatchString(r) || r == ".." || r == "." {
			return fmt.Errorf("nuvem.repos: %q inválido (use nome ou dono/nome)", r)
		}
	}
	return nil
}

// NuvemAtiva diz se o forge está ligado para sessões na nuvem.
func (c ForgeConfig) NuvemAtiva() bool { return c.Nuvem != nil && c.Nuvem.Ativo }

// NuvemOuPadrao devolve a configuração da nuvem, vazia quando a chave não existe.
func (c ForgeConfig) NuvemOuPadrao() NuvemConfig {
	if c.Nuvem == nil {
		return NuvemConfig{}
	}
	return *c.Nuvem
}
