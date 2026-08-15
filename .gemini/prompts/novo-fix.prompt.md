# Prompt: novo-fix

**Uso:** Peça "/novo-fix <descrição do bug>"

**Ação:**
1. Acione a lógica de **Specifier**.
2. **PASSO 1 IMPERATIVO:** Crie a branch do fix localmente (`git checkout -b fix/<nome>`) no terminal **antes** de criar qualquer arquivo de especificação ou atualizar logs.
3. **Identificação do ID de Nomenclatura:**
   - Leia a configuração de `naming_convention` em `sdd/.sddrc`.
   - Se `naming_convention` for `workitem`, **pergunte obrigatoriamente ao usuário** qual é o ID do Workitem a usar antes de prosseguir (ex: prefixo `1024` gerando `fix-1024-nome.md`).
   - Se `naming_convention` for `hash`, gere um hash de 4 dígitos hexadecimais únicos (ex: `5ae2`).
   - Se `naming_convention` for `sequencial` ou ausente, use o próximo número sequencial (XX) incremental com base no histórico de arquivos.
4. **Criação e Agrupamento:** Se a correção for complexa e requerer quebra em subtasks, crie uma subpasta com o nome do fix (ex: `sdd/features/fix-ID-<nome>/`) e armazene as tasks lá dentro (ex: `sdd/features/fix-ID-<nome>/task-YY-<nome>.md`). Caso contrário, crie o arquivo diretamente na raiz sob `sdd/features/fix-ID-<nome>.md`.
5. **Atualização de Logs:** Atualize o índice de features e o progresso.
6. **Gravação de Métricas:** Se a telemetria estiver habilitada em `sdd/.sddrc`, execute `forge-sdd session record --feature "<arquivo fix criado>" --outcome blocked --criterio-atendido=false` (o fix ainda não foi implementado) — garante telemetria mesmo se a sessão terminar aqui.

**Handoff Final:**
Gere um Handoff para o comando `/proxima-feature`, indicando o nome da branch `fix/*` e as tarefas prioritárias.
