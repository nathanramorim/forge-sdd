# feat-123-02 — Mod forge-painel: status line, alertas, telemetria, painel, aviso e /gastos

Tasks 1.3, 2.1, 2.2, 2.3 e 2.4. Trilha C. Depende de feat-123-01.

## Modelo do módulo
**Enxuto** (`ARQ-11`). Módulos: .agents/plugins/forge-painel (TypeScript, fora do binário).

## Decisões do projeto aplicadas
RN-1 a RN-12 da spec; ADR-0030; SEG-05; SEG-08.

## O que já existe
Criar o mod em `.agents/plugins/forge-painel/` com `.claude-plugin/plugin.json`, `hooks/hooks.json`, `hooks/register.tsx`, `types/index.d.ts` e testes. Status line via `session.measure`; toast em 70% e 90%; telemetria no schema 1.0 só da sessão interativa e subagentes (mudo com `FORGE_WORKER`); painel lendo `.forge/runtime/` e logs sem escrever; aviso em `agent.spawn` sem bloquear; `/gastos` lendo `sdd/.metrics`. Sem rede, sem `$.model`.

## Dependências novas
Nenhuma de runtime (`ARQ-04`, `SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`SEG-05`, `SEG-08`, `DUP-04`, `FLX-04`

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
claude plugin validate .agents/plugins/forge-painel
claude plugin test .agents/plugins/forge-painel
! grep -rnE "fetch\(|\$\.model\." .agents/plugins/forge-painel/hooks
```
