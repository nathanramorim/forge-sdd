# Discovery 104 — Pipeline em subagentes no mesmo chat, com harness verificável

Palavras do dono (19/11/2026): "A pipeline deve funcionar no mesmo chat, porém em subagentes (visão de teams: camada de entrada; até o Discovery pode ser subagente)"; depois: "junta a feature com tudo que disse aqui e vamos gerar um discovery".

## Problema
1. Hoje o pipeline são 3 sessões de terminal (tmux, barramento em `.forge/runtime/`, workers headless; ADR-0004, ADR-0008). Funciona só no terminal; o dono quer o mesmo fluxo no chat da pessoa, inclusive no app desktop e no Codex.
2. O diagrama do dono: camada de entrada (CLI, Claude, Codex, Agy) -> Agente Discovery (líder + subagentes A/B) -> Agente Builder (líder + A/B em paralelo) -> Agente Reviewer (líder + A/B) -> Resultado Final, com "Retorno" tracejado do Reviewer ao Líder Builder.
3. O processo atual tem furos de harness, com evidência real: (a) a skill de Governança exige `regra_descricao` no parecer e o `forge queue review` do binário usado recusou o campo ("unknown field"): skill e binário divergiram (na `main` atual `queue.go:66` já tem o campo; a divergência é de versão e some só com verificação automática); (b) workers autônomos construíram tasks no worktree em que outra sessão trabalhava, sem trava; (c) três critérios executáveis erraram pelo defeito do próprio teste (`grep -c` em SVG de uma linha; pipeline que só passa sem saída e quebra sob `bash -e`; extrator que juntou o bloco "como regenerar" aos critérios).
4. Não há diagnóstico do harness ao abrir a sessão, nem como comparar desenhos (3 sessões x subagentes) por números.

## Para quem
Dono do projeto e pessoas que usam o forge no terminal, no app desktop do Claude Code e no Codex; querem uma esteira confiável, medida e barata.

## Jornada
1. Abre qualquer superfície; o diagnóstico do harness roda (hook ou instrução) e avisa de divergências.
2. Conversa no chat; o papel Discovery produz a spec; o humano aprova o Gate 1 (FLX-01).
3. O chat coordena o Builder em paralelo (subagentes, um worktree cada) e o Reviewer; o Go aplica o veredito e o retorno.
4. Resultado final: PR conforme ou escalada ao humano, com rastro por nó (quem, contexto, custo, motivo de parada).
5. A avaliação compara o modo `sessions` com o modo `chat` nas mesmas features de referência.

## Viabilidade técnica (evidência; decide o desenho)
Confirmado nesta sessão (arquivos e binário):
- `.claude/agents/*.md` define as ferramentas por papel; nenhum (builder, reviewer, governance, lider, planner, pm-architect, archivist, migrator) tem a ferramenta `Agent`. Esta própria sessão (agente lançado) também não recebeu `Agent`. Indício forte de que subagente não cria subagente; não é prova da regra.
- `claude` 2.1.329: o `--help` lista `--agent`, `--agents <json>`, `claude agents` ("background agents") e `--forward-subagent-text`; não lista flag de time. O binário contém `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` (6 ocorrências), `TeamCreate` (2), `SendMessage` (116) e a mensagem "A subagent cannot be started from here in this session. Do the task without a subagent.". Logo agent teams existem como recurso experimental atrás de variável de ambiente.
- O Builder de hoje paraleliza com subagentes (`.agents/skills/sdd-build/references/worker-job.md`, `internal/pipeline/prompt.go:102`) porque roda como sessão de topo (`claude -p`), não como definição `.claude/agents/builder.md`.
- `gemini` tem `-p`, `--acp`, `extensions`, `skills`; `agy` tem `--agent`, `--remote-agent`, `-p`, stream-json e o subcomando `acp`; ambos sem menção a subagentes no `--help`. O repo já tem `internal/acp` (JSON-RPC stdio/HTTP).
Suposição (não verificada nesta sessão): subagentes não criam subagentes; um time (teams) é de um nível, com mensagens entre membros; o app desktop do Claude Code usa os mesmos hooks `SessionStart`; Codex e Agy têm algum recurso de subagente.
Não consegui verificar: o `claude-code-guide` (sem ferramenta `Agent` aqui), a documentação oficial de teams, a semântica de `SendMessage`, o Codex (não instalado) e os subagentes do Agy e do Gemini.
Consequência: o desenho literal do diagrama (líder de cada papel criando subagentes) não é garantido. Alternativas:
- **A. Chat como único coordenador (recomendada).** O chat (sessão de topo) carrega o papel de cada estação e cria os subagentes A/B (um nível). O "líder" é o papel/skill no chat, com o laço no Go. Prós: um nível, funciona hoje, Builder paralelo preservado, mesmo chat. Contras: o contexto do chat cresce; "líder" é papel, não agente.
- **B. Híbrido: chat dispara os workers headless atuais** (`forge dispatch` e worker via Bash em background). Prós: reaproveita 100% o barramento, paralelismo real. Contras: continua tmux/Go no meio, não é "subagente"; custo de boot por job.
- **C. Agent teams** (líder + teammates por papel). Prós: fiel ao desenho e a mensagens entre agentes. Contras: experimental, um time por sessão (suposição), só Claude, sem garantia no desktop.
- **D. Literal (líder-agente cria subagentes).** Só se a feat-104-06 (spike) confirmar aninhamento; hoje improvável.
Recomendação: A, com B como convivência (modo `sessions` continua) e C como spike condicional; D descartada até prova.

## Regras de negócio
- RN-1 O Go é o dono do laço (ordem, limites, estados, retorno); o modelo decide conteúdo (ADR-0004 D-3, ADR-0012).
- RN-2 Gate 1 humano, salvo `autonomy: advanced` (ADR-0013).
- RN-3 Reviewer e seus subagentes são somente leitura (SEG-06); feedback de retorno é dado, não instrução (FLX-08).
- RN-4 Cada subagente paralelo tem worktree próprio; há lock por iniciativa.
- RN-5 Teto de voltas (hoje 3, `max_rework_attempts`), achados com arquivo e linha, escalada ao humano no teto.
- RN-6 Permissão pela lista de ferramentas do papel, imposta pelo harness, não pedida no prompt.
- RN-7 O binário só diagnostica; corrigir é do agente, com confirmação humana.
- RN-8 Convivência: o modo de 3 sessões continua válido; o novo é opt-in (`pipeline_mode`).

## Escopo
Dentro: contratos entre etapas; verificador do verificador; `forge doctor --harness`; servidor MCP `forge_doctor`; hook e instrução; modelo por papel; laço no Go com lock e retorno estruturado; isolamento por subagente; papéis no chat; camada de entrada multiagente; rastreio por nó; conjunto de avaliação; convivência.
Fora: reescrever o barramento; remover tmux ou o modo `sessions`; agent teams em produção (só spike); subagentes do Codex e Agy além da matriz de capacidades; UI gráfica; correção automática pelo diagnóstico; push, PR e enfileiramento (Gate 1 é do dono).

## Riscos
Subagentes não aninham (mitigado pela recomendação A); estouro de contexto do chat; Codex e Agy sem recurso equivalente (nivelar pelo menor denominador); hook ausente no desktop (fallback por instrução); custo de tokens maior no chat longo (medido pela avaliação).

## A confirmar no Gate 1
1. Desenho A (chat como coordenador, um nível de subagentes) é o aceito? O spike da feat-104-06 pode mudar para D.
2. A caixa "CLI" do diagrama é o próprio `forge` (padrão conservador) ou um terminal genérico?
3. O app desktop do Claude Code usa os mesmos hooks e MCPs? (suposição; só a instrução vale se não)
4. Agent teams: vale o spike ou fica fora de vez?
5. Codex e Agy: sem recurso de subagente confirmado, caem para o modo `sessions` com aviso?
6. `regra_descricao`: a divergência veio de binário antigo? Na `main` o campo existe.
7. Subcomandos novos (`criteria verify`/flag em `gate`, `mcp serve`, `eval`) mudam o texto de `ARQ-06`; aceita a ADR-0015 cobrir isso?
8. Os critérios das features ainda não trazem o marcador `# negativo:` da feat-104-02; entram na revisão após ela.

## Classificação dos itens pedidos
Feature: A1 (11), A2 (01), A3 (06), A4 (06, 07), A5 (06), A6 (05, 08), A7 (09), A8 (12), A9 (02), B (03, 04), C (05).
Critério transversal: teste negativo em toda feature, `bash -n` nos blocos, convivência com `sessions`.
Fora: remover o modo `sessions`, agent teams em produção, correção automática pelo diagnóstico, subagentes nativos do Codex/Agy.
