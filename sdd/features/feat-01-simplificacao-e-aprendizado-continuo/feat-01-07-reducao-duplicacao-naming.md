# Feature 01-07 — Redução de Duplicação (Lógica de Nomenclatura)

Elimina a lógica de nomenclatura (`sequencial`/`hash`/`workitem`) hoje copiada literalmente em `discovery.prompt.md`, `nova-feature.prompt.md` e `novo-fix.prompt.md` — quando o comportamento muda, hoje é preciso lembrar de editar em múltiplos lugares (causa raiz confirmada de pelo menos 2 fixes anteriores).

## Critérios de Aceitação Executáveis

1. A lógica de nomenclatura passa a viver em um único bloco referenciado pelos três prompts (ex.: incluído via template parcial, ou citado como referência a uma única regra da Constituição), sem duplicar o texto.
2. Nenhuma capacidade removida: os três comandos continuam suportando `sequencial`, `hash` e `workitem` exatamente como hoje.
3. Replicado de forma consistente nos três agentes (Claude, Gemini, Copilot).

## Status: todo
