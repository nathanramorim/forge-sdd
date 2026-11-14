# Comando: discovery

**Uso:** `/discovery <descrição da demanda>`.

1. **Brief primeiro:** rode `forge-sdd brief` no início; se o estado for `rascunho`, proponha `/constitution` antes de qualquer spec.
2. **Clarify:** aplique `sdd/memory/clarify.md`.
3. **ID:** rode `forge-sdd next-id` e crie a branch `feat/<id>-<nome>` (pergunte a branch de partida, padrão `main`; se já existir branch da iniciativa, pergunte se retoma).
4. **ADRs e rules:** antes de propor qualquer coisa, leia as ADRs e as rules aplicáveis ao tema; cite as decisões aplicadas na spec; ADR depreciada ou substituída não vale como base.
5. **Produto (`discovery.md`):** problema, para quem, jornada, regras de negócio, escopo (dentro e fora) e riscos.
6. **Arquitetura (`criteria.md`):** restrições citando IDs da Constitution; seções `Modelo do módulo`, `Decisões do projeto aplicadas`, `O que já existe` (tabela `Procurei | Onde | Achei | Decisão | Justificativa`) e `Dependências novas`; diagramas C4 em Mermaid (contexto, contêineres e sequência) e critérios de aceitação executáveis.
7. **Plano (`plan.md`):** quebra ordenada de baixo para cima (limite de 7 tasks por onda), dependências, paralelismo e rules/IDs por task.
8. **Estado:** crie `status.md` (até 1 KB) com `Status`, `Próximo passo` e `A confirmar`.
9. Se `Nível de Linguagem` na Constitution for `iniciante`, explique C4 e critério de aceitação com exemplos simples, sem alterar os critérios.
10. **Telemetria:** `forge-sdd session record --feature "sdd/discovery/<id>-<nome>/discovery.md" --outcome approved --criterio-atendido=true`.
11. **Gate 1 (`FLX-01`):** submeta ao humano; após aprovar, sugira `/split-features`. Em modo pipeline (existe `.forge/config/forge.json`) esta sessão nunca constrói: o `/split-features` termina entregando ao Builder.
