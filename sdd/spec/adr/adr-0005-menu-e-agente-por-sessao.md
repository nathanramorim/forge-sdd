# ADR-0005: Menu Navegável e Agente por Sessão do Pipeline

- **Data:** 2026-11-14
- **Status:** `aprovado`
- **Regras afetadas (IDs da constitution):** `ARQ-06` (o menu não é subcomando novo; `up` e `worker` ganham flags), `DUP-01`, `SEG-03`, `SEG-04`, `GOV-01` (este ADR; nenhum ID foi criado ou removido)
- **Autor / Responsável:** Líder (decisão do dono do produto no discovery 71)

## Contexto

A 4.1.0-beta abria as três sessões com uma CLI só (`claude` ou `agy`), os comandos eram muitos para decorar e `forge` sem argumentos caía direto em `init` ou `update`. Quem usa Gemini, Copilot ou Codex não conseguia iniciar o fluxo, e misturar agentes por sessão exigia editar JSON. Referência: `sdd/discovery/71-menu-interativo-e-agente-por-sessao/` e `feat-71-01` a `feat-71-05`.

## Decisão

- **D-1 Perfil com dois modos.** `CLIProfile` ganha `interactive_args` (Discovery) ao lado de `args` (worker); marcadores `{agent}`, `{name}`, `{prompt}` e `{permissions}` (este expande em zero ou mais argumentos e só vale como argumento inteiro). Argv explícito, nunca shell (`SEG-03`).
- **D-2 `sessions` por papel.** `forge.json` ganha `sessions` (`discovery`, `builder`, `reviewer`), `permissions` e `allow_commands`. `CLIFor(papel)` é o único ponto de resolução: `sessions[papel]` ou `cli`. Um `forge.json` da 4.1 continua válido e escolhe o mesmo agente nas três sessões (testado na carga e na resolução; o launcher e o worker foram testados com configurações novas, não com um arquivo da 4.1); o `interactive_args` que ele não traz é completado pelo padrão da CLI.
- **D-3 Perfis pesquisados na documentação oficial** (tabela abaixo). **A confirmar:** nenhum foi executado de ponta a ponta com o worker; os testes cobrem a montagem do comando.
- **D-4 Permissões em dois níveis**, tabela única em `internal/config` (`PermissionArgs`): `project` (padrão) libera edição e os comandos de `allow_commands` onde a CLI aceita allowlist; `full` usa o bypass documentado de cada CLI, só para ambiente isolado. O menu mostra o que cada nível libera antes de abrir.
- **D-5 Escolha lembrada.** O seletor grava `sessions` em `forge.json` (`SaveForgeConfig`, escrita atômica) só depois da confirmação e só se a checagem não tiver bloqueio.
- **D-6 `--cli` sempre explícito no worker.** O launcher gera `forge-sdd worker --role <papel> --name <sessão> --cli <agente>`. Na 4.1 o worker lia o `cli` de `forge.json` mesmo quando a Discovery tinha sido aberta com `--cli agy`, e as três sessões podiam usar agentes diferentes sem aviso.
- **D-7 Menu é o mesmo código do subcomando.** Cada ação executa o `RunE` do subcomando (`rootCmd.Find` + `ParseFlags` com as flags zeradas), sem subprocesso e sem lógica duplicada. O menu roda uma ação por vez e não é painel ao vivo. Sem terminal interativo, `forge` imprime a ajuda curta e sai com 0.
- **D-8 `Preflight`.** Antes de abrir: CLI no PATH (bloqueio), arquivos de instrução do agente (aviso, com `forge-sdd update --agent <x>`), definições de agente do Claude usadas pela sessão (bloqueio) e `forge.json` válido (bloqueio). Só usa `LookPath` e leitura de arquivos; não executa a CLI.
- **D-9 `AGENTS.md` para o Codex.** O Codex lê `AGENTS.md`, não `OPENAI.md`; o scaffold do agente `openai` passa a gerar `AGENTS.md` (conteúdo) e `OPENAI.md` vira ponteiro. Um `AGENTS.md` do usuário nunca é sobrescrito nem removido pelo `destroy` (só o gerado, identificado pelo cabeçalho `# AGENTS.md — `).

### Perfis por CLI

| CLI | Worker | Discovery | Agente nomeado | Nome da sessão | Instruções |
|---|---|---|---|---|---|
| `claude` | `claude -p {prompt} --agent {agent} -n {name}` | `claude -n {name} --agent {agent}` | `--agent` | `-n` | `CLAUDE.md` |
| `gemini` | `gemini -p {prompt}` | `gemini -i {prompt}` | sem flag: papel no prompt | só título ANSI | `GEMINI.md` |
| `copilot` | `copilot -p {prompt} -n {name} --no-ask-user` | `copilot -i {prompt} -n {name}` | `--agent` existe, pastas não confirmadas: papel no prompt | `-n` | `.github/copilot-instructions.md`, `AGENTS.md` |
| `codex` | `codex exec {permissions} {prompt}` | `codex {prompt}` | sem flag: papel no prompt | só título ANSI | `AGENTS.md` |

`-p` é o prompt em `claude`, `gemini` e `copilot`, mas perfil no `codex` (o modo não interativo dele é `codex exec`).

## Consequências

- **Positivas:** o pipeline abre com o agente que cada pessoa tem, cada sessão pode usar um diferente, e `forge` vira a porta de entrada para os dois fluxos (pipeline ou estrutura padrão), sem decorar comando.
- **Negativas / Trade-offs:** as flags de Gemini, Copilot e Codex dependem da documentação e podem mudar; no Gemini o nível `project` libera só edição (comandos de shell pedem confirmação e travam o worker), então o Builder com Gemini precisa de `full`; o nível `full` desliga a confirmação da CLI; o `agy` sai do seletor mas continua aceito.
- **Desvio da spec 71-03:** o aviso "agente diferente do declarado em `agents` do `.sddrc`" não foi implementado; a ausência do arquivo de instrução do agente já cobre o caso.
- **Ações de acompanhamento:** executar cada CLI de ponta a ponta com o worker e confirmar as flags; ver `sdd/discovery/71-menu-interativo-e-agente-por-sessao/status.md` (backlog do uso real: abrir por tmux, visão limpa de cada sessão, entrega do split ao Builder, motivo legível de falha).
