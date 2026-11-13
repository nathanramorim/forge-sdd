# Discovery 01 — Simplificar o Forge-SDD e Fechar o Ciclo de Aprendizado Contínuo

Esta descoberta analisa o Forge-SDD sob a ótica de um Engenheiro de Software Sênior e de um UX Sênior. A discovery anterior (`discovery-5ae2`) atacou a curva de aprendizado **adicionando** superfície nova (autopilot, modo iniciante, tutorial, cheat-sheet, gate de graduação) — todas entregues e marcadas `done`. O problema persiste, agora na direção oposta: o projeto acumulou 44 tags em ~2 meses, e o próprio usuário/mantenedor não consegue mais acompanhar a evolução do que ele mesmo criou. Esta discovery propõe **cortar e consolidar sem perder capacidade**, e fechar uma lacuna concreta que a discovery anterior não endereçou: os agentes nunca aprendem com os próprios erros, e a plataforma de dados que permitiria isso (telemetria) falha silenciosamente.

## 1. Contexto e Problema (Porquê)

Evidências levantadas nesta análise, no próprio repositório:

- **Três descrições de lifecycle divergentes coexistem:** `CLAUDE.md` descreve um ciclo de 5 passos (READ-MIN → PLAN → ACT → WRITE → CLOSE); `sdd/FLOW.md` desenha um fluxograma de 7 estágios; o chatmode Orquestrador define um protocolo interno de 13 passos numerados. Nenhum dos três é gerado a partir dos outros — são três fontes da verdade que já não coincidem exatamente.
- **A própria Constituição já violou seu limite declarado:** o cabeçalho diz "Regras (máx. 12)", e hoje há 14.
- **12 comandos de chat, triplicados por agente** (Claude/Gemini/Copilot) resultam em ~36 arquivos de prompt quase idênticos. A lógica de nomenclatura (sequencial/hash/workitem) está copiada literalmente em `discovery`, `nova-feature` e `novo-fix` — quando um comportamento muda, é preciso lembrar de editar em 3 a 9 lugares (evidência real: o fix da telemetria só-Gemini precisou tocar `CLAUDE.md`, `.claude/commands/`, `.github/prompts/`).
- **Telemetria é escrita mas nunca lida.** `sdd/.metrics/session-*.json` acumula um arquivo por sessão desde a Fase 1.7, mas não existe nenhum comando, script ou seção de prompt que agregue esses dados. É uma fonte de dados morta.
- **A gravação de telemetria em si é frágil por construção — confirmado nesta análise.** Investigação dedicada nesta sessão mostrou que gravar `session-<ISO8601>.json` **não é código determinístico**: é uma instrução em linguagem natural, posicionada como o **último** passo de um prompt longo (passo 5 de 6 no Claude; passo 9 ou 13 nos chatmodes Orquestrador de Gemini/Copilot), disparada **apenas** ao final de `/proxima-feature`. Sessões que terminam em `/revisar` ou `/novo-fix` nunca gravam telemetria — comportamento hoje tratado como "intencional" na documentação de uma correção anterior (fix-50). Isso explica de forma direta o sintoma relatado pelo usuário: telemetria habilitada em `.sddrc` (confirmado correto no código Go de `init`) mas o arquivo às vezes simplesmente não aparece — porque depende de o modelo lembrar de executar um passo tardio de um fluxo com 6 a 13 etapas, e de a sessão chegar até `/proxima-feature`. Isso já gerou pelo menos 4 correções pontuais (incluindo fix-50) sem eliminar a causa raiz.
- **MCPs e VCS são tratados como universais — também confirmado nesta análise.** `context7` e `git` MCP são citados como obrigatórios em `CLAUDE.md`, no chatmode Builder e na Regra 5 da Constituição, sem nenhuma checagem em tempo real de que estão realmente respondendo. `sdd/memory/mcps.md` existe como uma tabela ("ativo"/"ativo"), mas nenhum prompt a consulta antes de usar um MCP — é decoração, não configuração. Da mesma forma, a criação de PR está hardcoded em `gh pr create --fill` em praticamente todos os fluxos; só um chatmode tem um fallback, e mesmo esse fallback ainda assume GitHub como destino. O projeto já tem `naming_convention: workitem` para IDs no estilo Azure DevOps, mas essa configuração nunca chega a influenciar como o PR é aberto — os dois pontos não se conversam. Na prática, cada projeto real do usuário roda num ambiente diferente (nem todo MCP configurado funciona; nem todo projeto usa GitHub), e o Forge-SDD não tem hoje nenhum lugar para declarar isso.
- Isso é sintoma do mesmo problema estrutural identificado na discovery anterior (deriva entre o que está documentado e o que é praticado) — mas agora manifestado como **falha operacional silenciosa**, não apenas confusão de nomenclatura.

## 2. Para Quem

- **O próprio mantenedor do Forge-SDD** — perde a visão de conjunto do que o projeto faz, porque a superfície cresce mais rápido do que ele consegue revisar/documentar de cabeça.
- **Times pequenos e devs solo já adotantes** — pagam o custo de manter 3 cópias de cada prompt sincronizadas manualmente, sem perceber (o fix-48 mostrou que `/novo-fix` simplesmente não existia para o Copilot por meses).
- **Qualquer usuário rodando em ambiente heterogêneo** (MCP parcialmente configurado, ou VCS diferente de GitHub) — hoje encontra falhas silenciosas (telemetria ausente, PR que não abre, MCP que não responde) sem nenhum mecanismo do framework para declarar "meu ambiente é diferente".

## 3. Lacunas Identificadas

1. **Três fontes de verdade para o mesmo lifecycle**, sem geração/derivação entre elas — divergência é inevitável a cada edição.
2. **Telemetria write-only**: dado é gravado, nunca lido, nenhum valor extraído.
3. **Telemetria é probabilística, não determinística**: depende de aderência do modelo a um passo tardio de um prompt longo, com um único ponto de disparo (`/proxima-feature`). Sessões que terminam em `/revisar`/`/novo-fix`, ou que abortam, não deixam rastro — quebrando qualquer plano futuro de "aprender com os dados de sessão".
4. **Zero mecanismo de aprendizado entre sessões.** Nenhum arquivo persiste "isso já quebrou antes, cuidado" — cada fix começa do zero, mesmo quando o mesmo tipo de causa raiz (ex.: duplicação de lógica entre prompts) já gerou mais de uma correção.
5. **MCPs assumidos como sempre disponíveis**, sem verificação nem fallback declarado — `mcps.md` existe mas não é lido por ninguém.
6. **VCS/work-item system hardcoded em GitHub (`gh`)** — `naming_convention: workitem` já reconhece a existência de fluxos estilo Azure DevOps na nomenclatura, mas isso nunca se propaga para a etapa de abertura de PR.
7. **Duplicação de lógica entre prompts** (nomenclatura, telemetria, MCP) — qualquer correção de comportamento precisa ser replicada manualmente em múltiplos arquivos, o que já causou pelo menos 2 fixes reativos (telemetria só-Gemini, `/novo-fix` ausente no Copilot).
8. **Superfície de onboarding grande**: 17 comandos entre CLI e chat, sem que nenhum seja claramente "opcional" ou "avançado" — tudo parece igualmente obrigatório para um novo usuário.

## 4. Oportunidades de Evolução (Como, macro)

- **Consolidar o lifecycle em uma única fonte da verdade** (um documento canônico de 5 estágios: Problema → Proposta → Refinamento → Execução → Entrega), do qual `CLAUDE.md` e os chatmodes apenas referenciam o estágio atual, em vez de reescrever os passos. Isso torna o fluxo **estável** — mudanças de detalhe de implementação não obrigam reescrever a narrativa que o usuário aprendeu.
- **Tornar a gravação de telemetria code-enforced**, não LLM-prompted: um mecanismo determinístico (subcomando do binário Go, ou hook) disparado em múltiplos pontos de saída de sessão (`/proxima-feature`, `/revisar`, `/novo-fix`), não apenas no último passo de um prompt longo. Isso é pré-requisito de tudo que depende de dados de sessão confiáveis.
- **Fechar o ciclo de telemetria** com um agregador mínimo (`/status` ou `/doctor` já existentes, sem criar comando novo) que resuma outcomes e uso — transformando dado morto em sinal útil.
- **Introduzir um artefato de aprendizado** (`sdd/memory/lessons.md`, budget pequeno) atualizado automaticamente quando um fix é aprovado, e consultado por Builder/Revisor no início da sessão — para que o mesmo tipo de erro não precise ser redescoberto.
- **Tornar MCPs e VCS configuráveis na Constituição**: o usuário declara, por projeto, quais MCPs realmente respondem e qual sistema de VCS/work-item usa (GitHub, Azure DevOps, ou nenhum). Os prompts passam a checar essa declaração antes de assumir `context7`/`git`/`gh`, com fallback explícito em vez de falha silenciosa.
- **Reduzir duplicação, não capacidade**: extrair a lógica repetida (nomenclatura, telemetria, MCP) para um único bloco referenciado pelos 3 agentes, em vez de copiado — sem remover nenhum comando existente nesta rodada, só sua reimplementação redundante.
- **Auditar comandos sobrepostos** como recomendação (não remoção automática) para uma rodada futura de simplificação de superfície.

## 5. Riscos e Trade-offs

- Consolidar o lifecycle em uma fonte única exige tocar em `CLAUDE.md`, `sdd/FLOW.md` e todos os chatmodes Orquestrador ao mesmo tempo — risco de regressão de comportamento se a extração não preservar cada passo hoje implícito.
- Mover telemetria para código determinístico é uma mudança de superfície pública do CLI (novo subcomando ou hook) — precisa respeitar a Regra 10 da Constituição (comandos públicos do CLI já fechados) ou justificar sua extensão.
- Tornar MCP/VCS configuráveis adiciona um campo novo à Constituição — vai contra o espírito de "cortar, não adicionar" desta discovery; mitigar tratando como consolidação (substitui suposições hardcoded espalhadas por um único campo declarado), não como feature nova.
- Fallback para Azure DevOps/"nenhum VCS" no lugar de `gh pr create` não deve quebrar o caminho GitHub já testado e documentado.

**Handoff:** Próximo passo (`/split-features`) deve quebrar estas oportunidades em features dentro de `sdd/features/feat-01-simplificacao-e-aprendizado-continuo/`, priorizando telemetria code-enforced primeiro (pré-requisito de dados confiáveis para as demais), seguida pelo artefato de lições e pela configuração de MCP/VCS na Constituição, com a consolidação do lifecycle e a auditoria de comandos por último (maior impacto de simplificação, mas também maior superfície de revisão).
