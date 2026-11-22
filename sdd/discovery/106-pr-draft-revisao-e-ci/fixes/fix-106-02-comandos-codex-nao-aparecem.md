# fix-106-02 — Ações Forge SDD não aparecem no seletor do Codex

**Iniciativa:** 106-pr-draft-revisao-e-ci  
**Branch da spec:** `fix/106-02-spec-codex`; **Branch da correção:** `fix/106-02-impl-codex`, aberta pelo Builder a partir da principal  
**Origem:** relato do usuário e auditoria do scaffold Codex (21/11/2026)  
**GitHub:** [issue #200](https://github.com/nathanramorim/forge-sdd/issues/200)  
**Tamanho:** `L`  
**Status:** `spec` (Conferência 1 em 22/11/2026 devolveu ajustes; Gate 1 depende do `[Seguir]` do dono e da confirmação manual no Codex)

## Descrição de negócio (release notes)

Quem usa Forge SDD no Codex poderá encontrar e acionar os fluxos de concepção e operação pelo mecanismo que o Codex realmente suporta, sem receber instruções para comandos que não aparecem no seletor.

## Problema

O scaffold do Forge SDD anuncia `/discovery`, `/implementar`, `/revisar`, `/status` e outros comandos no `AGENTS.md`, mas essa lista não registra comandos no Codex. Na instalação auditada, `.openai/prompts/` só contém `constitution.prompt.md` e `.agents/skills/` contém skills por estação, sem entradas específicas para todos esses fluxos. Por isso, digitar `/` no Codex não mostra os comandos do Forge.

O mecanismo antigo de custom prompts do Codex é local ao `CODEX_HOME`, não é compartilhado pelo repositório e está oficialmente obsoleto. A instalação não pode presumir que arquivos em `.openai/prompts/` serão descobertos pelo Codex nem prometer que skills criem aliases slash personalizados.

## Feature Correlacionada

- Origem: `sdd/discovery/106-pr-draft-revisao-e-ci/features/feat-106-26a-comandos-de-concepcao.md` e o scaffold/instruções Codex derivados de `internal/scaffold/`.
- Causa-raiz: a lista de slash commands foi tratada como instrução de prompt, sem artefato registrado na superfície de comandos do agente. Os arquivos `.openai/prompts/` não correspondem ao diretório local `CODEX_HOME/prompts` que o recurso legado lia; as skills existentes têm nomes de estação e não registram `/discovery` como alias.
- Evidência: `AGENTS.md` lista os aliases; `.openai/prompts/` não contém esses comandos; o Codex 0.171.0 instalado não os mostra ao digitar `/`. A documentação atual recomenda Skills e marca custom prompts como obsoletos.
- Sequenciamento: revisar a solução com a `feat-106-26a` antes de implementar; não editar os mesmos templates em paralelo.

## Modelo do módulo

**Enxuto** — correção de integração e artefatos gerados para Codex, sem novo serviço ou estado persistente (`ARQ-11`).

## Evidência sobre o Codex (a confirmar pelo dono)

Verificado: `codex --version` = 0.171.0; `~/.codex/prompts` não existe e `~/.codex/skills` está vazio; `internal/scaffold/templates/agents/openai/AGENTS.md.tmpl` (linha 19) lista `/discovery`, `/split-features`, `/implementar`, `/revisar`, `/status` e termina com `(.openai/prompts/)`; `cheatsheet.go` (linha 49) mapeia o destino `templates/agents/openai/.openai/prompts`.
**Não verificado** (só o dono confirma numa sessão limpa do Codex, e a spec cita a fonte ao confirmar): (a) que custom prompts estão oficialmente obsoletos; (b) que o seletor de `/` não mostra skills; (c) que `$nome-da-skill` invoca a skill. Enquanto (a)-(c) não forem confirmadas, o Builder **não** altera o texto do `AGENTS.md`.

## Decisões do projeto aplicadas

| Decisão | ID | Aplicação |
|---|---|---|
| Adaptadores finos sobre fonte única | `DUP-10`, ADR-0021 | A superfície de ativação do Codex aponta para `.agents/skills/` sem copiar o corpo dos comandos. (ADR-0003 e 0004 estão em parte substituídas pela ADR-0021.) |
| Template e golden no mesmo PR | `TST-02` | Mudou o texto gerado, muda o golden. |
| Caso negativo | `TST-03` | Os outros agentes mantêm os comandos. |
| Dependência nova exige consulta | `SEG-07` | Sem dependência; a superfície do Codex é confirmada na fonte citada. |

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Lista de comandos do Codex | `templates/agents/openai/AGENTS.md.tmpl` | anuncia `/` comandos sem registro | **corrigir** (só após a confirmação acima) |
| Prompts do Codex | `templates/agents/openai/.openai/prompts` | vazio no template; `constitution.prompt.md` só em instalação antiga | **decidir**: manter vazio e listar em `cleanObsoleteFiles` o que sair |
| Atualização sem impacto | fix-106-09 (`shouldPreserve`, `cleanObsoleteFiles`, `--dry-run`) | contrato de atualização | **seguir**: `AGENTS.md` sem o marcador `<!-- forge-sdd:generated -->` permanece intocado |
| Nome de `/status` | fix-106-09 | pode virar `/forge-status` | **usar o nome vigente na main** ao publicar a lista |
| References de concepção | feat-113-06 (mesclada) | já alteradas na main | **partir da main atual** |

## Dependências e ordem

- Não "depende" da `feat-106-26a` (ela mesma está bloqueada). **Colide** com ela nos goldens de `agents/openai`; vale a ordem de merge: quem chegar primeiro na principal, o outro rebaseia. A `feat-106-37` está superada pelo fix-106-10 e não conta.
- Estado: `spec`. O Builder só começa depois da confirmação manual do Codex e do Gate 1.

## Critérios de Aceitação Executáveis

Os testes novos ficam **vermelhos até o Builder** (antes da correção o golden do `AGENTS.md` do Codex ainda cita `/discovery` e `(.openai/prompts/)`). `-list` só prova existência; a aprovação é o `-run`.

```bash
set -o pipefail
# 1. o AGENTS.md do Codex não anuncia alias "/" sem registro
go test ./internal/scaffold -run 'TestCodexAgentsSemAliasNaoRegistrado' -count=1
# 2. o caminho suportado aparece (skill, apontando para .agents/skills)
go test ./internal/scaffold -run 'TestCodexCaminhoSuportadoApontaParaSkills' -count=1
# 3. os outros agentes mantêm os comandos (caso negativo, TST-03)
go test ./internal/scaffold -run 'TestOutrosAgentesMantemComandos' -count=1
# 4. atualização N-1 para N: AGENTS.md sem marcador intocado; prompts obsoletos listados; --dry-run lista o que sai
go test ./internal/scaffold ./cmd/forge-sdd -run 'TestCodexUpdateDeN1' -count=1
# 5. existência dos 4
[ "$(go test ./internal/scaffold ./cmd/forge-sdd -list 'TestCodex|TestOutrosAgentesMantemComandos' | grep -c '^Test')" -ge 4 ]
go test ./internal/scaffold -run 'Golden|ContextBudget' -count=1
```

6. **Manual, não verificável por CI:** numa sessão limpa do Codex (0.171.0 ou superior), o dono digita `/` e `$` e confirma o que aparece; o resultado e a fonte entram nesta spec antes do Gate 1.

## Fora do escopo

- Mudar os adaptadores de Claude Code, Gemini/agy e Copilot.
- Reescrever as skills `sdd-*`.
