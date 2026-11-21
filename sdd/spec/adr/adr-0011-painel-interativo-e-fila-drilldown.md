# ADR-0011: Painel Interativo, Stream de Chat nos Workers e Fila com Drill-Down

- **Data:** 2026-11-15
- **Status:** `aceito` (2026-11-20, aprovado pelo dono no Gate 1 da 105)
- **Regras afetadas (IDs da constitution):** `ARQ-01`, `COD-01`, `DUP-01`, `SEG-05`, `GOV-01` (este ADR; nenhum ID foi criado ou removido)
- **Autor / Responsável:** Líder (discovery 84)

## Contexto

O painel `forge status --live` truncava a lista ("... e mais 19 tarefas") e misturava tabelas de processos (CPU, RAM, PID) com o andamento da entrega. Os painéis do Builder e do Reviewer mostravam contadores secos e saída crua, enquanto a Discovery tem uma conversa legível. Não havia como inspecionar uma task executada (arquivos, laudo do Reviewer, tempo) sem vasculhar pastas. Referência: `sdd/discovery/84-painel-interativo-chat-e-fila-drilldown/` (`feat-84-01` a `feat-84-07`).

## Opções Consideradas

1. **Só ajustar o texto atual:** barato, mas não resolve a inspeção por task.
2. **TUI completa própria:** poderosa, mas pesada e fora do orçamento de dependências.
3. **Stream narrativo + fila com drill-down reutilizando `huh`/`bubbles`:** sem dependência nova, no padrão do menu (ADR-0005).

## Decisão

- **D-1 Stream de chat:** `monitor.ChatFeedFormatter` converte eventos do worker (ler spec, editar arquivos, rodar testes) em linhas curtas de chat; o log bruto continua completo à parte (`SEG-05`).
- **D-2 Painel executivo:** `pipeline.FormatRichDashboard` separa o fluxo de entrega dos processos e usa `queue/activity.json` como fonte do que cada estação faz.
- **D-3 Fila com drill-down:** `survey.QueueDrilldownTUI` lista as tasks em duas abas, **Na Fila** e **Concluídas**; `TaskDetailView` mostra arquivos, laudo, tempo, tokens por etapa e o diagrama do pipeline da task.
- **D-4:** a lógica fica em `internal/`; `cmd/` só encaminha (`ARQ-01`).

## Consequências

- **Positivas:** andamento legível de relance; inspeção de qualquer task sem sair do terminal; nenhuma dependência nova.
- **Negativas / Trade-offs:** mais código de apresentação em `internal/survey` e `internal/monitor`, com testes de renderização sensíveis a largura e cor.
- **Ações de acompanhamento:** o painel só é tão correto quanto o `queue.json`; ver ADR-0012.
