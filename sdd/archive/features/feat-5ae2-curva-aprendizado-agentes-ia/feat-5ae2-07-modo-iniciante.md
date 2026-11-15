# Feature 5ae2-07 — Modo Iniciante (Linguagem Simplificada)

Aproximação do público não-técnico (PMs, founders "vibecoding") que hoje esbarra em jargão como "C4 Model" e "critério executável" logo no primeiro uso.

## Critérios de Aceitação Executáveis

1. A pergunta de idioma já existente no `/constitution` (introduzida em `3cbe5ef`) deve ganhar uma segunda pergunta opcional: nível de linguagem (`padrão` ou `iniciante`).
2. Quando `iniciante` for escolhido, os templates de prompt gerados devem usar linguagem simplificada e exemplos concretos no lugar de jargão técnico, sem alterar os critérios de aceitação executáveis reais (que continuam técnicos e testáveis).
3. A escolha deve ser persistida em `sdd/memory/constitution.md` para reaproveitamento em `/upgrade-sdd`.

## Status: done

Adicionado o passo `0b` (nível de linguagem: `padrão`/`iniciante`) aos três prompts `/constitution`, espelhando o padrão já existente da pergunta de idioma (passo `0`), persistido na mesma seção Regras/Regras de Ouro de `constitution.md`. Para o comando com maior densidade de jargão (`/discovery`, citado na Lacuna #7 da Discovery), adicionada instrução explícita para verificar `Nível de Linguagem` e simplificar termos como "C4 Model" e "critério de aceitação" com exemplos concretos, sem alterar os critérios de aceitação em si. Demais comandos seguem a convenção geral estabelecida no `/constitution` (respeitar a escolha em toda invocação futura), sem duplicar templates completos por nível — mesmo padrão de "escolha persistida, comportamento seguido" já usado para o idioma. Golden fixtures do Copilot regeneradas.
