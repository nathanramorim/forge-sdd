# Feature 113-03 — Higienizador do corpo da issue e gravador idempotente de `card-evidence.json`

## Objetivo
Montar título e corpo comercial da issue a partir da spec, sem segredos nem controle/ANSI, e gravar `card-evidence.json` de forma atômica e idempotente.

## Modelo do módulo
**Enxuto** (ARQ-11): funções em `internal/pipeline` (dono de `CardEvidence`, DUP-01) que reusam `gate.MaskSecrets` e a escrita segura já existente. Sem pacote novo.

## Decisões do projeto aplicadas
- **ADR-0002** (`adr-0002-layout-e-estado-por-iniciativa.md`, existe): a evidência fica em `sdd/discovery/<id>-<nome>/card-evidence.json`.
- **ADR-0018** (`adr-0018-gravacao-segura-de-arquivos-do-usuario.md`, existe): escrita atômica, confinada ao projeto, sem seguir symlink.
- **ADR-0017** (existe): dado externo é não confiável; só o princípio se aplica.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| `CardEvidence` | `internal/pipeline/dispatch.go:42-49`, `CardEvidenceFile` em `:29` | Campos `provider,id,url,status,feature,checked_at`; hoje só LIDA por `loadEvidence` (`:431`), nunca gravada | Estender | Acrescentar o gravador no mesmo pacote |
| `loadEvidence` (esquema) | `dispatch.go:431-497` | `DisallowUnknownFields`, teto 16 KB, rejeita `\x00\r\n\x1b`, provider em `pm_mcp.supported`, URL http(s), `checked_at` RFC3339 | Reusar | O gravador deve produzir o que o leitor aceita; validar na borda com as mesmas regras |
| `gate.MaskSecrets` | `internal/gate/gate.go:58` | Existe | Reusar | SEG-05 |
| Remoção de ANSI/controle | `internal/notify/notify.go` (`Sanitize`, citado no criteria) | NÃO verifiquei com grep neste split | Reusar se existir | O Builder confirma; senão cria helper em `internal/gate` |
| Escrita segura | `internal/scaffold/hooks.go:262` (`WriteFileSafe`), `internal/queue/atomic.go:55` (`WriteJSONAtomic`), `internal/gate/context.go:115` (`ConfineToDir`) | Existem. Divergência: o criteria cita `hooks.go:235`; a linha real é 262 | Reusar | SEG-02; cuidado com ciclo de importação (ARQ-02): escolher o pacote que `pipeline` já importa |
| Extração de seção da spec | grep em `internal/` | Não verificado | Criar | Extrair título e `## Descrição de negócio` |

## Dependências novas
Nenhuma (SEG-07, ARQ-04): stdlib Go.

## Rules e IDs da Constituição Aplicáveis
`SEG-01`, `SEG-02`, `SEG-05`, `SEG-08`, `DUP-03`, `TST-03`, `TST-04`.

## Critérios de Aceitação Executáveis
Testes com `t.TempDir()`, sem rede e sem `gh` (TST-04).
```bash
set -o pipefail
go build ./... && go vet ./...
go test ./internal/pipeline -run 'TestCorpoIssueExtraiDescricaoDeNegocio|TestCorpoIssueMascaraSegredosEAnsi|TestCorpoIssueSemCaminhoAbsoluto|TestCorpoIssueSemDescricaoDeNegocioFalha|TestGravaCardEvidenceAtomico|TestGravaCardEvidenceIdempotenteReaproveitaExistente|TestGravaCardEvidenceRecusaSymlinkEFora|TestGravaCardEvidenceCompativelComLoadEvidence'
go test ./internal/pipeline ./internal/gate

# Os testes citados existem de verdade (sem isso o -run passa vazio):
test "$(go test ./internal/pipeline -list 'TestCorpoIssueExtraiDescricaoDeNegocio|TestCorpoIssueMascaraSegredosEAnsi|TestCorpoIssueSemCaminhoAbsoluto|TestCorpoIssueSemDescricaoDeNegocioFalha|TestGravaCardEvidenceAtomico|TestGravaCardEvidenceIdempotenteReaproveitaExistente|TestGravaCardEvidenceRecusaSymlinkEFora|TestGravaCardEvidenceCompativelComLoadEvidence' | grep -c '^Test')" -ge 8   # pelo menos 8 (go test -run sem correspondência sai com 0)
```
Casos: corpo só com título, descrição de negócio e link relativo; token/segredo mascarado, ANSI e controle removidos (SEG-05/08); caminho absoluto nunca aparece; spec sem a seção falha com erro claro (negativo); symlink ou caminho fora do projeto recusado (SEG-02); evidência existente é reaproveitada sem reescrever (RN-4); o arquivo gravado passa em `loadEvidence`.

Casos adicionais (observações do Reviewer da 113-01 e desvios do código real):
- `vcs.ErrSemPR` de `CriarIssue`/`ProcurarIssueAberta` (`nenhum`) é "sem efeito, sem erro": nada criado, nada gravado, `Feito=false` com `Motivo`; difere de `ProcurarPRAberto`, que devolve `(nil, nil)` (`TestCriarIssueConcepcaoSemPRSemEfeito`).
- `ProcurarIssueAberta` vê só 30 itens do `gh issue list` e filtra no cliente: issue antiga fora da janela pode ser duplicada (falha segura); o título-base é o `# ` da spec, estável; follow-up proposto: `--limit` maior em `internal/vcs`.
- Desvio: o provider gravado é `github-projects` (o `github` não está em `pm_mcp.supported`, e o `loadEvidence` o recusaria); `Status` inicial `Spec`; `Feature` é o nome da pasta da iniciativa.
- Desvio: sem `WriteFileSafe` (vive em `scaffold`, ciclo de import): usa `queue.WriteJSONAtomic` + `gate.ConfineToDir` na pasta e `Lstat` no alvo. Código em `internal/pipeline/corpo_issue.go` (`MontaCorpoIssue`, `GravaCardEvidence`, `CriarIssueConcepcao`).

## Depende de
113-01, 113-02.
