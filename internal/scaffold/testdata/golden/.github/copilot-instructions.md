# Copilot Instructions — demo

## Contexto
demo — stack principal: go.

## Lifecycle (todo agente)
1. **READ-MIN:** ler `sdd/memory/progress.md`
2. **PLAN:** reportar intenção, aguardar confirmação
3. **ACT:** executar no escopo do papel
4. **WRITE:** editar apenas arquivos do escopo
5. **CLOSE** (Orquestrador): atualizar progress, métricas, archive se necessário

## Arquivos críticos
- `sdd/memory/progress.md` — estado ativo
- `sdd/memory/constitution.md` — regras imutáveis
- `sdd/features/feat-XX-*.md` — tarefa atual

## MCPs
- **context7** — obrigatório antes de lib externa
- **git** — status antes de iniciar/encerrar feature

## Orçamentos
- `progress.md` ≤ 1 KB → exceder dispara `/archive`
- `chatmode` ≤ 500 tokens → detalhe vai para `skills/`
