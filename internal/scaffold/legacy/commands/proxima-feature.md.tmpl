Leia `sdd/memory/progress.md`, identifique a próxima feature com status `todo` (ou a próxima task na fila `sdd/.queue/queue.json`), reporte as tasks e aguarde confirmação.

Quando confirmado:
1. **Branch (`Regra 15` / `FLX-02`):**
   - Se `sdd/features/feat-XX.md` for um arquivo único, use o nome de branch definido no campo **Branch** desse arquivo.
   - Se a feature corrente for uma subpasta (`sdd/features/feat-XX-nome/*.md`, agrupando várias subtarefas), trate a pasta inteira como unidade de execução: **uma única branch cobre todas as subtarefas da pasta** — nunca crie uma branch por subtarefa.
   - **Branch de partida:** pergunte ao usuário qual branch usar como ponto de partida (default `main`), a menos que já tenha sido indicada explicitamente nesta sessão.
   - **Retomada:** rode `git branch --list <prefixo>/*` para essa feature; se já existir uma branch dela (de sessão anterior, possivelmente com subtarefas parcialmente concluídas), pergunte se deve continuar a partir dela em vez de recriar.
   - Execute `git checkout -b <branch>` (ou `git checkout <branch>` se estiver retomando).
2. **Execução pela Fila do Builder + Governança:**
   - Para cada task da fila (`forge-sdd queue next`):
     - O **Builder** (`.agents/roles/builder.md`) executa em contexto isolado carregando apenas a spec da task, `sdd/memory/lessons.md`, as rules (`.agents/rules/`) e skills (`.agents/skills/`) aplicáveis, e marca `forge-sdd queue complete <task-id>`.
     - Rode a **Camada A (determinística):** `forge-sdd gate`.
     - O **Analista de Governança** (`.agents/roles/governance.md` — Camada B read-only) audita Arquitetura, Boas Práticas e Segurança (`S-01..S-07`), verifica se há atividade padronizada repetida candidata a Auto-Skill (`DUP-05`), e registra o parecer JSON via `forge-sdd queue review <task-id>`.
3. **Retorno ao Líder e Handoff:**
   - O **Líder** apresenta o Painel Executivo ao humano (*1. O que já foi feito e aprovado pela Governança*, *2. Como está cada etapa*, *3. O que ainda não foi desenvolvido*, *4. Exceções e Auto-Skills propostas*).
   - Ao concluir todas as tasks da pasta, marque como concluídas em `progress.md` e `index.md`, e escreva uma release note concisa em linguagem de produto no topo de `sdd/releases/history.md` (sob a seção `## Entregas`).
4. **PR Automático:** Faça o commit das alterações e execute `git push origin <branch>`. Leia o campo "VCS / Work Item System" em `sdd/memory/constitution.md` antes de abrir o PR:
   - `github` (ou campo ausente/default): crie o PR com `gh pr create --fill` (respeitando a Regra 11: versões `-beta` mantêm o PR aberto sem mergear em `main`).
   - `azure-devops`: use `az repos pr create` no lugar de `gh pr create`.
   - `nenhum`: apenas informe o usuário de que a branch está pronta e com push feito.
5. **Gravação de Métricas (determinística):** Se a telemetria estiver habilitada em `sdd/.sddrc` (`telemetry.enabled` como `true`), execute `forge-sdd session record`:
   ```
   forge-sdd session record --feature "<caminho relativo completo da feature/fix ou task>" --phase "<ID>" \
     --agent-path "orquestrador,builder,governance" --outcome approved|rejected|blocked --criterio-atendido=true|false \
     --tokens-input <estimativa> --tokens-output <estimativa> --turns <n> --duration-seconds <n> \
     --model "<modelo>" --files-touched "<arquivo1,arquivo2,...>"
   ```
   - Se a sessão for inativa, cancelada, sofrer timeout ou encerrar sem atingir a finalização da feature/fix, use `--criterio-atendido=false` e `--outcome blocked` ou `--outcome rejected`.
   - Estime `--tokens-input`/`--tokens-output` de forma realista (1 token ≈ 4 caracteres ou 0.75 palavras), nunca zerados se houve interação.
6. **Fallbacks:** Se o `gh` não estiver disponível, informe o link da branch ao usuário.
