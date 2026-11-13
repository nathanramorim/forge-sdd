# feat/smart-upgrade-detection

**Branch:** `feat/smart-upgrade-detection`
**Fase:** 22
**Depende de:** `feat-07-release`, `feat-08-npx`
**Status:** `done`

## Objetivo

Aprimorar o comando `/upgrade-sdd` para que o agente identifique automaticamente o ambiente de instalação do CLI (Homebrew vs NPX) e forneça a instrução de atualização correta para o binário, além de realizar a migração estrutural dos arquivos.

## Critério de conclusão

- [x] O comando `/upgrade-sdd` detecta o ambiente (ex: via `which forge-sdd` ou checagem de caches).
- [x] Se instalado via Homebrew, sugere `brew upgrade forge-sdd`.
- [x] Se executado via npx, informa que a versão `@latest` já garante a atualização ou sugere limpar o cache `rm -rf ~/.cache/forge-sdd`.
- [x] Os templates de prompt de todos os agentes (Gemini, Copilot, Claude) foram atualizados com esta lógica.

## Tarefas

- [x] **22-1** Atualizar a skill `Migrator` para incluir a etapa de "Environment Detection".
- [x] **22-2** Atualizar o prompt `/upgrade-sdd` no Gemini, Copilot e Claude para incluir as instruções condicionais de binário.
- [x] **22-3** Implementar a lógica de detecção.
- [x] **22-4** Atualizar a documentação `README.md` e `HOWTO.md` com o novo fluxo de upgrade inteligente.

## Arquivos gerados/modificados

```
internal/scaffold/templates/agents/*/prompts/upgrade-sdd.prompt.md.tmpl
internal/scaffold/templates/agents/*/skills/migrator.chatmode.md.tmpl
docs/metodologia-sdd.md
README.md
```
