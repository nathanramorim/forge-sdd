# Discovery 93 — Coleta de contexto (a "garbage collection" das conversas)

Nasceu de um `/nova-feature`. Veio com discovery, critérios e plano porque mexe em medição, ganchos do Claude Code, worker e Settings.

## Problema
A cada chamada, o agente relê a conversa inteira. Conversa longa é entrada cara em toda mensagem. Medido em 17/11:

1. **Esta sessão do Líder é o exemplo.** No momento da medição eram 329 chamadas ao modelo, com o contexto em 607 mil tokens por chamada (média da sessão: 214 mil). Somando as chamadas, foram 71 milhões de tokens de entrada, 68 milhões relidos do cache. Em 135 das 329 chamadas o contexto passava de 200 mil.
2. **Só houve uma coleta, e por acaso.** Uma compactação automática aos 213 mil tokens, que derrubou o contexto para 23 mil. Depois disso ele subiu até 607 mil sem nenhuma outra.
3. **Nada no Forge cuida disso na conversa.** O `.claude/settings.json` do projeto só tem o gancho do inbox, e não há ajuste de compactação. O selo de tokens do tmux mostra o acumulado da telemetria, não o tamanho do contexto.
4. **O agente não consegue limpar sozinho.** `/compact` e `/clear` são comandos que só a pessoa digita. O pipeline só limpa contexto entre jobs, e só no modo `live` (`deliverContextReset`).
5. **Nos jobs o problema é menor, mas existe.** Cada job começa com 11 mil tokens. Os de fix não passam de 50 mil; os de iniciativa incham: pico de 285 mil na 90, 165 mil na 88 e 122 mil na 89. Nenhum dos 14 jobs compactou.
6. **As peças existem e não são usadas.** O Claude Code instalado (2.1.326) tem os ganchos `Stop`, `PreCompact`, `PostCompact` e `SessionStart` (com os gatilhos `compact` e `clear`) e ajustes de compactação automática (`CLAUDE_AUTOCOMPACT_PCT_OVERRIDE`, `CLAUDE_CODE_AUTO_COMPACT_WINDOW`, `CLAUDE_CODE_IDLE_COMPACT_MIN_TOKENS`). Os nomes foram achados no binário; o que cada um faz precisa ser conferido na documentação oficial.

**Quanto se ganharia:** simulando esta mesma sessão com uma coleta sempre que o contexto passasse de 150 mil tokens, a entrada cairia de 71 para cerca de 28 milhões (60% a menos, com 9 coletas). É estimativa: ela não conta o que o agente relê depois de uma coleta nem a qualidade do resumo.

## Gancho ou habilidade?
- **Gancho (hook):** um comando que o Claude Code roda sozinho quando algo acontece ("o agente terminou de responder", "a conversa foi compactada"). Não depende de o modelo lembrar. Serve para o que tem de acontecer sempre, na hora certa.
- **Habilidade (skill):** um texto de instruções que o modelo lê quando o assunto aparece. Serve para ensinar como fazer. Depende de o modelo decidir usar.
- **Aqui:** a hora de coletar é trabalho de gancho e de código Go. O que guardar no resumo é instrução. A coleta usa os dois, cada um na sua parte.

## Para quem
Dono do projeto que mantém a conversa com o Líder aberta por horas ou dias e deixa a esteira rodando.

## Jornada
1. Em **Settings**, liga "Coleta de contexto". Os valores iniciais já vêm preenchidos (meta de 100 mil e teto de 150 mil tokens).
2. O selo do painel e o `forge status` passam a mostrar o contexto real de cada sessão.
3. Quando um comando fecha (estado salvo em arquivo) e o contexto passou da meta, a coleta acontece naquele intervalo. Se o Claude Code não oferecer isso de forma pública, o Forge avisa em uma linha.
4. Se o contexto chegar ao teto no meio de um trabalho, o próprio Claude Code compacta, guiado pelas instruções do projeto.
5. Depois de qualquer coleta, o Forge devolve à conversa um cartão de retomada: branch, iniciativa, `status.md` e arquivos sem commit.
6. Nos jobs do Builder e do Reviewer vale o mesmo teto, e o relatório mostra o pico de contexto e as coletas de cada job.

## Regras de negócio
- **RN-1 Quem decide a hora é o código** (gancho mais Go), com medida real. O modelo não precisa lembrar.
- **RN-2 Dois limites:** a **meta** vale em ponto seguro (comando fechado e estado em arquivo); o **teto** vale a qualquer hora.
- **RN-3 Compactar, não apagar:** a coleta automática sempre resume. Limpar tudo (`/clear`) continua sendo escolha da pessoa.
- **RN-4 Antes e depois:** as instruções do projeto dizem o que o resumo guarda (decisões e preferências da pessoa, premissas a confirmar, pendências, branch e iniciativa) e o que descarta (saídas de ferramenta, diffs, logs). Depois da coleta, o estado do projeto volta pelo cartão de retomada.
- **RN-5 Sem digitar no painel da pessoa:** no teto, quem compacta é o próprio Claude Code. Na meta, usa-se a compactação em ociosidade do Claude Code se ela for pública; senão, um aviso de uma linha, no máximo a cada 10 minutos.
- **RN-6 Só ajuste documentado:** nenhum ajuste interno ou experimental do Claude Code entra sem estar na documentação oficial.
- **RN-7 Jobs:** teto por job. Um job por task (91 e 92) continua sendo a principal prevenção.
- **RN-8 Prevenir antes de coletar:** com a coleta ligada, o trabalho pesado do Líder (ler código, escrever discovery e specs) roda em subagente e volta só o resumo e os arquivos. Depende da 91-02.
- **RN-9 Medir:** contexto atual, pico e coletas, por sessão e por job, no selo e no relatório. Do transcrito só se leem contadores (`SEG-05`).
- **RN-10 Nada muda para quem não ligar.** Nesta onda vale para o `claude`; nas outras CLIs o Settings avisa que não tem efeito.

## Escopo
**Dentro:** medidor de contexto real (sessão e job); decisão de coleta em Go e comando `forge session gc`; ganchos e teto mesclados no `.claude/settings.json`; instruções de compactação; cartão de retomada depois de compactar ou limpar; teto nos jobs; chave no Settings, selo e relatório; subagente por padrão no trabalho pesado do Líder; seção "Gancho ou habilidade?" na documentação; ADR-0018.

**Fora:** o Forge digitar `/compact` no painel da pessoa pelo tmux (pode colidir com o que ela está escrevendo); limpar a conversa sozinho; coleta em Gemini, Codex e Copilot; encurtar a saída das ferramentas e o mapa de código (a maior parte do que entra no contexto); mudar o reset por job do modo `live`.

## Riscos
| Risco | Mitigação |
|---|---|
| O resumo perder uma decisão sua | Instruções de compactação, cartão de retomada e `status.md` como fonte; a meta só vale em ponto seguro |
| Ajuste do Claude Code mudar de nome ou não ser público | RN-6; sem ele, sobram o aviso e a medição |
| Coletar demais custar mais do que economiza | Meta e teto configuráveis; o relatório mostra as coletas e o contexto médio |
| O teto ser em percentual e depender da janela do modelo | Preferir ajuste em tokens; senão, usar a janela registrada nos logs dos jobs |
| O gancho deixar cada resposta mais lenta | Lê só o fim do transcrito; limite de 100 ms em teste |
| O transcrito conter dado sensível | Só contadores são lidos; nada dele vai para log ou telemetria |
| Subagente devolver resumo pobre | Os artefatos ficam em arquivo; o Líder confere pelo arquivo, não pelo resumo |
