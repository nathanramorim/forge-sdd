# Discovery 59 — Módulo Visual Plugável das 5 Estações e Personagens (`Forge Studio`)

Este discovery especifica a concepção de produto e arquitetura para um **módulo visual desacoplado e plugável (zero-impacto)** capaz de exibir em tempo real o trabalho de cada agente do Forge-SDD, utilizando exatamente os **5 personagens 3D compactos no padrão `dock-ia`**, o **trilho processual de Handoff (Entrada/Saída)** e o **Inspetor de Governança (Camadas A, B e C)**.

---

## 1. Contexto e Problema (Porquê)

- **Invisibilidade operacional durante a execução autônoma:** Com a chegada do Forge-SDD `v3.0` (`discovery-58`), o ciclo passou a operar em 5 estações coordenadas (`01. O Líder`, `02. PM & Arquiteto`, `03. O Planejador`, `04. Fila de Builders`, `05. Analista de Governança`). Porém, quem acompanha pelo terminal ou chat da IDE enxerga apenas logs textuais esparsos e precisa abrir arquivos Markdown/JSON manualmente para entender quem está trabalhando, qual contrato de entrada/saída está em vigor e qual foi o veredito da Governança.
- **Identidade visual e tangibilidade já aprovadas no `showcase-v3-lp.html` e no `dock-ia`:** O usuário já validou a linguagem visual e os 5 personagens 3D compactos (`Three.js` com estética `dock-ia`, olhos que acompanham o cursor, marcas peitorais `ring`, `diamond`, `bar`, `wave`, `tri` e animações por estado). Falta transformar essa vitrine estática em um **observador vivo** do repositório.
- **Risco de acoplamento no Core da CLI:** O `forge-sdd` tem como princípio inegociável da Constituição (`ARQ-03`, `ARQ-04`) ser um binário leve, determinístico e sem dependências externas em runtime. Portanto, a interface visual **jamais pode ser um requisito bloqueante** ou alterar a rotina dos agentes: ela deve funcionar como um **módulo sidecar 100% plugável e opcional**.

---

## 2. Para Quem

- **Engenheiro / Mantenedor Humano (Camada C no Comando Estratégico):** Deseja abrir uma janela simples (no navegador ou companion app) quando quiser acompanhar visualmente em qual estação o projeto está, quais tarefas estão sendo construídas na fila, o que o Analista de Governança aprovou ou vetou e quais Auto-Skills emergiram — e fechar a janela a qualquer momento sem afetar em nada o funcionamento do `forge-sdd`.
- **Times e Stakeholders em Demonstrações / Pair Programming:** Conseguem enxergar o processo de engenharia orientada por especificação (SDD) acontecendo ao vivo através dos 5 personagens e dos contratos de Handoff (`Recebe` → `Entrega`).

---

## 3. O que já existe (Reuse Check — `DUP-01` / `DUP-02`)

Antes de propor qualquer componente novo, mapeamos os ativos já prontos no `forge-sdd` e no repositório irmão `/Users/nathanramorim/git/dock-ia`:

| Ativo Existente | Onde reside | O que já entrega pronto | Decisão (`reusar`, `estender`, `criar`) |
| --- | --- | --- | --- |
| **Fila Determinística (`queue.json`)** | `internal/queue/queue.go` (`sdd/.queue/queue.json`) | Persiste `feature_id`, `branch`, `spec_status`, `updated_at` e `items[]` com `task_id`, `title`, `spec_path`, `status` (`queued`, `building`, `review_pending`, `approved`, `changes_requested`), `rules`, `skills_used`, `attempts` e `governance_verdict` (`conforme`, `violacao`, `duvida` + `achados[]` e `candidate_skill`). | **`reusar` (leitura passiva)** — fonte primária das Estações 04 e 05. |
| **Estado e Métricas SDD** | `sdd/memory/progress.md`, `sdd/features/index.md`, `sdd/.metrics/session-*.json` (`internal/telemetry`) | Indica a feature ativa, histórico de sessões, comandos executados, duração e status macro do projeto. | **`reusar` (leitura passiva)** — fonte para identificar transições das Estações 01, 02 e 03. |
| **Palco 3D dos 5 Personagens + Esteira Processual** | `docs/showcase-v3-lp.html` | Contém a implementação completa e aprovada dos **5 mascotes 3D em Three.js** (`buildCompactBot`, geometrias procedurais, marcas peitorais, animações por estado, rastreamento de mouse), o **Trilho Processual (`01 → 02 → 03 → 04 → 05`)**, os cards de contrato (`Recebe` / `Entrega`) e o **Inspetor de Governança**. | **`reusar` e `estender`** — extrair o front-end já construído para ser alimentado por eventos reais da pasta `sdd/` em vez de apenas simulação local. |
| **Design System e Estados de Mascote `dock-ia`** | `/Users/nathanramorim/git/dock-ia/produto/design-system.md` e `app/packages/ui-kit/src/mascot.ts` | Tokens oficiais (`--bg: #0d0f14`, `--bg2: #12151d`, `--bg3: #1a1e29`, `--accent: #4f8ef7`, `--ok: #3ecf8e`, `--warn: #f5b04a`, `--bad: #f2677a`), tipografia (`Montserrat` + `JetBrains Mono`) e máquina de estados (`idle`, `working`, `thinking`, `searching`, `approval`, `finished`, `error`). | **`reusar`** — manter 100% de fidelidade visual e semântica com o ecossistema `dock-ia`. |
| **Ponte de Observação em Tempo Real (File Watcher → UI)** | *Não existe no `forge-sdd`* | Falta apenas um observador passivo (read-only) que leia a pasta `sdd/` e transmita o snapshot JSON via SSE/HTTP para a interface visual. | **`criar` (desacoplado)** — micro-servidor/módulo plugável sem impacto nos comandos atuais. |

---

## 4. Como (Visão Macro da Solução e Desacoplamento Zero-Impacto)

### 4.1 Regra de Ouro: Acoplamento Zero-Impacto (Observer Sidecar Passivo)
1. **Independência Total da Rotina SDD:** Nenhum comando do `forge-sdd` (`init`, `update`, `doctor`, `status`, `queue`, `gate`, `skills`, `session`) nem qualquer prompt de agente (`.agents/commands/`, `.agents/roles/`) precisa saber se a interface visual está aberta, fechada ou sequer instalada.
2. **Comunicação Exclusiva via Sistema de Arquivos (`sdd/`):** O `forge-sdd` continua apenas lendo e gravando seus arquivos normais (`sdd/.queue/queue.json`, `sdd/memory/progress.md`, `sdd/discovery/*`, `sdd/features/*`, `sdd/.metrics/*`, `.agents/skills/*`). O módulo visual atua **estritamente em modo somente-leitura (Read-Only Observer)** sobre o diretório `sdd/`.
3. **Pulso Opcional Não-Bloqueante (`live-state.json`):** Quando um agente estiver no meio de uma etapa conversacional antes de gravar `queue.json` (ex.: PM & Arquiteto gerando um Discovery), o módulo infere a estação ativa pela modificação recente de arquivos em `sdd/discovery/` ou `sdd/memory/progress.md` (e, opcionalmente, suporta ler um arquivo efêmero `sdd/.queue/live-state.json` caso o usuário deseje granularidade extra no futuro, sem jamais exigir sua presença).

### 4.2 Os 5 Personagens e o Mapeamento Determinístico de Estados

A interface apresenta exatamente os **5 personagens validados** em `docs/showcase-v3-lp.html`:

| Estação | Personagem | Cor (`color`) | Marca Peitoral (`mark`) | Estado Ativo (`dock-ia`) | Gatilho Real no Repositório (`sdd/`) | Contratos Visíveis no Card (`Recebe` → `Entrega`) |
| --- | --- | --- | --- | --- | --- | --- |
| **01** | **O Líder** (*Orquestrador*) | `#f2a65a` (Âmbar) | `ring` (Anel) | `approval` / `working` / `idle` | Ativo quando aguarda Gate 1 (`spec_status != "reviewed"`), quando analisa retorno da fila ou quando todas as tasks chegam a `approved` (Camada C). | **Recebe:** Demanda do usuário + Parecer JSON da Governança.<br>**Entrega:** Aprovação Gate 1 (`FLX-01`), despacho de fila e release. |
| **02** | **PM & Arquiteto** (*Concepção*) | `#8aa4ff` (Azul) | `diamond` (Diamante) | `thinking` / `finished` / `idle` | Ativo quando há arquivo `sdd/discovery/discovery-*.md` ou `criteria-*.md` sendo criado/editado sem `plan-*.md` correspondente ou em revisão de Spec. | **Recebe:** Intenção de produto + `constitution.md` + `concepts.md`.<br>**Entrega:** `discovery-ID.md` + `criteria-ID.md` (Reuse Check `DUP-02` + C4). |
| **03** | **O Planejador** (*Plan & Decomposição*) | `#c9a7ff` (Violeta) | `bar` (Barras) | `working` / `finished` / `idle` | Ativo quando `plan-*.md` está sendo gerado ou quando `sdd/.queue/queue.json` acaba de ser populado com itens em status `queued`. | **Recebe:** Spec aprovada no Gate 1 + `.agents/rules/*.md`.<br>**Entrega:** `plan-ID.md` + `sdd/.queue/queue.json` com rules mapeadas. |
| **04** | **Fila de Builders** (*Executores Isolados*) | `#4fd1b5` (Turquesa) | `wave` (Onda) | `working` / `finished` / `error` / `idle` | `working` quando existe task com `status == "building"`; `error` quando `status == "changes_requested"`; `finished` quando todas estão `approved`. | **Recebe:** Task isolada da fila + `lessons.md` + Rules/Skills da task.<br>**Entrega:** Código + Testes (`go test`/`go vet`) + status `review_pending`. |
| **05** | **Analista de Governança** (*Auditor Read-Only*) | `#ff8a5c` (Coral) | `tri` (Triângulo) | `searching` / `finished` / `error` / `idle` | `searching` quando existe task em `review_pending`; `error` se o último `governance_verdict.veredito == "violacao"`; `finished` se `"conforme"`. | **Recebe:** `<spec>` + `<diff>` (não confiáveis `FLX-08`) + check `P-xx`.<br>**Entrega:** Parecer JSON (`S-01..S-07`) com evidência + proposta de Auto-Skill. |

### 4.3 Componentes da Interface Visual
1. **Header de Conexão e Modo (Ao Vivo vs. Simulação):**
   - Mostra o nome do projeto/branch atual lido de `queue.json` / `.sddrc`, indicador `LIVE (SSE)` conectado ao diretório `./sdd` e um botão de **Modo Simulação (Demo)** para quando o usuário quiser testar ou apresentar as animações dos 5 agentes mesmo com o repositório parado.
2. **Trilho Processual de Handoff (`01 → 02 → 03 → 04 → 05 ↺`):**
   - Barra horizontal mostrando o fluxo sequencial e o loop de retorno (`05 ↺ 04` em caso de `changes_requested` e `05 → 01` em caso de `approved`), destacando com pulso luminoso a estação ativa no momento.
3. **Grid das 5 Estações com Personagens 3D (`Three.js`):**
   - Cada estação exibe seu mascote 3D interativo no topo, badge de estado em tempo real (`EXECUTANDO`, `ARQUITETANDO`, `AUDITANDO`, `AGUARDANDO GATE`, `CONCLUÍDO`, `VIOLAÇÃO`), a task ou arquivo que o agente está manipulando agora e o bloco de contratos **`Recebe (Entrada)`** / **`Entrega (Saída)`**.
4. **Inspetor de Governança e Fila em Tempo Real:**
   - Painel inferior exibindo as tasks de `sdd/.queue/queue.json`, o JSON real do último `governance_verdict` (achados, arquivo, linha, regra violada e severidade) e as Auto-Skills propostas (`candidate_skill` / `.agents/skills/`).

---

## 5. Comparativo de Opções Tecnológicas (Simples e Desacopladas)

Avaliamos três caminhos arquiteturais para entregar o módulo mantendo simplicidade máxima e zero impacto no fluxo do `forge-sdd`:

| Critério | **Opção A (Recomendada): Módulo Web Local Zero-Config (`npx` ou Subcomando Sidecar via `embed.FS` + SSE)** | **Opção B: Plugin/Adapter Nativo dentro do app Desktop `dock-ia` (Tauri)** | **Opção C: Extensão VS Code (Webview Panel)** |
| --- | --- | --- | --- |
| **Como funciona** | Um micro-servidor HTTP local lê a pasta `./sdd` em modo read-only e transmite mudanças via **Server-Sent Events (SSE)** para uma SPA HTML/JS/Three.js derivada diretamente de `docs/showcase-v3-lp.html`. Pode rodar como pacote independente (`npx @forge-sdd/studio`) ou comando opcional (`forge-sdd studio`). | O app desktop `dock-ia` ganha um conector que assiste `sdd/.queue/queue.json` e renderiza as 5 estações dentro da janela/notch do `dock-ia`. | Uma extensão `.vsix` abre uma aba Webview dentro do VS Code/Cursor e observa os arquivos do workspace. |
| **Impacto na rotina do `forge-sdd`** | **Zero.** Não altera nenhum comando nem template do SDD. Só consome memória/CPU quando o usuário abre a URL local (`http://localhost:4321`). | **Zero** no `forge-sdd`, mas exige que o usuário tenha o app desktop `dock-ia` instalado e em execução. | **Zero** no CLI, mas fica preso a IDEs baseadas no VS Code (não funciona se o usuário estiver apenas no terminal com Claude Code ou outra IDE). |
| **Simplicidade de implementação e manutenção** | **Altíssima.** 90% do código visual já está pronto em `docs/showcase-v3-lp.html` (um único arquivo HTML/CSS/Three.js sem build complexo). O backend precisa apenas de ~100 linhas para servir o HTML e emitir um JSON consolidado via SSE quando arquivos em `sdd/` mudarem. | **Média/Alta.** Exige sincronizar releases entre dois repositórios (`forge-sdd` e `dock-ia`) e adaptar o layout de notch para caber o fluxo das 5 estações + contratos. | **Média.** Requer pipeline de empacotamento de extensão VS Code, API de Webview e permissões de workspace. |
| **Portabilidade de Tecnologia** | Funciona em qualquer OS (macOS, Linux, Windows), qualquer navegador e também pode ser embutido num `<webview>` do `dock-ia` ou VS Code depois. | Restrito ao ambiente Desktop onde o `dock-ia` roda. | Restrito ao VS Code / forks. |

> **Recomendação da Arquitetura (Opção A com ponte aberta para Opção B):**
> Adotar a **Opção A** como entrega principal: um módulo web autônomo ultraleve (Front-end em **HTML5 + CSS Tokens `dock-ia` + Three.js** evoluído de `docs/showcase-v3-lp.html`, alimentado por um endpoint passivo `/api/state` e `/api/events` via **SSE** que lê `./sdd`).
> Como o contrato de estado emitido por `/api/state` é um JSON limpo e padronizado, o `dock-ia` (Opção B) também poderá consumi-lo diretamente no futuro sem reescrever nenhuma regra.

---

## 6. Escopo (In / Out)

### In Scope (Dentro do Escopo)
- Módulo visual desacoplado com os **5 personagens 3D compactos** (`01. O Líder`, `02. PM & Arquiteto`, `03. O Planejador`, `04. Fila de Builders`, `05. Analista de Governança`) usando os tokens e animações do `dock-ia`.
- Leitura passiva (Read-Only) em tempo real de `sdd/.queue/queue.json`, `sdd/memory/progress.md`, `sdd/discovery/*`, `sdd/features/index.md` e `.agents/skills/*`.
- Atualização automática da interface via **Server-Sent Events (SSE)** ou polling leve de fallback, sem necessidade de recarregar a página.
- Visualização do **Trilho Processual (Handoff De → Para)**, **Contratos de Entrada/Saída de cada estação**, **Fila de Tasks** e **Inspetor de Governança (Camadas A, B e C)** com destaque para violações e evidências.
- Modo dual na própria interface: **Modo Ao Vivo (`LIVE`)** sincronizado com a pasta `./sdd` do projeto e **Modo Simulação (`DEMO`)** interativo para explorar os estados manualmente quando nenhum agente estiver rodando.

### Out of Scope (Fora do Escopo neste Discovery)
- Transformar a interface visual em um disparador que executa prompts de IA diretamente (a interface é **observacional/read-only** para garantir risco zero e desacoplamento total).
- Alterar o formato obrigatório dos arquivos markdown ou exigir que os agentes chamem APIs de rede para reportar progresso.
- Substituir o app desktop `dock-ia` completo (notificações de sistema, notch nativo do macOS e integrações externas do `dock-ia` continuam no repositório próprio dele).

---

## 7. Questões Abertas para Decisão Humana no Gate 1 (`FLX-01`)

Antes de acionar o **Planejador (`Estação 03`)** para gerar o `plan-59`, precisamos da sua decisão em **2 pontos simples**:

1. **Forma de distribuição/empacotamento preferida (dentro da Opção A):**
   - **A1 (Subcomando Opcional no próprio binário Go):** `forge-sdd studio` (ou `forge-sdd ui --port 4321`) embutido via `embed.FS` usando apenas a `net/http` da stdlib do Go (zero dependência externa, já vem pronto em qualquer máquina que tenha o `forge-sdd`, mas sem interferir em nenhum outro comando).
   - **A2 (Pacote Node/NPX ou pasta separada):** Um pacote/repositório ou subpasta JS independente (ex.: `npx forge-sdd-studio` ou `node studio/server.mjs`) totalmente fora do binário Go.
2. **Granularidade nas Estações 01, 02 e 03:**
   - A inferência automática apenas observando os arquivos existentes (`progress.md`, `sdd/discovery/*`, `sdd/.queue/queue.json`) já é suficiente para você, ou deseja também que a CLI grave opcionalmente um arquivo efêmero `sdd/.queue/live-state.json` quando comandos CLI forem chamados?
