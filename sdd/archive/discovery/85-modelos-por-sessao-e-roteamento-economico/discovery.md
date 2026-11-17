# Discovery 85 — Seleção de Modelos por Sessão do Pipeline e Roteamento Econômico (Advisor / Orchestrator)

## 1. Problema
No estado atual do pipeline multi-sessão do Forge-SDD (`ADR-0005`, `ADR-0008`):
1. **Seleção restrita apenas ao binário da CLI:** O usuário consegue escolher qual agente (`claude`, `gemini`/`agy`, `copilot`, `codex`) roda em cada sessão (`discovery`, `builder`, `reviewer`), mas **não consegue especificar o modelo nem o nível de esforço (`effort`)** por sessão.
2. **Risco de queima desnecessária de tokens:** Sem controle explícito de modelo/esforço por papel, todas as sessões acabam rodando no modelo padrão global da CLI (frequentemente um modelo topo de linha de alto custo, como `claude-opus` ou `gemini-4-hydrogen-high`), gastando cota e tokens caros em tarefas mecânicas ou repetitivas do Builder/Reviewer.
3. **Diferença de paradigma entre Claude ("Advisor") e Antigravity (`agy`):**
   - No **Claude Code**, o usuário costuma usar o padrão *Advisor* / delegação (`opus` para arquitetar/decidir na Discovery ou revisar casos críticos, e `sonnet` com `--effort medium/low` para construir código no Builder, ou `--fallback-model`).
   - No **Antigravity (`agy`)**, a CLI expõe tanto um meta-modelo roteador dedicado (`gemini-pro-orchestrator-g1-pro` — *Orchestrator (Argon)*) quanto variantes explícitas com nível de raciocínio embutido (`gemini-4-argon-low`, `gemini-4-argon-medium`, `gemini-4-argon-high`, `gemini-3.8-flash-high`, além de `claude-sonnet-5-5-low/medium` via `agy`) e a flag `--effort (low|medium|high|xhigh|max)`. Hoje o usuário não tem visibilidade nem presets no Forge para aproveitar esse roteamento sem queimar tokens à toa.

## 2. Para Quem
Desenvolvedores e líderes técnicos que operam a esteira multi-agente do Forge-SDD e precisam equilibrar **qualidade arquitetural** (Discovery/Reviewer) com **eficiência e baixo consumo de tokens** (Builder/Workers), tanto no Claude quanto no Antigravity (`agy`), Copilot e Codex.

## 3. Como Funciona em Cada CLI (Mapeamento Técnico)
| CLI | Como seleciona modelo | Como controla custo / "Advisor" | Modelos de referência |
|---|---|---|---|
| **Antigravity (`agy` / `gemini`)** | `--model <id>` (listados via `agy models`) | 1. **Orchestrator (`gemini-pro-orchestrator-g1-pro`):** a própria interface do AGY decide quando acionar subagentes/modelos.<br>2. **Tier Econômico por Papel:** `gemini-4-argon-medium` na Discovery e `gemini-4-argon-low` ou `gemini-3.8-flash-high` no Builder/Reviewer + `--effort low/medium`. | `gemini-pro-orchestrator-g1-pro`, `gemini-4-argon-low/medium/high`, `gemini-3.8-flash-high`, `claude-sonnet-5-5-medium` |
| **Claude (`claude`)** | `--model <alias-ou-nome>` | **Padrão Advisor Econômico:** `sonnet` (ou `opus` apenas no planejamento) + `--effort medium/low` nos workers para evitar raciocínio estendido desnecessário. | `sonnet`, `opus`, `haiku`, `claude-sonnet-4-5`, `claude-opus-4-5` |
| **GitHub Copilot (`copilot`)** | `--model <id>` | Seleção de modelo leve por papel. | `claude-sonnet-4.5`, `gpt-4.1`, `o4-mini` |
| **Codex (`codex`)** | `-m` / `--model <id>` | Modelo enxuto para execução rápida em sandbox. | `o4-mini`, `gpt-4.1`, `codex-mini-latest` |

## 4. Jornada do Usuário (Agregação ao Fluxo Existente, Sem Novo Fluxo)
A escolha de modelo **não cria um menu separado nem altera o fluxo atual**. Ela se agrega como um passo leve e opcional dentro do próprio `chooseAgents` (usado em *"Iniciar esteira (up)"* e em *"Escolher o agente de cada sessão"*):

1. **No mesmo fluxo atual de escolher o agente de cada sessão (`chooseAgents`):**
   - Para cada sessão (`forge-discovery`, `forge-builder`, `forge-reviewer`), você escolhe o agente (`claude`, `gemini`, `copilot`, `codex`) exatamente como já faz hoje.
   - Logo em seguida, para aquela sessão, aparece a seleção de **Modelo** com a primeira opção sempre sendo **"Padrão da CLI / Auto (mantém o comportamento atual)"**, seguida de:
     - **Sugestão Econômica / Advisor** para aquele papel (ex.: `gemini-4-argon-low (effort: low)` no Builder para não queimar token à toa, ou `gemini-pro-orchestrator-g1-pro` no AGY / `sonnet` no Claude);
     - Modelos listados pelo catálogo curado + `agy models` (quando o agente for `gemini`/`agy`);
     - Opção de digitar um ID customizado.
   - Se o usuário apenas der `Enter` em "Padrão da CLI", **absolutamente nada muda** nos comandos gerados nem no `forge.json`.
2. **Execução Transparente e 100% Compatível (`forge up` e `forge worker`):**
   - Só adiciona `--model <id>` (e `--effort <nivel>` quando aplicável) se a sessão tiver modelo escolhido.
   - Se `models` estiver vazio ou omitido no `forge.json`, o `BuildArgv`, `SessionsFor` e `BuildTmuxPlan` produzem byte a byte os mesmos comandos de hoje.

## 5. Regras de Negócio
- **RN-01 (Zero Quebra / Agregação Pura):** Nenhum comando, flag, teste ou arquivo `forge.json` existente pode quebrar. Sem modelo configurado na sessão, o `argv` gerado pelo launcher (`tmux` / nativo) e pelo `worker` deve ser **idêntico** ao atual.
- **RN-02 (Não é Outro Fluxo):** Não criar subcomandos novos nem menus paralelos. A seleção de modelo é agregada diretamente em `chooseAgents` (`cmd/forge-sdd/menu.go`) e como flags opcionais em `forge up` / `forge worker`.
- **RN-03 (Proteção contra Queima de Tokens):** Na lista de modelos de `builder` e `reviewer`, destacar no topo a recomendação econômica para o papel (modelos `low`/`flash-high` ou `sonnet` com `--effort low`) para evitar queimar tokens caros em tarefas autônomas.
- **RN-04 (Catálogo Resiliente sem Bloqueio):** A consulta a `agy models` tem timeout máximo de 3 segundos e limpa sequências ANSI/spinners; se falhar ou demorar, usa imediatamente a lista curada estática.
- **RN-05 (Segurança de Argumentos — `SEG-01`, `SEG-03`):** Nomes de modelos e níveis de `effort` são validados por regex estrita (`^[A-Za-z0-9._:-]{1,80}$`) e passados no `argv` sem shell.

## 6. Escopo
### Dentro do Escopo
1. **Esquema e Resolução em `internal/config`:**
   - Campos opcionais `models` (`map[string]string` por papel: `discovery`, `builder`, `reviewer`) e `efforts` (`map[string]string` por papel: `low`, `medium`, `high`) em `ForgeConfig`.
   - Marcador `{model}` nos perfis de CLI (`CLIProfile.Args` e `CLIProfile.InteractiveArgs`) que expande para `--model <id>` (e `--effort <level>` quando suportado pela CLI e configurado).
   - Presets de Advisor Econômico (`ApplyAdvisorPreset`) para Claude, Gemini/AGY, Copilot e Codex.
2. **Descoberta e Catálogo de Modelos (`internal/config/models.go`):**
   - Catálogo curado por CLI com descrição de custo/benefício em pt-BR (ex.: `gemini-4-argon-low — Rápido e econômico (ideal para Builder/Reviewer)`).
   - Parser para `agy models` (removendo códigos ANSI de spinner e extraindo ID + Nome amigável).
3. **Menu Interativo e Flags CLI (`cmd/forge-sdd/menu.go`, `up.go`, `worker.go`):**
   - Seleção de Preset Advisor ou Modelo por Sessão em `Configurações` e opcionalmente no `Iniciar esteira (up)`.
   - Flags `--discovery-model`, `--builder-model`, `--reviewer-model`, `--model` e `--effort` em `forge up` e `forge worker`.
   - Campo opcional de entrada de texto (`Input`) no `prompter` (`internal/survey`) para modelo customizado.
4. **ADR e Testes Automatizados:**
   - Registro do `ADR-0009` documentando a arquitetura de seleção de modelos por sessão e o preset Advisor econômico.

### Fora do Escopo
- Roteamento por proxy HTTP próprio ou interceptação de chamadas de rede das CLIs.
- Alterar a cobrança interna das contas Anthropic/Google/OpenAI fora das flags oficiais de cada CLI (`--model`, `--effort`).

## 7. Riscos e Mitigações
| Risco | Impacto | Mitigação |
|---|---|---|
| `agy models` imprimir spinners ANSI antes da lista no stdout | Médio | Filtrar linhas com sequências de controle ANSI (`\x1b[...`) e validar cada ID extraído com regex de modelo antes de exibir no menu. |
| Uma CLI não suportar `--effort` (ex.: `copilot` ou versão antiga) | Médio | A tabela `ModelArgs(cli, model, effort)` só emite `--effort` para CLIs que o suportam oficialmente (`claude`, `gemini`/`agy`). |
| Usuário digitar nome de modelo inválido manualmente | Baixo | Validação por regex (`SEG-01`) antes de salvar no `forge.json` e exibição clara do comando gerado no `--dry-run` e no cabeçalho do worker. |
