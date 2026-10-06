# Plano de Discovery 56 — Papéis Especialistas, Agent Teams e Jev (opcional)

Roadmap preliminar (revisão 3). Ordem do menor risco ao maior. **O Jev é opcional e fica por último**: A, B, D e E não dependem dele.

## Subfeatures sugeridas para o `/split-features`

### A — Papéis comuns (sem Jev, sem teams)
- [ ] **A1:** Extrair os 7 papéis dos chatmodes duplicados (Gemini/Copilot) para `.agents/roles/<papel>.md`, gerado pelo scaffold e embutido via `embed.FS`.
- [ ] **A2:** Gerar `.claude/agents/<papel>.md` (frontmatter `name`, `description`, `tools`, `model` + ponteiro ao papel canônico). Revisor sem `Edit`/`Write`; Archivist só Read/Edit. `update` preserva agentes do usuário; `doctor` detecta papel ausente.
- [ ] **A3:** Converter os chatmodes de Gemini e Copilot em adaptadores finos do mesmo papel canônico, com golden tests e regressão de comportamento.
- [ ] **A4:** Atualizar `CLAUDE.md`/`FLOW.md` (papéis como definições reais; subagents por padrão).

### B — Base determinística e economia de tokens (sem Jev)
- [ ] **B1:** Ledger por feature (`sdd/.runs/`) com estado por estação, lease/heartbeat e `forge-sdd run status`.
- [ ] **B2:** Motor de regras (conflito por `files_touched`, Regra 15, tasks pendentes, conclusão) e `forge-sdd run verify` (exit 2 se o critério falhar).
- [ ] **B3:** `forge-sdd dispatch` **sem modelo**: devolve etapa, papel, ocupação, conflito e conclusão só com regras.
- [ ] **B4:** Modelo por papel (`roles.<papel>.model` no `.sddrc`, padrões econômicos nos mecânicos) e `forge-sdd report --by-role` para ajustar com dados.
- [ ] **B5:** Registro de decisões em `sdd/.decisions/` e telemetria correlacionada por `feature` (N sessões/teammates → N `session-*.json`).

### C — Especialistas de domínio (um agente por especialidade)
- [ ] **C1:** `forge-sdd agents sync`: gera `.claude/agents/esp-<dominio>.md` a partir de `.agents/rules/*.md`, sem nunca alterar as regras.
- [ ] **C2:** Órfãos, `doctor` (limite de especialistas, descrições sobrepostas) e orientação de `description` curta para delegação precisa.
- [ ] **C3:** Adaptadores equivalentes para Gemini/Copilot (mesmo conteúdo, mecanismo próprio de cada agente).

### D — Agent Teams (opt-in, depende de decisão do usuário)
- [ ] **D1:** `forge-sdd init --claude-teams`: grava a variável experimental em `.claude/settings.json` com aviso de custo, de recurso experimental, de indisponibilidade em `-p` e do efeito sobre subagents nomeados. Desligado por padrão.
- [ ] **D2:** Hooks opcionais: `TaskCompleted` → `run verify`; `TeammateIdle` → checa handoff no ledger.
- [ ] **D3:** Playbooks em `sdd/FLOW.md`: teams só para revisão por lentes, discovery paralelo e features com arquivos disjuntos (3-5 teammates); a esteira de uma feature segue em subagents.

### E — Jev opcional via OpenRouter (por último)
- [ ] **E1:** Cliente OpenRouter isolado em pacote próprio (`net/http`, chave por env, timeout, schema validado, snapshot sem código) e bloco `dispatcher` no `.sddrc` (`enabled=false`).
- [ ] **E2:** `dispatch` consulta o Jev **só quando as regras não resolvem**; fallback determinístico/humano em falha; níveis `L0`/`L1`.
- [ ] **E3 (piloto, decisão do usuário):** autonomia `L2` no Claude, com gates humanos em merge/release, escopo ambíguo, conflito e baixa confiança.
- [ ] **E4:** Documentação, release notes (Regra 12) e, se necessário, ajuste da Constituição (hoje 15/15, exigiria consolidar).

## Observação de Priorização

- **A resolve o pedido central** (papéis do Forge também no Claude) e elimina duplicação, sem depender de nada experimental ou externo.
- **B reduz tokens por construção:** estado e decisão de processo saem do ledger e do git, não de um LLM; o modelo por papel e o relatório por papel permitem ajustar o custo com dados.
- **C entrega "um agente por especialidade"** sem duplicar conhecimento: o agente aponta para a regra de domínio, que continua sendo a fonte única.
- **D e E são opcionais e desligados por padrão.** Teams custa mais tokens; o Jev depende de terceiro. Nenhum deles é pré-requisito dos demais.

## Métrica de Sucesso

- `forge-sdd init` para Claude gera os 7 papéis em `.claude/agents/`; os três agentes apontam para a **mesma** definição (zero corpo duplicado).
- `forge-sdd dispatch` responde sem nenhuma chamada de rede com o Jev desligado.
- `forge-sdd report --by-role` mostra tokens por papel; o modelo padrão de cada papel é ajustado a partir desses dados.
- Um especialista é gerado por regra de domínio, com `.agents/rules/` intacto após o comando.
- Revisor do Claude sem ferramentas de escrita; zero merge/release autônomo; sem `--claude-teams` e sem Jev, nenhuma mudança de comportamento nos comandos atuais.
