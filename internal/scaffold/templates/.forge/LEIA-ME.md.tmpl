# Pipeline multi-sessão do Forge SDD (guia rápido)

Três sessões de terminal trabalham juntas: **você conversa numa, e outras duas constroem e revisam sozinhas.**

| Sessão | Quem faz | O que faz |
|---|---|---|
| `forge-discovery` | você + agente | conversa, refina a feature e dispara `/implementar` |
| `forge-builder` | worker | pega o job da fila, constrói em paralelo e abre o PR |
| `forge-reviewer` | worker | confere o PR contra spec, ADRs e rules; devolve ou aprova |

## Começar em 3 passos

**Jeito mais fácil: rode `forge` e navegue pelo menu** (setas, Enter, Esc volta). Em **Pipeline multi-sessão → Iniciar as 3 sessões** você escolhe o agente de cada sessão (Claude, Gemini, Copilot ou Codex), o nível de permissão, vê a checagem (CLI instalada, arquivos do agente) e confirma. A escolha fica lembrada para a próxima vez.

Sem menu (scripts):

```bash
forge-sdd pipeline init                                  # 1. ativa o modo (cria .forge/config/forge.json)
forge-sdd up --discovery claude --builder codex --reviewer gemini   # 2. abre as 3 sessões
# 3. na sessão forge-discovery:
/implementar nome-da-feature
```

`forge-sdd up --cli <agente>` usa o mesmo agente nas três; sem flags vale o que está em `sessions` de `forge.json`. Cada sessão pode usar um agente diferente.

| Agente | Sessão interativa (Discovery) | Worker (Builder e Reviewer) | Instruções que lê |
|---|---|---|---|
| `claude` | sim | `claude -p` | `CLAUDE.md` |
| `gemini` | sim | `gemini -p` | `GEMINI.md` |
| `copilot` | sim | `copilot -p` | `.github/copilot-instructions.md`, `AGENTS.md` |
| `codex` | sim | `codex exec` | `AGENTS.md` (gerado pelo `forge-sdd update --agent openai`) |

Gemini, Copilot e Codex seguem a documentação oficial das CLIs, mas **ainda não foram executados de ponta a ponta com o worker**: os perfis estão em `cli_profiles` de `forge.json` e podem ser ajustados.

**Permissões:** o worker roda o agente sem você aprovar cada passo. `project` (padrão) libera edição e só os comandos listados em `allow_commands` (`git`, `gh`, `forge-sdd`); `full` usa o bypass de cada CLI e só deve ser usado em ambiente isolado. No Gemini, o nível `project` libera só edição: para rodar testes no Builder é preciso `full`.

Builder e Reviewer mostram `Aguardando jobs na fila...` até chegar trabalho. Para só ver os comandos de abertura, sem abrir nada: `forge-sdd up --dry-run`. Os scripts `.forge/forge-up.sh` (macOS, Linux, WSL, Git Bash) e `.forge/forge-up.ps1` (Windows) fazem o mesmo.

Sem terminal automático? Abra você mesmo, um terminal por papel:

```bash
forge-sdd worker --role builder
forge-sdd worker --role reviewer
```

## O que acontece depois do `/implementar`

1. **Gate pré-build:** confere spec, tasks, ADRs, rules e a evidência de que o card existe no seu gerenciador (ClickUp, GitHub Projects, Azure DevOps ou Jira, via MCP). Sem card, o despacho é bloqueado e a lista de pendências aparece.
2. **Builder:** divide em tarefas independentes, roda build, lint e testes e abre o PR.
3. **Reviewer:** não conforme → devolve ao Builder com feedback (até `max_rework_attempts`, padrão 3); depois disso, ou em dúvida, chama você.
4. **Conforme:** você é avisado na sessão `forge-discovery` (`/inbox` mostra de novo), com toast, e-mail ou webhook se estiverem ligados.

Você decide e aprova o PR. O pipeline nunca mescla nada.

## Configuração

- `.forge/config/forge.json`: agente de cada sessão (`sessions`), nível de permissão, tarefas em paralelo, worktrees, tentativas de rework, status do card. Copie de `forge.json.example` se quiser recomeçar.
- `.forge/config/notifications.json`: canais (inbox sempre ligado; toast, e-mail, webhook e comentário no card opcionais). URLs e senhas **só por variável de ambiente** (`url_env`, `password_env`), nunca no arquivo.
- **Permissões do agente:** o worker roda o agente sem interação. O perfil padrão do `claude` não traz flags de permissão: ajuste `cli_profiles` em `forge.json` para ele poder editar arquivos e rodar comandos sem esperar você.
- Sem `forge.json`, o `/implementar` volta ao fluxo de sessão única da 4.0.0-beta.

## Quando algo não anda

| Sintoma | O que fazer |
|---|---|
| Job parado em `processing` (worker caiu) | 15 minutos depois (o prazo do lease), `forge-sdd bus recover --role builder` (ou `reviewer`) devolve ao `pending` |
| Ver a fila | `forge-sdd bus list --role builder` |
| O que o worker fez | logs em `.forge/runtime/logs/` |
| `/implementar` bloqueado | leia a lista de pendências: falta card, ADR, rule ou evidência do card |
| Avisos não aparecem | `forge-sdd inbox` mostra as notificações pendentes |

O estado de execução (filas, logs, worktrees) fica em `.forge/runtime/`, fora do git.

## Limites desta versão (beta)

Nenhuma das CLIs foi testada de ponta a ponta com o worker (só com agente simulado e montagem de comandos); confirme os argumentos no seu ambiente. O `agy` continua aceito se já estiver em `forge.json`. Detalhes e decisões: `sdd/spec/adr/adr-0004-pipeline-multi-sessao.md`.
