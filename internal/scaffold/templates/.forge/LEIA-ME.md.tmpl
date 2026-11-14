# Pipeline multi-sessão do Forge SDD (guia rápido)

Três sessões de terminal trabalham juntas: **você conversa numa, e outras duas constroem e revisam sozinhas.**

| Sessão | Quem faz | O que faz |
|---|---|---|
| `forge-discovery` | você + agente | conversa, refina a feature e dispara `/implementar` |
| `forge-builder` | worker | pega o job da fila, constrói em paralelo e abre o PR |
| `forge-reviewer` | worker | confere o PR contra spec, ADRs e rules; devolve ou aprova |

## Começar em 3 passos

**Primeiro passo em projeto novo: a Constituição (`/constitution`)**
Antes de qualquer discovery ou feature, o projeto precisa ter a sua missão, stack, módulos (`sdd/spec/modules.md`), fluxos (`sdd/spec/flows.md`) e decisões arquiteturais (`sdd/spec/decisions.md`) definidos. O menu do `forge` e o `forge-sdd doctor` alertam quando a constituição é rascunho e destacam `/constitution` como passo 1.

O comando determinístico de contexto:
```bash
forge-sdd brief           # imprime o resumo: missão, stack, módulos, fluxos e decisões
forge-sdd brief --check   # confere prontidão (sai 0 pronta, 3 rascunho)
```
Toda sessão de Discovery inicia executando `forge-sdd brief` para ler a missão e as decisões arquiteturais antes de propor novas mudanças.

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
| `gemini` | sim | `agy -p` | `GEMINI.md` |
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

## O que cada painel mostra e contingência de terminal

Com `tmux` instalado, `forge` e `forge-sdd up` abrem **uma tela com três painéis**, cada um com o nome da sessão no topo: `forge-discovery` (grande, à esquerda), `forge-builder` e `forge-reviewer` (empilhados à direita).

**Instalação do tmux e contingência de terminal:**
- Em modo `auto`, se o `tmux` não estiver no PATH, o Forge detecta o gerenciador de pacotes do sistema (Homebrew no macOS, apt-get, dnf, pacman, apk ou zypper no Linux) e pergunta se você deseja instalá-lo com confirmação interativa.
- Se você recusar ou a instalação falhar, o Forge cai com graça (**contingência do tmux**) para **janelas nativas do sistema**. No Windows com Windows Terminal (`wt.exe`), ele abre as 3 sessões em abas/painéis divididos do Windows Terminal; no Windows nativo sem `wt.exe`, abre janelas separadas de `cmd.exe`.
- No **WSL** (Windows Subsystem for Linux), o tmux funciona nativamente dentro da distribuição Linux.
- Você pode forçar a escolha com `forge-sdd up --terminal tmux|native|auto` ou definindo `"terminal": "tmux"` ou `"native"` em `forge.json`.

Dentro do tmux: `Ctrl+b` e as setas trocam de painel, `Ctrl+b d` solta a tela sem encerrar nada, e `tmux ls` mostra o nome da sessão (`forge-<projeto>-<6 hex do caminho>`) para reconectar com `tmux attach -t =<nome>`; rodar `forge` ou `forge-sdd up` de novo também reconecta.

**Identidade visual e alertas no tmux:**
- A barra inferior exibe o identificador visual `📦 <projeto>` e o nome da janela é padronizado como `forge:<projeto>`.
- As bordas dos painéis mostram `[ <projeto> ❯ <papel> ]` para você saber exatamente em qual projeto e estação está trabalhando.
- **Alertas visuais de aprovação:** quando um job aguarda revisão humana (ex.: PR aberto pelo Builder ou veredito do Reviewer), a borda do painel pisca com `[ 🔔 REVISÃO NECESSÁRIA ]` ou `[ 🔔 AÇÃO NECESSÁRIA ]`, voltando ao normal assim que a ação for destravada.

Nos painéis do Builder e do Reviewer o worker mostra o que o agente está fazendo, em linhas curtas (a saída completa fica em `.forge/runtime/logs/`):

```
▶ job auth-oauth · feature_build · tentativa 1 · codex · permissões project
  ✎ src/domain/unit.entity.ts
  $ npm test
  … trabalhando há 2m10s · 14 arquivos alterados
■ concluído em 4m12s · 14 arquivos alterados
```

As linhas vêm dos eventos do agente (Claude, Gemini e Codex) e, para qualquer agente, do acompanhamento do `git status`; segredos em comandos aparecem mascarados. `forge-sdd worker --raw` mostra a saída bruta.

**Limite de uso:** se o agente avisa que bateu o limite (como o de 5 horas do Claude), o job volta à fila sem gastar tentativa, o painel mostra `⏸ limite de uso do claude: retomo às 22:40`, o inbox avisa e o worker retoma sozinho (espera de até 6 horas; um mesmo job é devolvido no máximo 3 vezes, e limite semanal ou com data faz o job falhar com o motivo).

**A Discovery não executa o build.** Em modo pipeline, o `/split-features` termina em `forge-sdd dispatch` e entrega ao Builder; `queue next` e `queue complete` só funcionam sob um worker (ou com `--local`, por sua conta). Com `pm_mcp.require_existing_card: false` o despacho não pede card nem evidência.

## O que acontece depois do `/implementar`

1. **Gate pré-build:** confere spec, tasks, ADRs, rules e a evidência de que o card existe no seu gerenciador (ClickUp, GitHub Projects, Azure DevOps ou Jira, via MCP). Sem card, o despacho é bloqueado e a lista de pendências aparece.
2. **Builder:** divide em tarefas independentes, roda build, lint e testes e abre o PR.
3. **Reviewer:** não conforme → devolve ao Builder com feedback (até `max_rework_attempts`, padrão 3); depois disso, ou em dúvida, chama você.
4. **Conforme:** você é avisado na sessão `forge-discovery` (`/inbox` mostra de novo), com toast, e-mail ou webhook se estiverem ligados.

Você decide e aprova o PR. O pipeline nunca mescla nada.

## Modo Live e Despachante de Entrega

Por padrão, os workers operam em modo `headless` (executam uma nova chamada de comando da CLI a cada job). Com `session_mode: live` e `tmux`:

*   **CLIs Abertas nos Painéis:** Builder e Reviewer mantêm suas sessões interativas abertas continuamente na janela principal do tmux, sem pagar o custo de boot a cada job.
*   **Despachante Oculto (`forge-ctl`):** uma janela de controle oculta entrega os prompts aos painéis usando buffers raw do tmux (`load-buffer` + `paste-buffer -p` + `Enter`), garantindo integridade total de aspas e quebras de linha.
*   **Limpeza de Contexto (`context: clear`):** envia o comando de reset do perfil (`/clear` ou `/new`) entre jobs e espera a ociosidade do painel antes do próximo prompt. Para manter o histórico de raciocínio, use `context: continuous`.
*   **Falhas Lidas da Tela:** o despachante inspeciona a tela em tempo real para detectar limites de cota (`DetectUsageLimit` → pausa automática) ou indisponibilidade de conta/tier (`DetectAgentUnavailable` → aviso no inbox).
*   **Queda Suave:** se `live` estiver ativado sem tmux ou em terminal nativo, o pipeline recai suavemente para `headless` com um aviso no terminal.

## Monitor de Processos e Consumo (`forge-sdd ps`)

Para ter visibilidade total sobre o que está rodando no computador relacionado ao projeto:
```bash
forge-sdd ps            # exibe tabela com PID, papel, CPU %, RAM (MB), tempo e estado
forge-sdd ps --json     # exporta em JSON para dashboards e automações
forge-sdd ps --kill-zombies  # encerra processos orfãos ou presos de sessões anteriores
```
O monitor também está disponível no menu interativo (`forge`) sob a opção de diagnóstico.

## Paralelismo de Filas por DAG

O pipeline suporta dependências explícitas entre tarefas na fila (`**Depende de:** task-1, task-2`).
- As tarefas são resolvidas respeitando o grafo acíclico dirigido (DAG), impedindo que tarefas bloqueadas iniciem antes da conclusão de suas dependências.
- `forge-sdd queue next --batch <max>` extrai até N tarefas independentes prontas para execução em paralelo simultâneo no Builder.

## Configuração

- `.forge/config/forge.json`: agente de cada sessão (`sessions`), modo de sessão (`session_mode`: `live` ou `headless`), limpeza de contexto (`context`: `clear` ou `continuous`), nível de permissão, tarefas em paralelo, worktrees, tentativas de rework, status do card. Copie de `forge.json.example` se quiser recomeçar.
- `.forge/config/notifications.json`: canais (inbox sempre ligado; toast, e-mail, webhook e comentário no card opcionais). URLs e senhas **só por variável de ambiente** (`url_env`, `password_env`), nunca no arquivo.
- **Permissões do agente:** escolha o nível `project` ou `full` (veja acima, no menu ou em `permissions` de `forge.json`); só edite `cli_profiles` para ajustes finos.
- Sem `forge.json`, o `/implementar` volta ao fluxo de sessão única da 4.0.0-beta.

## Quando algo não anda

| Sintoma | O que fazer |
|---|---|
| Job parado em `processing` (worker caiu) | 15 minutos depois (o prazo do lease), `forge-sdd bus recover --role builder` (ou `reviewer`) devolve ao `pending` |
| Ver a fila | `forge-sdd bus list --role builder` |
| O que o worker fez | o painel do Builder (ou do Reviewer) e os logs completos em `.forge/runtime/logs/` |
| Painel parado com `⏸` | limite de uso do agente: ele retoma sozinho no horário mostrado (ou `forge-sdd bus requeue` para devolver um job à mão) |
| `/implementar` bloqueado | leia a lista de pendências: falta card, ADR, rule ou evidência do card |
| Avisos não aparecem | `forge-sdd inbox` mostra as notificações pendentes |

O estado de execução (filas, logs, worktrees) fica em `.forge/runtime/`, fora do git.

## Limites desta versão (beta)

Nenhuma das CLIs foi testada de ponta a ponta com o worker (só com agente simulado e montagem de comandos); confirme os argumentos no seu ambiente. O `agy` continua aceito se já estiver em `forge.json`. Detalhes e decisões: `sdd/spec/adr/adr-0004-pipeline-multi-sessao.md`.
