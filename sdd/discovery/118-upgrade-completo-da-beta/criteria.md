# 118 — Critérios e arquitetura

## Restrições (IDs da Constituição)

`ARQ-03` (templates por `embed.FS`), `ARQ-05` (`--dry-run` sem efeito e completo), `ARQ-06` (subcomando novo pede ADR), `ARQ-11`, `SEG-01`, `SEG-02` (gravação atômica; preserva domínio do usuário), `TST-02`, `TST-03`, `DUP-02`, `GOV-01`.

## Modelo do módulo

**Completo**: toca `scaffold`, `config`, `doctor`, `update` e os testes, com um teste ponta a ponta do upgrade. Justificativa: o defeito atravessa as camadas.

## Decisões do projeto aplicadas

ADR-0021 (3 sessões), ADR-0022/0023 (PR e CI em dois pontos), ADR-0025 (instalação limitada ao projeto), fix-106-09 (contrato de atualização sem impacto), fix-106-14a e 14b (hook e script).

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Entrega do hook e do script | fix-106-14a/14b (spec em PR #256) | cobre os achados 1, 7 (parte) e 8 (hook) | **reusar** como ondas 1 e 2 | evita duplicar spec |
| Contrato de atualização | fix-106-09 | furos 1 a 6; sem implementação | **estender** (achados 3, 6, 10) | mesma causa-raiz |
| Persistência da configuração | `.sddrc` / `internal/config/config.go:159` | não guarda stack e banco | **estender** | achado 2 |
| Mesclagem segura de JSON | `hooks.go` `mergeClaudeHook*` | preserva, não duplica | **reusar** | padrão correto |
| Teste de upgrade ponta a ponta | nenhum | `testdata` só tem o golden atual | **criar** fixture 5.0 | fix-106-09 furo 6 |

## Dependências novas

Nenhuma (`SEG-07`).

## Critérios de aceitação executáveis (iniciativa)

```bash
set -o pipefail
# num projeto 5.0 (fixture), update --upgrade deixa tudo coerente
go test ./internal/scaffold ./cmd/forge-sdd -run 'TestUpgradeDe50' -count=1
# nenhum hook sem script, em init e em update
go test ./internal/scaffold -run 'TestNenhumHookSemScript' -count=1
# update nunca perde stack/banco nem edição em arquivo preservado
go test ./internal/scaffold -run 'TestUpdateNaoPerdeDadoDoUsuario' -count=1
```

Os critérios por feature vivem em cada `features/feat-118-NN-*.md`.

## Diagramas

```mermaid
sequenceDiagram
    actor U as Usuário (5.0)
    participant N as npm (@beta)
    participant F as forge update
    participant A as Agente (/upgrade-sdd)
    U->>N: instala 5.2.0-beta
    U->>F: forge update --upgrade
    F->>F: grava scripts/ (pr_guard) ANTES do hook
    F->>F: persiste stack/banco; acrescenta o que falta
    F-->>U: plano completo no --dry-run
    U->>A: /upgrade-sdd (CI, política)
    A-->>U: workflow próprio (opt-in), nada reescrito
    U->>F: forge doctor (confere arquivos, não só texto)
```
