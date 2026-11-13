# CLAUDE.md — forge-sdd

forge-sdd: stack go. Fluxo por feature em `sdd/FLOW.md`.

## Ciclo por comando
1. **READ-MIN:** `forge-sdd status` e, para Builder e Governança, `sdd/memory/lessons.md`.
2. **PLAN:** reporte a intenção, aguarde confirmação e pergunte se a atividade vai para um subagente (decisão do usuário).
3. **ACT:** siga a skill da estação em `.agents/skills/sdd-*`, carregando só a referência do comando.
4. **WRITE:** apenas arquivos do escopo.
5. **CLOSE:** atualize o `status.md` da iniciativa e grave telemetria (`forge-sdd session record`, se `telemetry.enabled`).

## Onde está cada coisa
- `sdd/memory/constitution.md`: regras por ID (`ARQ`, `SEG`, `DUP`, `COD`, `TST`, `FLX`, `GOV`); leia só os IDs da tarefa.
- `sdd/discovery/<id>-<nome>/`: discovery, criteria, plan, status, queue, features e fixes.
- `.agents/rules/*.md`: como cumprir a Constitution, por tarefa.
- `.agents/skills/sdd-*`: concepcao, planejamento, build, governanca e operacao.

## Comandos
`/discovery`, `/split-features`, `/implementar [feature]`, `/revisar`, `/status`, `/nova-feature`, `/novo-fix`, `/archive`, `/doctor`, `/upgrade-sdd`, `/constitution`, `/c4-architecture`, `/tutorial`. Adaptadores em `.claude/commands/` e `.claude/agents/`.

## Regras
- Nunca commite em `main`; uma branch por iniciativa.
- Uma estação, uma responsabilidade: PM/Arquiteto concebe, Planejador planeja, Builder implementa, Governança audita só leitura.
- MCPs `context7` e `git`; estado real em `sdd/memory/mcps.md`.
