# Plano Preliminar 02 — Fonte Única de Agente (`.agent/`) e Branch por Feature

Roadmap para `v2.2.0-beta` (bump explícito pedido pelo usuário a partir de `2.0.0-beta`, ver riscos em `discovery-02`). Ordem sugerida: base de menor risco (`.agent/rules/`, sem conteúdo pré-existente para migrar) → base de maior risco (`.agent/commands/` + adaptadores) → migração em `update` → comportamento de branch (maior superfície de prompt, mais validação).

## Estimativa de Quebra de Features

1. **`.agent/rules/` — convenção e scaffold em `init`**
   Criar `internal/scaffold/templates/.agent/rules/` com exemplo mínimo (`README.md` explicando a convenção + 1 arquivo de amostra comentado, ex: `design-system.md.example`). Cobrir em `forge-sdd init` (novo projeto) e golden fixtures. Sem risco de conteúdo pré-existente — pasta nova.

2. **`.agent/commands/` — corpo canônico por comando + adaptadores finos**
   Para cada um dos 12 comandos SDD, extrair o corpo de instrução comum para `internal/scaffold/templates/.agent/commands/<comando>.md.tmpl` (fonte única). Reescrever os templates hoje divergentes de `.claude/commands/`, `.gemini/prompts/`, `.github/prompts/` como adaptadores finos: mantêm frontmatter/sintaxe própria da ferramenta, e no corpo instruem "leia `.agent/commands/<comando>.md` e siga essas instruções". Validar que nenhum comando perde comportamento na transição (comparar cobertura de conteúdo antes/depois).

3. **Referência de rules nos três agentes**
   Adicionar instrução curta e equivalente em `CLAUDE.md.tmpl`, `GEMINI.md.tmpl` e chatmode(s) Copilot relevante(s) apontando para `.agent/rules/*.md` — consulta sob demanda antes de agir. Pode ser feita junto da feature 1 ou 2, já que é a mesma mecânica de referência.

4. **Migração aditiva em `forge-sdd update`**
   Estender o fluxo de `update` (mesmo ponto do feat-31 — preservação de domínio) com dois passos idempotentes: (a) cria `.agent/rules/` só se ausente, nunca sobrescreve regra existente; (b) cria `.agent/commands/` e converte adaptadores de `.claude/`/`.gemini/`/`.github/` para a forma fina **só quando o conteúdo atual bater com o template anterior conhecido** (sem customização manual detectada) — caso contrário, pula e reporta. Nenhum passo toca `sdd/features/`, `sdd/discovery/`, `sdd/fix/*`, `progress.md`. Cobrir com teste de idempotência (rodar `update` 2x), teste de detecção de customização, e `--dry-run` reportando sem escrever.

5. **Regra de lifecycle: branch única por pasta de feature quebrada**
   Atualizar `/nova-feature`, `/proxima-feature`, `/novo-fix` (agora via corpo canônico em `.agent/commands/`, propagando automaticamente aos três agentes) para detectar subpasta de feature (`sdd/features/feat-XX-nome/*.md`) e tratar como unidade de uma única branch (`feat/XX-nome`) agrupando as subtarefas. Preservar comportamento atual para features de arquivo único.

6. **Pergunta obrigatória de branch de partida e retomada**
   Na mesma mudança da feature 5: perguntar branch de partida (default `main`) e checar/perguntar sobre branch de feature já existente (`git branch --list feat/XX-*`) antes de criar uma nova. Documentar a regra correspondente em `sdd/memory/constitution.md` (nova regra, respeitando o limite declarado — revisar contagem atual antes de adicionar).

7. **Versão, documentação e release**
   Bump `sdd/.sddrc` e `internal/config/config.go` (`SddVersion`) para `2.2.0-beta`; atualizar `npm/package.json`; release notes em `sdd/releases/history.md` citando `.agent/` (rules + commands), o novo modelo de adaptador fino, e a nova pergunta de branch, identificando agente/comando conforme Regra 12 da Constituição.

## Observações para o Refinamento (`/split-features`)

- Features 1 e 3 são de baixo risco e podem entrar na mesma branch/PR (regras não têm conteúdo pré-existente a migrar).
- Feature 2 é a de maior superfície (toca os 12 comandos × 3 agentes) — considerar isolar em PR próprio, com um comando piloto primeiro (ex: `discovery`) antes de converter os outros 11, para validar o padrão de adaptador fino na prática.
- Feature 4 depende de 1 e 2 (precisa dos templates canônicos existirem para poder migrar/converter via `update`).
- Features 5 e 6 podem ser uma única feature (mesma superfície de prompt) ou duas, se o revisor preferir isolar "detecção de pasta" de "pergunta interativa" para reduzir risco de regressão por PR. Como já dependem do corpo canônico (feature 2), naturalmente vêm depois.
- Feature 7 fecha o pacote — só depois das demais estarem `done`, seguindo o padrão dos pacotes anteriores (feat-01, discovery-5ae2).
- Confirmar com o usuário, antes de codar a feature 7, se o pulo de `2.0.0-beta` para `2.2.0-beta` (sem passar por `2.1.0-beta`) é definitivo ou se prefere `2.1.0-beta` seguindo o padrão incremental já usado no projeto.

**Handoff:** Arquivos gerados nesta discovery — `discovery-02-agent-rules-e-branch-por-feature.md`, `criteria-02-agent-rules-e-branch-por-feature.md`, `plan-02-agent-rules-e-branch-por-feature.md` (em `sdd/discovery/`). Próximo passo: `/split-features`, organizando as features quebradas em `sdd/features/feat-02-agent-rules-e-branch-por-feature/`.
