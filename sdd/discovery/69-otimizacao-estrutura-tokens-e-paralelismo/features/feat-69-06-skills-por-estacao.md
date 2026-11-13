# feat-69-06 — Skills por Estação

**Branch:** `feat/69-otimizacao-estrutura-tokens-e-paralelismo`
**Fase:** 69-06
**Tamanho:** `G`
**Depende de:** feat-69-02
**Status:** `approved` (FLX-01 aprovado)

> Rule of 7: tamanho `G`. A Governança pode exigir dividir por skill no Portão.

## Contexto & Motivação
13 comandos e 7 papéis (cerca de 65 KB em `.agents/`) são replicados em 63 arquivos de adaptadores em `.claude/`, `.gemini/` e `.github/`, com camadas adaptador → comando → papel → rules. Eles viram 5 skills por estação, com carregamento sob demanda e adaptadores finos.

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** muda a fonte única dos artefatos de agente, o scaffold e os adaptadores de três ferramentas, e remove diretórios inteiros; exige contrato de formato e testes de sincronização (`ARQ-11`, `DUP-10`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Distribuição de skills | `internal/skills` (`Propose`, `Approve`, `Sync`, `List`) | Sincroniza `.agents/skills/` para cada agente | reusar | Mesmo mecanismo serve às skills de estação |
| Índice de skills | `sdd/skills/index.md` | 1 entrada aprovada | estender | Recebe as 5 skills |
| Comandos e papéis | `.agents/commands/*.md` (13), `.agents/roles/*.md` (7) | Fonte atual dos adaptadores | remover após consolidar | Conteúdo migra para `SKILL.md` e `references/` |
| Rules de domínio | `.agents/rules/*.md`, check P-23 | Gate valida frontmatter nesse caminho | reusar sem mover | Skills apenas citam os arquivos (`GOV-04`) |
| Adaptadores | `.claude/{agents,commands}`, `.gemini/{prompts,skills}`, `.github/{prompts,chatmodes}` | 54 arquivos finos mais 9 chatmodes | estender | Um adaptador por skill e por ferramenta, sem corpo |
| Instruções raiz | `CLAUDE.md`, `GEMINI.md`, `.github/copilot-instructions.md` | 4,5 KB cada, carregados sempre | estender | Reduzir ao mínimo e apontar para as skills |
| Fluxo | `sdd/FLOW.md` | 9,6 KB, §8 do Studio removido em 69-01 | estender | Reduzir ao pipeline e à tabela de estações |
| Clarify e nomenclatura | `sdd/memory/clarify.md`, `naming-convention.md` | Lógica única referenciada por 3 comandos | reusar | Viram `references/` da skill de concepção |
| Divergência de nome | `pm-architect.md` vs `/discovery` | Papel proíbe `discovery.md`; comando exige | corrigir | A skill adota `discovery.md` dentro da pasta da iniciativa |

## Dependências novas
- Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- `DUP-05`, `DUP-06`, `DUP-10`, `GOV-04`, `COD-01`, `COD-04`, `TST-02`, `TST-04`, `SEG-02`.
- Rules: `sdd-reuso-e-duplicacao.md`, `sdd-codigo-e-testes.md`.
- Gerais: `FLX-01`, `FLX-02`, `COD-02`, `COD-05`, `TST-05`.

## Escopo
1. Criar em `.agents/skills/`: `sdd-concepcao` (discovery, nova-feature, novo-fix, c4-architecture, constitution + `pm-architect`), `sdd-planejamento` (split-features + `planner`), `sdd-build` (proxima-feature + `builder`), `sdd-governanca` (revisar + `governance`) e `sdd-operacao` (status, doctor, archive, upgrade-sdd, tutorial + `lider`, `archivist`, `migrator`).
2. Cada `SKILL.md` ≤ 2.048 bytes com gatilho, passos e links; detalhe em `references/` carregado sob demanda.
3. Remover `.agents/commands/` e `.agents/roles/`; manter `.agents/rules/` e `.agents/skills/novo-subcomando-cli-sdd`.
4. Adaptadores de comando por ferramenta ≤ 300 bytes e definições de papel (`.claude/agents`, chatmodes) ≤ 420 bytes, ambos sem repetir o corpo; `CLAUDE.md`, `GEMINI.md`, `OPENAI.md` e `copilot-instructions.md` renderizados ≤ 1.536 bytes. Os adaptadores de comando e as definições de subagente/chatmode (inclusive o `tools:` somente leitura da Governança) mantêm formato e local de cada ferramenta; só o destino do ponteiro muda.
5. Atualizar templates embutidos e goldens; `forge-sdd skills sync` gera os adaptadores.
6. **Comando `/implementar [nome-da-feature]`** na skill `sdd-build`, que substitui `/proxima-feature` (o nome antigo fica como alias de uma linha no adaptador, sem corpo, por `DUP-10`):
   - **Sem argumento:** executa a fila da iniciativa da branch atual. Para cada task em ordem de dependência: `forge-sdd queue next`, implementa, roda os critérios da spec e `forge-sdd gate`, `forge-sdd queue complete`, e passa a task à Governança; para ao esgotar a fila ou ao falhar o portão.
   - **Com argumento:** resolve o nome contra `sdd/discovery/*<nome>*` (e, no legado, `sdd/archive/`). Se houver mais de uma correspondência, lista e pergunta. Confirma a branch única da iniciativa (`Regra 15`: pergunta o ponto de partida e reaproveita branch existente) antes de executar a fila dela.
   - Antes de iniciar, o comando reporta o plano (iniciativa, branch, tasks na fila) e aguarda confirmação (`CLAUDE.md`, passo PLAN); `--yes` dispensa a confirmação em modo autônomo.
   - Grava telemetria ao fechar cada task (`forge-sdd session record`).

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./internal/skills/... ./internal/scaffold/...
test ! -d .agents/commands && test ! -d .agents/roles
for s in sdd-concepcao sdd-planejamento sdd-build sdd-governanca sdd-operacao; do test "$(wc -c < .agents/skills/$s/SKILL.md)" -le 2048; done
go test ./internal/skills/... ./internal/scaffold/... -run 'Sync|Orcamento|Legados|Apontam'
```
- Segunda execução do `sync` não altera arquivos (idempotência).
- Existe o adaptador `implementar` em cada ferramenta (≤ 300 bytes) apontando para `sdd-build`; `proxima-feature` é apenas alias. `grep -c 'queue next' .agents/skills/sdd-build/SKILL.md` ≥ 1 e o `SKILL.md` descreve os modos com e sem argumento.
- Teste (`budget_test.go`): nenhum adaptador excede o orçamento, nenhum arquivo gerado cita `.agents/commands/` ou `.agents/roles/` e todo ponteiro `.agents/skills/...` aponta para arquivo existente (`DUP-10`).
- Skills de estação têm `scope: core`: o `Sync` não as copia para `.claude/skills` nem `.gemini/skills`, e `skills approve` as recusa. As demais skills em `.agents/skills/` seguem preservadas no `update`; só `sdd-*` é regenerada.
- Check P-23 continua passando (rules intactas).
- Violações (`TST-03`): skill sem frontmatter, rule citada inexistente e `SKILL.md` acima do limite falham.
