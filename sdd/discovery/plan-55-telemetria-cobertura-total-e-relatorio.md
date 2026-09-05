# Plano 55 — Telemetria com Cobertura Total e Relatório de Métricas

## Roadmap preliminar

1. **Classificação de tipo (base):** função `ClassifySessionType(feature
   string) string` em `cmd/forge-sdd/session.go` — infere
   discovery/feature/fix pelo prefixo do path, sem quebrar
   `AggregateSessionMetrics` existente.
2. **Comando `forge-sdd report`:** novo `cmd/forge-sdd/report.go`,
   reaproveitando a leitura de `sdd/.metrics/` e a classificação da
   task 1; imprime tabela (tokens, modelos, duração) + idade medida do
   projeto. Testes unitários com fixtures mistas.
3. **Cobertura de telemetria em `/discovery`:** passo "Gravação de
   Métricas" no comando canônico + template, outcome
   `approved`/`criterio-atendido=false` (discovery não tem "critério
   atendido" no sentido de feature — decisão de nomenclatura do outcome
   fica para o Builder na implementação, seguindo o padrão mais próximo
   já usado em `novo-fix`).
4. **Cobertura de telemetria em `/split-features`, `/nova-feature` e
   `/archive`:** mesmo padrão da task 3, um comando por vez.
5. **Regeneração de golden testdata:** `internal/scaffold/testdata/golden/`
   precisa refletir os novos passos nos arquivos de comando afetados.
6. **Documentação:** `sdd/FLOW.md` e `README.md` citam o novo comando
   `forge-sdd report` como parte do ciclo de observabilidade.

## Estimativa de quebra em features

- **Bounded contexts distintos:** motor de agregação/relatório (Go, CLI)
  vs. cobertura de telemetria nos prompts (Markdown, 4 comandos × N
  agentes). Candidato a 2 features dentro de uma pasta `feat-55-*/`:
  - `feat-55-01-relatorio-metricas` (tasks 1–2, código Go + testes)
  - `feat-55-02-cobertura-telemetria-comandos` (tasks 3–5, prompts +
    golden tests)
  - Task 6 (documentação) pode entrar em qualquer uma das duas ou ficar
    como subtask final de `feat-55-02`.
- Sem dependência circular: `feat-55-02` só depende de `feat-55-01`
  existir (usa o mesmo `session record`, já existente hoje) — pode até
  rodar em paralelo se o Builder preferir, já que não compartilham
  arquivo de código.
- Decisão final de quebra e nomes fica para `/split-features`.
