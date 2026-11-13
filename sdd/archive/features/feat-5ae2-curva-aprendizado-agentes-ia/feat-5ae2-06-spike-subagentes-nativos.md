# Feature 5ae2-06 — Spike: Subagentes Nativos para Orquestrador/Builder/Revisor

Investigação técnica (não obrigatoriamente shippada nesta rodada) sobre substituir as "personas simuladas num único chat" por delegação real de tarefas, aproveitando primitivos nativos de subagentes já disponíveis em parte do ecossistema de IA agêntica.

## Critérios de Aceitação Executáveis

1. Documento de spike mapeando, para cada agente suportado (Claude, Gemini, Copilot), se existe hoje um primitivo nativo equivalente a "subagente com contexto isolado" (ex: `.claude/agents/*.md` no Claude Code).
2. Tabela de suporte por agente + recomendação objetiva de ir/não ir adiante com a migração, registrada como anexo desta task.
3. Não é necessário implementar a migração completa nesta feature — a saída é a decisão informada, não o código.

## Status: done (spike concluído, sem código)

### Tabela de Suporte por Agente

| Agente | Primitivo nativo de subagente com contexto isolado? | Notas |
|---|---|---|
| **Claude (Claude Code)** | **Sim.** Suporta definições de subagente em `.claude/agents/*.md` (front-matter com `name`, `description`, `tools`, `model`), invocados via delegação de tarefa (`Agent`/`Task` tool) em uma janela de contexto separada da conversa principal. | Primitivo maduro e documentado; é literalmente o mecanismo usado para orquestrar agentes especializados (ex: `Explore`, `Plan`) dentro do próprio Claude Code. |
| **Gemini (Gemini CLI / Antigravity)** | **Parcial / não confirmado.** O Antigravity oferece `/goal` para execução autônoma em background (single-agent, não multi-agente com contexto isolado). Não há confirmação de um primitivo público de "definição de subagente em arquivo" equivalente ao `.claude/agents/`. | Precisa de validação direta na documentação oficial mais recente antes de qualquer investimento — o ecossistema Gemini evolui rápido e pode já ter fechado essa lacuna. |
| **GitHub Copilot** | **Não, hoje.** Copilot Chat suporta "chat modes" (`.github/chatmodes/*.md`, já usado neste repositório) e Copilot Workspace/Coding Agent para tarefas multi-passo, mas sem um primitivo público de subagente com contexto isolado e invocável por definição de arquivo. | Chat modes já cobrem parte do ganho (especialização de persona), mas sem isolamento de contexto real. |

### Recomendação

**Ir adiante apenas com um piloto restrito ao Claude.** Migrar Orquestrador/Builder/Revisor para `.claude/agents/orquestrador.md`, `.claude/agents/builder.md`, `.claude/agents/revisor.md` é viável hoje e resolveria diretamente a Lacuna 5 da Discovery (personas simuladas vs. delegação real) — mas apenas para usuários de Claude Code. Gemini e Copilot continuam no modelo atual de "personas simuladas num único chat" até que seus respectivos ecossistemas confirmem um primitivo equivalente.

**Não migrar os três agentes de uma vez.** Uma migração assimétrica (Claude com subagentes nativos, Gemini/Copilot com prompts sequenciais) quebraria a promessa de paridade de comportamento entre agentes que o Forge-SDD sustenta hoje — introduzir essa divergência é uma decisão de produto, não só técnica, e deve ser levada ao usuário antes de qualquer implementação.

**Ação de acompanhamento sugerida:** reexecutar este spike a cada ciclo de `/doctor` ou de discovery futuro (o espaço de agentes de IA muda rápido o suficiente para que uma investigação de poucos meses atrás já possa estar desatualizada).
