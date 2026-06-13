# GEMINI.md — forge-sdd

## Contexto do projeto
forge-sdd — CLI Go para scaffolding de estruturas SDD (Software Design Doc) v1.3.2.

## Lifecycle (todo agente)
1. **READ-MIN:** ler `sdd/memory/progress.md`
2. **PLAN:** reportar intenção, aguardar confirmação
3. **ACT:** executar no escopo do papel (use as skills em `.gemini/skills/`)
4. **WRITE:** editar apenas arquivos do escopo
5. **CLOSE** (Orquestrador): atualizar progress, métricas, archive se necessário

## Arquivos críticos
- `sdd/memory/progress.md` — estado ativo (leia primeiro a cada sessão)
- `sdd/memory/constitution.md` — regras imutáveis
- `sdd/features/feat-XX-*.md` — tarefa atual

## Orçamentos
- `progress.md` ≤ 1 KB
- Resposta de planejamento ≤ 500 tokens — detalhe vai para `sdd/skills/`

## Regras
- Nunca commite direto em `main`
- Consulte `sdd/memory/constitution.md` antes de decisões arquiteturais
- Use ferramentas de leitura antes de editar qualquer arquivo
- Siga rigorosamente a responsabilidade única por artefato (Specifier escreve specs, Builder escreve código, etc.)

## Skills & Papéis
Este projeto define papéis específicos em `.gemini/skills/`. Carregue a instrução correspondente ao iniciar uma fase:
- **Orquestrador:** `.gemini/skills/orquestrador.chatmode.md`
- **Builder:** `.gemini/skills/builder.chatmode.md`
- **Revisor:** `.gemini/skills/revisor.chatmode.md`
- **Archivist:** `.gemini/skills/archivist.chatmode.md`
- **Specifier:** `.gemini/skills/specifier.chatmode.md`
- **Migrator:** `.gemini/skills/migrator.chatmode.md`

## Prompts (Comandos)
Se o usuário solicitar um comando (ex: `/status`), consulte a instrução em `.gemini/prompts/`:
- `/status` -> `.gemini/prompts/status.prompt.md`
- `/proxima-feature` -> `.gemini/prompts/proxima-feature.prompt.md`
- `/nova-feature` -> `.gemini/prompts/nova-feature.prompt.md`
- `/revisar` -> `.gemini/prompts/revisar.prompt.md`
- `/archive` -> `.gemini/prompts/archive.prompt.md`
- `/doctor` -> `.gemini/prompts/doctor.prompt.md`
- `/upgrade-sdd` -> `.gemini/prompts/upgrade-sdd.prompt.md`
