# Feature 5ae2-07 — Modo Iniciante (Linguagem Simplificada)

Aproximação do público não-técnico (PMs, founders "vibecoding") que hoje esbarra em jargão como "C4 Model" e "critério executável" logo no primeiro uso.

## Critérios de Aceitação Executáveis

1. A pergunta de idioma já existente no `/constitution` (introduzida em `3cbe5ef`) deve ganhar uma segunda pergunta opcional: nível de linguagem (`padrão` ou `iniciante`).
2. Quando `iniciante` for escolhido, os templates de prompt gerados devem usar linguagem simplificada e exemplos concretos no lugar de jargão técnico, sem alterar os critérios de aceitação executáveis reais (que continuam técnicos e testáveis).
3. A escolha deve ser persistida em `sdd/memory/constitution.md` para reaproveitamento em `/upgrade-sdd`.
