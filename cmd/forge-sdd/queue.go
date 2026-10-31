package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/forge-sdd/cli/internal/queue"
	"github.com/spf13/cobra"
)

var queueCmd = &cobra.Command{
	Use:   "queue",
	Short: "Gerencia a fila determinística do Builder (sdd/.queue/queue.json)",
	Long:  `Subcomando para enfileirar, consumir, completar, revisar e exibir o painel executivo da fila de tarefas do Builder.`,
}

var queueEnqueueCmd = &cobra.Command{
	Use:   "enqueue",
	Short: "Valida a Spec (P-02/P-03) e enfileira uma task na fila do Builder",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		featureID, _ := cmd.Flags().GetString("feature")
		branch, _ := cmd.Flags().GetString("branch")
		taskID, _ := cmd.Flags().GetString("task-id")
		title, _ := cmd.Flags().GetString("title")
		specPath, _ := cmd.Flags().GetString("task")
		if specPath == "" {
			specPath, _ = cmd.Flags().GetString("spec")
		}
		rulesStr, _ := cmd.Flags().GetString("rules")

		if specPath == "" {
			return fmt.Errorf("--task (ou --spec) com o caminho da Spec é obrigatório")
		}

		var rules []string
		for _, r := range strings.Split(rulesStr, ",") {
			r = strings.TrimSpace(r)
			if r != "" {
				rules = append(rules, r)
			}
		}

		if err := queue.Enqueue(dir, featureID, branch, taskID, title, specPath, rules); err != nil {
			return err
		}
		fmt.Printf("✓ Task enfileirada com sucesso (%s)\n", specPath)
		return nil
	},
}

var queueNextCmd = &cobra.Command{
	Use:   "next",
	Short: "Retorna a próxima task pendente na fila e muda seu status para building",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		asJSON, _ := cmd.Flags().GetBool("json")

		item, err := queue.Next(dir)
		if err != nil {
			return err
		}

		if asJSON {
			data, err := json.MarshalIndent(item, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("Task: %s (%s)\n", item.TaskID, item.Title)
		fmt.Printf("Spec: %s\n", item.SpecPath)
		fmt.Printf("Status: %s (tentativa #%d)\n", item.Status, item.Attempts)
		if len(item.Rules) > 0 {
			fmt.Printf("Rules: %s\n", strings.Join(item.Rules, ", "))
		}
		return nil
	},
}

var queueCompleteCmd = &cobra.Command{
	Use:   "complete <task-id>",
	Short: "Marca uma task em building como review_pending",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		taskID := args[0]

		if err := queue.Complete(dir, taskID); err != nil {
			return err
		}
		fmt.Printf("✓ Task %s movida para %s\n", taskID, queue.StatusReviewPending)
		return nil
	},
}

var queueReviewCmd = &cobra.Command{
	Use:   "review <task-id>",
	Short: "Registra e valida o parecer semântico da Camada B (Analista de Governança)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		verdictFile, _ := cmd.Flags().GetString("verdict-file")
		verdictJSONStr, _ := cmd.Flags().GetString("verdict-json")
		taskID := args[0]

		var raw []byte
		var err error
		switch {
		case verdictJSONStr != "":
			raw = []byte(verdictJSONStr)
		case verdictFile != "":
			raw, err = os.ReadFile(verdictFile)
			if err != nil {
				return fmt.Errorf("falha ao ler --verdict-file %q: %w", verdictFile, err)
			}
		default:
			return fmt.Errorf("informe --verdict-file ou --verdict-json")
		}

		item, err := queue.Review(dir, taskID, raw)
		if err != nil {
			return err
		}
		fmt.Printf("✓ Parecer registrado para %s: veredito=%s → status=%s\n", item.TaskID, item.GovernanceVerdict.Veredito, item.Status)
		return nil
	},
}

var queueListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista todas as tasks da fila e seus estados atuais",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		asJSON, _ := cmd.Flags().GetBool("json")

		q, err := queue.Load(dir)
		if err != nil {
			return err
		}

		if asJSON {
			data, err := json.MarshalIndent(q, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		if len(q.Items) == 0 {
			fmt.Println("Fila vazia.")
			return nil
		}
		fmt.Printf("Feature: %s | Branch: %s | Spec: %s\n", q.FeatureID, q.Branch, q.SpecStatus)
		for _, it := range q.Items {
			fmt.Printf("  - [%s] %s: %s (tentativas: %d, spec: %s)\n", it.Status, it.TaskID, it.Title, it.Attempts, it.SpecPath)
		}
		return nil
	},
}

var queuePanelCmd = &cobra.Command{
	Use:   "panel",
	Short: "Exibe o Painel Executivo do Líder (Camada C)",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		q, err := queue.Load(dir)
		if err != nil {
			return err
		}
		fmt.Print(queue.FormatExecutivePanel(q))
		return nil
	},
}

func init() {
	queueEnqueueCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	queueEnqueueCmd.Flags().String("feature", "", "ID da feature (ex: feat-58)")
	queueEnqueueCmd.Flags().String("branch", "", "Branch da feature (ex: feat/58-nome)")
	queueEnqueueCmd.Flags().String("task-id", "", "ID da task (opcional; inferido do nome do arquivo se omitido)")
	queueEnqueueCmd.Flags().String("title", "", "Título da task (opcional; inferido do cabeçalho da Spec se omitido)")
	queueEnqueueCmd.Flags().String("task", "", "Caminho para o arquivo de Spec da task")
	queueEnqueueCmd.Flags().String("spec", "", "Alias para --task (caminho da Spec)")
	queueEnqueueCmd.Flags().String("rules", "", "Lista de rules/IDs separados por vírgula")

	queueNextCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	queueNextCmd.Flags().Bool("json", false, "Exibir saída em formato JSON")

	queueCompleteCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")

	queueReviewCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	queueReviewCmd.Flags().String("verdict-file", "", "Caminho para arquivo JSON com o parecer da Camada B")
	queueReviewCmd.Flags().String("verdict-json", "", "String JSON com o parecer da Camada B")

	queueListCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	queueListCmd.Flags().Bool("json", false, "Exibir saída em formato JSON")

	queuePanelCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")

	queueCmd.AddCommand(queueEnqueueCmd)
	queueCmd.AddCommand(queueNextCmd)
	queueCmd.AddCommand(queueCompleteCmd)
	queueCmd.AddCommand(queueReviewCmd)
	queueCmd.AddCommand(queueListCmd)
	queueCmd.AddCommand(queuePanelCmd)
}
