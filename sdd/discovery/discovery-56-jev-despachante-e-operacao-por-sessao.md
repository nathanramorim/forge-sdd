# Discovery 56 — Papéis Especialistas, Agent Teams do Claude e Jev (opcional) como Despachante

Responde à pergunta: *como tornar o Forge mais independente em seus agentes e operar por sessão, incluindo o Jev como tomador de decisões?* Complementa o **discovery-53** (esteira Spec → Act → Revisor em 3 sessões), que definiu as estações e o handoff, mas **não definiu quem decide qual estação roda, quando e com qual agente**. Este discovery preenche essa lacuna.

> **Revisão 2 (agent teams):** o usuário pediu para assimilar a responsabilidade de *agent teams* do Claude Code no CLI e gerar os papéis do Forge-SDD também para o Claude. Ver seção 6; a seção 7 reavalia o desenho do Jev e da esteira à luz dos teams.

> **Revisão 3 (Jev opcional, economia de tokens, um agente por especialidade):** o usuário definiu que o Jev é **opcional**, que reduzir consumo de tokens é prioridade e que cada especialidade deve ter seu próprio agente. Ver seção 8.

> **Premissas registradas (Clarify, `sdd/memory/clarify.md`):** o usuário respondeu só à pergunta 1 (quem é o Jev). As demais foram assumidas com padrões conservadores e **precisam de confirmação no `/split-features`**: (a) este é um discovery novo que referencia o 53, não o substitui; (b) o Jev **recomenda e delega**, mas ações irreversíveis (merge, release, escopo ambíguo) continuam com o humano; (c) a decisão é assíncrona — a esteira registra a decisão e segue ou aguarda conforme o nível de autonomia.

## 1. Contexto e Problema (Porquê)

- **Quem é o Jev:** um modelo de IA novo, acessível provavelmente **só via OpenRouter**. Ele é um decisor externo aos agentes de código (Claude, Gemini, Copilot). Hoje nenhum dos três decide *sobre o processo*: o Orquestrador, dentro do próprio chat, assume mentalmente "em que etapa estamos" e "quem faz o quê".
- **Dependência do humano e do chat único.** O lifecycle (`CLAUDE.md`) exige confirmação no passo PLAN a cada comando, e o estado do pipeline vive em `progress.md` + memória da conversa. Não há registro máquina-legível de "estação X da feature Y está ocupada/concluída".
- **Agentes não são independentes entre si.** Cada agente executa o comando que o usuário digita; nenhum sabe se outro está trabalhando na mesma feature, se há conflito de arquivos entre duas features em andamento, nem se a feature já terminou por inteiro.
- **O discovery-53 cria as estações, mas sem despachante.** Sessões Claude Code Remote separadas "não se auto-coordenam" (risco registrado no 53). O Jev é a peça que faltava: o coordenador.

## 2. Para Quem

- **O mantenedor** — deixa de ser o despachante manual (qual comando agora? em qual chat? o outro já acabou?) e passa a aprovar só os pontos críticos.
- **Quem roda várias features em paralelo** — o Jev detecta impacto/conflito entre features antes de delegar.
- **Usuários de Gemini/Copilot** — recebem a decisão do Jev como instrução ("abra o Builder para a feat-X"), sem orquestração automática (paridade de resultado, não de mecanismo).

## 3. O que o Jev decide (escopo funcional)

Antes de **cada delegação**, o Jev recebe um snapshot do estado e devolve uma decisão estruturada sobre:

1. **Etapa atual** do pipeline da feature (Discovery → Split → Spec → Act → Revisor → PR).
2. **Para qual agente/estação delegar** a próxima ação.
3. **Disponibilidade:** a estação/agente está ocupado? (lease/heartbeat do ledger de sessões).
4. **Entregas:** a estação anterior terminou e entregou o handoff exigido?
5. **Impacto e conflito:** a feature em desenvolvimento colide com outra (mesmos arquivos, mesma branch, mesma spec, dependência não resolvida)?
6. **Conclusão:** a feature inteira está concluída (todas as tasks `[x]`, critério executável passando, revisão `approved`)?
7. **Ação resultante:** `delegar` | `aguardar` | `escalar ao humano` | `concluir`.

**O Jev não escreve código, specs nem faz commit.** Preserva a responsabilidade única por artefato (`CLAUDE.md`): o Jev decide; Specifier/Builder/Revisor executam.

## 4. Como (macro)

- **Ledger de execução por feature** (`sdd/.runs/<feature>.json`): estado de cada estação (`pending|running|done|blocked`), `session_id`, branch, heartbeat, handoff entregue. É a fonte de verdade "ocupado ou não"; não depende de o modelo lembrar.
- **Duas camadas de decisão:** primeiro **regras determinísticas** (barato, auditável): tasks pendentes, lease vigente, sobreposição de `files_touched`, branch existente (Regra 15). O **Jev entra só no que exige julgamento**: conflito semântico, prioridade, ambiguidade de próxima etapa. Isso reduz custo, latência e dependência do OpenRouter.
- **Registro de decisões** (`sdd/.decisions/<feature>-<ISO8601>.json`): entrada resumida, decisão, justificativa, confiança. Auditável e reaproveitável pela telemetria.
- **Níveis de autonomia** configuráveis em `sdd/.sddrc` (`dispatcher.autonomy`):
  - `L0 advisor` — Jev só recomenda; humano executa.
  - `L1 delegate-confirm` — Jev delega, humano confirma cada passo (comportamento atual do PLAN).
  - `L2 autonomous` — Jev delega e a esteira avança sozinha; **gates humanos obrigatórios** em merge/release (Regra 1 e 11), escopo ambíguo (Clarify), confiança baixa e conflito detectado.
- **Provedor via OpenRouter** com `dispatcher.provider`, `dispatcher.model` e chave em variável de ambiente (`OPENROUTER_API_KEY`) — **nunca em arquivo ou binário** (Regra 4). Modelo configurável, pois o Jev pode mudar de ID ou sair do ar.
- **Degradação segura:** Jev indisponível, timeout ou resposta inválida → cai para as regras determinísticas e, se ainda ambíguo, escala ao humano (L1). Nunca avança em silêncio.
- **Operação por sessão:** cada estação roda em sessão própria (discovery-53). Claude Code: `create_session`/`send_message` automatizam a delegação. Gemini/Copilot: o Jev emite o comando e o prompt de handoff; o humano abre o chat.

## 5. Riscos e Trade-offs

- **Dependência de terceiro (OpenRouter + modelo novo):** disponibilidade, mudança de ID, custo e latência. Mitigação: camada determinística primeiro, fallback ao humano, modelo configurável.
- **Privacidade:** enviar código/specs a um provedor externo. Mitigação: o Jev recebe **somente metadados de estado** (nomes de arquivo, status, resumos), nunca conteúdo de código nem secrets; redação antes do envio.
- **Decisão errada com autonomia alta:** o Jev pode delegar a estação errada. Mitigação: gates humanos em ações irreversíveis, `confidence` mínima, log de decisões e nível padrão `L1`.
- **Complexidade operacional:** ledger e leases exigem tratar sessão travada (heartbeat expirado → `blocked`, escala ao humano).
- **Paridade entre agentes:** automação plena só no Claude Code; declarar explicitamente, sem fingir paridade (mesmo risco do 53).
- **Regras da Constituição preservadas:** Regra 15 (uma branch por feature, todas as estações nela), Regra 14 (telemetria incondicional — decisões e sessões do Jev também gravam), Regra 8 (cliente HTTP só com stdlib `net/http`).

## 6. Papéis comuns e Agent Teams do Claude (revisão 2)

### 6.1 Lacuna encontrada no repositório
- Os templates definem **7 papéis**: Orquestrador, Specifier, Builder, Revisor, Archivist, Migrator e C4-architecture. Existem para **Gemini** (`.gemini/skills/*.chatmode.md.tmpl`) e **Copilot** (`.github/chatmodes/*.chatmode.md.tmpl`), com o corpo **duplicado** nos dois.
- O **Claude não recebe nenhum papel**: só `CLAUDE.md` e `.claude/commands/*`. Os papéis aparecem ali apenas como menção. O Claude é justamente o agente com o primitivo mais maduro para papéis isolados.

### 6.2 O que são agent teams (docs oficiais, conferidas em `code.claude.com/docs/en/agent-teams`)
- **Recurso experimental, desligado por padrão.** Habilita-se com `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` (shell ou `settings.json`). Não funciona em modo `-p` nem em sessões do Agent SDK.
- Uma sessão é o **lead**; os **teammates** são instâncias Claude Code separadas, cada uma com seu contexto, que trocam mensagens (`SendMessage`/mailbox) e compartilham uma **lista de tarefas** (com dependências e file locking).
- **Não existe arquivo de time no projeto.** `~/.claude/teams/...` é estado de runtime, removido ao fim da sessão, e não deve ser editado. Papéis reutilizáveis se definem como **subagent definitions** em `.claude/agents/*.md`, citadas ao pedir o teammate ("spawn a teammate using the `builder` agent type").
- **O que de uma definição vale para o teammate:** `tools` (limita ferramentas), `model`, `disallowedTools` e `effort` (in-process), e o **corpo** (anexado ao prompt de sistema). **`skills` não é aplicado**; `mcpServers` só em split-pane. Teammates carregam `CLAUDE.md` do projeto.
- **Hooks de qualidade:** `TeammateIdle`, `TaskCreated` e `TaskCompleted` (exit 2 bloqueia e devolve feedback).
- **Limitações relevantes:** sem retomada de teammates in-process (`/resume`), status de tarefa pode atrasar, uma equipe por sessão, sem times aninhados, custo de tokens proporcional ao número de teammates, e editar o mesmo arquivo em dois teammates gera sobrescrita.
- **Atenção:** com teams ligado, um subagent que o Claude *nomeia* passa a abrir como teammate, mesmo sem você pedir.

### 6.3 Decisão proposta: uma definição de papel, três adaptadores
- **Fonte canônica única** em `.agents/roles/<papel>.md` (mesmo padrão já usado por `.agents/commands/`). Gemini, Copilot e Claude passam a ter adaptadores finos que apontam para ela, eliminando a duplicação atual.
- **Claude:** o scaffold gera `.claude/agents/<papel>.md` com frontmatter (`name`, `description`, `tools`, `model`) e o corpo apontando para o papel canônico. A mesma definição serve como **subagent** (padrão) e como **teammate** (opt-in).
- **O Orquestrador é o lead** (a sessão principal), cujo comportamento já está no `CLAUDE.md`; a definição dele também é gerada para uso explícito.
- **Responsabilidade única passa a ser imposta por ferramenta, não só por texto.** Ex.: o Revisor sem `Edit`/`Write`; o Archivist só com leitura e edição. Ressalva: `Bash` ainda permite escrita via shell, então é uma barreira parcial, reforçada por hook e pela revisão.
- **Teams é opt-in, nunca padrão:** o scaffold não liga a variável por conta própria. O comando de init oferece a opção e avisa do custo e do efeito sobre subagents nomeados.

## 7. Reavaliação: teams, esteira por sessão e Jev

- **Onde teams ajuda de verdade** (a própria doc recomenda): revisão em paralelo com lentes distintas (Constituição, testes, segurança), pesquisa/discovery paralelo, depuração por hipóteses e **features paralelas com arquivos disjuntos**.
- **Onde teams não ajuda:** a esteira Spec → Act → Revisor de **uma** feature é sequencial e edita os mesmos arquivos — exatamente o caso em que a doc indica sessão única ou subagents. Portanto: **subagents são o padrão da esteira; teams é o modo para paralelismo.** Isso corrige a inclinação do discovery-53 para sessões Remote separadas como mecanismo principal no Claude.
- **O ledger do Forge continua necessário.** Os teams são efêmeros (config removido ao fim da sessão, sem retomada de teammates), e o ledger em `sdd/.runs/` é o que sobrevive entre sessões e serve Gemini/Copilot. Na sessão com teams, a lista de tarefas nativa é um espelho de trabalho; o ledger é a fonte de verdade.
- **Jev × lead:** o Jev continua sendo o decisor externo (OpenRouter) que **recomenda** etapa, papel, ocupação e conflito; o **lead (Claude) executa** a decisão (criar teammate/subagent com o papel indicado). O Jev não cria teammates nem edita arquivos. Em Gemini/Copilot o Jev apenas emite a instrução.
- **Gates por hook:** `TaskCompleted` pode rodar o critério executável da feature (Regra 6) e bloquear a conclusão; `TeammateIdle` pode checar se o handoff foi entregue. É a forma nativa de materializar o "entregou?" que o Jev avalia.
- **Telemetria:** cada teammate é uma sessão separada, então vira vários `session-*.json` amarrados pelo mesmo `feature` (já previsto no discovery-53, Tarefa 3).

## 8. Revisão 3: Jev opcional, economia de tokens e um agente por especialidade

### 8.1 Jev opcional (desacoplado)
- **Sem o Jev, o Forge despacha igual**, só que por regras determinísticas: o motor lê o ledger e responde etapa, papel, ocupação, conflito e conclusão sem nenhuma chamada de modelo. O Jev só entra, quando habilitado, nos casos que as regras não resolvem (conflito semântico, prioridade, ambiguidade).
- **Fronteira de módulo:** as subfeatures A (papéis), B (base determinística) e D (teams) **não importam nada do Jev**. O cliente OpenRouter vive isolado na subfeature C e pode ser removido sem quebrar o resto. Padrão `dispatcher.enabled=false`; sem `OPENROUTER_API_KEY`, o Forge funciona normalmente, sem erro.
- Consequência: o ganho de "independência de agentes" não depende de um modelo novo e de terceiros.

### 8.2 Reduzir consumo de tokens (princípios de projeto)
1. **Determinístico antes de modelo:** estado, ocupação, conflito e conclusão saem do ledger e do `git`, não de um LLM. O Jev, quando ligado, recebe só metadados (nomes de arquivo, status), nunca código.
2. **Modelo certo por papel:** cada papel declara `model` no frontmatter, com padrão econômico para os mecânicos. Proposta inicial (sobrescrevível em `.sddrc`): Archivist e Migrator em modelo leve; Builder e Specifier em modelo intermediário; Revisor e Orquestrador herdando o modelo da sessão (`inherit`) ou um modelo mais forte, pois é onde o erro custa mais. Os valores finais devem sair da telemetria, não de suposição.
3. **Contexto mínimo por especialista:** cada agente carrega só o que sua especialidade exige (READ-MIN já existente) e consulta `.agents/rules/<dominio>.md` sob demanda, em vez de o lead carregar tudo.
4. **Subagent por padrão, teams só quando paga:** a documentação oficial indica que subagents devolvem resultado resumido ao contexto principal (custo menor), enquanto cada teammate é uma instância separada (custo proporcional ao número de teammates). Teams fica opt-in, limitado a 3-5 teammates e a tarefas que justificam o paralelismo.
5. **Medir:** estender `forge-sdd report` (que já agrega tokens, modelo e duração) para quebrar por **papel/agente**, permitindo ver onde o gasto está e ajustar o modelo de cada papel com dados.
6. **Orçamentos:** respostas de planejamento e `progress.md` já têm limite; acrescentar limite de resumo de handoff por papel.

### 8.3 Um agente por especialidade
- **Núcleo (processo SDD):** os 7 papéis atuais (orquestrador, specifier, builder, revisor, archivist, migrator, c4-architecture), cada um com responsabilidade única e ferramentas restritas.
- **Especialistas de domínio (novo):** o projeto já declara conhecimento de domínio em `.agents/rules/*.md` (design system, frontend, arquitetura, acessibilidade...). Proposta: cada regra de domínio pode virar um **agente especialista** (`.claude/agents/esp-<dominio>.md`) com o corpo apontando para a regra correspondente, e o Orquestrador/Builder delegar a ele o trabalho daquela especialidade. Assim o conhecimento fica no arquivo de regra (fonte única) e o agente só o carrega quando acionado.
- **Geração explícita, não automática:** `.agents/rules/` é conteúdo do usuário e o `update` nunca o toca. Os especialistas são gerados por comando próprio (`forge-sdd agents sync`), que lê os arquivos de regra e **nunca modifica as regras**. Remover ou renomear uma regra marca o especialista como órfão, sem apagar.
- **Delegação por especialidade:** o lead escolhe o agente pelo `description` do frontmatter (mecanismo nativo de delegação). Descrições curtas e específicas evitam acionar o agente errado e gastar contexto à toa.
- **Limite saudável:** poucos especialistas bem definidos; muitos agentes com descrições parecidas pioram a delegação e aumentam custo. O `doctor` avisa quando há mais especialistas que o recomendado ou descrições sobrepostas.

**Handoff:** próximo passo é `/split-features`, quebrando em `sdd/features/feat-56-jev-despachante-e-operacao-por-sessao/` (sugestão de 5 subfeatures no `plan-56`, começando pelos **papéis comuns**, que não dependem do Jev nem de teams; o Jev é a última e opcional). Antes de codar o piloto de autonomia (L2), confirmar com o usuário as premissas (a)-(c) acima e o nível de autonomia inicial. Depende do handoff estruturado do discovery-53 (Tarefas 1-3).
