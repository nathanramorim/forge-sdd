# feat-82-03 — Prompts, Regras e Documentação

**Branch:** `feat/82-unificacao-comando-forge`
**Fase:** 82-03
**Tamanho:** `M`
**Depende de:** 82-01
**Status:** `queued`

## Contexto & Motivação
Para que o ecossistema do Forge-SDD seja 100% coerente, os arquivos de contexto raiz (`GEMINI.md`, `CLAUDE.md`), as rules de governança (`.agents/rules/`), as skills operacionais (`.agents/skills/`), a documentação de referência (`docs/comandos.md`) e o `README.md` devem instruir consistentemente `forge <comando>` (ex: `forge brief`, `forge status`, `forge gate`, `forge up`, `forge session record`), eliminando a confusão de nomes entre agentes e desenvolvedores.

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** Modifica prompts e documentação transversal do repositório (`COD-05`, `GOV-01`).

## Decisões do projeto aplicadas
- `ADR-0005`: Porta de entrada única via `forge`.
- `COD-05`: Clareza e padronização em pt-BR.

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Contexto Gemini | `GEMINI.md` | Comandos `forge-sdd brief`, etc. | Modificar | Padronizar para `forge brief`, `forge status`, etc. |
| Contexto Claude | `CLAUDE.md` | Comandos `forge-sdd` | Modificar | Padronizar para `forge`. |
| Documentação CLI | `docs/comandos.md` | Lista de comandos `forge-sdd` | Modificar | Padronizar para `forge`. |
| Regras dos agentes | `.agents/rules/` | Referências a `forge-sdd` | Modificar | Padronizar para `forge`. |
| Skills operacionais | `.agents/skills/` | Comandos de apoio | Modificar | Padronizar para `forge`. |

## Dependências novas
- Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `COD-05`: Comunicação clara em pt-BR.
- `DUP-01`: Fonte única da verdade para nomenclatura de comandos.
- `FLX-05`: Governança e integridade dos prompts.
- `GOV-01`: Cumprimento estrito da Constitution.

## Critérios de Aceitação Executáveis
1. `GEMINI.md` e `CLAUDE.md` ensinam o ciclo com `forge <comando>` (ex: `forge brief`, `forge status`, `forge session record`).
2. `docs/comandos.md` e `README.md` documentam `forge` como comando primário.
3. `forge gate` e suíte de testes (`go test ./...`) passam 100% verdes com 0 violações de governança.
