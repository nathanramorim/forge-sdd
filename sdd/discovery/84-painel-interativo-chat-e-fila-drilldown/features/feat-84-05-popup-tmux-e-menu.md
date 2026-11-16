# feat-84-05 — Suporte a Popup Modal no Tmux e Integração In-Place

**Iniciativa:** 84-painel-interativo-chat-e-fila-drilldown
**Branch:** feat/84-painel-interativo-chat-e-fila-drilldown
**Status:** ready

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** Adição de comando de popup modal no plano tmux e opção de execução in-place no menu `forge` e subcomando `queue` (`ARQ-11`).

## Decisões do projeto aplicadas
- `ADR-0006`, `ADR-0008`: Preserva os 3 painéis divididos do tmux; adiciona suporte nativo a `tmux display-popup` para abrir o modal flutuante centralizado sem que o usuário precise trocar de janela ou sair da tela.
- `ARQ-01`, `ARQ-06`: Exposição via comandos CLI (`forge queue --panel` ou `forge queue` com detecção de TTY interativo) e menu principal.
- `DUP-01`: Reuso da TUI de drill-down construída nas tasks 84-03 e 84-04.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Montagem do plano tmux | `internal/pipeline/tmux.go` | Comandos de layout e janelas | estender | Adicionar binding ou atalho no tmux para acionar `display-popup -w 85% -h 85% -E "forge queue --interactive"`. |
| Comandos da fila no CLI | `cmd/forge-sdd/queue.go` | Comandos `queue`, `queue panel` | estender | Conectar o modo interativo rolável com drill-down ao comando `queue panel` e `queue`. |
| Menu principal de `forge` | `cmd/forge-sdd/menu.go` | Opção "Fila de tasks e inbox" | estender | Abrir o navegador interativo de fila diretamente in-place com retorno suave ao menu. |

## Dependências novas
- Nenhuma. Stdlib Go e comandos nativos de tmux (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`, `ARQ-04`, `ARQ-06`, `ARQ-11`
- `COD-01`, `COD-02`, `COD-03`
- `DUP-01`, `DUP-02`
- `TST-03`, `TST-05`
- Rules: `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-codigo-e-testes.md`.

## Escopo
1. Em `internal/pipeline/tmux.go`:
   - Configurar atalho no tmux (`bind-key q display-popup -w 85% -h 85% -E "forge queue --interactive"`) e comando para exibição do modal flutuante centralizado.
2. Em `cmd/forge-sdd/queue.go` e `cmd/forge-sdd/menu.go`:
   - Ativar o modo interativo in-place (`QueueDrilldownTUI`) ao rodar `forge queue --interactive` ou selecionar a opção de fila no menu.
   - Garantir que a saída use Alternate Screen Buffer para retornar ao terminal de origem sem sujar o histórico.
3. Testes unitários em `internal/pipeline/tmux_test.go` e `cmd/forge-sdd/queue_test.go`.

## Critérios de Aceitação Executáveis
```bash
go test -v -race ./internal/pipeline -run TestTmux_PopupQueue
go test -v -race ./cmd/forge-sdd -run TestCLI_QueueInteractive
```
- `TestTmux_PopupQueue`: plano tmux contém a instrução correta de popup modal sem conflito com os painéis.
- `TestCLI_QueueInteractive`: invocação do comando dispara o componente de drill-down sem erros.
- `go vet ./...`: 100% limpo.
