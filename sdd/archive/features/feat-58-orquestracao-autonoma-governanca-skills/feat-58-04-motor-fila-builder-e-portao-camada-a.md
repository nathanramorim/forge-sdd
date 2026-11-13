# feat-58-04 — Motor Determinístico da Fila (`forge-sdd queue`) e Camada A do Portão (`forge-sdd gate`)

**Branch:** `feat/58-orquestracao-autonoma-governanca-skills`
**Fase:** 58-04
**Depende de:** `feat-58-01`
**Status:** `done`

## Objetivo

Implementar no binário Go (`forge-sdd`) o gerenciador determinístico da fila do Builder (`internal/queue` + subcomando `forge-sdd queue`) persistido em `sdd/.queue/queue.json`, e o motor da **Camada A do Portão de Governança** (`internal/gate` + subcomando `forge-sdd gate`), automatizando as verificações determinísticas `P-xx` com suporte aos modos `observacao` e `bloqueio`.

## Modelo do módulo (`ARQ-11`)
- **Modelo:** `completo` (pacotes de domínio/aplicação `internal/queue` e `internal/gate` + subcomandos Cobra em `cmd/forge-sdd/`).
- **Justificativa:** Contém regras de negócio de transição de estados da fila (`queued → building → review_pending → approved | changes_requested`) e regras de validação determinística do portão (`P-01..P-23`).

## O que já existe (`DUP-02`)

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
| --- | --- | --- | --- | --- |
| Diagnóstico do projeto | `internal/doctor/` (`forge-sdd doctor`) | Valida estrutura de pastas SDD, naming convention e resumo de telemetria | `reusar` / `estender` | Reaproveitar helpers de leitura de `.sddrc` e `progress.md`, mantendo `doctor` para saúde da instalação e criando `internal/gate` focado nos checks `P-xx` de PR/task |
| Especificação da Camada A | `busize-solutions/planning-brain-governanca` (`portao/portao-jornada-constitution.md` §4 Camada A) | Mapa de verificações `P-01` a `P-24` com modos `observacao` e `bloqueio` | `reusar` (genérico) | Implementar em Go stdlib as verificações universais do Forge-SDD (`P-01`, `P-02`, `P-03`, `P-13`, `P-14`, `P-15`, `P-17`, `P-19`, `P-20`, `P-21`, `P-23`) |
| Gerenciador de Fila do Builder | `internal/` | Não existe pacote de fila no repositório | `criar` | Criar `internal/queue` para persistir e validar `sdd/.queue/queue.json` |

## Dependências novas (`SEG-07`)
Nenhuma (apenas Go `stdlib` + `spf13/cobra` e `stretchr/testify` já existentes).

## Rules e IDs da Constituição Aplicáveis
- **Rules:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-codigo-e-testes.md`, `.agents/rules/sdd-fluxo-e-portao.md`
- **IDs:** `ARQ-03`, `DUP-02`, `SEG-04`, `SEG-07`, `COD-02`, `TST-01..05`, `FLX-01..06`, `GOV-01..04`.

## Critérios de Aceitação Executáveis

1. **Subcomando `forge-sdd queue` (`internal/queue`):**
   - `forge-sdd queue enqueue --feature <id> --task <path> --rules <r1,r2>`: valida que a Spec existe e possui a seção `"O que já existe"` preenchida (`P-02`, `P-03`), adicionando o item em `sdd/.queue/queue.json` com status `queued`.
   - `forge-sdd queue next`: retorna (em texto ou `--json`) a próxima task `queued` com suas `rules` e `skills`, mudando o status para `building`.
   - `forge-sdd queue complete <task-id>`: muda o status de `building` para `review_pending`.
   - `forge-sdd queue list`: lista todas as tasks da fila e seus estados atuais.
2. **Subcomando `forge-sdd gate` (`internal/gate` — Camada A):**
   - Executa as verificações determinísticas:
     - `P-01`: branch atual não é `main` e segue `<prefixo>/<id>-*`;
     - `P-02` e `P-03`: spec da feature/task existe e tem a seção `## O que já existe` preenchida com tabela não-vazia;
     - `P-13`: ausência de padrões de segredos/tokens privados nos arquivos rastreados;
     - `P-14`: se `go.mod` foi alterado em relação à base, a spec contém seção `## Dependências novas` justificada;
     - `P-15` (opcional via `--run-tests`): `go vet ./...` e `go test ./...` passam;
     - `P-20`: qualquer bloco `constitution-exception` encontrado possui `regra`, `motivo`, `prazo` (data futura válida `AAAA-MM-DD`) e `responsavel`;
     - `P-21`: `sdd/memory/progress.md` ≤ 1024 bytes;
     - `P-23`: todos os arquivos `.agents/rules/*.md` com frontmatter `regras: [...]` citam IDs que existem em `sdd/memory/constitution.md`.
   - Suporta `--mode observacao|bloqueio` (em `observacao` imprime achados e retorna exit code `0`; em `bloqueio` retorna exit code `!= 0` se houver falha).
3. Testes unitários em `internal/queue` e `internal/gate` cobrindo casos conformes e violações de cada check; `go test ./...` e `go vet ./...` passando 100%.

## Handoff
Depende de `feat-58-01`. Pode rodar em paralelo com `feat-58-06`. Desbloqueia `feat-58-05`.
