# feat-105-07 — Registro CSV dos comandos depreciados (vale para o forge inteiro)

**Iniciativa:** 105-experiencia-cli-e-harness (onda 1, task 1.7)
**Branch:** feat/105-experiencia-cli-e-harness (única; `FLX-02`)
**Depende de:** feat-105-03 (PR #126, branch `feat/105-03-avisar-depreciados`; o Builder parte dela, pois `cmd/forge-sdd/deprecated.go` só existe lá)
**Paralelo com:** nenhuma
**Sequenciar com:** feat-105-06 (sem arquivos em comum). 106/107: se mexerem em `queue.go`/`attach.go`, rebase e série.

## Problema
O dono depreciou comandos e precisa saber quais seguem em uso, por quem, para quê e o que os substitui, para decidir depois se a depreciação faz sentido. Hoje o aviso da 1.3 só imprime no stderr e some. Pedido literal: "gerar uma lista dos mapas de comandos depreciados (esse usou, pra que, qual agente usou, qual substitui); pode ser um arquivo csv; sempre que identificar preencher; vale pro forge inteiro".

## Comportamento
1. **Ponto único.** O helper da 1.3 (`avisoDepreciado`) passa a consultar um **catálogo** de depreciados (comando, substituto, "desde", ADR). Ele imprime o aviso (igual à 1.3) E acrescenta uma linha ao CSV. Depreciar um comando novo no futuro = registrar uma entrada no catálogo e ligar o helper no `PreRun`; nenhum código de log é duplicado. O catálogo nasce com `queue stations`, `queue handoff` e `attach` (desde a ADR-0021).
2. **Arquivo:** `sdd/memory/logs/depreciados.csv` do projeto (`--dir`, ou `.`). Não há `sdd/memory/logs/` hoje (`ls sdd/memory` só tem `clarify.md`, `constitution.md`, `lessons.md`, `mcps.md`, `naming-convention.md`, `progress-log.md`); o pedido do dono ("memory/logs/") é interpretado como `sdd/memory/logs/` e a pasta é criada na primeira gravação. Só grava se `sdd/` existir no alvo (não cria árvore em pasta qualquer).
3. **Colunas** (cabeçalho na primeira linha, UTF-8, `,` e aspas CSV padrão; abre em planilha):
   | Coluna | Conteúdo | Justificativa |
   |---|---|---|
   | `data_utc` | RFC 3339 UTC | ordenar e filtrar |
   | `comando` | caminho do comando depreciado (`queue stations`) | "esse usou" |
   | `contexto` | só o nome do subcomando e das flags usadas (`--dir`, `--from`), NUNCA valores nem argumentos posicionais | "pra quê" sem vazar dado (`SEG-01`, `SEG-04`) |
   | `agente` | `claude-code` se `CLAUDECODE` estiver definida; senão `desconhecido` | "qual agente usou" |
   | `substituto` | o que usar no lugar (do catálogo) | pedido do dono |
   | `desde` | ADR/versão da depreciação (`ADR-0021`) | rastrear |
   | `versao_forge` | versão do binário | correlacionar com releases |
4. **Detecção do agente:** por variável de ambiente, em tabela única e testável. `CLAUDECODE` é a variável que o Claude Code define (a verificar no Builder com `env` dentro de uma sessão; não provada aqui). **A verificar:** variáveis de Gemini CLI, Copilot e Codex (não confirmadas; até provar, esses agentes caem em `desconhecido` e a tabela traz apenas as entradas comprovadas). Nunca grava o valor da variável, só o rótulo fixo.
5. **Robustez:**
   - Escrita só por `scaffold.WriteFileSafe`/`checkSafeTarget` (`SEG-02`): recusa link simbólico e destino fora do projeto; gravação atômica (arquivo temporário + rename; ler, acrescentar, regravar).
   - Tamanho máximo **256 KiB** (proposta; ~2 mil linhas): ao exceder, não grava mais e avisa uma vez no stderr; sem rotação nem apagamento (não perde dado).
   - Cada célula passa por `notify.Sanitize` (sem quebra de linha nem ANSI) e, se iniciar com `=`, `+`, `-` ou `@`, ganha prefixo `'` (injeção de fórmula em planilha).
   - Falha de escrita NUNCA quebra o comando: o aviso e a execução seguem; só sai uma linha de aviso no stderr (`COD-03`: erro tratado e enriquecido, não descartado).
   - `--dry-run` não grava o CSV (`ARQ-05`).
   - Nenhum dado pessoal: sem usuário, host, caminho absoluto, e-mail nem valores de argumento.
   - Concorrência: dois processos simultâneos podem perder uma linha (ler+rename). Aceito e documentado; sem lock nesta task.
6. **Aliases de texto** (`/implement`, `/proxima-feature` em `.claude/commands/`): não passam pelo Go, logo **não geram linha de uso**. Entram só no catálogo documentado (comando, substituto `/implementar`, desde ADR-0021) para a revisão do dono. **A verificar:** se existe hook de comando de barra que permita registrar uso; sem prova, fica fora (registrar como limitação).
7. **Revisão pelo dono:** o CSV abre em planilha. Subcomando de leitura (`forge deprecated`) fica **fora de escopo** (custo > valor agora; `ARQ-06` pede decisão própria para um novo comando).

## Modelo do módulo
**Novo** `internal/deprecation`: catálogo, detecção do agente, montagem e saneamento da linha, append com limite. `cmd/forge-sdd/deprecated.go` só liga o `PreRun` e passa `--dir`/`--dry-run` (`ARQ-01`, `ARQ-10`). O módulo importa `internal/scaffold` só para `WriteFileSafe` (a verificar no Builder se isso respeita o sentido das dependências; senão, mover o helper seguro para pacote comum, sem duplicar).

## Decisões do projeto aplicadas
ADR-0021 (obsolescência, da 1.1), ADR-0002 (estado por iniciativa/`sdd/`), ADR-0010 (comando único `forge`), ADR-0004.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Helper de aviso | `cmd/forge-sdd/deprecated.go` (branch da 1.3): `avisoDepreciado(usar)` | Só imprime no stderr | **Estender** | `DUP-01` |
| Testes do aviso | `cmd/forge-sdd/deprecated_test.go` (1.3): `runSep`, `TestComandosDepreciadosAvisam` | Execução cobra separando stdout/stderr | **Reusar** | `TST-03` |
| Escrita segura | `internal/scaffold/hooks.go`: `checkSafeTarget`, `writeFileSafe`, `WriteFileSafe` exportada | Atômica, recusa link simbólico | **Reusar** | `SEG-02` |
| Saneamento | `internal/notify/notify.go:122` `Sanitize(s, max, keepNewlines)` | Remove controle/ANSI | **Reusar** | `DUP-01` |
| Pasta de logs | `sdd/memory/` | Sem `logs/`; sem convenção | **Criar** `sdd/memory/logs/` | pedido do dono |
| Detecção de agente | busca de `CLAUDECODE` em `*.go`/`*.md` | Nada | **Criar** tabela | a verificar |
| CSV | `encoding/csv` (stdlib) | Disponível | **Reusar** | `SEG-07` |

## Dependências novas
Nenhuma (`SEG-07`); só stdlib (`encoding/csv`).

## Riscos
- **Arquivo versionado:** cada uso modifica um arquivo rastreado: ruído em commits e conflito de merge entre PRs paralelos. Mitigações (sem decidir): (a) ignorar no git (`.gitignore`), perde histórico compartilhado; (b) versionar com `.gitattributes` `merge=union` (append-only costuma mesclar sem conflito); (c) um arquivo só, append-only, sem reescrever linhas. Posição pedida ao dono em "A confirmar" do `status.md`.
- Variáveis de agente erradas geram `desconhecido` (degradação segura).
- Perda de linha em concorrência (item 5).

## Rules e IDs da Constituição Aplicáveis
`ARQ-01`, `ARQ-05`, `ARQ-10`, `COD-03`, `COD-05`, `SEG-01`, `SEG-02`, `SEG-04`, `TST-03`. Rules: `sdd-arquitetura`, `sdd-codigo`, `sdd-seguranca`, `sdd-testes`.

## Critérios de Aceitação Executáveis
Testes usam `t.TempDir()` e variáveis de ambiente por `t.Setenv`; não dependem da identidade do git nem do `forge` do PATH.
```bash
set -o pipefail
go build ./... && go vet ./...
go test -list 'TestDepreciadoGravaLinhaCSV' ./internal/deprecation | grep -F 'TestDepreciadoGravaLinhaCSV'
go test ./internal/deprecation -run 'TestDepreciadoGravaLinhaCSV'
go test -list 'TestComandoDepreciadoRegistraEAvisa' ./cmd/forge-sdd | grep -F 'TestComandoDepreciadoRegistraEAvisa'
go test ./cmd/forge-sdd -run 'TestComandoDepreciadoRegistraEAvisa'
go test -list 'TestAgenteDetectadoPorVariavel' ./internal/deprecation | grep -F 'TestAgenteDetectadoPorVariavel'
go test ./internal/deprecation -run 'TestAgenteDetectadoPorVariavel'
# negativo: contexto nunca contém valores de argumento, caminho do usuário nem segredo
go test -list 'TestContextoSemValoresDeArgumento' ./internal/deprecation | grep -F 'TestContextoSemValoresDeArgumento'
go test ./internal/deprecation -run 'TestContextoSemValoresDeArgumento'
# negativo: célula iniciada por = + - @ é neutralizada; quebra de linha e ANSI saneados
go test -list 'TestCelulaNeutralizaFormulaESaneia' ./internal/deprecation | grep -F 'TestCelulaNeutralizaFormulaESaneia'
go test ./internal/deprecation -run 'TestCelulaNeutralizaFormulaESaneia'
# negativo: link simbólico no destino é recusado e o comando segue funcionando
go test -list 'TestRegistroRecusaLinkSimbolico' ./internal/deprecation | grep -F 'TestRegistroRecusaLinkSimbolico'
go test ./internal/deprecation -run 'TestRegistroRecusaLinkSimbolico'
# negativo: falha de escrita não quebra o comando (código 0, aviso mantido)
go test -list 'TestFalhaDeEscritaNaoQuebraComando' ./cmd/forge-sdd | grep -F 'TestFalhaDeEscritaNaoQuebraComando'
go test ./cmd/forge-sdd -run 'TestFalhaDeEscritaNaoQuebraComando'
# negativo: arquivo acima do limite não cresce
go test -list 'TestRegistroRespeitaTamanhoMaximo' ./internal/deprecation | grep -F 'TestRegistroRespeitaTamanhoMaximo'
go test ./internal/deprecation -run 'TestRegistroRespeitaTamanhoMaximo'
# negativo: --dry-run não cria o CSV (ARQ-05)
go test -list 'TestDryRunNaoGravaRegistro' ./cmd/forge-sdd | grep -F 'TestDryRunNaoGravaRegistro'
go test ./cmd/forge-sdd -run 'TestDryRunNaoGravaRegistro'
# negativo: comando vigente não grava linha
go test -list 'TestComandoVigenteNaoRegistra' ./cmd/forge-sdd | grep -F 'TestComandoVigenteNaoRegistra'
go test ./cmd/forge-sdd -run 'TestComandoVigenteNaoRegistra'
go test ./...
```
- Aviso em pt-BR, stderr, sem ANSI (herdado da 1.3); saída de dados inalterada.
- O Builder registra no `status.md` o que ficou "a verificar" (variáveis de Gemini, Copilot e Codex; hook dos aliases).
