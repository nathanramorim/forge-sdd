# Copilot Instructions — forge-sdd

## Contexto
forge-sdd — stack principal: .

## Lifecycle (todo agente)
_Protocolo por comando. O pipeline por feature (Discovery → ... → PR) está em `sdd/FLOW.md` — fonte única da verdade._

1. **READ-MIN:** ler `sdd/memory/progress.md` (Builder e Revisor também leem `sdd/memory/lessons.md` — padrões de erro já corrigidos, antes de implementar/revisar)
2. **PLAN:** reportar intenção, aguardar confirmação
3. **ACT:** executar no escopo do papel
4. **WRITE:** editar apenas arquivos do escopo
5. **CLOSE** (Orquestrador): atualizar progress, métricas, archive se necessário

## Arquivos críticos
- `sdd/memory/progress.md` — estado ativo
- `sdd/memory/constitution.md` — regras imutáveis
- `sdd/features/feat-XX-*.md` — tarefa atual

## MCPs
- **context7** — obrigatório antes de lib externa, desde que `sdd/memory/mcps.md` o liste como `ativo`; se `indisponível`, use a documentação oficial da lib em vez de assumir resposta do MCP
- **git** — status antes de iniciar/encerrar feature

## Orçamentos
- `progress.md` ≤ 1 KB → exceder dispara `/archive`
- `chatmode` ≤ 500 tokens → detalhe vai para `skills/`
