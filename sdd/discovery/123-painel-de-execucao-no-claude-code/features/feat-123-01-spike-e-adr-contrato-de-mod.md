# feat-123-01 — Spike de carregamento e ADR-0030 do contrato de mod

Tasks 1.1 e 1.2. Trilha C.

## Modelo do módulo
**Enxuto** (`ARQ-11`). Módulos: sdd/spec/adr, sdd/discovery/123.

## Decisões do projeto aplicadas
ADR-0025, ADR-0003, DUP-10, FLX-05.

## O que já existe
Nenhum mod hoje. Spike: como o Claude Code carrega um plugin do projeto sem `/plugin install` (configuração do projeto, `--plugin-dir`, `claude -p`); registrar achados no `status.md`. ADR-0030: manifesto do mod (nome, versão, versões mínimas do Claude Code e do forge, permissões: rede, escrita, ganchos), alinhado ao vocabulário da 109 e à ADR-0025.

## Dependências novas
Nenhuma (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`FLX-05`, `GOV-01`, `SEG-07`

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
test -f sdd/spec/adr/adr-0030-*.md
grep -q "Spike" sdd/discovery/123-painel-de-execucao-no-claude-code/status.md
```
