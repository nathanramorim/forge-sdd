# feat-70-01 — Layout e Configuração do Forge (`.forge/`)

**Branch:** `feat/70-pipeline-multi-sessao-orquestrada`
**Fase:** 70-01
**Tamanho:** `M`
**Depende de:** —
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
O pipeline guarda configuração versionada e estado de execução em `.forge/`, ao lado de `sdd/`. Esta task cria a base: caminhos no `internal/layout`, configuração tipada com validação estrita em `internal/config`, e os templates que não ativam o modo pipeline sozinhos.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** extensão de dois pacotes existentes (caminhos e configuração), só stdlib; sem concorrência nem efeito no SO (`ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Dono dos caminhos | `internal/layout` | Todos os caminhos de `sdd/` e a validação de ids | estender | `.forge/` entra aqui (`DUP-04`) |
| Dono da configuração | `internal/config` (`ReadSddrc`) | `.sddrc` e tipos de configuração | estender | Um único dono para os arquivos de configuração (`DUP-01`) |
| Estado fora do git | `sdd/.runtime/.gitignore` (scaffold) | Diretório que ignora tudo menos o próprio `.gitignore` | reusar o padrão | `.forge/runtime/.gitignore` idêntico |
| Segredo literal | `internal/gate` (P-13) | Varredura de padrões de segredo | reusar a abordagem | Aqui o esquema estrito já impede segredo literal; `*_env` só aceita nome de variável |

## Dependências novas
- Nenhuma. Apenas stdlib Go (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`, `ARQ-02`, `DUP-01`, `DUP-04`, `SEG-01`, `SEG-04`, `COD-03`, `TST-02`, `TST-03`, `TST-04`.
- Rules: `sdd-arquitetura.md`, `sdd-seguranca.md`, `sdd-codigo-e-testes.md`.
- Gerais: `FLX-01`, `FLX-02`, `COD-02`, `COD-05`, `TST-05`.

## Escopo
1. `internal/layout`: constantes e funções para `.forge/`, `config/forge.json`, `config/notifications.json`, `runtime/queues/<papel>/<estado>`, `runtime/queues/discovery/{inbox,archive}` e `runtime/worktrees/`. Papel em `builder|reviewer|discovery` e estado em `pending|processing|done|failed|inbox|archive` validados contra lista fechada (`SEG-01`).
2. `internal/config`: `ForgeConfig` (`cli`, `cli_profiles`, `pm_mcp`, `builder`, `reviewer`, `adr_dirs`) e `NotificationsConfig` (canais `discovery_session`, `email`, `webhook`, `pm_comment`). `LoadForgeConfig(dir)` devolve `(cfg, found, err)`; `found=false` significa modo 4.0.0-beta. Defaults da especificação; decodificação estrita (campo desconhecido é erro). Campos de segredo só por variável (`url_env`, `user_env`, `password_env`) com nome validado por `^[A-Z_][A-Z0-9_]*$`. Limites: `max_parallel_tasks` de 1 a 8 e `max_rework_attempts` de 1 a 10.
3. Templates: `.forge/config/forge.json.example`, `.forge/config/notifications.json.example` e `.forge/runtime/.gitignore`. **Os exemplos não ativam o modo pipeline**: só um `forge.json` real ativa.
4. `forge-sdd pipeline init [--dry-run]`: grava `forge.json` e `notifications.json` a partir dos defaults, sem sobrescrever o que existe (`SEG-02`, `ARQ-05`).

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./internal/layout/... ./internal/config/... ./cmd/forge-sdd/...
```
- Sem `forge.json`: `found=false`, sem erro.
- Conformes: arquivo com os defaults e arquivo com todos os campos preenchidos.
- Violações (`TST-03`): campo desconhecido, `cli` fora de `claude|agy`, limites fora da faixa, `url_env` com valor que parece URL ou token, JSON inválido, papel ou estado fora da lista (`../x`), segredo literal recusado.
- `pipeline init --dry-run` não escreve; segunda execução não altera nada.
- Goldens do scaffold atualizados (`TST-02`).
