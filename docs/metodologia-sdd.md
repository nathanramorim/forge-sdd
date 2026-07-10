# Forge-SDD — Metodologia v1.7.1-beta.3

Documento substituto e consolidado da v1.0. Principais mudanças:

- Escopo do CLI: apenas `forge-sdd init`. Todo o restante (upgrade, doctor, archive, métricas) é acionado via chatmodes/prompts do GitHub Copilot.
- Orçamentos de tokens explícitos por artefato.
- Ciclo de sessão unificado (todos os chatmodes seguem o mesmo lifecycle).
- Telemetria com responsável definido (Orquestrador escreve via filesystem MCP).
- Sincronização automática da documentação wiki no repositório lp-forge-sdd via GitHub Actions.
- Skills index para descoberta sob demanda.
- Chatmodes ampliados: Archivist (manutenção) e Migrator (upgrade de versão).
- Protocolos de paralelismo e rollback definidos.
- Templates revisados (fechamento de blocos, regras duplicadas removidas).

**Versão:** 1.7.1-beta.3
**Última atualização:** 10 de julho de 2026

---

## 1. Diferenças v1.0 → v1.1

| Área | v1.0 | v1.1 |
|------|------|------|
| CLI | Implícito (init, upgrade, doctor) | Apenas init |
| Upgrade entre versões | `forge-sdd upgrade` | Prompt `/upgrade-sdd` (chatmode Migrator) |
| Health check | `forge-sdd doctor` | Prompt `/doctor` (chatmode Orquestrador) |
| Arquivamento de progress.md | Manual | Prompt `/archive` (chatmode Archivist) |
| Métricas | "endpoint local" vago | Schema fixo + escritor definido (Orquestrador) |
| Skills | Sem índice | `skills/index.md` obrigatório |
| Token budget | Texto solto | Tabela canônica (§5) |
| Paralelismo | Mencionado | Protocolo formal (§9) |
| Rollback | Ausente | Protocolo formal (§10) |

---

## 2. Princípios fundamentais

### Mantidos da v1.0

- **P1 — Hierarquia de leitura mínima.** Agente lê apenas o necessário por padrão; o resto é buscado sob demanda.
- **P2 — SDD particionado.** SDD dividido por domínio; agente carrega só o chunk relevante.
- **P3 — Estado vivo separado de histórico.** `progress.md` curto e ativo; `progress-log.md` arquiva.
- **P4 — Chatmodes minimalistas + Skills sob demanda.** Chatmode = papel + protocolo (≤500 tokens). Skills = conhecimento detalhado.
- **P5 — Critério de conclusão executável.** Toda feature tem comando que comprova término.
- **P6 — MCPs como memória externa.** Conhecimento mutável vive em MCPs.
- **P7 — Medir é parte da metodologia.** Cada sessão grava métricas.

### Adicionados em v1.1

- **P8 — CLI mínimo, runtime via agente.** O CLI só prepara o solo. Ciclo de vida é 100% conduzido por chatmodes
- **P9 — Escritor único por artefato.** Cada arquivo tem um único responsável.
- **P10 — Orçamento de tokens é contrato.** Exceder o budget dispara `/archive` ou particionamento.

---

## 3. Estrutura de arquivos canônica

```
/
├── .github/
│   ├── copilot-instructions.md
│   ├── chatmodes/
│   │   ├── orquestrador.chatmode.md
│   │   ├── specifier.chatmode.md
│   │   ├── builder.chatmode.md
│   │   ├── revisor.chatmode.md
│   │   ├── archivist.chatmode.md   # NOVO
│   │   └── migrator.chatmode.md    # NOVO
│   └── prompts/
│       ├── proxima-feature.prompt.md
│       ├── nova-feature.prompt.md
│       ├── revisar.prompt.md
│       ├── status.prompt.md
│       ├── archive.prompt.md       # NOVO
│       ├── doctor.prompt.md        # NOVO
│       └── upgrade-sdd.prompt.md   # NOVO
│
├── .vscode/
│   └── mcp.json
│
└── sdd/
    ├── .sdd-version
    ├── .sddrc
    ├── README.md
    ├── HOWTO.md
    ├── memory/
    │   ├── constitution.md
    │   ├── progress.md
    │   ├── progress-log.md
    │   └── mcps.md
    ├── spec/
    │   ├── overview.md
    │   ├── stack.md
    │   ├── modules.md
    │   ├── flows.md
    │   └── decisions.md
    ├── plan.md
    ├── features/
    │   ├── index.md
    │   ├── feat-00-foundation.md
    │   ├── feat-XX-*.md
    │   └── feat-XX-nome-feature/    # Subpasta opcional para agrupamento de features/tasks
    │       ├── feat-XX-YY-*.md      # Subfeatures aninhadas
    │       └── task-YY-*.md         # Tasks aninhadas
    ├── skills/
    │   ├── index.md                 # NOVO — obrigatório
    │   └── *.md
    └── .metrics/
        ├── schema.json              # NOVO — contrato de telemetria
        └── session-*.json
```

---

## 4. Responsabilidade única por artefato

| Artefato | Único escritor | Leitores |
|----------|----------------|----------|
| `constitution.md` | humano | todos |
| `progress.md` | Orquestrador | todos |
| `progress-log.md` | Archivist | nenhum (auditoria humana) |
| `mcps.md` | humano (init) + Migrator | todos |
| `spec/*.md` | Specifier + humano | Builder, Revisor |
| `plan.md` | humano | Orquestrador, Specifier |
| `features/index.md` | Specifier + Orquestrador | todos |
| `features/feat-XX.md` (ou aninhados) | Specifier (cria), Builder (tasks), Revisor (status final) | todos |
| `skills/*.md` | humano | Builder |
| `skills/index.md` | humano + Specifier | Builder |
| `.metrics/session-*.json` | Orquestrador | benchmark externo |
| `.sdd-version` | CLI (init) + Migrator | Orquestrador |

**Regra:** se dois agentes editam o mesmo arquivo na mesma sessão, falhou o protocolo. Reabrir o ciclo.

---

## 5. Orçamento de tokens (contrato)

| Arquivo | Limite duro | Ação ao exceder |
|---------|-------------|-----------------|
| `copilot-instructions.md` | 1.5 KB | Mover detalhe para `skills/` |
| `constitution.md` | 2 KB | Recusar nova regra; revisar conjunto |
| `progress.md` | 1 KB | `/archive` automático |
| `chatmode.md` (cada) | 500 tokens (~2 KB) | Mover detalhe para `skills/` |
| `spec/<arquivo>.md` | 2 KB | Particionar |
| `feat-XX.md` | 2 KB | Quebrar em duas features |
| `skills/<arquivo>.md` | sem limite | — |

O Orquestrador valida `progress.md` ao final de toda sessão. Se exceder, dispara Archivist antes de encerrar.

---

## 6. Lifecycle unificado de sessão

Todo chatmode (exceto Migrator) segue:

```
READ-MIN
  Sempre: progress.md
  Se relevante: feat ativa, constitution

PLAN
  Reportar entendimento + intenção
  Aguardar confirmação humana (exceto se prompt explícito)

ACT
  Executar dentro do escopo do papel
  Consultar MCPs sob demanda (context7, git)

WRITE
  Editar apenas arquivos do seu escopo (§4)

CLOSE (apenas Orquestrador)
  Atualizar progress.md
  Validar budgets (§5)
  Gravar .metrics/session-<timestamp>.json
  Disparar /archive se progress.md > 1 KB
```

Chatmodes não-Orquestrador devolvem controle ao Orquestrador ao terminar; não fecham a sessão.

---

## 7. Chatmodes

### 7.1 Orquestrador

Lê estado, decide, delega, fecha sessão. Único que escreve `progress.md` e `.metrics/`. Nunca implementa.

### 7.2 Specifier

Cria `feat-XX-*.md` e atualiza `index.md`. Pode editar `spec/*` quando uma feature exige novo módulo/decisão.

**Bloqueios — NÃO criar se:**
- Critério de conclusão for subjetivo
- Não couber em uma sessão
- Conflitar com a constitution

### 7.3 Builder

Implementa código. Lê `feat-XX.md` + `spec/*` necessário. Consulta context7 antes de qualquer lib externa. Marca tasks `[x]`.

**Quando o Builder DEVE usar context7:**
1. Ao instalar/usar nova dependência
2. Ao implementar integração com API externa
3. Quando a constitution exige versão específica
4. Em caso de erro de tipo/assinatura inesperado

### 7.4 Revisor

Roda critério executável. Valida aderência à constitution. Atualiza Status da feature. Aprova/reprova.

### 7.5 Archivist *(NOVO)*

Aciona-se via `/archive` ou automaticamente quando `progress.md` > 1 KB.

- Move entradas antigas (>5 sessões) para `progress-log.md`
- Compacta a seção "Última sessão" em uma linha
- Não toca em features ou specs

### 7.6 Migrator

Aciona-se via `/upgrade-sdd <versão>`.

- **Environment Detection:** Identifica se o binário está em cache do NPX para sugerir limpeza de cache (`rm -rf ~/.cache/forge-sdd`) ou reinstalação via `npx @nathanramorim/forge-sdd@latest`.
- Lê `.sdd-version`
- Aplica diff entre versões da metodologia (templates, novos arquivos, renames)
- Atualiza `.sdd-version`
- Não modifica conteúdo de domínio (apenas estrutura)

---

## 8. Prompts oficiais

| Prompt | Modo | Aciona | Uso |
|--------|------|--------|-----|
| `/proxima-feature` | agent | Orquestrador | Continuar trabalho |
| `/nova-feature` | agent | Specifier | Criar feat a partir de descrição |
| `/revisar` | agent | Revisor | Validar feature ativa |
| `/status` | ask | Orquestrador (read-only) | Diagnóstico sem ação |
| `/archive` | agent | Archivist | Compactar progress.md |
| `/doctor` | ask | Orquestrador | Health check (budgets, refs, MCPs) |
| `/upgrade-sdd` | agent | Migrator | Migrar versão |
| `/discovery` | agent | Specifier | Discovery (Produto + Eng Senior) |
| `/constitution` | agent | Specifier | Architectural Alignment (Codebase Scan) |

---

## 9. Protocolo de Guardrails (Segurança e Qualidade)

Toda sessão deve respeitar as travas automáticas por fase:

| Fase | Guardrail | Responsável | Ação em caso de falha |
|------|-----------|-------------|----------------------|
| **Plan** | Budget de `progress.md` (≤ 1 KB) | Orquestrador | Bloquear e exigir `/archive` |
| **Act** | Validação Local (`go vet` / linter) | Builder | Corrigir antes de reportar conclusão |
| **Review** | Escopo de Arquivos (Gerados vs Modificados) | Revisor | Rejeitar se houver drift de escopo |
| **Close** | Registro de Telemetria (`.metrics/`) | Orquestrador | Gravar antes de encerrar sessão |

---

## 10. Protocolo de Handoff (Transição de Contexto)

O Handoff é obrigatório ao final de cada fase para garantir que a inteligência gerada seja aproveitada pelo próximo agente:

| Transição | Conteúdo do Handoff | Destino |
|-----------|----------------------|---------|
| **Discovery → Spec** | Resumo do produto, arquivos gerados em `discovery/` e riscos. | `/nova-feature` |
| **Spec → Act** | Nome da branch, feature ID e tarefas prioritárias. | `/proxima-feature` |
| **Act → Review** | Lista de arquivos alterados e pontos de atenção para testes. | `/revisar` |
| **Sessão → Sessão** | Resumo do estado atual e bloqueios. | `progress.md` (Handoff Context) |

---

## 11. Protocolo de paralelismo

Quando `index.md` declara janela de paralelismo (duas features sem dependência mútua):

1. Orquestrador identifica par compatível (mesma fase, sem deps cruzadas).
2. Cria branches `feat/A` e `feat/B` a partir do mesmo commit base.
3. Cada feature é executada em sessões separadas (uma por feature). Nunca duas no mesmo turno.
4. `progress.md` lista ambas como `in-progress` com branch distinta.
5. Merge segue ordem de aprovação do Revisor.
6. Conflito de merge → Orquestrador delega ao Builder da feature mais recente.

---

## 10. Protocolo de rollback

Quando o Revisor reprovar e Builder não conseguir corrigir em ≤2 turnos:

1. Orquestrador marca a feature como `blocked` em `progress.md`.
2. Reverte commits da branch (`git reset --hard <base>`), mantém branch viva.
3. Registra incidente em `progress-log.md` (via Archivist).
4. Sugere ao humano: redividir feature, revisar critério ou reescrever do zero.
5. Não tenta novamente sem confirmação humana.

---

## 11. Telemetria (schema fixo)

Arquivo `.metrics/schema.json` — contrato versionado:

```json
{
  "$schema": "forge-sdd/metrics/1.0",
  "feature": "string",
  "phase": "integer",
  "agent_path": ["orquestrador", "builder", "revisor"],
  "tokens_input": "integer",
  "tokens_output": "integer",
  "turns": "integer",
  "duration_seconds": "integer",
  "criterio_atendido": "boolean",
  "model": "string",
  "rework_lines": "integer",
  "context7_calls": "integer",
  "git_calls": "integer",
  "files_touched": ["string"],
  "outcome": "approved | rejected | blocked",
  "sdd_version": "string"
}
```

**Escritor:** Orquestrador, na fase CLOSE.
**Como:** filesystem MCP (já nativo). Nome: `session-<ISO8601>.json`.

Telemetria desabilitada (`.sddrc.telemetry.enabled = false`) → fase CLOSE pula a escrita.

*Nota sobre Gravação de Métricas:*
1. **Identificação com Subpastas:** O campo `"feature"` deve conter o caminho relativo completo da especificação ou subtask correspondente (ex: `sdd/features/feat-1234-auth/task-01.md`).
2. **Sessões Incompletas/Timeout:** As métricas devem ser gravadas obrigatoriamente a cada fim de sessão (fase `Close`), mesmo que a sessão seja interrompida, cancelada ou sem progresso (`criterio_atendido: false`), anotando `outcome: blocked` ou `outcome: rejected` para rastreamento de esforço e timeouts.

### Hipóteses verificáveis (benchmark)

- **H1:** Forge-SDD reduz tokens em ≥30% vs readme-only em tasks ≥4h
- **H2:** Forge-SDD reduz turnos até conclusão em ≥40% vs vibe coding
- **H3:** Forge-SDD aumenta aderência ao critério em ≥25%
- **H4:** Forge-SDD reduz drift de contexto em ≥60%

## 12. Visualização Arquitetural (C4 Model)

O SDD utiliza o framework **C4 Model** representado em **Mermaid** para documentar arquiteturas:

1. **Nível 1 (Contexto):** Flowcharts (`graph TB`) para interações macro.
2. **Nível 2 (Container):** Flowcharts para detalhar aplicações e persistência.
3. **Nível 3 (Componente):** Sequence Diagrams (`sequenceDiagram`) para fluxos de dados e interações temporais.
4. **Nível 4 (Código):** Markdown estruturado ou diagramas ER simplificados.

---

## 13. Skills index

`sdd/skills/index.md` é obrigatório. Builder consulta-o antes de carregar qualquer skill.

Exemplo:

```markdown
# Skills Index

| Skill | Quando usar | Tamanho |
|-------|-------------|---------|
| python-conventions.md | Python ≥3.11, type hints, ruff | 3 KB |
| stripe-integration.md | Pagamentos, webhooks Stripe | 5 KB |
| pgvector-queries.md | Busca semântica Postgres | 4 KB |
```

**Sem index, Builder não carrega skill (regra dura).**

---

## 13. CLI — escopo definitivo

Único comando: `forge-sdd init [opções]`.

### 13.1 O que `init` faz

1. Pergunta interativamente (ou via flags):
   - Nome do projeto
   - Stack principal (runtime, db, config, secrets)
   - Versão do Forge-SDD a usar (default: latest)
   - Telemetria (sim/não)
   - Idioma dos templates (pt-BR/en)
2. Cria toda a árvore da §3.
3. Preenche templates com valores informados.
4. Inicializa `.sdd-version`, `.sddrc`, `.metrics/schema.json`.
5. Não roda `git init` (responsabilidade do humano).
6. Não instala dependências de runtime do projeto.
7. Imprime próximos passos: abrir VS Code, instalar Copilot, rodar `/proxima-feature`.

### 13.2 Flags propostas

```
forge-sdd init
  --stack node|python|go|rust|other
  --db    postgres|sqlite|mongo|none
  --telemetry on|off
  --lang  pt-BR|en
  --version 1.7.1-beta.3
  --yes       # não-interativo, usa defaults
  --dry-run   # mostra árvore sem criar
```

### 13.3 O que `init` NÃO faz

| Operação | Onde fica |
|----------|-----------|
| Upgrade de versão | `/upgrade-sdd` (Migrator) |
| Health check | `/doctor` (Orquestrador) |
| Adicionar feature | `/nova-feature` (Specifier) |
| Compactar progress | `/archive` (Archivist) |
| Validar estrutura | `/doctor` |
| Reset/wipe | manual (humano) |

**Justificativa:** manter o CLI minúsculo, sem dependência de runtime, sem manutenção de features que evoluem rápido. Toda evolução vive em chatmodes versionados junto com a metodologia.

### 13.4 Distribuição sugerida

- Binário único (Go ou Rust), ou
- `npx forge-sdd init` (Node), ou
- `pipx run forge-sdd init` (Python)

**Recomendação:** Go — binário estático, sem runtime, cross-platform. Templates embutidos via `embed.FS`.

---

## 14. Templates revisados

### 14.1 `feat-XX.md`

```markdown
# feat/<nome>

**Branch:** `feat/<nome>`
**Fase:** N
**Depende de:** —
**Status:** `todo`

## Objetivo

[Uma frase]

## Critério de conclusão

```bash
<comando executável>
```

## Tarefas

- [ ] **XX-1**
- [ ] **XX-2**

## Arquivos gerados

```
arquivo1
arquivo2
```

## Skills relevantes

(consultar `skills/index.md`)
```

### 14.2 `constitution.md`

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

## Regras (máx. 10)

1. Sem commits diretos em main
2. Branch por feature
3. Config centralizado em <arquivo>
4. Secrets em .env (nunca commit)
5. Antes de usar lib externa, consultar context7 com versão exata
6. Toda feature tem critério executável
7. <regra específica>
```

### 14.3 `progress.md`

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

### 14.4 `copilot-instructions.md`

```markdown
# Copilot Instructions — <projeto>

## Contexto

<2 linhas>

## Lifecycle (todo agente)

1. **READ-MIN:** ler `sdd/memory/progress.md`
2. **PLAN:** reportar intenção, aguardar confirmação
3. **ACT:** executar no escopo do papel
4. **WRITE:** editar apenas arquivos do escopo
5. **CLOSE** (Orquestrador): atualizar progress, métricas, archive se necessário

## Arquivos críticos

- `sdd/memory/progress.md` — estado ativo
- `sdd/memory/constitution.md` — regras imutáveis
- `sdd/features/feat-XX-*.md` — tarefa atual

## MCPs

- **context7** — obrigatório antes de lib externa
- **git** — status antes de iniciar/encerrar feature

## Orçamentos

- `progress.md` ≤ 1 KB → exceder dispara `/archive`
- `chatmode` ≤ 500 tokens → detalhe vai para `skills/`
```

### 14.5 `archivist.chatmode.md` *(NOVO)*

```markdown
---
description: "Compacta progress.md e move histórico para progress-log.md"
tools: [read, edit]
mcps: []
---

Você é o Archivist. Sua única responsabilidade é manter `progress.md` ≤ 1 KB.

## Antes

1. Leia `sdd/memory/progress.md`
2. Leia `sdd/memory/progress-log.md` (últimas 3 entradas)

## O que fazer

1. Identifique entradas em "Última sessão" / histórico com >5 sessões
2. Mova-as para `progress-log.md` (topo, com data)
3. Compacte "Última sessão" em uma linha resumo
4. Mantenha intactas: Status, Features ativas, Próximo passo, Bloqueios

## Bloqueios

- Não modifique features
- Não modifique specs
- Não escreva métricas

## Ao finalizar

Reporte ao Orquestrador: tamanho final de `progress.md`.
```

### 14.6 `migrator.chatmode.md` *(NOVO)*

```markdown
---
description: "Migra a estrutura SDD de uma versão para outra"
tools: [read, edit, search]
mcps: [git]
---

Você é o Migrator. Aplica diffs estruturais entre versões do Forge-SDD.

## Antes

1. Leia `sdd/.sdd-version`
2. Leia o diff oficial da versão alvo
3. Confirme com o usuário a versão alvo

## O que fazer

1. Aplique renomes, novos arquivos, novos chatmodes/prompts
2. Atualize `.sdd-version`
3. NÃO altere conteúdo de domínio (constitution, features, specs)
4. Em caso de ambiguidade, pergunte

## Ao finalizar

1. Liste arquivos criados/movidos/renomeados
2. Devolva controle ao Orquestrador
3. Sugira rodar `/doctor`
```

### 14.7 `mcp.json`

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

---

## 15. Fluxo de desenvolvimento

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
Se progress.md > 1 KB → dispara [Archivist]
   ↓
git commit + push + PR
```

---

## 16. Checklist de novo projeto (pós-init)

```
[ ] forge-sdd init executado sem erros
[ ] sdd/.sdd-version == versão esperada
[ ] sdd/memory/constitution.md preenchido (não placeholder)
[ ] sdd/spec/overview.md + stack.md preenchidos
[ ] sdd/features/feat-00-foundation.md com critério executável
[ ] sdd/skills/index.md existe (mesmo vazio)
[ ] .github/chatmodes/ tem 6 arquivos
[ ] .github/prompts/ tem 7 arquivos
[ ] .vscode/mcp.json com context7 + git
[ ] git init + commit inicial feito (manual)
[ ] /doctor reporta verde
[ ] /proxima-feature inicia foundation corretamente
```

---

## 17. Anti-padrões

| Anti-padrão | Problema | Solução |
|-------------|----------|---------|
| `progress.md` gigante | Lê arquivo grande toda sessão | Mover histórico para `progress-log.md` |
| SDD monolítico | Carrega tudo sempre | Particionar em `sdd/spec/` |
| `constitution` com 30 regras | Agente ignora por excesso | Máximo 10 regras |
| Chatmode com 2000 tokens | Carrega contexto pesado | Mover detalhe para `skills/` |
| Features sem critério executável | Agente nunca sabe se terminou | Adicionar comando de validação |
| Builder sem context7 | Usa API antiga | Tornar consulta obrigatória na constitution |
| Branches sem PR | Difícil rastrear | `feat/*` → PR → main |
| Config hardcoded | Agente não sabe onde ajustar | Centralizar em arquivo de config |
| Sem telemetria | Não dá para otimizar | Habilitar `.metrics/` desde dia 1 |
| Templates poluídos com comentários | Tokens desperdiçados | Comentários em `HOWTO.md` |
| CLI inflado com update, add-feature | Manutenção dupla (CLI + chatmode) | Manter CLI = `init` apenas |
| Builder escrevendo em `progress.md` | Quebra escritor único | Builder reporta; Orquestrador escreve |
| Skills sem index | Builder carrega tudo ou nada | `skills/index.md` obrigatório |
| Métricas em formato livre | Impossível benchmarkar | Schema fixo (§11) |
| Migrator alterando constitution | Mistura estrutura com domínio | Migrator só toca estrutura |
| Sessão paralela na mesma thread | Estado inconsistente | Uma feature por sessão (§9) |
| Reprovação infinita do Revisor | Loop de retrabalho | Rollback em ≤2 turnos (§10) |

---

## 18. Roadmap para implementar o CLI

Ordem sugerida de features do próprio `forge-sdd`:

| Feature | Descrição |
|---------|-----------|
| `feat-00-foundation` | repo Go, layout `cmd/forge-sdd` |
| `feat-01-templates-embed` | `embed.FS` dos templates v1.1 |
| `feat-02-init-interactive` | survey + render |
| `feat-03-init-flags` | modo não-interativo |
| `feat-04-dry-run` | preview da árvore |
| `feat-05-versioning` | escrita de `.sdd-version` |
| `feat-06-self-test` | golden test (init + diff vs fixture) |
| `feat-07-release` | goreleaser, binários multi-OS |

Cada uma com critério executável (ex.: `forge-sdd init demo --yes && diff -r demo/ tests/fixtures/expected/`).

---

## 19. Versionamento

`sdd/.sdd-version = 1.7.1-beta.3` para projetos criados com este documento.

Mudanças seguem semver:

- **MAJOR:** muda estrutura de pastas ou contratos de chatmode
- **MINOR:** adiciona artefato sem quebrar existentes
- **PATCH:** ajusta templates, melhora textos

### Migração 1.0 → 1.1 (executada por Migrator)

1. Criar `archivist.chatmode.md`, `migrator.chatmode.md`
2. Criar `archive.prompt.md`, `doctor.prompt.md`, `upgrade-sdd.prompt.md`
3. Criar `skills/index.md` (vazio com cabeçalho)
4. Criar `.metrics/schema.json`
5. Atualizar `copilot-instructions.md` com seção Lifecycle
6. Atualizar `.sdd-version` para `1.7.1-beta.3`

Nenhum conteúdo de domínio é tocado.

### Releases Beta vs Estáveis

Para garantir a estabilidade do fluxo principal:
- **Lançamentos Beta (`-beta`):** A tag correspondente (ex: `v1.7.1-beta.3`) é criada e enviada diretamente a partir de sua branch de feature. O Pull Request (PR) correspondente deve ser mantido **aberto** no GitHub para testes e validações em ambiente real por um período. Opcionalmente, múltiplos fixes e features podem ser agrupados e acumulados na mesma branch e versão beta antes de realizar um novo bump ou tag.
- **Lançamentos Estáveis/Oficiais:** Apenas após a consolidação dos testes no canal beta, o PR da feature é mergeado na branch `main` e a tag estável (ex: `v1.7.1`) é publicada como oficial.

---

## 20. Resumo executivo das decisões-chave

1. **CLI = só `init`.** Tudo o mais (upgrade, doctor, archive) virou prompt + chatmode. Reduz superfície de manutenção e garante que a metodologia funcione mesmo sem o CLI instalado.
2. **2 chatmodes novos:** Archivist (mantém `progress.md` enxuto) e Migrator (upgrade de versão).
3. **3 prompts novos:** `/archive`, `/doctor`, `/upgrade-sdd`.
4. **Escritor único por arquivo (§4)** — elimina conflitos.
5. **Lifecycle unificado (§6)** — todo chatmode segue READ-MIN → PLAN → ACT → WRITE → CLOSE.
6. **Schema fixo de métricas (§11)** — habilita benchmark real.
7. **Skills index obrigatório (§12)** — descoberta sob demanda real.
8. **Protocolos de paralelismo (§9) e rollback (§10)** — preenchem buracos da v1.0.
9. **Stack sugerida do CLI:** Go + `embed.FS`, distribuído como binário estático.

---

**Forge-SDD v1.7.1-beta.3** — Última atualização: 10 de julho de 2026
