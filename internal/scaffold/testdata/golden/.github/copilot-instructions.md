# Copilot Instructions — demo

## Contexto
demo — stack principal: go.

## Lifecycle (todo agente)
1. **READ-MIN:** ler `sdd/memory/progress.md` (Builder e Revisor também leem `sdd/memory/lessons.md` — padrões de erro já corrigidos, antes de implementar/revisar)
2. **PLAN:** reportar intenção, aguardar confirmação
3. **ACT:** executar no escopo do papel
4. **WRITE:** editar apenas arquivos do escopo
5. **CLOSE** (Orquestrador): atualizar progress, métricas, archive se necessário

## Arquivos críticos
- `sdd/memory/progress.md` — estado ativo
- `sdd/memory/constitution.md` — regras imutáveis
- `sdd/features/feat-XX-*.md` (ou aninhados em subpastas) — tarefa atual

## MCPs
- **context7** — obrigatório antes de lib externa, desde que `sdd/memory/mcps.md` o liste como `ativo`; se `indisponível`, use a documentação oficial da lib em vez de assumir resposta do MCP
- **git** — status antes de iniciar/encerrar feature

## Orçamentos
- `progress.md` ≤ 1 KB → exceder dispara `/archive`
- `chatmode` ≤ 500 tokens → detalhe vai para `skills/`

## Regras
- Ao quebrar features em tasks ou discovery em features, agrupar arquivos em subpastas sob `sdd/features/` (a pasta da feature gerada pelo discovery deve refletir o nome do discovery).
