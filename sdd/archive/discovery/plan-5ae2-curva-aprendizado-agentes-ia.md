# Plano de Discovery 5ae2 — Reduzindo a Curva de Aprendizado do Forge-SDD

Este plano descreve o roadmap preliminar de features sugeridas, priorizadas por maior impacto na curva de aprendizado com menor risco arquitetural primeiro.

## Tarefas de Homologação Sugeridas

- [ ] **Tarefa 1 (rápido ganho):** Implementar cheat-sheet de comandos SDD impresso ao final de `forge-sdd init`, gerado a partir dos `.prompt.md.tmpl` já embutidos.
- [ ] **Tarefa 2:** Evoluir o prompt `/status` (nos três agentes) para sempre concluir com "Próximo comando sugerido", calculado a partir de `sdd/memory/progress.md`.
- [ ] **Tarefa 3:** Auditoria de deriva de convenção no `forge-sdd doctor` — detectar coexistência de nomenclatura sequencial (`feat-NN`) e hash (`feat-xxxx`) no mesmo projeto e reportar. Usar o próprio repositório do Forge-SDD (PR #26 aberta vs. `feat-43a2`/`discovery-9b2f` já em uso) como caso de teste real.
- [ ] **Tarefa 4:** Criar o prompt `/tutorial` e a flag `forge-sdd init --tutorial`, com discovery/feature de exemplo isolados em `sdd/features/_tutorial/`.
- [ ] **Tarefa 5:** Implementar gate de graduação para `.sdd-auto-pilot` (mínimo N ciclos `done` registrados em telemetria antes de permitir o flag, com bypass explícito documentado).
- [ ] **Tarefa 6 (spike técnico):** Investigar viabilidade de Orquestrador/Builder/Revisor como subagentes nativos por agente (Claude/Gemini/Copilot), produzindo tabela de suporte e recomendação — sem obrigação de shippar a migração completa nesta rodada.
- [ ] **Tarefa 7 (opcional/paralela):** Adicionar opção de "modo iniciante" (linguagem simplificada) na pergunta de idioma já existente no `/constitution`.

## Observação de Priorização

Tarefas 1-3 resolvem lacunas já comprovadas neste próprio repositório (fratura CLI/chat, falta de próximo-passo, deriva de nomenclatura real) e têm baixo risco — devem vir primeiro. Tarefas 4-5 têm maior valor pedagógico mas exigem mais desenho de UX de prompt. Tarefa 6 é exploratória e não bloqueia as demais.
