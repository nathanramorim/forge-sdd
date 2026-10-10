# Discovery 57 — Harness Studio: app desktop para conduzir sessões do Claude Code com perfis de harness

Nova vertente **fora do Forge SDD público**: este discovery é a semente de um **repositório privado novo** (Electron + React + TypeScript). Fica registrado aqui apenas para alimentar `/split-features` e depois ser levado ao repo novo. Linguagem: padrão (Constituição, regra 0).

> **Clarify (`sdd/memory/clarify.md`) — respostas do usuário:** (a) artefatos ficam em `sdd/discovery/` deste repo; (b) MVP = **painel de sessões + iniciar/retomar sessão** com perfis de harness aplicados como arquivos; (c) público = **uso pessoal/time, 100% local**.

## 1. Contexto e Problema (Porquê)

- Quem usa Claude Code em várias frentes hoje abre terminais soltos: não há visão única de sessões, custo, duração, nem status ao vivo ("terminou?", "pediu permissão?").
- O **harness** (CLAUDE.md, agentes, skills, hooks, MCP, permissões) é o que diferencia metodologias como Arsenal IA e Forge SDD, mas é aplicado copiando arquivos à mão, sem versão, sem diff e sem revisão do que executa código na máquina.
- Em ambiente com governança forte, os históricos (JSONL) contêm código e prompts: a ferramenta precisa ser **local por construção**.

## 2. Para Quem

- **O próprio mantenedor** (usuário principal): alternar metodologias por projeto e acompanhar sessões em um painel.
- **Time próximo**: receber perfis de harness versionados e aplicá-los com segurança.
- Fora de escopo agora: distribuição comercial, licenciamento, auto-update público.

## 3. Proposta de valor (macro)

1. **Painel de sessões**: lista, busca e visualização de sessões antigas com custo e duração, lendo `~/.claude/projects/<projeto>/*.jsonl` (somente leitura, tempo real).
2. **Conduzir sessões**: criar, retomar e encerrar sessões via **Claude Agent SDK** (main process), com stream de mensagens/tool calls e permissões por callback.
3. **Perfis de harness**: ao criar a sessão, escolher um perfil ("Forge SDD", "Auditoria", ...) que é aplicado ao projeto como arquivos (`CLAUDE.md`, `.claude/agents|skills|settings.json`, `.mcp.json`).
4. **Status ao vivo**: hooks `SessionStart`, `Stop`, `Notification` enviam eventos ao app por HTTP local/socket.

Pós-MVP (backlog): injeção de harness via opções do SDK, editor visual com validação, diff entre harnesses, import/export como plugin/zip, marketplace próprio, terminal embutido (`node-pty` + `xterm.js`).

## 4. Como (macro)

Combinação recomendada: **JSONL para histórico + Agent SDK para conduzir + hooks para tempo real**. O CLI como processo filho (`claude -p --output-format stream-json --resume`) fica como fallback, por ser mais frágil. O formato de plugin do Claude Code é o candidato a formato de perfil (pós-MVP); no MVP, o perfil é uma pasta versionada em Git copiada para o projeto.

## 5. Riscos e Trade-offs

- **Formato JSONL é interno** e pode mudar entre versões do Claude Code → parser tolerante, testes com fixtures por versão, somente leitura.
- **SDK no main process**: superfície de ataque do Electron → `contextBridge` + IPC tipado, `nodeIntegration` desligado, `contextIsolation` ligado, CSP estrita.
- **Perfis executam código** (hooks, MCP, comandos) → pin de versão (commit/hash), revisão obrigatória antes de aplicar, nunca aplicar silenciosamente.
- **Sobrescrever arquivos do projeto** ao aplicar perfil → detectar conflitos e pedir confirmação (mesmo princípio do `init` do Forge SDD).
- **Dependência do SDK/CLI** (mudanças de API) → isolar em um adaptador único (`SessionDriver`).
- **Privacidade**: nenhuma telemetria externa; tudo local.

## 6. Critérios de aceitação macro

- [ ] O painel lista sessões reais de `~/.claude/projects` com custo e duração e atualiza em tempo real.
- [ ] É possível iniciar uma sessão nova e retomar uma existente pelo app, vendo o stream e respondendo a pedidos de permissão.
- [ ] Ao criar a sessão, o usuário escolhe um perfil; o app mostra o que será gravado (diff) e só aplica após confirmação.
- [ ] Eventos de hook mudam o status da sessão no painel em até ~1 s.
- [ ] Nenhum dado sai da máquina; renderer sem acesso direto a Node.

## 7. Fora de escopo (MVP)

Editor visual, diff entre perfis, marketplace, import/export zip/plugin, terminal embutido, auto-update, build assinado/notarizado.

## 8. Próximos passos propostos

1. Esqueleto Electron + React + TS com `SessionDriver` (Agent SDK) e IPC seguro.
2. Desenho do **formato de perfil de harness** e do fluxo de importação/aplicação.
