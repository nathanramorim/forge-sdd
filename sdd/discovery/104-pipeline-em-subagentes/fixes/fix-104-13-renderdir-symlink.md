# fix-104-13 — renderDir segue link simbólico e grava fora do projeto (SEG-02)

**Branch:** `fix/104-renderdir-symlink`
**Origem:** Iniciativa 104, apontado pela Governança na feat-104-04; falha anterior à 104
**Tamanho:** `S`
**Status:** `review`

## Problema

`renderDir` (`internal/scaffold/scaffold.go`) grava os templates com `os.MkdirAll` + `os.WriteFile`. Com `.claude` (ou qualquer pai) como link simbólico para fora do projeto, `forge init --agent claude` grava no destino externo. A escrita também não é atômica e segue link simbólico no próprio arquivo, FIFO trava e link pendente cria o alvo.

## Evidência

Reproduzida pela Governança: `.claude -> /fora`, `forge init --agent claude` cria `/fora/agents/*` e `/fora/commands/*`. Os testes `TestRenderDir*` em `cmd/forge-sdd/renderdir_seg_test.go` reprovam sem a correção (7 de 8).

## Feature Correlacionada

- Origem: `internal/scaffold/scaffold.go` (`renderDir`, `UpdateSddrc`, `migrateLegacyAgentDir`), commits de maio e junho.
- Causa-raiz: dado externo (estado do disco do usuário) usado como destino sem validação na borda; `checkSafeTarget`/`writeFileSafe` (feat-104-04) só cobriam os hooks do Claude.

## Decisões do projeto aplicadas

- SEG-02: recusar destino simbólico, fora do projeto ou não regular.
- DUP-01: reuso de `checkSafeTarget`/`writeFileSafe`; `writeFileSafeMode` só parametriza o modo novo (0755 dos `.sh`).

## Correção

1. `renderDir`, `UpdateSddrc` e a mescla de `.agent` legado gravam via `writeFileSafe[Mode]` (temporário exclusivo + rename, modo preservado).
2. `checkSafeTarget` passa a validar todo componente existente entre o projeto e o destino (pai do pai simbólico) e roda antes do `MkdirAll`.
3. Recusa vira aviso em stderr; os demais arquivos seguem. `--dry-run` não grava; `shouldPreserve` inalterado.
4. Fora de escopo: `copyTree`/`ApplyLayoutMigration` (já `O_EXCL`, recusam não regulares), `cmd/forge-sdd/pipeline.go` (`.forge/`, só `Lstat` no folha) e `lessons.go`/`governance.go`/`autopilot.go` (gravam arquivos do fluxo do usuário, não templates); registrados para fix próprio.

## Critérios de Aceitação

```bash
set -o pipefail
go test -list 'TestRenderDir' ./cmd/forge-sdd | grep -F 'TestRenderDirInitClaudeSymlinkParaForaNaoGrava' >/dev/null
go test ./cmd/forge-sdd -run 'TestRenderDirCasoNormalInstalaTudoEPreservaModo' -count=1
# negativo: .claude simbólico para fora (init e update), pai do pai, arquivo simbólico, link pendente, FIFO e diretório no lugar do arquivo não gravam fora e o resto instala
go test -list 'TestRenderDirInitClaudeSymlinkParaForaNaoGrava|TestRenderDirUpdateClaudeSymlinkParaForaNaoGrava|TestRenderDirPaiDoPaiSymlinkNaoCria|TestRenderDirArquivoSymlinkNaoSegue|TestRenderDirSymlinkPendenteNaoCriaAlvo|TestRenderDirFIFONaoTravaNemSubstitui|TestRenderDirDiretorioNoLugarDoArquivoRecusado' ./cmd/forge-sdd | grep -F 'TestRenderDirFIFONaoTravaNemSubstitui' >/dev/null
go test ./cmd/forge-sdd -run 'TestRenderDir' -count=1
go vet ./... && go test ./internal/scaffold/ -count=1
```
