# Plano de Discovery 01 — Simplificação e Aprendizado Contínuo do Forge-SDD

Este plano descreve o roadmap preliminar de features sugeridas, priorizado por dependência de dados confiáveis primeiro, depois por impacto direto nos problemas relatados, deixando a consolidação de maior superfície por último.

## Tarefas de Homologação Sugeridas

- [ ] **Tarefa 1 (pré-requisito):** Tornar a gravação de telemetria code-enforced — mecanismo determinístico no binário Go (ou hook), disparado em `/proxima-feature`, `/revisar` e `/novo-fix`, não apenas no último passo de um prompt longo. Resolve diretamente o sintoma relatado: telemetria habilitada mas ausente de forma inconsistente entre projetos e agentes.
- [ ] **Tarefa 2:** Agregador mínimo de telemetria em `/status`/`/doctor`, consumindo os dados agora confiáveis (contagem por `outcome`, comandos mais usados).
- [ ] **Tarefa 3:** Introduzir `sdd/memory/lessons.md` e a escrita determinística ao final de `/revisar`/`/novo-fix` quando um fix é aprovado (usa o mesmo mecanismo da Tarefa 1).
- [ ] **Tarefa 4:** Consultar `lessons.md` no READ-MIN de Builder e Revisor.
- [ ] **Tarefa 5:** Ferramentas configuráveis na Constituição — campo de status real de MCPs (lendo `mcps.md`) e de VCS/work-item system (`github`/`azure-devops`/`nenhum`), perguntado em `/constitution`. Prompts que assumem `gh pr create` ou `context7`/`git` MCP passam a checar esse campo antes de agir, com fallback declarado.
- [ ] **Tarefa 6:** Unificar o lifecycle em uma fonte única (`sdd/FLOW.md`), com `CLAUDE.md` e os chatmodes Orquestrador citando o estágio em vez de reescrever a sequência completa.
- [ ] **Tarefa 7:** Extrair a lógica de nomenclatura duplicada (`discovery`, `nova-feature`, `novo-fix`) para um único bloco referenciado.
- [ ] **Tarefa 8 (recomendação, não execução automática):** Auditoria de comandos sobrepostos entre os 12 existentes, entregando candidatos a fusão/aposentadoria para decisão do usuário numa rodada futura.

## Observação de Priorização

Tarefas 1-4 formam a cadeia "telemetria confiável → dado agregado → lição persistida → lição consultada" — sem a Tarefa 1, as Tarefas 2-4 constroem sobre uma fonte de dados que falha silenciosamente, repetindo o padrão já visto em 4 correções pontuais anteriores. A Tarefa 5 resolve o segundo problema relatado (MCPs/VCS hardcoded) e é independente da cadeia de telemetria — pode rodar em paralelo às Tarefas 2-4. As Tarefas 6-7 são consolidação de superfície (o pedido original de "simplificar sem perder qualidade") e têm maior risco de regressão por tocarem múltiplos arquivos ao mesmo tempo — ficam por último, quando a base de dados/lições já está estável para validar que nada quebrou. A Tarefa 8 é só recomendação: não remove nenhum comando nesta rodada, apenas produz a lista para decisão explícita do usuário.

## Métrica de Sucesso

- Número de projetos onde telemetria falha silenciosamente cai a zero (toda sessão, mesmo interrompida em `/revisar`, produz um `session-*.json`).
- Número de comandos/conceitos que um novo usuário precisa aprender no onboarding (hoje 17) não aumenta, com tendência de queda após a Tarefa 8.
- `mcps.md` e o campo de VCS na Constituição passam a ser efetivamente lidos por pelo menos um prompt cada (hoje: zero).
