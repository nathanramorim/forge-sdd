# feat-58-06 — Motor de Auto-Skills (`forge-sdd skills`) via Regra de Três (`DUP-05`) e Fim de Onda (`GOV-03`)

**Branch:** `feat/58-orquestracao-autonoma-governanca-skills`
**Fase:** 58-06
**Depende de:** `feat-58-02`
**Status:** `done`

## Objetivo

Implementar o motor de **Auto-Skills** (`internal/skills` + subcomando `forge-sdd skills`) para transformar atividades padronizadas e repetidas pelo Builder (detectadas pelo Analista de Governança via **Regra de Três `DUP-05`** ou na revisão de **Fim de Onda `GOV-03`**) em skills reutilizáveis em `.agents/skills/<nome>/SKILL.md`, sincronizando-as automaticamente para os diretórios de skills dos agentes configurados (`.claude/skills/` e `.gemini/skills/`) após validação humana no Líder.

## Modelo do módulo (`ARQ-11`)
- **Modelo:** `completo` (pacote `internal/skills` + subcomando Cobra `forge-sdd skills` + índice em `sdd/skills/index.md`).
- **Justificativa:** Gerencia ciclo de vida de skills (`proposed → approved → synced`), validação de frontmatter e propagação idempotente entre múltiplos agentes.

## O que já existe (`DUP-02`)

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
| --- | --- | --- | --- | --- |
| Aprendizado contínuo (`lessons.md`) | `internal/lessons/` (`forge-sdd lessons add`, `feat-01-03`) | Registra lições curtas de fixes em `sdd/memory/lessons.md` | `reusar` (padrão CLI) | Seguir a mesma ergonomia de comando determinístico da CLI, diferenciando *lição* (erro a evitar) de *skill* (procedimento padronizado reutilizável) |
| Regra de Três e Fim de Onda | `busize-solutions/planning-brain-governanca` (`DUP-05` e `GOV-03`) | Critério objetivo para extrair padrão na 3ª repetição e consolidar aprendizados no fim de onda | `reusar` | Usar como gatilho no Analista de Governança (`S-07`) e no `/archive` para propor novas skills |
| Gerenciador de Skills | `internal/` | `feat-35` removeu o antigo `install-skill` externo por URL; não há gerenciador de skills locais do projeto | `criar` | Criar `internal/skills` operando 100% local sobre `.agents/skills/<nome>/SKILL.md` (`DUP-10`) |

## Dependências novas (`SEG-07`)
Nenhuma.

## Rules e IDs da Constituição Aplicáveis
- **Rules:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-reuso-e-duplicacao.md`, `.agents/rules/sdd-codigo-e-testes.md`
- **IDs:** `DUP-05`, `DUP-10`, `GOV-03`, `TST-01`.

## Critérios de Aceitação Executáveis

1. **Subcomando `forge-sdd skills` (`internal/skills`):**
   - `forge-sdd skills propose --name <slug> --description "<desc>" --steps "<markdown>"`: cria `.agents/skills/<slug>/SKILL.md` com frontmatter (`name`, `description`, `status: proposed`) e registra em `sdd/skills/index.md`.
   - `forge-sdd skills approve <slug>`: muda o status para `approved` e executa o `sync` automaticamente.
   - `forge-sdd skills sync`: para cada skill `approved` em `.agents/skills/<slug>/SKILL.md`, gera/atualiza os adaptadores nos agentes configurados em `sdd/.sddrc` (`.claude/skills/<slug>/SKILL.md` e `.gemini/skills/<slug>/SKILL.md`), preservando skills manuais criadas diretamente pelo usuário.
   - `forge-sdd skills list`: exibe skills propostas e ativas.
2. Atualizado `.agents/commands/archive.md` (e seu `.md.tmpl`) para incluir a revisão de **Fim de Onda (`GOV-03`)**: verificar padrões repetidos nas tasks arquivadas e sugerir `forge-sdd skills propose` quando aplicável.
3. Testes unitários em `internal/skills` cobrindo `propose`, `approve`, `sync` e `list`; `go test ./...` e `go vet ./...` passando 100%.

## Handoff
Pode rodar em paralelo com `feat-58-04`. Desbloqueia o fechamento na `feat-58-07`.
