# fix-76 — Worker trata agente indisponível (autenticação, plano ou conta)

**Branch:** `fix/76-worker-agente-indisponivel` (a partir da `feat/72`)
**Origem:** Discovery 72 (visão limpa do worker, limite de uso)
**Tamanho:** `M`
**Status:** `implemented`

## Feature Correlacionada
- **Arquivo de origem corrigido:** `internal/pipeline/worker.go` (`processJob`, caminho de falha) e `internal/pipeline/usagelimit.go` (único classificador de falha de agente).
- **Causa-raiz:** o worker só distingue **limite de uso**. Qualquer outra saída diferente de zero vira `bus.Fail`, com o motivo "o agente saiu com código N: <última linha>". Para uma falha de **ambiente** (conta sem direito, plano descontinuado, não logado, chave inválida) isso é errado em três pontos: (1) o job vai para `failed` e só volta com intervenção manual; (2) o worker segue no laço e queima o próximo job do mesmo jeito; (3) a Discovery não é avisada, a pessoa descobre olhando o painel.
- **Evidência (relato de uso real, 14/11/2026):** `sessions.builder: "gemini"`; o Gemini CLI falhou na autenticação antes de ler qualquer task com `IneligibleTierError: This client is no longer supported for Gemini Code Assist for individuals. ... migrate to the Antigravity suite of products`. O job `job-20261114-043237-abcc-…` foi para `builder/failed`; as 24 tasks seguiram na fila com 0 tentativas e nenhum arquivo foi alterado. No painel `forge-builder` apareceu o dump bruto (`tierId: 'free-tier'`, `tierName: ...`) e dezenas de linhas `Skill conflict detected` entre `~/.agents/skills` e `~/.gemini/skills`. Na d15b, três jobs seguidos já tinham falhado por limite de uso do Claude.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** extensão do worker e do `LiveView` com um classificador puro e um arquivo novo pequeno, sem processo novo nem concorrência nova (`ARQ-11`).

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Classificador de falha do agente | `internal/pipeline/usagelimit.go` (`DetectUsageLimit`, regex forte e genérica, última linha útil) | Só limite de uso | reusar o padrão | Novo `DetectAgentUnavailable` com a mesma regra: padrão forte vale em qualquer das últimas linhas; genérico só na última linha ou com marcador de erro. |
| Devolver o job sem gastar tentativa | `internal/bus/store.go` (`Requeue`), `worker.go` (`deferJob`) | `Requeue` mantém `attempt` e registra evento | reusar | A falha de ambiente usa `Requeue` com o motivo `agente indisponível`. |
| Aviso à Discovery | `internal/notify` (`TypeInfo`, `Dispatch`), `deferJob` | Aviso `info` no inbox, id determinístico | reusar | O worker avisa o inbox com o motivo e a ação. |
| Resultado do job | `worker.go` (`OnJobDone`, `OutcomeDone/Failed/Deferred`) | Três resultados | estender | `OutcomeUnavailable`. |
| Ruído do agente na tela | `internal/pipeline/liveview.go` (`hookNoiseRe`, `hookShown`) | Colapsa o ruído de hooks globais em uma linha | estender | Mesmo mecanismo para as linhas de inicialização do Gemini. |
| Mascarar segredo no que sai | `LiveView` + `gate.MaskSecrets` | Já mascara | reusar | A mensagem do agente exibida passa pela máscara (`SEG-05`). |
| Dica de CLI ausente | `internal/pipeline/preflight.go` (`agentFiles[...].InstallIt`) | Ação "instale" por CLI | reusar | A ação de "agente indisponível" segue o mesmo padrão de texto por CLI. |
| Skills globais do usuário | `~/.agents/skills`, `~/.gemini/skills` | Pastas do usuário; o Forge só sincroniza `.claude/skills` e `.gemini/skills` **do projeto** (`internal/skills/skills.go`) | não mexer | `SEG-02`: o Forge não edita configuração global. Os conflitos vêm da duplicação feita fora do Forge. |

## Dependências novas
- Nenhuma. Stdlib Go (`regexp`, `strings`) (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `SEG-05`, `SEG-02`, `COD-01`, `COD-03`, `TST-03`, `TST-04`, `DUP-01`, `DUP-02`.
- Rules: `sdd-seguranca.md`, `sdd-codigo-e-testes.md`, `sdd-reuso-e-duplicacao.md`.
- Gerais: `FLX-02`, `COD-02`, `TST-05`.

## Escopo
1. **Classificador `DetectAgentUnavailable(tail []string, cli string) (AgentUnavailable, bool)`** (arquivo novo `internal/pipeline/agentstate.go`). Padrões fortes (vale em qualquer das últimas 12 linhas): `IneligibleTierError`, `no longer supported for`, `migrate to the Antigravity`, `not logged in`, `please (log ?in|run .* login|authenticate)`, `invalid api key`, `authentication (failed|error|required)`. Padrões genéricos (só na última linha ou com marcador de erro): `\b401\b`, `unauthorized`, `forbidden`, `permission[_ ]denied` junto de `api|token|credential`. Só vale com saída diferente de zero.
2. **Resultado e `Hint` por CLI:** `AgentUnavailable{Message, Hint}`; a `Message` é a linha saneada e mascarada. Hints: gemini → "o plano gratuito do Gemini Code Assist foi descontinuado para este cliente: troque o agente desta sessão (forge → Configurar) ou use o agy (Antigravity)"; claude → "rode claude login"; copilot → "rode gh auth login"; codex → "rode codex login"; agy → "entre pelo agy e tente de novo"; demais → "revise o login da CLI ou troque o agente desta sessão".
3. **Caminho do worker:** em `processJob`, antes do `bus.Fail`, se saída ≠ 0 e `DetectAgentUnavailable` casar: `bus.Requeue` (sem gastar tentativa, motivo `agente indisponível: <cli>`), aviso `info` no inbox com a mensagem e a ação, linha `✗ <cli> indisponível: <mensagem>` e `→ <hint>` na tela, `OutcomeUnavailable`, e o **worker encerra** com código ≠ 0 e a mesma mensagem (não segue consumindo a fila). Se o `Requeue` falhar, cai no `bus.Fail` de hoje.
4. **Ruído de inicialização do Gemini na tela:** o `LiveView` colapsa em uma linha cada família, uma vez por job: `Skill conflict detected` → `(N skills duplicadas nas suas pastas globais: ignorado; o log bruto guarda tudo)`; `YOLO mode is enabled` e `Ripgrep is not available` → omitidas. O log bruto continua com tudo; `--raw` mostra tudo.
5. **Não faz parte:** instalar, remover ou editar skills fora do projeto; sonda de autenticação antes do job (custa uma chamada ao agente e muda por CLI); trocar de agente sozinho. A troca continua sendo escolha da pessoa.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./internal/pipeline/... ./internal/bus/... ./cmd/forge-sdd/...
```
- `TestDetectAgentUnavailable`: o texto do relato (`IneligibleTierError ... migrate to the Antigravity`) é detectado com o hint do gemini; `not logged in` e `invalid api key` idem; texto de código ou conversa que cita "401" no meio da saída não casa; saída com código 0 nunca casa.
- `TestWorker_AgenteIndisponivel`: agente falso falha com a mensagem do Gemini; o job volta a `pending` com `attempt` intacto (não vai a `failed`), o inbox recebe um `info` com a ação, o worker retorna erro e **não** consome o segundo job da fila.
- `TestWorker_AgenteIndisponivelRequeueFalha`: se o `Requeue` falha, o job vai para `failed` com o motivo e a mensagem.
- `TestLiveView_RuidoGemini`: 40 linhas `Skill conflict detected` viram uma linha com a contagem; `YOLO mode`/`Ripgrep` somem da tela e continuam no log bruto; com `--raw` aparecem todas.
- `TestUsageLimit_Inalterado`: os casos atuais de limite de uso continuam passando (sem regressão); limite de uso tem precedência quando a mesma saída traz as duas mensagens.
- `forge-sdd gate` sem falhas; nenhum segredo na mensagem exibida (`TestLiveView_MascaraSegredo` continua verde).
