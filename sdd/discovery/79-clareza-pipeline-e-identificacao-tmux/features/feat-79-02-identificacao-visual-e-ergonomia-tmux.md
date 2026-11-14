# feat-79-02 — Identificação Visual do Projeto e Ergonomia no Tmux

**Branch:** `feat/79-clareza-pipeline-e-identificacao-tmux`
**Fase:** 79-02
**Tamanho:** `M`
**Depende de:** —
**Status:** `completed` (FLX-01 aprovado)

## Contexto & Motivação
Ao rodar múltiplos projetos com o Forge SDD, as janelas e sessões do tmux são visualmente idênticas, causando confusão sobre qual projeto está aberto. Além disso, ter dois painéis inteiros ocupados com loops ruidosos de espera desperdiça espaço. Esta feature adiciona identificação inequívoca do projeto na barra inferior do tmux, nas bordas dos painéis e no nome da janela, além de introduzir a estrutura de monitoramento compacto de esteira.

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** Modifica a criação do plano de comandos do tmux (`internal/pipeline/tmux.go`), manipulação de strings de terminal, sanitização e testes unitários com runner simulado (`ARQ-11`, `COD-01`).

## Decisões do projeto aplicadas
- `ADR-0006`: Visibilidade, tmux e Entrega ao Builder.
- `ADR-0008`: Sessões com CLI Aberta no Tmux (Modo Live).
- `ARQ-01`, `ARQ-04`: Regras de execução em `internal/pipeline/`, sem dependências externas adicionais.

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Construção do plano tmux | `internal/pipeline/tmux.go` | `BuildTmuxPlan` | estender | Adicionar opções de sessão (`status-left`, `status-left-length`, `pane-border-format`, nome de janela). |
| Nome do projeto saneado | `internal/pipeline/tmux.go` | `TmuxSessionName` | reusar | Extrair base limpa do nome para exibição amigável. |
| Testes de plano tmux | `internal/pipeline/tmux_test.go` | Asserções de argv gerados | estender | Validar inclusão dos novos comandos `set-option` e `set-window-option`. |

## Dependências novas
- Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`, `ARQ-04`, `COD-01`, `COD-03`, `TST-04`.
- Rules: `sdd-arquitetura.md`, `sdd-codigo-e-testes.md`.

## Escopo
1. Em `BuildTmuxPlan`:
   - Configurar `status-left` no alvo da sessão (`-t exact`): `#[fg=colour16,bg=colour48,bold] 📦 <projeto> #[default] `.
   - Configurar `status-left-length` para 60 para não truncar nomes longos de projeto.
   - Configurar `window-name` para `forge:<projeto>` tanto no modo `live` quanto `headless`.
   - Atualizar `pane-border-format` para exibir `[ <projeto> ❯ #{pane_title} ]`.
2. Garantir que todas as diretivas usem o alvo exato da sessão (`-t =name`), sem aplicar opções globais (`-g`), preservando o ambiente global do usuário.
3. Testes unitários em `internal/pipeline/tmux_test.go` cobrindo o plano com as opções visuais.

## Critérios de Aceitação Executáveis
```bash
go test -v ./internal/pipeline -run TmuxPlanProjectIdentity
```
- O plano inclui as diretivas de status e bordas com o nome do projeto.
- `go test -v ./internal/pipeline` passa sem regressões.
