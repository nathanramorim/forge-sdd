# Plano de Discovery 57 — Harness Studio

Roadmap preliminar. Ordem do menor risco ao maior; o MVP termina na fase C.

## Subfeatures sugeridas para o `/split-features`

Pasta sugerida: `sdd/features/feat-57-harness-studio-electron/`

### A — Fundação
- [ ] **A1:** Esqueleto Electron + React + TS (Vite), preload com `contextBridge`, IPC tipado e validado, flags de segurança + CSP, teste de segurança (critério 6).
- [ ] **A2:** Estado local em `userData` e estrutura de serviços no main.

### B — Painel de sessões (somente leitura)
- [ ] **B1:** `HistoryService`: parser JSONL tolerante, índice, custo e duração (critério 1).
- [ ] **B2:** Watcher com `chokidar` e atualização em tempo real (critério 2).
- [ ] **B3:** UI: lista, busca, detalhe da sessão.

### C — Conduzir sessões (fecha o MVP)
- [ ] **C1:** `SessionDriver` com Agent SDK: start/resume/stop, stream, permissões por callback (critério 3).
- [ ] **C2:** `HookReceiver` local com token efêmero e status ao vivo (critério 5).
- [ ] **C3:** `ProfileService`: formato de perfil, preview/diff, apply com confirmação, revisão de hooks/MCP, backup/desfazer (critério 4).
- [ ] **C4:** UI: seletor de perfil ao criar sessão + tela de revisão/diff.
- [ ] **C5:** E2E do fluxo completo (critério 7).

### D — Pós-MVP (backlog, não estimado)
- [ ] **D1:** Harness injetado via opções do SDK (`systemPrompt`, `agents`, `mcpServers`, `hooks`, `settingSources`) sem tocar o disco.
- [ ] **D2:** Perfil como plugin do Claude Code; instalar de Git; marketplace próprio.
- [ ] **D3:** Editor visual de CLAUDE.md/agentes/hooks com validação; diff entre harnesses; import/export zip.
- [ ] **D4:** Terminal embutido (`node-pty` + `xterm.js`).

## Dependências e decisões em aberto

- **Criar o repo privado** e decidir nome, licença e CI (fora deste repo). Os arquivos acima devem ser migrados para lá.
- Confirmar a API vigente do Agent SDK (opções e callbacks) na primeira tarefa de C1, antes de implementar.
- Decidir formato do estado local (JSON simples vs SQLite) em A2.
- Perfis iniciais a empacotar: **Forge SDD** e **Arsenal IA** (apenas conteúdo existente, sem reescrever metodologia).

## Observação de priorização

- **A+B entregam valor sozinhos** (dashboard read-only) com risco baixo.
- **C é o salto de risco** (SDK, hooks, aplicação de perfis no disco do usuário); por isso o diff/confirmação é parte do MVP, não pós-MVP.
- **D** só depois de validar o uso real do MVP.
