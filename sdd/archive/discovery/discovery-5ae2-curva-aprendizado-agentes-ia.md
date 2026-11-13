# Discovery 5ae2 — Reduzindo a Curva de Aprendizado do Forge-SDD na Era dos Agentes Autônomos

Esta descoberta analisa o Forge-SDD sob a ótica de um Engenheiro de Software Sênior, considerando a transição do mercado de "chat assistido" para "agentes que operam com autonomia real" (subagents, background agents, MCP, auto-pilot). A proposta central não é adicionar features novas ao produto, mas **remover atrito de adoção**: hoje o valor da metodologia só é percebido depois que o usuário memoriza 11 comandos e a ordem correta entre eles. Isso é o principal fator limitante de crescimento do projeto.

## 1. Contexto e Problema (Porquê)

O Forge-SDD scaffolda uma metodologia rica (Orquestrador → Builder → Revisor, ciclo `READ-MIN → PLAN → ACT → WRITE → CLOSE`) através de 11 prompts de slash-command (`/constitution`, `/discovery`, `/split-features`, `/nova-feature`, `/proxima-feature`, `/revisar`, `/status`, `/doctor`, `/archive`, `/upgrade-sdd`, `/c4-architecture`) espalhados em três agentes (Claude, Gemini, Copilot). O único comando exposto pelo binário Go (`forge-sdd init/doctor/destroy/update`) apenas cria os arquivos — não ensina a sequência de uso.

Evidência concreta do problema, encontrada nesta análise: o próprio repositório hoje tem **três linhas de desenvolvimento divergentes** a partir do mesmo commit-base (`8a3c953`) — a convenção de nomenclatura por hash de 4 dígitos (`feat-43a2`, `discovery-9b2f`) já está em uso prático em commits recentes, mas a PR que formaliza essa regra (#26) segue **aberta e não mesclada** na `main`. Ou seja, mesmo o mantenedor do projeto — que escreveu a metodologia — está operando com um descompasso entre "o que está documentado" e "o que está sendo praticado". Se isso acontece com o autor, o efeito num adotante novo é ainda mais severo: ele lê a Constituição, vê `feat-XX` sequencial, mas encontra exemplos reais com hash. A curva de aprendizado começa quebrada pela própria inconsistência interna do sistema.

## 2. Para Quem

- **Desenvolvedores solo / times pequenos** adotando IA agêntica pela primeira vez como prática de engenharia — querem estrutura, mas se perdem em cerimônia.
- **PMs e founders "vibecoding"** — o mundo de agentes de IA está reduzindo a barreira de quem escreve código. Esse público não tem vocabulário de "C4 Model" ou "critério executável", mas é cada vez mais quem inicia o `npx forge-sdd init`.
- **Times que já usam agentes autônomos** (Claude Code subagents, Cursor background agents, Antigravity `/goal`, Copilot Workspace) e querem que o Forge-SDD converse nativamente com essas primitivas, não apenas com "chat + slash command".

## 3. Lacunas Identificadas

1. **Nenhum comando ensina a sequência.** `/status` informa o estado, mas não prescreve o próximo passo. Falta um "next step" explícito que substitua a necessidade de memorizar a ordem dos 11 prompts.
2. **Zero exemplo vivo.** Não existe um projeto-exemplo ("golden path") com um discovery→features→done completo para o usuário ler antes de fazer o seu. Aprendizado por exemplo é ordens de magnitude mais rápido que aprendizado por manual.
3. **Fratura entre CLI Go e prompts de chat.** `forge-sdd --help` não menciona nenhum dos 11 slash-commands — eles só existem como arquivos escondidos dentro de `.claude/`, `.gemini/`, `.github/`. Um usuário nunca veria a lista de comandos sem ler a Wiki externa.
4. **Deriva de convenção não detectada.** Como mostrado na Seção 1, o próprio repositório permitiu que uma convenção nova (hash) fosse usada em produção antes de ser formalizada e sem que `doctor` acusasse a inconsistência entre branches/PRs divergentes.
5. **Orquestrador/Builder/Revisor são "personas" simuladas num único chat**, não agentes isolados de fato. Isso é uma limitação técnica que aumenta a curva de aprendizado indiretamente: o usuário precisa entender que é "uma IA trocando de chapéu", um conceito não-óbvio, em vez de delegação real de tarefas — que é exatamente o que frameworks modernos (subagents, Claude Agent SDK, MCP) já resolvem nativamente.
6. **Autopilot pode ensinar o atalho antes da regra.** Com o modo `.sdd-auto-pilot` chegando (branch `feat/cli-autopilot-autonomy`), existe risco de um iniciante pular direto para autonomia total sem nunca entender o ciclo manual — o que esvazia o propósito pedagógico original da metodologia.
7. **Idioma e densidade de jargão.** Prompts em PT-BR com termos como "critério executável" e "C4 Model" pressupõem bagagem de engenharia que o público-alvo, cada vez mais amplo, nem sempre tem.

## 4. Oportunidades de Evolução (Como, macro)

- Criar um comando/fluxo de **onboarding guiado** (`/tutorial` ou `forge-sdd init --tutorial`) que roda um discovery e uma feature de exemplo de ponta a ponta dentro do próprio projeto do usuário, com dados fictícios — "aprender fazendo" em vez de "aprender lendo".
- Evoluir `/status` para **sempre recomendar o próximo comando exato** com base no estado real de `progress.md` (ex: "Você tem 2 features `todo`; rode `/proxima-feature`").
- Expor a lista de comandos SDD no próprio terminal ao final do `forge-sdd init` (cheat-sheet impresso), fechando a fratura entre CLI e prompts de chat.
- Migrar os papéis de Orquestrador/Builder/Revisor de "personas simuladas" para **definições nativas de subagentes** (ex: arquivos `.claude/agents/*.md`, equivalentes ao Agent tool usado por ferramentas como o Claude Code), aproveitando isolamento de contexto e paralelismo real — convergindo com o trabalho de autopilot já em andamento.
- Adicionar um **"modo iniciante"** de linguagem nos templates, com opção explícita durante o `/constitution` (paralelo à pergunta de idioma já implementada em `3cbe5ef`).
- Fazer o `doctor` também auditar **deriva de convenção** entre a Constituição/templates instalados e o padrão real de nomenclatura em uso (sequencial vs. hash), avisando o usuário quando os dois divergirem — o mesmo tipo de inconsistência encontrada nesta análise.
- Gate de "graduação" antes do autopilot: `doctor` só permite criar `.sdd-auto-pilot` depois que N ciclos manuais completos (`done`) forem registrados em telemetria, preservando a função didática da metodologia.

## 5. Riscos e Trade-offs

- Simplificar demais a linguagem pode diluir a precisão técnica que dá credibilidade ao método — mitigar com "modo iniciante" opcional, não substituto.
- Migrar para subagentes nativos é uma mudança arquitetural relevante e específica por agente (nem todo agente-alvo do Forge-SDD suporta subagents/Skills hoje) — precisa de investigação técnica antes de comprometer prazo (ver `criteria-5ae2`).
- Gatilhos de "graduação" para autopilot podem frustrar usuários avançados que já dominam SDD de outros projetos — considerar flag de bypass explícito e consciente (`--i-know-what-im-doing`).

**Handoff:** Próximo passo (`/split-features`) deve quebrar estas oportunidades em features dentro de `sdd/features/feat-5ae2-curva-aprendizado-agentes-ia/`, priorizando os itens do `plan-5ae2` na ordem sugerida (maior impacto na curva de aprendizado primeiro, menor risco arquitetural primeiro).
