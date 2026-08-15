# Feature 01-01 — Telemetria Code-Enforced

Pré-requisito das demais features desta discovery. Substitui a gravação de `sdd/.metrics/session-<ISO8601>.json` — hoje uma instrução em linguagem natural no último passo de um prompt longo, disparada apenas por `/proxima-feature` — por um mecanismo determinístico, resolvendo o sintoma relatado (telemetria habilitada em `.sddrc` mas ausente de forma inconsistente entre projetos e agentes).

## Critérios de Aceitação Executáveis

1. Existe um mecanismo determinístico (subcomando do binário Go, ex. `forge-sdd session record`, ou script/hook) que grava `sdd/.metrics/session-<ISO8601>.json` respeitando o schema já existente em `sdd/.metrics/schema.json`, sem depender de instrução LLM para o passo de escrita em si.
2. O mecanismo é acionado nos três pontos de saída de sessão hoje existentes: fim de `/proxima-feature`, fim de `/revisar`, fim de `/novo-fix` — não apenas no primeiro.
3. Uma sessão que termina em `/revisar` (sem chegar a `/proxima-feature`) produz um `session-*.json` válido, com `outcome` refletindo o estado real (`approved`/`rejected`/`blocked`).
4. `go build` e `go vet ./...` passam; teste cobrindo o novo mecanismo de gravação.
5. Não quebra o contrato atual dos comandos públicos do CLI (Regra 10 da Constituição) — se novo subcomando, é aditivo.

## Status: todo
