# fix-80-01 — Renomear o slash `/status` para `/forge-status`

**Branch:** `fix/80-01-renomear-slash-status`
**Origem:** Iniciativa 80 (`/status` tríplice de baixo consumo de tokens)
**Tamanho:** `M` (muitos arquivos, nenhuma lógica nova)
**Status:** `spec` (aguarda Gate 1)

## Sintoma

O comando `/status` que o forge-sdd instala nos agentes tem o **mesmo nome** do `/status` nativo das CLIs dos agentes. Quando os dois existem, não fica claro qual roda: depende da CLI e da versão. O dono relatou o conflito ao usar a CLI.

## Esclarecimento (clarify)

Respostas do dono nesta rodada, registradas para não se perderem:

- **Conflito:** com o comando nativo da CLI do agente (não com o subcomando `forge status` do binário).
- **Novo nome:** `/forge-status`. O `forge status` do binário **não muda**.
- **Branch de partida:** `main`, em worktree própria (a árvore principal estava com mudanças não commitadas de outra iniciativa).

## Feature Correlacionada

- **Arquivo de origem:** `.agents/skills/sdd-operacao/references/status.md` e os arquivos de comando `status` de cada agente (`.claude/commands/status.md`, `.gemini/prompts/status.prompt.md`, `.github/prompts/status.prompt.md`), todos gerados de `internal/scaffold/templates/`.
- **Causa-raiz:** o nome foi escolhido sem checar se as CLIs de agentes já o usam. Dois comandos com o mesmo nome disputam a resolução.
- **Evidência do defeito:** o relato do dono. **A confirmar na implementação:** que Claude Code e Codex expõem um `/status` nativo. Isto não foi verificado na documentação oficial nesta spec; o critério 6 exige a conferência antes do merge.

## Modelo do módulo

**Enxuto** (`ARQ-11`): é renomeação de arquivos de template, texto de documentação e goldens. Não há pacote novo, função nova de domínio nem dependência.

## Decisões do projeto aplicadas

- `DUP-01` — o nome do comando passa a existir em **um** só lugar por agente (o arquivo do template); a documentação só referencia.
- `COD-04` — o arquivo antigo `status` sai na mesma entrega; não fica alias nem arquivo morto.
- `FLX-07` — renomear comando que o usuário digita é mudança visível: entra primeiro numa **beta** e na nota de release (`sdd/releases/history.md`).
- `GOV-04` — a explicação operacional fica em `references/`, a Constituição não muda.
- `TST-03` — teste com caso positivo (projeto novo recebe `forge-status`) e negativo (projeto antigo perde o `status` antigo no `update`).
- ADR-0021 (obsolescência das 5 estações) — precedente de remover comando com nota de release.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Arquivos de comando `status` por agente | `internal/scaffold/templates/agents/claude/.claude/commands/status.md.tmpl`, `.../gemini/.gemini/prompts/status.prompt.md.tmpl`, `.../.github/prompts/status.prompt.md.tmpl` | 3 templates | **renomear** para `forge-status` |
| Referência da skill | `.agents/skills/sdd-operacao/references/status.md` (+ `.tmpl`) e `SKILL.md` | tabela `/status` → `references/status.md` | **renomear** o slash; manter o arquivo `references/status.md` (o nome do arquivo não é digitado pelo usuário) |
| Texto que cita o slash | `CLAUDE.md`, `GEMINI.md`, `AGENTS.md`, `copilot-instructions.md` (+ templates), `README.md`, `docs/comandos.md`, `sdd/HOWTO.md.tmpl`, `sdd/README.md.tmpl`, `references/toca-ai.md` | ~20 ocorrências, lista fechada no critério 1 | **trocar** o texto |
| Lista de comandos do scaffold | `internal/scaffold/cheatsheet.go` (`commandOrder`) | `status` **não** está na lista; o cheat-sheet não o exibe | **não mexer** nessa lista |
| Remoção de arquivo obsoleto | `internal/scaffold/scaffold.go:365` (`cleanObsoleteFiles`) | precedente exato: o `.prompt.md` antigo do Claude é removido ali | **reusar**: acrescentar `status` antigo à lista de obsoletos |
| Preservação de prompts | `shouldPreserve` em `scaffold.go` | cobre `.claude/commands` (sempre regenerado); **a conferir** para `.gemini/prompts` e `.github/prompts` | **conferir** na implementação antes de remover o antigo |
| Registro de depreciados | `internal/deprecation` | registra uso de **subcomando da CLI** (`queue stations`), não de slash | **não usar**: o slash roda no agente, fora do binário |
| Goldens | `internal/scaffold/testdata/golden/` | 6 arquivos citam `/status` | **regenerar** com `go test ./internal/scaffold -update` |
| Histórico de release | `sdd/releases/history.md` | cita `/status` em versões antigas | **não alterar** o passado; só acrescentar a nota nova |

## Dependências novas (`SEG-07`)

`Nenhuma`.

## Rules e IDs da Constituição Aplicáveis

`DUP-01`, `DUP-02`, `COD-04`, `COD-05`, `FLX-07`, `GOV-04`, `TST-01`, `TST-03`.

Rules: `sdd-codigo-e-testes`, `sdd-fluxo-e-portao`.

## Escopo

**Dentro:**
1. Renomear os 3 arquivos de comando (e os templates) de `status` para `forge-status`.
2. Trocar `/status` por `/forge-status` na documentação e nos templates da lista fechada do critério 1.
3. Remover o arquivo `status` antigo em projetos que já o têm (`update`).
4. Regenerar os goldens e acrescentar a nota na `history.md` da próxima beta.

**Fora:**
- O subcomando `forge status` do binário e o `forge status --live` (não conflitam e continuam).
- O nome do arquivo `references/status.md` e o `status.md` das iniciativas (`sdd/discovery/<id>/status.md`), que são outra coisa.
- Alias `/status` apontando para o novo: seria o mesmo conflito de volta (`COD-04`).

## Critérios de Aceitação Executáveis

```bash
set -o pipefail
# 1. O slash antigo saiu da documentação e dos templates (não pega status.md).
! grep -rInE '(^|[^a-zA-Z0-9/_.-])/status([^.a-zA-Z0-9_-]|$)' \
    internal/scaffold/templates .agents .claude .gemini .github \
    README.md docs/comandos.md CLAUDE.md GEMINI.md AGENTS.md || exit 1
# 2. Os arquivos novos existem e os antigos não.
test -f internal/scaffold/templates/agents/claude/.claude/commands/forge-status.md.tmpl
test -f internal/scaffold/templates/agents/gemini/.gemini/prompts/forge-status.prompt.md.tmpl
test -f internal/scaffold/templates/.github/prompts/forge-status.prompt.md.tmpl
test ! -f internal/scaffold/templates/agents/claude/.claude/commands/status.md.tmpl
# 3. Projeto antigo perde o comando antigo no update; projeto novo recebe o novo.
go test ./internal/scaffold/... -run 'TestUpdateRemoveStatusAntigo|TestInitCriaForgeStatus'
# 4. Goldens em dia.
go test ./internal/scaffold/...
# 5. O subcomando do binário continua.
go build ./... && ./forge-sdd status --help >/dev/null
# 6. Conferência manual antes do merge (não automatizável aqui):
#    confirmar na documentação da CLI de cada agente que o /status nativo existe.
go vet ./... && go test ./...
```

- **Teste negativo obrigatório (`TST-03`):** `TestUpdateRemoveStatusAntigo` cria um projeto com `.claude/commands/status.md`, roda o `update` e prova que o arquivo antigo sumiu e o `forge-status.md` apareceu, **sem apagar** outros arquivos do usuário na mesma pasta.
- **Risco:** quem tem o hábito de digitar `/status` recebe a nativa da CLI, que é outra coisa. A nota de release precisa dizer isso com todas as letras.
