# Forge-SDD — Metodologia

Guia completo para replicar uma estrutura de planejamento orientado a agentes em qualquer projeto de software, otimizada para economia de tokens, validação por benchmark e uso com GitHub Copilot (extensível para Cursor e outros editores).

**Versão:** 1.0.0
**Última atualização:** 30 de maio de 2026

---

## O que é Forge-SDD

Um sistema de documentação viva que combina cinco elementos:

1. **SDD particionado (Software Design Document)** — especificação técnica fragmentada por domínio para leitura sob demanda
2. **Memória de projeto** — estado ativo curto, histórico arquivado e regras imutáveis
3. **Agentes Copilot (chatmodes)** — especialistas com contexto mínimo necessário
4. **MCPs integrados** — capacidades externas dos agentes (docs atualizadas, git, etc.)
5. **Telemetria + Benchmark** — medição contínua de eficiência (tokens, turnos, aderência)

O resultado é um fluxo de desenvolvimento onde o Copilot retoma qualquer sessão, entende onde está e executa a próxima tarefa **com o mínimo de tokens possível**, sem perda de contexto.

---

## Princípios fundamentais

### P1 — Hierarquia de leitura mínima
Agente lê apenas o necessário por padrão. Tudo o mais é buscado sob demanda.

```
Sempre lê:         progress.md (≤1KB)
Lê se relevante:   constitution.md + feat-XX atual
Sob demanda:       spec/*, skills/*, outras features
Nunca por padrão:  progress-log.md, HOWTO.md, .metrics/
```

### P2 — SDD particionado
Em vez de um arquivo monolítico, o SDD é dividido por domínio. O agente carrega só o chunk relevante.

### P3 — Estado vivo separado de histórico
`progress.md` é curto e ativo. `progress-log.md` guarda o histórico arquivado.

### P4 — Chatmodes minimalistas + Skills sob demanda
Chatmode = papel + protocolo (curto, ≤500 tokens). Skills = conhecimento detalhado, carregado quando necessário.

### P5 — Critério de conclusão executável
Toda feature tem um comando que comprova que terminou. Sem comando, não é uma feature válida.

### P6 — MCPs como memória externa
Conhecimento que muda (docs de libs, estado do repo) vive em MCPs, não no contexto do modelo.

### P7 — Medir é parte da metodologia
Cada sessão grava métricas. Sem dados, não há otimização real.

---

## Estrutura de arquivos

```
<projeto>/
├── .github/
│   ├── copilot-instructions.md          # ≤1.5KB, instruções globais
│   ├── chatmodes/                       # agentes (formato Copilot)
│   │   ├── orquestrador.chatmode.md
│   │   ├── specifier.chatmode.md
│   │   ├── builder.chatmode.md
│   │   └── revisor.chatmode.md
│   └── prompts/                         # prompts oficiais
│       ├── proxima-feature.prompt.md
│       ├── nova-feature.prompt.md
│       ├── revisar.prompt.md
│       └── status.prompt.md
│
├── .vscode/
│   └── mcp.json                         # declaração dos MCPs
│
└── sdd/
   ├── .sdd-version                     # versão da metodologia
   ├── .sddrc                           # config + telemetria
   ├── README.md                        # navegação curta
   ├── HOWTO.md                         # tutoriais (não lido por agentes)
   ├── memory/
   │   ├── constitution.md              # imutável (≤2KB)
   │   ├── progress.md                  # estado ATIVO (≤1KB)
   │   ├── progress-log.md              # histórico arquivado
   │   └── mcps.md                      # catálogo de MCPs
   ├── spec/                            # SDD particionado
   │   ├── overview.md
   │   ├── stack.md
   │   ├── modules.md
   │   ├── flows.md
   │   └── decisions.md
   ├── plan.md                          # roadmap faseado
   ├── features/
   │   ├── index.md                     # tabela de features
   │   ├── feat-00-foundation.md
   │   └── feat-XX-*.md
   ├── skills/                          # conhecimento sob demanda
   │   └── *.md
   └── .metrics/                        # telemetria local
       └── session-*.json
```

---

## Os 6 artefatos essenciais

### 1. `sdd/memory/constitution.md` — A Lei

Documento imutável. Escreva antes de qualquer código.

```markdown
# Constituição — <projeto>

## Missão
[Uma frase descrevendo o que o sistema faz]

## Stack
| Camada | Escolha | Motivo |
|--------|---------|--------|
| Runtime | | |
| DB | | |
| Config | | |
| Secrets | | |

## Decisões de arquitetura resolvidas
| Decisão | Resolução |
|---------|-----------|

## Regras de desenvolvimento
1. Sem commits diretos em main
2. Cada feature em sua própria branch
3. Config centralizado em arquivo (nunca hardcode)
4. Prompts em arquivos separados
5. Secrets em .env (nunca commitar)
6. Antes de usar lib externa, consultar context7 com a versão exata
7. [regras específicas do projeto]
```

**Regra de tamanho:** ≤2KB. Máximo 10 regras.
**Critério:** se alguém novo lê este arquivo, entende o que pode e não pode fazer.

---

### 2. `sdd/memory/progress.md` — O Mapa Vivo

Atualizado ao fim de cada sessão. É o primeiro arquivo lido pelo agente.

```markdown
# Progress — <projeto>

## Status
```
Fase 0 — Foundation   [x] done
Fase 1 — Core         [ ] in-progress
Fase 2 — Integrações  [ ] todo
```

## Features ativas
| Feature | Branch | Status |
|---------|--------|--------|
| feat-01-core | feat/core | in-progress |
| feat-02-api  | feat/api  | todo |

## Próximo passo
**Iniciar:** continuar `feat-01-core` (task 01-3)
**Bloqueios:** nenhum

## Última sessão
- 2026-05-30 — concluído feat-00, iniciado feat-01 até task 01-2

> Histórico completo em `progress-log.md`
```

**Regra de tamanho:** ≤1KB. Tudo acima de 5 sessões antigas vai para `progress-log.md`.
**Critério:** abrir nova sessão, ler só este arquivo, saber exatamente o que fazer.

---

### 3. `sdd/memory/progress-log.md` — O Arquivo Morto

Histórico completo. Não é lido por padrão. Existe para auditoria humana.

```markdown
# Progress Log — <projeto>

## 2026-05-30 — Foundation completa
- Criada estrutura base
- Smoke test passando
- Próximo: feat-01-core

## 2026-05-29 — Setup inicial
- ...
```

---

### 4. `sdd/features/feat-XX-nome.md` — A Tarefa

Um arquivo por feature. Estrutura padrão:

```markdown
# feat/nome

**Branch:** `feat/nome`
**Fase:** N
**Depende de:** `feat/outra` (mergeada)
**Status:** `todo` | `in-progress` | `done`

## Objetivo
[Uma frase do que esta feature entrega]

## Critério de conclusão
```bash
# Comando executável que prova que está pronto
python test_X.py  # → saída esperada
```

## Tarefas
- [ ] **XX-1** Descrição da tarefa
- [ ] **XX-2** Descrição da tarefa

## Arquivos gerados
```
arquivo1.py
arquivo2.py
```

## Skills relevantes
- `skills/python-conventions.md`
```

**Regra de ordem:** o critério vem ANTES das tarefas. Se já está atendido, encerra sem implementar.
**Critério:** o agente sabe exatamente o que criar, como testar, quando parar.

---

### 5. `sdd/features/index.md` — O Mapa de Features

```markdown
# Index de Features

## Dependency graph
```
main
└─ feat/foundation
     ├─ feat/core-a
     └─ feat/core-b
         feat/pipeline (deps: core-a + core-b)
```

## Janelas de paralelismo
| Após | Pode rodar em paralelo |
|------|------------------------|
| foundation | core-a + core-b |

## Índice
| # | Arquivo | Branch | Fase | Status |
|---|---------|--------|------|--------|
| 00 | feat-00-foundation.md | feat/foundation | 0 | done |
| 01 | feat-01-core.md | feat/core | 1 | in-progress |
```

---

### 6. `sdd/spec/` — SDD Particionado

Em vez de um único `sdd-<projeto>.md`, o SDD é dividido em arquivos pequenos e focados:

#### `spec/overview.md`
```markdown
# Overview

[2-3 linhas descrevendo o sistema]

## Índice
- `stack.md` — tecnologias
- `modules.md` — componentes
- `flows.md` — fluxos principais
- `decisions.md` — decisões de design
```

#### `spec/stack.md`
Tecnologias, versões, assets existentes.

#### `spec/modules.md`
Cada módulo com responsabilidades em `[ ]` / `[x]`.

#### `spec/flows.md`
Pseudocódigo dos fluxos principais.

#### `spec/decisions.md`
Decisões resolvidas com `[x]` e abertas com `[ ]`.

**Por que particionar:** o agente carrega só o chunk relevante. Implementando um módulo? Lê `modules.md`, ignora os outros.

---

## Os agentes (chatmodes)

### Estrutura mínima

```markdown
---
description: "Quando invocar este agente"
tools: [read, edit, search, execute]
mcps: [context7, git]
---

Você é [papel]. Sua responsabilidade é [objetivo].

## Antes de começar
1. [arquivo a ler]
2. [arquivo a ler]

## O que fazer
[instruções específicas]

## Ao finalizar
1. Atualize progress.md (apenas seção alterada)
2. Marque tasks em feat-XX.md
3. Atualize index.md
```

**Regra de tamanho:** ≤500 tokens por chatmode. Detalhes vão para `skills/`.

### Os 4 chatmodes essenciais

| Chatmode | Função | User-invocable |
|----------|--------|----------------|
| **Orquestrador** | Lê estado, decide, delega. Nunca implementa. | sim |
| **Specifier** | Cria novas features (feat-XX) a partir de descrição | sim |
| **Builder** | Implementa uma feature por vez | sim |
| **Revisor** | Valida código contra SDD e critério de conclusão | sim |

### Orquestrador — protocolo obrigatório

```markdown
## Protocolo de início de sessão
1. Leia progress.md → identifique próxima feature
2. Se em dúvida, leia constitution.md
3. Leia feat-XX.md indicado
4. Reporte status ao usuário e aguarde confirmação
5. Delegue para Builder (não implemente)
6. Após conclusão, invoque Revisor
7. Atualize progress.md, feat-XX.md, index.md
8. Grave métricas em .metrics/session-<timestamp>.json
```

### Specifier — protocolo de criação de feature

```markdown
## Antes de criar
1. Leia constitution.md (regras e stack)
2. Leia features/index.md (próximo número, dependências)
3. Leia spec/overview.md (escopo do sistema)

## Bloqueios — NÃO criar se:
- Critério de conclusão for subjetivo
- Não couber em uma sessão
- Conflitar com a constitution

## Ao finalizar
1. Criar feat-XX-<nome>.md
2. Adicionar em index.md
3. Sugerir ao Orquestrador iniciar (não iniciar sozinho)
```

### Builder — protocolo de implementação

```markdown
## Antes de implementar
1. Leia o feat-XX.md alvo
2. Leia o critério de conclusão PRIMEIRO
3. Se já atendido, encerre sem implementar
4. Carregue spec/modules.md SE precisar de detalhe arquitetural
5. Para libs externas, consulte context7 com a versão da constitution

## Durante
- Implemente apenas o necessário para o critério passar
- Não modifique arquivos fora da lista declarada na feat
- Use skills/* sob demanda

## Ao finalizar
1. Rode o critério de conclusão
2. Marque tasks como [x]
3. Reporte ao Orquestrador
```

### Revisor — protocolo de validação

```markdown
## Validação
1. Rodar critério de conclusão (deve passar)
2. Conferir aderência à constitution
3. Conferir se modificou apenas arquivos declarados
4. Levantar issues por gravidade (bloqueante / aviso / sugestão)

## Ao finalizar
- Aprovar (passa para merge) OU
- Reprovar com lista de correções específicas
```

---

## Os prompts oficiais

`.github/prompts/*.prompt.md` — invocações rápidas que o usuário usa via `/comando`.

### `proxima-feature.prompt.md`
```markdown
---
description: "Inicia sessão na próxima feature pendente"
mode: agent
---
Leia progress.md, identifique a próxima feature, reporte e aguarde confirmação.
```

### `nova-feature.prompt.md`
```markdown
---
description: "Cria nova feature a partir de descrição"
mode: agent
---
Invoque o Specifier para criar feat-XX a partir desta descrição: ${input}
```

### `revisar.prompt.md`
```markdown
---
description: "Revisa a feature em andamento"
mode: agent
---
Invoque o Revisor sobre a feature ativa em progress.md.
```

### `status.prompt.md`
```markdown
---
description: "Mostra status atual sem implementar nada"
mode: ask
---
Leia progress.md e reporte: features ativas, próximo passo, bloqueios.
```

---

## MCPs integrados

Forge-SDD recomenda os seguintes MCPs como padrão:

### `.vscode/mcp.json`
```json
{
 "servers": {
   "context7": {
     "command": "npx",
     "args": ["-y", "@upstash/context7-mcp"]
   },
   "git": {
     "command": "uvx",
     "args": ["mcp-server-git", "--repository", "."]
   }
 }
}
```

### Stack de MCPs

| MCP | Status | Função |
|-----|--------|--------|
| **context7** | obrigatório | Docs de libs/frameworks atualizadas |
| **git** | obrigatório | Estado do repositório |
| **github** | opcional | Issues, PRs, branches remotos |
| **memory** | opcional | Memória persistente entre sessões |
| **filesystem** | opcional | Já nativo no Copilot |

### Quando o Builder DEVE usar context7
1. Ao instalar/usar nova dependência
2. Ao implementar integração com API externa
3. Quando a constitution exige versão específica
4. Em caso de erro de tipo/assinatura inesperado

### `sdd/memory/mcps.md`
```markdown
# MCPs do projeto

| MCP | Status | Usado por |
|-----|--------|-----------|
| context7 | ativo | Builder, Specifier |
| git | ativo | Orquestrador, Revisor |
```

---

## Telemetria e Benchmark

### Telemetria local (opt-in)

`sdd/.sddrc`:
```json
{
 "version": "1.0.0",
 "telemetry": {
   "enabled": true,
   "anonymous": true,
   "endpoint": "local"
 }
}
```

Cada sessão grava `sdd/.metrics/session-<timestamp>.json`:
```json
{
 "feature": "feat-03-stripe",
 "tokens_input": 12450,
 "tokens_output": 3210,
 "turns": 4,
 "duration_seconds": 380,
 "criterio_atendido": true,
 "model": "claude-sonnet-4",
 "rework_lines": 12,
 "context7_calls": 2
}
```

### Métricas capturadas

| Métrica | Significado |
|---------|-------------|
| `tokens_input` | Total enviado ao modelo |
| `tokens_output` | Total gerado pelo modelo |
| `turns` | Mensagens trocadas até conclusão |
| `duration_seconds` | Tempo wall-clock |
| `criterio_atendido` | Se o critério executável passou |
| `rework_lines` | Linhas removidas/reescritas |
| `context7_calls` | Quantas consultas a docs externas |

### Benchmark — estrutura padrão

```
benchmarks/
├── tasks/
│   ├── task-01-cli-tool/
│   ├── task-02-rest-api/
│   ├── task-03-data-pipeline/
│   ├── task-04-multi-agent/
│   └── task-05-frontend-component/
├── baselines/
│   ├── no-methodology/
│   ├── plain-readme/
│   ├── spec-kit/
│   └── kiro/
├── runs/
│   └── YYYY-MM-DD-<task>-<method>/
│       ├── metrics.json
│       ├── transcript.md
│       └── final-code/
└── results/
   └── report.md
```

### Protocolo de execução do benchmark

1. Mesmo modelo em todos os runs
2. Mesma temperatura, mesmo seed quando possível
3. Mesma task description palavra por palavra
4. Sem intervenção humana além de responder perguntas do agente
5. Stop: critério passa OU 50 turnos OU $5 gastos
6. Gravar tudo: input, output, tool calls, tempo

### Hipóteses verificáveis

- **H1:** Forge-SDD reduz tokens em ≥30% vs readme-only em tasks ≥4h
- **H2:** Forge-SDD reduz turnos até conclusão em ≥40% vs vibe coding
- **H3:** Forge-SDD aumenta aderência ao critério em ≥25%
- **H4:** Forge-SDD reduz drift de contexto em ≥60%

---

## O fluxo de desenvolvimento

```
Sessão nova
   ↓
[Orquestrador] lê progress.md (apenas)
   ↓
Identifica próxima feature (status: todo)
   ↓
Lê feat-XX.md + constitution.md (se necessário)
   ↓
Reporta tarefas → aguarda confirmação
   ↓
Delega para [Builder]
   ↓
Builder lê critério → consulta context7 se preciso → implementa → testa
   ↓
[Revisor] valida contra SDD + critério
   ↓
Orquestrador atualiza progress.md + feat-XX + index
   ↓
Grava .metrics/session-*.json
   ↓
git commit + push + PR
```

---

## Como criar um novo projeto Forge-SDD

### Passo 1 — Defina a missão (15 min)

> "Este sistema [faz o quê] para [quem] usando [como]."

### Passo 2 — Escreva a constitution (30 min)

- Stack
- Como config é gerenciado
- Onde ficam segredos
- 6-10 regras imutáveis (incluindo regra de context7)

### Passo 3 — Particione o SDD (1h)

Crie `sdd/spec/` com 5 arquivos. Não escreva tudo agora — escreva o overview e o módulo principal. O resto cresce com as fases.

### Passo 4 — Mapeie as fases (1h)

Decomponha em 4-8 fases sequenciais. Cada fase entrega algo testável.

```
Fase 0 — Foundation
Fase 1 — Core Domain
Fase 2 — Integrações
Fase 3 — Interface
Fase 4 — Observabilidade
Fase 5 — Robustez
```

### Passo 5 — Crie feat-00-foundation

Sempre comece pela foundation. Critério executável obrigatório.

### Passo 6 — Configure agentes e MCPs

- 4 chatmodes em `.github/chatmodes/`
- 4 prompts em `.github/prompts/`
- `mcp.json` com context7 + git
- `copilot-instructions.md` com 3 seções obrigatórias

### Passo 7 — Teste

Abra sessão nova. Invoque `/proxima-feature`. O Orquestrador deve saber exatamente o que fazer.

---

## Regras de ouro

### 1. progress.md é fonte de verdade — e curto
Nunca encerre sessão sem atualizar. Mantenha ≤1KB.

### 2. constitution.md é imutável
Só muda com justificativa explícita.

### 3. Critério de conclusão é executável
Sem comando executável, não é uma feature válida.

### 4. Uma feature = um entregável testável
Não agrupe por conveniência.

### 5. Agentes têm contexto mínimo
Builder não conhece outros builders. Recebe só o que precisa.

### 6. Orquestrador nunca implementa
Lê, decide, delega. Se cria arquivos, a separação quebrou.

### 7. Builder consulta context7 antes de usar lib externa
Não confia em conhecimento prévio do modelo.

### 8. Toda sessão grava métricas
Sem dados, não há otimização.

### 9. Particione cedo
Se um arquivo passar de 2KB, divida.

### 10. Skills sob demanda
Conhecimento detalhado não fica em chatmode.

---

## Checklist para novo projeto

```
[ ] sdd/.sdd-version criado com "1.0.0"
[ ] sdd/.sddrc com telemetria configurada
[ ] sdd/memory/constitution.md (≤2KB)
[ ] sdd/memory/progress.md (≤1KB) com "Próximo passo"
[ ] sdd/memory/progress-log.md (vazio inicialmente)
[ ] sdd/memory/mcps.md
[ ] sdd/spec/overview.md + stack.md
[ ] sdd/features/index.md
[ ] sdd/features/feat-00-foundation.md com critério executável
[ ] sdd/plan.md
[ ] .github/copilot-instructions.md (≤1.5KB)
[ ] .github/chatmodes/{orquestrador, specifier, builder, revisor}.chatmode.md
[ ] .github/prompts/{proxima-feature, nova-feature, revisar, status}.prompt.md
[ ] .vscode/mcp.json com context7 + git
[ ] Testou: sessão nova + /proxima-feature funciona
```

---

## Anti-padrões a evitar

| Anti-padrão | Problema | Solução |
|-------------|----------|---------|
| progress.md gigante | Lê arquivo grande toda sessão | Mover histórico para progress-log.md |
| SDD monolítico | Carrega tudo sempre | Particionar em sdd/spec/ |
| constitution com 30 regras | Agente ignora por excesso | Máximo 10 regras |
| Chatmode com 2000 tokens | Carrega contexto pesado | Mover detalhe para skills/ |
| Features sem critério executável | Agente nunca sabe se terminou | Adicionar comando de validação |
| Builder sem context7 | Usa API antiga | Tornar consulta obrigatória na constitution |
| Branches sem PR | Difícil rastrear | feat/* → PR → main |
| Config hardcoded | Agente não sabe onde ajustar | Centralizar em arquivo de config |
| Sem telemetria | Não dá para otimizar | Habilitar .metrics/ desde dia 1 |
| Templates poluídos com comentários | Tokens desperdiçados | Comentários em HOWTO.md |

---

## Templates

### Template `constitution.md`
```markdown
# Constituição — <projeto>

## Missão
[Uma frase]

## Stack
| Camada | Escolha | Motivo |
|--------|---------|--------|
| Runtime | | |
| DB | | |
| Config | | |
| Secrets | | |

## Decisões resolvidas
| Decisão | Resolução |
|---------|-----------|

## Regras
1. Sem commits diretos em main
2. Branch por feature
3. Config centralizado em [arquivo]
4. Secrets em .env
5. Antes de usar lib externa, consultar context7 com a versão exata
6. [regra específica]
```

### Template `progress.md`
```markdown
# Progress — <projeto>

## Status
```
Fase 0 — Foundation   [ ] todo
```

## Features ativas
| Feature | Branch | Status |
|---------|--------|--------|
| feat-00-foundation | feat/foundation | todo |

## Próximo passo
**Iniciar:** feat-00-foundation
**Bloqueios:** —

## Última sessão
- Projeto criado via Forge-SDD CLI
```

### Template `feat-XX.md`
```markdown
# feat/nome

**Branch:** `feat/nome`
**Fase:** N
**Depende de:** —
**Status:** `todo`

## Objetivo
[Uma frase]

## Critério de conclusão
```bash
[comando executável]
```

## Tarefas
- [ ] **XX-1** Tarefa 1
- [ ] **XX-2** Tarefa 2

## Arquivos gerados
```
arquivo1
arquivo2
### Template `copilot-instructions.md`
```markdown
# Copilot Instructions — <projeto>

## Contexto
[2-3 linhas]

## Antes de qualquer tarefa
1. Leia `sdd/memory/progress.md`
2. Para a feature ativa, leia `sdd/features/feat-XX-*.md`
3. Em dúvidas de regra, leia `sdd/memory/constitution.md`

## Ao finalizar
1. Atualize `sdd/memory/progress.md` (apenas seção alterada)
2. Marque tasks em `sdd/features/feat-XX-*.md`
3. Atualize `sdd/features/index.md`
4. Grave métricas em `sdd/.metrics/session-<timestamp>.json`

## MCPs
- **context7** — obrigatório antes de usar lib externa
- **git** — consultar status antes de iniciar feature

## Regras
[copiar regras da constitution]
```

---

## Versionamento da metodologia

Cada projeto tem `sdd/.sdd-version` indicando a versão do Forge-SDD usada.

Mudanças seguem semver:
- **MAJOR**: muda estrutura de pastas ou contratos de chatmode
- **MINOR**: adiciona artefato sem quebrar existentes
- **PATCH**: ajusta templates, melhora textos

O CLI oferece `forge-sdd upgrade` para migrar entre versões.

---

**Forge-SDD v1.0.0** — Última atualização: 30 de maio de 2026