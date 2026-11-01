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

var queueHandoffCmd = &cobra.Command{
	Use:   "handoff",
	Short: "Transfere um artefato de forma idempotente entre duas estações",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		id, _ := cmd.Flags().GetString("id")
		title, _ := cmd.Flags().GetString("title")
		artifact, _ := cmd.Flags().GetString("artifact")
		input, _ := cmd.Flags().GetString("input")
		output, _ := cmd.Flags().GetString("output")
		idempotencyKey, _ := cmd.Flags().GetString("idempotency-key")

		if from == "" || to == "" || id == "" {
			return fmt.Errorf("flags --from, --to e --id são obrigatórias")
		}

		item, hit, err := queue.Handoff(dir, from, to, id, title, artifact, input, output, idempotencyKey)
		if err != nil {
			return err
		}

		hitMsg := ""
		if hit {
			hitMsg = " (hit idempotente detectado)"
		}
		fmt.Printf("✓ Handoff concluído (%s -> %s): %s [%s]%s\n", from, to, item.ID, item.IdempotencyKey, hitMsg)
		return nil
	},
}

var queueStationsCmd = &cobra.Command{
	Use:   "stations",
	Short: "Lista o estado e versão das 5 filas de estações",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		stations := queue.CanonicalStations()

		fmt.Println("Filas Dedicadas por Estação:")
		for _, st := range stations {
			sq, err := queue.LoadStationQueue(dir, st)
			if err != nil {
				fmt.Printf("  - %s: erro ao carregar (%v)\n", st, err)
				continue
			}
			activeCount := 0
			for _, it := range sq.Items {
				if it.Status != queue.StatusApproved {
					activeCount++
				}
			}
			fmt.Printf("  - [%s] versão: %d | total itens: %d (ativos: %d) | chaves: %d\n",
				st, sq.Version, len(sq.Items), activeCount, len(sq.ProcessedKeys))
		}
		return nil
	},
}

var queueActivityCmd = &cobra.Command{
	Use:     "activity",
	Aliases: []string{"log"},
	Short:   "Registra ou higieniza o histórico de atividades e telemetria",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		sanitize, _ := cmd.Flags().GetBool("sanitize")
		if sanitize {
			count, err := queue.SanitizeOrphanActivities(dir)
			if err != nil {
				return err
			}
			fmt.Printf("✓ %d atividades órfãs reconciliadas para 'success'\n", count)
			return nil
		}

		station, _ := cmd.Flags().GetString("station")
		tool, _ := cmd.Flags().GetString("tool")
		action, _ := cmd.Flags().GetString("action")
		details, _ := cmd.Flags().GetString("details")
		status, _ := cmd.Flags().GetString("status")
		command, _ := cmd.Flags().GetString("command")
		filePath, _ := cmd.Flags().GetString("file-path")
		output, _ := cmd.Flags().GetString("output")
		errMsg, _ := cmd.Flags().GetString("error")
		durationMs, _ := cmd.Flags().GetInt64("duration")

		if station == "" {
			station = queue.StationLider
		}
		if status == "" {
			status = "success"
		}
		if tool == "" {
			return fmt.Errorf("--tool é obrigatório (ex: run_command, read_file, edit_file, gate, git)")
		}

		if command != "" || filePath != "" || output != "" || errMsg != "" || durationMs > 0 {
			item := queue.ActivityItem{
				Station:    station,
				Tool:       tool,
				Action:     action,
				Details:    details,
				Status:     status,
				Command:    command,
				FilePath:   filePath,
				Output:     output,
				Error:      errMsg,
				DurationMs: durationMs,
			}
			if err := queue.RecordActivityDetailed(dir, item); err != nil {
				return err
			}
		} else {
			if err := queue.RecordActivity(dir, station, tool, action, details, status); err != nil {
				return err
			}
		}
		fmt.Printf("✓ Atividade registrada para estação '%s': tool=%s\n", station, tool)
		return nil
	},
}

var queueLogCmd = queueActivityCmd

func init() {
	queueHandoffCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	queueHandoffCmd.Flags().String("from", "", "Estação de origem (ex: lider, pm-architect, planner, builder, governance)")
	queueHandoffCmd.Flags().String("to", "", "Estação de destino (ex: lider, pm-architect, planner, builder, governance)")
	queueHandoffCmd.Flags().String("id", "", "ID do item ou artefato em transição")
	queueHandoffCmd.Flags().String("title", "", "Título descritivo do artefato")
	queueHandoffCmd.Flags().String("artifact", "", "Caminho do arquivo de artefato gerado")
	queueHandoffCmd.Flags().String("input", "", "Descrição do contrato de entrada")
	queueHandoffCmd.Flags().String("output", "", "Descrição do contrato de saída")
	queueHandoffCmd.Flags().String("idempotency-key", "", "Chave de idempotência (opcional; gerada via SHA-256 se omitida)")

	queueStationsCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")

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

	queueActivityCmd.Flags().String("dir", ".", "Diretório alvo do projeto Forge-SDD")
	queueActivityCmd.Flags().String("station", "lider", "Estação responsável (lider, pm-architect, planner, builder, governance)")
	queueActivityCmd.Flags().String("tool", "", "Nome da ferramenta executada (run_command, read_file, edit_file, gate, git, etc)")
	queueActivityCmd.Flags().String("action", "", "Argumento, comando ou arquivo operado")
	queueActivityCmd.Flags().String("details", "", "Descrição resumida da intenção da operação")
	queueActivityCmd.Flags().String("status", "success", "Status da execução: running, success, error")
	queueActivityCmd.Flags().Bool("sanitize", false, "Reconcilia e encerra atividades órfãs pendentes em 'running'")
	queueActivityCmd.Flags().String("command", "", "Comando completo executado")
	queueActivityCmd.Flags().String("file-path", "", "Caminho do arquivo afetado")
	queueActivityCmd.Flags().String("output", "", "Saída de texto capturada")
	queueActivityCmd.Flags().String("error", "", "Mensagem de erro capturada")
	queueActivityCmd.Flags().Int64("duration", 0, "Duração da execução em milissegundos")

	queueCmd.AddCommand(queueEnqueueCmd)
	queueCmd.AddCommand(queueNextCmd)
	queueCmd.AddCommand(queueCompleteCmd)
	queueCmd.AddCommand(queueReviewCmd)
	queueCmd.AddCommand(queueListCmd)
	queueCmd.AddCommand(queuePanelCmd)
	queueCmd.AddCommand(queueHandoffCmd)
	queueCmd.AddCommand(queueStationsCmd)
	queueCmd.AddCommand(queueActivityCmd)
}
