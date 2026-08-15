# Feature 01-07 — Redução de Duplicação (Lógica de Nomenclatura)

Elimina a lógica de nomenclatura (`sequencial`/`hash`/`workitem`) hoje copiada literalmente em `discovery.prompt.md`, `nova-feature.prompt.md` e `novo-fix.prompt.md` — quando o comportamento muda, hoje é preciso lembrar de editar em múltiplos lugares (causa raiz confirmada de pelo menos 2 fixes anteriores).

## Critérios de Aceitação Executáveis

1. A lógica de nomenclatura passa a viver em um único bloco referenciado pelos três prompts (ex.: incluído via template parcial, ou citado como referência a uma única regra da Constituição), sem duplicar o texto.
2. Nenhuma capacidade removida: os três comandos continuam suportando `sequencial`, `hash` e `workitem` exatamente como hoje.
3. Replicado de forma consistente nos três agentes (Claude, Gemini, Copilot).

## Status: done

Criado `sdd/memory/naming-convention.md` (dogfood + template) como fonte única da lógica sequencial/hash/workitem. Os 14 pontos onde essa lógica estava duplicada literalmente (`discovery`, `nova-feature`, `novo-fix` — Claude e Gemini completos, mais `novo-fix` do Copilot, cada um em dogfood + template) foram substituídos por uma referência de uma linha ao arquivo canônico. Nenhuma capacidade removida: os três valores (`sequencial`/`hash`/`workitem`) continuam suportados exatamente como antes. Nota: `discovery`/`nova-feature` do Copilot já não tinham essa lógica implementada (gap pré-existente, fora do escopo desta feature — não é duplicação, é ausência). Golden fixtures regeneradas.
