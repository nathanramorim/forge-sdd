# Feature 115-03-01 — Contratos de tier por task

## Modelo do módulo

Enxuto (`ARQ-11`): estender `internal/config` e `internal/bus`; sem pacote ou dependência nova.

## Decisões do projeto aplicadas

- ADR-0004: jobs trafegam pelo bus e são consumidos por workers.
- ADR-0005: CLI e argv são configurados por papel.
- ADR-0014 D-5: seleção estática por papel permanece como fallback.
- ADR-0024 (proposta): tier lógico por task, sem fallback entre provedores.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Tipos de job | `internal/bus/job.go` | Job sem complexidade/tier | Estender com campos opcionais | Preservar compatibilidade |
| Configuração de modelos | `internal/config/forge_profiles.go` | Modelo/esforço estático por papel | Adicionar mapa tier→modelo por CLI | Seleção validada no dono da config |

## Dependências novas

Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis

`ARQ-01`, `ARQ-11`, `COD-03`, `DUP-01`, `FLX-01`, `SEG-01`, `TST-01`, `TST-03`.

## Critérios de Aceitação Executáveis

1. Campos opcionais `complexity` e `model_tier` serializam e desserializam sem quebrar jobs legados.
2. Configuração valida tiers conhecidos, valores não vazios e nomes de modelo específicos do perfil sem expor segredos.
3. Critérios verificáveis:

```sh
go test ./internal/bus/... ./internal/config/...
go vet ./internal/bus/... ./internal/config/...
```
