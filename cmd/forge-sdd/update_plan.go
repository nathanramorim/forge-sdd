package main

import (
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/forge-sdd/cli/internal/config"
	"github.com/forge-sdd/cli/internal/scaffold"
)

// printUpgradePlan imprime o que a atualização de estrutura moveria e removeria, sem escrever
// nada (ARQ-05). Devolve true quando há alguma ação planejada.
func printUpgradePlan(w io.Writer, targetDir string, cfg config.Config) (bool, error) {
	migration, err := scaffold.PlanLayoutMigration(targetDir)
	if err != nil {
		return false, err
	}
	obsolete, err := scaffold.PlanObsoleteAgentArtifacts(targetDir, cfg)
	if err != nil {
		return false, err
	}

	if len(migration.Moves) > 0 {
		fmt.Fprintln(w, "Migração do layout (legado para sdd/archive/, com backup em sdd/backups/):")
		fmt.Fprintln(w, "  atenção: todo o conteúdo de sdd/features/ é arquivado, inclusive features em andamento; confira antes de aplicar.")
		for _, m := range migration.Moves {
			fmt.Fprintf(w, "  mover   %s -> %s\n", m.From, m.To)
		}
	}
	for _, s := range migration.Skipped {
		fmt.Fprintf(w, "  mantido %s (%s)\n", s.Path, s.Reason)
	}
	if len(obsolete.Remove) > 0 {
		fmt.Fprintln(w, "Artefatos de agente obsoletos (idênticos ao template anterior):")
		for _, p := range obsolete.Remove {
			fmt.Fprintf(w, "  remover %s\n", p)
		}
	}
	for _, p := range obsolete.Preserve {
		fmt.Fprintf(w, "  mantido %s (editado pelo usuário ou fora do padrão)\n", p)
	}
	if slices.Contains(cfg.Agents, config.AgentClaude) {
		if v, ok := scaffold.PRGuardAdotavel(targetDir); ok {
			fmt.Fprintf(w, "  atualizar scripts/pr_guard.py (igual ao entregue pelo forge na %s)\n", v)
		} else if scaffold.PRGuardUserOwned(targetDir) {
			fmt.Fprintln(w, "  mantido scripts/pr_guard.py (sem o marcador forge-sdd:generated; é do usuário e não é atualizado; a versão nova vai para scripts/pr_guard.novo.py)")
		}
	}

	for _, rel := range scaffold.InstrucoesDoUsuario(targetDir, cfg) {
		fmt.Fprintf(w, "  mantido %s (sem o marcador forge-sdd:generated; é seu e não é atualizado; o conteúdo novo vai para o .novo.md)\n", rel)
	}

	atualizar, manter := scaffold.PlanoRegras(targetDir, cfg)
	for _, rel := range atualizar {
		fmt.Fprintf(w, "  atualizar %s (igual ao entregue pelo forge; recebe a versão nova)\n", rel)
	}
	for _, rel := range manter {
		fmt.Fprintf(w, "  mantido %s (editado por você; a versão nova vai para o .novo)\n", rel)
	}

	acted := len(migration.Moves) > 0 || len(obsolete.Remove) > 0
	if !acted {
		fmt.Fprintln(w, "Nenhuma migração de layout nem limpeza de obsoletos necessária.")
	}
	return acted, nil
}

// printPlanoCompleto lista o plano inteiro do update além do layout (ARQ-05: nada é gravado): arquivos
// reescritos, criados e preservados, campos de modo da Constituição, hook do inbox, workflow --ci e a
// versão que seria gravada. Instruções e rules preservadas já saíram em printUpgradePlan. Só imprime;
// a decisão vem de scaffold.PlanoDoUpdate (ARQ-01).
func printPlanoCompleto(w io.Writer, targetDir string, cfg config.Config, ci bool, versao string) error {
	itens, err := scaffold.PlanoDoUpdate(targetDir, cfg)
	if err != nil {
		return err
	}
	var reescrever, criar, preservar []string
	iguais, mantidosDoProjeto := 0, 0
	for _, it := range itens {
		switch it.Acao {
		case scaffold.AcaoReescrever:
			reescrever = append(reescrever, it.Rel)
		case scaffold.AcaoCriar:
			criar = append(criar, it.Rel)
		case scaffold.AcaoIgual:
			iguais++
		case scaffold.AcaoPreservar:
			if it.Tipo == "outro" {
				mantidosDoProjeto++
			} else {
				preservar = append(preservar, it.Rel)
			}
		}
	}
	fmt.Fprintf(w, "Serão reescritos (%d):\n", len(reescrever))
	for _, rel := range reescrever {
		fmt.Fprintf(w, "  reescrever %s\n", rel)
	}
	fmt.Fprintf(w, "Serão criados (%d):\n", len(criar))
	for _, rel := range criar {
		fmt.Fprintf(w, "  criar %s\n", rel)
	}
	fmt.Fprintf(w, "Serão preservados (editados por você) (%d):\n", len(preservar))
	for _, rel := range preservar {
		fmt.Fprintf(w, "  preservar %s (a versão nova vai para o .novo)\n", rel)
	}
	fmt.Fprintf(w, "Sem mudança: %d arquivo(s) idêntico(s); %d do projeto (sdd/, skills e rules suas) nunca são sobrescritos.\n", iguais, mantidosDoProjeto)

	rotulos, err := scaffold.PlanoCamposDeModo(targetDir, io.Discard)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "Serão acrescentados na Constituição (%d):\n", len(rotulos))
	for _, r := range rotulos {
		fmt.Fprintf(w, "  acrescentar %s\n", r)
	}

	if slices.Contains(cfg.Agents, config.AgentClaude) {
		if seria, herr := scaffold.PlanoHookInbox(targetDir); herr == nil && seria {
			fmt.Fprintln(w, "hook UserPromptSubmit seria instalado")
		}
	}
	fmt.Fprintf(w, "Versão que seria gravada: %s\n", versao)
	if ci {
		fmt.Fprintln(w, "Workflow (--ci):")
		dest := scaffold.ForgeProntoPath(targetDir)
		if _, err := os.Lstat(dest); err == nil && !scaffold.IsGeneratedCI(dest) {
			fmt.Fprintln(w, "  preservar .github/workflows/forge-pronto.yml (não é do forge-sdd; sem o marcador)")
		} else {
			fmt.Fprintln(w, "  criar .github/workflows/forge-pronto.yml")
		}
	}
	return nil
}
