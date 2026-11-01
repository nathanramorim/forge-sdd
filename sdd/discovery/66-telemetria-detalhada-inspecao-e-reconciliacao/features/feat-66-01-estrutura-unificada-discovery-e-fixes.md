# feat-66-01 — Estrutura Unificada de Discovery em Pastas e Fixes Correlacionados no Forge-SDD

**Branch:** `feat/66-telemetria-detalhada-inspecao-e-reconciliacao`
**Fase:** 66-01
**Tamanho:** `M`
**Depende de:** Nenhuma
**Status:** `ready`

## Contexto & Motivação

Historicamente, o comando `/discovery` gerava 3 arquivos soltos diretamente na raiz de `sdd/discovery/` (`discovery-XX.md`, `criteria-XX.md`, `plan-XX.md`), enquanto `/split-features` criava pastas em `sdd/features/` e `/novo-fix` gerava correções avulsas em `sdd/features/fix-*.md`. Isso causava dispersão e perda de rastreabilidade.

Esta subfeature estabelece o novo padrão arquitetural unificado no próprio `forge-sdd`:
1. Cada iniciativa de Discovery vive em sua pasta autocontida `sdd/discovery/<id>-<resume>/`:
   - `discovery.md` (ou `discovery-<id>-<resume>.md`)
   - `criteria.md`
   - `plan.md`
   - `features/` (onde o Planejador fatia as subfeatures para o Builder)
   - `fixes/` (onde correções correlacionadas a esta iniciativa são gravadas)
2. Atualização do comando `/novo-fix`: exige correlação com a feature original e gera o fix dentro da pasta correspondente (`sdd/discovery/<id>-<resume>/fixes/fix-*.md`).
3. Atualização no código Go (`internal/config/config.go`, `cmd/forge-sdd/doctor.go`): suporte a `ClassifyNamingConvention` e `DetectNamingConvention` reconhecendo a nova árvore de pastas.
4. Atualização dos templates de scaffolding (`internal/scaffold/templates/`) e comandos em `.agents/commands/`.

## Modelo do módulo

- **Modelo:** `enxuto` em `internal/config`, `cmd/forge-sdd`, `internal/scaffold` e `.agents/commands`.
- **Justificativa:** Trata-se de padronização estrutural de diretórios e convenções de nomenclatura sem introduzir dependências externas (`ARQ-03`, `ARQ-04`, `ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
|---|---|---|---|---|
| Reconhecimento de Convenção | `internal/config/config.go` | `ClassifyNamingConvention`, `DetectNamingConvention` | `estender` | Reconhecer `discovery.md`, `criteria.md`, `plan.md` quando em subpastas de discovery e validar nomes de pastas. |
| Comando Doctor | `cmd/forge-sdd/doctor.go` | `scanConventionDir` | `estender` | Varrer subpastas `sdd/discovery/` e `features/` / `fixes/` recursivamente. |
| Templates de Scaffolding | `internal/scaffold/templates/` | `discovery.md.tmpl`, `split-features.md.tmpl`, `novo-fix.md.tmpl` | `estender` | Atualizar os templates canônicos com a nova estrutura em pastas. |
| Comandos de Agentes | `.agents/commands/` | `discovery.md`, `split-features.md`, `novo-fix.md` | `estender` | Padronizar as instruções dos agentes para operar na nova estrutura unificada. |

## Dependências novas

- **Nenhuma dependência externa nova.** (`SEG-07`, `ARQ-04`).

## Regras Constitucionais e IDs Aplicáveis

- **Rules:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-governanca.md`, `.agents/rules/sdd-codigo-e-testes.md`.
- **IDs:** `ARQ-01`, `ARQ-03`, `ARQ-04`, `ARQ-06`, `ARQ-11`, `DUP-02`, `DUP-05`, `SEG-01`, `SEG-02`, `FLX-01`, `FLX-02`.

## Critérios de Aceite Gherkin

```gherkin
Cenário: Reconhecimento da nova estrutura de pastas pelo doctor
  Dado um projeto com iniciativa em sdd/discovery/66-telemetria-detalhada/
  E arquivos discovery.md, criteria.md e plan.md dentro da subpasta
  Quando o comando forge-sdd doctor for executado
  Então a convenção deve ser identificada como consistente
  E nenhum falso-positivo de deriva de nomenclatura deve ser reportado

Cenário: Novo fix correlacionado dentro da pasta da feature
  Dado uma demanda de correção para a iniciativa 66
  Quando o comando novo-fix for disparado
  Então o arquivo deve ser criado em sdd/discovery/66-telemetria-detalhada/fixes/fix-66-01-correcao.md
  E a seção "## Feature Correlacionada" deve apontar para o arquivo de origem

Cenário: Templates de scaffolding atualizados
  Dado a execução do teste de scaffolding do forge-sdd
  Quando os templates forem processados
  Então os comandos discovery, split-features e novo-fix devem conter as instruções da nova estrutura unificada
```

## Plano de Implementação

1. Estender `internal/config/config.go` e `cmd/forge-sdd/doctor.go` para suportar `discovery.md`, `criteria.md`, `plan.md` em subpastas.
2. Atualizar `.agents/commands/discovery.md`, `split-features.md`, `novo-fix.md`, `nova-feature.md` e `.gemini/prompts/`.
3. Atualizar templates correspondentes em `internal/scaffold/templates/`.
4. Executar testes de unidade e validação com `go test -race ./...`.

## Comandos de Verificação Determinísticos

```bash
go test -v -race ./cmd/forge-sdd/... ./internal/config/... ./internal/scaffold/... && go vet ./...
```
