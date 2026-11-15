# feat-79-03 — Menu Unificado e Enxuto (Mapa De/Para)

**Branch:** `feat/79-clareza-pipeline-e-identificacao-tmux`
**Fase:** 79-03
**Tamanho:** `M`
**Depende de:** 79-02
**Status:** `completed` (FLX-01 aprovado)

## Contexto & Motivação
O menu atual em `cmd/forge-sdd/menu.go` divide as opções entre "Pipeline multi-sessão" e "Estrutura padrão", obrigando o desenvolvedor a navegar em submenus para ações frequentes (iniciar sessões, checar status, ver fila ou rodar doctor). Esta feature implementa a unificação completa dos menus conforme o mapa De/Para aprovado, exibindo cabeçalho contextual com o projeto e branch atual.

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** Refatoração da máquina de navegação do menu CLI (`cmd/forge-sdd/menu.go`), integração de subcomandos via `runSubcommand`, prompter simulado e testes de interface (`ARQ-11`, `COD-01`).

## Decisões do projeto aplicadas
- `ADR-0005`: Menu Navegável e Agente por Sessão do Pipeline.
- `D2`: Framework CLI Cobra v1.8+.
- `D8`: Idioma pt-BR padrão em opções e mensagens.

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Lógica de menu | `cmd/forge-sdd/menu.go` | `runMenu`, `pipelineMenu`, `standardMenu`, `configMenu` | refatorar | Fundir submenus no menu raiz conforme mapa De/Para; manter `configMenu` para opções secundárias. |
| Leitura de brief e estado | `internal/gate` | `BuildBrief` | reusar | Extrair estado do projeto (pronta / rascunho) para o cabeçalho do menu. |
| Testes de menu interativo | `cmd/forge-sdd/menu_test.go` | Simulação com prompter | estender | Testar o fluxo direto do novo menu raiz. |

## Dependências novas
- Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`, `COD-01`, `COD-05`, `TST-04`.
- Rules: `sdd-arquitetura.md`, `sdd-codigo-e-testes.md`.

## Escopo
1. Atualizar o menu raiz (`runMenu`) para apresentar:
   - Cabeçalho: `Forge SDD <version> — <projeto> (<estado> · branch: <branch>)`
   - Opções diretas:
     1. Iniciar esteira (up)
     2. Status das iniciativas e esteira
     3. Fila de tasks e inbox de notificações
     4. Diagnóstico de saúde e regras (Doctor & Gate)
     5. Configurações (agentes, terminal, modo ao vivo/headless)
     6. Ajuda e comandos do CLI
     7. Sair
2. Eliminar os submenus redundantes "Pipeline multi-sessão" e "Estrutura padrão".
3. Unificar fila e inbox num fluxo único; unificar doctor e gate num submenu de diagnóstico.
4. Preservar o atalho para `/constitution` quando o estado for `rascunho`.
5. Atualizar os testes de menu em `cmd/forge-sdd/menu_test.go`.

## Critérios de Aceitação Executáveis
```bash
go test -v ./cmd/forge-sdd -run Menu
```
- A navegação do menu raiz reflete o novo mapa De/Para.
- Nenhum subcomando de linha de comando (`status`, `up`, `doctor`, `gate`) é quebrado.
- `go build ./... && go vet ./...` sem erros.
