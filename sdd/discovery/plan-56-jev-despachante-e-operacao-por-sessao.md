# Plano de Discovery 56 — Jev como Despachante/Decisor e Operação por Sessão

Roadmap preliminar, do menor risco (sem Jev) ao maior (autonomia). Tarefas 1-3 têm valor mesmo que o Jev não seja adotado.

## Tarefas de Homologação Sugeridas

- [ ] **Tarefa 1 (pré-requisito, sem Jev):** Ledger de execução por feature (`sdd/.runs/`) com estados por estação, lease e heartbeat, e comando `forge-sdd run status` — responde "ocupado? entregou? concluído?" de forma determinística.
- [ ] **Tarefa 2 (sem Jev):** Motor de regras determinístico: detecção de conflito por `files_touched`, branch existente (Regra 15), tasks pendentes e critério de conclusão da feature.
- [ ] **Tarefa 3:** Schema da decisão e registro em `sdd/.decisions/`, mais extensão da telemetria para amarrar sessões/decisões à mesma `feature` (reaproveita discovery-53, Tarefa 3).
- [ ] **Tarefa 4:** Cliente Jev via OpenRouter (`net/http`, chave por env, timeout, validação de schema, redação do snapshot) e bloco `dispatcher` no `.sddrc` com `enabled=false` por padrão.
- [ ] **Tarefa 5:** Comando `forge-sdd dispatch` em autonomia `L0`/`L1`: o Jev recomenda a etapa e o agente, o humano confirma; fallback determinístico/humano em falha.
- [ ] **Tarefa 6:** Prompts de handoff por estação e integração com `/nova-feature`, `/proxima-feature` e `/revisar` (Claude, Gemini, Copilot), declarando o que é automático só no Claude.
- [ ] **Tarefa 7 (piloto, decisão de produto primeiro):** Autonomia `L2` no Claude Code via sessões Claude Code Remote (`create_session`/`send_message`), com gates humanos em merge/release, escopo ambíguo, conflito e baixa confiança.
- [ ] **Tarefa 8 (documentação):** Atualizar `sdd/FLOW.md` (o despachante e as estações), README/CLAUDE.md/GEMINI.md/Copilot, release notes (Regra 12) e a Constituição se uma nova regra for necessária (hoje 15/15 — exigiria consolidar).

## Observação de Priorização

Tarefas 1-3 constroem a base **sem depender do Jev**: se o OpenRouter ou o modelo ficarem indisponíveis, o Forge ainda sabe quem está ocupado e o que foi entregue. A 4 e a 5 introduzem o Jev apenas como conselheiro (`L0/L1`), o menor risco. A 7 (autonomia) só começa após confirmação explícita do usuário sobre níveis de autonomia, privacidade dos dados enviados ao provedor e assimetria entre agentes. O limite de 8 tarefas excede a regra dos 7 (`FLOW.md`): o `/split-features` deve quebrar em subfeatures, por exemplo **(A) base determinística** (T1-3), **(B) Jev advisor** (T4-6) e **(C) autonomia + docs** (T7-8).

## Métrica de Sucesso

- `forge-sdd run status` mostra, para qualquer feature, estado de cada estação, dono, última atividade e conclusão — hoje: impossível sem ler o chat.
- Em testes, 100% das decisões inválidas/timeout do Jev caem no fallback sem avançar a esteira.
- Zero conflito de branch (Regra 15) e zero merge/release autônomo em qualquer nível de autonomia.
- Com `dispatcher.enabled=false`, nenhuma mudança de comportamento nos comandos atuais.
