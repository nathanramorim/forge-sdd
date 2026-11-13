# feat-69-05 — Migração do Legado e Templates no Novo Layout

**Branch:** `feat/69-otimizacao-estrutura-tokens-e-paralelismo`
**Fase:** 69-05
**Tamanho:** `G`
**Depende de:** feat-69-03, feat-69-04
**Status:** `approved` (FLX-01 aprovado)

> Rule of 7: tamanho `G` por tocar `update`, templates e goldens juntos. A Governança pode exigir dividir em duas fases no Portão (migração; templates).

## Contexto & Motivação
Hoje 14 discoveries ficam soltos (42 arquivos) e `sdd/features/` tem 115 arquivos. O novo layout mantém tudo de uma demanda em `sdd/discovery/<id>-<nome>/`. Projetos em v3 recebem a mudança pelo `update` sem perder o conteúdo de domínio.

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** envolve movimentação em disco, backup, idempotência e dry-run, além de templates e goldens; o risco de perda de dados do usuário exige fronteiras e testes próprios (`ARQ-11`, `SEG-02`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Motor de scaffold e `update` | `internal/scaffold` (`Walk`, `Run`, `shouldPreserve`, `UpdateSddrc`) | Materializa templates e preserva domínio | estender | A migração roda como etapa do `update` |
| Backup | `.gitignore` (`sdd/backups/`) | Pasta já ignorada | reusar | Destino do backup antes de mover |
| `--dry-run` | `internal/scaffold`, `ARQ-05` | Imprime árvore sem escrever | reusar | Migração imprime movimentos sem tocar o disco |
| Detecção de versão | `sdd/.sdd-version`, `.sddrc` | Versão do scaffold instalada | reusar | Decide se há migração a aplicar |
| Caminhos | `internal/layout` (69-02) | Constantes e `Archive` | reusar | Destino `sdd/archive/` |
| Templates e goldens | `internal/scaffold/templates/sdd/**`, `testdata/golden/**` | Árvore v3 com `features/` e `memory/progress.md` | estender | Novo layout; goldens regenerados (`TST-02`) |
| Estado do 69-03 | `status.md`, `queue.json` | Definidos em 69-03 | reusar | Templates criam estes arquivos na iniciativa nova |

## Dependências novas
- Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-03`, `ARQ-05`, `SEG-01`, `SEG-02`, `COD-03`, `TST-02`, `TST-03`, `TST-04`, `FLX-02`, `GOV-01`.
- Rules: `sdd-arquitetura.md`, `sdd-seguranca.md`, `sdd-fluxo-e-portao.md`, `sdd-codigo-e-testes.md`.
- Gerais: `FLX-01`, `COD-02`, `COD-05`, `TST-05`.

## Escopo
1. `forge-sdd update` migra: discoveries soltos (`discovery-*.md`, `criteria-*.md`, `plan-*.md`) e `sdd/features/**` para `sdd/archive/`, preservando nomes; cria backup em `sdd/backups/<timestamp>/` antes.
2. Pastas de iniciativa que já existem (66, 67, 68, 69) permanecem no lugar.
3. Idempotente: segunda execução não altera nada. `forge-sdd update --dry-run` lista os movimentos e as remoções e não escreve (`ARQ-05`).
4. Templates e goldens passam a gerar o layout novo; remover `sdd/features/` e `memory/progress.md` do scaffold e criar `sdd/.runtime/.gitignore` (ignora o estado de execução sem editar o `.gitignore` da raiz, o que cumpre o escopo 2 da 69-03).
5. **Limpeza de artefatos obsoletos no `update`** (consequência da 69-06): o `update` hoje não remove arquivos que saíram dos templates, e um projeto atualizado ficaria com comandos e papéis antigos ao lado das skills, duplicando instruções e custo de tokens. Manter no scaffold uma lista explícita de caminhos obsoletos (`.agents/commands/*.md`, `.agents/roles/*.md` e os adaptadores antigos de `.claude/`, `.gemini/` e `.github/`) e removê-los apenas quando o conteúdo for idêntico ao do template da versão anterior (os templates anteriores ficam embutidos em `internal/scaffold/legacy/`, `ARQ-03`) ou carregar marcador gerenciado; arquivo modificado pelo usuário é preservado e listado no relatório (`SEG-02`). `--dry-run` imprime as remoções sem executar (`ARQ-05`).
6. ADR `adr-0002-layout-e-estado-por-iniciativa.md`; revisar texto das Regras 13 e 15 e de `FLX-02` na ADR (a edição da Constitution fica na 69-07).

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./internal/scaffold/... ./cmd/forge-sdd/...
go test ./internal/scaffold -run 'TestMigrate|TestGolden'
```
- Fixture v3 em `t.TempDir()`: `update --dry-run` imprime os movimentos e a árvore fica byte a byte igual (`ARQ-05`).
- `update` real move o legado, cria backup e a segunda execução não muda nada.
- Violações (`TST-03`): arquivo de domínio do usuário fora do padrão não é movido; destino já existente não é sobrescrito; link simbólico para fora do projeto é recusado (`SEG-02`).
- Merge de duas branches com iniciativas distintas, no fixture migrado, termina sem conflito.
- Fixture v3 com `.agents/commands/` e `.agents/roles/` intactos: o `update` remove os idênticos aos templates antigos, preserva o arquivo editado pelo usuário e é idempotente.
