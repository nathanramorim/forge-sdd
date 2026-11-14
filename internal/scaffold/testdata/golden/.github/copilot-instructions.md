# Copilot Instructions — demo

demo: stack go. Fluxo por feature em `sdd/FLOW.md`.

## Ciclo por comando
1. **READ-MIN:** Discovery começa por `forge-sdd brief`; `forge-sdd status`; Builder e Governança leem também `sdd/memory/lessons.md`.
2. **PLAN:** reporte a intenção, aguarde confirmação e pergunte se vai para um subagente.
3. **ACT:** siga a skill da estação em `.agents/skills/sdd-*`, só a referência do comando.
4. **WRITE:** apenas arquivos do escopo.
5. **CLOSE:** atualize o `status.md` da iniciativa e grave telemetria (`forge-sdd session record`).

## Onde está cada coisa
- `sdd/memory/constitution.md`: regras por ID; leia só os IDs da tarefa (`forge-sdd context`).
- `sdd/discovery/<id>-<nome>/`: discovery, criteria, plan, status, queue, features e fixes.
- `.agents/rules/`: como cumprir a Constitution. `.agents/skills/sdd-*`: as 5 estações.

## Comandos
`/discovery`, `/split-features`, `/implementar [feature]`, `/revisar`, `/status`, `/nova-feature`, `/novo-fix`, `/archive`, `/doctor`, `/upgrade-sdd`, `/constitution`, `/c4-architecture`, `/tutorial` (`.github/prompts/`, `.github/chatmodes/`).

## Regras
- Nunca commite em `main`; uma branch por iniciativa.
- Uma estação, uma responsabilidade; a Governança só lê.
- Respostas de planejamento até 500 tokens; o detalhe vai para as skills.
- MCPs `context7` e `git` (estado em `sdd/memory/mcps.md`).
