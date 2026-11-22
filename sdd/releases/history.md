# Release Notes — forge-sdd

Este arquivo registra o histórico de entregas de produto deste projeto.

## Entregas

### Versão 5.2.0 — O CI Sai do Caminho, a Revisão Vem Antes e Cada Feature Tem o Seu PR (Beta)

Compatível com a 5.0.0. Consolida a iniciativa 106 inteira: o CI deixa de rodar em trabalho não conferido, a revisão acontece em dois pontos (a spec e a feature) e cada feature do split tem branch e PR próprios.

*   **Agora a LP só vai ao ar quando você manda.** Antes, cada push publicava a página; agora o deploy é um comando que regera, confere e publica, e avisa quando a LP está desatualizada (`scripts/lp-deploy.sh`, com `--dry-run` e `--status`; o workflow `Deploy LP` ficou só manual).
*   **Agora você escolhe o tmux de volta.** O padrão continua sendo as janelas do sistema, mas com `--terminal tmux` ou `terminal: tmux` no `forge.json`, o `forge up` abre as três sessões em tmux. Sem o tmux instalado, ele avisa e usa as janelas, sem erro (o modo `live` só vale com o tmux; `--dry-run` só imprime).
*   **Acabou gastar CI em trabalho que ninguém conferiu.** A validação roda na sua máquina e deixa um selo do commit; o `gh pr ready` exige esse selo junto do parecer. O GitHub Actions roda uma vez, quando o PR sai do rascunho, e não roda mais em cada push (`scripts/ci-local.sh`, `concurrency`, job `pronto?` recusando PR em draft; `scripts/ci-politica.py` mostra os runs por dia).
*   **Agora a spec volta cedo, em PR próprio.** O `/discovery`, o `/nova-feature` e o `/novo-fix` abrem só a spec, em PR draft. Passada a Conferência 1 e o seu `[Seguir]`, a spec entra na principal e a construção começa, sem ficar presa numa branch até o fim (o Gate 1 virou uma linha explícita no `status.md`).
*   **Agora cada feature tem a sua branch e o seu PR em draft.** O Builder abre uma branch por feature do split, sempre a partir da principal, e nenhum PR depende de outro. Uma feature só começa quando as dependências já foram mescladas, e o `[Continuar]` vem agrupado por onda.
*   **Agora você instala uma skill ou plugin só no projeto.** Você escolhe o que instalar e onde, sem rede e sem mexer no `~/.claude`; se o hub já tem, nada se repete (`forge-sdd skills destinos` e `forge-sdd skills instalar <skill|plugin> <slug> --de <dir>`, ADR-0025).
*   **O `/toca-ai` deixa de supor que há workers rodando.** Sem worker ativo, você é avisado de que as etapas ficam na fila, e o chat orquestra subagentes em paralelo em vez de mandar você rodar o `forge up` (o `forge doctor` também avisa fila parada).
*   **Acabou marcar "pronto" sem parecer válido.** O `gh pr ready` e o `gh pr merge` só passam com parecer `conforme` do commit atual; se o código mudou depois da revisão, os dois SHA aparecem lado a lado na recusa.
*   **Agora a barreira de PR fica instalada sozinha.** O hook que barra o `gh` entra no `.claude/settings.json` sem você pedir, e o `forge doctor` mostra em quais agentes ele vale; nos outros, o job `pronto?` do CI segura (hook `PreToolUse`).
*   **Agora a issue nasce junto com a spec.** Com `create_on_conception: true` no `forge.json` (desligado por padrão), a issue é criada ou reaproveitada e nunca duplicada, e o dispatch lê a evidência. Falha externa vira aviso e não trava a spec (`card-evidence.json`; com gestor por MCP, `forge queue card --feature <id> --provider <p> --id <id> --url <https>`; para renovar, apague o arquivo).
*   **Agora a skill de diagramas vive no repositório.** Ela pergunta o motor, Mermaid (padrão) ou diagram-studio, e não instala nada sozinha (`diagramas-c4-mermaid` em `.agents/skills/`).
*   **Agora o atestado é conferido em segundos no CI.** No modo `local`, o parecer do PR de feature é verificado na hora; o modo `desligado` diz em voz alta que nada foi conferido, e `vcs` é recusado como adiado (`scripts/atestado.py`).
*   **Agora você vê quanto o CI custa, com número que se defende.** O cálculo cruza a cobrança do GitHub com o custo por job e mostra a diferença; jobs pulados e cancelados ficam à parte e o CSV abre seguro na planilha (`scripts/ci-politica.py`; células iniciadas por `=`, `+`, `@` ou `-` são neutralizadas).
*   **Agora a Constituição guarda os modos do projeto.** A tabela `Ferramentas e Integrações` registra o modo do atestado, da revisão humana, o WIP humano, os módulos permitidos e o início automático do Builder, e as lições passam a ficar num arquivo por regra (`GOV-03`, em `sdd/memory/licoes/`).
*   **O modo sem humano só liga depois de você ver os riscos.** O aviso dos três riscos precisa ser apresentado e aceito; recusar mantém o padrão seguro, `com-humano`.
*   **Agora o `/upgrade-sdd` leva a política de PR para o seu projeto.** Você recebe o hook, o job `pronto?`, os scripts e os campos de modo sem perder o seu `ci.yml` nem o `.claude/settings.json`; o forge mescla, nunca sobrescreve, e JSON inválido não é tocado.
*   **Agora o README tem um prompt para cada ferramenta.** São quatro blocos "cole isto na sua IA", um por agente, cada um apontando para o arquivo de instruções que ele lê.
*   **Como atualizar:** `npm install -g @nathanramorim/forge-sdd@beta` para a prévia; nos projetos existentes, `forge update --upgrade` e depois `/upgrade-sdd`, que leva a política de PR. Rode `forge doctor` no fim: a seção "Barreira de PR" diz o que ficou instalado.
*   **Avisos da beta:** esta versão não está na `main` (`FLX-07`); quem usa `npm install -g @nathanramorim/forge-sdd` sem `@beta` continua na linha estável.

### Versão 5.1.0 — CLI Mais Claro, Métricas de Discovery e Parecer com Proveniência e Fila por Branch de Feature (Beta)

Compatível com a 5.0.0 (Iniciativas 97, 104, 105 e 106). Beta: a versão não está na `main`; a estável continua sendo a 5.0.0.

*   **`forge doctor` compara versões e fala menos (#125):** o diagnóstico compara a versão do projeto com a do programa e avisa quando o projeto está atrás. Os caminhos de features e tasks só aparecem com `--verbose`.
*   **`forge status` esconde o que já terminou (#124):** iniciativa com a fila concluída sai da lista; `forge status --all` mostra todas, marcando as concluídas.
*   **Aviso de comando depreciado (#126, #129):** `queue stations`, `queue handoff` e `attach` mostram o substituto ao serem usados, e o aviso aparece também na documentação dos atalhos `/implement` e `/proxima-feature`. Os usos dos comandos do programa (`queue stations`, `queue handoff` e `attach`) são registrados em `sdd/memory/logs/depreciados.csv`, para você ver o que ainda está em uso antes da remoção; os atalhos de barra não geram linha no CSV. A decisão está na ADR-0021 (#123).
*   **"3 sessões" no lugar de "5 estações" (#127, #128):** a ajuda do programa e os textos entregues passam a falar de Discovery, Builder e Reviewer; a descrição do agente `lider` segue o mesmo caminho. Nenhum comando foi renomeado.
*   **Métricas de Discovery (#131):** as métricas das sessões de concepção ganham o bloco `spec`.
*   **Parecer do Reviewer com proveniência (feature 106-11, #159):** o parecer guarda `head_sha`, `ponto` (`spec` ou `entrega`), `modo` e `feature`, gravados pelo Go a partir do HEAD da branch. Veredito sem commit válido é recusado.
*   **A fila acha a iniciativa em branch por feature (fix-106-01, #163):** em branches como `feat/106-11-...` ou `fix/106-01-...`, os comandos que leem a fila (`forge queue list`, `forge queue next` e os demais subcomandos de `queue`) passam a achar a iniciativa casando o número do começo da branch com a pasta em `sdd/discovery/`. Se duas pastas tiverem o mesmo número, o comando falha com erro claro e lista as pastas. `forge status`, `forge doctor`, `forge gate`, `forge archive` e `forge context` conferem isso antes de ler a fila. Nova flag global `--iniciativa <id>` e variável `FORGE_INICIATIVA` escolhem a iniciativa à mão (a flag vence a variável, e a variável vence a branch); id que não existe dá erro.
*   **README mais claro (#130, #132):** a esteira é contada uma vez só, e os diagramas usam as cores e as marcas dos robôs da página.
*   **O que vem a seguir (em desenho, sem data):** contrato de módulos, régua de estimativa de tokens, painel de métricas, orquestrador e política de PR em rascunho com CI. Nada disso está nesta beta.
*   **Como atualizar:** `npm install -g @nathanramorim/forge-sdd@beta` para instalar a prévia; nos projetos existentes, rode `forge update --upgrade` e depois `forge doctor` para conferir se o projeto está em sincronia com o programa. Para voltar à estável, `npm install -g @nathanramorim/forge-sdd` sem `@beta`.
*   **Avisos da beta:** esta versão não está na `main` (`FLX-07`); quem usa `npm install -g @nathanramorim/forge-sdd` sem `@beta` continua na 5.0.0 estável.

### Versão 5.0.0 — Wiki Get Started, PR Automático e Edição Missões (Estável)

Compatível com a 4.8.0. Consolida em versão estável tudo o que foi entregue na linha 4.9.x e nas Iniciativas 97 a 100.

*   **Wiki Get Started (`forge-sdd.org/wiki`):** uma jornada guiada de instalação, primeira feature e ciclo completo. Cada passo traz um bloco "Prompt para IA", pronto para copiar e colar no seu agente. Inclui o modo Hub, para usar o forge-sdd em vários repositórios, e uma seção sobre `/constitution`. O botão "Get Started" da landing page leva direto para lá (Iniciativa 99).
*   **PR automático depois da revisão (`forge queue review`):** quando o veredito é `conforme`, o comando envia a branch e abre o PR com `gh pr create`, sem perguntar. Use `--no-pr` para pular (Iniciativa 100).
*   **Revise a spec sem sair da conversa:** ao encerrar `/discovery`, `/nova-feature`, `/novo-fix` ou `/split-features`, o agente pergunta se você quer revisar os artefatos gerados antes de seguir para o Gate 1, sem navegar pelas pastas (`feat-97-03`).
*   **Worktrees ociosos não se acumulam:** ao iniciar `forge up`, `forge worker` ou `forge inbox`, o forge detecta worktrees em `.forge/runtime/worktrees/` que podem ser removidos com segurança e oferece a limpeza (`feat-97-02`).
*   **Métricas de sessão com tokens de verdade:** corrigidos os `tokens_input: 0` e `tokens_output: 0` em `sdd/.metrics/session-*.json`. O campo `complexidade` fica reservado no schema, ainda sem preenchimento (`feat-97-01`).
*   **Diagramas C4 com dois motores:** a skill de diagramas pergunta, a cada uso, se você quer Mermaid ou diagram-studio (Iniciativa 98).
*   **`forge rules` volta a achar os IDs da Constituição** (`fix-88-05`).
*   **Vindo da linha 4.9.x:** o modo pipeline avançado (experimental), o menu Settings e os tetos desligados por padrão passam a fazer parte da versão estável. O modo normal não muda.
*   **Edição Missões:** o README ganha um banner temático e as sessões do pipeline aparecem como missões: Discovery é o plano, Builder é a obra, Reviewer é o vigia, e a sua aprovação fecha com "Missão cumprida". O design system da edição fica em `docs/design/edicao-missoes/`. É uma homenagem ao clima de pôr do sol, neon e palmeiras, com marca, tipografia e ilustrações próprias.
*   **Como atualizar:** `npm install -g @nathanramorim/forge-sdd` e, nos projetos existentes, `forge update`.

### Versão 4.9.1 — Correções do Modo Avançado: Fila, Teto Diário e Tetos no Settings (Beta)

Compatível com a 4.9.0-beta (Iniciativas 84 e 88).

*   **A fila passa a refletir a aprovação do Reviewer (`fix-84-03`):** o job guardava as tasks como `88-01` e a fila como `feat-88-01-<nome>`, então nenhuma casava e tudo seguia `queued` mesmo com o PR aprovado. O casamento agora ignora o prefixo `feat-`, e um aviso aparece quando um job traz tasks que não pertencem à fila. O auto-archive volta a funcionar.
*   **O teto diário de jobs não derruba mais a esteira (`fix-88-01`):** a contagem passa a ser de jobs distintos (antes, cada nova reivindicação do mesmo job contava), o dia é o do seu fuso e, ao atingir o teto, a esteira entra em pausa até o dia seguinte e retoma sozinha, em vez de encerrar. O `STOP` continua encerrando de forma limpa, inclusive no modo ao vivo.
*   **Todos os tetos desligados por padrão e no Settings (`fix-88-02`):** o teto diário de jobs, o de tokens por job e o de tempo por job nascem desativados e se ligam no menu **Settings**. Sem nenhum teto ligado, só o `STOP` encerra a esteira avançada. O timeout do runner (`builder.job_timeout_minutes`) não mudou.
*   **Como atualizar:** `npm install -g @nathanramorim/forge-sdd@beta`. Esteiras que já têm `max_jobs_per_day` e `max_job_tokens` gravados no `forge.json` mantêm esses valores; para desligar, use o Settings.

### Versão 4.9.0 — Modo Pipeline Avançado e Menu Settings (Beta)

Compatível com a 4.8.0 (Iniciativa 88). Experimental: o modo normal não muda.

*   **Modo pipeline avançado (escondido no menu):** em "Iniciar esteira (up)", a seta **→** liga e a seta **←** desliga (`autonomy: advanced` em `forge.json`). Exige `permissions: full` e a confirmação de que o ambiente é isolado. Pensado para deixar a esteira rodando por dias sem você responder a cada passo.
*   **Reviewer revisa a spec:** no modo avançado, depois da Discovery o Reviewer confere a spec contra Constitution, rules, ADRs e arquitetura (job `spec_review`, parecer `aprovado`, `ajustar` ou `escalar`). Depois da entrega do Builder, confere o código de novo contra a spec. Descrito na ADR-0013 (proposta).
*   **Task `blocked` não para a fila:** ao esgotar as tentativas ou escalar, a task fica `blocked` com o motivo no inbox e a esteira segue nas tasks independentes.
*   **Salvaguardas:** tetos de tokens e de tempo por job e de jobs por dia, e parada limpa por `.forge/runtime/STOP`. O merge do PR continua manual.
*   **Menu Settings e fim da pergunta de card por padrão:** `pm_mcp.require_existing_card` deixa de nascer `true`; o `forge pipeline init` pergunta se você usa gerenciador de projeto, e o novo menu **Settings** liga e desliga a chave. Esteiras existentes mantêm o valor gravado.
*   **Como atualizar:** `npm install -g @nathanramorim/forge-sdd@beta` e, nos projetos existentes, `forge update`.

### Versão 4.8.0 — Painel Interativo e Fila que Acompanha a Aprovação (Estável)

Compatível com a 4.6.1. Consolida em versão estável tudo o que foi entregue na linha 4.7.x.

*   **Veja o que cada sessão está fazendo, em linguagem de gente:** o Builder e o Reviewer mostram um stream narrativo, estilo chat (analisando, editando, validando, concluído), no lugar de contadores secos e saída crua de terminal.
*   **Fila interativa com drill-down:** `forge queue -i` e `forge status -i` listam todas as tasks, sem o corte "... e mais N tarefas", em duas abas: **Na Fila** e **Concluídas**. Um `Enter` abre o detalhe da task: duração, arquivos alterados, laudo do Reviewer, tokens por etapa e o caminho Discovery → Builder → Reviewer.
*   **Popup no tmux:** `Ctrl+b q` abre a fila por cima dos três painéis e devolve você exatamente onde estava.
*   **Painel que não mente:** quando o Reviewer aprova um job, a fila, o `forge status` e o painel passam a mostrar as tasks como aprovadas (antes ficavam em 0/N). A iniciativa concluída pode ser arquivada sozinha, e o veredito fica guardado em `sdd/discovery/<iniciativa>/verdicts/`, sem o Reviewer precisar escrever no repositório.
*   **Menos surpresas no tmux:** abrir as sessões com tokens acumulados não derruba mais o painel para janelas soltas do sistema.
*   **Documentação com cara de produto:** novo README com banner, selos e guia de instalação e uso em linguagem simples; `LICENSE` MIT; especificação técnica (`sdd/spec/`) alinhada ao código e quatro novas decisões de arquitetura registradas (ADR-0009 a ADR-0012).
*   **Como atualizar:** `npm install -g @nathanramorim/forge-sdd` e, nos projetos existentes, `forge update`.

### Versão 4.7.7 — Fila e Painel Acompanham a Aprovação do Job (Beta)

Compatível com a 4.7.6-beta (Iniciativa 84 — `fix-84-02`).

*   **A aprovação do Reviewer agora atualiza a fila (`internal/pipeline/review.go`):** ao aplicar o veredito `conforme`, as tasks do job (`TaskIDs`, da iniciativa do job) passam a `approved` no `queue.json`, então `forge status`, `forge queue list` e o painel deixam de ficar em 0/N. A mudança é idempotente e nunca faz uma task aprovada regredir.
*   **O PR aberto entrega as tasks à revisão (`forge pipeline pr-opened`):** as tasks do job passam a `review_pending` quando o Builder registra o PR.
*   **Veredito preservado:** o Go guarda uma cópia do veredito em `sdd/discovery/<iniciativa>/verdicts/<job>.json`; o Reviewer segue somente leitura e o arquivo temporário pode ser descartado.
*   **Fila da iniciativa do job, não da branch atual (`internal/queue/taskstatus.go`):** `SetTaskStatuses` abre o `queue.json` da iniciativa pelo nome, já que Builder e Reviewer podem estar em outra branch.

### Versão 4.7.6 — Legibilidade e Alinhamento em Colunas na Lista de Tarefas (Beta)

Compatível com a 4.7.5-beta (Iniciativa 84 — `feat-84-07`).

*   **Alto contraste no cabeçalho de abas (`internal/survey/queue_drilldown.go`):** move o indicador da aba ativa (`ABA ATIVA: ✅ CONCLUÍDAS (25) │ ⏳ Na Fila / Em Andamento (0)`) para o `Title` principal em destaque, em vez do `Description` acinzentado do `huh`.
*   **Linhas limpas e alinhadas em colunas sem repetição visual:** na aba de Concluídas, remove a repetição de `[✓] Concluído` e `Disc→Bld→Rev ✓` em todas as 25 linhas e alinha o número da tarefa, o identificador (`50` colunas) e a coluna de tokens (`│ 🪙 10k tok`) verticalmente.

### Versão 4.7.5 — Troca de Aba no Topo da Lista, Sem Quebra de Linha e Dica Clara de Tela Cheia no Tmux (Beta)

Compatível com a 4.7.4-beta (Iniciativa 84 — `feat-84-07`).

*   **Botão de trocar de aba fixo no topo da lista (`internal/survey/queue_drilldown.go`):** a opção **`⇄ [Trocar Aba] Ir para Aba 1 / Aba 2`** agora é sempre o **1º item no topo da lista** (com o cursor iniciando direto na 1ª tarefa logo abaixo), permitindo alternar entre as abas com apenas uma seta para cima (`↑` + `Enter`) sem precisar rolar até o final quando há 25+ tarefas.
*   **Eliminação de repetição de ID e quebra de linha na lista (`cleanTaskDisplay`):** remove a repetição redundante quando o título em `queue.json` repete o `TaskID` (`Feature: <task_id>`), mantendo todas as linhas da fila em linha única.
*   **Linguagem clara na dica do atalho `Ctrl+b z` (`internal/pipeline/tmux.go`):** substitui o termo ambíguo `"z zoom"` por **`Ctrl+b z tela cheia/voltar`** (`[z] Tela Cheia/Voltar`), deixando claro que o atalho expande o painel atual em tela cheia e volta aos 3 painéis ao pressionar novamente.

### Versão 4.7.4 — Layout Sem Bordas Direitas e Separação em 2 Abas (Na Fila vs. Concluídas) (Beta)

Compatível com a 4.7.3-beta (Iniciativa 84 — `feat-84-07`).

*   **Blocos de detalhe abertos à direita (`internal/survey/task_detail.go`):** substitui a caixa fechada nos 4 lados (`┌─...─┐` / `│...│`) por blocos com borda apenas à esquerda (`╭───` / `│` / `├───` / `╰───`), eliminando em definitivo qualquer desalinhamento ou quebra de layout causada por emojis, acentos ou largura do terminal.
*   **Separação em 2 Abas (`1. Na Fila / Em Andamento` e `2. Concluídas`) (`internal/survey/queue_drilldown.go`):** divide a listagem interativa em duas listas independentes com alternância direta no menu (`⇄ Ver Aba 2: ✅ Concluídas` / `⇄ Ver Aba 1: ⏳ Na Fila / Em Andamento`) e formatação fluida que não quebra quando o `TaskID` é extenso.

### Versão 4.7.3 — UX dos Botões de Ação no Painel e Menu Interativo (Beta)

Compatível com a 4.7.2-beta (Iniciativa 84 — `feat-84-07`).

*   **Ações claras no card de detalhes da tarefa (`internal/survey/task_detail.go` e `cmd/forge-sdd/queue.go`):** substitui os dois botões idênticos `[ Voltar ]  [ Voltar ]` por ações distintas e identificadas: **`← Voltar à lista`** (retorna para a lista de tarefas) e **`✕ Fechar painel`** (fecha o popup/modal imediatamente).
*   **Remoção do botão duplicado na pausa do menu (`internal/survey/select.go`):** substitui o componente de confirmação com dois botões repetidos `[ Voltar ao menu ]  [ Voltar ao menu ]` por uma opção única **`← Voltar ao menu`**.

### Versão 4.7.2 — UX da Fila (Separação Ativas vs. Concluídas, Esteira Discovery→Builder→Reviewer, Tokens por Etapa e Dicas no Tmux) (Beta)

Compatível com a 4.7.1-beta (Iniciativa 84 — `feat-84-07`).

*   **Separação de tarefas `Na Fila / Em Andamento` vs. `Concluídas` (`internal/survey/queue_drilldown.go`):** as tarefas ativas ou aguardando na fila aparecem agrupadas no topo (`[FILA]`) e as concluídas ficam separadas abaixo (`[CONCLUÍDA]`), exibindo em qual etapa da esteira `Discovery ──► Builder ──► Reviewer` cada tarefa está e seu consumo de tokens (`🪙 Xk`).
*   **UX e Tokens por Etapa no Card de Detalhes da Tarefa (`internal/survey/task_detail.go` e `internal/monitor/tokens.go`):**
    *   Seção visual dedicada da esteira: `[✓ 1. Discovery] ──► [✓ 2. Builder] ──► [✓ 3. Reviewer]`.
    *   Seção de consumo de tokens discriminado por etapa (`1. Discovery`, `2. Builder`, `3. Reviewer`) e **Total da Tarefa**, além de estimativa automática gravada pelo worker (`cmd/forge-sdd/worker.go`) ao encerrar cada job.
    *   Alinhamento exato das bordas direitas (`│`) com cálculo de largura visual Unicode/emojis (78 colunas).
*   **Dicas dinâmicas de atalhos na barra inferior direita do Tmux (`internal/pipeline/tmux.go`):** exibe dicas rotativas (`💡 Dica: Ctrl+b q abre Fila`, `💡 Dica: Ctrl+b Setas navega · z zoom`, e guia imediato `⌨️ [q] Fila/Detalhes · [Setas] Painel · [z] Zoom` ao pressionar `Ctrl+b`) ao lado do resumo de tokens, usando `-t <sessão>` sem prefixo `=`.

### Versão 4.7.1 — Fix no Tmux: Correção do Alvo em status-right com Tokens Acumulados (Beta)

Compatível com a 4.7.0-beta.

*   **Correção definitiva da abertura agrupada no tmux quando há tokens acumulados:** substitui o alvo com prefixo `=` (`=forge-<projeto>-<hash>`) pelo nome direto da sessão (`forge-<projeto>-<hash>`) nos passos `tmux set-option -t <sessão> status-right` e `status-right-length` em ambos os modos (`live` e `headless`). No tmux 3.8+, `set-option -t =<sessão>` falha com `exit status 1 (no such session)`, o que derrubava a sessão recém-criada e disparava indevidamente o fallback de 3 janelas separadas do sistema sempre que `totalTokens > 0`.
*   **Teste de integração real com tmux (`FORGE_TMUX_REAL=1`):** inclui verificação ponta a ponta com métricas de tokens populadas em `sdd/.metrics/` contra um servidor tmux real e corrige a inspeção de painéis com `base-index 1`.

### Versão 4.7.0 — Painel Interativo: Stream Narrativo nos Workers, Fila com Drill-Down e Popup no Tmux (Beta)

Compatível com a 4.6.1 (Iniciativa 84). Sem breaking changes.

*   **Stream narrativo estilo chat no Builder e no Reviewer:** o heartbeat monótono (`trabalhando há 4m40s · 3 arquivos alterados`) dá lugar a etapas legíveis (💭 analisando, 🔨 editando, 🧪 validando, ✅ concluído, ❌ falha com as últimas linhas relevantes). Há fallback ASCII para terminais sem Unicode.
*   **Fila interativa com drill-down:** `forge queue -i` e `forge status -i` listam 100% das tasks (sem o corte "... e mais N tarefas") com badge de status. Setas navegam, `Enter` abre o detalhe da task (duração, arquivos alterados, commit, laudo do Reviewer) e `Esc`/`q` volta. Usa Alternate Screen Buffer e não suja o histórico do terminal.
*   **Popup no tmux sem sair da tela:** `Ctrl+b q` abre a fila em popup (85% da janela) por cima dos 3 painéis e, ao fechar, você volta exatamente onde estava.
*   **Governança:** PR #78 aprovado pela Camada B na tentativa 2 de 3 (duplicação de `bind-key` corrigida, `COD-01`).

### Versão 4.6.0 — Unificação Canônica do Comando `forge`, Painel Vivo Executivo e Gestão Suave de Tokens

Compatível com a 4.5.2-beta (Iniciativas 80, 81 e 82).

*   **Unificação e Padronização do Comando `forge` (Iniciativa 82):**
    *   Comando raiz Cobra padronizado para `rootCmd.Use = "forge"` (`cmd/forge-sdd/main.go`), garantindo que todas as mensagens de usage e help usem unicamente `forge`.
    *   Padronização de todas as mensagens de dicas e erros (ex: `forge ps --kill-zombies`).
    *   Templates de scaffold (`internal/scaffold/templates/`), golden files, prompts de agentes (`GEMINI.md`, `CLAUDE.md`, `.agents/rules/`, `.agents/skills/`) e documentação oficial (`docs/comandos.md`, `README.md`) instruindo 100% o comando canônico `forge`.
*   **Painel Executivo Vivo e Status Tríplice (Iniciativa 80):**
    *   Separação clara entre `forge status` (snapshot instantâneo na CLI) e `forge status --live` (living dashboard atualizado em tempo real a cada 2s).
    *   Monitoramento transparente de CPU % e RAM (MB) por sessão do projeto (`forge-discovery`, `forge-builder`, `forge-reviewer`) e seus subprocessos.
    *   Contabilidade acumulada de tokens consumidos por papel.
    *   Barra de progresso visual DAG (`[████████████████████] 100%`) com checklist detalhado de status das tarefas.
    *   Nova opção direta no menu interativo `forge`: `"Painel vivo da esteira (status --live)"`.
*   **Contabilidade de Tokens por Janela, Auto-Archive e Reset Suave (Iniciativa 81):**
    *   Monitoramento contínuo de consumo de tokens por janela de contexto dos modelos.
    *   Auto-archive automático de iniciativas quando todas as tarefas da esteira são concluídas com sucesso.
    *   Reset suave in-place nos painéis de agentes sem necessidade de reiniciar o terminal ou recriar sessões tmux.

### Versão 4.5.2 — Fix no Monitor ps: Isolamento Estrito de Processos entre Projetos (Beta)

Compatível com a 4.5.1-beta.

*   **Isolamento Estrito no `forge ps`:** impede o vazamento de processos de outros projetos que compartilham o mesmo servidor daemon do Tmux. Ao inspecionar processos de um projeto, o comando obtém a correlação exata de sessão via `tmux list-panes -a -F "#{session_name}|#{pane_pid}|..."` e nunca ingere processos de outras sessões nem o processo raiz compartilhado do daemon Tmux como raiz de busca de filhos.

### Versão 4.5.1 — Fix no Tmux: Correção na sintaxe do set-option (Beta)

Compatível com a 4.5.0-beta.

*   **Correção de prefixo em comandos de estilização do Tmux:** substitui o alvo com prefixo de igualdade (`=session`) pelo nome direto da sessão nos comandos `tmux set-option -t <session> status-left` e `status-left-length`, prevenindo o erro `exit status 1: no such session` em versões modernas do tmux (ex: tmux 3.8 no macOS) e garantindo que o launcher abra os 3 painéis integrados no tmux sem recair indevidamente para janelas nativas avulsas do sistema.

### Versão 4.5.0 — Clareza no Pipeline, Identificação Visual no Tmux, Menu Direto, DAG e Monitor ps (Beta)

Compatível com a 4.4.1-beta (Discovery 79, features `79-01` a `79-07`).

*   **Diretrizes Pipeline-First nos Prompts e Skills (`79-01`):** eliminação definitiva da confusão cognitiva sobre subagentes; os prompts e skills instruem os agentes a utilizarem os painéis especializados da esteira (`forge-builder` e `forge-reviewer`) por meio da fila determinística, reservando subagentes estritamente para tarefas efêmeras de leitura e pesquisa em paralelo.
*   **Identificação Visual do Projeto no Tmux (`79-02`):** inclusão automática do identificador do projeto no `status-left` do tmux (`[forge:meu-projeto]`), formatação dos títulos de borda dos painéis (`pane-border-format`) e padronização do nome da janela ativa (`forge:<projeto>`), evitando alternâncias acidentais entre múltiplos projetos abertos.
*   **Menu Unificado e Direto (`79-03`):** simplificação da interface do comando `forge`, eliminando a divisão artificial entre "Pipeline multi-sessão" e "Estrutura padrão". As ações principais agora aparecem diretamente no primeiro nível com números de atalho claros e diretos.
*   **Alertas Visuais e Fluxo de Entrega Não Bloqueante (`79-04`):** mecanismo de feedback em tempo real para a sessão de Discovery durante o despacho para o Builder, evitando loops ociosos de terminal e garantindo que o desenvolvedor tenha a linha de comando sempre liberada.
*   **Paralelismo de Fila por DAG e Ondas de Execução (`79-05`):** suporte nativo ao atributo `depends_on` no `queue.json`, cálculo determinístico de grafos de dependência e liberação automática de tarefas em ondas paralelas quando suas dependências são satisfeitas.
*   **Monitor Nativo de Processos `forge-sdd ps` (`79-06`):** novo comando CLI para inspecionar em tempo real os processos, servidores em background (ex: dev servers, workers, compilações) e threads ativas associadas à sessão e ao repositório, com flags `--json` e `--kill`.
*   **Status Executivo Tríplice Live (`80-01` a `80-03`):** subcomando `forge status --live` para consolidação instantânea (< 100ms e < 200 tokens) combinando captura de tela do worker ativo no tmux, commits git recentes e fila de tarefas da esteira.
*   **Documentação Encantadora e Simulador Interativo Atualizado (`79-07`):** atualização do guia do pipeline (`docs/pipeline.md`, `LEIA-ME.md.tmpl`), documentação de comandos (`docs/comandos.md`) e simulação em tempo real no emulador web da landing page (`site/src/data/emulator.ts`).

### Versão 4.4.1 — Fix no Pipeline: Gemini abre CLI agy e auto-cura em instalações anteriores (Beta)

Compatível com a 4.4.0-beta (Fix 78).

*   **Mapeamento Gemini -> agy (Antigravity CLI):** ao selecionar o agente Gemini na pipeline ou no menu, o Forge-SDD garante a invocação do executável oficial `agy` em vez da CLI descontinuada `gemini`.
*   **Tratamento de flags interativas:** omite flags de prompt interativo soltas (`-i`) quando a sessão inicializa sem prompt inicial (como nos painéis de Builder e Reviewer em modo `live`), evitando encerramento abrupto da CLI.
*   **Auto-cura e migração de instalações anteriores:** ao atualizar a versão (`forge-sdd update`) ou ao carregar qualquer configuração do pipeline (`LoadForgeConfig`), o Forge-SDD detecta perfis com `command: "gemini"` legado e normaliza/persiste automaticamente o arquivo `.forge/config/forge.json` no disco com `command: "agy"`.

### Versão 4.4.0 — Sessões com a CLI Aberta (Builder e Reviewer ao Vivo) (Beta)

Compatível com a 4.3.0-beta (Discovery 77, features `77-01` a `77-05`). Sem alteração em `forge.json`, o comportamento padrão permanece `headless`.

*   **Modo `live` de sessão (`session_mode: live`):** em vez de inicializar um novo processo da CLI a cada job, o Builder e o Reviewer mantêm seus terminais interativos abertos nos painéis do tmux, eliminando a sobrecarga de boot e permitindo acompanhar o raciocínio do agente ao vivo na tela.
*   **Despachante de entrega ao vivo (`worker --deliver <target>`):** executado em janela oculta de controle (`forge-ctl`), entrega o prompt ao painel alvo via buffers raw do tmux (`tmux load-buffer` + `paste-buffer -p` + `Enter`), garantindo integridade de aspas, quebras de linha e caracteres especiais.
*   **Limpeza de contexto automática (`context: clear`):** envia o comando de reset do perfil (`/clear` para Claude, Gemini, Copilot e Agy; `/new` para Codex) via buffer entre jobs e aguarda a estabilização do painel antes da próxima tarefa. Suporta também `context: continuous` para manter o histórico de raciocínio.
*   **Conclusão por estado compartilhado:** Builder conclui quando registra o PR no barramento (`UpdatePR` / `forge-sdd pipeline pr-opened`), enfileirando a revisão; Reviewer conclui quando aplica o veredito via `bus review` (`ApplyVerdict`).
*   **Detecção de falhas lidas da tela:** inspeciona o terminal para identificar limites de cota (`DetectUsageLimit` → requeue e pausa), agente indisponível (`DetectAgentUnavailable` → alerta de login/tier no inbox) ou painel fechado inesperadamente.
*   **Contingência suave:** quando configurado em `live` mas sem tmux ou em terminal nativo, executa fallback transparente para `headless` com aviso no terminal.

Ver `sdd/spec/adr/adr-0008-sessoes-com-cli-aberta-no-tmux.md`.

### Versão 4.3.0 — Contexto do Projeto na Discovery, Briefing e Contingência de Terminal (Beta)

Compatível com a 4.2.0-beta (Discovery 75, features `75-01` a `75-06`).

*   **`forge-sdd brief` e `--check`:** comando determinístico que lê a missão da constituição (`sdd/memory/constitution.md`), módulos (`sdd/spec/modules.md`), fluxos (`sdd/spec/flows.md`) e decisões (`sdd/spec/decisions.md`) sem alterar arquivos. O flag `--check` sai com código 0 (`pronta`) ou 3 (`rascunho`), validando se o projeto tem missão e arquitetura preenchidas antes de novas frentes.
*   **Protocolo de abertura da Discovery:** toda sessão de Discovery agora inicia executando `forge-sdd brief` para assimilar decisões anteriores, e toda especificação gerada passa a registrar explicitamente a seção `## Decisões do projeto aplicadas`.
*   **Constituição como primeiro passo (`/constitution`):** em projetos recém-criados ou com constituição em rascunho, o menu interativo do `forge` e o `forge-sdd doctor` destacam `/constitution` como o passo número 1 para evitar avanço sem diretrizes arquiteturais. Os templates de comandos de agentes foram otimizados dentro do orçamento de 300 bytes (`P-24`).
*   **Instalação assistida do tmux:** em modo `auto`, quando o tmux não está no PATH, o Forge detecta o gerenciador de pacotes do sistema (Homebrew no macOS; apt-get, dnf, pacman, apk ou zypper no Linux) e oferece instalá-lo com confirmação interativa e sudo quando necessário.
*   **Contingência suave de terminal (graceful fallback):** se o usuário recusar a instalação do tmux ou se ela falhar, as sessões abrem em janelas nativas do sistema e o Forge oferece salvar a escolha em `forge.json` (`terminal: "native"`). No Windows com Windows Terminal (`wt.exe`), as três sessões abrem em abas/painéis divididos na mesma janela. No WSL, o tmux funciona nativamente dentro do Linux.
*   **Suporte ao CLI oficial Antigravity (`agy`):** o agente Gemini passa a mapear para o binário oficial `agy` com flags correspondentes.

Ver `sdd/spec/adr/adr-0007-contexto-do-projeto-e-contingencia-tmux.md`.

### Versão 4.2.0 — Menu Navegável, Agente por Sessão e Visibilidade do Pipeline (Beta)

Compatível com a 4.1.0-beta: um `forge.json` da 4.1 continua válido e escolhe o mesmo agente nas três sessões (Discovery 71, features `71-01` a `71-05`, e Discovery 72, features `72-01` a `72-05`). Sem `forge.json` o fluxo de implementação segue como na 4.0.0-beta.

*   **`forge` abre um menu navegável (setas, Enter, Esc volta):** dois caminhos, **Pipeline multi-sessão** (iniciar as 3 sessões, ver os comandos, reconfigurar, fila e inbox, recuperar jobs presos, diagnóstico, guia) e **Estrutura padrão** (status, painel da fila, gate, doctor, instalar ou atualizar). Num projeto sem Forge a primeira opção é instalar. Sem terminal interativo, `forge` imprime a ajuda curta e sai, sem travar.
*   **Um agente por sessão:** Claude, Gemini, Copilot ou Codex em `forge-discovery`, `forge-builder` e `forge-reviewer`, escolhidos no menu (a última escolha vem marcada e é gravada em `sessions` de `forge.json`) ou por flags: `forge-sdd up --discovery claude --builder codex --reviewer gemini`.
*   **Checagem antes de abrir:** a CLI escolhida está no PATH? O projeto tem o arquivo de instruções dela? CLI ausente bloqueia; arquivo ausente avisa e indica `forge-sdd update --agent <x>`.
*   **Permissões em dois níveis:** `project` (padrão: edição e só os comandos de `allow_commands`) e `full` (bypass da CLI, só em ambiente isolado); o menu mostra o que cada um libera antes de abrir.
*   **Correção:** o worker agora usa exatamente o agente da própria sessão (`--cli` explícito). Na 4.1, abrir a Discovery com `--cli agy` deixava Builder e Reviewer no agente do `forge.json`.
*   **Cada sessão mostra o que está fazendo:** no terminal do Builder e do Reviewer aparecem linhas curtas (`✎ arquivo`, `$ comando`, batimento com o tempo decorrido) em vez de silêncio; a saída completa continua em `.forge/runtime/logs/`. Vale para qualquer agente (acompanhamento do `git status`) e, para Claude, Gemini e Codex, também pelos eventos do próprio agente. Segredos em comandos aparecem mascarados; `forge-sdd worker --raw` mostra a saída bruta.
*   **Abre pelo tmux, em três painéis:** com `tmux` instalado, `forge` e `forge-sdd up` criam uma tela com `forge-discovery`, `forge-builder` e `forge-reviewer`, cada painel com o nome da sessão (`--terminal tmux|native|auto`).
*   **Limite de uso não queima mais o job:** quando o agente avisa que bateu o limite (como o de 5 horas do Claude), o job volta à fila sem gastar tentativa, o painel mostra quando o worker retoma, o inbox avisa e o worker espera sozinho (até 6 horas; no máximo 3 devoluções por job, e limite semanal ou com data falha com o motivo). Toda falha agora traz a última linha útil do agente, não só "código 1".
*   **A Discovery não executa o build:** em modo pipeline, o `/split-features` termina entregando o job ao Builder (`forge-sdd dispatch`), e `queue next` e `queue complete` só funcionam sob um worker (ou com `--local`). Com `pm_mcp.require_existing_card: false`, o despacho não pede card nem evidência (antes exigia o arquivo mesmo assim).
*   **Codex:** o scaffold do agente `openai` passa a gerar `AGENTS.md`, que é o arquivo que o Codex lê; `OPENAI.md` vira ponteiro. Um `AGENTS.md` seu nunca é sobrescrito.

**Observações:** os perfis de Gemini, Copilot e Codex seguem a documentação oficial das CLIs, mas **nenhuma delas foi executada de ponta a ponta com o worker**, só a montagem dos comandos foi testada. No Gemini, o nível `project` libera só edição: o Builder com Gemini precisa de `full` para rodar testes. Ver `sdd/spec/adr/adr-0005-menu-e-agente-por-sessao.md` e `sdd/spec/adr/adr-0006-visibilidade-tmux-e-entrega-ao-builder.md`. **O tmux foi verificado contra o tmux 3.8 no macOS** (painéis, títulos, reconexão, `pane-base-index 1`), mas não o `attach` interativo num terminal real nem outras versões; os formatos de evento e as mensagens de limite de uso do Gemini, do Copilot e do Codex estão a confirmar. Um `forge.json` que já existe mantém os argumentos antigos dos perfis: a visão cai no acompanhamento por git até você atualizar `cli_profiles`.

### Versão 4.1.0 — Pipeline Multi-Sessão: Discovery, Builder e Reviewer em Terminais Separados (Beta)

Compatível com a 4.0.0-beta: **sem `.forge/config/forge.json` o fluxo de implementação não muda**. O pipeline só liga com `forge-sdd pipeline init`. Em contrapartida, `init` e `update` passam a instalar, em todo projeto, a pasta `.forge/` (launcher e `.gitignore` do runtime), os comandos `/inbox` e `/implement`, o agente `reviewer` e, para Claude, o hook do inbox em `.claude/settings.json` (mesclado sem sobrescrever o seu) (Discovery 70, features `70-01` a `70-07`).

*   **Três sessões que se falam por uma fila local:** `forge-discovery` (você conversa e decide), `forge-builder` e `forge-reviewer` (workers em laço) trocam jobs em `.forge/runtime/queues`. A reivindicação é atômica, com lease e recuperação de jobs presos; `high` (rework) passa na frente de `normal`.
*   **`/implementar <feature>` com gate pré-build:** confirma spec e quebra prontas, ADRs, rules e IDs existentes e exige a **evidência de que o card existe**: o agente consulta o gerenciador pelo MCP (ClickUp, GitHub Projects, Azure DevOps ou Jira) e grava `card-evidence.json`, que o Go valida (formato, provider, feature e prazo). É uma checagem de processo, não de autenticidade: quem escreve no disco pode forjar o arquivo. Sem a evidência o despacho é bloqueado e as pendências são listadas; com ela, o status é movido e o job entra na fila do Builder.
*   **Builder paralelo:** planeja ondas de tarefas independentes pelo cabeçalho `**Depende de:**`, orienta o agente a executar em paralelo (subagentes e, com `use_git_worktrees`, `forge-sdd pipeline worktree add|remove`) até `max_parallel_tasks`, valida build, lint e testes e abre o PR ligado ao card.
*   **Reviewer com ciclo de correção limitado:** confere o PR contra spec, ADRs, rules e padrões. Não conforme: devolve ao Builder com feedback estruturado (`pr_rework`, prioridade alta) até `max_rework_attempts`; no limite, ou em dúvida, escala para você. Conforme: move o card e avisa.
*   **Notificação onde você está:** o PR revisado chega ao inbox da sessão Discovery (hook no Claude, `/inbox` no Gemini e no agy) com o link, e opcionalmente toast do sistema, e-mail, webhook e comentário no card. Segredos só por variável de ambiente.
*   **`forge-up`:** `.forge/forge-up.sh` (macOS, Linux, WSL e Git Bash) e `.forge/forge-up.ps1` (Windows) abrem as três sessões nomeadas, com `claude` ou `agy`.

**Prévia dos comandos**
```bash
forge-sdd pipeline init                       # ativa o modo pipeline (forge.json e notifications.json)
.forge/forge-up.sh --cli claude               # abre as três sessões (ou .forge\forge-up.ps1)
forge-sdd dispatch auth-oauth                 # gate pré-build e despacho
forge-sdd worker --role builder               # sessão forge-builder
forge-sdd bus review --id <job> --verdict-file v.json   # aplica o veredito do Reviewer
forge-sdd inbox                               # PR revisado pronto para aprovação
```
No chat: `/implementar` e `/implement` (despacho), `/inbox` (para Claude em `.claude/commands/`, Gemini em `.gemini/prompts/` e Copilot em `.github/prompts/`) e o agente `reviewer`.

**Observações:** o ciclo completo foi testado com agente falso; **nem `claude` nem `agy` foram executados de ponta a ponta**, e o perfil padrão não traz flags de permissão (o worker precisa que o agente possa editar arquivos e rodar comandos em modo não interativo: ajuste `cli_profiles` ao seu ambiente). Os argumentos de `agy` precisam ser confirmados. O launcher de Windows e WSL só teve a geração dos comandos testada; a abertura real dos terminais depende do sistema (`forge-sdd up --dry-run` mostra os comandos).

### Versão 4.0.0-beta — Estrutura por Iniciativa, Skills por Estação e Orçamento de Contexto (Beta, mudança incompatível)

Primeira versão da linha 4.x. **Contém mudanças incompatíveis** (veja "Como migrar"). Reestrutura o fluxo para trabalhar com várias frentes em paralelo sem conflito de arquivos de estado e para consumir muito menos contexto por comando (Discovery 69, features `69-01` a `69-07`).

*   **Estado por iniciativa, sem conflito entre branches (`forge-sdd status`, `forge-sdd next-id`):** tudo de uma demanda vive em `sdd/discovery/<id>-<nome>/` (`discovery.md`, `criteria.md`, `plan.md`, `status.md`, `queue.json`, `features/` e `fixes/`). O `index.md` e o `progress.md` globais saem do fluxo; a visão geral é calculada na leitura e `--all-branches` enxerga frentes ainda não mescladas. `next-id` consulta todas as branches e não repete número. O estado de execução fica em `sdd/.runtime/`, fora do git.
*   **5 skills por estação e o comando `/implementar [feature]`:** os 13 comandos e 7 papéis viram `sdd-concepcao`, `sdd-planejamento`, `sdd-build`, `sdd-governanca` e `sdd-operacao` em `.agents/skills/`, com referências carregadas sob demanda e adaptadores finos para Claude, Gemini e Copilot. `/implementar` executa a fila da iniciativa da branch atual ou a feature indicada (`/proxima-feature` continua como alias).
*   **Menos contexto por comando:** o conjunto de leitura do `/discovery` cai de cerca de 66 KB para 22 KB e o do Builder de cerca de 32 KB para cerca de 13 KB (medido em bytes, ADR 0003). `forge-sdd context <task>` entrega só a spec, os IDs citados da Constituição, as rules e as skills da task. O check `P-24` do gate verifica os orçamentos.
*   **Migração segura (`forge-sdd update --dry-run`):** o `update` move discoveries soltos e `sdd/features/` para `sdd/archive/` com backup em `sdd/backups/`, sem sobrescrever e recusando links simbólicos, e remove comandos e papéis antigos só quando idênticos ao template anterior.
*   **Constituição enxuta:** cada regra escrita uma vez, por ID (45 IDs preservados), em até 9 KB. ADRs 0001, 0002 e 0003 registram as decisões.

**Prévia dos comandos**
```bash
forge-sdd status --all-branches      # iniciativas de todas as branches
forge-sdd next-id                    # próximo ID sem colidir com outras branches
forge-sdd context <task>             # contexto mínimo de uma task
forge-sdd update --dry-run           # plano da migração, sem gravar
```
No chat: `/implementar` (para Claude em `.claude/commands/`, Gemini em `.gemini/prompts/` e Copilot em `.github/prompts/`).

**Mudanças incompatíveis**
*   O **Forge Studio foi removido**: o comando `forge-sdd studio` não existe mais. O ACP (`forge-sdd acp`) permanece e passa a manter o heartbeat do pool.
*   `.agents/commands/` e `.agents/roles/` deixam de existir (viraram skills); `sdd/features/`, `sdd/memory/progress.md` e a fila global `sdd/.queue/` deixam de ser o fluxo.
*   A fila e o status passam a ficar na pasta da iniciativa; `sdd/.queue/{activity,stations}` migram para `sdd/.runtime/`.

**Como migrar:** rode `forge-sdd update --dry-run` para ver o plano e `forge-sdd update --upgrade` para aplicar. Arquivos de comando ou papel que você editou são preservados e listados.

### Versão 3.2.1-beta — Mascotes 3D Three.js no Studio e Visualização de Tarefa Ativa por Agente (Beta)

Atualização incremental da linha v3.2 (versão 3.2.1-beta) trazendo melhorias visuais e ergonômicas ao **Forge Studio**:

*   **Motor 3D Three.js Embarcado para os 5 Mascotes:** Integração nativa do Three.js embutido em `internal/studio/assets/three.min.js` via `embed.FS` (100% offline). Renderização 3D idêntica à do site oficial com floating bob, sombra dinâmica, olhar reativo ao mouse (`mousemove`), piscadas naturais, oscilação de antena, reações de salto (*hop*) e fallback SVG 2D automático (`CLS = 0`).
*   **Visualização de Atividade Viva por Agente:** Cada card das 5 estações agora conta com uma caixa destacada indicando exatamente o que o agente está executando no momento (`EM EXECUÇÃO`, `RETRABALHO SOLICITADO`, `AUDITANDO` ou `STANDBY`), com ID da tarefa, título e ponto pulsante em tempo real.
*   **Modo DEMO Enriquecido:** A simulação percorre sequencialmente o ciclo de vida real demonstrando o trabalho de cada um dos 5 especialistas com contratos de transição e animações dedicadas.
*   **Porta Dinâmica em Testes:** Servidor HTTP agora suporta `Port: 0` para portas efêmeras do SO, eliminando colisões de bind quando o usuário está com o Studio aberto durante execuções de teste.

### Versão 3.2.0-beta — Forge Studio (Módulo Visual Plugável), Filas Atômicas por Estação e Handoff Determinístico (Beta)

Entrega do ciclo v3.2 (versão 3.2.0-beta) integrando o módulo visual plugável do Forge Studio e o motor de filas atômicas com handoff determinístico entre as 5 estações (Feature 59, 4 entregas: `feat-59-01` a `feat-59-04`).

*   **Forge Studio & Servidor Read-Only HTTP/SSE (`forge-sdd studio`):** Servidor local isolado em `127.0.0.1:4321` sem nenhuma dependência Go externa (`net/http` e `embed.FS` nativos da stdlib — `ARQ-03`, `ARQ-04`). Rejeita mutações com `405 Method Not Allowed` (`SEG-01`, `SEG-06`) e distribui snapshots via Server-Sent Events (`/api/events`).
*   **Interface Web Embarcada com 5 Mascotes & Trilho de Handoff:** Frontend completo pré-embutido em `internal/studio/assets/` com visualização dos 5 mascotes (`Líder`, `PM & Arquiteto`, `Planejador`, `Builder` e `Governança`) usando fallback SVG 2D de alta fidelidade e `CLS = 0`, trilho processual animado com pacote em trânsito e loop de retrabalho `05 ↺ 04`, cards de status e chaveador de modo `LIVE / DEMO`.
*   **Filas Individuais por Estação com Idempotência Estrita (`forge-sdd queue handoff` e `stations`):** Persistência atômica via arquivo temporário + `os.Rename` em `sdd/.queue/stations/<station>.json`, cálculo de `IdempotencyKey` via SHA-256 e sincronização bidirecional com a fila do Builder mantendo total retrocompatibilidade com `sdd/.queue/queue.json`.
*   **Auditoria Contínua da Governança:** Cartão do Inspetor de Governança apresentando o parecer em tempo real da Camada B e conformidade determinística com os 7 pilares constitucionais (`ARQ`, `SEG`, `DUP`, `COD`, `TST`, `FLX`, `GOV`).

### Versão 3.0.1-beta.1 — Site Estático Oficial (Astro 5 SSG), Vídeo Player e Correções Visuais (Beta)

Segunda entrega beta da linha v3.0 baseada na branch `release/v3.0.1` (Feature 60 e Fixes 61 e 62). Entrega a aplicação estática oficial em `/site` para `https://forge-sdd.org`, com zero impacto sobre o binário Go (`ARQ-04`).

*   **Aplicação Estática Astro 5 SSG em `/site` (`https://forge-sdd.org`):** Arquitetura pré-renderizada no build com Core Web Vitals otimizados (`LCP < 1.2s`, `CLS = 0`), eliminação de scripts CDN externos bloqueantes e hidratação progressiva (`interactions.ts` e `mascot-3d.ts` com tree-shaking ESM e pausa via `IntersectionObserver`).
*   **SEO Técnico Completo & Schemas JSON-LD:** Canonical tags, OpenGraph, Twitter Cards, `robots.txt`, `sitemap-index.xml`, `llms.txt` e schemas Schema.org (`SoftwareApplication`, `WebSite`, `FAQPage`).
*   **Player de Vídeo Nativo Oficial:** Nova seção 06 com demonstração em vídeo (1080p Full HD, áudio stereo e controles nativos) e atalho de navegação no Header com ScrollSpy.
*   **Fix 61 & Fix 62 (Handoff Responsivo & Limpeza):** Barra viva de handoff reformulada com empilhamento fluido (`max-width: 1180px`), eliminando qualquer sobreposição horizontal entre a rota dos agentes e os artefatos; link oficial do Instagram atualizado para `@nathan.ramorim` e player de vídeo limpo sem metadados redundantes.
*   **Borda Cloudflare Workers:** Configuração declarativa em `site/wrangler.jsonc` com Static Assets e cabeçalhos de segurança HTTP com cache imutável de 1 ano para bundles em `/_astro/*`.

### Versão 3.0.0-beta — Agentes Independentes, Fila do Builder, Portão de Governança (A/B/C), Constituição por IDs e Auto-Skills (Beta)

Inaugura o ciclo major **v3.0.0-beta** (Discovery 58, 7 entregas: `feat-58-01` a `feat-58-07`), consolidando e substituindo os discoveries 53 e 56 sem dependência de modelos externos e incorporando ao core do Forge-SDD os mecanismos universais validados em `busize-solutions/planning-brain-governanca`. Publicada como Beta diretamente na branch `feat/58-orquestracao-autonoma-governanca-skills` com Pull Request mantido aberto para estabilização (Regra 11).

*   **Topologia de 5 Estações Independentes (`.agents/roles/` para Claude, Gemini e Copilot):** os papéis do Forge-SDD agora têm fonte canônica única em `.agents/roles/` (`lider`, `pm-architect`, `planner`, `builder`, `governance`, `archivist`, `migrator`), com adaptadores nativos em `.claude/agents/*.md`, `.gemini/skills/*.chatmode.md` e `.github/chatmodes/*.chatmode.md`. O **Líder** coordena o fluxo, o **PM + Arquiteto** gera `discovery` e `criteria` sujeitos ao Gate obrigatório de Revisão de Spec (`FLX-01`), o **Planejador** gera o `plan` mapeando rules por task, o **Builder** executa pela fila e o **Analista de Governança** audita em modo somente leitura.
*   **Constituição por IDs Estáveis + Rules Modulares (`GOV-04` e `DUP-10`):** `sdd/memory/constitution.md` passa a declarar "o quê e por quê" com IDs canônicos estáveis (`ARQ-*`, `SEG-*`, `DUP-*`, `COD-*`, `TST-*`, `FLX-*`, `GOV-*`), enquanto `.agents/rules/sdd-*.md` explicam "como cumprir" citando os IDs no frontmatter sem duplicar o texto da Constituição. Inclui no scaffold o Catálogo de Conceitos (`sdd/spec/concepts.md`), modelo de ADR (`sdd/spec/adr/adr-0000-modelo.md`) e template de PR com bloco de exceção com prazo (`constitution-exception`).
*   **Reuse Check Obrigatório na Spec (`/discovery`, `/split-features` e `/nova-feature` para Claude, Gemini e Copilot):** toda especificação exige as seções `Modelo do módulo` (`completo | enxuto` — `ARQ-11`), `O que já existe` (`DUP-02` — decisão de reusar, estender ou criar com justificativa) e `Dependências novas` (`SEG-07`).
*   **Motor Determinístico da Fila do Builder (`forge-sdd queue`):** novo comando CLI (`forge-sdd queue enqueue|next|complete|review|list|panel`) que gerencia `sdd/.queue/queue.json`, bloqueando o enfileiramento de specs sem Reuse Check e coordenando o ciclo `queued → building → review_pending → approved | changes_requested`.
*   **Portão de Governança em 3 Camadas (`forge-sdd gate` + `/revisar` + `/status` para Claude, Gemini e Copilot):**
    *   *Camada A (Determinística):* `forge-sdd gate --mode observacao|bloqueio` verifica automaticamente convenção de branch (`P-01`), existência e Reuse Check da spec (`P-02`/`P-03`), segredos (`P-13`), dependências novas (`P-14`), exceções dentro do prazo (`P-20`), orçamento de `progress.md` (`P-21`) e integridade de IDs entre `.agents/rules/` e a Constituição (`P-23`).
    *   *Camada B (Analista de Governança Semântico):* audita Arquitetura, Boas Práticas e Segurança (`S-01..S-07`) com proteção contra prompt injection (`FLX-08`), exigindo evidência concreta (`arquivo`, `linha`, `evidencia`, `regra`) validada por `forge-sdd queue review`.
    *   *Camada C (Painel Executivo no Líder):* `/status` para Claude, Gemini e Copilot (e `forge-sdd queue panel`) devolve ao humano a visão consolidada do que já foi feito e aprovado pela Governança, como está cada etapa ativa e o que ainda não foi desenvolvido.
*   **Motor de Auto-Skills (`forge-sdd skills`):** atividades padronizadas que se repetem pela Regra de Três (`DUP-05`) ou na revisão de Fim de Onda (`GOV-03` em `/archive` para Claude, Gemini e Copilot) são propostas e aprovadas via `forge-sdd skills propose|approve|sync|list`, gerando a skill canônica em `.agents/skills/<slug>/SKILL.md` e sincronizando automaticamente para `.claude/skills/` e `.gemini/skills/`.
*   **Importação Idempotente de Governança Externa (`forge-sdd governance`):** suporte nativo ao padrão `manifest.yaml` (`bloco` demarcado na `constitution.md`, `substituir`, `criar_se_ausente`, `mesclar`, `criar_uma_vez` e registro em `docs/governanca/VERSAO`).

### Versão 2.4.0 (Estável) — Telemetria com Cobertura Total e Relatório de Métricas

Promovida direto a estável (2026-09-05), sem passar por tag beta — desvio pontual da Regra 11, autorizado explicitamente pelo usuário; PR #53 mergeado em `main` via rebase com bypass de review (`--admin`). Revisão de código (`/code-review`) no diff antes do merge encontrou e corrigiu 4 achados: 4 prompts do Copilot (`discovery`/`archive`/`nova-feature`/`split-features`) estavam com conteúdo antigo duplicado (nunca migrado para o padrão de adaptador fino, faltando o passo de telemetria), `forge-sdd report` escrevia parte da saída fora do writer configurável do comando, `ClassifySessionType` classificava por substring em vez de prefixo (risco de falso positivo tipo `feat-42-fix-race-condition.md`), e um teste sem `gofmt`.

*   **Nenhuma sessão perdida:** `/discovery`, `/split-features`, `/nova-feature` e `/archive` agora gravam telemetria automaticamente (junto de `/proxima-feature`, `/revisar` e `/novo-fix`, que já gravavam) — antes, qualquer sessão que passasse só por esses quatro comandos não deixava rastro nenhum em `sdd/.metrics/`.
*   **Novo comando `forge-sdd report`:** mostra, por feature/fix/discovery, quantos tokens foram gastos, quais modelos de IA foram usados, quanto durou cada sessão, e há quanto tempo (medido pela telemetria) o projeto está ativo — sem precisar abrir cada arquivo de métrica manualmente.

### Versão 2.3.0 (Estável) — Ergonomia de Comandos e Sincronização

Promove para `main` (2026-09-05) todo o ciclo beta acumulado desde a v1.9.4: v2.0.0-beta (Forge-SDD Slim), v2.2.0-beta (Agent Rules e branch por feature) e esta entrega (fix-52 + discovery-03, feat-03, 6 entregas: comando do Claude quebrado, `/status` sem sincronização remota, clarify ausente, e confirmação de delegação a subagente). PRs #47, #48 e #50 mergeados em sequência via rebase merge, com bypass de review (`--admin`) autorizado explicitamente pelo usuário. Antes da promoção, uma revisão de código (`/code-review`) no diff acumulado encontrou e corrigiu 3 problemas: descrição de comando quebrada no cheat-sheet do `init` (regressão visível a todo usuário), corte de orçamento de `lessons.md` que não truncava uma entrada única grande demais, e sobrescrita silenciosa de telemetria quando duas gravações caem no mesmo segundo.

*   **`.agent/` renomeada para `.agents/` (fix-52):** ajuste de nome pedido antes de qualquer publicação estável — a fonte única de agente (`v2.2.0-beta`) nunca chegou a ser publicada com o nome antigo. `forge-sdd update` migra automaticamente projetos que já tinham `.agent/`, preservando `rules/` do usuário sem perda de conteúdo.
*   **`/nova-feature` para Claude:** `.claude/commands/*.md` (sem `.prompt`) — o Claude Code descobre slash commands pelo nome do arquivo sem `.md`, então `.claude/commands/nova-feature.prompt.md` registrava `/nova-feature.prompt`, não `/nova-feature`, quebrando o comando recomendado por todo handoff (`/nova-feature`, `/status`, etc.). Gemini (`.gemini/prompts/*.prompt.md`) e Copilot (`.github/prompts/*.prompt.md`) não mudam — mantêm a extensão correta de cada ferramenta. `forge-sdd update` remove o nome antigo automaticamente.
*   **`/status` para Claude/Gemini/Copilot:** roda `git fetch` e compara ahead/behind da branch atual e de `main` contra `origin` antes do relatório; se o VCS configurado for GitHub, cruza `gh pr list` com `sdd/features/index.md` e reporta em uma nova seção "Divergência Remota" branches órfãs ou PRs abertos não referenciados no índice — achado real desta sessão: uma branch remota com uma correção completa ficou sem PR nem entrada no índice.
*   **Clarify em `/nova-feature`, `/novo-fix` e `/discovery` (Claude/Gemini/Copilot):** os três comandos avaliam a descrição recebida contra sinais objetivos (critério de aceitação ausente, escopo com mais de uma leitura plausível, dependência externa não mencionada) e só perguntam ao usuário quando detectam lacuna real — heurística única em `sdd/memory/clarify.md`, referenciada pelos três, nunca copiada.
*   **Confirmação de delegação a subagente (Claude/Gemini/Copilot/OpenAI):** o passo `PLAN` do lifecycle (`CLAUDE.md`/`GEMINI.md`/`OPENAI.md`/`copilot-instructions.md`) passa a perguntar objetivamente, antes de cada próxima atividade/comando, se deve ser delegada a um subagente — decisão deixa de ser implícita, com critério documentado e fallback "não aplicável" para ferramentas sem esse conceito.
*   **Política de retenção de releases (feat-51):** só as 3 versões estáveis mais recentes ficam visíveis nas Releases do GitHub (mais qualquer beta ativa publicada depois delas) — limpeza automática a cada nova release. Toda versão beta passa a exibir o selo "Pre-release" e é obrigada a trazer uma lista real do que foi melhorado.

### Versão 2.2.0-beta — Agent Rules e Branch por Feature (Beta)

Pacote da discovery-02: elimina a duplicação de conteúdo de comandos entre agentes e formaliza a estratégia de branch para features quebradas em subpastas. Publicado como beta (Regra 11) sobre a `v2.0.0-beta`, ainda não promovida.

*   **`.agent/rules/*.md`:** pasta neutra na raiz do projeto para regras de domínio (design system, arquitetura, acessibilidade, ...) — qualquer agente configurado (Claude, Gemini, Copilot, OpenAI) consulta os mesmos arquivos, sem duplicar conteúdo entre `.claude/`, `.gemini/`, `.github/`. `forge-sdd update` cria a pasta se ausente e nunca sobrescreve regra já existente.
*   **`.agent/commands/*.md`:** os 13 comandos SDD passam a ter um corpo de instrução canônico único. `.claude/commands/`, `.gemini/prompts/` e `.github/prompts/` continuam existindo — é o mecanismo de descoberta de cada ferramenta (`/comando` no autocomplete de cada agente não muda) — mas agora contêm só o frontmatter/sintaxe específicos e uma referência ao corpo compartilhado. Uma mudança de comportamento de comando passa a ser feita uma vez, não três.
*   **Branch única por pasta de feature quebrada:** `/nova-feature`, `/proxima-feature` e `/novo-fix` agora tratam `sdd/features/<prefixo>-ID-<nome>/` como uma única unidade de branch, agrupando todas as subtarefas — nova Regra 15 da Constituição.
*   **Pergunta obrigatória de branch de partida e retomada:** antes de criar uma branch, os três comandos acima perguntam de onde partir (default `main`) e checam (`git branch --list <prefixo>/*`) se já existe uma branch da mesma feature/fix de sessão anterior a retomar.

### Versão 2.0.0-beta — Forge-SDD Slim (Beta)

Pacote completo de simplificação e aprendizado contínuo, resultado da discovery-01 (análise como Eng. de Software Sênior + UX Sênior). **Breaking change controlado** — publicado como beta antes de promover a versão estável.

*   **Telemetria que não falha mais em silêncio:** a gravação de `sdd/.metrics/session-*.json` deixou de depender de um passo tardio de um prompt longo — agora é feita por um comando determinístico do próprio binário (`forge-sdd session record`), disparado em `/proxima-feature`, `/revisar` e `/novo-fix`. Quem já tinha "telemetria ativada" mas via o arquivo não aparecer vai notar a diferença.
*   **`forge-sdd doctor` agora lê a telemetria, não só grava:** novo resumo agregado (sessões aprovadas/reprovadas/bloqueadas) direto no diagnóstico do projeto.
*   **Os agentes aprendem com os próprios fixes:** novo `sdd/memory/lessons.md`, atualizado automaticamente a cada correção aprovada (`forge-sdd lessons add`) e consultado por Builder/Revisor antes de implementar ou revisar — o mesmo tipo de erro não precisa ser redescoberto duas vezes.
*   **MCPs e VCS configuráveis por projeto:** a Constituição agora tem um campo para declarar se o projeto usa GitHub, Azure DevOps ou nenhum VCS automatizado, e se MCPs como `context7`/`git` estão de fato respondendo — sem mais assumir `gh pr create` e MCPs sempre disponíveis.
*   **Um fluxo, uma fonte da verdade:** `sdd/FLOW.md` passa a ser a referência única do pipeline por feature (Discovery → Split → Branch → Especificação → Build → Revisão → PR), citado por `CLAUDE.md`/`GEMINI.md`/`copilot-instructions.md` em vez de reescrito em três lugares divergentes.
*   **Menos duplicação, mesma capacidade:** a lógica de nomenclatura (sequencial/hash/workitem), antes copiada em ~14 lugares, agora vive em um único arquivo referenciado pelos três agentes.
*   **Auditoria de comandos publicada:** mapeamento dos 12 comandos de chat com recomendações de simplificação para uma próxima rodada (sem remoção nesta versão).

### Versão 1.9.4 (Estável)

**Telemetria funcionando para todos os agentes:** Antes, a gravação de métricas de uso ao final de uma sessão só acontecia de fato ao usar o agente Gemini — Claude e Copilot silenciosamente deixavam de registrar telemetria mesmo com a opção ativada no projeto. Agora os três agentes gravam a telemetria corretamente ao concluir uma feature ou correção.

### Versão 1.9.3 (Estável)

Release apenas de documentação: sincroniza a seção "Novidades" do `README.md`/`npm/README.md` (que estava travada na v1.9.0-beta) com o estado real do projeto, para que a página do pacote no NPM reflita as últimas mudanças. Sem alteração de código/comportamento.

### Versão 1.9.2 (Estável)

Esta release traz duas correções pontuais no comando `update`/scaffold.

*   **Correção na Troca de Convenção de Nomenclatura:** Agora é possível trocar a convenção de nomenclatura (`sequencial`, `hash` ou `workitem`) de um projeto já inicializado usando `forge-sdd update --naming-convention <valor>` — antes essa opção era aceita mas silenciosamente ignorada em projetos existentes.
*   **Comando `/novo-fix` disponível para Copilot:** O comando `/novo-fix`, disponível desde a v1.9.1 para Claude e Gemini, também passa a existir para o agente Copilot (o default do CLI) — antes o arquivo nunca era gerado para esse agente.

### Versão 1.9.1 (Estável)

Esta release estável consolida todo o ciclo de betas desde a v1.7.0: nomenclatura configurável, redução da curva de aprendizado para quem está começando, correções de robustez na telemetria e no diagnóstico do projeto, e correções no comando `update`.

*   **Convenção de Nomenclatura Configurável:** Escolha entre nomenclatura `sequencial`, `hash` ou `workitem` no `init`, com auto-healing e detecção de deriva de convenção no `doctor` para projetos existentes.
*   **Onboarding Mais Simples:** `init` agora imprime um cheat-sheet dos comandos disponíveis, `/status` sempre sugere o próximo passo, e um novo comando `/tutorial` guia um ciclo SDD completo fictício para quem está começando.
*   **Modo Iniciante:** `/constitution` pode gerar explicações em linguagem simplificada, com exemplos no lugar de jargão técnico.
*   **Comando `forge-sdd autopilot`:** Ativa o modo autopilot somente após um número mínimo de ciclos completos registrados em telemetria, com bypass consciente via flag.
*   **Comando `/novo-fix`:** Fluxo dedicado para criar branches e especificações de correção (`fix-`).
*   **Telemetria Mais Confiável:** Estimativa real de tokens de entrada/saída e ativação/desativação dinâmica baseada no `.sddrc`.
*   **Diagnósticos Adicionais no `doctor`:** Detecção de nome padrão de projeto não alterado (`"meu-projeto"`), de métricas granulares corretamente referenciadas em subpastas, e verificação do arquivo de métricas no local correto do projeto.
*   **Atualização Confiável para Versões Beta:** Agora é possível atualizar o CLI diretamente para a última versão beta publicada usando `update --upgrade`, sem precisar informar o número da versão manualmente. Falhas de conexão ao verificar novas versões passam a ser avisadas claramente, em vez de serem ignoradas silenciosamente.

### Versão 1.9.1-beta (Beta)

Esta release beta adiciona convenção de nomenclatura configurável, auto-healing no doctor e melhorias na telemetria.

*   **Convenção de Nomenclatura Configurável:** Configuração de nomenclatura (`sequencial`, `hash` ou `workitem`) no `init` e preservação/auto-healing nos comandos do `doctor`.
*   **Melhorias na Telemetria:** Estimativa real do input/output de tokens e ativação/desativação dinâmica de telemetria baseada no `.sddrc`.
*   **Comando `/novo-fix`:** Adicionado o comando `/novo-fix` e melhor suporte a branches e arquivos `fix-` no fluxo SDD.

### Versão 1.9.0-beta (Beta)

Esta release beta reduz a curva de aprendizado do Forge-SDD para quem está começando, na era dos agentes de IA autônomos (Discovery 5ae2).

*   **Cheat-Sheet de Comandos:** O `forge-sdd init` agora imprime a lista completa dos comandos SDD disponíveis para os agentes escolhidos ao final da inicialização.
*   **`/status` Prescritivo:** O comando `/status` (Copilot, Claude e Gemini) agora sempre sugere o próximo comando a ser executado, com base no estado real do progresso do projeto.
*   **Diagnóstico de Deriva de Nomenclatura:** O comando `forge-sdd doctor` passa a detectar quando um projeto mistura a nomenclatura sequencial (`feat-NN`) com a nomenclatura por hash (`feat-xxxx`), alertando o usuário sobre a inconsistência.
*   **Onboarding Guiado:** Novo comando `/tutorial` (Copilot, Claude e Gemini) guia o usuário por um ciclo SDD completo e fictício, isolado dos dados reais do projeto.
*   **Comando `forge-sdd autopilot`:** Novo comando CLI que ativa o modo autopilot somente após um número mínimo de ciclos completos registrados em telemetria, com bypass consciente disponível via flag (independente do loop de autopilot em si, que segue em teste na branch `feat/cli-autopilot-autonomy`).
*   **Modo Iniciante:** O comando `/constitution` (Copilot, Claude e Gemini) passa a perguntar, opcionalmente, se o usuário prefere explicações em linguagem simplificada, com exemplos concretos no lugar de jargão técnico.

### Versão 1.7.1-beta.5 (Beta)

Esta release beta traz correções e alinhamento no fluxo de descobertas (discovery).

*   **Plano de Discovery Padronizado:** O prompt de `/discovery` de todos os agentes (Copilot, Claude e Gemini) foi atualizado para gerar obrigatoriamente o arquivo `plan-XX-*.md` na pasta `sdd/discovery/` contendo o roadmap e sugestão de quebra de tarefas/features.

### Versão 1.7.1-beta.4 (Beta)

Esta release beta traz atualizações na governança de releases de teste.

*   **Agrupamento Opcional de Releases Beta:** Flexibilizada a Regra 11 da Constituição para permitir, de forma opcional, agrupar e acumular múltiplos fixes ou features na mesma branch ou versão beta antes de gerar um novo bump ou tag, otimizando o fluxo.

### Versão 1.7.1-beta.3 (Beta)

Esta release beta traz melhorias de diagnósticos da integridade do projeto.

*   **Diagnóstico de Nome Padrão do Projeto:** O comando `doctor` e os prompts `/doctor` dos agentes passam a verificar se o projeto ainda utiliza o nome genérico `"meu-projeto"`, listando as correções e os cabeçalhos de agentes recomendados a serem renomeados.

### Versão 1.7.1-beta.2 (Beta)

Esta release beta traz a correção e robustez no mapeamento de métricas de telemetria das sessões.

*   **Métricas Granulares em Subpastas:** O Orquestrador agora referencia o caminho relativo completo da feature ou subtask (ex: `sdd/features/feat-1234-auth/task-01.md`) nas métricas registradas.
*   **Registro Incondicional de Sessões:** Métricas são gravadas incondicionalmente no encerramento de sessão, inclusive cobrindo timeouts, inatividade ou encerramentos sem progresso, registrando o esforço de telemetria.

### Versão 1.7.0 (Estável)

Esta release oficial e estável consolida o ciclo de desenvolvimento de diagnósticos, refinamento e agrupamento físico de especificações.

*   **Comandos `doctor` e `destroy`:** Novo utilitário CLI para diagnóstico de saúde da estrutura SDD do projeto (`doctor`) e purga segura e interativa dos scaffolds (`destroy`).
*   **Refinamento do Comando `init`:** Criação automática e isolamento de pastas de projetos com base nos parâmetros ou formulários interativos, além de sumário de fechamento visual das stacks configuradas.
*   **Agrupamento de Features/Discoveries:** Suporte completo para agrupar fisicamente tarefas e subfeatures complexas em pastas nomeadas para organizar o escopo de entregas e planos de discovery (Regra 13 da Constituição).
*   **Prompts de IA Atualizados:** Templates de comandos (`/split-features`, `/nova-feature`, `/discovery`) e instruções de sistema atualizados para todos os 4 agentes de IA suportados (Gemini, Claude, Copilot, OpenAI).

### Versão 1.6.1-beta.3 (Beta)

- **Fase 41 — Agrupamento de Features e Discovery**:
  - **Agrupamento Físico de Features/Tasks:** As subfeatures e tarefas geradas por quebra de plano ou features complexas agora são agrupadas em subpastas com o nome da feature (ex: `sdd/features/feat-XX-nome-feature/task-YY.md`).
  - **Agrupamento pelo Discovery:** Ao quebrar um plano de discovery, a subpasta de feature criada sob `sdd/features/` reflete exatamente o nome do discovery original (ex: `sdd/features/feat-XX-nome-discovery/`).
  - **Varredura Recursiva no CLI:** O comando CLI `doctor` foi refatorado para varrer recursivamente todas as subpastas em busca de arquivos de feature (`feat-`) e task (`task-`).
  - **Instruções e Prompts Atualizados:** Ajustados os prompts de sistema e chat (`/split-features`, `/nova-feature`, `/discovery`) nos 4 agentes suportados (Gemini, Claude, Copilot e OpenAI) para impor a nova regra física.

### Versão 1.6.1-beta.2 (Beta)

- **Fase 40 — Refinamento do Comando `init` e Relatórios**:
  - **Inteligência de Pastas no CLI:**
    - `forge-sdd init meu-projeto`: Cria a subpasta `./meu-projeto/` e inicializa o scaffolding dentro dela.
    - `forge-sdd init` (sem argumentos): Abre o formulário interativo e, ao final, cria uma subpasta com o nome digitado para o projeto, inicializando nela.
    - `forge-sdd init .`: Inicializa na pasta atual, detectando automaticamente o nome do projeto com base no nome do diretório físico.
  - **Relatórios de Fechamento:** Ao concluir `init` ou `update`, exibe um resumo visual consolidado das escolhas (Nome do projeto, Stack, BD, Idioma, Agentes e Telemetria).

### Versão 1.6.1-beta.1 (Beta)

- **Fase 39 — Comandos doctor e destroy no CLI**: 
  - Adicionado o comando `doctor` para diagnóstico completo de saúde dos componentes da metodologia, detecção de arquivos faltantes e progresso de features locais.
  - Adicionado o comando `destroy` para remover com segurança e de forma interativa a estrutura inteira e configurações de agentes do repositório (com suporte a `--dry-run` e `--yes`).
- **Ajustes de UI de Update**: O prompt de atualização de agentes agora exibe os agentes existentes desmarcados com o sufixo `(Já instalado)`, evitando sobrescritas acidentais e permitindo a seleção seletiva.
- **Purga de Arquivos Obsoletos**: O CLI apaga automaticamente arquivos de comandos descontinuados (como o `/install-skill`) para evitar incompatibilidades.

### Versão 1.6.1-beta.0 (Beta)

- **Fase 38 — Detecção Inteligente e Upgrade no CLI**: O CLI agora detecta automaticamente se o diretório alvo já possui a metodologia estruturada de SDD e redireciona para a interface de upgrade.
- **Consulta ao NPM Registry**: O CLI consulta de forma assíncrona as versões oficiais (`latest`) e de teste (`beta`) direto no repositório NPM para que o usuário selecione para qual deseja atualizar.

### Versão 1.6.0 (Estável)

- **Fase 36 — Documentação Estruturada**: Estruturação de diagramas conceituais Mermaid detalhados para visualização clara de fluxos e arquitetura.
- **Fase 35 — Remoção de Skill Incompleta**: Limpeza completa da funcionalidade de instalação de skill (`/install-skill`) em todos os prompts, agentes e documentação.
- **Fase 34 — Reestruturação de Prompts por Agente**: Migração dos prompts do Copilot e unificação dos prompts de sistema e habilidades customizadas específicas para Gemini, Claude e Copilot.
- **Fase 33 — Sincronização Automática da Wiki**: A documentação oficial (diretório docs/ e README) agora é enviada e atualizada automaticamente na wiki do repositório lp-forge-sdd a cada nova tag publicada.
- **Fase 32 — Geração de Release Notes**: Agora o agente documenta as entregas de cada feature de forma automática e em linguagem clara de produto ao encerrar as tarefas.
- **Fase 31 — Preservação de Domínio no Upgrade**: As atualizações do Forge-SDD ou do CLI protegem e preservam todas as suas especificações e arquivos de domínio do usuário.
- **Fase 30 — Correções de Fluxo de Branches**: Ajustado o fluxo de desenvolvimento automático do agente para garantir criação prévia de branches e integração robusta de PRs via `gh` CLI.

### Versões Anteriores (v1.5.3 para trás)
- **Fase 29 — Consolidação de Distribuição NPM**: Remoção do suporte ao Homebrew para focar a distribuição e execução do CLI inteiramente via `npx` simplificando pipelines.
- **Fase 28 — Guia de Fluxo de Desenvolvimento**: Lançamento do Guia de Fluxo do SDD (`sdd/FLOW.md`) detalhando boas práticas de splits de feature e discovery.
- **Fase 27 — Atualização de Node no CI**: Upgrade dos workflows do GitHub Actions para utilizar Node 24 para garantir o pipeline livre de depreciation warnings.
- **Fase 26 — Posicionamento Open Source**: Nova documentação e README estruturado que estabelece formalmente o forge-sdd como código aberto para controle de fluxos IA.
- **Fase 24 — Handoff Automático**: Encerramento automático de features com geração de commit semântico e criação direta de Pull Request via linha de comando.
- **Fase 23 — MCP e Habilidades por Agente**: Configurações de ferramentas MCP e habilidades específicas isoladas e customizadas para Gemini e Claude.
- **Fase 22 — Detecção Inteligente de Instalação**: Comando `/upgrade-sdd` detecta automaticamente se o CLI foi instalado via NPX ou Brew, instruindo a forma certa de atualizar.
- **Fase 21 — Instalação de Skills via URL**: Novo comando `/install-skill` para baixar e adaptar skills de IA diretamente de links públicos do GitHub para o seu projeto.
- **Fase 20 — Refinamento C4 Model**: Diagramas Mermaid C4 simplificados e padronizados para garantir perfeita compatibilidade de renderização visual.
- **Fase 19 — Integração de Arquitetura C4**: Capacidade de descrever arquiteturas complexas usando diagramas de fluxo e sequência C4 Model representados via Mermaid.
- **Fase 18 — Protocolo de Handoff**: Mecanismo de transição estruturado entre fases de desenvolvimento para prevenir perda de contexto e memória entre agentes.
- **Fase 17 — Onboarding Simplificado**: Guias rápidos ("Getting Started") reestruturados para facilitar o onboarding em repositórios novos e já existentes.
- **Fase 16 — Comando de Constituição Automática**: Comando `/constitution` para varredura do codebase do projeto, gerando regras e o stack tecnológico automaticamente.
- **Fase 15 — Guardrails de SDD**: Validações automáticas para escrita de especificações e features mantendo a qualidade técnica das fases.
- **Fase 14 — Comando /discovery**: Introdução do `/discovery` para criar automaticamente especificações de produto e critérios técnicos refinados a partir de breves ideias.
- **Fase 13 — Documentação Multi-Agente**: Atualização dos READMEs principais e do wrapper NPM para documentar e apoiar o uso do Gemini, Claude e Copilot.
- **Fase 12 — Deploy Automatizado da Landing Page**: Pipeline automatizado no Vercel e NPM que garante a publicação contínua da landing page e a atualização do pacote NPM.
- **Fase 11 — Estruturas Unificadas de Agentes**: Padronização dos diretórios de MCP, prompts e skills entre Gemini, Claude e Copilot.
- **Fase 10 — Landing Page Oficial**: Criação da landing page oficial em `site/` para visualização online do forge-sdd e seus comandos.
- **Fase 9 — Suporte Multi-Agente**: Possibilidade de escolher Claude, Gemini ou GitHub Copilot durante o scaffold inicial para gerar configurações focadas no agente escolhido.
- **Fase 8 — Execução via NPX**: Wrapper NPM que possibilita executar o CLI via `npx forge-sdd init` sem necessidade de ambiente local com Go pré-instalado.
- **Fase 7 — Compilação Multi-Plataforma**: Automação com GoReleaser para gerar binários e publicar em GitHub Releases a partir de tags Git.
- **Fase 6 — Teste de Conformidade (Self-test)**: Cobertura de testes automatizados com golden files para atestar que atualizações não quebram o layout do scaffolding.
- **Fase 5 — Arquivos de Configuração de Versão**: Geração e controle da versão via `.sdd-version` e configurações do projeto no arquivo `.sddrc`.
- **Fase 4 — Pré-visualização com Dry-run**: Flag `--dry-run` para listar quais arquivos seriam gerados no disco sem de fato gravá-los.
- **Fase 3 — Execução Silenciosa com Flags**: Flag `--yes` para execução não interativa e simplificada do CLI ideal para automações.
- **Fase 2 — Prompt Interativo Moderno**: Interface amigável baseada em TUI (usando a biblioteca `huh`) para coletar parâmetros durante o `init`.
- **Fase 1 — Templates Embutidos no Executável**: Uso de `embed.FS` para carregar todos os templates no binário compilado eliminando dependência externa de assets.
- **Fase 0 — Base Estrutural**: Estruturação inicial do projeto Go, CLI integrado via Cobra CLI e definição dos stubs internos.
