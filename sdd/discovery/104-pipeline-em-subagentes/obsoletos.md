# Obsoletos e enganosos (20/11/2026), fonte para o README

Modelo atual: **3 sessões** (Discovery, Builder, Reviewer) mais a **aprovação humana**; entrada por `forge pipeline init` e `forge up`; modo `chat` opt-in (ADR-0015). "Funciona" = existe no binário compilado da branch; nada abaixo foi executado de ponta a ponta (`não verificado`).

## Papéis atuais (fonte: definição no arquivo)

| Papel | O que faz | Aciona | Definição |
|---|---|---|---|
| Discovery | Conversa, gera discovery, critérios, plano e tasks; você aprova a spec | `/discovery`, `/split-features`; `forge dispatch` entrega ao Builder | `.claude/agents/pm-architect.md`, `planner.md` |
| Builder | Implementa as tasks da fila em ondas paralelas e abre o PR | `forge up`, `forge worker`; `/implementar` | `.claude/agents/builder.md` |
| Reviewer | Camada A (`forge gate`) mais Camada B somente leitura; emite veredito JSON | job `pr_review`; `forge bus review`; `/revisar` | `.claude/agents/reviewer.md`, `governance.md` |
| Aprovação humana | Gate 1 (spec) e aprovação final do PR | você, no chat e no PR | `sdd/FLOW.md` |

Modelo por papel: `agent_models` (`discovery`, `builder`, `reviewer`) em `internal/config/agent_models.go`.

## Obsoleto (não use)

| Item | Onde | Por que | Substituto | Funciona |
|---|---|---|---|---|
| "5 estações" como fluxo | `README.md` (linha 169, "Esteira de 5 especialistas", 3.0.0-beta), `.claude/agents/lider.md` ("coordena as 5 estações"), `sdd/spec/adr/adr-0003-*` | produto virou 3 sessões (ADR-0004; visão registrada em `overview.md`) | 3 sessões mais aprovação | texto apenas |
| Agente `lider` | `.claude/agents/lider.md` | descrito como orquestrador de 5 estações; o Go orquestra (ADR-0015 D-2) | `forge up` e o chat do modo `chat` | não verificado |
| Nome `forge-sdd` como comando | docs antigas | ADR-0010: `forge` é o nome oficial; `forge-sdd` é alias no npm | `forge` | alias: não verificado |
| Versão `v4.7.7-beta` nas specs | `sdd/spec/*.md` (corrigido aqui) | tag git é `v5.0.0` | v5.0.0 | n/a |
| `npm/package.json` `4.9.1` | `npm/package.json` | diverge da tag `v5.0.0`; não corrigido aqui | bump no checklist de release | n/a |

## Legado que ainda funciona

| Item | Onde | Evidência | Substituto | Funciona |
|---|---|---|---|---|
| Filas por estação | `forge queue stations` ("5 filas de estações"), `queue handoff`, `internal/queue` | aparece no `forge queue --help` do binário compilado | `queue.json` por iniciativa e barramento de jobs (ADR-0002, 0004) | listado no `--help`; não executado |
| Skills `sdd-*` por estação | `.agents/skills/sdd-concepcao`, `-planejamento`, `-build`, `-governanca`, `-operacao` | ADR-0003; os comandos `.claude/commands/*` ainda apontam para elas | seguem como apoio das 3 sessões | existem; não verificado em execução |
| Modo `sessions` (tmux) | `internal/pipeline/tmux.go`, ADR-0008 | convive com `chat` (ADR-0015); é o padrão | modo `chat` é opt-in | não verificado |
| `/implement`, `/proxima-feature` | `.claude/commands/implement.md`, `proxima-feature.md` | apontam para o mesmo arquivo de `/implementar` (alias) | `/implementar` | alias; não verificado |

## Papel que continua válido

| Item | Onde | Observação |
|---|---|---|
| `archivist` | `.claude/agents/archivist.md` | mantém `status.md` até 1 KB (GOV-03); auto-archive em `internal/pipeline/archive.go` |
| `migrator` | `.claude/agents/migrator.md` | apoia `/upgrade-sdd`; `internal/scaffold/migrate.go` |
| `/toca-ai` | `.claude/commands/toca-ai.md`, `forge pipeline toca-ai` | planeja trabalho longo e entrega à esteira |
| `/inbox`, `/status`, `/doctor`, `/tutorial`, `/archive`, `/nova-feature`, `/novo-fix`, `/c4-architecture`, `/constitution`, `/upgrade-sdd` | `.claude/commands/` | comandos de apoio, fora das 3 sessões |
| Skill `novo-subcomando-cli-sdd` | `.agents/skills/novo-subcomando-cli-sdd/` | guia interno para criar subcomando; não é estação |

## Enganosos a corrigir no README (não editado aqui)
- Contagem de comandos e ADRs: o índice só ia até ADR-0012; agora vai até 0020.
- `forge review` não existe; é `forge bus review`.
- A LP em `docs/` tem a Edição Missões ativa até 2026-11-23 (ADR-0019); depois some sozinha, sem tocar o README.
- `site/` (Astro) e a LP estática (`docs/`) são coisas distintas; a LP é a que o Worker publica (ADR-0020).
- Não verificado: conteúdo atual de `README.md` e de `npm/README.md` (não lidos).
