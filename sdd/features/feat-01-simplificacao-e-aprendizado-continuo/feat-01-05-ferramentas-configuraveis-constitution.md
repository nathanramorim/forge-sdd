# Feature 01-05 — Ferramentas e VCS Configuráveis na Constituição

Resolve a suposição incondicional de que `context7`/`git` MCP e `gh`/GitHub estão sempre disponíveis. Hoje `sdd/memory/mcps.md` é uma tabela estática nunca lida, e `naming_convention: workitem` (estilo Azure DevOps) nunca influencia como o PR é aberto. Independente da cadeia de telemetria (feat-01-01 a 04) — pode rodar em paralelo.

## Critérios de Aceitação Executáveis

1. `sdd/memory/constitution.md` (e seu `.tmpl`) ganha uma seção declarando, por projeto: (a) status real de cada MCP configurado (`ativo`/`indisponível`, populado a partir de `mcps.md`, que passa a ser efetivamente lido); (b) o VCS/work-item system em uso (`github`/`azure-devops`/`nenhum`).
2. `constitution.prompt.md` (nos 3 agentes) pergunta isso ao gerar/atualizar a constitution, no mesmo padrão já usado para `naming_convention` e nível de linguagem.
3. Prompts que hoje assumem `gh pr create --fill` incondicionalmente passam a checar esse campo antes de executar: `github` mantém comportamento atual; `azure-devops` usa instrução equivalente documentada; `nenhum` deixa a branch pronta e informa o usuário, sem tentar comando de VCS algum.
4. Prompts que hoje assumem `context7`/`git` MCP incondicionalmente (Regra 5 da Constituição, `CLAUDE.md`, chatmode Builder) passam a checar o status declarado antes de instruir o uso; se `indisponível`, usa alternativa (documentação já conhecida) em vez de assumir resposta do MCP.
5. Default do campo de VCS é `github` — projetos existentes que não passarem por `/constitution` novamente não têm regressão de comportamento.

## Status: todo
