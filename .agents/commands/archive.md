# Comando: archive

**Uso:** Peça "/archive" ou "limpar progresso"

**Ação:**
1. Limpe o histórico antigo de `progress.md` movendo para `progress-log.md` (mantendo `progress.md` ≤ 1 KB).
2. **Revisão de Fim de Onda (`GOV-03` e `DUP-05`):** Inspecione as tasks arquivadas e `sdd/memory/lessons.md`:
   - Se uma atividade padronizada se repetiu nas entregas concluídas e ainda não é uma skill em `.agents/skills/`, proponha-a com `forge-sdd skills propose --name <slug> --description "<desc>" --steps "<passos>"` e informe ao Líder/humano para aprovação (`forge-sdd skills approve <slug>`).
3. **Gravação de Métricas (determinística):** Se a telemetria estiver habilitada em `sdd/.sddrc`, execute `forge-sdd session record --feature "sdd/memory/progress.md" --outcome approved --criterio-atendido=true`.
