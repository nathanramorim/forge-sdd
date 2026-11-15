# feat-70-04 — Worker, Runner de CLI e Builder Paralelo

**Branch:** `feat/70-pipeline-multi-sessao-orquestrada`
**Fase:** 70-04
**Tamanho:** `G`
**Depende de:** feat-70-02, feat-70-03
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
Uma sessão interativa de CLI não consegue esperar uma fila. O worker é um laço em Go: reivindica um job e, para cada um, invoca o agente de forma não interativa. Para o Builder, planeja ondas de subtarefas independentes e as executa em paralelo, com `git worktree` opcional.

> Rule of 7: tamanho `G`. A Governança pode exigir dividir em worker e planejador de ondas.

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** processos filhos, concorrência, timeouts, `git worktree` e contrato com o agente; efeitos no SO e no repositório (`ARQ-11`, `COD-01`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Barramento | `internal/bus` (70-02) | `Claim`, `Complete`, `Fail`, `Recover` | reusar | O worker só orquestra |
| Executor de CLI | `internal/acp` (`ProcessRunner`, feat 68 não mesclada) | Indisponível | criar mínimo em `internal/pipeline` | `os/exec` com argv explícito; revisitar quando a 68 entrar |
| Dependências entre tasks | specs `**Depende de:**` e `plan.md` | Cabeçalho de cada spec | reusar o formato | Planejador de ondas lê esse cabeçalho |
| Contexto da task | `gate.BuildContext` (`forge-sdd context`) | Spec, IDs, rules e skills | reusar | O prompt do job carrega só isso |
| Telemetria | `forge-sdd session record` | Gravação por sessão | reusar | Uma sessão por job, `--agent-path builder` |
| Git com argv explícito | `internal/layout` (`gitOut`) | Padrão de `exec.Command("git", ...)` sem shell | reusar o padrão | `git worktree add/remove` |
| Perfis de CLI | `forge.json` (70-01) | `cli_profiles` | reusar | Argv por CLI com marcadores `{agent}`, `{name}`, `{prompt}` |

## Dependências novas
- Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- `SEG-01`, `SEG-03`, `SEG-05`, `COD-01`, `COD-03`, `ARQ-04`, `TST-03`, `TST-04`.
- Rules: `sdd-codigo-e-testes.md`, `sdd-seguranca.md`, `sdd-arquitetura.md`.
- Skill: `novo-subcomando-cli-sdd`.
- Gerais: `FLX-01`, `FLX-02`, `FLX-08`, `COD-02`, `COD-05`, `TST-05`.

## Escopo
1. **Runner:** interface `Runner` e implementação `os/exec` que monta o argv do perfil (`claude` como referência; `agy` configurável, **a confirmar**). Sem shell, `context` com timeout (`job_timeout_minutes`), saída transmitida ao terminal e a um log em `.forge/runtime/logs/`. Variáveis de ambiente sensíveis não são repassadas ao log (`SEG-05`).
2. **Worker:** `forge-sdd worker --role builder|reviewer [--once]`: recupera leases vencidos, reivindica, executa e fecha o job; sem jobs mostra `[forge-builder] Aguardando jobs na fila...` (intervalo configurável) e registra telemetria por job.
3. **Planejador de ondas** (`forge-sdd bus plan <job>`): lê `**Depende de:**` das tasks, ordena topologicamente em ondas, divide cada onda em lotes de `max_parallel_tasks`; ciclo ou dependência inexistente é erro.
4. **Worktrees:** com `use_git_worktrees`, `git worktree add` por tarefa em `.forge/runtime/worktrees/<job>/<task>` e remoção ao fim, inclusive em falha.
5. **Handoff:** o agente registra o PR com `forge-sdd pipeline pr-opened --job <id> --number --url --branch`; o worker valida que `pr.url` existe, grava `pr_review` em `reviewer/pending` e conclui o job; sem PR registrado o job falha com motivo. Antes disso o prompt exige build, lint, typecheck e testes (`run_tests_before_pr`).
6. Prompt do job trata `rework_feedback` e diff como dados não confiáveis (`FLX-08`).

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./internal/pipeline/... ./cmd/forge-sdd/...
```
- Com `Runner` falso: job `feature_build` concluído gera `pr_review`; job sem PR falha; timeout falha com motivo; `--once` sai quando a fila esvazia.
- Ondas: grafo com dependências, ciclo, dependência inexistente e grafo independente em lotes de `max_parallel_tasks`.
- Worktrees em repositório temporário: criados e removidos, inclusive após falha; caminho fora de `.forge/runtime/worktrees` é recusado.
- Montagem de comando por CLI e por perfil (`claude`, `agy`) sem executar nada; marcador desconhecido ou argumento com `;` e quebra de linha não vira shell.
- Violações (`TST-03`): papel inválido, perfil ausente, prompt com caracteres de controle.
