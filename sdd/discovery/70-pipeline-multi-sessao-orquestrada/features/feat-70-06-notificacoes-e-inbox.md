# feat-70-06 — Notificações e Inbox (`internal/notify`, `forge-sdd inbox`)

**Branch:** `feat/70-pipeline-multi-sessao-orquestrada`
**Fase:** 70-06
**Tamanho:** `M`
**Depende de:** feat-70-01
**Status:** `approved` (FLX-01 aprovado)

## Contexto & Motivação
Quando um PR fica pronto (ou escala), a pessoa precisa ver na sessão Discovery, sem procurar. O canal padrão grava no inbox e um hook exibe o link; toast, e-mail e webhook são reforços opcionais. Nenhum segredo pode ir para arquivo, log ou job.

## Modelo do módulo
- **Modelo:** `completo`.
- **Justificativa:** efeitos no SO (toast), rede (SMTP e HTTP), segredos por ambiente e escrita de configuração do usuário (hook); precisa de fronteira própria e testes por canal (`ARQ-11`, `COD-01`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Notificações | — | Inexistente | criar `internal/notify` | Conceito novo; entra em `concepts.md` |
| Caminhos do inbox | `internal/layout` (70-01) | `Inbox`, `Archive` | reusar | Única fonte |
| Configuração dos canais | `internal/config` (70-01) | `NotificationsConfig` | reusar | Defaults e validação |
| Escrita atômica | `internal/queue` (`WriteJSONAtomic`) | Gravação segura | reusar | Cada notificação é um arquivo |
| Hook do Claude | `.claude/settings.local.json` (local) | Não há `settings.json` no scaffold | criar mesclagem segura | Adiciona o hook sem sobrescrever (`SEG-02`) |
| Comando de leitura | adaptadores `.claude/commands`, `.gemini/prompts`, `.github/prompts` | Padrão de adaptador fino | reusar | `/inbox` como adaptador; Gemini e agy sem hook usam `/inbox` |
| Comentário no card | MCP, só pelo agente | Go não alcança o MCP | reusar D-2 | `pm_comment` vira instrução gravada na notificação |

## Dependências novas
- Nenhuma (`net/smtp`, `net/http` e `os/exec` da stdlib) (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis
- `SEG-02`, `SEG-04`, `SEG-05`, `ARQ-10`, `COD-01`, `COD-03`, `TST-03`, `TST-04`.
- Rules: `sdd-seguranca.md`, `sdd-arquitetura.md`, `sdd-codigo-e-testes.md`.
- Gerais: `FLX-01`, `FLX-02`, `COD-02`, `COD-05`, `TST-05`.

## Escopo
1. `internal/notify`: `Dispatch` por canal. `discovery_session` grava `inbox/<timestamp>-<id>.json` (`tipo`, `titulo`, `mensagem`, `pr_url`, `lida=false`); com `desktop_toast` dispara toast (`osascript` no macOS, `notify-send` no Linux, PowerShell no Windows) por executor injetável, com argv explícito e texto sanitizado. `email` por SMTP e `webhook` por HTTP (Slack e genérico), com timeouts; URL e credenciais só por variável de ambiente; `http` só para loopback. `pm_comment` não é executado pelo Go: grava a instrução para o agente.
2. Falha de um canal opcional não impede os demais nem a gravação no inbox; o erro é registrado sem expor URL nem credencial.
3. `forge-sdd inbox [--unread] [--mark-read <id>|--all] [--archive] [--hook]`: lista com o link do PR e a mensagem "revisado, sem inconformidades e pronto para aprovação final"; `--hook` imprime formato compacto para o hook.
4. Mesclagem do hook no `.claude/settings.json` (`UserPromptSubmit` chamando `forge-sdd inbox --unread --hook`): cria se faltar, acrescenta sem tocar nas outras chaves, não duplica e nunca sobrescreve JSON inválido (reporta). A ligação ao `init` e `update` é feita na 70-07.
5. Mensagens limitadas a 2 KB; caracteres de controle e sequências de terminal removidos.

## Critérios de Aceitação Executáveis
```bash
go build ./... && go vet ./... && go test -race ./internal/notify/... ./internal/scaffold/... ./cmd/forge-sdd/...
```
- Inbox: grava, lista não lidas, marca lida, arquiva; a notificação aprovada traz o link do PR.
- Webhook com `httptest`, SMTP com servidor de teste em processo, toast com executor falso; canal desabilitado não dispara.
- Falha de e-mail não bloqueia inbox nem webhook.
- Violações (`TST-03`): URL `http` fora de loopback, variável de ambiente ausente, mensagem com ESC e quebra de linha, tamanho acima do limite, id de notificação com `..`.
- Hook: arquivo novo, arquivo existente com outras chaves preservadas, segunda execução idempotente, JSON inválido intocado.
- Nenhum segredo em arquivo de job, inbox nem log (teste procura o valor da variável de ambiente nas saídas).
