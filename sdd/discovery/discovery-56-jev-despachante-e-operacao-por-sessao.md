# Discovery 56 — Jev como Despachante/Decisor e Operação por Sessão

Responde à pergunta: *como tornar o Forge mais independente em seus agentes e operar por sessão, incluindo o Jev como tomador de decisões?* Complementa o **discovery-53** (esteira Spec → Act → Revisor em 3 sessões), que definiu as estações e o handoff, mas **não definiu quem decide qual estação roda, quando e com qual agente**. Este discovery preenche essa lacuna.

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

**Handoff:** próximo passo é `/split-features`, quebrando em `sdd/features/feat-56-jev-despachante-e-operacao-por-sessao/`. Antes de codar o piloto de autonomia (L2), confirmar com o usuário as premissas (a)-(c) acima e o nível de autonomia inicial. Depende do handoff estruturado do discovery-53 (Tarefas 1-3).
