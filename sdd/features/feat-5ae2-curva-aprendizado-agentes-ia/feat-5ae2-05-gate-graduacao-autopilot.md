# Feature 5ae2-05 — Gate de Graduação para Autopilot

Preserva a função didática da metodologia: evita que um iniciante ative autonomia total (`.sdd-auto-pilot`) sem nunca ter praticado o ciclo manual.

## Critérios de Aceitação Executáveis

1. A criação do arquivo `sdd/.sdd-auto-pilot` (via CLI ou skill do Orquestrador) deve ser bloqueada por padrão até que existam **N ciclos completos** (`outcome: done`) registrados em `sdd/.metrics/session-*.json`, com N configurável e default 3.
2. Deve existir uma flag explícita de bypass consciente (ex: `--skip-graduation`), documentada como decisão deliberada do usuário avançado, não um atalho oculto.
3. Mensagem de bloqueio deve informar quantos ciclos faltam para a graduação automática.
4. Teste cobrindo: menos de N ciclos done (bloqueado), N ou mais ciclos done (liberado), bypass explícito (liberado independente da contagem).

## Status: done

Implementado como novo comando público `forge-sdd autopilot [diretório]` (`cmd/forge-sdd/autopilot.go`), independente da execução do próprio loop de autopilot (que segue apenas na branch `feat/cli-autopilot-autonomy`, ainda em teste). Conta ciclos com `"outcome": "approved"` em `sdd/.metrics/*.json` (schema real de `sdd/.metrics/schema.json.tmpl`); bloqueia a criação de `sdd/.sdd-auto-pilot` abaixo do mínimo (default 3, configurável via `--min-cycles`), com bypass consciente via `--skip-graduation`. Idempotente se o arquivo já existir. Testes em `cmd/forge-sdd/autopilot_test.go`. Quando o autopilot real for mesclado ao main, ele só precisa checar a existência de `sdd/.sdd-auto-pilot` — já é o contrato assumido no design da branch de autopilot.
