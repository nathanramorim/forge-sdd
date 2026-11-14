# Discovery 71 — Menu Navegável e Agente por Sessão do Pipeline (v4.2.0)

Status: `FLX-01` aprovado em 2026-11-13 (decisões 1 a 4 aceitas); fatiado em 5 features. Nenhum código alterado.

## Problema
Na 4.1.0-beta o pipeline abre as três sessões com **uma única CLI para todas** (`--cli` ou `forge.json`), só `claude` ou `agy`, e quem usa Gemini, Copilot ou Codex não consegue iniciar o fluxo. Além disso:
1. **Não há como escolher agente por sessão.** Quem quer `claude` na Discovery, `codex` no Builder e `gemini` no Reviewer precisa editar JSON e conhecer flags de cada CLI.
2. **Há comando demais para decorar.** `forge-sdd` tem dezenas de subcomandos; `forge` sem argumentos hoje cai direto em `init` ou `update`, sem mostrar o que mais existe.
3. **Os dois fluxos se misturam.** O pipeline (três sessões) e a estrutura padrão (sessão única) são caminhos diferentes, mas nada ajuda a escolher entre eles.

## Para quem
- **Pessoa que conduz o produto:** quer rodar `forge`, escolher com as setas, definir qual agente fica em cada sessão e abrir tudo, sem decorar comando nem editar arquivo.
- **Time com agentes diferentes:** cada pessoa usa o agente que tem licença ou prefere; o pipeline não pode prender todo mundo a uma CLI.

## Decisões já tomadas (Clarify, 2026-11-13)
| Tema | Decisão |
|---|---|
| Escolha do agente | **Perguntar e lembrar:** o menu pergunta a CLI de cada sessão com a última escolha já marcada e grava em `forge.json` para o próximo uso. |
| Agentes oferecidos | **Claude, Gemini, Copilot e Codex.** Flags de modo não interativo pesquisadas na documentação oficial; o que não for testado fica marcado "a confirmar" e editável. O `agy` continua aceito se já estiver em `forge.json`, mas não aparece no seletor. |
| Base | `feat/70` (a 4.1.0-beta só existe nela), em worktree próprio `../forge-sdd-71`. |
| Menu | O comando `forge` sem argumentos abre o menu, **só para escolher e executar**: não é painel de acompanhamento (a 4.0.0-beta removeu a interface de acompanhamento de propósito). |

## Proposta (o "como" macro)
**1. Menu em dois caminhos.** `forge` (terminal interativo) mostra:
- **Pipeline multi-sessão:** iniciar as 3 sessões (escolhendo o agente de cada uma), só ver os comandos de abertura, ativar ou reconfigurar (`pipeline init`), ver fila e inbox, recuperar jobs presos, diagnóstico (CLI instalada, arquivos do agente, `forge.json` válido) e abrir o guia.
- **Estrutura padrão (sessão única):** status das iniciativas, painel da fila, gate (Camada A), doctor, instalar ou atualizar a estrutura e "ver todos os comandos".
- Num projeto sem Forge, a primeira opção é instalar. Setas e Enter navegam, Esc volta um nível (no menu principal, sai) e a opção Sair encerra. Sem terminal interativo (CI, pipe) o comando imprime a ajuda curta e sai, sem travar.

**2. Seletor de agente por sessão.** Três perguntas (`forge-discovery`, `forge-builder`, `forge-reviewer`), cada uma com Claude, Gemini, Copilot e Codex; a última escolha vem marcada, e as CLIs que não estão no PATH aparecem com aviso. Na linha de comando, o mesmo vale sem menu: `forge-sdd up --discovery claude --builder codex --reviewer gemini`, para scripts.

**3. Perfil por agente.** `forge.json` ganha `sessions` (o agente de cada sessão) e `cli_profiles` ganha perfis para `gemini`, `copilot` e `codex`, cada um com o comando não interativo (worker) e o comando interativo (Discovery). O worker passa a usar o perfil da **sua** sessão, não um perfil único.

**4. Checagem antes de abrir.** Antes de abrir os terminais: a CLI escolhida existe no PATH e o projeto tem os arquivos que ela lê (`CLAUDE.md`, `GEMINI.md`, `AGENTS.md`, `.github/copilot-instructions.md`, definições de agente). Faltando, o menu oferece `forge-sdd update --agent <x>` e só abre depois.

**5. Reaproveitar o que existe.** O menu usa o `huh` que o `init` já usa (sem dependência nova), chama as mesmas funções dos subcomandos (sem duplicar lógica) e o launcher continua sendo `internal/pipeline` (`forge-up.sh` e `.ps1` seguem finos).

## Escopo
**Dentro:** `forge` sem argumentos (menu), seletor por sessão, `sessions` e novos perfis em `forge.json`, flags de `up`, checagem de CLI e arquivos do agente, worker por perfil da sessão, `.forge/LEIA-ME.md` e docs, ADR 0005, release notes 4.2.0.

**Fora:** painel ou interface de acompanhamento ao vivo; mudar o fluxo do Reviewer, do Builder ou do barramento; criar novos agentes no scaffold além do que `forge-sdd update --agent` já instala; aprovar ou mesclar PR pelo agente.

## Jornada alvo
1. `forge` → menu: **Pipeline multi-sessão**.
2. **Iniciar as 3 sessões** → Discovery: `claude` (marcado da última vez) → Builder: `codex` → Reviewer: `gemini`.
3. Resumo: "Discovery claude · Builder codex · Reviewer gemini. Codex não está instalado: [instalar guia] [escolher outro]". Escolheu outro, confirma.
4. Os três terminais abrem; a escolha fica salva para a próxima vez.
5. O menu mostra, a qualquer hora, fila e inbox, sem comando para decorar.

## Riscos
| Risco | Mitigação |
|---|---|
| Flags não interativas de Gemini, Copilot e Codex mudam ou não foram testadas | Perfis em configuração, pesquisados na documentação oficial, marcados "a confirmar" e cobertos por teste de montagem de comando, não de execução (como o `agy` na 4.1). |
| O worker precisa de permissão para editar e rodar comandos sem confirmar, e cada CLI chama isso de um jeito | O perfil traz a flag mais restrita que funciona e o menu avisa o que ela libera; nada é liberado sem o usuário ver. |
| Menu quebra em terminal sem TTY ou no Windows | Detecção de TTY, saída limpa sem terminal e testes com entrada simulada. |
| Menu duplica lógica dos subcomandos | Cada ação chama a função já existente do subcomando; teste garante que o menu e o comando fazem o mesmo. |
| Escolha de agente sem os arquivos dele no projeto (Codex sem `AGENTS.md`) | Checagem prévia e oferta de `forge-sdd update --agent`. |
| Quebrar quem usa `forge` sem argumentos para instalar | Instalar vira a primeira opção do menu num projeto sem Forge; `forge init` e `forge update` continuam iguais. |

## Decisões em aberto para o Gate 1
1. **`forge` sem argumentos:** o menu substitui o `init`/`update` automático (que ficam como opções do menu). Aceita?
2. **Versão:** `4.2.0-beta` (minor), reunindo a mensagem pós-instalação e o guia da `4.1.0-beta`. Aceita?
3. **Permissões do worker:** o perfil de cada CLI traz a flag de permissão documentada e o menu mostra o que ela libera antes de abrir. Aceita, ou prefere começar sem flag e mandar o usuário configurar?
4. **`agy`:** continua funcionando se estiver em `forge.json`, mas sai do seletor. Aceita?
