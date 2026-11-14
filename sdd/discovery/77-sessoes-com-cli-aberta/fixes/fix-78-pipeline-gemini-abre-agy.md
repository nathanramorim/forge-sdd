# fix-78 — Pipeline com Gemini abre o CLI agy (Antigravity)

**Branch:** `fix/78-pipeline-gemini-abre-agy` (a partir da `feat/77-sessoes-com-cli-aberta`)
**Origem:** Discovery 77 / Feat 77 (modo live, sessões abertas no tmux) e Fix 76 (unificação do Gemini no agy)
**Tamanho:** `S`
**Status:** `implemented`

## Feature Correlacionada
- **Arquivo de origem corrigido:** `internal/config/forge_profiles.go` (`Profile`, `defaultProfiles`), `internal/config/forge.go` (`LoadForgeConfig`), `internal/pipeline/runner.go` (`expandArgs`), `internal/pipeline/preflight.go` (`agentFiles`), `internal/scaffold/templates/.forge/config/forge.json.example.tmpl` e testdata golden correspondente.
- **Causa-raiz:**
  1. **Persistência e retrocompatibilidade de `forge.json`:** Ao carregar um `.forge/config/forge.json` pré-existente (criado antes do fix-76 ou a partir de templates/exemplos antigos), o `strictDecode` deserializa `cli_profiles["gemini"]` com `Command: "gemini"`. O método `Profile(cli)` preenchia apenas `InteractiveArgs` e `ResetCommand` quando vazios, mas nunca normalizava `Command: "gemini"` para `"agy"`. Como resultado, qualquer seleção de `"gemini"` pelo menu ou flags continuava instanciando o binário legado `gemini`.
  2. **Incompatibilidade de flags interativas de `agy`:** O binário `agy` aceita `-i <prompt>` (`--prompt-interactive`), que exige obrigatoriamente um argumento de prompt e falha com `flag needs an argument: -i` se invocado sem valor. Em `SessionModeLive` (ADR-0008), Builder e Reviewer iniciam a CLI interativa no tmux sem prompt inicial (`spec.Prompt == ""`). O método `expandArgs` omitia apenas `"{prompt}"`, deixando a flag `"-i"` solta na linha de comando (`agy -i`), provocando saída imediata da CLI com erro. Além disso, o perfil legado `CLIAgy` mantinha `InteractiveArgs: ["--agent", "{agent}", "-i"]`, que falha duplamente (`error: unknown agent "lider"` e `flag needs an argument: -i`).
  3. **Verificação de PATH no Preflight e Menu:** O preflight valida a disponibilidade da CLI no PATH através de `profile.Command`. Quando o perfil continha `Command: "gemini"`, verificava a presença de `gemini` em vez de `agy`. Se o usuário instalou o pacote `@google/antigravity-cli`, o comando `agy` está no PATH mas `gemini` não está, gerando falso positivo de bloqueio; caso o binário antigo `gemini` estivesse instalado, a validação passava e abria o cliente descontinuado que falha com `IneligibleTierError`.
- **Evidência do defeito (relato de uso real):**
  Ao selecionar a opção "gemini" na configuração ou abertura do pipeline (`forge` → Pipeline multi-sessão → Iniciar as 3 sessões → escolher `gemini`), o Forge abre janelas executando a CLI `gemini` em vez de `agy`. A sessão `forge-discovery` falha com `IneligibleTierError: This client is no longer supported for Gemini Code Assist for individuals... migrate to the Antigravity suite of products`, e as sessões em modo live de Builder e Reviewer falham ao inicializar.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** Correção cirúrgica de normalização de configuração e expansão de argumentos CLI em `internal/config` e `internal/pipeline`, sem novos subsistemas, novos processos em background ou concorrência nova (`ARQ-11`).

## Decisões do projeto aplicadas
- `D-1, D-3 (ADR-0005)`: Perfis por CLI e `CLIProfile` com `interactive_args`. A seleção de agente `gemini` pelo usuário no menu representa a família Google Antigravity e deve mapear para o executável oficial `agy`.
- `D-1, D-2 (ADR-0008)`: Sessões em modo `live` iniciam a CLI interativa aberta nos painéis; Builder e Reviewer iniciam prontos e sem prompt inicial (`spec.Prompt == ""`), aguardando entrega dos despachantes.
- `ARQ-01`, `COD-01`, `DUP-01`: Normalização centralizada em `internal/config` (`Profile` e `LoadForgeConfig`), garantindo que o resto da esteira (`launcher`, `tmux`, `worker`, `preflight`, `menu`) receba a configuração sanitizada sem duplicação de regras.
- `SEG-03`: Invocação por argv explícito sem interpolação de shell.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Perfis padrão de CLI | `internal/config/forge_profiles.go` (`defaultProfiles`) | `CLIGemini` e `CLIAgy` com comandos e argumentos | estender | Unificar `CLIGemini` e `CLIAgy` para `Command: "agy"`, `InteractiveArgs: ["-i", "{prompt}"]` e `ResetCommand: "/clear"`. |
| Resolução e retrocompatibilidade de perfil | `internal/config/forge_profiles.go` (`Profile`) | Completa `InteractiveArgs` e `ResetCommand` vazios | estender | Normalizar `Command: "gemini"` para `"agy"` e substituir `args` legados do gemini (`-o stream-json`) pelos args corretos do agy. |
| Carga e validação de `forge.json` | `internal/config/forge.go` (`LoadForgeConfig`) | Lê e decodifica estritamente o JSON | estender | Garantir que perfis existentes no JSON passem pela normalização do `Profile`. |
| Expansão de placeholders de comandos | `internal/pipeline/runner.go` (`expandArgs`) | Omite `{prompt}` quando `spec.Prompt == ""` | estender | Se o argumento anterior for `"-i"` e `spec.Prompt == ""`, omitir também a flag `"-i"`, evitando passar opção sem valor. |
| Checagem de ferramentas e arquivos no Preflight | `internal/pipeline/preflight.go` (`agentFiles`, `Preflight`) | Validação de PATH e arquivos de instrução | reusar | Com `profile.Command` normalizado para `"agy"`, o `Preflight` e o menu verificam corretamente o binário `agy`. |
| Templates de scaffold de configuração | `internal/scaffold/templates/.forge/config/forge.json.example.tmpl` | Template de exemplo | manter | Já possui `command: "agy"`, garantir sincronia em testdata golden. |

## Dependências novas
- Nenhuma. Stdlib Go pura (`fmt`, `strings`, `regexp`) (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`, `ARQ-04`, `ARQ-11`: Separação de responsabilidades em `internal/`, sem dependências externas, modelo enxuto.
- `COD-01`, `COD-02`, `COD-03`: Responsabilidade única por função, go vet limpo, tratamento de erro idiomático.
- `DUP-01`, `DUP-02`: Fonte única de resolução de perfis (`Profile`), reuso obrigatório.
- `SEG-03`: Argv explícito para execução sem shell.
- `TST-03`, `TST-04`, `TST-05`: Testes unitários com assertions claras, isolamento e cobertura executável.
- `FLX-01`, `FLX-02`: Spec antes do código, ciclo completo.
- Rules: `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-codigo-e-testes.md`, `.agents/rules/sdd-reuso-e-duplicacao.md`, `.agents/rules/sdd-fluxo-e-portao.md`.

## Escopo
1. **Normalização do perfil `gemini` em `ForgeConfig.Profile`:**
   - Se `cli == CLIGemini` e `profile.Command == "gemini"` (ou vazio): normalizar automaticamente para `Command: "agy"`.
   - Se `profile.Args` contiver marcadores ou flags exclusivas da CLI descontinuada (ex.: `-o stream-json`): substituir pelos `Args` padrão do `CLIGemini` (`-p {prompt} {permissions}`).
   - Se `c.CLIProfiles["gemini"]` estiver ausente mas `c.CLIProfiles["agy"]` existir: utilizar as definições do agy como fallback.
2. **Correção de argumentos interativos de `CLIGemini` e `CLIAgy`:**
   - Em `defaultProfiles()`:
     - `CLIGemini`: `Command: "agy"`, `Args: ["-p", "{prompt}", MarkerPermissions]`, `InteractiveArgs: ["-i", "{prompt}"]`, `ResetCommand: "/clear"`.
     - `CLIAgy`: idêntico a `CLIGemini` (remove `"--agent", "{agent}"` que não é suportado pelo agy).
3. **Tratamento de flag dependente em `expandArgs` (`runner.go`):**
   - Ao iterar sobre o template de argumentos, se um argumento for `"-i"` seguido de `"{prompt}"` e `spec.Prompt == ""`: omitir tanto o `"-i"` quanto o `"{prompt}"`.
   - Garantir que Builder e Reviewer no modo `live` recebam apenas `agy` (ou `agy {permissions}`) sem o flag `-i` solto.
   - Preservar o comportamento de Discovery que recebe `agy -i 'Voce e a sessao...'`.
4. **Verificação no Preflight e Menu:**
   - Como `Profile("gemini")` passa a retornar `Command: "agy"` mesmo para `forge.json` antigo, `chooseAgents` e `Preflight` checam a existência do executável `agy` no PATH.
   - Mensagem de erro de preflight quando ausente: orienta instalação via `npm install -g @google/antigravity-cli`.
5. **Não faz parte:**
   - Remover a constante `CLIGemini` ou a opção visual "gemini" do menu (o nome "gemini" continua sendo oferecido ao usuário, mas seu motor de execução é o `agy`).
   - Alterar perfis ou lógica de invocação de Claude, Copilot ou Codex.

## Critérios de Aceitação Executáveis
```bash
go test -v -race ./internal/config/... ./internal/pipeline/... ./cmd/forge-sdd/...
```
- `TestProfile_NormalizaGeminiLegadoParaAgy`: um `ForgeConfig` contendo `cli_profiles["gemini"]` com `Command: "gemini"` e flags antigas retorna perfil com `Command: "agy"`, argumentos não-interativos atualizados e `ResetCommand: "/clear"`.
- `TestExpandArgs_FlagPromptVazio`: ao expandir `[]string{"-i", "{prompt}"}` com `spec.Prompt == ""`, o resultado é vazio (`[]string{}`), sem deixar `"-i"` orfão.
- `TestSessions_LiveModeBuilderReviewerGeminiAbreAgySemFlagI`: com `session_mode: "live"` e `sessions: {"builder": "gemini", "reviewer": "gemini"}`, as linhas de comando geradas para os painéis executam `agy` (com flags de permissão quando aplicável) e não contêm `-i` nem `gemini`.
- `TestSessions_DiscoveryGeminiAbreAgyComPromptI`: para a Discovery, a linha de comando gerada executa `agy -i 'Voce e a sessao forge-discovery...'`.
- `TestPreflight_GeminiChecaBinarioAgy`: com perfil `gemini`, `Preflight` consulta `LookPath("agy")`; ausência de `gemini` no PATH não causa erro se `agy` estiver presente.
- `go build ./... && go vet ./...`: compilação e verificação estática 100% limpas.
