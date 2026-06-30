# Arquitetura, Fluxo e Metodologia Estruturada do Forge-SDD

O **Forge-SDD** é um framework operacional para times de desenvolvimento e de produto focado no ciclo contínuo de **Discovery → Spec → Test → Code**. Ele organiza a interação com Inteligências Artificiais e agentes autônomos por meio de fases bem definidas com artefatos persistidos no próprio repositório, mitigando a perda de contexto inerente a chats convencionais de LLM.

---

## 1. Domínios Estruturais (Mapeamento de Pastas e Arquivos)

O ecossistema do Forge-SDD organiza o repositório em domínios específicos para garantir a separação entre descobertas de negócio, especificações técnicas, regras de arquitetura e código de produção.

```
/
├── .github/ ou .gemini/ ou .claude/  # Domínio de Instruções e Prompts dos Agentes
│   ├── chatmodes/                  # (Copilot/Gemini) Definições de papéis de agentes (Orquestrador, etc.)
│   ├── prompts/                    # (Copilot/Gemini) Gatilhos estruturados de prompt (/status, etc.)
│   └── commands/                   # (Claude Code em .claude/) Prompts específicos do Claude Code
│
├── CLAUDE.md ou GEMINI.md           # Instruções de bootstrapping (para Claude ou Gemini na raiz)
│
├── .vscode/                        # Integração com VS Code (opcional se Copilot)
│   └── mcp.json                    # Declaração local de servidores MCP
│
├── docs/                           # Domínio de Documentação Pública e Guias
│   ├── introducao.md               # Onboarding de novos membros do time
│   ├── comandos.md                 # Referência rápida de terminal e chat
│   └── arquitetura-e-fluxo-forge-sdd.md # Este documento de arquitetura e conceitos
│
└── sdd/                            # Domínio do Core do Framework (SDD)
    ├── .sdd-version                # Rastreamento de versão do framework
    ├── .sddrc                      # Configurações locais (Ex: habilitar telemetria)
    ├── README.md                   # Resumo rápido do SDD local
    ├── HOWTO.md                    # Instruções de operação para a IA
    │
    ├── memory/                     # Memória Persistente do Projeto
    │   ├── constitution.md         # A "lei universal" do projeto (Stack e Regras)
    │   ├── progress.md             # Estado ativo do projeto (Curto, máx 1 KB)
    │   ├── progress-log.md         # Histórico arquivado das sessões antigas
    │   └── mcps.md                 # Configurações de Model Context Protocol ativas
    │
    ├── spec/                       # Especificações e Arquitetura de Alto Nível
    │   ├── overview.md             # Visão geral de negócio e glossário
    │   ├── stack.md                # Tecnologias, linguagens e bibliotecas oficiais
    │   ├── modules.md              # Mapeamento de módulos e fronteiras de contexto
    │   ├── flows.md                # Fluxos críticos do sistema
    │   └── decisions.md            # Índice central de Architecture Decision Records (ADRs)
    │
    ├── features/                   # Backlog de Desenvolvimento por Feature
    │   ├── index.md                # Índice e grafo de dependência de features
    │   └── feat-XX-*.md            # Especificações individuais de funcionalidade
    │
    └── skills/                     # Habilidades e Guias Técnicos Específicos
        ├── index.md                # Índice e critérios de carregamento das skills
        └── *.md                    # Regras técnicas profundas (Ex: stripe-integration.md)
```

---

## 2. Divisão de Papéis e Responsabilidades

O ciclo de vida do Forge-SDD é gerenciado por agentes virtuais que assumem papéis complementares ao longo da sessão. Cada agente possui um escopo restrito de leitura e escrita para evitar conflitos de merge e desvios de escopo.

| Papel | Responsabilidade Principal | Ações Frequentes | Arquivos Sob sua Escrita |
|---|---|---|---|
| **Orquestrador** | Lê o estado do projeto, planeja a próxima ação, delega tarefas aos demais agentes, grava métricas e encerra a sessão. | `/status`, `/doctor`, `/archive` | `sdd/memory/progress.md`, `.metrics/session-*.json` |
| **Specifier** | Conduz debates de produto e engenharia (discovery), gera especificações estruturadas e detalha tarefas técnicas. | `/discovery`, `/nova-feature`, `/split-features` | `sdd/features/feat-XX-*.md`, `sdd/features/index.md`, `sdd/discovery/*` |
| **Builder** | Codifica a solução com base nas especificações técnicas fornecidas. Aplica o ciclo TDD de forma estrita. | `/proxima-feature` | Código de produção, testes unitários, `sdd/features/feat-XX-*.md` (tasks) |
| **Revisor** | Executa validações automatizadas e auditorias de conformidade com a constituição, regras OWASP e qualidade de código. | `/revisar` | `sdd/features/feat-XX-*.md` (status final) |
| **Archivist** | Mantém a leveza do arquivo de progresso ativo, compactando o histórico da sessão e movendo dados antigos. | `/archive` | `sdd/memory/progress.md`, `sdd/memory/progress-log.md` |
| **Migrator** | Conduz a atualização de versão e layout da estrutura do framework no projeto sem afetar os dados de domínio. | `/upgrade-sdd` | `.sdd-version`, prompts e templates estruturais do framework |

---

## 3. Fluxo de Etapas (Pipeline Operacional)

Uma nova funcionalidade passa por um pipeline operacional sequencial com **Portões de Qualidade (Gates)**, garantindo que o código só comece a ser escrito após as especificações estarem estruturadas e aprovadas pelo time.

```
[Problema ou Ideia]
       │
       ▼
1. Execução do comando `/discovery`
       │
       ▼
2. Análise e documentação (discovery/*.md)
       │
       ├─► SE "Descartar / Diferir" ──► Fim do Processo
       │
       └─► SE "Construir" ──► [GATE 1: Aprovação de Produto]
                                   │
                                   ▼
                      3. Execução de `/nova-feature`
                                   │
                                   ▼
                      Geração de Story & Spec BDD (feat-XX-*.md)
                                   │
                                   ▼
                             [GATE 2: Spec Aprovada]
                                   │
                                   ▼
                      4. Ciclo TDD (Comando `/proxima-feature`)
                           ├── RED (Escrever teste baseado na spec)
                           ├── GREEN (Implementar lógica simples - KISS)
                           └── REFACTOR (Melhorar código - SOLID + DRY)
                                   │
                                   ▼
                      5. Auditoria de Código (Comando `/revisar`)
                           ├── Validação das especificações
                           └── Checagem de segurança (OWASP Top 10)
                                   │
                                   ▼
                      6. Fechamento da Sessão (Orquestrador)
                           ├── Commit e PR Automático
                           ├── Escrita de Telemetria
                           └── Execução de `/archive` se progress.md > 1 KB
```

---

## 4. Diagramas Conceituais e Técnicos

Os diagramas abaixo ilustram o comportamento, ciclo de vida e a arquitetura operacional do framework.

### Diagrama 1: Modo Tradicional vs. Forge-SDD
Demonstra o contraste entre o desenvolvimento assistido por IA convencional e o fluxo estruturado do Forge-SDD.

```mermaid
graph TD
    subgraph Tradicional["Modo Tradicional (Vibe Coding)"]
        T1["Ideia Vaga no Chat de IA"] --> T2["IA gera código direto sem spec"]
        T2 --> T3["Desenvolvedor copia e cola no projeto"]
        T3 --> T4["Bugs complexos e regressões"]
        T4 --> T5["Chat longo reexplicando tudo (perda de contexto)"]
        T5 --> T1
        style Tradicional fill:#ffe6e6,stroke:#ff8080,stroke-width:2px
    end
    subgraph Forge["Forge-SDD (Fluxo Estruturado)"]
        F1["Discovery estruturado (/discovery)"] --> F2["Especificação BDD (/nova-feature)"]
        F2 --> F3["Blueprint de Arquitetura & ADR"]
        F3 --> F4["Ciclo TDD (Red -> Green -> Refactor)"]
        F4 --> F5["Auditoria de Segurança & Qualidade (/revisar)"]
        F5 --> F6["Handoff Automatizado e PR via CLI"]
        style Forge fill:#e6ffe6,stroke:#80ff80,stroke-width:2px
    end
```

### Diagrama 2: Ciclo do Agente/TDD — Princípio por Fase
Ilustra como os princípios de engenharia de software (KISS, SOLID, DRY) são acoplados a cada fase do ciclo de desenvolvimento orientado a testes do agente.

```mermaid
graph LR
    RED["RED (Test-First)"] -->|"KISS (Keep It Simple, Stupid)"| GREEN["GREEN (Make it Pass)"]
    GREEN -->|"SOLID + DRY (Don't Repeat Yourself)"| REFACTOR["REFACTOR (Clean Code)"]
    REFACTOR -->|"Próxima task ou cenário"| RED
    
    style RED fill:#ffe6e6,stroke:#ff8080,stroke-dasharray: 5 5,stroke-width:2px
    style GREEN fill:#e6ffe6,stroke:#80ff80,stroke-width:2px
    style REFACTOR fill:#e6f2ff,stroke:#80b3ff,stroke-width:2px
```

### Diagrama 3: Feature Pipeline — Discovery First
Mapeia a ordem cronológica de ações e a transição de responsabilidade desde a ideia até o código verificado.

```mermaid
graph TD
    Idea["Ideia Vaga / Requisito bruto"] -->|"1. Executar /discovery"| Disc["Sessão de Discovery (sdd/discovery/)"]
    Disc --> G1{"GATE 1: Construir?"}
    G1 -->|"Sim"| Feature["2. Executar /nova-feature"]
    Feature -->|"Escreve"| Spec["Story & Spec BDD (sdd/features/feat-XX.md)"]
    Spec --> G2{"GATE 2: Spec Validada?"}
    G2 -->|"Sim"| Dev["3. Executar /proxima-feature (Builder)"]
    Dev -->|"Ciclo TDD"| Code["Código e Testes criados"]
    Code -->|"4. Executar /revisar (Revisor)"| Rev{"GATE 3: Aprovado?"}
    Rev -->|"Sim"| Close["5. Handoff & Auto-PR (Orquestrador)"]
    Rev -->|"Não"| Dev
```

### Diagrama 4: Closed Loop (Laço Fechado)
Garante que a especificação em markdown e o código-fonte de produção nunca divirjam. Qualquer desvio no código é detectado e a especificação é mantida como a única fonte da verdade.

```mermaid
graph LR
    Spec["Especificação BDD (sdd/features/)"] -->|"Especifica / Valida"| Code["Código de Produção"]
    Code -->|"Gera logs de teste"| Review["Revisor (/revisar)"]
    Review -->|"Detecta Drift"| Spec
    Sensors["Sensores (Background Sync)"] -->|"Alinhamento em tempo real"| Spec
    Sensors -->|"Alinhamento em tempo real"| Code
    
    style Sensors fill:#fff2cc,stroke:#d6b656,stroke-width:2px
```

### Diagrama 5: Segurança — Dimensão Transversal
A segurança é tratada como um elemento transversal contínuo por meio de duas frentes de sensores.

```mermaid
graph TD
    subgraph SensoresComputacionais["Sensores Computacionais (Automáticos / Git Hooks)"]
        C1["Secret Detection (bloqueio de tokens/chaves expostas)"]
        C2["Static Linter & Dependency Vulnerability Scan (CVEs)"]
    end
    
    subgraph SensoresInferenciais["Sensores Inferenciais (Semânticos baseados em LLM / /revisar)"]
        I1["Broken Access Control & Lógica de Autenticação"]
        I2["Insecure Design / Validação contra OWASP Top 10"]
    end
    
    Code["Código Submetido"] --> SensoresComputacionais
    Code --> SensoresInferenciais
    
    style SensoresComputacionais fill:#f8cecc,stroke:#b85450,stroke-width:1px
    style SensoresInferenciais fill:#dae8fc,stroke:#6c8ebf,stroke-width:1px
```

### Diagrama 6: Arquitetura do Framework e Suporte Multi-IDE
Mostra como o Forge-SDD mantém a independência de IDE por meio de adaptadores específicos que consomem as regras centrais estruturadas em `sdd/`.

```mermaid
graph TB
    subgraph Core["Core do Framework (Agnóstico)"]
        Rules["Leis e Regras (sdd/memory/constitution.md)"]
        Templates["Templates e Playbooks (sdd/memory/ & sdd/skills/)"]
    end
    
    subgraph AdaptadoresIDE["Adaptadores e Motores de IDE"]
        Cursor["Cursor IDE Adaptador (.cursor/rules/)"]
        Claude["Claude Code Adaptador (CLAUDE.md & .claude/)"]
        Copilot["GitHub Copilot Adaptador (.github/instructions/ & prompts/)"]
        Antigravity["Antigravity Adaptador (GEMINI.md & .gemini/)"]
    end
    
    Core -->|"Injeta contexto e regras"| AdaptadoresIDE
```

---

## 5. Console de Comandos e CLI

O framework disponibiliza comandos divididos em duas categorias: Comandos de Chat (Slash Commands executados pelos agentes dentro do chat da IDE) e Comandos de Terminal (CLI executada localmente).

### Comandos de Chat (IDE / Agente)

* **`/status`**: Analisa e resume o progresso ativo de `sdd/memory/progress.md`.
* **`/discovery <ideia>`**: Abre debate de design de produto e gera arquivos conceituais de hipótese e viabilidade em `sdd/discovery/`.
* **`/nova-feature <nome>`**: Cria branch do git e gera scaffold inicial da funcionalidade em `sdd/features/feat-XX.md`.
* **`/split-features`**: Quebra especificações que ficaram grandes ou complexas em pequenas tarefas independentes.
* **`/proxima-feature`**: Coloca o Builder para atuar na próxima feature da fila (`status: todo`), configurando a branch local.
* **`/revisar`**: Roda o linter, testes, e faz a análise estática e de conformidade de segurança contra OWASP Top 10.
* **`/constitution`**: Executa escaneamento da base de código para garantir o alinhamento de tecnologias listadas na constituição.
* **`/c4-architecture`**: Atualiza ou desenha diagramas Mermaid de alto nível de Contexto e Container (C4 Model).
* **`/doctor`**: Health check da integridade dos arquivos e diretórios da estrutura local.
* **`/archive`**: Compacta e arquiva sessões finalizadas de `progress.md` em `progress-log.md` (limite de 1 KB).
* **`/upgrade-sdd <versao>`**: Aciona o Migrator para alinhar a estrutura física local para a versão do framework desejada.

### CLI Local (`forge-sdd`)

* **`forge-sdd init`**: Cria a estrutura de diretórios e templates do zero.
    * `--yes`: Não-interativo (usa valores default).
    * `--dry-run`: Exibe a árvore de arquivos sem de fato alterá-los.
    * `--stack`, `--db`, `--agent`, `--lang`: Flags para customização da inicialização do repositório.
* **`forge-sdd update`**: Atualiza regras estruturais de agentes, preservando estritamente os diretórios de domínio do projeto (`sdd/`).
* **`forge-sdd version`**: Exibe a versão instalada da CLI do framework.

---

## 6. Mecanismos de Automação (Hooks e Portabilidade)

Para garantir que as regras da constituição e os portões de qualidade não sejam pulados, o framework emprega mecanismos de automação:

1. **Bootstrap de Contexto (IDE Hooks)**: Ao iniciar qualquer interação, as ferramentas de IDE compatíveis leem a memória e o glossário em `sdd/memory/` para injetar terminologias do projeto nas mensagens de sistema do agente.
2. **Process Guard**: Trava de segurança que impede novas tarefas se houver modificações pendentes não revisadas ou testes quebrados no repositório local.
3. **Session & Metrics Tracker**: Telemetria automática que registra o tempo decorrido, commits gerados e consumo de tokens para fins de auditoria de produtividade.
4. **Portabilidade (Fallbacks)**: Em ambientes sem suporte nativo a hooks de lifecycle (como Copilot ou Cursor), as regras de verificação de pré-condições são embutidas nos prompts do kit de instruções do agente, induzindo a verificação manual por parte da IA.
