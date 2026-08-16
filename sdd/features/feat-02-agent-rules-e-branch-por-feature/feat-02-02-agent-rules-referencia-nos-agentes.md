# Feature 02-02 — Referência a `.agent/rules/` nos Três Agentes

Fecha o valor de feat-02-01: de nada adianta ter `.agent/rules/` se nenhum agente é instruído a consultar. Depende de feat-02-01.

## Critérios de Aceitação Executáveis

1. `CLAUDE.md.tmpl`, `GEMINI.md.tmpl` e o(s) chatmode(s)/instruções Copilot relevantes ganham uma instrução curta e equivalente (não copiada em conteúdo, só em intenção) do tipo: "antes de agir, consulte os arquivos relevantes em `.agent/rules/*.md`".
2. A leitura é sob demanda (mesmo mecanismo já usado hoje para `sdd/memory/progress.md`/`lessons.md`) — não exige carregar todos os arquivos de `.agent/rules/` sempre, só os pertinentes à tarefa corrente.
3. Nenhuma sintaxe proprietária de agente é usada na instrução — é leitura de Markdown puro, válida para qualquer modelo por trás dos três agentes.
4. Replicado nos três agentes (dogfood + templates) mantendo paridade de comportamento.

## Status: done

`CLAUDE.md.tmpl`, `GEMINI.md.tmpl`, `OPENAI.md.tmpl` e `.github/copilot-instructions.md.tmpl` (dogfood + templates) ganharam a mesma linha, na seção **Arquivos críticos**: referência a `.agent/rules/*.md` com instrução de consulta sob demanda antes de agir sobre código/design. Leitura de Markdown puro, sem sintaxe proprietária — mesmo padrão já usado para `sdd/memory/progress.md`/`lessons.md`. Golden fixtures regeneradas; dogfood deste repositório atualizado via `forge-sdd update`.
