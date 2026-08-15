---
description: "Cria uma fix para forge-sdd a partir da descrição de um bug"
agent: agent
---

Invoque o Specifier para criar um fix-ID a partir desta descrição de bug: ${input:descrição do bug}

1. **PASSO 1 MANDATÓRIO:** Crie a branch do fix localmente no terminal (`git checkout -b fix/<nome>`) **antes** de criar qualquer especificação ou logs.
2. **Nomenclatura:** Siga a lógica única descrita em `sdd/memory/naming-convention.md`.
3. **Criação:** Crie o arquivo `sdd/features/fix-ID-<nome>.md`. Se o fix for complexo, crie uma subpasta `sdd/features/fix-ID-<nome>/` e salve as subtasks lá dentro. Atualize o progresso e índice.
4. **Gravação de Métricas:** Se a telemetria estiver habilitada em `sdd/.sddrc`, execute `forge-sdd session record --feature "<arquivo fix criado>" --outcome blocked --criterio-atendido=false` (o fix ainda não foi implementado) — garante telemetria mesmo se a sessão terminar aqui.

**Handoff:**
Ao finalizar, gere um resumo para o comando `/proxima-feature`.
