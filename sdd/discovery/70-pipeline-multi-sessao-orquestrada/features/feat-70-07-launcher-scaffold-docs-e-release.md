# feat-70-07 — Launcher, Scaffold, Skills, ADR 0004 e Release 4.1.0

**Branch:** `feat/70-pipeline-multi-sessao-orquestrada`
**Fase:** 70-07
**Tamanho:** `G`
**Depende de:** feat-70-03, feat-70-04, feat-70-05, feat-70-06
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
Falta entregar o pipeline ao usuário: abrir as três sessões em qualquer sistema, instalar `.forge/` e os agentes com `init` e `update`, atualizar as skills, documentar a decisão e fechar a versão `4.1.0` com um teste ponta a ponta simulado.

> Rule of 7: tamanho `G`. A Governança pode exigir dividir em launcher e scaffold, e em docs e release.

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** toca o scaffold, os templates, goldens, três ferramentas de agente, documentação e release; é o ponto de integração (`ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Lançamento de terminais | — | Nada; `forge-sdd attach` é de ACP | criar `forge-sdd up` e os scripts | Uma implementação em Go, scripts finos (D-8) |
| Motor de scaffold | `internal/scaffold` (`Run`, `shouldPreserve`) | Materializa templates e preserva domínio | estender | `.forge/` e os agentes novos instalados sem sobrescrever |
| Skills de estação | `.agents/skills/sdd-*` | Cinco skills com `references/` | estender | `concepcao` carrega ADRs e rules; `build` ganha o job do worker; `operacao` ganha `up` e `inbox` |
| Adaptadores finos | `.claude/commands`, `.gemini/prompts`, `.github/prompts` | Um ponteiro por comando | reusar | `inbox` e `implement` como adaptadores |
| Orçamentos | `internal/gate` (`P-24`, `managedAdapterNames`) | Limites de bytes | estender | Novos adaptadores e `reviewer` na lista gerenciada |
| Hook do Claude | 70-06 | Mesclagem segura | reusar | Ligada ao `Run` |
| Catálogo de conceitos | `sdd/spec/concepts.md` | Dono por conceito | estender | `bus`, `pipeline` e `notify` (`DUP-01`) |
| Release notes e versão | `sdd/releases/history.md`, `cmd/forge-sdd/main.go` | Padrão da 4.0.0-beta | reusar | Entrada `4.1.0` com prévia dos comandos |

## Dependências novas
- Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
- `DUP-10`, `DUP-01`, `FLX-05`, `FLX-07`, `GOV-01`, `ARQ-03`, `ARQ-05`, `SEG-02`, `TST-02`, `TST-05`.
- Rules: `sdd-reuso-e-duplicacao.md`, `sdd-fluxo-e-portao.md`, `sdd-arquitetura.md`, `sdd-seguranca.md`.
- Gerais: `FLX-01`, `FLX-02`, `COD-02`, `COD-05`, `TST-04`.

## Escopo
1. `forge-sdd up [--cli claude|agy] [--dry-run]`: monta a lista de comandos de abertura por sistema (macOS `Terminal.app` e iTerm2, Linux `gnome-terminal`, `konsole` e `x-terminal-emulator`, Windows `wt.exe` e `cmd.exe /c start`) e por CLI, com título `forge-discovery`, `forge-builder` e `forge-reviewer` (`-n` no `claude`, escape ANSI no `agy`), e executa; `--dry-run` só imprime. `forge-up.sh` e `forge-up.ps1` são envoltórios finos que chamam `forge-sdd up`, e o `.ps1` roda sem Git Bash.
2. Scaffold: instala `.forge/` (exemplos, `.gitignore` do runtime, scripts com bit de execução), os agentes `reviewer`, os adaptadores `inbox` e `implement` e liga a mesclagem do hook; `update` não sobrescreve configuração existente nem ativa o modo pipeline.
3. Skills e referências: `sdd-concepcao` carrega ADRs e rules no início do Discovery e do Split; `sdd-build` com `references/worker-job.md`; `sdd-operacao` com `up` e `inbox`; limites de `P-24` respeitados.
4. Constitution: `ARQ-06` lista `dispatch`, `worker`, `bus`, `inbox`, `pipeline` e `up`. `concepts.md` registra os três pacotes. ADR `adr-0004-pipeline-multi-sessao` (D-1 a D-8).
5. Release: entrada `### Versão 4.1.0` em `sdd/releases/history.md` (destaques, prévia dos comandos, agente explícito) e bump de versão para `4.1.0`.
6. `TestE2E_PipelineCompleto`: dispatch, builder com runner falso, reviewer com uma reprovação e uma aprovação, inbox exibido.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./...
go test ./internal/pipeline -run 'TestLauncher|TestE2E_PipelineCompleto'
bash -n .forge/forge-up.sh && forge-sdd gate
```
- Launcher: comandos corretos por sistema e por CLI, título presente, `--dry-run` sem efeito, `bash -n` limpo e teste de conteúdo do `.ps1`.
- Scaffold: goldens atualizados (`TST-02`); `update` em projeto existente instala `.forge/` sem sobrescrever e sem criar `forge.json`; segunda execução idempotente.
- Modo: sem `forge.json`, o fluxo da 4.0.0-beta permanece (testes de scaffold e de comando).
- `forge-sdd gate` com `P-24` ✓; IDs da Constituição inalterados (teste existente).
- Release notes passam na extração do workflow (`### Versão 4.1.0`), com prévia de comandos e agente explícito.
