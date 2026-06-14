# feat/install-skill-from-url

**Branch:** `feat/install-skill-from-url`
**Fase:** 21
**Depende de:** `feat-11-unify-agent-structures`
**Status:** `todo`

## Objetivo

Implementar o comando `/install-skill` que permite ao agente importar e adaptar skills diretamente de URLs do GitHub (como o repositório oficial de skills da Anthropic). O comando deve ser capaz de ler o conteúdo remoto, salvá-lo em `sdd/skills/` e garantir que a instrução seja válida para Copilot, Gemini e Claude.

## Critério de conclusão

- [ ] Novo comando `/install-skill <url>` implementado em todos os agentes.
- [ ] O agente lê a URL, extrai a instrução e cria o arquivo `.md` em `sdd/skills/`.
- [ ] A skill instalada é automaticamente adicionada ao `sdd/skills/index.md`.
- [ ] A instrução da skill é adaptada (se necessário) para ser compatível com o formato mult-agente do Forge-SDD.

## Tarefas

- [ ] **21-1** Criar prompt `/install-skill` para Gemini, Copilot e Claude.
- [ ] **21-2** Instruir o `Specifier` a utilizar ferramentas de leitura web (como `web_fetch` ou similar, dependendo do agente) para acessar o GitHub.
- [ ] **21-3** Definir o protocolo de adaptação: converter instruções específicas de um agente (ex: Claude) para um formato agnóstico compatível com o SDD.
- [ ] **21-4** Atualizar `sdd/skills/index.md` após cada instalação bem-sucedida.
- [ ] **21-5** Adicionar documentação no README e HOWTO sobre como estender as habilidades do agente via URLs externas.

## Arquivos gerados/modificados

```
internal/scaffold/templates/agents/*/prompts/install-skill.prompt.md.tmpl
sdd/skills/index.md
README.md
docs/metodologia-sdd.md
```
