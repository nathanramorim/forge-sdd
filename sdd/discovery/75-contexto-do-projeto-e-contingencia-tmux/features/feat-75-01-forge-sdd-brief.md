# feat-75-01 — `forge-sdd brief`

**Branch:** `feat/75-contexto-do-projeto-e-contingencia-tmux`
**Fase:** 75-01
**Tamanho:** `M`
**Depende de:** —
**Status:** `pending` (FLX-01 aprovado)

## Contexto & Motivação
A Discovery esquece as decisões do projeto porque nada entrega um resumo delas no começo da sessão. Esta task cria o comando que monta esse resumo, curto e determinístico, e o estado do contexto (`rascunho` ou `pronta`).

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** pacote novo de leitura (`internal/gate/brief.go`) com várias fontes, orçamento e contratos testáveis por camada (`ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Linhas de ID da Constitution | `internal/gate/context.go` (`ConstitutionIDLines`) | Regex `- **ID:** texto` | reusar | O brief lista ID e primeira frase com a mesma função (`DUP-01`). |
| Frontmatter das rules | `internal/gate/context.go` (`parseRuleFrontmatterRegras`) | Lê `regras:` | reusar | Cada rule aparece com `description` e IDs cobertos. |
| ADRs | `sdd/spec/adr/*.md` (template com `Status`) | Nenhum código lê | criar (pequeno) | Extrai título, status e a linha de decisão; `substituido` fica fora. |
| Módulos, fluxos, decisões | `sdd/spec/{modules,flows,decisions,stack,overview}.md` | Existem; nada lê | reusar | Resume tabelas e títulos e aponta o arquivo. |
| Comando irmão | `cmd/forge-sdd/context.go` | `context <task>` | estender | `brief` é subcomando novo seguindo `novo-subcomando-cli-sdd` (`DUP-05`). |

## Dependências novas
- Nenhuma. Stdlib Go (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`, `DUP-01`, `COD-01`, `TST-01`, `TST-04`.
- Rules: `sdd-arquitetura.md`, `sdd-codigo-e-testes.md`.
- Gerais: `FLX-02`, `COD-02`, `TST-05`.
- Skill: `novo-subcomando-cli-sdd`.

## Escopo
1. `internal/gate/brief.go`: `BuildBrief(dir, maxBytes) (Brief, error)` somente leitura. Seções em ordem: Estado do contexto, Missão e Stack, Regras do projeto, IDs canônicos (ID + 1ª frase), ADRs vigentes, Módulos, Fluxos, Decisões, Segurança, Rules. Orçamento padrão 6 KB; corte por seção, nunca no meio de linha, terminando com `… N itens omitidos; leia <arquivo>`.
2. Estado: `rascunho` quando a Constitution mantém a missão-modelo (`[Uma frase`), não tem decisão resolvida, ou `modules.md`/`flows.md` seguem com o texto-modelo; `pronta` caso contrário. `Brief.Missing` lista o que falta.
3. `cmd/forge-sdd/brief.go`: `forge-sdd brief [--dir] [--max-bytes] [--check]`. `--check` imprime só o estado e o que falta e sai com código 3 em `rascunho` e 0 em `pronta`, pelo mecanismo de saída já usado em `main.go`.
4. Nunca escreve em disco; arquivo ausente vira seção vazia, não erro.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./internal/gate/... ./cmd/forge-sdd/...
```
- `TestBrief`: projeto completo gera todas as seções na ordem, dentro do orçamento, com truncamento por seção e linha `… N itens omitidos`; ADR `substituido` não é vigente.
- `TestBriefEstado`: scaffold recém-gerado é `rascunho` com a lista do que falta; depois de preencher missão, decisões, módulos e fluxos é `pronta`.
- `TestBriefComando`: `brief` imprime o resumo; `--check` sai com 3 em rascunho e 0 em pronta; árvore de arquivos idêntica antes e depois.
- `forge-sdd gate` sem falhas.
