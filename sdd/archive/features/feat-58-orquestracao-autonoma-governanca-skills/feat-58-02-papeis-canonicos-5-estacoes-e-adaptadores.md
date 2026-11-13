# feat-58-02 — Papéis Canônicos das 5 Estações (`.agents/roles/`) e Adaptadores Multi-Agente

**Branch:** `feat/58-orquestracao-autonoma-governanca-skills`
**Fase:** 58-02
**Depende de:** —
**Status:** `done`

## Objetivo

Unificar as definições de papéis do Forge-SDD numa única fonte da verdade em `.agents/roles/` (`DUP-10`), estruturando as 5 estações coordenadas (**Líder**, **Concepção PM + Arquiteto**, **Planejador**, **Builder** e **Analista de Governança** read-only com contrato de Revisor Semântico Camada B), e gerar adaptadores finos para Claude (`.claude/agents/*.md`), Gemini (`.gemini/skills/*.chatmode.md`) e Copilot (`.github/chatmodes/*.chatmode.md`).

## Modelo do módulo (`ARQ-11`)
- **Modelo:** `completo` (altera a árvore gerada pelo `internal/scaffold` para todos os agentes suportados).
- **Justificativa:** Envolve lógica de seleção de arquivos por agente no scaffold Go e sincronização de adaptadores.

## O que já existe (`DUP-02`)

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
| --- | --- | --- | --- | --- |
| Chatmodes Gemini e Copilot | `.gemini/skills/*.chatmode.md` e `.github/chatmodes/*.chatmode.md` | 7 papéis com corpo duplicado entre os dois agentes | `estender` | Extrair o corpo para `.agents/roles/<papel>.md` e transformar os arquivos específicos de agente em adaptadores finos (mesmo padrão de `.agents/commands/` da `feat-02-03`) |
| Prompt do Revisor Semântico (Camada B) | `busize-solutions/planning-brain-governanca` (`portao/prompt-portao-github-action.md` Parte B) | Instrução de revisor com entradas delimitadas, defesa contra prompt injection (`FLX-08`), checks `S-01..S-06`, exigência de evidência e saída JSON | `reusar` (genérico) | Incorporar ao papel canônico `.agents/roles/governance.md` |
| Subagentes nativos Claude | `.claude/` | Não existe `.claude/agents/` no scaffold atual | `criar` | Gerar `.claude/agents/<papel>.md` com frontmatter (`name`, `description`, `tools`, `model`) apontando para `.agents/roles/<papel>.md` |

## Dependências novas (`SEG-07`)
Nenhuma.

## Rules e IDs da Constituição Aplicáveis
- **Rules:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-seguranca.md`, `.agents/rules/sdd-reuso-e-duplicacao.md`
- **IDs:** `ARQ-03`, `SEG-04`, `DUP-10`, `FLX-08`, Regra 3 (`embed.FS`).

## Critérios de Aceitação Executáveis

1. Criados em `.agents/roles/` (e em `internal/scaffold/templates/.agents/roles/`) os papéis canônicos:
   - `lider.md` (evolução do `orquestrador`, coordenador das estações, mapa da Constituição e painel humano);
   - `pm-architect.md` (evolução de `specifier` + `c4-architecture`, responsável por `discovery` e `criteria` com Reuse Check);
   - `planner.md` (Planejador, responsável pelo `plan` e mapeamento de rules/IDs por task);
   - `builder.md` (Executor orientado pela fila em contexto isolado);
   - `governance.md` (Analista de Governança / Revisor Semântico Camada B read-only, com proteção anti-injection `FLX-08`, checks `S-01..S-07` e contrato JSON com evidência);
   - `archivist.md` e `migrator.md`.
2. Para o agente `claude`, o scaffold gera `.claude/agents/*.md` com frontmatter válido (`governance.md` restrito a ferramentas de leitura/verificação, sem `Edit`/`Write` em código de produção).
3. Para `gemini` e `copilot`, `.gemini/skills/*.chatmode.md` e `.github/chatmodes/*.chatmode.md` passam a ser adaptadores finos que referenciam `.agents/roles/*.md` (zero corpo duplicado).
4. `CLAUDE.md`, `GEMINI.md`, `OPENAI.md` e `.github/copilot-instructions.md` atualizados com a referência a `.agents/roles/`.
5. Golden tests atualizados; `go test ./internal/scaffold/...` e `go vet ./...` passam sem erros.

## Handoff
Pode rodar em paralelo com `feat-58-01`. Desbloqueia `feat-58-03` e `feat-58-05`.
