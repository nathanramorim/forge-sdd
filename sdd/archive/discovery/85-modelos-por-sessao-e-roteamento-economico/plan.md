# Plano de Implementação — Iniciativa 85

Quebra estruturada de baixo para cima (limite máximo de 7 tasks por onda), respeitando fronteiras `internal/config` → `internal/pipeline` → `cmd/forge-sdd`.

---

## Onda 1: Domínio de Configuração, Catálogo de Modelos e Presets Econômicos (`internal/config`)

### Task 85-01: Esquema `models` e `efforts` por Sessão e Expansão de `ModelArgs` em `internal/config`
- **Escopo:**
  - Estender `ForgeConfig` (`internal/config/forge.go` e `forge_profiles.go`) com `Models map[string]string` (`json:"models,omitempty"`) e `Efforts map[string]string` (`json:"efforts,omitempty"`).
  - Adicionar `MarkerModel = "{model}"` aos `validMarkers` e incluir `ModelFor(role)`, `EffortFor(role)`, `ModelArgs(cli, model, effort)` e `SessionModelArgs(role)`.
  - Validar formatos de modelo (`^[A-Za-z0-9._:-]{1,80}$`) e níveis de `effort` (`low`, `medium`, `high`, `xhigh`, `max`) em `validateSessions()`.
- **Rules e IDs aplicáveis:** `ARQ-01`, `ARQ-02`, `ARQ-04`, `ARQ-10`, `ARQ-11`, `COD-01`, `COD-02`, `COD-03`, `DUP-01`, `DUP-02`, `SEG-01`, `SEG-03`, `SEG-04`, `TST-01`, `TST-03`, `TST-05`.
- **Dependências:** Nenhuma.
- **Critério Executável:** `go test -v -race ./internal/config -run TestForgeConfig_Models`

### Task 85-02: Catálogo Híbrido de Modelos (`agy models`), Guia de Custo e Presets Advisor (`internal/config/models.go`)
- **Escopo:**
  - Criar `internal/config/models.go` e `models_test.go` com:
    - Catálogo curado de modelos por CLI (`claude`, `gemini`/`agy`, `copilot`, `codex`) com rótulos explicativos de consumo de tokens.
    - Parser `ParseAgyModelsOutput(raw string)` que limpa spinners ANSI e extrai modelos dinâmicos do `agy models`, e `DiscoverCLIModels(ctx, cli, runner)` com timeout de 3s e fallback seguro para a lista curada.
    - Presets Advisor (`PresetEconomic`, `PresetOrchestrator`, `PresetDefault`) via `ApplyAdvisorPreset(cfg, preset)` que configuram modelos econômicos e `effort` controlado por sessão para não queimar tokens à toa.
- **Rules e IDs aplicáveis:** `ARQ-01`, `ARQ-04`, `COD-01`, `COD-02`, `COD-03`, `DUP-01`, `DUP-02`, `SEG-01`, `SEG-03`, `TST-01`, `TST-03`, `TST-05`.
- **Dependências:** Task 85-01.
- **Critério Executável:** `go test -v -race ./internal/config -run "TestModelCatalog|TestAdvisorPreset"`

---

## Onda 2: Propagação no Runner, Launcher, Tmux e Worker (`internal/pipeline`)

### Task 85-03: Expansão de Modelo e Effort no Runner, Launcher e Plano Tmux (`internal/pipeline`)
- **Escopo:**
  - Estender `RunSpec` (`internal/pipeline/runner.go`) com `ModelArgs []string` e atualizar `expandArgs` para expandir `{model}` (ou anexar `spec.ModelArgs` se o perfil não contiver `{model}` explícito, garantindo compatibilidade com `forge.json` existentes).
  - Atualizar `interactiveLineForRole` e `SessionsFor` (`internal/pipeline/launcher.go`), `BuildTmuxPlan` (`internal/pipeline/tmux.go`) e `RunWorker` / `RunDeliverWorker` (`internal/pipeline/worker.go`, `deliver.go`) para propagar `--model` e `--effort` da sessão tanto em modo `headless` quanto em modo `live`, exibindo o modelo ativo no cabeçalho do worker.
- **Rules e IDs aplicáveis:** `ARQ-01`, `ARQ-02`, `ARQ-04`, `COD-01`, `COD-02`, `COD-03`, `DUP-01`, `DUP-02`, `SEG-01`, `SEG-03`, `TST-01`, `TST-03`, `TST-05`.
- **Dependências:** Task 85-01.
- **Critério Executável:** `go test -v -race ./internal/pipeline -run "TestBuildArgv_Model|TestSessionsFor_Model|TestBuildTmuxPlan_Model|TestRunWorker_Model"`

---

## Onda 3: Menu Interativo, Flags de CLI (`up` / `worker`) e ADR (`cmd/forge-sdd`, `sdd/spec/adr`)

### Task 85-04: Agregação da Seleção de Modelo em `chooseAgents` e Flags Opcionais em `up` / `worker`
- **Escopo:**
  - Agregar a escolha de modelo diretamente dentro de `chooseAgents` (`cmd/forge-sdd/menu.go`), logo após a escolha do agente de cada sessão, sem criar novos menus nem alterar a estrutura de navegação existente. Se o usuário escolher `"Padrão da CLI"`, nenhuma chave extra é exigida e o comportamento permanece 100% idêntico ao atual.
  - Adicionar flags opcionais `--discovery-model`, `--builder-model`, `--reviewer-model`, `--model` e `--effort` em `cmd/forge-sdd/up.go` e `--model` / `--effort` em `cmd/forge-sdd/worker.go`, mantendo todos os testes existentes de `menu_test.go`, `up_test.go` e `worker_test.go` passando sem regressão.
- **Rules e IDs aplicáveis:** `ARQ-01`, `ARQ-06`, `ARQ-10`, `COD-01`, `COD-02`, `COD-03`, `COD-05`, `DUP-01`, `DUP-02`, `SEG-01`, `TST-01`, `TST-03`, `TST-05`.
- **Dependências:** Tasks 85-02 e 85-03.
- **Critério Executável:** `go test -v -race ./cmd/forge-sdd -run "TestMenu|TestUp|TestWorker"`

### Task 85-05: Registro do ADR-0009, Atualização de `concepts.md` e Validação Global do Portão
- **Escopo:**
  - Criar `sdd/spec/adr/adr-0009-modelos-por-sessao-e-roteamento-economico.md` documentando as decisões arquiteturais de seleção de modelos, mapeamento Claude Advisor vs. AGY Orchestrator e proteção contra queima de tokens.
  - Atualizar `sdd/spec/concepts.md` na linha `Perfis de Agente por Sessão` com os novos contratos (`ModelFor`, `EffortFor`, `ModelArgs`, `ApplyAdvisorPreset`).
  - Rodar suíte completa de testes, `go vet ./...` e `forge gate`.
- **Rules e IDs aplicáveis:** `COD-02`, `COD-04`, `DUP-01`, `FLX-05`, `GOV-01`, `TST-01`, `TST-05`.
- **Dependências:** Task 85-04.
- **Critério Executável:** `go build ./... && go vet ./... && go test -v -race ./... && forge gate`
