package main

import (
	"fmt"

	"github.com/forge-sdd/cli/internal/skills"
	"github.com/spf13/cobra"
)

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Motor de Auto-Skills locais (DUP-05 Regra de Três / GOV-03 Fim de Onda)",
	Long: `Gerencia o ciclo de vida de skills locais reutilizáveis em .agents/skills/<slug>/SKILL.md
(DUP-10), mantendo o índice em sdd/skills/index.md e sincronizando skills aprovadas
para os adaptadores dos agentes (.claude/skills/ e .gemini/skills/).`,
}

var skillsProposeCmd = &cobra.Command{
	Use:   "propose",
	Short: "Propõe uma nova skill padronizada (status: proposed) em .agents/skills/<slug>/SKILL.md",
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir, _ := cmd.Flags().GetString("dir")
		if targetDir == "" {
			targetDir = "."
		}
		slug, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		steps, _ := cmd.Flags().GetString("steps")

		if slug == "" || description == "" || steps == "" {
			return fmt.Errorf("--name, --description e --steps são obrigatórios")
		}

		path, err := skills.Propose(targetDir, slug, description, steps)
		if err != nil {
			return err
		}
		fmt.Printf("✓ Skill %q proposta em %s (registrada em sdd/skills/index.md como proposed)\n", slug, path)
		return nil
	},
}

var skillsApproveCmd = &cobra.Command{
	Use:   "approve <slug>",
	Short: "Aprova uma skill proposta e sincroniza automaticamente para os agentes configurados",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir, _ := cmd.Flags().GetString("dir")
		if targetDir == "" {
			targetDir = "."
		}
		var slug string
		if len(args) > 0 {
			slug = args[0]
		} else {
			slug, _ = cmd.Flags().GetString("name")
		}
		if slug == "" {
			return fmt.Errorf("informe o <slug> da skill a aprovar")
		}

		synced, err := skills.Approve(targetDir, slug, nil)
		if err != nil {
			return err
		}
		fmt.Printf("✓ Skill %q aprovada e sincronizada (%d adaptador(es) atualizado(s))\n", slug, len(synced))
		for _, p := range synced {
			fmt.Printf("  - %s\n", p)
		}
		return nil
	},
}

var skillsSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sincroniza skills aprovadas (.agents/skills/) para os diretórios dos agentes (.claude/skills/ e .gemini/skills/)",
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir, _ := cmd.Flags().GetString("dir")
		if targetDir == "" {
			targetDir = "."
		}

		synced, err := skills.Sync(targetDir, nil)
		if err != nil {
			return err
		}
		fmt.Printf("✓ Sincronização concluída: %d adaptador(es) atualizado(s)\n", len(synced))
		for _, p := range synced {
			fmt.Printf("  - %s\n", p)
		}
		return nil
	},
}

var skillsListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista skills propostas e aprovadas registradas em .agents/skills/",
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir, _ := cmd.Flags().GetString("dir")
		if targetDir == "" {
			targetDir = "."
		}

		items, err := skills.List(targetDir)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			fmt.Println("Nenhuma skill encontrada em .agents/skills/.")
			return nil
		}
		for _, item := range items {
			fmt.Printf("- %s [%s]: %s (%s)\n", item.Name, item.Status, item.Description, item.Path)
		}
		return nil
	},
}

func init() {
	skillsProposeCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	skillsProposeCmd.Flags().String("name", "", "Slug em kebab-case da skill (obrigatório)")
	skillsProposeCmd.Flags().String("description", "", "Descrição curta de quando usar a skill (obrigatório)")
	skillsProposeCmd.Flags().String("steps", "", "Conteúdo/passos padronizados em Markdown (obrigatório)")

	skillsApproveCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	skillsApproveCmd.Flags().String("name", "", "Slug da skill a aprovar (alternativa ao argumento posicional)")

	skillsSyncCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	skillsListCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")

	skillsCmd.AddCommand(skillsProposeCmd)
	skillsCmd.AddCommand(skillsApproveCmd)
	skillsCmd.AddCommand(skillsSyncCmd)
	skillsCmd.AddCommand(skillsListCmd)
}
