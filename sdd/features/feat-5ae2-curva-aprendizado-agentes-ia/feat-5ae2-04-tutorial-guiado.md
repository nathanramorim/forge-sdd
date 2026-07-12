# Feature 5ae2-04 — Onboarding Guiado (`/tutorial`)

Ensina o ciclo completo do SDD por exemplo vivo, dentro do próprio projeto do usuário, em vez de exigir leitura prévia da Wiki.

## Critérios de Aceitação Executáveis

1. Deve existir um prompt `/tutorial` (replicado nos três agentes) que executa um discovery e uma feature de exemplo fictícios, percorrendo o ciclo `PLAN → ACT → CLOSE` de ponta a ponta.
2. Os artefatos gerados pelo tutorial devem ficar isolados em `sdd/features/_tutorial/` e `sdd/discovery/_tutorial/`, sem tocar `sdd/features/index.md` real nem a telemetria de sessões normais.
3. `forge-sdd init` deve aceitar uma flag `--tutorial` que, ao final do scaffold, já deixa sinalizado (arquivo de flag ou instrução impressa) que o usuário deve rodar `/tutorial` no agente escolhido.
4. `--dry-run` continua não criando arquivos mesmo com `--tutorial` (Regra 9 da Constituição).
