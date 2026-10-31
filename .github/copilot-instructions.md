# Copilot Instructions — forge-sdd

## Contexto
forge-sdd — stack principal: .

## Lifecycle (todo agente)
_Protocolo por comando. O pipeline por feature (Discovery → ... → PR) está em `sdd/FLOW.md` — fonte única da verdade._

1. **READ-MIN:** ler `sdd/memory/progress.md` (Builder e Analista de Governança também leem `sdd/memory/lessons.md` — padrões de erro já corrigidos, antes de implementar/revisar)
2. **PLAN:** reportar intenção, aguardar confirmação — e perguntar objetivamente se a atividade deve ser delegada a um subagente (varredura/pesquisa extensa tende a "sim"; edição pontual e pequena tende a "não"; decisão final é sempre do usuário). Em ferramentas sem esse conceito, item não aplicável.
3. **ACT:** executar no escopo do papel canônico em `.agents/roles/` (adaptado em `.github/chatmodes/`)
4. **WRITE:** editar apenas arquivos do escopo
5. **CLOSE** (Líder / Orquestrador): atualizar progress, métricas, archive se necessário

## Arquivos críticos
- `sdd/memory/progress.md` — estado ativo
- `sdd/memory/constitution.md` — regras imutáveis numeradas por IDs (`ARQ-*`, `SEG-*`, `DUP-*`, `COD-*`, `TST-*`, `FLX-*`, `GOV-*`)
- `sdd/features/feat-XX-*.md` (ou aninhados em subpastas) — tarefa atual
- `.agents/rules/*.md` — regras de domínio do projeto (design system, arquitetura, acessibilidade, ...). Antes de agir sobre código ou design, consulte os arquivos relevantes à tarefa corrente.
- `.agents/roles/*.md` — fonte única dos papéis das 5 estações (Líder, Concepção PM/Arquiteto, Planejador, Fila do Builder e Analista de Governança).

## Papéis Canônicos e 5 Estações (`.agents/roles/` → `.github/chatmodes/`)
- **Líder (Orquestrador):** `.github/chatmodes/orquestrador.chatmode.md` → `.agents/roles/lider.md`
- **Concepção (PM + Arquiteto):** `.github/chatmodes/specifier.chatmode.md` / `.github/chatmodes/c4-architecture.chatmode.md` → `.agents/roles/pm-architect.md`
- **Planejador (`Plan`):** `.github/chatmodes/planner.chatmode.md` → `.agents/roles/planner.md`
- **Fila do Builder:** `.github/chatmodes/builder.chatmode.md` → `.agents/roles/builder.md`
- **Analista de Governança (Camada B Read-Only):** `.github/chatmodes/governance.chatmode.md` / `.github/chatmodes/revisor.chatmode.md` → `.agents/roles/governance.md`
- **Archivist / Migrator:** `.github/chatmodes/archivist.chatmode.md` e `.github/chatmodes/migrator.chatmode.md`

## MCPs
- **context7** — obrigatório antes de lib externa, desde que `sdd/memory/mcps.md` o liste como `ativo`; se `indisponível`, use a documentação oficial da lib em vez de assumir resposta do MCP
- **git** — status antes de iniciar/encerrar feature

## Orçamentos
- `progress.md` ≤ 1 KB → exceder dispara `/archive`
- `chatmode` ≤ 500 tokens → detalhe vai para `skills/`

## Regras
- Ao quebrar features em tasks ou discovery em features, agrupar arquivos em subpastas sob `sdd/features/` (a pasta da feature gerada pelo discovery deve refletir o nome do discovery).
