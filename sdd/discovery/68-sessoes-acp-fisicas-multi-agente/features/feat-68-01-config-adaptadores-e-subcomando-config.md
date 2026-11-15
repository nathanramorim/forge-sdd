# feat-68-01 — Configuração de Adaptadores ACP e Subcomando `forge-sdd config`

**Branch:** `feat/68-sessoes-acp-fisicas-multi-agente`
**Fase:** 68-01-config-adaptadores-e-subcomando-config
**Tamanho:** `M`
**Depende de:** —
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
Hoje não existe `forge-sdd config` e `Sddrc` não conhece agentes ACP. Esta task define o bloco `acp` em `sdd/.sddrc` (adaptadores com `command`, `args`, `env`, `healthCheck`; adaptador padrão; override por estação), presets embutidos (copilot, claude, agy, gemini, codex) e o subcomando `config` (`list`, `get`, `set`, `unset` para chaves `acp.*`).

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** extensão de `internal/config` com validação pura e um subcomando fino em `cmd/`; sem novo pacote (`ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Leitura do .sddrc | `internal/config/config.go` | `Sddrc`, `ReadSddrc` | estender | Novo campo `ACP` com `omitempty`; retrocompatível |
| Subcomando config | `cmd/forge-sdd/` | Inexistente | criar | Procedimento da skill `novo-subcomando-cli-sdd`; lógica em `internal/config` |
| Presets embutidos | `internal/scaffold/templates` | Templates via `embed.FS` | estender | Preset JSON embutido (`ARQ-03`); init/update preservam o .sddrc do usuário (`SEG-02`) |
| Registro de comandos | `cmd/forge-sdd/main.go` | `rootCmd.AddCommand` | estender | Registrar `configCmd`; atualizar `ARQ-06`/docs/golden se aplicável (`TST-02`) |

## Dependências novas
- Nenhuma. Apenas stdlib Go (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`, `ARQ-03`, `ARQ-06`, `ARQ-10` (validação em `internal/config`), `SEG-01` (validar command/args/env/nomes), `SEG-02`, `SEG-05`, `DUP-05`, `TST-02`, `TST-03`.
- Rules: `sdd-arquitetura.md`, `sdd-seguranca.md`, `sdd-reuso-e-duplicacao.md`, `sdd-codigo-e-testes.md`.
- Skill: `novo-subcomando-cli-sdd`.
- Gerais: `FLX-01`, `FLX-02`, `COD-02`, `COD-05` (pt-BR, sem emojis), `TST-05`.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race -run "TestACPConfig|TestConfigCmd" ./internal/config/... ./cmd/...
```
- Config sem bloco `acp` continua válida (teste negativo e positivo).
- Rejeita adaptador com `command` vazio, `command` com metacaracteres de shell ou nome de adaptador inválido.
- `forge-sdd config set acp.defaultAdapter gemini` persiste e `get` devolve o valor; `--dry-run` não grava (`ARQ-05`).

### Decisões FLX-01 incorporadas
- Chaves `acp.*`: `defaultAdapter`, `adapters.<nome>.{command,args,env,healthCheck.{method,timeoutMs,intervalMs}}`, `stations.<estacao>`, `warmupTimeoutMs`, `restart.{maxAttempts,initialBackoffMs,maxBackoffMs,multiplier,jitter,stableResetMs}`. Omitidos herdam defaults (timeouts herdam o warmup atual do Studio).
- `healthCheck.method`: `initialize` (ACP) ou `liveness` (processo vivo + stdin aberto, fallback para agentes sem ACP via stdio). Valor inválido é rejeitado.
- Verbos: `list` mostra chaves `acp.*` efetivas; `get <chave>` mostra uma; `set <chave> <valor>` valida e grava (respeita `--dry-run`); `unset <chave>` remove o valor do usuário e volta ao preset/default. Chaves fora de `acp.*` são rejeitadas (evolução futura: demais chaves do `.sddrc`).
- Presets: gemini, copilot, claude, codex confirmados; agy marcado "a confirmar" com `liveness` (detalhes e fontes em `criteria.md`).
- Testes adicionais: `TestConfigCmdRejectsNonACPKey`, `TestACPConfigRestartValidation` (multiplier >= 1, jitter entre 0 e 1, `maxBackoffMs >= initialBackoffMs`).

- (68-06) Presets e config ganham `enabledAgents`, `requires`, `versionArgs`, `install.{command,args}`; validação rejeita metacaracteres de shell em `install`.
