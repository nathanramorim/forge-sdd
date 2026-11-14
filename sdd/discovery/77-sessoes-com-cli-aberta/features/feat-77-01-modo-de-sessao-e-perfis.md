# feat-77-01 — Modo de sessão e perfis

**Branch:** `feat/77-sessoes-com-cli-aberta`
**Fase:** 77-01
**Tamanho:** `M`
**Depende de:** —
**Status:** `pending` (FLX-01 aprovado)

## Contexto & Motivação
Configuração de `session_mode` (`live` vs `headless`), `context` (`clear` vs `continuous`), tempos de prontidão e `reset_command` por perfil CLI, garantindo compatibilidade retroativa para projetos com `forge.json` antigo.

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** Novos campos de configuração, validações, compatibilidade retroativa, extensão de perfis interativos e menu interativo (`ARQ-11`, `COD-01`).

## Decisões do projeto aplicadas
- `D-8 (ADR-0007)`: Config antiga sem `session_mode` mantém padrão `headless`.
- `D-5 (ADR-0007)`: Limpeza de contexto (`clear` padrão) usando comando do perfil (`/clear` no claude/gemini/copilot, `/new` no codex).
- `ARQ-01`, `COD-03`: Regras e parsing em `internal/config/`, handlers de CLI finos.

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Struct de config do forge | `internal/config/forge.go` | `ForgeConfig` | estender | Adicionar `session_mode`, `context`, `live` e `reset_command` no `CLIProfile`. |
| Perfis padrão de CLI | `internal/config/forge_profiles.go` | `DefaultProfiles` | estender | Preencher `reset_command` e argumentos interativos para Builder e Reviewer. |
| Menu interativo | `cmd/forge-sdd/menu.go` | `pipelineMenu` | estender | Permitir configurar modo de sessão e contexto. |

## Dependências novas
- Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `COD-01`, `COD-03`, `TST-04`, `ARQ-01`.
- Rules: `sdd-codigo-e-testes.md`, `sdd-arquitetura.md`.

## Escopo
1. Estender `ForgeConfig` com `SessionMode`, `ContextMode`, `LiveConfig` (`StartupWaitSeconds`, `ReadyQuietSeconds`) e `ResetCommand` em `CLIProfile`.
2. Validar valores permitidos: `session_mode` (`live`, `headless`), `context` (`clear`, `continuous`).
3. Manter default `headless` quando ausente em arquivos pré-existentes.
4. Adicionar argumentos interativos de Builder e Reviewer nos perfis.
5. Adicionar opções no menu do forge para configurar modo e contexto.

## Critérios de Aceitação Executáveis
```bash
go test -v ./internal/config -run SessionMode
```
- Testes cobrem validação de enum, padrões retrocompatíveis e reset_command por CLI.
- `go build ./... && go vet ./...` sem erros.
