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
4. `checkSafeTarget` e `writeFileSafe*` tornam `targetDir` e o destino absolutos (`filepath.Abs`) antes de comparar com o `EvalSymlinks` do projeto: `init .`, `./x`, `a/../b`, `/var` e `/tmp` simbólicos e caminhos com espaço usam a mesma base.
5. Se houver recusa e o projeto ficar sem `sdd/.sddrc` (`Run`), ou nada for instalado (`RunAgents`), o comando sai com código diferente de 0 explicando o motivo. Recusa pontual continua aviso com saída 0.
6. Avisos passam por `notify.Sanitize` (SEG-05).
7. `pipeline init` (`.forge/`) grava via `scaffold.WriteFileSafe` (exportação fina de `writeFileSafe`).
8. Fora de escopo: `copyTree`/`ApplyLayoutMigration` (já `O_EXCL`, recusam não regulares) e `lessons.go`/`governance.go`/`autopilot.go` (arquivos do fluxo do usuário, não templates).

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Checagem de destino seguro | `internal/scaffold/hooks.go` `checkSafeTarget` | Só pai imediato, caminho relativo quebrava | Estender (ancestrais, `Abs`) |
| Gravação atômica | `writeFileSafe` | Modo novo fixo 0644 | Reusar |
| Modo novo parametrizado | `writeFileSafeMode` | Não existia | Estender (`writeFileSafe` o chama) |
| Preservar arquivos do usuário | `shouldPreserve` | Correto, usa `os.Stat` | Reusar, inalterado |
| Mescla de `.agent` legado | `migrateLegacyAgentDir` | `MkdirAll`+`WriteFile` e leitura seguindo link | Estender (usa `writeFileSafe`) |
| Cópia de backup | `copyTree` | `O_EXCL`, recusa não regular | Reusar, sem mudança |
| Migração de layout | `ApplyLayoutMigration` | `Lstat` no destino e `Rename` | Reusar, sem mudança |

## Decisões da implementação

- `.claude` (ou outro pai) simbólico apontando para DENTRO do próprio projeto passa a ser recusado; a base gravava pelo link. É o comportamento da spec, mas é mudança de comportamento.
- `pipeline init` com `.forge` simbólico falha com erro em vez de gravar fora.

## Critérios de Aceitação Executáveis

```bash
set -o pipefail
go test -list 'TestRenderDir' ./cmd/forge-sdd | grep -F 'TestRenderDirInitClaudeSymlinkParaForaNaoGrava' >/dev/null
go test ./cmd/forge-sdd -run 'TestRenderDirCasoNormalInstalaTudoEPreservaModo' -count=1
# negativo: .claude simbólico para fora (init e update), pai do pai, arquivo simbólico, link pendente, FIFO e diretório no lugar do arquivo não gravam fora e o resto instala
go test -list 'TestRenderDirInitClaudeSymlinkParaForaNaoGrava|TestRenderDirUpdateClaudeSymlinkParaForaNaoGrava|TestRenderDirPaiDoPaiSymlinkNaoCria|TestRenderDirArquivoSymlinkNaoSegue|TestRenderDirSymlinkPendenteNaoCriaAlvo|TestRenderDirFIFONaoTravaNemSubstitui|TestRenderDirDiretorioNoLugarDoArquivoRecusado' ./cmd/forge-sdd | grep -F 'TestRenderDirFIFONaoTravaNemSubstitui' >/dev/null
go test ./cmd/forge-sdd -run 'TestRenderDirInitClaudeSymlinkParaForaNaoGrava|TestRenderDirUpdateClaudeSymlinkParaForaNaoGrava|TestRenderDirPaiDoPaiSymlinkNaoCria|TestRenderDirArquivoSymlinkNaoSegue|TestRenderDirSymlinkPendenteNaoCriaAlvo|TestRenderDirFIFONaoTravaNemSubstitui|TestRenderDirDiretorioNoLugarDoArquivoRecusado' -count=1
# negativo: `init .` instala árvore idêntica ao caminho absoluto, recusa só o simbólico e sai não zero quando tudo é recusado
go test -list 'TestInitPontoIdenticoAoCaminhoAbsoluto|TestInitPontoClaudeSimbolicoRecusaSoOQueDeve|TestInitTodosRecusadosSaiNaoZero|TestPipelineInitForgeSimbolicoNaoGrava|TestAvisoSanitizaCaminhoComEscEBidi|TestUpdatePontoIdenticoAoCaminhoAbsoluto' ./cmd/forge-sdd | grep -F 'TestInitTodosRecusadosSaiNaoZero' >/dev/null
go test ./cmd/forge-sdd -run 'TestInitPontoIdenticoAoCaminhoAbsoluto|TestInitPontoClaudeSimbolicoRecusaSoOQueDeve|TestInitTodosRecusadosSaiNaoZero|TestPipelineInitForgeSimbolicoNaoGrava|TestAvisoSanitizaCaminhoComEscEBidi|TestUpdatePontoIdenticoAoCaminhoAbsoluto' -count=1
go vet ./... && go test ./internal/scaffold/ -count=1
```
