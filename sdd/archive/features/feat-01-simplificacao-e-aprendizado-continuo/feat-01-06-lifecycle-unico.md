# Feature 01-06 — Lifecycle Único (Fonte Única da Verdade)

Consolida as três descrições de lifecycle hoje divergentes (`CLAUDE.md` 5 passos, `sdd/FLOW.md` 7 estágios, chatmode Orquestrador com 13 passos) em uma única fonte, tornando o fluxo estável para o usuário acompanhar. Maior superfície de revisão — priorizada por último, depois que a base de dados/lições (feat-01-01 a 04) já está estável para validar que nada quebrou.

## Critérios de Aceitação Executáveis

1. `sdd/FLOW.md` passa a ser a fonte única, com os 5 estágios já usados em `CLAUDE.md` (Problema → Proposta → Refinamento → Execução → Entrega) mapeados aos estágios reais do processo (Discovery → Especificação → Split/Refinamento → Build → Review/Handoff).
2. `CLAUDE.md` e os chatmodes Orquestrador (Claude, Gemini, Copilot) passam a citar o estágio atual, não reescrever a sequência completa.
3. Nenhum passo hoje implícito nas três descrições é perdido — cada um é mapeado explicitamente para o novo documento único antes de remover a duplicata.

## Status: done

`sdd/FLOW.md` (dogfood, é o próprio documento do projeto — sem template porque é específico deste repositório) passou a se declarar explicitamente como fonte única da verdade, com uma tabela de mapeamento entre a narrativa de 5 estágios (Problema → Proposta → Refinamento → Execução → Entrega) e seus 7 passos técnicos. `CLAUDE.md`/`GEMINI.md`/`copilot-instructions.md` (dogfood + templates) e os chatmodes Orquestrador (Gemini/Copilot, dogfood + templates) ganharam uma linha citando `sdd/FLOW.md` como fonte do pipeline por feature, deixando claro que o ciclo READ-MIN→CLOSE é um eixo diferente (protocolo por comando). Nenhum passo operacional foi removido — apenas adicionada a referência cruzada, evitando perda de instrução concreta em qualquer agente. Golden fixtures regeneradas.
