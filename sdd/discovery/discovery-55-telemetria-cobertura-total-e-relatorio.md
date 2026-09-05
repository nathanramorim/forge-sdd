# Discovery 55 — Telemetria com Cobertura Total e Relatório de Métricas

## Por quê

Hoje a telemetria do Forge-SDD só é gravada em 3 dos 13 comandos
(`/proxima-feature`, `/revisar`, `/novo-fix`). Qualquer sessão que passa
por `/discovery`, `/split-features`, `/nova-feature` ou `/archive` sem
depois passar por um desses três nunca gera `sdd/.metrics/session-*.json`
— métrica perdida silenciosamente, sem aviso. Além disso, não existe
nenhum comando que responda "quantos tokens gastei nesta feature?",
"qual modelo foi usado em cada sessão?" ou "há quanto tempo esse projeto
está ativo?" — só um contador simples (aprovado/reprovado/bloqueado)
embutido no `doctor`.

## Para quem

Mantenedores de projetos Forge-SDD (a começar pelo próprio forge-sdd)
que quantificam custo/esforço de IA por entrega, e querem auditar em
retrospecto onde o tempo/token foi gasto.

## Como (macro)

1. **Cobertura total nos comandos que mudam estado do SDD:** estender a
   gravação determinística de telemetria (`forge-sdd session record`,
   já usada em 3 comandos) para `/discovery`, `/split-features`,
   `/nova-feature` e `/archive` — os comandos que criam ou modificam
   artefatos do ciclo SDD. Comandos de leitura/diagnóstico (`/status`,
   `/doctor`, `/constitution`, `/c4-architecture`, `/upgrade-sdd`,
   `/tutorial`) ficam fora de escopo por decisão do usuário nesta sessão
   (não produzem entrega, gerariam ruído no relatório).
2. **Novo comando `forge-sdd report`:** determinístico (Go, não
   depende de o LLM montar nada), agrega todos os `sdd/.metrics/session-
   *.json` e imprime no terminal:
   - Tokens de entrada/saída somados por feature/fix/discovery.
   - Modelo(s) usado(s) em cada um.
   - Duração de cada sessão e duração total.
   - Tempo decorrido desde a métrica mais antiga registrada até a mais
     recente ("idade medida do projeto" — não a idade real do
     repositório, que pode ser maior que o histórico de telemetria).
3. Sem geração de arquivo por padrão — saída direta no terminal (decisão
   do usuário nesta sessão), mantendo o padrão de comandos determinísticos
   do CLI (mesmo estilo de `forge-sdd doctor`).

## Fora de escopo (decisões desta sessão)

- Comandos de leitura/diagnóstico não gravam telemetria.
- Sem exportação para arquivo (Markdown/CSV) na primeira entrega — pode
  virar um fix futuro se a necessidade aparecer.
- Sessões já gravadas antes desta feature (sem campo `type`) continuam
  válidas — o relatório infere o tipo (discovery/feature/fix) a partir do
  caminho em `feature`, sem exigir re-gravação nem migração de schema.
