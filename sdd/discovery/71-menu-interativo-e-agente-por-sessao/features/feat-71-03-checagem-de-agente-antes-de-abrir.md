# feat-71-03 — Checagem de Agente Antes de Abrir (`Preflight`)

**Branch:** `feat/71-menu-interativo-e-agente-por-sessao`
**Fase:** 71-03
**Tamanho:** `M`
**Depende de:** 71-01
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
Escolher um agente que não está instalado, ou cujo arquivo de instruções não existe no projeto, abre uma sessão que não funciona. Esta task cria a checagem que bloqueia ou avisa antes de abrir e corrige a lacuna conhecida do Codex, que lê `AGENTS.md` enquanto o scaffold gera `OPENAI.md`.

## Modelo do módulo
- **Modelo:** `enxuto`.
- **Justificativa:** função de verificação em `internal/pipeline` com `exec.LookPath` e leitura de arquivos; uma mudança pequena de template no scaffold (`ARQ-11`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Diagnóstico do projeto | `forge-sdd doctor` | Check-up com itens verdes e vermelhos | reusar o formato | `Preflight` devolve itens `ok`, `aviso` e `bloqueio` |
| Detecção de CLI | `exec.LookPath` (stdlib), `runner.go` | O runner já recusa CLI fora do PATH com mensagem clara | reusar a mensagem | a checagem antecipa o mesmo erro, antes de abrir terminais |
| Agentes do scaffold | `internal/scaffold/templates/agents/{claude,gemini,openai}`, `.github/` | `CLAUDE.md`, `GEMINI.md`, `OPENAI.md`, `copilot-instructions.md` | estender | o agente `openai` passa a gerar também `AGENTS.md` (o Codex não lê `OPENAI.md`); `OPENAI.md` fica por compatibilidade |
| Instalar agente | `forge-sdd update --agent <x>` | Adiciona agente a projeto existente | reusar | `Preflight` oferece este comando quando falta o arquivo do agente |

## Dependências novas
- Nenhuma. Apenas stdlib Go (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `SEG-01`, `DUP-01`, `ARQ-03`, `TST-02`, `TST-03`, `TST-04`.
- Rules: `sdd-seguranca.md`, `sdd-arquitetura.md`, `sdd-codigo-e-testes.md`.
- Gerais: `FLX-01`, `FLX-02`, `COD-02`, `TST-05`.

## Escopo
1. `Preflight(dir, sessions) []Item` com `Item{Sessao, Agente, Nivel (ok|aviso|bloqueio), Mensagem, Acao}`: CLI no PATH (ausente: bloqueio com o guia de instalação); arquivo de instruções do agente presente (`CLAUDE.md`, `GEMINI.md`, `AGENTS.md`, `.github/copilot-instructions.md`; ausente: aviso com `forge-sdd update --agent <x>`); `forge.json` válido (bloqueio); agente diferente do declarado em `agents` do `.sddrc` (aviso).
2. `PreflightBlocked(items)` e um formatador de texto curto para o terminal e para o menu.
3. Scaffold: o agente `openai` gera `AGENTS.md` com o mesmo conteúdo-fonte do `OPENAI.md` (um ponteiro curto para `.agents/`, dentro do orçamento de `P-24`); goldens atualizados.
4. Nenhuma chamada de rede e nenhuma execução da CLI: só `LookPath` e leitura de arquivos.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./internal/pipeline/... ./internal/scaffold/...
```
- `TestPreflight`: CLI ausente (bloqueio), arquivo de instrução ausente (aviso com a ação), `forge.json` inválido (bloqueio), tudo certo; `PATH` controlado no teste.
- Cada uma das quatro CLIs aponta para o arquivo certo (`codex` → `AGENTS.md`).
- Scaffold: projeto com `--agent openai` gera `AGENTS.md`; `P-24` continua ✓; goldens atualizados (`TST-02`).
- Segurança: a checagem não executa a CLI nem lê o ambiente.
