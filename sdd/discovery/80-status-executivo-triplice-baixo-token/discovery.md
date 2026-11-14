# Discovery 80 — Status Executivo Tríplice Padronizado e de Baixo Consumo de Tokens

## Problema
1. **Divergência de Tratativa entre CLIs de IA:** Quando o usuário pergunta "qual o status das demandas?" em diferentes ferramentas (Claude Code, Gemini CLI, Copilot CLI ou Codex), cada agente executa comandos diferentes, alguns lendo arquivos JSON imensos, outros listando árvores inteiras de diretórios ou gerando respostas longas e prolixas. Isso consome muitos tokens de contexto e gera fricção na experiência.
2. **Falta de Visibilidade do Trabalho em Tempo Real:** O comando `/status` clássico lê apenas o arquivo de estado em disco, mas não sabe o que o Builder ou Reviewer estão fazendo *neste exato segundo* no terminal (ex.: rodando testes, compilando, criando arquivos).
3. **Alto Custo de Tokens:** Processar logs e estados brutos gasta de 3x a 5x mais tokens do que consultar diretamente as 3 fontes essenciais e sintetizá-las em um template padronizado de tamanho fixo.

## A Solução: O Padrão Tríplice de Status
Quando o status é solicitado, o agente consulta exatamente três pontos objetivos de baixo custo:
1. **Ao Vivo no Painel do Worker (`tmux capture-pane -pt <sessao>:0.1 -S -15`):** Traz as últimas 15 linhas do painel do Builder/Reviewer, mostrando a ação exata em execução.
2. **Checklist da Fila (`forge-sdd queue panel`):** Checklist de tasks concluídas (`[x]`) vs pendentes (`[ ]`).
3. **Commits Recentes (`git log --oneline -n 5`):** O que já foi commitado e integrado na branch da iniciativa.

## Para quem
Desenvolvedores e líderes de projeto que consultam o status do pipeline frequentemente e precisam de respostas instantâneas, previsíveis, idênticas em qualquer CLI e com o menor consumo de tokens possível.

## Jornada Alvo
1. O desenvolvedor digita `/status` ou pergunta "qual o status das demandas?" em qualquer agente (Claude, Gemini, Copilot, Codex).
2. O agente executa os 3 comandos alvos de leitura mínima:
   - `tmux capture-pane` (se em sessão tmux)
   - `forge-sdd queue panel`
   - `git log --oneline -n 5`
3. O agente renderiza a resposta seguindo estritamente o **Template Padrão de Baixo Token** (máximo de 250 tokens de saída).
4. A resposta é idêntica, clara, sem rodeios e de custo computacional mínimo.

## Regras de Negócio
- **RN-01 (Template Canônico Rígido):** Todos os prompts e templates de status (`.gemini/`, `.claude/`, `.github/`, `internal/scaffold/templates`) devem adotar a mesma estrutura visual de 3 seções: `1. Ao Vivo no Painel`, `2. Commits Recentes no Git`, `3. Fila de Tarefas`.
- **RN-02 (Orçamento Estrito de Tokens):** A resposta de status não deve ultrapassar 250 tokens de saída. Leituras de disco não devem carregar JSONs brutos completos na memória do prompt quando a CLI dispõe de utilitários sumarizados.
- **RN-03 (Tolerância a Ambientes sem tmux):** Se o ambiente estiver sem tmux (modo headless ou terminal nativo), a seção `1. Ao Vivo` recai suavemente para a leitura das últimas linhas do log do job em `.forge/runtime/logs/` ou status do processo.
- **RN-04 (Suporte Nativo na CLI):** A CLI Go deve fornecer a flag `forge-sdd status --live` para consolidar esses 3 pontos em um único comando executável de 1 chamada.

## Escopo
- **Dentro:**
  - Padronização de prompts em `.gemini/prompts/status.prompt.md`, `.claude/commands/status.md`, `.github/prompts/status.prompt.md`.
  - Atualização dos templates de scaffold em `internal/scaffold/templates/`.
  - Atualização de `.agents/skills/sdd-operacao/references/status.md`.
  - Implementação da flag `forge-sdd status --live` em `cmd/forge-sdd/status.go`.
  - Documentação no `LEIA-ME`.
- **Fora:**
  - Alterações na máquina de estados da fila.
  - Modificações nas rotinas de build ou revisão.

## Riscos e Mitigações
| Risco | Mitigação |
|---|---|
| Sessão tmux não existir ou ter nome diferente | Resolução automática via `TmuxSessionName(dir)` com fallback silencioso para os logs em `.forge/runtime/logs/`. |
| Prompt do agente ignorar o template canônico | Regra no prompt do agente com formato "Few-Shot" explícito e restrição de tokens. |
