package main

import (
	"fmt"
	"io"

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

	acted := len(migration.Moves) > 0 || len(obsolete.Remove) > 0
	if !acted {
		fmt.Fprintln(w, "Nenhuma migração de layout nem limpeza de obsoletos necessária.")
	}
	return acted, nil
}
