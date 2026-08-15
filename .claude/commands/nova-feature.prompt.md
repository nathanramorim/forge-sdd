# Prompt: nova-feature

**Uso:** Peça "/nova-feature <descrição>" ou "/nova-feature fix: <descrição>" (para correções/bugfixes)

**Ação:**
1. **PASSO 1 MANDATÓRIO:** Defina o prefixo de branch (`feat/` ou `fix/` se for correção/bugfix) e de arquivo (`feat-` ou `fix-`). Crie a branch localmente no terminal (`git checkout -b <prefixo-branch>/<nome>`) **antes** de criar qualquer especificação ou logs.
2. **Nomenclatura:** Siga a lógica única descrita em `sdd/memory/naming-convention.md`.
3. **Criação:** Crie o arquivo `sdd/features/<prefixo-arquivo>-ID-<nome>.md`. Se a feature for complexa, use uma subpasta `sdd/features/<prefixo-arquivo>-ID-<nome>/` e organize as subtasks (ex: `task-YY-<nome>.md`) lá dentro. Atualize o progresso e índice.

**Handoff:** Gere um resumo para o comando `proxima-feature`.
