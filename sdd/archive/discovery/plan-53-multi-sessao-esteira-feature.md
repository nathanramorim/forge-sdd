# Plano de Discovery 53 — Esteira de Execução de Feature em 3 Sessões (Spec → Act → Revisor)

Este plano descreve o roadmap preliminar de features sugeridas, priorizado por menor risco/dependência de mecanismo de sessão primeiro (handoff estruturado, que vale mesmo na sessão única de hoje), deixando a automação de sessão real e a decisão de paridade entre agentes por último.

## Tarefas de Homologação Sugeridas

- [ ] **Tarefa 1 (pré-requisito, sem dependência de mecanismo de sessão):** Handoff estruturado Act → Revisor — formato mínimo determinístico (arquivos alterados, comando(s) de teste + resultado, pontos de atenção) anexado à seção Handoff do arquivo `feat-XX.md`/`fix-XX.md`. Vale hoje mesmo em sessão única — melhora a qualidade da revisão independente de qualquer isolamento real de sessão.
- [ ] **Tarefa 2:** Documentar explicitamente as 3 estações (Spec/Act/Revisor) como unidades de sessão no `sdd/FLOW.md`, especializando a linha genérica "Sessão → Sessão" do Protocolo de Handoff em 3 handoffs nomeados — sem exigir nenhuma mudança de mecanismo, só de documentação/nomenclatura.
- [ ] **Tarefa 3:** Telemetria correlacionada por `feature` — validar/ajustar `sdd/.metrics/schema.json` e o agregador em `/status`/`/doctor` para reconstruir a esteira completa de uma feature a partir de múltiplos `session-*.json` com o mesmo valor de `feature`.
- [ ] **Tarefa 4 (piloto, decisão de produto primeiro):** Piloto de subagentes in-session só para Claude (`.claude/agents/*.md` para Builder/Revisor) — depende de decisão explícita do usuário sobre aceitar a assimetria entre agentes (per riscos do `discovery-53` e do spike `feat-5ae2-06`).
- [ ] **Tarefa 5 (piloto, opcional, maior escopo):** Piloto de sessões Claude Code Remote separadas (`create_session`/`send_message`) para uma esteira com estações realmente isoladas em sessões top-level distintas — inclui decidir quem orquestra a criação da próxima estação e o que fazer se uma travar.
- [ ] **Tarefa 6:** Validar explicitamente (teste de aceitação) que qualquer piloto de sessão isolada (Tarefa 4 ou 5) preserva a Regra 15 da Constituição — todas as estações operam na mesma branch, nunca uma branch por estação.
- [ ] **Tarefa 7 (documentação):** Declarar explicitamente no `README.md`/`CLAUDE.md`/`GEMINI.md`/docs de Copilot que Gemini e Copilot seguem sem isolamento automatizável de sessão nesta rodada — sem fingir paridade que não existe.

## Observação de Priorização

Tarefas 1-3 não dependem de nenhum mecanismo de sessão isolada e têm valor mesmo se o usuário decidir não avançar com pilotos de sessão real — priorizadas primeiro porque reduzem risco (handoff melhor documentado, telemetria correlacionável) sem exigir nenhuma decisão de produto sensível. As Tarefas 4 e 5 são pilotos que dependem de uma decisão explícita do usuário sobre aceitar assimetria entre agentes (mesmo risco já levantado pelo spike `feat-5ae2-06`, nunca resolvido) — não devem ser iniciadas sem essa confirmação. A Tarefa 6 é um teste de aceitação transversal a qualquer piloto, não uma feature isolada. A Tarefa 7 fecha o ciclo de transparência independente de quais pilotos avançarem.

## Métrica de Sucesso

- Toda feature revisada via `/revisar` passa a ter, no arquivo `feat-XX.md`/`fix-XX.md`, uma seção de handoff estruturada (arquivos alterados + testes + pontos de atenção) — hoje: inconsistente/informal.
- `/status`/`/doctor` conseguem listar, para uma feature em andamento, quais das 3 estações (Spec/Act/Revisor) já produziram telemetria e com qual `outcome` — hoje: impossível sem ler múltiplos arquivos manualmente.
- Se um piloto de sessão isolada (Tarefa 4 ou 5) avançar: zero regressão de comportamento quando a esteira roda em sessão única (compatibilidade retroativa), e zero violação da Regra 15 (branch única por feature) nos testes de aceitação.
