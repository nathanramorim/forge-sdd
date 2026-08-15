# Feature 01-03 — Artefato de Aprendizado (`sdd/memory/lessons.md`)

Introduz o mecanismo de os agentes aprenderem com os próprios fixes: hoje nenhum arquivo persiste "isso já quebrou antes" — cada sessão de fix começa do zero, mesmo quando a mesma causa raiz já gerou mais de uma correção (ex.: duplicação de lógica entre prompts). Depende de feat-01-01 para o mecanismo determinístico de escrita.

## Critérios de Aceitação Executáveis

1. Novo arquivo `sdd/memory/lessons.md`, orçamento ≤ 2 KB, formato de lista curta: `padrão → correção → referência da feature/fix`.
2. Atualizado de forma determinística (mesmo mecanismo de feat-01-01) ao final de `/revisar` ou `/novo-fix` quando o outcome é `approved` e a causa raiz é significativa/recorrente.
3. Não exige leitura de todo o histórico de fixes — apenas o resumo curto persistido.

## Status: todo
