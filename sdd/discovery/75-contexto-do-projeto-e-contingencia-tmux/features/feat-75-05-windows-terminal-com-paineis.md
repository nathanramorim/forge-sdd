# feat-75-05 — Windows Terminal com painéis

**Branch:** `feat/75-contexto-do-projeto-e-contingencia-tmux`
**Fase:** 75-05
**Tamanho:** `P`
**Depende de:** 75-04
**Status:** `pending` (FLX-01 aprovado)

## Contexto & Motivação
No Windows nativo não existe tmux e o Forge abre três janelas separadas. Com o Windows Terminal (`wt.exe`) dá para abrir as três sessões como painéis divididos de uma única janela, o equivalente mais próximo do tmux.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** novo ramo no montador de argv do launcher, sem pacote novo (`ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Terminais por SO | `internal/pipeline/launcher.go` (`terminalArgv`, `LaunchPlanFor`, `Launch`) | Windows: um `wt.exe new-tab` por sessão ou `cmd /c start` | estender | Com `wt.exe`, uma chamada com `new-tab` e dois `split-pane`. |
| Quoting de comandos | `launcher.go` (`quoteArg`, `winUnsafe`) | Recusa caracteres inseguros no Windows (inclui `;`) | reusar | O `Inner` de cada sessão segue validado; o `;` do `wt` é um argumento separado, fora do `Inner`. |
| Detecção de `wt.exe` | `launcher.go` (`has`) | `wt.exe` ou `wt` | reusar | A mesma checagem. |

## Dependências novas
- Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `SEG-01`, `COD-03`, `TST-04`.
- Rules: `sdd-seguranca.md`, `sdd-codigo-e-testes.md`.
- Gerais: `FLX-02`, `COD-02`, `TST-05`.

## Escopo
1. Em `GOOS=windows` com `wt.exe`, `LaunchPlanFor` devolve **um** comando: `wt.exe -w new new-tab --title forge-discovery cmd.exe /k <inner> ; split-pane -V --title forge-builder cmd.exe /k <inner> ; split-pane -H --title forge-reviewer cmd.exe /k <inner>`, com cada `;` como argumento próprio.
2. Sem `wt.exe`: `cmd /c start` por sessão, como hoje. Dentro do WSL vale o caminho Linux (tmux).
3. `up --dry-run` mostra o comando; o rótulo do menu diz "em painéis do Windows Terminal" quando for o caso.
4. Documentar que o comportamento em Windows real é **a confirmar** (só o argv é testado aqui).

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./internal/pipeline/... ./cmd/forge-sdd/...
```
- `TestLauncherWindows`: com `wt.exe` o argv é uma única chamada com `new-tab` e dois `split-pane` e os títulos das três sessões; sem `wt.exe`, `cmd /c start` por sessão; `Inner` com caractere inseguro continua recusado.
- `forge-sdd gate` sem falhas.
