# Comando: doctor

**Uso:** `/doctor` faz o check-up do projeto e reporta verde ou vermelho, sem corrigir.

Rode `forge doctor` e reporte verde ou vermelho para cada item, sem corrigir nada:
1. Orçamentos: `status.md` da iniciativa ativa em até 1 KB.
2. Adaptadores dos comandos existentes para cada agente configurado e skills de estação em `.agents/skills/`.
3. `context7` e `git` configurados nas instruções do agente.
4. Iniciativa em andamento sem branch ativa no git.
5. `sdd/.metrics/schema.json` presente.
6. Nome do projeto no `.sddrc` ainda no padrão `meu-projeto`.
7. Inclua o resumo de telemetria (sessões approved, rejected e blocked).
8. `forge doctor --harness [--json]` diagnostica o harness do pipeline (versão do binário, skills, contrato skill x binário, pareceres gravados, worktrees e workers) e sai com 3 se houver falha. Só lê: não corrija nem apague por conta própria.
