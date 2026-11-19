# feat-97-01 — fix-90-02: tokens zerados em sessões de métricas e campo complexidade reservado

**Iniciativa:** 97-finalizar-implementacoes-pendentes
**Status:** ready
**Tamanho:** M

## Modelo do módulo

**Enxuto** (ARQ-11): sem pacote novo. Altera arquivos existentes em `internal/monitor/`, `cmd/forge-sdd/`, `.agents/skills/`, `internal/scaffold/templates/` e `sdd/.metrics/`. Nenhum novo package Go é criado.

## Objetivo

Corrigir o bug de `tokens_input: 0` / `tokens_output: 0` em `sdd/.metrics/session-*.json` nos paths de skill-template call, e reservar o campo `complexidade` no schema e na struct Go (sem lógica de preenchimento).

## Contexto

`forge session record` aceita zero sem estimar quando as flags são omitidas. Seis dos sete templates de skill não passam `--tokens-input`/`--tokens-output`. Resultado: a maioria das sessões de PM/Arquiteto, Governança, Planejador e Líder grava tokens zerados.

## Parte A — Bug: tokens zerados

### 1. Adicionar `--tokens-input` / `--tokens-output` em 7 arquivos de skill (vivos + template)

Nos 7 pares de arquivos (vivo em `.agents/skills/.../references/` + `.tmpl` em `internal/scaffold/templates/.agents/skills/.../references/`), adicionar a instrução de coleta de tokens ao final de cada fluxo — igual ao modelo de `implementar.md`:

```
Ao encerrar, chame:
forge session record --feature <id> --tokens-input <N> --tokens-output <N> --files-touched <lista>
Nunca passe zeros: se não souber os tokens, omita as flags (o CLI vai estimar).
```

Arquivos a corrigir:
- `.agents/skills/sdd-concepcao/references/novo-fix.md` + `.tmpl`
- `.agents/skills/sdd-concepcao/references/nova-feature.md` + `.tmpl`
- `.agents/skills/sdd-concepcao/references/discovery.md` + `.tmpl`
- `.agents/skills/sdd-governanca/references/revisar.md` + `.tmpl`
- `.agents/skills/sdd-operacao/references/lider.md` + `.tmpl` (parcial — completar)
- `.agents/skills/sdd-planejamento/references/split-features.md` + `.tmpl`
- `.agents/skills/sdd-operacao/references/archive.md` + `.tmpl`

### 2. `cmd/forge-sdd/session.go` — verificação de `cmd.Flags().Changed()`

No handler de `forge session record`, após parsear `tokens_input` e `tokens_output`:
```go
if !cmd.Flags().Changed("tokens-input") || !cmd.Flags().Changed("tokens-output") {
    est := monitor.EstimateSessionUsage(dir, feature, filesTouched)
    if !cmd.Flags().Changed("tokens-input") {
        tokensInput = est.InputTokens
    }
    if !cmd.Flags().Changed("tokens-output") {
        tokensOutput = est.OutputTokens
    }
}
```

### 3. `internal/monitor/usage.go` — nova função `EstimateSessionUsage`

```go
// EstimateSessionUsage estima o uso de tokens de uma sessão de skill a partir do tamanho
// do arquivo de feature (proxy para input) e dos arquivos tocados (proxy para output).
// Compartilha as mesmas constantes e lógica de EstimateJobUsage (DUP-01).
// reservado; preenchimento definido por ADR futura
func EstimateSessionUsage(dir, feature string, filesTouched []string) JobUsage {
    // tokens_input: bytes do arquivo de feature * 4 (heurística: 1 token ≈ 4 bytes)
    // fallback: 4000 bytes (mesmo padrão de EstimateJobUsage)
    // tokens_output: soma dos bytes de filesTouched * 4
    // fallback: 2000 bytes (mesmo padrão de EstimateJobUsage)
}
```

Testes em `internal/monitor/usage_test.go`:
- `TestEstimateSessionUsage_ComFeature`: feature existente → tokens_input = bytes * 4 (aproximado)
- `TestEstimateSessionUsage_SemFeature`: feature ausente → usa fallback 4000 bytes
- `TestEstimateSessionUsage_ComFilesTouched`: arquivos existentes → tokens_output = soma bytes * 4
- `TestEstimateSessionUsage_SemFiles`: files vazio → usa fallback 2000 bytes

## Parte B — Campo `complexidade` reservado

### 4. `internal/monitor/session_record.go` — adicionar campo à struct

```go
type SessionMetrics struct {
    // ... campos existentes ...
    Complexidade *string `json:"complexidade"` // reservado; preenchimento definido por ADR futura
}
```

Sem `omitempty` (distingue "não preenchido" de ausente).

### 5. `sdd/.metrics/schema.json` — adicionar propriedade

No objeto `properties`, adicionar:
```json
"complexidade": {
    "type": ["string", "null"],
    "description": "Complexidade da tarefa (reservado; preenchimento via ADR futura)"
}
```

Também sincronizar campos opcionais de feat-90-02 ausentes no template:
`tokens_cache_read`, `tokens_cache_creation`, `cost_usd_tabela`, `medicao`.

### 6. `internal/scaffold/templates/sdd/.metrics/schema.json.tmpl` — espelhar acima

Mesmas adições de `complexidade` + campos opcionais de feat-90-02.

## Regenerar goldens

Após atualizar os `.tmpl`, rodar:
```bash
go test ./internal/scaffold/... -run TestGolden -update
```

## Rules e IDs

`ARQ-01`, `ARQ-11`, `COD-04`, `COD-06`, `DUP-01`, `GOV-01`, `TST-01`, `TST-04`

## Critérios de Aceitação Executáveis

```bash
go build ./...
go test ./internal/monitor/... -run TestEstimateSessionUsage
go test ./internal/scaffold/... -run TestGolden
grep -r "tokens-input" .agents/skills/ --include="*.md" | wc -l
# deve retornar >= 7
```
