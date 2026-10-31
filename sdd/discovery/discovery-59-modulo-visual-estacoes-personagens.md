# Discovery 59 — Módulo Visual Plugável das 5 Estações e Personagens (`Forge Studio`) + Filas Idempotentes por Agente

Este discovery especifica a concepção de produto e arquitetura para o **módulo visual desacoplado e plugável (`forge-sdd studio`)** capaz de exibir em tempo real o trabalho de cada agente do Forge-SDD utilizando os **5 personagens compactos das estações**, o **trilho processual de Handoff (Entrada/Saída)** e o **Inspetor de Governança (Camadas A, B e C)**, sustentado por **filas dedicadas e estritamente idempotentes por estação (`sdd/.queue/stations/<station>.json`)**.

> **Decisões Resolvidas no Portão de Revisão Humana (`FLX-01` — Gate 1 Aprovado):**
> 1. **Empacotamento escolhido (Opção A1):** Subcomando opcional no próprio binário Go (`forge-sdd studio`, via `net/http` + `embed.FS` + `SSE` da stdlib). Permanece 100% inerte e sem impacto na rotina caso não seja invocado.
> 2. **Filas Dedicadas por Agente com Idempotência Estrita (`sdd/.queue/stations/<station>.json`):** Em vez de inferir indiretamente as Estações 01, 02 e 03 por data de modificação de arquivos markdown ou concentrar tudo em um único arquivo com risco de duplicação/contenção, **cada uma das 5 estações passa a ter sua própria fila persistida** (`lider.json`, `pm-architect.json`, `planner.json`, `builder.json`, `governance.json`) com **chave de idempotência determinística (`idempotency_key`)** e **escrita atômica (`write temp + rename`)**.

---

## 1. Contexto e Problema (Porquê)

- **Invisibilidade operacional durante a execução autônoma:** Com a chegada do Forge-SDD `v3.0` (`discovery-58`), o ciclo passou a operar em 5 estações coordenadas (`01. O Líder`, `02. PM & Arquiteto`, `03. O Planejador`, `04. Fila de Builders`, `05. Analista de Governança`). Porém, quem acompanha pelo terminal ou chat da IDE enxerga apenas logs textuais esparsos e precisa abrir arquivos Markdown/JSON manualmente para entender quem está trabalhando, qual contrato de entrada/saída está em vigor e qual foi o veredito da Governança.
- **Assimetria entre a Fila do Builder e as demais Estações + Risco de Duplicação (Idempotência):** Até a Feature 58, apenas o Builder possuía fila formal (`sdd/.queue/queue.json`). Se um comando ou subagente de Concepção, Planejamento ou Governança fosse reexecutado (retry de LLM, reconexão ou duplo disparo), não havia uma fila própria por estação nem garantia de idempotência (`idempotency_key`) para evitar reprocessamento ou eventos duplicados.
- **Identidade visual e tangibilidade já aprovadas no `showcase-v3-lp.html`:** O usuário já validou a linguagem visual e os 5 personagens compactos (olhos que acompanham o cursor, marcas peitorais `ring`, `diamond`, `bar`, `wave`, `tri` e animações por estado). Falta conectar essa vitrine diretamente às filas reais das 5 estações.
- **Risco de acoplamento no Core da CLI:** O `forge-sdd` tem como princípio inegociável da Constituição (`ARQ-03`, `ARQ-04`) ser um binário leve, determinístico e sem dependências externas em runtime. Portanto, a interface visual **jamais pode ser um requisito bloqueante** para os agentes: o `forge-sdd studio` atua como um **observador passivo (Read-Only Sidecar)** sobre as filas das 5 estações.

---

## 2. Para Quem

- **Engenheiro / Mantenedor Humano (Camada C no Comando Estratégico):** Roda `forge-sdd studio` quando quiser acompanhar visualmente no navegador em qual estação o projeto está, o que está na fila de cada um dos 5 agentes, o que o Analista de Governança aprovou ou vetou e quais Auto-Skills emergiram — podendo fechar a aba ou encerrar o comando a qualquer momento sem interromper nada no `forge-sdd`.
- **Agentes das 5 Estações (Claude, Gemini, Copilot):** Passam a contar com filas individuais determinísticas e idempotentes (`sdd/.queue/stations/<station>.json`), garantindo handoffs seguros (`De → Para`) sem duplicação de tarefas em caso de retries.

---

## 3. O que já existe (Reuse Check — `DUP-01` / `DUP-02`)

Antes de propor qualquer componente novo, mapeamos os ativos já prontos no `forge-sdd`:

| Ativo Existente | Onde reside | O que já entrega pronto | Decisão (`reusar`, `estender`, `criar`) |
| --- | --- | --- | --- |
| **Fila Determinística do Builder (`queue.json`)** | `internal/queue/queue.go` (`sdd/.queue/queue.json`) | Persiste `feature_id`, `branch`, `spec_status`, `updated_at` e `items[]` com `task_id`, `title`, `spec_path`, `status`, `rules`, `skills_used`, `attempts` e `governance_verdict`. | **`estender`** — evoluir `internal/queue` para suportar as **5 filas por estação** (`sdd/.queue/stations/<station>.json`) com `idempotency_key`, escrita atômica e retrocompatibilidade total com `sdd/.queue/queue.json`. |
| **Estado e Métricas SDD** | `sdd/memory/progress.md`, `sdd/features/index.md`, `sdd/.metrics/session-*.json` (`internal/telemetry`) | Indica a feature ativa, histórico de sessões, comandos executados, duração e status macro do projeto. | **`reusar` (leitura passiva)** — contexto suplementar para o cabeçalho do projeto no Studio. |
| **Palco dos 5 Personagens + Esteira Processual** | `docs/showcase-v3-lp.html` | Contém a implementação completa e aprovada dos **5 personagens das estações** (`createForgeMascot3D`, marcas peitorais, animações por estado, rastreamento de mouse), o **Trilho Processual (`01 → 02 → 03 → 04 → 05`)**, os cards de contrato (`Recebe` / `Entrega`), os tokens do Design System (`--bg: #0d0f14`, `--bg2: #12151d`, `--bg3: #1a1e29`, `Montserrat` + `JetBrains Mono`) e o **Inspetor de Governança**. | **`reusar` e `estender`** — embutir via `embed.FS` no subcomando `forge-sdd studio` alimentado em tempo real via SSE pelas 5 filas. |
| **Servidor Local Read-Only (`forge-sdd studio`)** | *Não existe no `forge-sdd`* | Servidor HTTP/SSE na `stdlib` Go que observa `sdd/.queue/stations/*.json` e projeta o estado vivo para o navegador. | **`criar` (`internal/studio` + `cmd/forge-sdd/studio.go`)** — subcomando opcional zero-impacto seguindo a skill `novo-subcomando-cli-sdd`. |

---

## 4. Como (Visão Macro da Solução: Filas por Agente + Idempotência + `forge-sdd studio`)

### 4.1 Filas Dedicadas por Estação com Idempotência Estrita (`sdd/.queue/stations/<station>.json`)

Cada um dos 5 agentes possui sua própria fila dedicada em disco, eliminando inferências frágeis e contenção de escrita entre agentes paralelos:

1. **`sdd/.queue/stations/lider.json` (`01. O Líder`):**
   - Recebe demandas iniciais, pedidos de aprovação humana (Gate 1 `FLX-01` e Camada C `FLX-06`), consolidações de fim de onda e decisões de release.
2. **`sdd/.queue/stations/pm-architect.json` (`02. PM & Arquiteto`):**
   - Recebe itens de concepção (`discovery-ID` e `criteria-ID`), registrando a execução do Reuse Check (`DUP-02`) e diagramas C4.
3. **`sdd/.queue/stations/planner.json` (`03. O Planejador`):**
   - Recebe Specs aprovadas no Gate 1 para decomposição em `plan-ID.md` e geração das tasks mapeadas com IDs da Constitution e `.agents/rules/`.
4. **`sdd/.queue/stations/builder.json` (`04. Fila de Builders`):**
   - Fila de execução isolada de código e testes das tasks (`queued`, `building`, `review_pending`, `approved`, `changes_requested`), mantida em sincronia espelhada com `sdd/.queue/queue.json` para retrocompatibilidade 100% com scripts e gates existentes.
5. **`sdd/.queue/stations/governance.json` (`05. Analista de Governança`):**
   - Fila de auditoria Read-Only (Camada A + Camada B Semântica `S-01..S-07`), armazenando os itens aguardando parecer, o veredito JSON com evidências e as propostas de Auto-Skills (`DUP-05`).

#### Garantias de Idempotência Estrita:
- **Chave Determinística (`idempotency_key`):** Todo item ou handoff entre estações calcula um hash canônico `sha256(station + ":" + item_id + ":" + operation + ":" + artifact_path)`.
- **No-Op Seguro em Retries:** Se um agente ou comando CLI tentar enfileirar (`enqueue`), transferir (`handoff`), assumir (`claim`), concluir (`complete`) ou revisar (`review`) um item com a mesma `idempotency_key` já processada naquela estação, o motor `internal/queue` reconhece a chave, **não duplica o item**, **não incrementa contadores indevidamente** e retorna o estado existente com sucesso idempotente.
- **Escrita Atômica (`write temp + rename`) e Versão Monotônica (`version`):** Cada arquivo `sdd/.queue/stations/<station>.json` é gravado primeiro em um arquivo temporário no mesmo diretório e renomeado atomicamente via `os.Rename`, incrementando o campo inteiro `version`. Isso garante que o `forge-sdd studio` (SSE) nunca leia um JSON corrompido ou pela metade, mesmo com Builders paralelos.

### 4.2 Regra de Ouro: Acoplamento Zero-Impacto do `forge-sdd studio`
- **100% Opcional e Passivo:** O core do `forge-sdd` e as filas `sdd/.queue/stations/*.json` funcionam no disco independentemente de o `forge-sdd studio` estar rodando ou não.
- **Observação Somente-Leitura (Read-Only):** Quando o usuário executa `forge-sdd studio`, o servidor HTTP local (`127.0.0.1:4321`) apenas observa as 5 filas (`sdd/.queue/stations/*.json`) e transmite qualquer mudança de `version` / `idempotency_key` via **Server-Sent Events (`SSE`)** para os 5 personagens 3D no navegador.

### 4.3 Os 5 Personagens Conectados Diretamente às Suas Filas

| Estação | Personagem | Cor (`color`) | Marca (`mark`) | Fila Dedicada | Estados Derivados da Própria Fila | Contratos Visíveis no Card (`Recebe` → `Entrega`) |
| --- | --- | --- | --- | --- | --- | --- |
| **01** | **O Líder** (*Orquestrador*) | `#f2a65a` (Âmbar) | `ring` | `stations/lider.json` | `approval` (aguardando Gate 1 ou Camada C) · `working` (orquestrando handoff) · `finished` · `idle` | **Recebe:** Demanda do usuário + Parecer JSON da Governança.<br>**Entrega:** Aprovação Gate 1 (`FLX-01`), despacho de fila e release. |
| **02** | **PM & Arquiteto** (*Concepção*) | `#8aa4ff` (Azul) | `diamond` | `stations/pm-architect.json` | `thinking` (elaborando Discovery/Criteria/C4) · `finished` (Spec entregue) · `idle` | **Recebe:** Intenção de produto + `constitution.md` + `concepts.md`.<br>**Entrega:** `discovery-ID.md` + `criteria-ID.md` (Reuse Check `DUP-02` + C4). |
| **03** | **O Planejador** (*Plan & Decomposição*) | `#c9a7ff` (Violeta) | `bar` | `stations/planner.json` | `working` (quebrando tasks e mapeando rules) · `finished` (plano e fila gerados) · `idle` | **Recebe:** Spec aprovada no Gate 1 + `.agents/rules/*.md`.<br>**Entrega:** `plan-ID.md` + Handoff idempotente para `builder.json`. |
| **04** | **Fila de Builders** (*Executores Isolados*) | `#4fd1b5` (Turquesa) | `wave` | `stations/builder.json` (`queue.json`) | `working` (`building`) · `error` (`changes_requested`) · `finished` (todas `approved`) · `idle` | **Recebe:** Task isolada da fila + `lessons.md` + Rules/Skills da task.<br>**Entrega:** Código + Testes (`go test`/`go vet`) + Handoff para `governance.json`. |
| **05** | **Analista de Governança** (*Auditor Read-Only*) | `#ff8a5c` (Coral) | `tri` | `stations/governance.json` | `searching` (auditando item em `review_pending`) · `error` (`violacao`) · `finished` (`conforme`) · `idle` | **Recebe:** `<spec>` + `<diff>` (não confiáveis `FLX-08`) + check `P-xx`.<br>**Entrega:** Parecer JSON (`S-01..S-07`) com evidência + proposta de Auto-Skill. |

---

## 5. Decisão Tecnológica Consolidada (Opção A1 — `forge-sdd studio`)

| Critério | **Opção A1 Escolhida: Subcomando Opcional `forge-sdd studio` (`net/http` + `embed.FS` + `SSE`)** | Opção B (Descartada): App Desktop Externo | Opção C (Descartada): Extensão VS Code |
| --- | --- | --- | --- |
| **Como funciona** | Subcomando Go opcional que serve a SPA embutida via `embed.FS` (`ARQ-03`) e expõe `/api/state` e `/api/events` (SSE) lendo `sdd/.queue/stations/*.json`. | Aplicação desktop externa separada do binário CLI. | Preso a IDEs baseadas no VS Code. |
| **Dependências externas** | **Zero (`ARQ-04`, `SEG-07`).** Usa exclusivamente a biblioteca padrão do Go no binário. | Requer runtime desktop externo. | Requer Node/VSIX. |
| **Impacto na rotina do `forge-sdd`** | **Zero.** Totalmente desacoplado; só roda quando o usuário chama `forge-sdd studio`. | Zero no CLI. | Zero no CLI. |

---

## 6. Escopo (In / Out)

### In Scope (Dentro do Escopo)
- Filas dedicadas por estação em `sdd/.queue/stations/{lider,pm-architect,planner,builder,governance}.json` com controle estrito de idempotência (`idempotency_key`, histórico de chaves processadas, `version` monotônico e escrita atômica `temp + rename`).
- Retrocompatibilidade transparente entre `sdd/.queue/stations/builder.json` e `sdd/.queue/queue.json`.
- Subcomandos determinísticos na CLI (`forge-sdd queue station ...` / `handoff`) para consultar e transicionar itens entre as 5 estações de forma idempotente.
- Subcomando opcional `forge-sdd studio` (`--port`, `--host`, `--no-open`) servindo a interface visual com os **5 personagens das estações**, **Trilho Processual (`01 → 02 → 03 → 04 → 05 ↺`)**, **Contratos de Entrada/Saída**, **Filas por Agente** e **Inspetor de Governança**.
- Modo dual na interface: **Modo Ao Vivo (`LIVE`)** via SSE e **Modo Simulação (`DEMO`)** interativo.

### Out of Scope (Fora do Escopo)
- Execução de chamadas de LLM diretamente por dentro da UI web (o Studio é estritamente um observador Read-Only para manter segurança `SEG-01..06` e desacoplamento total).
- Dependências externas no `go.mod` (`ARQ-04`).

---

## 7. Handoff para `/split-features`

Com o Gate 1 (`FLX-01`) aprovado e o plano técnico elaborado pelo Planejador (`Estação 03`), os artefatos completos deste Discovery são:
- [discovery-59-modulo-visual-estacoes-personagens.md](file:///Users/nathanramorim/git/forge-sdd/sdd/discovery/discovery-59-modulo-visual-estacoes-personagens.md)
- [criteria-59-modulo-visual-estacoes-personagens.md](file:///Users/nathanramorim/git/forge-sdd/sdd/discovery/criteria-59-modulo-visual-estacoes-personagens.md)
- [plan-59-modulo-visual-estacoes-personagens.md](file:///Users/nathanramorim/git/forge-sdd/sdd/discovery/plan-59-modulo-visual-estacoes-personagens.md)

**Próximo passo:** executar `/split-features` para materializar as subtarefas em `sdd/features/feat-59-modulo-visual-estacoes-personagens/` e enfileirá-las na Fila do Builder.
