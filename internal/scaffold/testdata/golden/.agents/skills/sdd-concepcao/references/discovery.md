# Comando: discovery

**Uso:** `/discovery <descrição da demanda>`.

1. **Clarify:** aplique `sdd/memory/clarify.md`.
2. **ID:** rode `forge-sdd next-id` e crie a branch `feat/<id>-<nome>` (pergunte a branch de partida, padrão `main`; se já existir branch da iniciativa, pergunte se retoma).
3. **ADRs e rules:** antes de propor qualquer coisa, leia as ADRs e as rules aplicáveis ao tema; ADR depreciada ou substituída não vale como base.
4. **Produto (`discovery.md`):** problema, para quem, jornada, regras de negócio, escopo (dentro e fora) e riscos.
5. **Arquitetura (`criteria.md`):** restrições citando IDs da Constitution; seções `Modelo do módulo`, `O que já existe` (tabela `Procurei | Onde | Achei | Decisão | Justificativa`) e `Dependências novas`; diagramas C4 em Mermaid (contexto, contêineres e sequência) e critérios de aceitação executáveis.
6. **Plano (`plan.md`):** quebra ordenada de baixo para cima (limite de 7 tasks por onda), dependências, paralelismo e rules/IDs por task.
7. **Estado:** crie `status.md` (até 1 KB) com `Status`, `Próximo passo` e `A confirmar`.
8. Se `Nível de Linguagem` na Constitution for `iniciante`, explique C4 e critério de aceitação com exemplos simples, sem alterar os critérios.
9. **Telemetria:** `forge-sdd session record --feature "sdd/discovery/<id>-<nome>/discovery.md" --outcome approved --criterio-atendido=true`.
10. **Gate 1 (`FLX-01`):** submeta ao humano; após aprovar, sugira `/split-features`.
