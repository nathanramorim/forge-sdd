# feat-114-01 — Reutilização do código existente como parte obrigatória da proposta e diretriz explícita nos agentes do Forge SDD

- **Iniciativa:** `114-reutilizacao-codigo-existente-na-proposta`
- **Branch:** `feat/114-reutilizacao-codigo-existente-na-proposta`
- **Status:** `spec_ready` (aguardando Gate 1 — `FLX-01`)

## Descrição de negócio

- **Problema:** Hoje a regra `DUP-02` exige a tabela `## O que já existe` em `criteria.md` e `feat-*.md`, mas a reutilização do código existente ainda é tratada apenas como um quadro burocrático de spec, sem aparecer como parte obrigatória da **proposta apresentada ao usuário** desde o primeiro momento (em `discovery.md`, `/nova-feature`, `/novo-fix`, `/split-features` e no resumo de Gate 1) nem como diretriz explícita nos adaptadores e instruções dos agentes do Forge SDD (`pm-architect`/`specifier`, `planner`, `builder`, `reviewer`/`governance`/`revisor`, `lider`/`orquestrador`).
- **Para quem:** Desenvolvedores, arquitetos e agentes de IA (Claude Code, Gemini CLI e GitHub Copilot) que utilizam o Forge SDD para conceber, planejar, implementar e auditar evoluções de software.
- **Valor:** Garante que toda proposta apresentada ao dono explicite primeiro quais pacotes, funções, contratos e testes existentes serão reutilizados ou estendidos antes de propor qualquer código novo, reduzindo duplicação estrutural (`DUP-01`, `DUP-02`, `DUP-03`), economizando tokens/retrabalho e alinhando todos os agentes das três ferramentas a partir de uma fonte única (`DUP-10`, `TST-02`).

## Problema

Embora a Constituição (`DUP-01`, `DUP-02`, `DUP-03`) e o check determinístico `P-03` (`internal/gate/gate.go`) validem a presença da tabela `## O que já existe` na spec, existem cinco lacunas operacionais nas instruções atuais dos agentes:

1. **Na Concepção (`sdd-concepcao` e `sdd-reuso-e-duplicacao.md`):** `discovery.md` e o resumo da proposta submetido ao usuário no Gate 1 (`references/discovery.md`, `references/nova-feature.md`, `references/novo-fix.md`) não exigem que a proposta destaque explicitamente o código existente que será reutilizado/estendido como parte central da solução proposta.
2. **No Planejamento (`sdd-planejamento`):** `SKILL.md` e `references/split-features.md` verificam a presença de `O que já existe` na spec de entrada, mas não exigem que cada task do `plan.md` e cada `feat-<id>-<nn>-*.md` explicitem os símbolos, pacotes e arquivos existentes que a task irá reutilizar ou estender.
3. **No Build (`sdd-build`):** `SKILL.md`, `references/implementar.md` e `references/worker-job.md` mencionam `O que já existe` de forma genérica, sem instruir explicitamente o Builder a priorizar a extensão/reúso dos pacotes e funções mapeados na proposta e a proibir a criação de função, struct ou helper duplicado (`DUP-03`).
4. **Na Governança e Revisão (`sdd-governanca`):** `references/revisar.md` e `references/pr-review.md` não detalham que a Conferência da spec e do código deve reprovar (`S-02` / `DUP-02` / `DUP-03`) qualquer proposta ou PR que omita o que será reutilizado do código existente ou que crie código novo quando já havia equivalente reutilizável.
5. **Nos Agentes do Forge SDD (`.claude/agents/*.md`, `.gemini/skills/*.chatmode.md`, `.github/chatmodes/*.chatmode.md` e templates do scaffold):** os adaptadores de papel das três ferramentas ainda não declaram de forma uniforme que a reutilização do código existente faz parte obrigatória de toda proposta, plano, entrega e auditoria.

## Comportamento

A implementação desta feature atualizará a fonte canônica em `.agents/`, os adaptadores das 3 ferramentas (`.claude/`, `.gemini/`, `.github/`), seus templates em `internal/scaffold/templates/` e os goldens em `internal/scaffold/testdata/golden/`:

### 1. Concepção e Regra de Reúso (`.agents/rules/sdd-reuso-e-duplicacao.md` e `.agents/skills/sdd-concepcao/`)
- **`.agents/rules/sdd-reuso-e-duplicacao.md`** (e template `.tmpl` + golden):
  - Atualizar o frontmatter para `regras: [DUP-01, DUP-02, DUP-03, DUP-05, DUP-10, GOV-04]` (incluindo `DUP-03` sem copiar linhas literais da Constituição, respeitando `P-23`).
  - Explicitar na seção 1 que **reutilizar o código existente é parte obrigatória de toda proposta**: tanto na tabela `## O que já existe` quanto na narrativa da proposta (`discovery.md`, spec de feature/fix e resumo entregue ao dono no Gate 1), o agente deve listar quais pacotes, funções, structs, contratos e testes existentes serão reutilizados ou estendidos antes de propor qualquer item novo, sendo `criar` permitido apenas quando provado que o código existente não atende (`DUP-01`, `DUP-02`, `DUP-03`).
- **`.agents/skills/sdd-concepcao/references/discovery.md`**, **`references/nova-feature.md`** e **`references/novo-fix.md`** (e templates `.tmpl` + goldens):
  - Em `discovery.md`: exigir que a proposta de produto/arquitetura (`discovery.md` e `criteria.md`) e a apresentação no Gate 1 sempre incluam o que será reutilizado ou estendido do código existente (`sdd/spec/concepts.md`, pacotes `internal/`, comandos `cmd/` e skills) antes de qualquer novo componente (`DUP-01..03`).
  - Em `nova-feature.md` e `novo-fix.md`: exigir que a proposta apresentada ao usuário antes do Gate 1 explicite os pacotes, funções e contratos existentes reutilizados/estendidos junto com `## O que já existe`, vedando criar símbolo novo se houver equivalente reutilizável (`DUP-02`, `DUP-03`).
- **`.agents/skills/sdd-concepcao/SKILL.md`** (e template `.tmpl` + golden):
  - Atualmente com **2.045/2.048 bytes**: se editado, enxugar palavras equivalentes para manter estritamente `<= 2048` bytes (`P-24`), reforçando no item 4/5 que toda proposta explicita o reúso do código existente (`DUP-02`, `DUP-03`).

### 2. Planejamento (`.agents/skills/sdd-planejamento/`)
- **`.agents/skills/sdd-planejamento/SKILL.md`** e **`references/split-features.md`** (e templates `.tmpl` + goldens):
  - Exigir que cada task do `plan.md` e cada arquivo `features/feat-<id>-<nn>-<nome>.md` indique explicitamente os arquivos, pacotes e símbolos existentes que serão reutilizados ou estendidos na task (`DUP-02`, `DUP-03`), propagando o mapeamento da proposta para a fila.

### 3. Build (`.agents/skills/sdd-build/`)
- **`.agents/skills/sdd-build/SKILL.md`**, **`references/implementar.md`** e **`references/worker-job.md`** (e templates `.tmpl` + goldens):
  - Instruir o Builder a priorizar estender e reutilizar os pacotes, funções e contratos existentes mapeados na proposta e em `## O que já existe`, sendo proibido duplicar função, struct, constante ou helper já existente no repositório (`DUP-02`, `DUP-03`).

### 4. Governança e Reviewer (`.agents/skills/sdd-governanca/` e `.agents/skills/sdd-operacao/references/lider.md`)
- **`.agents/skills/sdd-governanca/references/revisar.md`** e **`references/pr-review.md`** (e templates `.tmpl` + goldens):
  - Explicitar que na revisão de spec (Conferência 1) e na revisão de código/PR (Conferência 2), o Analista de Governança / Reviewer reprova com `violacao` (`S-02` / `DUP-02` / `DUP-03`) qualquer proposta, spec ou PR que omita o que será reutilizado do código existente ou que crie pacote, função, struct ou helper novo quando já havia equivalente reutilizável.
- **`.agents/skills/sdd-governanca/SKILL.md`** (atualmente com **2.026/2.048 bytes**):
  - Manter `<= 2048` bytes (`P-24`); se atualizado, enxugar texto equivalente.
- **`.agents/skills/sdd-operacao/references/lider.md`** (e template `.tmpl` + golden):
  - No Gate 1 e no despacho da fila, o Líder/Orquestrador confirma que a proposta apresentada ao humano explicita a reutilização do código existente (`DUP-02`, `DUP-03`).

### 5. Adaptadores dos Agentes nas 3 Ferramentas + Templates e Goldens (`DUP-10`, `P-24`, `TST-02`)
Atualizar as descrições/instruções curtas dos agentes em Claude Code, Gemini CLI e GitHub Copilot (e seus pares em `internal/scaffold/templates/` e `internal/scaffold/testdata/golden/`) para declarar explicitamente a reutilização do código existente na proposta/plano/build/revisão, respeitando o teto `MaxRoleAdapterBytes = 420` bytes:

- **Claude Code (`.claude/agents/`):**
  - `pm-architect.md` (e `.tmpl` + golden)
  - `planner.md` (e `.tmpl` + golden)
  - `builder.md` (e `.tmpl` + golden)
  - `reviewer.md` (e `.tmpl` + golden)
  - `governance.md` (e `.tmpl` + golden)
  - `lider.md` (e `.tmpl` + golden)
- **Gemini CLI (`.gemini/skills/`):**
  - `specifier.chatmode.md` e `c4-architecture.chatmode.md` (e `.tmpl` + goldens)
  - `planner.chatmode.md` (e `.tmpl` + golden)
  - `builder.chatmode.md` (e `.tmpl` + golden)
  - `reviewer.chatmode.md`, `governance.chatmode.md` e `revisor.chatmode.md` (e `.tmpl` + goldens)
  - `orquestrador.chatmode.md` (e `.tmpl` + golden)
- **GitHub Copilot (`.github/chatmodes/`):**
  - `specifier.chatmode.md` e `c4-architecture.chatmode.md` (e `.tmpl` + goldens)
  - `planner.chatmode.md` (e `.tmpl` + golden)
  - `builder.chatmode.md` (e `.tmpl` + golden)
  - `reviewer.chatmode.md`, `governance.chatmode.md` e `revisor.chatmode.md` (e `.tmpl` + goldens)
  - `orquestrador.chatmode.md` (e `.tmpl` + golden)

### Tabela de Orçamento de Contexto `P-24` (`internal/gate/gate.go:785-790`)

| Arquivo / Grupo Alvo | Tamanho Atual (bytes) | Limite `P-24` (bytes) | Estratégia na Implementação |
|---|---|---|---|
| `.agents/skills/sdd-concepcao/SKILL.md` | 2.045 | 2.048 (`MaxSkillBytes`) | Enxugar palavras equivalentes se editado; detalhar nas `references/` |
| `.agents/skills/sdd-governanca/SKILL.md` | 2.026 | 2.048 (`MaxSkillBytes`) | Enxugar palavras equivalentes se editado; detalhar em `revisar.md` e `pr-review.md` |
| `.agents/skills/sdd-build/SKILL.md` | 1.511 | 2.048 (`MaxSkillBytes`) | Folga de 537 bytes; manter `<= 2048` bytes |
| `.agents/skills/sdd-operacao/SKILL.md` | 1.433 | 2.048 (`MaxSkillBytes`) | Folga de 615 bytes; detalhar em `references/lider.md` |
| `.agents/skills/sdd-planejamento/SKILL.md` | 1.247 | 2.048 (`MaxSkillBytes`) | Folga de 801 bytes; manter `<= 2048` bytes |
| `.claude/agents/{pm-architect,planner,builder,reviewer,governance,lider}.md` | 310 a 365 | 420 (`MaxRoleAdapterBytes`) | Folga de 55 a 110 bytes; manter cada arquivo `<= 420` bytes |
| `.gemini/skills/{specifier,c4-architecture,planner,builder,reviewer,governance,revisor,orquestrador}.chatmode.md` | 123 a 180 | 420 (`MaxRoleAdapterBytes`) | Folga de >240 bytes; manter cada arquivo `<= 420` bytes |
| `.github/chatmodes/{specifier,c4-architecture,planner,builder,reviewer,governance,revisor,orquestrador}.chatmode.md` | 238 a 339 | 420 (`MaxRoleAdapterBytes`) | Folga de 81 a 182 bytes; manter cada arquivo `<= 420` bytes |
| `.agents/rules/sdd-reuso-e-duplicacao.md` e `.agents/skills/sdd-*/references/*.md` | 744 a 3.078 | Sem teto fixo em `P-24` (`P-23` valida rules) | Respeitar `P-23` (sem cópia literal da Constituição) e manter instruções enxutas |

## Modelo do módulo

**Enxuto** (`ARQ-11`): a feature não cria pacote Go novo nem altera contratos de CLI; estende apenas as rules, skills/references canônicas em `.agents/`, os adaptadores de agentes em `.claude/`, `.gemini/` e `.github/`, e os respectivos templates (`internal/scaffold/templates/`) e goldens (`internal/scaffold/testdata/golden/`) já gerenciados pelo pacote `internal/scaffold`.

## Decisões do projeto aplicadas

- **ADR-0003** (`skills-por-estacao-e-orcamento`, `aprovado`): preserva a divisão nas 5 skills de estação (`sdd-concepcao`, `sdd-planejamento`, `sdd-build`, `sdd-governanca`, `sdd-operacao`), mantendo adaptadores finos por ferramenta e respeitando rigorosamente os tetos de bytes do check `P-24`.
- **ADR-0004** (`pipeline-multi-sessao`, `aprovado`): alinha as instruções tanto do fluxo de sessão única quanto dos workers do pipeline (`references/worker-job.md` e `references/pr-review.md`).
- **DUP-01, DUP-02, DUP-03**: reforça a consulta ao catálogo `sdd/spec/concepts.md`, a tabela `## O que já existe` e a obrigação de reutilizar/estender código existente como parte central da proposta antes de criar código novo.
- **DUP-10**: mantém o corpo canônico das instruções em `.agents/rules/` e `.agents/skills/sdd-*/`, usando `.claude/agents/`, `.gemini/skills/` e `.github/chatmodes/` apenas como adaptadores finos sincronizados com `internal/scaffold/templates/`.
- **GOV-04**: atualiza `.agents/rules/sdd-reuso-e-duplicacao.md` declarando IDs canônicos da Constituição no frontmatter (`regras: [...]`) e descrevendo apenas procedimentos práticos, sem copiar frases literais da Constituição (`P-23`).
- **TST-02**: toda alteração em templates embutidos (`internal/scaffold/templates/`) atualiza os arquivos goldens (`internal/scaffold/testdata/golden/`) e passa nos testes de paridade e orçamento do pacote `internal/scaffold` e `internal/gate`.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Regra prática de reúso e anti-duplicação | `.agents/rules/sdd-reuso-e-duplicacao.md` | Guia prático cobrindo `DUP-01`, `DUP-02`, `DUP-05`, `DUP-10` e `GOV-04` | **Estender** | Incluir `DUP-03` no frontmatter e explicitar que a reutilização do código existente deve constar sempre na proposta apresentada ao usuário e nas decisões de concepção, planejamento, build e revisão |
| Instruções da estação de Concepção | `.agents/skills/sdd-concepcao/SKILL.md` e `references/{discovery,nova-feature,novo-fix}.md` | Roteiros de `/discovery`, `/nova-feature` e `/novo-fix` que exigem a seção `## O que já existe`, mas não obrigam explicitar o reúso na narrativa da proposta ao usuário | **Estender** | Tornar obrigatório explicitar na proposta (e no Gate 1) quais pacotes, funções e contratos existentes serão reutilizados/estendidos antes de propor código novo |
| Instruções da estação de Planejamento | `.agents/skills/sdd-planejamento/SKILL.md` e `references/split-features.md` | Roteiro de `/split-features` que gera `plan.md` e `feat-<id>-<nn>-*.md` | **Estender** | Exigir que cada task do plano e cada spec de task explicite os símbolos e arquivos existentes que serão reutilizados ou estendidos |
| Instruções da estação de Build | `.agents/skills/sdd-build/SKILL.md` e `references/{implementar,worker-job}.md` | Roteiro do Builder que cita `O que já existe` (`DUP-02`) | **Estender** | Explicitar a prioridade de estender/reutilizar os símbolos mapeados na proposta e a proibição de duplicar função, struct ou helper existente (`DUP-03`) |
| Instruções da estação de Governança e Operação | `.agents/skills/sdd-governanca/references/{revisar,pr-review}.md` e `.agents/skills/sdd-operacao/references/lider.md` | Roteiros de auditoria Camada B (`S-01..S-07`), revisão de PR e coordenação do Líder | **Estender** | Explicitar que o Reviewer/Governança reprova (`S-02` / `DUP-02` / `DUP-03`) propostas ou PRs que omitam o código existente reutilizado ou criem duplicatas |
| Adaptadores de agentes nas 3 ferramentas | `.claude/agents/*.md`, `.gemini/skills/*.chatmode.md`, `.github/chatmodes/*.chatmode.md` | Adaptadores finos (123 a 365 bytes) que apontam para `.agents/skills/sdd-*/` | **Estender** | Explicitar no resumo do papel a diretriz de reutilização do código existente na proposta/entrega/revisão, mantendo `<= 420` bytes (`P-24`) |
| Templates de scaffold e goldens | `internal/scaffold/templates/` e `internal/scaffold/testdata/golden/` | Espelhos `.tmpl` e goldens testados por `TestGoldenInit` e `TestCheckP24_RepositorioDoForgeSddPassaNoProprioPortao` | **Estender** | Reusar o motor `internal/scaffold` existente e sincronizar templates e goldens (`TST-02`, `DUP-10`) sem criar pacote novo |
| Validador determinístico do Portão (Camada A) | `internal/gate/gate.go` (`checkP03ReuseTable`, `checkP23RulesGovernance`, `checkP24ContextBudgets`) | Checks `P-03`, `P-23` e `P-24` já implementados e testados em `internal/gate/gate_test.go` | **Reusar** | Os checks existentes já validam a tabela `## O que já existe`, a governança das rules e os limites de bytes de skills/adaptadores |

## Dependências novas

Nenhuma (`SEG-07`, `ARQ-04`).

## Rules e IDs da Constituição Aplicáveis

- **Rules (`.agents/rules/`):**
  - `.agents/rules/sdd-reuso-e-duplicacao.md` (`DUP-01`, `DUP-02`, `DUP-03`, `DUP-05`, `DUP-10`, `GOV-04`)
  - `.agents/rules/sdd-fluxo-e-portao.md` (`FLX-01`, `FLX-02`, `FLX-03`, `FLX-04`, `FLX-06`, `GOV-02`)
  - `.agents/rules/sdd-codigo-e-testes.md` (`COD-02`, `COD-05`, `TST-01`, `TST-02`, `TST-03`, `TST-05`)
- **IDs da Constituição:** `ARQ-03`, `ARQ-04`, `ARQ-11`, `DUP-01`, `DUP-02`, `DUP-03`, `DUP-10`, `GOV-04`, `FLX-01`, `FLX-02`, `FLX-06`, `SEG-07`, `COD-02`, `COD-05`, `TST-01`, `TST-02`, `TST-03`, `TST-05`.

## Critérios de Aceitação Executáveis

```bash
set -o pipefail

# 1. Rule sdd-reuso-e-duplicacao.md referencia DUP-03 e explicita reutilização na proposta
grep -F "DUP-03" .agents/rules/sdd-reuso-e-duplicacao.md
grep -iE "proposta.*reutiliz|reutiliz.*proposta" .agents/rules/sdd-reuso-e-duplicacao.md

# 2. References de Concepção, Planejamento, Build, Governança e Operação exigem reutilização do código existente na proposta/tasks/revisão
grep -iE "proposta.*reutiliz|reutiliz.*código existente|reutiliz.*codigo existente" .agents/skills/sdd-concepcao/references/discovery.md
grep -iE "proposta.*reutiliz|reutiliz.*código existente|reutiliz.*codigo existente" .agents/skills/sdd-concepcao/references/nova-feature.md
grep -iE "proposta.*reutiliz|reutiliz.*código existente|reutiliz.*codigo existente" .agents/skills/sdd-concepcao/references/novo-fix.md
grep -iE "reutiliz" .agents/skills/sdd-planejamento/references/split-features.md
grep -iE "DUP-03|reutiliz" .agents/skills/sdd-build/SKILL.md
grep -iE "DUP-03|reutiliz" .agents/skills/sdd-governanca/references/revisar.md
grep -iE "DUP-03|reutiliz" .agents/skills/sdd-governanca/references/pr-review.md

# 3. Agentes do Forge SDD (.claude/agents, .gemini/skills, .github/chatmodes) explicitam a reutilização do código existente
grep -iE "reutiliz|reúso|reuso" .claude/agents/pm-architect.md
grep -iE "reutiliz|reúso|reuso" .claude/agents/planner.md
grep -iE "reutiliz|reúso|reuso" .claude/agents/builder.md
grep -iE "reutiliz|reúso|reuso" .claude/agents/reviewer.md
grep -iE "reutiliz|reúso|reuso" .claude/agents/governance.md
grep -iE "reutiliz|reúso|reuso" .claude/agents/lider.md
grep -iE "reutiliz|reúso|reuso" .gemini/skills/specifier.chatmode.md
grep -iE "reutiliz|reúso|reuso" .gemini/skills/planner.chatmode.md
grep -iE "reutiliz|reúso|reuso" .gemini/skills/builder.chatmode.md
grep -iE "reutiliz|reúso|reuso" .gemini/skills/reviewer.chatmode.md
grep -iE "reutiliz|reúso|reuso" .gemini/skills/governance.chatmode.md
grep -iE "reutiliz|reúso|reuso" .gemini/skills/orquestrador.chatmode.md
grep -iE "reutiliz|reúso|reuso" .github/chatmodes/specifier.chatmode.md
grep -iE "reutiliz|reúso|reuso" .github/chatmodes/planner.chatmode.md
grep -iE "reutiliz|reúso|reuso" .github/chatmodes/builder.chatmode.md
grep -iE "reutiliz|reúso|reuso" .github/chatmodes/reviewer.chatmode.md
grep -iE "reutiliz|reúso|reuso" .github/chatmodes/governance.chatmode.md
grep -iE "reutiliz|reúso|reuso" .github/chatmodes/orquestrador.chatmode.md

# 4. Paridade de templates, goldens (TST-02) e orçamentos P-23 / P-24
go test ./internal/gate/... -list 'TestCheckP03ReuseTable|TestCheckP23RulesGovernance|TestCheckP24'
go test ./internal/scaffold/... -list 'TestGoldenInit'
# negativo: P-03, P-23 e P-24 devem reprovar specs sem tabela de reúso, rules com ID inválido/cópia literal e arquivos acima do orçamento de bytes
go test ./internal/gate/... -run 'TestCheckP03ReuseTable|TestCheckP23RulesGovernance|TestCheckP24'
go test ./internal/scaffold/... -run 'TestGoldenInit'
go vet ./...
```
