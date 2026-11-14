# feat-71-01 — Perfis por Agente e `sessions` em `forge.json`

**Branch:** `feat/71-menu-interativo-e-agente-por-sessao`
**Fase:** 71-01
**Tamanho:** `M`
**Depende de:** —
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
A 4.1 só conhece `claude` e `agy` e usa uma CLI para as três sessões. Esta task dá à configuração o vocabulário novo: perfis de `claude`, `gemini`, `copilot` e `codex` (modo não interativo e interativo), o agente de cada sessão (`sessions`) e o nível de permissão do worker.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** extensão de um pacote existente (`internal/config`), só stdlib, sem concorrência nem efeito no sistema operacional (`ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Perfil de CLI | `internal/config/forge.go` (`CLIProfile`, `DefaultForgeConfig`, `validateProfile`) | `claude` e `agy`, só `args` não interativos, `ForgeConfig.CLI` único | estender | `interactive_args`, `sessions`, `permissions` e perfis novos entram aqui; `cli` continua como padrão quando `sessions` falta |
| Substituição de marcadores | `internal/pipeline/runner.go` (`BuildArgv`) | `{agent}`, `{name}`, `{prompt}`, sem shell, prompt não vira opção | reusar e estender | ganha o marcador `{permissions}` que expande para zero ou mais argumentos vindos da tabela de permissões |
| Escrita atômica | `internal/queue/atomic.go` (`WriteJSONAtomic`) | Arquivo temporário e `os.Rename` | reusar | `SaveForgeConfig` grava `forge.json` por ela |
| Validação estrita | `internal/config/forge.go` (`strictDecode`, `Validate`) | Campo desconhecido é erro; limites e nomes de variável | reusar | `sessions` e `permissions` entram na mesma validação |

## Dependências novas
- Nenhuma. Apenas stdlib Go (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`, `ARQ-02`, `DUP-01`, `SEG-03`, `SEG-04`, `COD-03`, `TST-02`, `TST-03`, `TST-04`.
- Rules: `sdd-arquitetura.md`, `sdd-seguranca.md`, `sdd-codigo-e-testes.md`.
- Gerais: `FLX-01`, `FLX-02`, `COD-02`, `TST-05`.

## Escopo
1. `CLIProfile` ganha `interactive_args` (comando da sessão interativa, com `{prompt}`/`{name}`/`{agent}`). Constantes `CLIGemini`, `CLICopilot`, `CLICodex` ao lado de `CLIClaude` e `CLIAgy`.
2. `DefaultForgeConfig` traz os perfis pesquisados (tabela do `criteria.md`): `claude -p {prompt} --agent {agent} -n {name} {permissions}`; `gemini -p {prompt} {permissions}` e `gemini -i {prompt}`; `copilot -p {prompt} --agent {agent} -n {name} --no-ask-user {permissions}` e `copilot -i {prompt} -n {name}`; `codex exec {permissions} {prompt}` e `codex {prompt}`. Perfil escrito inteiro por CLI: `-p` é perfil no Codex. O `agy` permanece válido se estiver no arquivo, mas sai dos defaults oferecidos.
3. `ForgeConfig.Sessions` (`discovery`, `builder`, `reviewer`) e `ForgeConfig.Permissions` (`project` padrão ou `full`). `CLIFor(papel)` é o único ponto de resolução: `sessions[papel]` ou `cli`. Validação: cada agente de `sessions` precisa de perfil; papel e nível de lista fechada.
4. Tabela única `PermissionArgs(cli, nivel) []string` em `internal/config` (`DUP-01`): `project` libera edição e os comandos do fluxo por allowlist onde a CLI permite (`claude --permission-mode acceptEdits` mais ferramentas; `gemini --approval-mode auto_edit`; `copilot --allow-tool` granular; `codex --sandbox workspace-write`); `full` usa o bypass documentado de cada CLI. O que a documentação não confirma fica comentado `a confirmar`.
5. `BuildArgv` expande `{permissions}` em argumentos separados, mantém a recusa de prompt iniciado por `-` e de byte nulo.
6. `SaveForgeConfig(dir, cfg)` grava por escrita atômica, preservando os demais campos; `LoadForgeConfig` de uma `forge.json` da 4.1 (sem `sessions`) continua válida e idêntica no comportamento.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./internal/config/... ./internal/pipeline/...
```
- `TestForgeProfiles`: os quatro perfis têm comando e marcadores válidos; `forge.json` da 4.1 sem `sessions` carrega e `CLIFor` devolve `cli` nas três sessões.
- Conformes: `sessions` com agentes diferentes; `permissions` `project` e `full`.
- Violações (`TST-03`): agente de `sessions` sem perfil, papel desconhecido, `permissions` fora da lista, comando com metacaractere, marcador desconhecido, prompt iniciado por `-`.
- `BuildArgv` por CLI e nível: o argv de cada perfil bate com a tabela (teste de montagem, não de execução).
- `SaveForgeConfig` preserva os campos e não deixa arquivo parcial em falha.
