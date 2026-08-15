# Feature 01-06 — Lifecycle Único (Fonte Única da Verdade)

Consolida as três descrições de lifecycle hoje divergentes (`CLAUDE.md` 5 passos, `sdd/FLOW.md` 7 estágios, chatmode Orquestrador com 13 passos) em uma única fonte, tornando o fluxo estável para o usuário acompanhar. Maior superfície de revisão — priorizada por último, depois que a base de dados/lições (feat-01-01 a 04) já está estável para validar que nada quebrou.

## Critérios de Aceitação Executáveis

1. `sdd/FLOW.md` passa a ser a fonte única, com os 5 estágios já usados em `CLAUDE.md` (Problema → Proposta → Refinamento → Execução → Entrega) mapeados aos estágios reais do processo (Discovery → Especificação → Split/Refinamento → Build → Review/Handoff).
2. `CLAUDE.md` e os chatmodes Orquestrador (Claude, Gemini, Copilot) passam a citar o estágio atual, não reescrever a sequência completa.
3. Nenhum passo hoje implícito nas três descrições é perdido — cada um é mapeado explicitamente para o novo documento único antes de remover a duplicata.

## Status: todo
