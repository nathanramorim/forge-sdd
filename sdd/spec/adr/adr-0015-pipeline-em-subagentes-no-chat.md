# ADR-0015: Pipeline em Subagentes no Mesmo Chat (modo `chat`)

- **Data:** 2026-11-19
- **Status:** `proposto`
- **Regras afetadas (IDs da constitution):** `ARQ-06` (subcomandos `mcp` e `eval`; flag `doctor --harness`; texto da lista muda), `SEG-06`, `FLX-01`, `FLX-08`, `DUP-01`, `GOV-01` (esta ADR). Nenhum ID é criado ou removido.
- **Autor / Responsável:** Concepção (iniciativa `104-pipeline-em-subagentes`); decisão do dono pendente (Gate 1)

## Relação com ADRs vigentes
- **ADR-0004 (pipeline multi-sessão): estende.** O barramento, `ApplyVerdict`, o teto de retorno e o Go como dono do laço continuam; o modo `chat` chama as mesmas funções.
- **ADR-0008 (sessões com CLI aberta no tmux): convive.** O modo `sessions` (live e headless) segue o padrão e válido; nada é substituído ou removido.
- ADR-0013 (modo avançado) vale nos dois modos; ADR-0005 e ADR-0014 fornecem modelo por papel.
- **Constituição:** muda só o texto de `ARQ-06` (lista de subcomandos); como exige `GOV-01`, esta ADR vai no mesmo PR da implementação.

## Contexto
Hoje o pipeline são 3 sessões de terminal (tmux, barramento em `.forge/runtime/`, workers headless). O dono quer o fluxo no mesmo chat, com cada papel como líder que coordena subagentes e com a camada de entrada aceitando CLI, Claude, Codex e Agy, mais diagnóstico do harness, contratos validados e avaliação.

Viabilidade (detalhe em `discovery.md`): **confirmado** que os agentes de `.claude/agents/` não têm a ferramenta `Agent`, que o builder paraleliza como sessão de topo, e que o `claude` 2.1.329 contém `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS`, `TeamCreate` e `SendMessage`; **suposição** de que subagente não cria subagente e de que teams é de um nível; **não verificado**: documentação oficial, Codex, subagentes do Agy e do Gemini, hooks no app desktop.

## Opções Consideradas
1. **A. Chat como único coordenador:** os papéis são skills no chat; o chat cria subagentes A/B (um nível); o Go é dono do laço.
2. **B. Híbrido:** o chat dispara os workers headless atuais por Bash.
3. **C. Agent teams** (experimental, só Claude).
4. **D. Literal:** líder-agente cria subagentes (depende do spike).

## Decisão
- **D-1:** adotar **A** como modo `pipeline_mode: chat` (opt-in; padrão `sessions`).
- **D-2:** o Go é o dono do laço: ordem, limites, lock por iniciativa, retorno (teto 3) e escalada; o modelo devolve conteúdo por contratos com esquema estrito.
- **D-3:** Reviewer e subagentes de revisão somente leitura pela lista de ferramentas (`SEG-06`); retorno é dado (`FLX-08`); Gate 1 humano, salvo `advanced`.
- **D-4:** um worktree por subagente paralelo; conflito escala.
- **D-5:** diagnóstico `forge doctor --harness` exposto por MCP `forge_doctor` e hook `SessionStart`; instrução onde não há hook.
- **D-6:** spike condicional sobre D e C; só vira decisão por nova ADR.

## Consequências
- **Positivas:** fluxo no mesmo chat e em mais de uma superfície; paralelismo preservado; falhas de contrato e de concorrência deixam de passar despercebidas; avaliação comparável.
- **Negativas / Trade-offs:** o contexto do chat cresce; "líder" é papel, não agente; Codex e Agy podem cair para `sessions`; mais superfície de código (`internal/harness`).
- **Ações de acompanhamento:** executar o spike; verificar hooks no app desktop; medir pela avaliação antes de qualquer mudança de padrão.
