# Fluxos Operacionais — forge-sdd (v3.0.0-beta)

## 1. Esteira Completa das 5 Estações + Portão em 3 Camadas (A / B / C)

```mermaid
sequenceDiagram
    autonumber
    actor H as Humano
    participant L as Estação 1: Líder
    participant PM as Estação 2: PM + Arquiteto
    participant PL as Estação 3: Planejador
    participant Q as CLI: forge-sdd queue
    participant B as Estação 4: Builder(s)
    participant GA as Camada A: forge-sdd gate
    participant GB as Camada B: Governança (Read-Only)
    participant SK as CLI: forge-sdd skills

    H->>L: Solicita Demanda / Discovery
    L->>PM: Executa Reuse Check (DUP-01) + Gera discovery.md e criteria.md
    PM-->>H: Portão FLX-01: Revisão Humana Obrigatória da Spec
    H->>L: Aprova Spec
    L->>PL: Solicita Plano Técnico
    PL-->>L: Gera plan.md (IDs Constitution + .agents/rules/ + Tarefas P/M/G)
    L->>Q: Enfileira subtarefas (sdd/.queue/queue.json)
    par Execução Paralela (itens sem conflito de arquivos)
        Q->>B: Builder 1 consome feat-XX-01
        Q->>B: Builder 2 consome feat-XX-02
    and Fila Dependente (depends_on)
        Q->>B: Builder aguarda dependências (feat-XX-03)
    end
    B->>GA: Roda Portão Determinístico (P-01, P-02, P-03, P-13, P-21, P-23)
    GA-->>L: PASS Determinístico
    L->>GB: Aciona Auditoria Semântica Read-Only (Diff + Spec + Rules)
    GB-->>L: Veredito JSON (conforme / ressalvas / bloqueado) + Auto-Skills candidatas
    L->>SK: Propõe/Sincroniza Auto-Skills (DUP-05) em .agents/, .claude/ e .gemini/
    L-->>H: Camada C: Painel Executivo (O que foi feito, status por etapa, o que falta)
```

---

## 2. Máquina de Estados da Fila do Builder (`sdd/.queue/queue.json`)

```mermaid
stateDiagram-v2
    [*] --> queued: forge-sdd queue enqueue
    queued --> claimed: forge-sdd queue claim (DependsOn = approved/done & sem conflito)
    claimed --> done: forge-sdd queue complete
    done --> approved: forge-sdd queue review --verdict approved (Camadas A + B OK)
    done --> rejected: forge-sdd queue review --verdict rejected
    rejected --> claimed: Retrabalho pelo Builder
    approved --> [*]
```

---

## 3. Ciclo de Vida de Auto-Skills (`DUP-05` / `GOV-03`)

```mermaid
flowchart LR
    E["3+ Ocorrências do Padrão<br/>(Regra de Três DUP-05)"] --> P["forge-sdd skills propose<br/>.agents/skills/<slug>/SKILL.md<br/>(status: candidate)"]
    P --> R["Aprovação no Fechamento de Onda<br/>(GOV-03)"]
    R --> A["forge-sdd skills approve<br/>(status: active + sdd/skills/index.md)"]
    A --> S["forge-sdd skills sync<br/>Propaga para .claude/skills/ e .gemini/skills/"]
```
