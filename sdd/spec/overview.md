# Overview — forge-sdd (v3.0.0-beta)

CLI Go e framework de **Spec-Driven Development (SDD)** multi-agente com **Orquestração Independente por 5 Estações**, **Constituição Distribuída por IDs estáveis**, **Portão de Governança em 3 Camadas (A/B/C)** e **Motor de Auto-Skills**.

## Índice da Especificação Arquitetural (`sdd/spec/`)
- `overview.md` — visão C4 (Contexto e Containers) e comparativo `v2.6.x` → `v3.0.0-beta`
- `modules.md` — pacotes Go (`cmd/` e `internal/`), templates embarcados e responsabilidades
- `flows.md` — fluxos operacionais (5 estações, fila determinística, Portão A/B/C e Auto-Skills)
- `concepts.md` — glossário canônico do domínio (`DUP-01` / `GOV-04`)
- `stack.md` — tecnologias, versões e layout do projeto
- `decisions.md` & `adr/` — registro de decisões arquiteturais (ADRs)

---

## Evolução Arquitetural: `v2.6.x` vs `v3.0.0-beta`

```mermaid
flowchart TB
    subgraph V2["Antes (v2.6.x) — Sessão Sequencial e Constitution Monolítica"]
        direction LR
        U2["Humano"] --> A2["Agente Único (troca de chapéu na mesma sessão)"]
        A2 --> C2["sdd/memory/constitution.md<br/>(Monolítica: regras 0..13 em prosa)"]
        A2 --> S2["Discovery → Plan → Implementação → /revisar"]
    end

    subgraph V3["Agora (v3.0.0-beta) — 5 Estações Independentes + Fila + Portão A/B/C + Auto-Skills"]
        direction TB
        H["👤 Humano"] <-->|"1. Aprova Spec (FLX-01)<br/>2. Recebe Painel Executivo (Camada C)"| L["🎯 Estação 1: Líder (Orquestrador)<br/>.agents/roles/lider.md"]

        L -->|"Aciona Concepção"| PM["📐 Estação 2: PM + Arquiteto<br/>.agents/roles/pm-architect.md<br/>(discovery + criteria + Reuse Check DUP-01)"]
        PM -->|"Portão Humano Obrigatório (FLX-01)"| L

        L -->|"Spec aprovada"| PL["🗺️ Estação 3: Planejador (Plan)<br/>.agents/roles/planner.md<br/>(Mapeia IDs Constitution + .agents/rules/sdd-*.md)"]
        PL -->|"Quebra em P/M/G + depends_on"| Q[("🗂️ Fila Determinística<br/>sdd/.queue/queue.json<br/>(forge-sdd queue)")]

        Q -->|"Onda Paralela (sem conflito)"| B1["🔨 Estação 4: Builder A<br/>.agents/roles/builder.md"]
        Q -->|"Onda Paralela / Fila"| B2["🔨 Estação 4: Builder B<br/>.agents/roles/builder.md"]

        B1 & B2 -->|"Entrega concluída"| GA["⚙️ Portão Camada A (Determinístico)<br/>forge-sdd gate (P-01..P-23)"]
        GA -->|"Se Camada A = PASS"| GB["🛡️ Estação 5: Analista de Governança<br/>.agents/roles/governance.md<br/>(Camada B — Estritamente Read-Only)"]
        GB -->|"Veredito JSON + Candidatas Auto-Skill"| L
        L -->|"Regra de Três (DUP-05 / GOV-03)"| SK["♻️ Motor Auto-Skills (forge-sdd skills)<br/>.agents/skills/ → .claude/skills/ & .gemini/skills/"]
    end
```

---

## Constituição Distribuída (`GOV-04`)

```mermaid
flowchart LR
    CI["sdd/memory/constitution.md<br/>Enxuta: 'O quê e Por quê'<br/>IDs estáveis: ARQ-*, SEG-*, DUP-*, COD-*, TST-*, FLX-*, GOV-*"]
    CI --> R1[".agents/rules/sdd-arquitetura.md<br/>(ARQ-01..05)"]
    CI --> R2[".agents/rules/sdd-seguranca.md<br/>(SEG-01..04)"]
    CI --> R3[".agents/rules/sdd-reuso-e-duplicacao.md<br/>(DUP-01..05)"]
    CI --> R4[".agents/rules/sdd-codigo-e-testes.md<br/>(COD-01..05, TST-01..04)"]
    CI --> R5[".agents/rules/sdd-fluxo-e-portao.md<br/>(FLX-01..05, GOV-01..05)"]
```

---

## Escopo do Binário (`forge-sdd`)
- `forge-sdd init` — inicializa a estrutura Forge-SDD em qualquer repositório
- `forge-sdd update` — atualiza artefatos de agentes preservando dados do usuário (`sdd/memory/`, `sdd/features/`, `sdd/discovery/`)
- `forge-sdd doctor` — verifica integridade dos artefatos SDD e versão instalada
- `forge-sdd status` / `forge-sdd report` — exibe progresso, métricas e telemetria de sessões
- `forge-sdd queue` — gerencia a fila determinística de tarefas do Builder (`sdd/.queue/queue.json`)
- `forge-sdd gate` — executa o Portão Determinístico Camada A (`P-01`, `P-02`, `P-03`, `P-13`, `P-21`, `P-23`)
- `forge-sdd skills` — propõe, aprova, lista e sincroniza Auto-Skills (`DUP-05` / `GOV-03`)
- `forge-sdd governance` — importa e audita padrões genéricos de governança sem acoplamento a cliente
