# feat-88-01 — Campo `autonomy` na configuração do pipeline

**Iniciativa:** 88-modo-pipeline-avancado
**Branch:** feat/88-modo-pipeline-avancado
**Status:** ready
**Depende de:** nenhuma

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** `internal/config` (ARQ-01): campo de um conceito já dono de `forge.json` (DUP-01).

## Decisões do projeto aplicadas
- ADR-0005 (permissões e `forge.json`), ADR-0013 (nova, task 88-03).
- Rules abaixo; nenhuma ADR depreciada usada como base.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Campo de modo em forge.json | `internal/config/forge.go` (`ForgeConfig`, `Validate`) | Campos `session_mode`, `context` com validação estrita | estender | Mesmo padrão de enumeração validada, sem pacote novo. |
| Permissão sem aprovação | `config.PermissionArgs` | `full` = bypass por CLI | reusar | `advanced` apenas exige `permissions: full`. |

## Dependências novas
- Nenhuma (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`, `SEG-01`, `COD-01`, `COD-03`, `TST-01`
- Rules: `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-codigo-e-testes.md`.

## Escopo
1. `ForgeConfig.Autonomy` (`normal` padrão | `advanced`) e `IsolatedConfirmed bool`, com `omitempty`; arquivo antigo sem a chave segue válido.
2. Validação: valor desconhecido é erro; `advanced` exige `permissions: full` e `isolated_confirmed: true`.
3. `SaveForgeConfig` grava de forma atômica (já existente).

## Critérios de Aceitação Executáveis
```bash
go test -v -race ./internal/config -run TestAutonomy
```
- `TestAutonomy_PadraoNormal`, `TestAutonomy_ValorInvalido`, `TestAutonomy_AvancadoExigeFullEIsolado`, `TestAutonomy_ArquivoAntigoValido`.
- `go vet ./...` limpo.
