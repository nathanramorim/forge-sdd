# Plano 44 — Aprimoramento do Fluxo Decisório do Forge-SDD

## Roadmap preliminar

| # | Feature | Depende de | Paralelizável (worktree) |
|---|---------|-----------|---------------------------|
| 44-01 | Sabatina ao final de `/discovery`, `/split-features`, `/nova-feature` | — | Não (mesmos 3 arquivos de prompt, edição sequencial evita conflito) |
| 44-02 | `sdd/scripts/worktree.sh` (open/close) + comando `/worktree` no chat | — | Sim, com 44-01 |
| 44-03 | `active_release_branch` em `.sddrc` + leitura em `/novo-fix`/`/nova-feature` | 44-01 (usa o mesmo ponto de sabatina para confirmar a branch) | Não |
| 44-04 | `/status` audita PRs abertos (`gh pr list`) e versão NPM não publicada | — | Sim, com 44-01/44-02 |

## Validação de dependências entre tasks

- **44-01** e **44-03** tocam os mesmos arquivos de prompt (`/nova-feature`, `/novo-fix`) em pontos próximos (final do fluxo) — recomendação: sequenciais, não paralelas, para evitar merge conflict.
- **44-02** e **44-04** não têm overlap de arquivos com as demais (novo script isolado + prompt `/status` isolado) — candidatas a rodar em worktrees paralelas entre si e em relação a 44-01/44-03.

## Sugestão de paralelismo (mecânico, via script)

Tasks sem dependência direta entre si podem ser desenvolvidas em worktrees isoladas. Comandos sugeridos (mecânicos — nenhuma lógica gerada por IA, apenas invocação):

```bash
sdd/scripts/worktree.sh open feat/44-02-worktree-script
sdd/scripts/worktree.sh open feat/44-04-status-git-audit
```

Ao concluir cada task, antes de abrir o PR:

```bash
sdd/scripts/worktree.sh close feat/44-02-worktree-script
sdd/scripts/worktree.sh close feat/44-04-status-git-audit
```

## Branch de destino dos PRs

Como estas 4 tasks compõem uma única entrega (aprimoramento de fluxo), sugere-se criar uma branch de release `feat/release-44-fluxo-sdd` e gravar `active_release_branch: "feat/release-44-fluxo-sdd"` em `sdd/.sddrc` antes de iniciar 44-01. Todos os PRs de 44-01 a 44-04 devem mirar essa branch; a branch de release é então mesclada em `main` como release única ao final.

## Sabatina desta sessão (registrada)

- Ativação da sabatina: **sempre ativa, com opção de pular** (usuário pode responder "pular" para aceitar o default).
- Worktree: script mecânico em `sdd/scripts/worktree.sh`, **acionável tanto pelo orquestrador (como sugestão de comando) quanto diretamente pelo usuário no chat**; fechamento sugerido ao final da implementação de cada task, junto da criação do PR.
- Direcionamento de PR: via **nova chave `active_release_branch` em `sdd/.sddrc`**.
- Checagem de release em `/status`: **comparação com o registro NPM** (`npm view forge-sdd version`).

## Handoff para `/split-features`

Arquivos gerados nesta discovery:
- `sdd/discovery/discovery-44-aprimoramento-fluxo-sdd.md`
- `sdd/discovery/criteria-44-aprimoramento-fluxo-sdd.md`
- `sdd/discovery/plan-44-aprimoramento-fluxo-sdd.md`

Próximo passo: rodar `/split-features` para quebrar as 4 features acima em specs individuais dentro de `sdd/features/feat-44-aprimoramento-fluxo-sdd/`, preservando o mapa de dependência/paralelismo e a decisão de branch de release já registrados aqui.
