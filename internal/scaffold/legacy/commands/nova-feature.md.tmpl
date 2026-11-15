# Comando: nova-feature

**Uso:** Peça "/nova-feature <descrição>" ou "/nova-feature fix: <descrição>" (para correções/bugfixes)

**Ação:**
0. **Clarify:** Antes do PASSO 1, avalie a descrição recebida contra a lógica única descrita em `sdd/memory/clarify.md`. Se algum sinal de lacuna for detectado, faça a rodada de perguntas antes de prosseguir; caso contrário, siga direto.
1. **PASSO 1 MANDATÓRIO — Branch:** Defina o prefixo de branch (`feat/` ou `fix/` se for correção/bugfix) e de arquivo (`feat-` ou `fix-`).
   - **Branch de partida:** Pergunte ao usuário qual branch usar como ponto de partida (default `main`). Só prossiga sem perguntar se o usuário já indicou explicitamente nesta mesma solicitação.
   - **Retomada:** Rode `git branch --list <prefixo-branch>/*` para essa feature/fix. Se já existir uma branch correspondente de uma sessão anterior, pergunte se deve continuar a partir dela em vez de criar uma nova.
   - Crie a branch localmente no terminal (`git checkout -b <prefixo-branch>/<nome>`, a partir da branch de partida escolhida) **antes** de criar qualquer especificação ou logs.
2. **Nomenclatura:** Siga a lógica única descrita em `sdd/memory/naming-convention.md`.
3. **Criação (PM/Arquiteto + Planejador):**
   - Se a feature pertencer a uma iniciativa existente de Discovery: crie diretamente dentro de `sdd/discovery/<id>-<nome>/features/feat-<id>-<nome>.md`.
   - Se for uma feature independente direta: crie uma pasta autocontida `sdd/features/feat-ID-<nome>/feat-ID-<nome>.md` (com subpastas para subtasks e fixes se necessário; suporta também arquivos legados `sdd/features/feat-ID-<nome>.md`).
   O arquivo deve conter obrigatoriamente as seções:
   - `## Modelo do módulo` (`completo | enxuto` + justificativa — `ARQ-11`);
   - `## O que já existe` (tabela de Reuse Check — `DUP-02`);
   - `## Dependências novas` (`SEG-07`);
   - `## Rules e IDs da Constituição Aplicáveis` (consultando `sdd/memory/constitution.md` e `.agents/rules/*.md`);
   - `## Critérios de Aceitação Executáveis`.
   A pasta inteira é a unidade de execução: uma única branch (a criada no passo 1) agrupa todas as subtarefas — nunca crie uma branch por subtarefa. Atualize o progresso e índice.
4. **Gate de Revisão de Spec (`FLX-01`) e Fila:** Submeta a Spec criada à revisão humana no Líder e, quando aprovada, enfileire para o Builder (`forge-sdd queue enqueue`).
5. **Gravação de Métricas (determinística):** Se a telemetria estiver habilitada em `sdd/.sddrc`, execute `forge-sdd session record --feature "<arquivo feat/fix criado>" --outcome blocked --criterio-atendido=false` (a feature/fix ainda não foi implementada nesse ponto do fluxo).

**Handoff:** Gere um resumo para o comando `proxima-feature`.
