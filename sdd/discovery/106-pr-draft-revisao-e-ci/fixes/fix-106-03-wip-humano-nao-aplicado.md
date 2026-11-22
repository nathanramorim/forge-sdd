# fix-106-03 — O WIP humano é lido, mas nada o aplica

**Branch:** `fix/106-03-wip-humano-nao-aplicado`
**Origem:** Iniciativa 106 (ADR-0022, `feat-106-08`)
**Tamanho:** `M`
**Status:** `spec` (aguarda Gate 1)

## Sintoma

A Constituição declara `WIP humano | 3` e a rule `sdd-modos-de-projeto.md` (seção 4) promete: "no limite o orquestrador para e pergunta em menu". Na prática, em 22/11/2026 havia **5 specs aguardando o humano** (PRs #218, #211, #179, #175 e #164, todos em draft) e nenhum comando parou para avisar. O limite existe no papel e não barra nada.

## Apuração: a configuração foi burlada?

O dono perguntou se o limite foi contornado ou se havia modo automático configurado que parou de funcionar. **Não e não**, com evidência:

| Hipótese | Evidência | Conclusão |
|---|---|---|
| Alguém mudou o valor do WIP | `git log -S'WIP humano' -- sdd/memory/constitution.md` mostra um único commit, `7cbe3af` (#177, 2026-11-21), com o valor padrão `3`; nunca editado depois | **Não foi alterado** |
| O campo foi desligado | O valor está `3`, não `desligado` | **Não foi desligado** |
| Havia modo automático ligado que quebrou | `Início automático do Builder` está `desligado` (padrão) e `.forge/config/forge.json` não tem chave de início automático | **Nunca esteve ligado** |
| O laço automático existe e falhou | O código do laço da 107 **não está na `main`**: só o discovery foi mesclado (#122). Nenhum código lê `Início automático do Builder` | **Não existe ainda; não quebrou** |
| O modo de revisão humana foi contornado | `com-humano` (padrão) e **tem código** (`internal/pipeline/modo_humano.go`, `volta_da_spec.go`), **mas nenhuma função exportada desses arquivos tem chamador fora de testes** (corrigido em 22/11/2026; ver `fix-106-06`) | **Não foi contornado; também não é aplicado** |

Resumo: os três campos de modo (`WIP humano`, `Início automático do Builder` e `Modo de revisão humana`) foram escritos na Constituição **sem a ligação ao fluxo**. No último, o código existe e é testado, mas nada o chama (`fix-106-06`). O critério da 106 só conferia que a linha existia (`grep -F '| WIP humano | 3 |'`), então a lacuna passou.

## Feature Correlacionada

- **Arquivo de origem:** `internal/vcs/modos.go` (`WIPHumano`, `WIPLigado()`) e `.agents/rules/sdd-modos-de-projeto.md` seção 4.
- **Causa-raiz:** o leitor do campo foi entregue (`feat-106-05`) sem a task que **usa** o valor. Nenhum comando de concepção, `forge brief`, `forge status` ou `forge doctor` consulta `WIPLigado()`. `grep -rIn 'WIPLigado\|WIPHumano' internal cmd scripts` fora de `modos.go` e dos testes não devolve nada.
- **Evidência do defeito:** 5 itens aguardando o humano, com limite 3, sem aviso. `grep -rIn -i wip .agents/skills .claude/agents` não devolve nada.

## Modelo do módulo

**Completo** (`ARQ-11`): toca três pacotes (`internal/vcs`, `internal/gate`, `cmd/forge-sdd`) e cria uma operação nova no adaptador de VCS. A dependência aponta para dentro: `cmd` → `gate` → `vcs`.

## Decisões do projeto aplicadas

- ADR-0022 e a rule `sdd-modos-de-projeto.md` seção 4: o limite existe e "reduzir ou desligar exige aviso dos riscos e registro no `status.md`".
- `FLX-01`: nada avança sem Gate 1 do humano; o WIP protege a capacidade de revisão desse humano.
- `FLX-02`: operação de VCS passa pelo adaptador `internal/vcs`, não por `gh` solto.
- `DUP-01`: **um** ponto de leitura (`internal/vcs`) e **um** ponto de relato (`forge brief`, que os três comandos de concepção já rodam primeiro). Nenhum subcomando novo, para não mudar `ARQ-06` nem exigir ADR.
- `SEG-01`: se `gh` estiver ausente ou falhar, a contagem **não bloqueia** (o guard não inventa recusa sem prova); avisa "contagem indisponível".
- `TST-03`: teste negativo do bloqueio (ver critérios).

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Leitor do limite | `internal/vcs/modos.go` | `Modos.WIPHumano`, `WIPLigado()`, `LeModos`, `LeModosDaBranchBase` | **reusar** como está |
| Listagem de PRs | `internal/vcs` (`Adaptador`) | só `ProcurarPRAberto(dir, branch)`; **sem listar** | **criar** `ListarPRsDraft(dir) ([]PR, error)` no adaptador; `nenhum` devolve vazio |
| Ponto de relato | `internal/gate/brief.go` (`BuildBrief`) | já é o primeiro passo dos 3 comandos de concepção (`READ-MIN`) | **estender**: nova seção curta "Aguardando você" |
| Painel de status | `cmd/forge-sdd/status.go` | mostra esteira e fila; sem WIP | **estender** com a linha `Aguardando você: N/L` |
| Check-up | `cmd/forge-sdd/doctor.go` | sem checagem de WIP | **estender** com aviso quando acima do limite ou contagem indisponível |
| Instrução dos comandos de concepção | `.agents/skills/sdd-concepcao/references/{discovery,nova-feature,novo-fix}.md` | nenhuma cita WIP | **acrescentar** um passo: se o brief marcar limite, `AskUserQuestion` antes de criar spec |
| Início automático do Builder | `Modos` lê o campo; sem consumidor | **não implementar**: pertence à 107 (laço) | **fora do escopo**, registrado |
| Subcomando novo `forge wip` | — | exigiria mudar `ARQ-06` e abrir ADR | **não criar** (`DUP-01`) |

## Dependências novas (`SEG-07`)

`Nenhuma`.

## Rules e IDs da Constituição Aplicáveis

`ARQ-01`, `ARQ-02`, `ARQ-11`, `DUP-01`, `DUP-02`, `FLX-01`, `FLX-02`, `SEG-01`, `COD-03`, `TST-01`, `TST-03`.

Rules: `sdd-arquitetura`, `sdd-codigo-e-testes`, `sdd-fluxo-e-portao`, `sdd-modos-de-projeto`.

## Escopo

**Dentro:**
1. `ListarPRsDraft` no adaptador `internal/vcs` (`github`: `gh pr list --state open --draft`; `nenhum`: vazio; `azure-devops`: "não verificado", como o resto do adaptador).
2. Contagem e comparação com o limite, em `internal/gate` (função pura, testável sem rede).
3. `forge brief` ganha a seção "Aguardando você: N/L" e um marcador explícito `LIMITE` quando `N >= L` e o WIP está ligado. Fica fora quando o WIP é `desligado`.
4. `forge status` e `forge doctor` mostram o mesmo número; o doctor avisa quando a contagem está indisponível.
5. Os 3 comandos de concepção param no marcador `LIMITE` e perguntam em menu: **resolver um pendente primeiro**, **criar mesmo assim (registra no `status.md`)**, **cancelar**.

**Fora:**
- Início automático do Builder e o laço da 107.
- Bloquear a criação do PR por hook (`pr_guard.py`). Fica como evolução, se o aviso no comando não bastar.
- Decidir o que fazer com os 5 PRs de hoje (é decisão do dono, em Gate 1 de cada um).
- Contar PRs que não estão em draft. **Observação:** o #214 (Discovery da 98) está aberto **sem** draft, contra a política da ADR-0022 (PR de spec nasce em draft). A contagem por draft não o vê. Registrado como achado, não corrigido aqui.

## Critérios de Aceitação Executáveis

```bash
set -o pipefail
# 1. Adaptador lista PRs em draft (github com runner falso; nenhum devolve vazio).
go test ./internal/vcs/... -run 'TestListarPRsDraft'
# 2. A função de contagem decide o limite (abaixo, no limite, acima, WIP desligado).
go test ./internal/gate/... -run 'TestAguardandoHumano'
# 3. O brief relata o número e marca LIMITE só quando deve.
go test ./internal/gate/... -run 'TestBriefMostraWIPHumano|TestBriefMarcaLimiteDeWIP'
# 4. Status e doctor mostram o mesmo número.
go test ./cmd/forge-sdd/... -run 'TestStatusMostraWIPHumano|TestDoctorAvisaWIPHumano'
# 5. Os três comandos de concepção citam o marcador.
test "$(grep -l 'LIMITE' .agents/skills/sdd-concepcao/references/{discovery,nova-feature,novo-fix}.md | wc -l)" = 3
# 6. Nenhum subcomando novo foi criado (ARQ-06 intacto).
! grep -rn 'Use: *"wip' cmd/forge-sdd || exit 1
go vet ./... && go test ./...
```

- **Teste negativo obrigatório (`TST-03`):** com `WIP humano | 3` e 5 PRs em draft, `TestBriefMarcaLimiteDeWIP` prova que o brief traz `LIMITE`; com `WIP humano | desligado`, prova que **não** traz; com `gh` falhando, prova que **não bloqueia** e diz "contagem indisponível".
- **Prova de que o critério antigo era fraco:** o `grep -F '| WIP humano | 3 |'` da 106 passa mesmo sem nenhum consumidor. Este fix troca a conferência de texto por conferência de comportamento.

## Riscos

- **R1 — Latência e rede:** o `forge brief` passa a chamar `gh`. Mitigação: limite de tempo curto e falha sem bloquear (`SEG-01`).
- **R2 — Falso positivo por PRs antigos em draft** (ex.: #164, release 5.1.0-beta, superada pela 5.2.0-beta). Mitigação: o aviso lista os PRs, e o menu permite resolver ou fechar o antigo.
- **R3 — O limite vira ritual:** a opção "criar mesmo assim" existe de propósito, mas **exige registro no `status.md`**, como a rule já pede para reduzir o limite.
