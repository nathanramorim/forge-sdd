# Stack — forge-sdd (v5.0.0)

## Dependências

| Camada | Lib | Versão | Import path |
|--------|-----|--------|-------------|
| Runtime | Go | `go 1.22` em `go.mod` (CI usa `go-version-file: go.mod`) | `GOTOOLCHAIN=local` |
| CLI | cobra | v1.9.1 | `github.com/spf13/cobra` |
| Prompts e menus | huh | v0.3.0 | `github.com/charmbracelet/huh` |
| Componentes de TUI | bubbles | v0.17.2 (pré-release) | `github.com/charmbracelet/bubbles` |
| Detecção de terminal | go-isatty | v0.0.20 | `github.com/mattn/go-isatty` |
| Templates | embed.FS | stdlib | `//go:embed all:templates` |
| Render | text/template | stdlib | `text/template` |
| Testes | testify | v1.9.x | `github.com/stretchr/testify` |
| Lint | golangci-lint | v1.57+ | via GitHub Actions |
| Release | goreleaser | v2.x | `.goreleaser.yaml` |

Dependências indiretas relevantes: `bubbletea` e `lipgloss` (renderização da TUI do `huh`).

## Ferramentas externas (opcionais, detectadas em runtime)
- **tmux**: abre as 3 sessões do pipeline em painéis numa tela só (`forge up`); sem ele, há contingência para janelas do sistema ou Windows Terminal (ADR-0007).
- **CLIs de agente**: `claude`, `gemini`, `copilot`, `codex` ou `agy`, uma por sessão (perfis em `internal/config/forge_profiles.go`).
- **git** e **gh**: branches por iniciativa, worktrees do Builder e criação de PR.

## Distribuição
- Pacote npm `@nathanramorim/forge-sdd`: um launcher (`bin/run.js`) baixa o binário da versão do GitHub Releases, confere o SHA256 e o guarda em `~/.cache/forge-sdd/<versão>/`.
- Binários: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64 e windows/amd64.
- Site: `site/` (Astro 5 SSG, `ARQ-04`) e **LP estática em `docs/`**, publicada em `https://forge-sdd.org` por **Cloudflare Workers** (`site/wrangler.jsonc` serve `../docs`), por deploy manual: `scripts/lp-deploy.sh` local ou o workflow `deploy-lp.yml` só com `workflow_dispatch` (ADR-0020, D-2 emendada).
- Workflows em `.github/workflows/`: `ci.yml`, `release.yml`, `npm-publish.yml`, `deploy-lp.yml`, `encerrar-edicao.yml`.
- Divergência registrada: `npm/package.json` ainda declara `4.9.1` enquanto a tag git é `v5.0.0` (não corrigido aqui).

## Layout do projeto
```
cmd/forge-sdd/           # camada Cobra: init, update, doctor, status, queue, bus, dispatch,
                         # worker, up, ps, inbox, review, pipeline, brief, context, gate,
                         # skills, session, report, archive, lessons, next-id, acp, mcp, eval, rules, menu...
internal/
  acp/                   # servidor Agent Client Protocol (stdio e HTTP) e MCP somente leitura (mcp.go)
  bus/                   # barramento de jobs entre as sessões
  config/                # .sddrc, forge.json, perfis de agente e agent_models por papel
  doctor/                # diagnóstico do harness (`forge doctor --harness`)
  harness/               # contratos entre etapas, lock por iniciativa, parecer
  gate/                  # Portão Camada A, brief e contexto mínimo
  governance/            # importador de regras externas
  layout/                # caminhos de sdd/ e .forge/, iniciativa ativa pela branch
  monitor/               # processos, tokens por janela, feed de chat, telemetria
  notify/                # inbox, toast, e-mail e webhook
  pipeline/              # dispatch, worker, review, tmux, handoff, auto-archive, status vivo,
                         # laço do modo chat (laco.go) e leque com worktree por task (leque.go)
  queue/                 # fila por iniciativa, filas por estação, atividades
  scaffold/              # templates embarcados, migração e golden tests
  skills/                # Motor de Auto-Skills
  survey/                # formulário do init e TUI da fila com drill-down
  testutil/              # auxiliares de teste
site/                    # Astro 5 (isolado, ARQ-04)
npm/                     # launcher do binário
sdd/                     # memória, specs, discovery, releases e métricas do próprio projeto
```

## Características do binário
- Estático (`CGO_ENABLED=0`), cerca de 10 MB.
- Rede só para notificações opcionais (webhook e e-mail) e para o launcher npm baixar o binário.
- Estado do pipeline em `.forge/` (jobs, logs, runtime) e do projeto em `sdd/`.
