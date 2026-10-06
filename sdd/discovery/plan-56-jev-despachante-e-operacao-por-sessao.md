# Plano de Discovery 56 — Papéis Comuns, Agent Teams do Claude e Jev como Despachante

Roadmap preliminar (revisão 2). Ordem do menor risco ao maior: primeiro o que **não depende nem do Jev nem de agent teams** (papéis comuns e base determinística), depois o Jev como conselheiro e, por último, teams e autonomia.

## Subfeatures sugeridas para o `/split-features` (excede a regra dos 7 em um único bloco)

### A — Papéis comuns (novo; sem dependência de Jev ou teams)
- [ ] **A1:** Extrair os 7 papéis dos chatmodes duplicados (Gemini/Copilot) para a fonte canônica `.agents/roles/<papel>.md`, gerada pelo scaffold e embutida via `embed.FS`.
- [ ] **A2:** Gerar `.claude/agents/<papel>.md` (frontmatter `name`, `description`, `tools`, `model: inherit` + ponteiro ao papel canônico). Revisor sem `Edit`/`Write`; Archivist só Read/Edit. `update` preserva agentes do usuário; `doctor` detecta papel ausente.
- [ ] **A3:** Converter os chatmodes de Gemini e Copilot em adaptadores finos para o mesmo papel canônico (remove a duplicação), com golden tests e regressão de comportamento.
- [ ] **A4:** Atualizar `CLAUDE.md`/`FLOW.md` para citar os papéis como definições reais (subagents por padrão) e resolver, nos docs, o risco de paridade do spike `feat-5ae2-06`.

### B — Base determinística (sem Jev)
- [ ] **B1:** Ledger por feature (`sdd/.runs/`) com estado por estação, lease/heartbeat e `forge-sdd run status` — fonte de verdade entre sessões.
- [ ] **B2:** Motor de regras: conflito por `files_touched`, branch existente (Regra 15), tasks pendentes, critério de conclusão; comando `forge-sdd run verify` (exit 2 se o critério falhar).
- [ ] **B3:** Schema e registro de decisões em `sdd/.decisions/`; telemetria correlacionada por `feature` (N sessões/teammates → N `session-*.json`).

### C — Jev como conselheiro (via OpenRouter)
- [ ] **C1:** Cliente OpenRouter (`net/http`, chave por env, timeout, schema validado, snapshot sem código) e bloco `dispatcher` no `.sddrc` (`enabled=false`).
- [ ] **C2:** `forge-sdd dispatch` em `L0`/`L1`: Jev recomenda etapa, **papel** e agente; o lead/usuário executa; fallback determinístico/humano em falha.

### D — Agent Teams e autonomia (opt-in, depende de decisão do usuário)
- [ ] **D1:** `forge-sdd init --claude-teams`: grava a variável experimental em `.claude/settings.json` com aviso de custo, de recurso experimental, de indisponibilidade em `-p` e do efeito sobre subagents nomeados. Desligado por padrão.
- [ ] **D2:** Hooks opcionais: `TaskCompleted` → `run verify`; `TeammateIdle` → checa handoff no ledger.
- [ ] **D3:** Playbooks em `sdd/FLOW.md` para quando usar teams: revisão paralela por lentes (Constituição/testes/segurança), discovery paralelo, features paralelas com arquivos disjuntos. Esteira Spec → Act → Revisor de uma feature segue em subagents.
- [ ] **D4 (piloto):** Autonomia `L2` no Claude: o lead cria teammates/subagents conforme a decisão do Jev, com gates humanos em merge/release, escopo ambíguo, conflito e baixa confiança.
- [ ] **D5:** Documentação, release notes (Regra 12) e, se necessário, ajuste da Constituição (hoje 15/15, exigiria consolidar).

## Observação de Priorização

- **A vem primeiro** porque resolve diretamente o pedido (papéis do Forge também no Claude), elimina duplicação existente e não depende de nada experimental. Entrega valor mesmo que o Jev e os teams nunca sejam adotados.
- **Teams não é o mecanismo padrão da esteira.** A documentação oficial indica sessão única ou subagents para trabalho sequencial e edição dos mesmos arquivos; teams entra para paralelismo (D). Por isso o desenho do discovery-53 (sessões Remote separadas como mecanismo principal) foi rebaixado a alternativa.
- **B antes de C:** se o OpenRouter ou o modelo Jev saírem do ar, o Forge ainda sabe quem está ocupado e o que foi entregue. D4 só começa após confirmação explícita do usuário sobre autonomia, privacidade e custo de tokens.

## Métrica de Sucesso

- `forge-sdd init` para Claude gera os 7 papéis em `.claude/agents/`, e os três agentes apontam para a **mesma** definição canônica (zero corpo duplicado).
- O Revisor gerado para o Claude não possui ferramentas de escrita de arquivo.
- `forge-sdd run status` mostra, para qualquer feature, o estado de cada estação, dono, última atividade e conclusão.
- 100% das decisões inválidas/timeout do Jev caem no fallback sem avançar a esteira; zero merge/release autônomo em qualquer nível.
- Sem `--claude-teams` e com `dispatcher.enabled=false`, nenhuma mudança de comportamento nos comandos atuais.
