# Pipeline multi-sessão do Forge SDD (v5.0.0)

> Este guia também é instalado em cada projeto, em `.forge/LEIA-ME.md`, e aparece resumido ao final de `forge-sdd init`, `forge-sdd update` e `forge-sdd pipeline init`.

**Instalar a versão com o pipeline:** `npm install -g @nathanramorim/forge-sdd@beta`. O dist-tag `latest` ainda aponta para a versão estável anterior.

Três sessões de terminal trabalham juntas: **você conversa numa, e outras duas constroem e revisam sozinhas.**

| Sessão | Quem faz | O que faz |
|---|---|---|
| `forge-discovery` | você + agente | conversa, refina a feature e dispara `/implementar` |
| `forge-builder` | worker | pega o job da fila, constrói em paralelo e abre o PR |
| `forge-reviewer` | worker | confere o PR contra spec, ADRs e rules; devolve ou aprova |

## Começar em 3 passos

**Primeiro passo em projeto novo: a Constituição (`/constitution`)**
Antes de qualquer discovery ou feature, o projeto precisa ter a sua missão, stack, módulos (`sdd/spec/modules.md`), fluxos (`sdd/spec/flows.md`) e decisões arquiteturais (`sdd/spec/decisions.md`) definidos. O menu do `forge` e o `forge-sdd doctor` alertam quando a constituição é rascunho e destacam o comando `/constitution` como passo 1.

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

## O que cada painel mostra, identificação do projeto e contingência de terminal

Com `tmux` instalado, `forge` e `forge-sdd up` abrem uma tela limpa, encantadora e com foco total no projeto:
- **Identificação Imediata:** A barra inferior exibe `[ 📦 <projeto> │ <branch> ]`, o nome da janela é `forge:<projeto>` e as bordas de cada painel estampam `[ <projeto> ❯ forge-discovery ]`, impossibilitando qualquer confusão mesmo ao alternar entre múltiplos projetos abertos.
- **Terminal Livre & Mini-Monitor:** Em repouso, o terminal direito fica desobstruído para o desenvolvedor trabalhar sem loops barulhentos de espera. O monitoramento da esteira resume o status passivo e se acende com alertas visuais amigáveis sempre que uma aprovação humana for demandada.
- **Visibilidade de Processos e Consumo:** Com `forge-sdd ps`, você inspeciona a qualquer instante todos os workers, CLIs de IA e testes rodando em background com consumo de CPU, memória RAM e tempo ativo.
- **Três Painéis Principais:** `forge-discovery` (amplo, à esquerda), `forge-builder` e `forge-reviewer` (à direita).

**Instalação do tmux e contingência de terminal:**
- Em modo `auto`, se o `tmux` não estiver no PATH, o Forge detecta o gerenciador de pacotes do sistema (Homebrew no macOS, apt-get, dnf, pacman, apk ou zypper no Linux) e pergunta se você deseja instalá-lo com confirmação interativa.
- Se você recusar ou a instalação falhar, o Forge cai com graça (**contingência do tmux**) para **janelas nativas do sistema**. No Windows com Windows Terminal (`wt.exe`), ele abre as 3 sessões em abas/painéis divididos do Windows Terminal; no Windows nativo sem `wt.exe`, abre janelas separadas de `cmd.exe`.
- No **WSL** (Windows Subsystem for Linux), o tmux funciona nativamente dentro da distribuição Linux.
- Você pode forçar a escolha com `forge-sdd up --terminal tmux|native|auto|painel` ou definindo `"terminal": "tmux"` ou `"native"` em `forge.json`.
- **Modo painel** (`--terminal painel`): nenhuma janela nem tmux. Builder e Reviewer rodam em segundo plano, com a saída em `.forge/runtime/logs/<papel>-worker.log`, e você acompanha pelo Claude Code com o mod `forge-painel` (instale com `forge-sdd skills instalar plugin forge-painel`, o mod vem no binário; o `forge-sdd up` imprime o comando com `--plugin-dir`). Fechar o Claude não para os workers.

Dentro do tmux: `Ctrl+b` e as setas trocam de painel, `Ctrl+b d` solta a tela sem encerrar nada, e `tmux ls` mostra o nome da sessão (`forge-<projeto>-<6 hex do caminho>`) para reconectar com `tmux attach -t =<nome>`; rodar `forge` ou `forge-sdd up` de novo também reconecta.

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

## Trabalho longo com `/toca-ai`

`/toca-ai <objetivo>` na sessão Discovery: o Líder gera o plano e as specs de todas as etapas, junta num pré-voo único o que só você pode dar (permissão de comando, ferramenta, arquivo e consentimento) e entrega à esteira com `forge pipeline toca-ai <iniciativa>`. O chat fica livre.

- **Você é interrompido** só no pré-voo (`forge pipeline toca-ai --preflight <iniciativa>` mostra a lista única; consentimento se responde com `forge pipeline consent <id> --iniciativa <iniciativa>`, só em terminal) e para ação irreversível fora do repositório, segredo, gasto externo, merge em `main` ou exceção a regra. Imprevisto vira etapa `blocked` com o item exato em `pendencias-humanas.md`; as independentes seguem.
- **Retomar:** rode `/toca-ai` ou `forge pipeline toca-ai <iniciativa>` de novo; os jobs não se duplicam. Acompanhe com `forge status`; pare com `.forge/runtime/STOP`.
- **Sem o modo avançado**, o comando exige a spec aprovada por você (Gate 1) antes de enfileirar; com ele, o Reviewer aprova a spec e o Go libera as etapas.

## Configuração

- `.forge/config/forge.json`: agente de cada sessão (`sessions`), modo de sessão (`session_mode`: `live` ou `headless`), limpeza de contexto (`context`: `clear` ou `continuous`), nível de permissão, tarefas em paralelo, worktrees, tentativas de rework, status do card. Copie de `forge.json.example` se quiser recomeçar.
- `.forge/config/notifications.json`: canais (inbox sempre ligado; toast, e-mail, webhook e comentário no card opcionais). URLs e senhas **só por variável de ambiente** (`url_env`, `password_env`), nunca no arquivo.
- **Permissões do agente:** escolha o nível `project` ou `full` (veja acima, no menu ou em `permissions` de `forge.json`); só edite `cli_profiles` para ajustes finos.
- **Card:** `pm_mcp.require_existing_card` nasce `false`; `forge pipeline init` pergunta no terminal e grava a resposta. Para alternar depois, use o item **Settings** do menu (`ativado`/`desativado`).
- **Modo avançado (experimental, ADR-0013):** `autonomy: advanced` faz o Reviewer revisar a spec no lugar da aprovação humana (job `spec_review`, parecer em `review-spec.md`, aplicado pelo Go). Exige `permissions: full` e `isolated_confirmed: true`; use só em contêiner, VM ou máquina descartável. Liga com → sobre "Iniciar esteira (up)" no menu, desliga com ←, ou por `forge pipeline autonomy advanced --isolated` / `normal`. Parecer `ajustar` volta à Discovery até `reviewer.max_rework_attempts`; `escalar` ou o limite marcam as tasks `blocked` (as independentes seguem). Salvaguardas (todas desligadas por padrão; ligam em Settings): `safeguards.max_job_tokens` e `safeguards.max_job_minutes` (ao passar, a esteira para com motivo no inbox), `safeguards.max_jobs_per_day` ( conta jobs distintos no dia local e, ao atingir o teto, a esteira **pausa até o dia seguinte** e retoma sozinha). `builder.job_timeout_minutes` segue só como timeout do runner, em todos os modos. Sem teto ligado, só o arquivo `.forge/runtime/STOP`, que para a esteira de forma limpa. `forge doctor` avisa se estiver ligado sem ambiente isolado confirmado.
- Sem `forge.json`, o `/implementar` volta ao fluxo de sessão única da 4.0.0-beta.

## Esteira econômica (ADR-0014)

- **PR pelo Go:** o Builder faz commit e roda `forge pipeline pr --id <job> --body-file <arq>`; o Go roda a Camada A, faz `git push origin <branch>:<branch>`, abre o PR (ou corrige a base de um PR já aberto) e o registra. Recusa branch `main` ou igual à base, corpo fora do `.github/pull_request_template.md` e branch sem commit à frente. `forge pipeline pr-opened` fica para quem abre o PR à mão.
- **Base da iniciativa:** `forge queue enqueue --base <branch>` grava `base_branch` na fila. A base vale nesta ordem: flag, fila, origem da branch no git ("Created from") e `base_branch` do `forge.json`; `dispatch`, `/toca-ai` e `pr` dizem de onde ela veio.
- **Fix no despacho:** `forge dispatch <iniciativa> --fix fix-88-01` confere a spec do fix, enfileira o job com a branch `fix/88-01-<nome>` e base na branch da iniciativa, e põe o fix na fila.
- **Uso real:** o worker lê o evento `result` do log do `claude` e grava tokens, cache, turnos, duração, modelo e custo de tabela (`medicao: real`); sem `result`, estima (`medicao: estimado`). `forge report --initiative <id> [--jobs]` soma separando real e estimado.
- **Status curto:** sem terminal, `forge status` imprime até 1,5 KB (só iniciativas ativas); `--full` mostra tudo.
- **Archive:** `forge archive` move a pasta da iniciativa para `sdd/archive/discovery/` (só na branch dela e sem task pendente; rodar de novo não muda nada). Ao aprovar o último job, só a fila `origem: toca-ai` arquiva sozinha; nas demais o inbox avisa "pronta para arquivar". `forge archive --concluidas --dry-run` lista as candidatas.
- **Modelo e esforço por sessão:** `models` e `efforts` em `forge.json` (por `discovery`, `builder`, `reviewer`) viram `--model` e `--effort` pelos marcadores `{model}` e `{effort}` do perfil (claude e agy; gemini só modelo; as demais CLIs ainda sem flag confirmada). Escolha em **Settings**; sem escolha, o comando é o de sempre. Perfil gravado sem os marcadores gera aviso no preflight.
- **Modelo por agente:** `agent_models` em `forge.json` (por `discovery`, `builder`, `reviewer`) grava o `model:` no cabeçalho de `.claude/agents/*.md` ao rodar `forge update` (`--dry-run` mostra sem gravar). discovery vale para pm-architect e planner; reviewer para reviewer e governance. O valor deve ser texto simples (começa por letra; `null`, `true`, `yes` e similares são recusados). Codex e Agy ainda não.

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
