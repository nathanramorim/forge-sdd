# feat-123-03 — Instalação do mod valida o manifesto

Task 1.4. Trilha C. Depende de feat-123-01.

## Modelo do módulo
**Enxuto** (`ARQ-11`). Módulos: internal/skills, cmd/forge-sdd/skills_instalar.go.

## Decisões do projeto aplicadas
ADR-0025, ADR-0030, SEG-01, SEG-02.

## O que já existe
`forge skills instalar plugin <slug> --de` já copia para `.agents/plugins/<slug>/` (ADR-0025). Estender: quando houver `.claude-plugin/plugin.json`, validar campos do contrato da ADR-0030 (permissões e versões mínimas) e recusar manifesto incompleto.

## Dependências novas
Nenhuma (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`SEG-01`, `SEG-02`, `TST-03`

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test ./internal/skills/... ./cmd/forge-sdd/ -run Instalar
go build ./... && go vet ./...
```
