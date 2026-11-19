# fix-88-05 — forge rules: lookup de IDs da Constituição

**Branch:** `fix/88-05-forge-rules-lookup`
**Origem:** Iniciativa 88, `internal/gate/brief.go`, seção `IDs` do `forge brief`
**Tamanho:** `S`
**Status:** `done`

## Sintoma

`forge brief` mostra apenas os primeiros itens da seção `IDs` e corta "… 29 itens omitidos; leia sdd/memory/constitution.md". Quando o Reviewer menciona `FLX-08`, `GOV-04` ou `SEG-01` em vereditos e comentários, o usuário não tem como entender o significado sem abrir manualmente o arquivo da Constituição. Comportamento esperado: existe um comando rápido para consultar o significado de um ID.

## Feature Correlacionada

- **Arquivo de origem:** `internal/gate/brief.go` — `briefIDs` trunca a 70 chars e o budget da seção `IDs` limita a exibição (29 de ~44 IDs omitidos).
- **Causa-raiz:** `forge brief` tem orçamento de tokens fixo; expandir o budget quebraria a premissa de brief compacto. O gap é a falta de um subcomando dedicado para lookup de IDs.
- **Evidência:**
  - `forge brief | grep "omitidos"` → `… 29 itens omitidos; leia sdd/memory/constitution.md`
  - `internal/gate/brief.go:467-480` — `briefIDs` limita descrição a 70 chars

## Modelo do módulo

**Enxuto** (ARQ-11): novo subcomando `forge rules [ID...]` em `cmd/forge-sdd/` usando `ConstitutionIDLines` já existente em `internal/gate/`. Sem nova dependência.

## Decisões do projeto aplicadas

- ARQ-01 (lógica em internal/, cmd/ só adapta flags)
- ARQ-06 (CLI expõe subcomandos de ciclo de vida)
- COD-01 (responsabilidade única)

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Função de lookup de IDs | `internal/gate/brief.go` | `ConstitutionIDLines(constitution string) map[string]string` já mapeia ID→descrição completa | **reusar** diretamente |
| Subcomando `rules` | `cmd/forge-sdd/` | Ausente | **criar** |
| Subcomando `constitution` | `cmd/forge-sdd/` | Ausente; existe `/constitution` como slash command | **não duplicar** — slash command já existe; `forge rules` é CLI rápido |
| Leitura da constitution | `cmd/forge-sdd/brief.go` | `loadConstitutionPath` ou similar | verificar e reusar |

## Dependências novas

Nenhuma. Reusa `ConstitutionIDLines` e a leitura da constitution já feita pelo `forge brief`.

## Rules e IDs da Constituição Aplicáveis

`ARQ-01`, `ARQ-06`, `COD-01`, `DUP-01`.

## Escopo

1. `cmd/forge-sdd/rules.go`: novo subcomando `forge rules [ID...]`.
   - Sem argumentos: lista todos os IDs ordenados com prefixo e descrição completa (sem truncar).
   - Com IDs como argumento (`forge rules FLX-08 GOV-04`): exibe só os IDs pedidos com descrição completa.
   - Formato: `FLX-08  AskUserQuestion obrigatório para todas as perguntas de múltipla escolha` (alinhado por colunas).
   - Saída em texto simples (não JSON por padrão); flag `--json` opcional.
2. Registrar o subcomando no root em `cmd/forge-sdd/main.go` ou equivalente.
3. Testes unitários básicos (tabela de IDs conhecidos).

## Fora do escopo

- Mudar o orçamento ou o formato do `forge brief`.
- Mudar como o Reviewer referencia IDs em vereditos.
- Subcomando `forge constitution` (slash command já cobre esse caso).

## Critérios de Aceitação Executáveis

```bash
# Lista todos os IDs sem truncar
forge rules | grep "FLX-08"
# Deve retornar: FLX-08  <descrição completa sem "…">

# Lookup de ID específico
forge rules FLX-08 GOV-04
# Deve exibir exatamente 2 linhas com FLX-08 e GOV-04

# ID inexistente
forge rules FOO-99
# Deve retornar exit code != 0 e mensagem clara

go build ./...
go vet ./...
go test ./...
```
