# Fluxo da entrega com CI local primeiro (fix-106-10)

Quem faz o quê, do chat até o merge, e o **único** ponto em que o GitHub Actions roda.

```mermaid
sequenceDiagram
    autonumber
    actor H as Dono (humano)
    participant D as Discovery (chat)
    participant B as Builder (subagente)
    participant L as Máquina local<br/>(ci-local.sh)
    participant R as Reviewer (subagente)
    participant G as GitHub<br/>(PR + Actions)

    H->>D: pede a feature ou o fix
    D->>G: PR da spec em draft (sem CI)
    D->>H: Conferência 1, pede [Seguir]
    H->>D: [Seguir]
    D->>G: marca pronto e mescla a spec

    loop uma branch e um PR por feature
        D->>B: entrega a feature (worktree própria)
        B->>G: push de commits no PR em draft (sem CI)
        B->>L: scripts/ci-local.sh (build, vet, test, lint)
        L-->>B: selo ci-local.json com head_sha
        B->>D: PR draft pronto para revisão
        D->>R: revisa o PR
        R->>L: reexecuta testes e critérios da spec
        R-->>D: parecer (conforme ou violação) com head_sha
        alt violação
            D->>B: corrige e repete
        else conforme
            D->>G: gh pr ready (exige parecer + selo do HEAD)
            G->>G: ÚNICO run do CI: pronto?, test, lint
            G-->>D: checks verdes
            D->>G: merge squash (sem CI no push da main)
        end
    end
```

## Como ler

- Linhas até o `gh pr ready`: **zero** execução de Actions. Todo o trabalho e a validação rodam na máquina.
- `gh pr ready` só passa com parecer **e** selo local do mesmo `head_sha` do HEAD (`pr_guard.py`).
- O CI do GitHub dispara só em `ready_for_review` (ou `workflow_dispatch`) e confirma o que a máquina já provou.
- Precisa de nova rodada no GitHub? Volte o PR a draft e marque pronto de novo.

## Limites

- Selo local é auto-atestado: não prova que a máquina estava limpa. O CI único é a segunda opinião.
- Sem `push` na `main`: regressão por merge conflitante só aparece no PR seguinte.
