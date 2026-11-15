# feat-72-02 — Visão Limpa do Worker (`LiveView`)

**Branch:** `feat/72-visibilidade-tmux-e-entrega-ao-builder`
**Fase:** 72-02
**Tamanho:** `G`
**Depende de:** 72-01
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
O worker mostra só "job iniciado" e o log do job tinha 0 bytes enquanto o agente criava arquivos. Esta task entrega, no próprio terminal da sessão, linhas curtas do que o agente faz, sem JSON cru, sem segredos e para qualquer CLI.

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** interpreta saída não confiável de quatro CLIs, tem relógio e acompanhamento do git, e mexe em segurança (máscara de segredos); contratos e testes por camada (`ARQ-11`, `COD-01`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Laço e log do worker | `internal/pipeline/worker.go` (`processJob`, `MultiWriter`) | Saída inteira para terminal e log; com agy não há saída | estender | o terminal passa por um `LiveView`; o log continua com a saída bruta completa |
| Runner | `internal/pipeline/runner.go` (`ExecRunner`) | `cmd.Stdout = out` | reusar | o `out` entregue é o `LiveView`; o argv não muda |
| Padrões de segredo | `internal/gate/gate.go` (`secretPatterns`) | Regexes de P-13 | reusar | exportar uma função de máscara no gate e usá-la no `LiveView` (`DUP-01`) |
| Perfis de CLI | `internal/config/forge_profiles.go` | Perfis sem saída estruturada | estender | `--output-format stream-json --verbose` (claude), `-o stream-json` (gemini), `--json` (codex), a confirmar |
| Git do projeto | `internal/layout` (`Git`) | Execução de git com argv explícito | reusar | acompanhamento do `git status --porcelain` com teto de tempo |

## Dependências novas
- Nenhuma. Apenas stdlib Go (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `SEG-05`, `SEG-03`, `DUP-01`, `COD-01`, `COD-03`, `TST-03`, `TST-04`.
- Rules: `sdd-seguranca.md`, `sdd-codigo-e-testes.md`, `sdd-arquitetura.md`.
- Gerais: `FLX-01`, `FLX-02`, `COD-02`, `TST-05`.

## Escopo
1. `LiveView` (`io.Writer`) em `internal/pipeline/liveview.go`: grava tudo, bruto, no log; interpreta linhas JSON por CLI e imprime resumos (`✎ arquivo`, `$ comando`, `✓/✗ resultado`, texto do agente em uma linha); evento desconhecido ou JSON inválido nunca aparece como JSON cru (é ignorado ou, se for texto comum, mostrado truncado e sem códigos ANSI).
2. Máscara de segredos com os padrões do gate, linha truncada em 200 caracteres, nenhuma variável de ambiente; ruído conhecido de hooks globais (`SessionEnd hook [...] failed: connect ECONNREFUSED`) vira uma linha só por job.
3. Acompanhamento independente de CLI: a cada 5 s (consulta com teto de 3 s) compara `git status --porcelain` no diretório do job e imprime só o que mudou; sem mudança e sem evento por 30 s, um batimento `… trabalhando há <tempo> · <n> arquivos alterados`. Relógio e consulta injetáveis.
4. Cabeçalho e rodapé por job: `▶ job <feature> · tentativa <n> · <agente> · permissões <nível>` e `■ concluído em <tempo> · <n> arquivos` ou `✗ falhou em <tempo>: <motivo>`.
5. `forge-sdd worker --raw` mantém a saída bruta de hoje; sem TTY a visão limpa também vale (só sem cores).
6. Perfis padrão ganham as flags de saída estruturada de claude, gemini e codex (marcadas a confirmar); goldens do exemplo de `forge.json` atualizados.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./internal/pipeline/... ./internal/gate/... ./internal/config/... ./cmd/forge-sdd/...
```
- `TestLiveView`: fixtures de eventos de claude, gemini e codex viram as linhas do contrato; JSON desconhecido não aparece; texto comum sai truncado e sem ANSI; segredo mascarado; ruído de hook colapsado em uma linha; o log recebe sempre a saída completa; `--raw` devolve a saída bruta.
- `TestGitWatch`: arquivo novo e alterado viram linhas; sem mudança, batimento a cada 30 s (relógio injetado); repositório sem git não quebra; consulta lenta respeita o teto.
- Perfis: o argv de cada CLI inclui a flag de saída estruturada (teste de montagem).
- `forge-sdd gate` com `P-13` e `P-24` ✓.
