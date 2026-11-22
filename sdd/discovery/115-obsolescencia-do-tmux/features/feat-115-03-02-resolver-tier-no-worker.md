# Feature 115-03-02 — Resolver tier no despacho e worker

## Modelo do módulo

Enxuto (`ARQ-11`): estender `internal/pipeline`, `internal/config` e o contrato do bus já preparado pela 115-03-01.

## Decisões do projeto aplicadas

- ADR-0004: worker executa job reivindicado.
- ADR-0005: cada CLI usa seus próprios argumentos.
- ADR-0014 D-5: seleção estática por papel é fallback.
- ADR-0024 (proposta): tier de task precede configuração estática; Reviewer só recebe após Camada A/PR.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Dispatch | `internal/pipeline/dispatch.go` | valida gates antes de criar job | Copiar metadados aprovados da task | Não perder decisão do split |
| Args do worker | `internal/pipeline/worker.go` | modelo resolvido por papel | Resolver tier na CLI ativa antes do fallback | Modelo certo por task |
| Gate A e Reviewer | `internal/pipeline/pr.go`, `worker.go` | handoff condicionado a validação/PR | Preservar e testar | Não revisar código não validado |

## Dependências novas

Nenhuma (`SEG-07`). Depende de feat-115-03-01 mesclada.

## Rules e IDs da Constituição Aplicáveis

`ARQ-01`, `COD-03`, `DUP-01`, `FLX-01`, `FLX-03`, `SEG-01`, `SEG-03`, `TST-03`, `TST-05`.

## Critérios de Aceitação Executáveis

1. Dispatch copia complexity/tier para o job; job sem tier usa o fallback estático.
2. Testes de argv cobrem modelos configurados por CLI, flags suportadas e ausência de fallback cruzado.
3. Reviewer não é enfileirado antes de Camada A e PR registrado.
4. Critérios verificáveis:

```sh
go test ./internal/config/... ./internal/pipeline/...
go vet ./...
```
