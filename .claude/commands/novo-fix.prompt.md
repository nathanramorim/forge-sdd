# Prompt: novo-fix

**Uso:** Peça "/novo-fix <descrição do bug>"

**Ação:**
1. **PASSO 1 MANDATÓRIO:** Crie a branch do fix localmente no terminal (`git checkout -b fix/<nome>`) **antes** de criar qualquer especificação ou logs.
2. **Nomenclatura:** Leia a configuração `naming_convention` em `sdd/.sddrc`. Se for `workitem`, **pergunte obrigatoriamente ao usuário** qual é o ID do Workitem a usar. Se for `hash`, gere um hash de 4 dígitos hexadecimais únicos (ex: `5ae2`). Se for `sequencial` ou ausente, use o próximo número sequencial (XX).
3. **Criação:** Crie o arquivo `sdd/features/fix-ID-<nome>.md`. Se o fix for complexo, crie uma subpasta `sdd/features/fix-ID-<nome>/` e salve as subtasks lá dentro. Atualize o progresso e índice.
4. **Gravação de Métricas:** Se a telemetria estiver habilitada em `sdd/.sddrc`, execute `forge-sdd session record --feature "<arquivo fix criado>" --outcome blocked --criterio-atendido=false` (o fix ainda não foi implementado) — garante telemetria mesmo se a sessão terminar aqui.

**Handoff:** Gere um resumo para o comando `proxima-feature`.
