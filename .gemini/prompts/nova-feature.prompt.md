# Prompt: nova-feature

**Uso:** Peça "/nova-feature <descrição>" ou "/nova-feature fix: <descrição>" (para correções/bugfixes)

**Ação:**
1. Acione a lógica de **Specifier**.
2. **Definição de Prefixo e Branch:**
   - Se for uma correção/bugfix (descrição contendo "fix" ou "correção"), use o prefixo de arquivo `fix-` (ex: `fix-ID-<nome>.md`) e o prefixo de branch `fix/` (ex: `fix/<nome>`).
   - Caso contrário, use o prefixo de arquivo `feat-` (ex: `feat-ID-<nome>.md`) e o prefixo de branch `feat/` (ex: `feat/<nome>`).
3. **PASSO 1 IMPERATIVO:** Crie a branch correspondente localmente (`git checkout -b <prefixo-branch>/<nome>`) no terminal **antes** de criar qualquer arquivo de especificação ou atualizar logs.
4. **Identificação do ID de Nomenclatura:**
   - Leia a configuração de `naming_convention` em `sdd/.sddrc`.
   - Se `naming_convention` for `workitem`, **pergunte obrigatoriamente ao usuário** qual é o ID do Workitem a usar antes de prosseguir (ex: prefixo `1024` gerando `feat-1024-nome.md`).
   - Se `naming_convention` for `hash`, gere um hash de 4 dígitos hexadecimais únicos (ex: `5ae2`).
   - Se `naming_convention` for `sequencial` ou ausente, use o próximo número sequencial (XX) incremental com base no histórico de arquivos.
5. **Criação e Agrupamento:** Se a demanda for complexa e requerer quebra em subtasks, crie uma subpasta com o nome da feature (ex: `sdd/features/<prefixo-arquivo>-ID-<nome>/`) e armazene as tasks lá dentro (ex: `sdd/features/<prefixo-arquivo>-ID-<nome>/task-YY-<nome>.md`). Caso contrário, crie o arquivo diretamente na raiz sob `sdd/features/<prefixo-arquivo>-ID-<nome>.md`.
6. **Atualização de Logs:** Atualize o índice de features e o progresso.

**Handoff Final:**
Gere um Handoff para o comando `/proxima-feature`, indicando o nome da branch e as tarefas prioritárias.
