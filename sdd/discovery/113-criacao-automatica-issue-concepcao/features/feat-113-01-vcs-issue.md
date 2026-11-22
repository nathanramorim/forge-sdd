# Feature 113-01 — Operações de Issue em `internal/vcs`

## Objetivo
Dar ao adaptador VCS as operações `CriarIssue` e `ProcurarIssueAberta`, para criar e localizar issues pelo `gh` sem duplicar, com `nenhum` como no-op sem chamada externa.

## Modelo do módulo
**Enxuto** (ARQ-11): só estende o pacote existente `internal/vcs` (dono das chamadas ao `gh`, DUP-01). Sem pacote, estado ou daemon novo.

## Decisões do projeto aplicadas
- Nenhuma ADR vigente governa esta task de forma direta; a camada `internal/vcs` vem da 106-02 e é reutilizada sem mudar os métodos de PR. ADR-0022/0023 tratam de PR em draft e estado isolado e não se aplicam: nenhuma operação de PR é alterada.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Métodos de Issue no adaptador | `internal/vcs/*.go` (grep `CriarIssue\|ProcurarIssue`) | Nenhum (só `CriarPRDraft`, `ProcurarPRAberto`, `CorrigirBase`, `MarcarPronto`, `Mesclar` em `github.go:15-67`) | Criar | Plano confirmado: não há método de Issue |
| Runner injetável e execução segura | `internal/vcs/vcs.go:19` (`Runner`), `vcs.go:99` (`RunGitOrGh`) | `type Runner func(dir, name string, args ...string) (string, error)`; `RunGitOrGh` só `git`/`gh` | Reusar | SEG-03, DUP-01 |
| Erro do VCS `nenhum` | `internal/vcs/vcs.go:22`, `nenhum.go` | `ErrSemPR`; `nenhum` devolve `ErrSemPR` nas operações de PR | Reusar | O plano manda `nenhum` devolver `ErrSemPR` (no-op sem chamada) |
| Interface `Adaptador` | `internal/vcs/vcs.go:45` | Interface com os 5 métodos de PR | Estender | Acrescentar 2 métodos exige atualizar todos os implementadores e fakes (checado por `go build`/`go vet`) |

## Dependências novas
Nenhuma (SEG-07, ARQ-04): stdlib Go e o `gh` já previsto em `AllowCommands`.

## Rules e IDs da Constituição Aplicáveis
`ARQ-01`, `ARQ-02`, `DUP-01`, `SEG-01`, `SEG-03`, `TST-03`, `TST-04`.

## Critérios de Aceitação Executáveis
Os testes usam `Runner` falso injetado; NUNCA chamam o `gh` real nem a rede (TST-04).
```bash
set -o pipefail
go build ./... && go vet ./...
go test ./internal/vcs -run 'TestGitHubCriarIssueComDescricaoComercial|TestGitHubProcurarIssueAbertaIdempotente|TestNenhumCriarIssueRetornaSemPR|TestGitHubCriarIssueErroDoGh|TestGitHubCriarIssueRecusaTituloVazio|TestGitHubProcurarIssueSaidaJSONInvalida'
go test ./internal/vcs

# Os testes citados existem de verdade (sem isso o -run passa vazio):
test "$(go test ./internal/vcs -list 'TestGitHubCriarIssueComDescricaoComercial|TestGitHubProcurarIssueAbertaIdempotente|TestNenhumCriarIssueRetornaSemPR|TestGitHubCriarIssueErroDoGh|TestGitHubCriarIssueRecusaTituloVazio|TestGitHubProcurarIssueSaidaJSONInvalida' | grep -c '^Test')" -ge 6   # pelo menos 6 (go test -run sem correspondência sai com 0)
```
Casos: sucesso (argv `gh issue create --title ... --body ...`, `Dir` fixo, sem shell); idempotência (`gh issue list --state open --search ... --json number,url,title` com título base ou `] <base>` devolve a issue existente sem criar); `nenhum` devolve `ErrSemPR` sem chamar o runner; erro do `gh` propagado; título vazio recusado (negativo); JSON inválido ou campo fora do esquema recusado (SEG-01). Os testes de PR existentes continuam passando.

## Depende de
Nenhuma (106-02 já na `main`).
