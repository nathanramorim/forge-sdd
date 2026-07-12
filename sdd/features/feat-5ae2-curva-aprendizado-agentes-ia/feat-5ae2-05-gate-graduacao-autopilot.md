# Feature 5ae2-05 — Gate de Graduação para Autopilot

Preserva a função didática da metodologia: evita que um iniciante ative autonomia total (`.sdd-auto-pilot`) sem nunca ter praticado o ciclo manual.

## Critérios de Aceitação Executáveis

1. A criação do arquivo `sdd/.sdd-auto-pilot` (via CLI ou skill do Orquestrador) deve ser bloqueada por padrão até que existam **N ciclos completos** (`outcome: done`) registrados em `sdd/.metrics/session-*.json`, com N configurável e default 3.
2. Deve existir uma flag explícita de bypass consciente (ex: `--skip-graduation`), documentada como decisão deliberada do usuário avançado, não um atalho oculto.
3. Mensagem de bloqueio deve informar quantos ciclos faltam para a graduação automática.
4. Teste cobrindo: menos de N ciclos done (bloqueado), N ou mais ciclos done (liberado), bypass explícito (liberado independente da contagem).
