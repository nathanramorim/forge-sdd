# Spike de aninhamento (feat-104-06)

Data: 2026-11-20. Só verificações locais e de documentação: nenhum modelo foi chamado e nenhum agente foi executado. Versão: `claude` 2.1.329 (`claude --version`).

## Comandos e saídas literais
```
$ claude --help | grep -i -E "agent|team|tools"
  --agent <agent>        Agent for the current session. Overrides the 'agent' setting.
  --agents <json-or-file>  JSON object defining custom agents ...
  --allowedTools, --allowed-tools <tools...>
  --disallowedTools, --disallowed-tools <tools...>
  --forward-subagent-text  Forward subagent text and thinking ...

$ strings <claude 2.1.329> | grep -c <termo>
CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS: 6      TeamCreate: 2      SendMessage: 78

$ strings <claude 2.1.329> | grep -o -i -E "...teammates cannot spawn..."
"Teammates cannot spawn other teammates — the team roster is flat. To spawn a subagent instead, omit the `name` parameter."
"In-process teammates cannot spawn background agents. Use run_in_background=false for synchronous subagents."
"`name` is unavailable here — teammates cannot spawn teammates."

$ grep -n "^tools:" .claude/agents/*.md
builder.md / lider.md / pm-architect.md / migrator.md: Read, Grep, Glob, Bash, Edit, Write
reviewer.md / governance.md: Read, Grep, Glob, Bash
planner.md / archivist.md: Read, Grep, Glob, Edit, Write
```

## Confirmado (evidência direta)
- Nenhum agente de `.claude/agents/` lista a ferramenta de criar subagente nas suas `tools`.
- O binário 2.1.329 contém `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS`, `TeamCreate` e `SendMessage`, ou seja, agent teams existe como recurso experimental.
- O texto das mensagens do binário diz que um teammate não cria outro teammate ("the team roster is flat") e que teammates em processo não criam agentes em segundo plano; a saída indicada é criar um subagente síncrono.

## Suposição (não provada por execução)
- Um subagente comum não cria subagente: é o que a lista de ferramentas dos papéis implica, mas isso é configuração nossa, não prova de que a plataforma bloqueia.
- Um teammate pode criar subagente síncrono (inferido da mensagem "To spawn a subagent instead, omit the name parameter"), com um nível de profundidade.

## Não verificado
- Execução real de subagente dentro de subagente e de teammate (exige chamar modelo; fora do escopo).
- Documentação oficial online, o comportamento em versões posteriores a 2.1.329, o Codex, os subagentes do Agy e do Gemini, e hooks no app desktop.

## Efeito no desenho
Nada na evidência obriga a mudar a Decisão da ADR-0015: continua a opção A (o chat coordena, um nível de subagentes, o Go dono do laço). A opção D (líder-agente que cria subagentes) fica sem base, e C (teams) fica como experimental e só Claude. Qualquer mudança de padrão exige nova ADR e execução real fora de `go test`.
