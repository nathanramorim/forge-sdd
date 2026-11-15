# feat-70-05 — Reviewer: Veredito, Rework e Anti-Loop

**Branch:** `feat/70-pipeline-multi-sessao-orquestrada`
**Fase:** 70-05
**Tamanho:** `M`
**Depende de:** feat-70-02, feat-70-04
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
O Reviewer confere o PR contra spec, ADRs, rules e padrões. O modelo julga; o Go aplica a regra de fluxo: aprovar, devolver ao Builder com tentativa incrementada ou escalar. É essa separação que torna os cenários 4, 5 e 6 testáveis sem modelo.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** uma máquina de estados pequena sobre o barramento e as notificações, mais definição de agente e skill (`ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Parecer semântico | `internal/queue` (`Verdict`, `Finding`), skill `sdd-governanca` | Contrato JSON com `veredito`, `achados`, evidência obrigatória | reusar | O veredito do PR usa o mesmo contrato |
| Verificações S-01..S-07 e FLX-08 | `.agents/skills/sdd-governanca` | Regras do revisor semântico | reusar | O `reviewer` herda e acrescenta ADRs e rules |
| Subagente somente leitura | `.claude/agents/governance.md` | `tools: Read, Grep, Glob, Bash` | reusar o formato | `reviewer` com as mesmas ferramentas; comenta no PR só por `gh` quando `comment_on_pr` |
| Barramento | `internal/bus` (70-02) | Reenfileirar e concluir | reusar | `pr_rework` e `escalation` são operações dele |
| Notificações | `internal/notify` (70-06) | Canais e inbox | reusar | Aprovação e escalonamento notificam por ele |
| Worker | `internal/pipeline` (70-04) | `--role reviewer` | reusar | Mesmo laço do Builder |

## Dependências novas
- Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- `FLX-06`, `FLX-08`, `SEG-06`, `SEG-05`, `DUP-10`, `COD-01`, `COD-03`, `TST-03`.
- Rules: `sdd-fluxo-e-portao.md`, `sdd-seguranca.md`, `sdd-codigo-e-testes.md`.
- Gerais: `FLX-01`, `FLX-02`, `COD-02`, `COD-05`, `TST-04`, `TST-05`.

## Escopo
1. `forge-sdd bus review <job> --verdict-file <arquivo>` (`ApplyVerdict`): 
   - `conforme`: conclui o job de revisão, grava `pr_approved` no inbox com o link do PR, aciona os canais e imprime a instrução de mover o card para `status_on_pr_approved`.
   - `violacao` com `attempt < max_rework_attempts`: `attempt + 1`, `type: pr_rework`, `priority: high`, `rework_feedback` estruturado (arquivo, linha ou trecho, regra ou ADR, descrição, ajuste esperado) e devolução a `builder/pending`.
   - `violacao` com `attempt >= max_rework_attempts`: nenhum job novo; grava `escalation` no inbox.
   - `duvida`: escala para a pessoa (sem rework automático), registrando o que precisa de confirmação.
2. Agente `reviewer` (definição para Claude, Gemini e Copilot) e `references/pr-review.md` da skill `sdd-governanca`: confere spec e critérios, ADRs aplicáveis (inclusive depreciadas), rules, padrões, segurança, cobertura e saúde dos testes; produz o veredito JSON; comenta no PR só se configurado.
3. `worker --role reviewer`: se o agente terminar sem aplicar veredito, o job falha com motivo.
4. Feedback e diff são dados não confiáveis; tamanho do feedback limitado e caracteres de controle removidos.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./internal/pipeline/... ./internal/bus/...
go test ./internal/pipeline -run 'TestReview_NaoConformeDevolveComTentativaIncrementada|TestReview_LimiteDeTentativasEscalaParaInbox|TestReview_AprovadoGravaInboxENotifica'
```
- Cenário 4: reprovação devolve com `attempt` incrementado, `pr_rework`, `high` e feedback preenchido; o job volta antes de qualquer `normal`.
- Cenário 5: em `attempt == max` não há novo job e o inbox recebe `escalation`; nunca há loop.
- Cenário 6: aprovação grava o link do PR no inbox e dispara só os canais habilitados.
- `duvida` escala sem rework; veredito sem evidência literal em `violacao` é recusado.
- Violações (`TST-03`): veredito inválido, job inexistente, veredito aplicado duas vezes (idempotente), feedback com injeção e controle.
- `.claude/agents/reviewer.md` mantém `tools` somente leitura (teste de conteúdo).
