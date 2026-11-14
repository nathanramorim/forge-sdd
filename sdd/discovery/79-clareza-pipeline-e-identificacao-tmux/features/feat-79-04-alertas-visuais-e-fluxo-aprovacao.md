# feat-79-04 — Alertas Visuais e Fluxo de Aprovação Não Bloqueante

**Branch:** `feat/79-clareza-pipeline-e-identificacao-tmux`
**Fase:** 79-04
**Tamanho:** `M`
**Depende de:** 79-02
**Status:** `completed` (FLX-01 aprovado)

## Contexto & Motivação
Hoje os workers em modo headless ou live ficam ociosos ocupando a tela com loops de espera ou bloqueando o terminal do desenvolvedor. Quando um agente demanda aprovação humana (Camada C) ou intervenção, a experiência deve ser limpa e não bloqueante: a esteira fica em repouso passivo sem poluir a tela, o terminal direito permanece disponível, e um badge/alerta visual no painel do worker indica claramente quando a atenção humana é necessária.

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** Toca a camada de worker (`internal/pipeline/worker.go`, `dispatcher.go`), formatação de mensagens de terminal e manipulação de títulos de painéis do tmux (`ARQ-11`, `COD-01`).

## Decisões do projeto aplicadas
- `ADR-0006`: Visibilidade, tmux e Entrega ao Builder.
- `ADR-0008`: Sessões com CLI Aberta no Tmux (Modo Live).
- `SEG-01`, `SEG-05`: Isolamento e integridade de entrega de jobs.

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Loop de worker | `internal/pipeline/worker.go` | `RunWorker`, loop de claim | estender | Silenciar loops barulhentos quando ocioso; emitir status apenas na mudança de estado. |
| Leitura de tela do painel | `internal/pipeline/deliver.go` | `capture-pane` | reusar | Detectar prompts de confirmação interativa ou parada de agente. |
| Avisos ao inbox | `internal/notify` | Notificações de escalonamento | estender | Refletir alertas de aprovação pendente no título do painel tmux. |

## Dependências novas
- Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `COD-01`, `COD-03`, `SEG-01`, `TST-04`.
- Rules: `sdd-codigo-e-testes.md`, `sdd-seguranca.md`.

## Escopo
1. Otimizar a saída de espera do worker para repouso discreto (sem poluir o histórico do terminal a cada segundo com mensagens repetitivas).
2. Adicionar atualização de título do painel no tmux quando o worker estiver aguardando aprovação humana (ex.: `[ ⚠️ APROVAÇÃO NECESSÁRIA ❯ forge-builder ]`).
3. Restaurar o título normal do painel assim que a aprovação for concedida e o trabalho for retomado.
4. Testes cobrindo a transição de status visual no mock de tmux runner.

## Critérios de Aceitação Executáveis
```bash
go test -v ./internal/pipeline -run WorkerApprovalAlert
```
- O título do painel é alterado temporariamente para sinalizar aprovação pendente.
- A saída em ociosidade não gera linhas repetidas infinitas no terminal.
- `go build ./... && go vet ./...` sem erros.
