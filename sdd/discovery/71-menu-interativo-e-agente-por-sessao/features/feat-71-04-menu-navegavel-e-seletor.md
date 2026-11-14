# feat-71-04 — Menu Navegável e Seletor de Agente

**Branch:** `feat/71-menu-interativo-e-agente-por-sessao`
**Fase:** 71-04
**Tamanho:** `G`
**Depende de:** 71-02, 71-03
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
`forge` sem argumentos cai direto em `init` ou `update`, e há dezenas de subcomandos. Esta task entrega o menu navegável que separa o pipeline da estrutura padrão, o seletor de agente por sessão (que lembra a última escolha) e a integração com a checagem prévia.

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** superfície interativa nova que chama várias funções existentes, com detecção de TTY e testes por entrada simulada; o risco é duplicar lógica dos subcomandos e travar em ambiente sem terminal (`ARQ-11`, `COD-01`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Prompts interativos | `internal/survey` (`huh` v0.3.0) | `Run`, `RunUpdate`, `RunConfirm`, `RunSmartUpgradePrompt` | reusar e estender | menu e seletor usam o mesmo `huh`; nenhuma dependência nova |
| `forge` sem argumentos | `cmd/forge-sdd/main.go` (`rootCmd.RunE`) | `runUpdateFlow` ou `runInitFlow` por `config.DetectProject` | estender | vira o menu; `init`/`update` ficam como opções que chamam as mesmas funções |
| Ações dos subcomandos | `cmd/forge-sdd/*.go`, `internal/*` | `status`, `gate`, `doctor`, `queue panel`, `bus list`, `inbox`, `up`, `pipeline init` | extrair o mínimo | lógica que está só no `RunE` sobe para função nomeada; o menu chama a função, não reexecuta o binário (`DUP-01`) |
| Persistência da escolha | `config.SaveForgeConfig` (71-01) | Escrita atômica de `forge.json` | reusar | grava `sessions` só depois da confirmação |
| Checagem prévia | `pipeline.Preflight` (71-03) | Itens ok, aviso e bloqueio | reusar | mostrada no resumo antes de abrir; bloqueio impede abrir |

## Dependências novas
- Nenhuma. `huh` já está em `go.mod`; `go-isatty` passa de indireta a direta, sem mudar a versão (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `ARQ-01`, `ARQ-02`, `DUP-01`, `COD-01`, `COD-03`, `TST-03`, `TST-04`.
- Rules: `sdd-arquitetura.md`, `sdd-codigo-e-testes.md`.
- Gerais: `FLX-01`, `FLX-02`, `COD-02`, `TST-05`.
- Skill: `novo-subcomando-cli-sdd`.

## Escopo
1. `forge` com TTY abre o menu: **Pipeline multi-sessão** (iniciar as 3 sessões, ver os comandos de abertura, ativar ou reconfigurar, fila e inbox, recuperar jobs presos, diagnóstico, abrir o guia) e **Estrutura padrão** (status das iniciativas, painel da fila, gate, doctor, instalar ou atualizar, ver todos os comandos). Num projeto sem Forge, instalar é a primeira opção. Setas e Enter navegam, Esc volta um nível (no menu principal, sai) e a opção Sair encerra.
2. Seletor: três perguntas (`forge-discovery`, `forge-builder`, `forge-reviewer`) com Claude, Gemini, Copilot e Codex; a escolha lembrada vem marcada; CLI fora do PATH aparece com aviso. Depois, escolha do nível de permissão com o que cada um libera, o resumo da checagem prévia e a confirmação; só então grava `sessions` e abre.
3. Sem TTY (`isatty` falso): imprime a ajuda curta e sai com 0, sem perguntar nada.
4. O menu chama funções (`runStatus`, `runGate`, `runDoctor`, `busList`, `inboxList`, `runUp`, `runPipelineInit`, `runInitFlow`, `runUpdateFlow`) e não reexecuta o binário; o menu roda uma ação e sai (não é painel ao vivo).
5. `forge init`, `forge update` e `forge-sdd up` explícitos não mudam.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./cmd/forge-sdd/... ./internal/survey/...
```
- `TestMenu_PipelineIniciar`: entrada simulada escolhe Pipeline, um agente por sessão, o nível de permissão e confirma; confere os três comandos gerados e a gravação de `sessions`.
- `TestMenu_EscVolta`, `TestMenu_SemTTYImprimeAjuda` (código 0, nenhuma pergunta).
- `TestMenu_BloqueioImpedeAbrir`: CLI ausente bloqueia; aviso pede confirmação.
- Equivalência: ações do menu e subcomandos produzem o mesmo resultado (`status`, `gate`, `doctor`, `bus list`, `inbox`).
- Lembrar a escolha: a próxima abertura já vem marcada; `sessions` preserva os outros campos de `forge.json`.
- Projeto sem Forge: a primeira opção é instalar e chama a mesma função de `init`.
