# GEMINI.md — forge-sdd

## Contexto do projeto
forge-sdd — CLI Go para scaffolding de estruturas SDD (Software Design Doc) v1.6.1-beta.1.

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
- **Architect:** `.gemini/skills/c4-architecture.chatmode.md`

## Prompts (Comandos)
Para executar uma tarefa, peça pelo nome do comando ou utilize uma das frases de exemplo. Eu consultarei a instrução correspondente em `.gemini/prompts/`:

- **Status:** "rodar o status", "qual o progresso?" → `status.prompt.md`
- **Próxima Feature:** "iniciar próxima tarefa", "proxima-feature" → `proxima-feature.prompt.md`
- **Nova Feature:** "criar nova feature [descrição]", "nova-feature" → `nova-feature.prompt.md`
- **Revisar:** "validar feature", "rodar o revisar" → `revisar.prompt.md`
- **Discovery:** "discovery [ideia]", "fazer discovery de..." → `discovery.prompt.md`
- **Constitution:** "alinhar arquitetura", "rodar constitution" → `constitution.prompt.md`
- **C4 Architecture:** "gerar diagrama C4", "desenhar arquitetura" → `c4-architecture.prompt.md`
- **Doctor:** "check-up do projeto", "rodar doctor" → `doctor.prompt.md`
- **Archive:** "limpar progresso", "rodar archive" → `archive.prompt.md`
- **Upgrade:** "atualizar sdd para vX", "upgrade-sdd" → `upgrade-sdd.prompt.md`
