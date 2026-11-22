# fix-106-05 — A limpeza de worktrees só enxerga 2 dos 61 e nunca roda no chat

**Branch:** `docs/106-specs-worktree-modo-ask` (PR único das três specs 05, 06 e 07)
**Origem:** Iniciativa 97, `feat-97-02` (`feat-96-01`, limpeza de worktrees), mesclada em 22/11/2026; e iniciativa 106 (branch e PR por feature)
**Tamanho:** `M`
**Status:** `spec` (aguarda Gate 1)

## Sintoma

Em 22/11/2026 havia **61 worktrees** além do principal. A limpeza entregue pela `feat-97-02` cobre **2** deles. O dono relatou: "tá ficando muito worktree sujo que já não é mais utilizado, isso também não deve acontecer".

Auditoria só de leitura (script `git worktree list --porcelain` + três testes de segurança):

| Origem | Qtde | Visível para a limpeza da 97-02? |
|---|---|---|
| `.claude/worktrees/agent-*` (agentes do Claude Code) | 23 | **não** |
| `/tmp/*` | 23 | **não** |
| `~/git/forge-sdd-*` (pastas irmãs) | 13 | **não** |
| `.forge/runtime/worktrees` (pipeline) | 2 | **sim** |

Classificação: **20 removíveis sem perda** (limpos, commits no remoto ou na `main`), **17 sujos** (mudanças não commitadas), **16 sem remoto** (commits só locais) e **8 com commits não enviados**. Os 20 seguros foram removidos manualmente nesta sessão (mais 4 criados nela); restam 37.

## Feature Correlacionada

- **Arquivo de origem:** `internal/pipeline/worktree_gc.go` (`ListStaleWorktrees`, `SweepWorktrees`) e `cmd/forge-sdd/up.go` (`checkStaleWorktrees`).
- **Causa-raiz (duas):**
  1. **Fonte errada.** `ListStaleWorktrees` lê `layout.Worktrees(dir)` (`.forge/runtime/worktrees/<job>/<task>`) e decide "seguro" pelo **estado do job** no barramento e da task na fila. Worktree de agente do Claude Code, de `/tmp` ou de pasta irmã **não tem job**, então nunca entra.
  2. **Ponto de entrada errado.** `checkStaleWorktrees` roda em `forge up`, `forge worker` e `forge inbox`, mas `inbox.go:30` faz `if !hook { checkStaleWorktrees(...) }`: o `inbox --hook`, que roda a cada prompt no chat, **pula de propósito**. No modo chat a limpeza não roda nunca.
- **Evidência do defeito:** `git worktree list | wc -l` devolve 62; a limpeza lista 2.

## Modelo do módulo

**Enxuto** (`ARQ-11`): estende `internal/pipeline/worktree_gc.go`, sem pacote novo, e acrescenta uma flag a um subcomando existente. A dependência aponta para dentro.

## Decisões do projeto aplicadas

- `DUP-01` — **um** dono do conceito "worktree ocioso": `worktree_gc.go`. A regra nova entra ali, não em um segundo coletor.
- `SEG-01` — tudo que vem do `git` é dado externo; falha ao consultar **não** remove (trata como ativo).
- `COD-03` — erro de `git worktree remove` é enriquecido e reportado; nada é descartado.
- `ARQ-01` — a classificação é função pura em `internal/pipeline`; o `cmd` só mostra o menu e chama.
- Regra do dono: toda decisão é componente de ask (ver fix-106-07). Remoção só depois de confirmação em menu.
- `FLX-02` — branch e PR por feature criam worktrees; quem cria também deve poder recolher.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Classificação por job e remoção | `internal/pipeline/worktree_gc.go` | `ListStaleWorktrees`, `SweepWorktrees`; regras por estado do job | **reusar** como regra para worktrees do pipeline |
| Subcomando de limpeza | `cmd/forge-sdd/worker.go:241` (`forge pipeline worktree gc`) | existe, lê só `.forge/runtime` | **estender** com `--todos` (sem subcomando novo, `ARQ-06` intacto) |
| Aviso na sessão | `cmd/forge-sdd/up.go` (`checkStaleWorktrees`) | chamado em `up`, `worker`, `inbox` sem `--hook` | **reusar**; acrescentar o ponto do chat |
| Lista real de worktrees | `git worktree list --porcelain` | enxerga **todas** as origens | **usar como fonte** (nova função `ListarWorktreesGit`) |
| Ponto de entrada do chat | `inbox --hook` (UserPromptSubmit) | pula a checagem de propósito | **não mudar** o hook por prompt (custo); usar o fim de entrega e o `doctor` |
| Seleção múltipla em menu | `internal/survey` (`RunMultiSelect`) | existe | **reusar** |
| Aviso de saúde | `cmd/forge-sdd/doctor.go` | sem checagem de worktrees | **estender** com a contagem |

## Dependências novas (`SEG-07`)

`Nenhuma`.

## Rules e IDs da Constituição Aplicáveis

`ARQ-01`, `ARQ-11`, `COD-03`, `COD-04`, `DUP-01`, `FLX-02`, `SEG-01`, `TST-01`, `TST-03`.

Rules: `sdd-arquitetura`, `sdd-codigo-e-testes`, `sdd-fluxo-e-portao`.

## Escopo

**Dentro:**
1. `ListarWorktreesGit(dir)`: lê `git worktree list --porcelain` e devolve todos, com origem (pipeline, agente, tmp, irmã).
2. `ClassificarWorktree(wt)` (função pura): **seguro** só se passar nos **três testes** — (a) sem mudança local (nem arquivo não rastreado), (b) sem commit à frente do remoto **ou** HEAD já na principal, (c) remoto existe. Qualquer dúvida ou erro do `git` vira **não seguro**. Para worktree do pipeline vale a regra por job **e** a nova; a mais restritiva vence.
3. Nunca remover: o worktree principal, o do diretório corrente, e qualquer travado (`locked`).
4. `forge pipeline worktree gc --todos` lista e remove, **sem `--force`**, só o que o usuário marcou em menu.
5. **Pontos de disparo no modo chat:** (a) ao fim de uma entrega (PR mesclado ou fix concluído), oferecer a limpeza em menu; (b) `forge doctor` avisa "N worktrees ociosos" e "M não removíveis (sujos ou sem remoto)".
6. Relatório curto por categoria (seguro, sujo, sem remoto, não enviado), como o da auditoria.

**Fora:**
- Remover worktree sujo, sem remoto ou com commit não enviado (perderia trabalho). Esses ficam listados para o dono decidir.
- Impedir o Claude Code de criar worktrees de agente (não é do forge).
- Rodar a limpeza a cada prompt no `inbox --hook`.
- Apagar branches (só a pasta do worktree sai).

## Critérios de Aceitação Executáveis

```bash
set -o pipefail
# 1. A fonte vê todas as origens, não só .forge/runtime.
go test ./internal/pipeline/... -run 'TestListarWorktreesGitIncluiTodasAsOrigens'
# 2. A regra dos três testes decide corretamente.
go test ./internal/pipeline/... -run 'TestClassificarWorktree'
# 3. Nunca toca o principal, o corrente nem travado.
go test ./internal/pipeline/... -run 'TestGcNaoTocaPrincipalNemCorrenteNemTravado'
# 4. Sem --force: sujo e sem remoto não saem.
go test ./internal/pipeline/... -run 'TestGcRecusaSujoESemRemoto'
# 5. Comando e doctor.
go test ./cmd/forge-sdd/... -run 'TestWorktreeGcTodos|TestDoctorAvisaWorktreesOciosos'
# 6. Subcomando novo não foi criado.
! grep -rn 'Use: *"worktrees' cmd/forge-sdd || exit 1
go vet ./... && go test ./...
```

- **Teste negativo obrigatório (`TST-03`):** worktree com um arquivo não rastreado, um com commit local sem remoto e um com `git` falhando **continuam em disco** depois do `gc --todos`.
- **Prova de cobertura:** num repositório de teste com worktrees de `.claude/worktrees`, `/tmp` e pasta irmã, a lista devolve as três origens (hoje devolve só a do pipeline).

## Riscos

- **R1 — Perda de trabalho** por classificar errado. Mitigação: três testes + dúvida vira "não seguro" + sem `--force` + confirmação em menu.
- **R2 — Worktree de agente em uso** por outra sessão. Mitigação: `locked` e o do cwd são intocáveis; remoção só com confirmação.
- **R3 — Custo do `git` em muitos worktrees.** Mitigação: consulta sob demanda (fim de entrega e `doctor`), não por prompt.
