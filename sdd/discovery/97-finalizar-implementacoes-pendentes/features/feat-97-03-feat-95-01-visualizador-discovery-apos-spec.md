# feat-97-03 — feat-95-01: visualizador de artefatos da spec após conclusão

**Iniciativa:** 97-finalizar-implementacoes-pendentes
**Status:** ready
**Tamanho:** P

## Objetivo

Ao encerrar `/discovery`, `/nova-feature`, `/novo-fix` ou `/split-features`, antes da submissão ao Gate 1, apresentar um `AskUserQuestion` que permite revisar os artefatos gerados sem navegar manualmente.

**Sem mudança no binário Go.** Somente arquivos de skill e templates de scaffold.

## Escopo

### 1. Quatro arquivos de referência de skill (vivos)

Adicionar um bloco de visualização após a criação dos artefatos finais e antes de qualquer menção ao Gate 1 / queue / status.

Arquivos a alterar:
- `.agents/skills/sdd-concepcao/references/discovery.md`
- `.agents/skills/sdd-concepcao/references/nova-feature.md`
- `.agents/skills/sdd-concepcao/references/novo-fix.md`
- `.agents/skills/sdd-planejamento/references/split-features.md`

Bloco a inserir (adaptar o nome da pasta conforme o contexto):

```markdown
**Visualizar artefatos (antes do Gate 1)**

Use AskUserQuestion com as opções:
- "Abrir pasta no Finder" → `open sdd/discovery/<id>-<nome>/`
- "Abrir no VS Code" → `command -v code && code sdd/discovery/<id>-<nome>/` (fallback: Finder)
- "Ver lista dos arquivos" → list de arquivos em `sdd/discovery/<id>-<nome>/`
- "Pular"

Só execute a ação escolhida; não aguarde — o Gate 1 segue normalmente.
```

A seção canônica "Como perguntar com AskUserQuestion" fica em `sdd/memory/clarify.md`.
Referência nos arquivos de skill: `Consulte sdd/memory/clarify.md → seção "Visualização pós-spec"`.

### 2. Quatro arquivos `.tmpl` correspondentes

Espelhar as mesmas adições em:
- `internal/scaffold/templates/.agents/skills/sdd-concepcao/references/discovery.md.tmpl`
- `internal/scaffold/templates/.agents/skills/sdd-concepcao/references/nova-feature.md.tmpl`
- `internal/scaffold/templates/.agents/skills/sdd-concepcao/references/novo-fix.md.tmpl`
- `internal/scaffold/templates/.agents/skills/sdd-planejamento/references/split-features.md.tmpl`

### 3. `sdd/memory/clarify.md` — seção nova

Adicionar seção `## Visualização pós-spec` com:
- As 4 opções de ação disponíveis
- Que `open` é macOS-only; em outros SOs, omitir a opção Finder
- Que `code` precisa estar no PATH; verificar com `command -v code`
- Fallback: se Finder e VS Code indisponíveis, mostrar apenas "Ver lista" e "Pular"

### 4. `internal/scaffold/templates/sdd/memory/clarify.md.tmpl` — espelhar acima

Mesma seção `## Visualização pós-spec`.

### 5. Regenerar goldens do scaffold

```bash
go test ./internal/scaffold/... -run TestGolden -update
```

Os goldens em `internal/scaffold/testdata/golden/` devem espelhar os templates atualizados (TST-02).

## Decisões do projeto aplicadas

- `ARQ-01`: sem toque em `cmd/`; mudança só em arquivos de skill e templates embed.
- `ARQ-04`: sem nova dependência de runtime Go; `open` e `code` são ferramentas de OS invocadas pelo agente.
- `ARQ-06`: sem novo subcomando CLI.
- `DUP-06`: descrição de comportamento em `clarify.md` uma única vez; os quatro arquivos de referência apontam para ela.
- `TST-02`: mudança em templates de scaffold atualiza goldens no mesmo PR.

## Rules e IDs

`ARQ-01`, `ARQ-03`, `ARQ-04`, `ARQ-06`, `DUP-06`, `TST-02`

## Critérios de Aceitação Executáveis

```bash
go build ./...
go test ./internal/scaffold/... -run TestGolden
grep -r "Visualizar artefatos" .agents/skills/ --include="*.md" | wc -l
# deve retornar 4
grep "Visualização pós-spec" sdd/memory/clarify.md
```
