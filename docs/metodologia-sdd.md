# Metodologia SDD — Guia de Replicação

Guia completo para replicar esta estrutura de planejamento orientado a agentes em qualquer projeto de software.

---

## O que é esta metodologia

Um sistema de documentação viva que combina três elementos:

1. **SDD (Software Design Document)** — especificação técnica do sistema
2. **Memória de projeto** — estado atual, regras imutáveis e histórico
3. **Agentes Copilot** — especialistas que executam partes do projeto com contexto correto

O resultado é um fluxo de desenvolvimento onde o Copilot consegue retomar qualquer sessão, entender onde está, e executar a próxima tarefa sem perda de contexto.

---

## Estrutura de arquivos

```
sdd/
├── README.md                  ← Guia de navegação do SDD
├── sdd-<projeto>.md           ← Especificação completa (a "bíblia")
├── plan-<projeto>.md          ← Roadmap faseado com tarefas numeradas
├── memory/
│   ├── constitution.md        ← Regras que nunca mudam
│   └── progress.md            ← Estado vivo do projeto (atualizado sempre)
└── features/
    ├── index.md               ← Tabela de todas as features
    ├── feat-00-foundation.md
    ├── feat-01-nome.md
    └── ...

.github/
├── copilot-instructions.md    ← Instruções globais para o Copilot
├── agents/
│   ├── orquestrador.agent.md
│   ├── builder-X.agent.md
│   └── revisor.agent.md
└── prompts/
    ├── proxima-feature.prompt.md
    └── ...
```

---

## Os 4 artefatos essenciais

### 1. `sdd/memory/constitution.md` — A Lei

O que nunca muda. Escreva aqui **antes de qualquer código**.

```markdown
## Missão
[Uma frase descrevendo o que o sistema faz]

## Stack (não alterar sem revisão)
| Camada | Escolha | Motivo |
...

## Decisões de arquitetura resolvidas
| Decisão | Resolução |
...

## Regras de desenvolvimento
1. Sem commits diretos em main
2. Cada feature em sua própria branch
3. Config centralizado em arquivo (nunca hardcode)
4. Prompts em arquivos separados
5. [suas regras específicas]
```

**Critério:** se alguém novo ler este arquivo, entende o que pode e não pode fazer.

---

### 2. `sdd/memory/progress.md` — O Mapa Vivo

Atualizado ao **fim de cada sessão ou feature**. É o primeiro arquivo que o agente lê.

```markdown
## Status geral
```
Fase 0 — Foundation   [x] done
Fase 1 — Core         [ ] todo
```

## Features
| Feature | Branch | Status | Notas |
|---------|--------|--------|-------|
| feat-00 | feat/foundation | done | — |
| feat-01 | feat/core       | todo | — |

## Próximo passo
**Iniciar por:** `feat/core`
...

## Histórico de sessões
### YYYY-MM-DD — O que foi feito
- item 1
- item 2
- **Próximo:** feat-XX
```

**Critério:** ao abrir uma nova sessão, lendo só este arquivo você sabe exatamente o que fazer.

---

### 3. `sdd/features/feat-XX-nome.md` — A Tarefa

Um arquivo por feature. Estrutura padrão:

```markdown
# feat/nome-da-feature

**Branch:** `feat/nome`
**Fase do plano:** N
**Depende de:** `feat/outra` (mergeada)
**Bloqueada por:** —
**Status:** `todo` | `in-progress` | `done`

## Objetivo
[Uma frase do que esta feature entrega]

## Tarefas
- [ ] **F0-1** Descrição da tarefa
  - Detalhe ou código de exemplo

## Critério de conclusão
```bash
# comando que comprova que funciona
python test_X.py  # → saída esperada
```

## Arquivos gerados por esta feature
```
arquivo1.py
arquivo2.py
```
```

**Critério:** ao ler este arquivo, o agente sabe exatamente o que criar, como testar e quando parar.

---

### 4. `sdd/features/index.md` — O Mapa de Features

```markdown
## Dependency graph
```
main
 └─ feat/foundation
      ├─ feat/core-a  (deps: foundation)
      └─ feat/core-b  (deps: foundation)
          feat/pipeline  (deps: core-a + core-b)
```

## Janelas de trabalho paralelo
| Janela | Branches simultâneas |
|--------|----------------------|
| Após foundation merge | core-a + core-b |

## Índice de features
| # | Arquivo | Branch | Fase | Status |
|---|---------|--------|------|--------|
| 00 | feat-00.md | feat/foundation | 0 | done |
```

---

## Os agentes Copilot

### Estrutura mínima de um agente (`.github/agents/nome.agent.md`)

```markdown
---
description: "Quando invocar este agente (triggers em linguagem natural)"
name: "Nome do Agente"
tools: [read, edit, search, execute]
user-invocable: true   # false = só invocado por outros agentes
---

Você é [papel]. Sua responsabilidade é [objetivo].

## Antes de começar
1. Leia [arquivo de contexto]
2. Leia [regras]

## O que fazer
[instruções específicas com exemplos de código se necessário]

## Ao finalizar
1. Atualize progress.md
2. Marque tasks em feat-XX.md
3. Atualize index.md
```

### Tipos de agente por responsabilidade

| Tipo | Função | Quantidade |
|------|--------|-----------|
| **Orquestrador** | Lê o estado, decide o que fazer, delega | 1 |
| **Builder** | Implementa uma fase/domínio específico | 1 por fase |
| **Revisor** | Valida o código gerado contra o SDD | 1 |
| **Gerador** | Usa o sistema pronto para produzir outputs | 1 por tipo de output |

### Orquestrador — protocolo obrigatório

```markdown
## Protocolo de início de sessão
1. Leia progress.md → identifique próxima feature
2. Leia constitution.md → relembre regras
3. Leia feat-XX.md → entenda as tarefas
4. Reporte ao usuário e aguarde confirmação
5. Delegue para o builder correto
6. Após conclusão, invoque o revisor
7. Atualize progress.md, feat-XX.md, index.md
```

### Prompt de entrada rápida (`.github/prompts/proxima-feature.prompt.md`)

```markdown
---
description: "Inicia sessão de desenvolvimento"
agent: "Orquestrador"
---
Inicie a sessão: leia progress.md, identifique a próxima feature,
reporte o status e aguarde confirmação.
```

---

## O fluxo de desenvolvimento

```
Sessão nova
    ↓
[Orquestrador] lê progress.md
    ↓
Identifica próxima feature (status: todo)
    ↓
Lê feat-XX.md + constitution.md
    ↓
Reporta tarefas → aguarda confirmação
    ↓
Delega para [Builder correspondente]
    ↓
Builder implementa + testa
    ↓
[Revisor] valida contra SDD
    ↓
Orquestrador atualiza progress.md + feat-XX.md + index.md
    ↓
git commit + push + PR
```

---

## Como criar o SDD de um novo projeto

### Passo 1 — Defina a missão (15 min)

Responda em uma frase:
> "Este sistema [faz o quê] para [quem] usando [como]."

### Passo 2 — Escolha a stack (30 min)

Documente em `constitution.md`:
- Runtime / linguagem
- Banco de dados (ou ausência)
- Frameworks principais
- Como a config será gerenciada
- Onde ficam segredos (`.env`)

### Passo 3 — Mapeie as fases (1h)

Decomponha o sistema em 4–8 fases sequenciais. Cada fase:
- Entrega algo **testável e funcional**
- Tem dependências claras das fases anteriores
- Pode ser paralelizada onde possível

Exemplo de fases para qualquer projeto:
```
Fase 0 — Foundation      (scaffolding, config, state)
Fase 1 — Core Domain     (a lógica central do sistema)
Fase 2 — Integrações     (APIs externas, banco, storage)
Fase 3 — Interface       (CLI, API HTTP, UI)
Fase 4 — Observabilidade (logs, métricas, retry)
Fase 5 — Robustez        (edge cases, rate limits, falhas)
```

### Passo 4 — Crie os arquivos de feature

Para cada tarefa significativa, crie `sdd/features/feat-XX-nome.md` com:
- Objetivo em uma frase
- Lista de tarefas atômicas (`- [ ]`)
- Critério de conclusão executável
- Lista de arquivos que serão criados

### Passo 5 — Crie os agentes

Mínimo viável:
1. `orquestrador.agent.md` — lê progress.md, roteia para builders
2. `builder-foundation.agent.md` — implementa feat-00
3. `builder-core.agent.md` — implementa feats do domínio principal
4. `revisor.agent.md` — valida contra SDD

### Passo 6 — Configure o Copilot

Em `.github/copilot-instructions.md`:

```markdown
## Contexto do projeto
[descrição em 2-3 linhas]

## Antes de começar qualquer tarefa
1. Leia sdd/memory/progress.md
2. Leia o feat-XX-*.md da feature atual

## Ao finalizar
1. Atualize sdd/memory/progress.md
2. Marque tasks em sdd/features/feat-XX-*.md
3. Atualize sdd/features/index.md

## Regras obrigatórias
[suas regras específicas]
```

---

## Regras de ouro da metodologia

### 1. Progress.md é a fonte de verdade
Nunca encerre uma sessão sem atualizar. É o que permite retomar sem perda de contexto.

### 2. Constitution.md é imutável
Só mude com justificativa explícita. Se você está mudando constantemente, as decisões não estavam maduras.

### 3. Critério de conclusão é executável
Todo feat-XX.md precisa ter um comando que você pode rodar agora e verificar se passou. Se o critério for subjetivo, o agente vai travar.

### 4. Uma feature = um entregável testável
Não agrupe features por conveniência. Se não dá para testar isoladamente, divida mais.

### 5. Agentes têm contexto mínimo necessário
Um builder não precisa conhecer outros builders. Dê a ele apenas o que precisa: o feat-XX.md, a constitution.md e os arquivos que vai modificar.

### 6. O orquestrador nunca implementa
Só lê, decide e delega. Se o orquestrador começa a criar arquivos, a separação de responsabilidades quebrou.

---

## Checklist para novo projeto

```
[ ] Criou sdd/memory/constitution.md com missão + stack + regras
[ ] Criou sdd/memory/progress.md com tabela de features e "Próximo passo"
[ ] Criou sdd/features/index.md com grafo de dependências
[ ] Criou sdd/features/feat-00-foundation.md com critério executável
[ ] Criou .github/copilot-instructions.md com as 3 seções obrigatórias
[ ] Criou orquestrador.agent.md com protocolo de sessão
[ ] Criou pelo menos um builder.agent.md
[ ] Criou revisor.agent.md
[ ] Testou: abriu sessão nova e o orquestrador soube o que fazer
```

---

## Anti-padrões a evitar

| Anti-padrão | Problema | Solução |
|-------------|----------|---------|
| progress.md desatualizado | Agente não sabe onde está | Atualizar **sempre** ao encerrar sessão |
| Features sem critério executável | Agente nunca sabe se terminou | Adicionar comando de validação em cada feat-XX.md |
| Constitution.md com 50 regras | Agente ignora por excesso | Máximo 10 regras, apenas as que realmente mudariam o código |
| Agente faz tudo | Perde especialização e contexto | Um builder por domínio/fase |
| Branches sem PR | Difícil rastrear o que foi mergeado | Sempre feat/* → PR → main |
| Config hardcoded no código | Agente não sabe onde ajustar | Centralizar em config.toml ou equivalente |

---

## Template de constitution.md

```markdown
# Constituição — [nome-do-projeto]

Documento imutável. Atualizar apenas com justificativa explícita.

## Missão
[Uma frase]

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
| [decisão] | [resolução] |

## Regras de desenvolvimento
1. Sem commits diretos em main
2. [regra específica do projeto]
3. Config centralizado em [arquivo]
4. Secrets em .env (nunca commitar)
```

## Template de feat-XX.md

```markdown
# feat/nome

**Branch:** `feat/nome`
**Fase:** N
**Depende de:** `feat/outra` (mergeada)
**Status:** `todo`

## Objetivo
[Uma frase]

## Tarefas
- [ ] **X-1** Tarefa 1
- [ ] **X-2** Tarefa 2

## Critério de conclusão
```bash
python test_X.py  # → saída esperada
```

## Arquivos gerados
```
arquivo1.py
arquivo2.py
```
```

---

**Última atualização:** 30 de maio de 2026
