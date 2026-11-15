# feat-75-02 — Protocolo de abertura da Discovery

**Branch:** `feat/75-contexto-do-projeto-e-contingencia-tmux`
**Fase:** 75-02
**Tamanho:** `M`
**Depende de:** 75-01
**Status:** `pending` (FLX-01 aprovado)

## Contexto & Motivação
Mesmo com o `brief`, a Discovery só o usa se toda sessão começar por ele. Hoje o prompt pede ADRs e rules só "antes de quebrar tarefas" e a leitura mínima manda carregar só os IDs que a tarefa cita. Esta task muda o protocolo e exige que a spec cite as decisões que aplica.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** textos de skill e templates, uma constante em Go e goldens; sem pacote novo (`ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Prompt da sessão | `internal/pipeline/launcher.go` (`DiscoveryPrompt`) | Pede ADRs e rules antes de quebrar tarefas | estender | Manda rodar `forge-sdd brief` no início e citar as decisões na spec. |
| Skill da Concepção | `.agents/skills/sdd-concepcao/SKILL.md`, `references/{discovery,nova-feature,novo-fix}.md` | "Leitura mínima: só os IDs da tarefa" | estender | Concepção abre com o `brief`; seção nova na spec. |
| Skill do split | `.agents/skills/sdd-planejamento/references/split-features.md` | Carrega ADRs e rules antes de fatiar | estender | Também abre com o `brief`. |
| Instruções por agente | `internal/scaffold/templates/agents/{claude,gemini,openai}/` | "Leia só os IDs da tarefa" | estender | READ-MIN da Discovery: `brief` primeiro. |
| Cópia viva em `.agents/` | `.agents/skills/` do repositório | Espelho dos templates | estender | Mantida em sincronia (`forge-sdd skills sync`). |
| Orçamento de contexto | `internal/gate/gate.go` (`P-24`) | Mede adaptadores e skills | reusar | As mudanças cabem no orçamento. |

## Dependências novas
- Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `FLX-01`, `GOV-01`, `DUP-02`, `TST-02`.
- Rules: `sdd-fluxo-e-portao.md`, `sdd-reuso-e-duplicacao.md`.
- Gerais: `FLX-02`, `COD-02`, `TST-05`.

## Escopo
1. `DiscoveryPrompt` (ASCII): primeira ação é `forge-sdd brief`; se o estado for `rascunho`, propor `/constitution` antes de qualquer spec; citar as decisões do projeto na spec; usar `/implementar` quando pronta.
2. `sdd-concepcao` e os `references` de discovery, nova-feature, novo-fix e split-features: passo "rode `forge-sdd brief`" antes de tudo; a regra de leitura mínima passa a valer só para Builder e Governança.
3. Seção obrigatória nova nas specs: `## Decisões do projeto aplicadas` (ADR, ID, módulo e fluxo que a spec respeita; conflito sem ADR não passa pelo Gate 1, `GOV-01`); atualizar os templates de spec e a regra "não aprove" da skill.
4. `CLAUDE.md`, `GEMINI.md` e `AGENTS.md` do scaffold: READ-MIN da Discovery começa pelo `brief`.
5. Cópia em `.agents/skills` sincronizada; goldens atualizados; adaptadores dentro do orçamento `P-24`.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./internal/pipeline/... ./internal/scaffold/... ./cmd/forge-sdd/...
```
- Teste de conteúdo: `DiscoveryPrompt`, `sdd-concepcao` (template e cópia), os quatro `references` e os três arquivos de instrução citam `forge-sdd brief`; a skill exige `Decisões do projeto aplicadas`.
- `TestScaffold_*` e goldens atualizados (`go test -update` só depois de revisar o diff).
- `forge-sdd gate` com `P-24` verde.
