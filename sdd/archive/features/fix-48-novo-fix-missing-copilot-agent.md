# fix/novo-fix-missing-copilot-agent

**Branch:** `fix/novo-fix-missing-copilot-agent`
**Fase:** 48
**Depende de:** `fix-5ae2-08-naming-telemetry-tokens`
**Status:** `done`

## Objetivo

Corrigir a ausência do comando `/novo-fix` para o agente Copilot. O comando foi
criado no commit `8029104` (fix-5ae2-08) apenas para Claude e Gemini — o
template `.github/prompts/` (usado por Copilot, o agente default do CLI)
nunca recebeu o arquivo correspondente. Projetos scaffoldados/atualizados com
Copilot nunca tiveram acesso ao `/novo-fix`, mesmo em versões que já o
anunciavam no release notes (1.9.1-beta e 1.9.1).

## Causa raiz

1. `internal/scaffold/templates/.github/prompts/` nunca ganhou um
   `novo-fix.prompt.md.tmpl` — só os diretórios de Claude e Gemini o têm.
2. `commandOrder` em `internal/scaffold/cheatsheet.go` não incluía
   `"novo-fix"`, então mesmo para os agentes que o possuíam, ele não aparecia
   no cheat-sheet impresso após `init`/`update`.
3. As seções "Comandos Customizados"/"Prompts (Comandos)" em
   `CLAUDE.md.tmpl` e `GEMINI.md.tmpl` também não citavam `/novo-fix`,
   deixando o agente sem forma de descobrir o comando lendo sua própria
   instrução de projeto.

OpenAI foi deliberadamente deixado de fora desta fix: o diretório
`.openai/prompts/` não tem nenhum comando implementado hoje (é um gap maior
e pré-existente, não específico do `novo-fix`).

## Critérios de Aceitação Executáveis

1. `internal/scaffold/templates/.github/prompts/novo-fix.prompt.md.tmpl`
   existe, com o mesmo comportamento (branch-first, leitura de
   `naming_convention`, criação de `fix-ID-*.md`) descrito nas versões
   Claude/Gemini.
2. `"novo-fix"` está presente em `commandOrder`
   (`internal/scaffold/cheatsheet.go`), logo após `"nova-feature"`.
3. `CLAUDE.md.tmpl` e `GEMINI.md.tmpl` citam `/novo-fix` na lista de
   comandos.
4. `TestCommandCheatSheet_CopilotOmitsMissingC4Architecture` continua
   verde (Copilot passa a ter `novo-fix`, mas segue sem `c4-architecture`,
   gap não relacionado a esta fix).
5. Golden fixtures (`internal/scaffold/testdata/golden/`) regeneradas
   incluindo `.github/prompts/novo-fix.prompt.md`.
6. Este repositório (dogfood) reflete a correção: `.github/prompts/novo-fix.prompt.md`
   criado, `CLAUDE.md`/`GEMINI.md` atualizados.

## Handoff

Fix pontual em templates + `cheatsheet.go`, sem impacto em `scaffold.Run`/
`RunAgents` (a lógica de cópia de arquivos já tratava `.github/prompts/`
corretamente — faltava apenas o arquivo). Sem mudança de comportamento para
Claude/Gemini, que já tinham o comando.
