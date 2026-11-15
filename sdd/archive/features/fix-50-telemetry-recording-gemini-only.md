# fix/telemetry-recording-gemini-only

**Branch:** `fix/telemetry-recording-gemini-only`
**Fase:** 50
**Depende de:** —
**Status:** `done`

## Objetivo

Corrigir a gravação de telemetria (`sdd/.metrics/session-<ISO8601>.json`)
funcionando de fato apenas para o agente Gemini, apesar de `sdd/.sddrc`
listar `telemetry.enabled` como uma configuração global válida para
qualquer agente. Reportado pelo usuário após observar o comportamento em
outro projeto scaffoldado pelo forge-sdd.

## Causa raiz

A instrução concreta de "Gravação de Métricas" (passo/schema de
`sdd/.metrics/session-<ISO8601>.json`, campo `feature`, tratamento de
`outcome: blocked/rejected`, estimativa de tokens) só existe, por escrito,
dentro do arquivo de papel **Orquestrador**:

- `internal/scaffold/templates/agents/gemini/.gemini/skills/orquestrador.chatmode.md.tmpl`
- `internal/scaffold/templates/.github/chatmodes/orquestrador.chatmode.md.tmpl` (Copilot)

Mas esse arquivo só é efetivamente acionado durante uma sessão real se o
prompt de entrada (`/proxima-feature`) delegar para ele:

1. **Gemini** — `.gemini/prompts/proxima-feature.prompt.md.tmpl` abre com
   "Acione a lógica de **Orquestrador**", então a gravação de métricas do
   passo "Guardrail (Close)" é de fato herdada em toda sessão. É o único
   agente onde isso acontece hoje.
2. **Copilot** — `.github/prompts/proxima-feature.prompt.md.tmpl` **não**
   segue o padrão de delegação usado pelos outros prompts do próprio
   Copilot (`revisar.prompt.md.tmpl` e `archive.prompt.md.tmpl` abrem com
   "Invoque o Revisor"/"Invoque o Archivist"). Em vez disso, inlina 4 passos
   próprios que nunca mencionam telemetria nem o guardrail de budget de
   `progress.md` (também presente no Orquestrador). Resultado: o chatmode
   `orquestrador.chatmode.md.tmpl` do Copilot existe no repositório, mas é
   órfão — nunca é referenciado por nenhum prompt, então nunca roda.
3. **Claude** — não existe um papel "Orquestrador" dedicado (a estrutura do
   Claude é só `.claude/commands/*.prompt.md`, sem chatmodes por papel).
   `CLAUDE.md.tmpl` resume o lifecycle em uma linha vaga ("CLOSE
   (Orquestrador): atualizar progress, **métricas**, archive se
   necessário") sem nenhuma instrução executável — sem caminho de schema,
   sem campos, sem estimativa de tokens. `proxima-feature.prompt.md.tmpl`
   do Claude também nunca menciona telemetria.

Ou seja: a gravação de métricas nunca foi migrada/replicada para os fluxos
reais de sessão do Claude e do Copilot quando esses agentes ganharam seus
próprios `proxima-feature.prompt.md.tmpl` independentes do Orquestrador.

## Critérios de Aceitação Executáveis

1. `internal/scaffold/templates/agents/claude/.claude/commands/proxima-feature.prompt.md.tmpl`
   ganha um passo final explícito de gravação de telemetria (condicionado a
   `telemetry.enabled` em `sdd/.sddrc`), com o mesmo nível de detalhe do
   Orquestrador do Gemini: caminho `sdd/.metrics/session-<ISO8601>.json`,
   campo `feature` com caminho relativo completo, `outcome: blocked/rejected`
   para sessões incompletas, e estimativa realista de `tokens_input`/
   `tokens_output`.
2. `internal/scaffold/templates/.github/prompts/proxima-feature.prompt.md.tmpl`
   passa a delegar para o Orquestrador (`Invoque o Orquestrador`), seguindo o
   mesmo padrão de `revisar.prompt.md.tmpl`/`archive.prompt.md.tmpl`, em vez
   de inlinar passos próprios — isso restaura tanto a gravação de métricas
   quanto o guardrail de budget de `progress.md` que hoje só o Gemini aplica.
3. `CLAUDE.md.tmpl` (passo 5 do Lifecycle) referencia o caminho concreto do
   schema de métricas (`sdd/.metrics/session-<ISO8601>.json`) em vez do
   termo genérico "métricas", para que o agente tenha a informação mesmo
   fora do comando `/proxima-feature`.
4. Sessão real de dogfood neste repositório (Claude) grava
   `sdd/.metrics/session-<ISO8601>.json` ao concluir esta fix, validando a
   correção fim a fim.
5. Testes de golden fixtures (`internal/scaffold/testdata/golden/`)
   regenerados refletindo os três arquivos alterados.

## Handoff

Fix de conteúdo de templates (prompts/chatmodes) — sem mudança em
`internal/scaffold/scaffold.go`/`cheatsheet.go`. Escopo deliberadamente
restrito ao fluxo `/proxima-feature` (onde vive o "CLOSE" da sessão); os
comandos `/nova-feature` e `/novo-fix` não gravam telemetria em nenhum
agente hoje (incluindo Gemini) pois eles apenas criam a especificação e
encerram com handoff para `/proxima-feature` — comportamento correto e
fora do escopo desta correção.
