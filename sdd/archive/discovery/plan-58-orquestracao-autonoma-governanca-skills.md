# Plano de Discovery 58 — Orquestração Independente, Fila do Builder, Portão de Governança (A/B/C), Constituição por IDs e Auto-Skills

Roadmap preliminar e proposta de quebra para o `/split-features` dentro de `sdd/features/feat-58-orquestracao-autonoma-governanca-skills/`.
Incorpora os padrões genéricos do repositório [`busize-solutions/planning-brain-governanca`](https://github.com/busize-solutions/planning-brain-governanca), respeita a **Rule of 7** (7 subfeatures/tasks ordenadas de baixo para cima) e opera sobre uma **única branch** (`feat/58-orquestracao-autonoma-governanca-skills`, conforme Regra 15).

---

## Quebra Preliminar de Tasks / Subfeatures

### Fase 1 — Fundação: Constituição por IDs, Rules Modulares e Templates de Governança

- [ ] **`feat-58-01` — Constituição por IDs Estáveis + Rules Modulares (`.agents/rules/`) + Templates de Governança**
  - **Escopo:**
    1. Estruturar `sdd/memory/constitution.md` com **IDs canônicos estáveis** (`ARQ-*`, `SEG-*`, `DUP-*`, `COD-*`, `TST-*`, `FLX-*`, `GOV-*`), tabela de **Divergências Conhecidas**, Mapa de Roteamento e suporte a blocos demarcados (`<!-- <pacote>-governanca:inicio -->`).
    2. Criar as rules nativas de "como cumprir" em `.agents/rules/sdd-{arquitetura,seguranca,reuso-e-duplicacao,codigo-e-testes,fluxo-e-portao}.md` com frontmatter (`description`, `globs`, `regras: [IDs]`) sem duplicar o texto da Constitution (`GOV-04`).
    3. Adicionar ao scaffold os templates genéricos do `planning-brain-governanca`: `sdd/spec/concepts.md` (Catálogo de Conceitos `DUP-01`), `sdd/spec/adr/adr-0000-modelo.md` (`GOV-01`) e `.github/pull_request_template.md` (`FLX-03`, `GOV-02`).
  - **Validação:** `go test ./internal/scaffold/...` e golden tests atualizados.

- [ ] **`feat-58-02` — Papéis Canônicos das 5 Estações (`.agents/roles/`) e Adaptadores Multi-Agente**
  - **Escopo:**
    1. Unificar os papéis em `.agents/roles/` (fonte única `DUP-10`): **Líder** (`lider`), **Concepção PM + Arquiteto** (`pm-architect`), **Planejador** (`planner`), **Builder** (`builder`), **Analista de Governança** (`governance` — com o prompt de Revisor Semântico Camada B blindado contra prompt injection `FLX-08`) e **Archivist/Migrator**.
    2. Gerar adaptadores finos para `.claude/agents/*.md` (com restrição de `tools` read-only no Analista de Governança), `.gemini/skills/*.chatmode.md` e `.github/chatmodes/*.chatmode.md`.
  - **Validação:** `go test ./internal/scaffold/...` e verificação via `forge-sdd doctor` dos papéis nos 3 agentes.

---

### Fase 2 — Concepção (Reuse Check), Gate de Revisão de Spec e Planejamento

- [ ] **`feat-58-03` — Fluxo Líder + PM/Arquiteto + Planejador com Reuse Check (`DUP-02`) e Gate de Spec (`FLX-01`)**
  - **Escopo:**
    1. Atualizar `.agents/commands/{discovery,split-features,nova-feature}.md` (e templates) para exigir nas Specs as seções **"Modelo do módulo" (`ARQ-11`)**, **"O que já existe" (`DUP-02` — Reuse Check)** e **"Dependências novas" (`SEG-07`)**.
    2. Instituir o **Gate de Revisão de Spec (`FLX-01`)** antes do planejamento/execução.
    3. Instruir o **Planejador (`Plan`)** a consultar os IDs da Constitution e o frontmatter de `.agents/rules/` para anotar em cada task do plano quais rules e skills o Builder e a Governança devem carregar, devolvendo ao **Líder** para decisão humana de quebra e envio à fila.
  - **Validação:** Golden tests de `.agents/commands/` e testes unitários de scaffold passando.

---

### Fase 3 — Motor da Fila do Builder e Portão de Governança (Camadas A, B e C)

- [ ] **`feat-58-04` — Motor Determinístico da Fila (`forge-sdd queue`) e Camada A do Portão (`forge-sdd gate`)**
  - **Escopo:**
    1. Implementar `internal/queue` e `forge-sdd queue` (`enqueue`, `list`, `next`, `complete`, `review`) gerenciando `sdd/.queue/queue.json`.
    2. Implementar `internal/gate` e `forge-sdd gate` com as verificações determinísticas da **Camada A** (`P-01` branch, `P-02` spec revisada antes do código, `P-03` seção "O que já existe" preenchida, `P-13` segredos, `P-14` dependências novas justificadas, `P-15` `go vet`/`go test`, `P-20` exceções com prazo válido, `P-21` orçamento `progress.md`, `P-23` rules citando IDs válidos da Constitution), suportando modos `observacao` e `bloqueio`.
  - **Validação:** Testes unitários em `internal/queue/...` e `internal/gate/...` com fixtures de casos conformes e não-conformes.

- [ ] **`feat-58-05` — Estação do Analista de Governança (Camada B JSON) e Painel Executivo do Líder (Camada C)**
  - **Escopo:**
    1. Validar em `forge-sdd queue review` o schema JSON do parecer da **Camada B (Analista de Governança)**: `veredito` (`conforme | violacao | duvida`), `achados[]` (exigindo `verificacao S-01..S-07`, `regra`, `arquivo`, `linha`, `evidencia`, `explicacao`, `correcao`), `candidate_skill` e `resumo`.
    2. Atualizar `/proxima-feature`, `/revisar` e `forge-sdd status` (`/status`) para que o **Líder** apresente ao humano a visão consolidada da **Camada C**: **O que já foi feito (aprovado nas Camadas A e B)**, **Como está cada etapa da fila**, **O que ainda não foi desenvolvido**, **Exceções/ADRs** e **Propostas de Auto-Skills**.
  - **Validação:** Testes de validação do schema JSON da Camada B e testes de saída de `forge-sdd status`.

---

### Fase 4 — Motor de Auto-Skills (`DUP-05` / `GOV-03`), Importação de Pacotes e Fechamento

- [ ] **`feat-58-06` — Motor de Auto-Skills (`forge-sdd skills`) via Regra de Três (`DUP-05`) e Fim de Onda (`GOV-03`)**
  - **Escopo:**
    1. Implementar `internal/skills` e `forge-sdd skills` (`list`, `propose`, `approve`, `sync`) sobre o diretório canônico `.agents/skills/<skill-name>/SKILL.md`.
    2. Conectar o campo `candidate_skill` do parecer do Analista de Governança e a revisão de fim de onda (`/archive` / `/revisar`) para que, após confirmação humana no Líder, a skill seja sincronizada automaticamente para `.claude/skills/` e `.gemini/skills/`.
  - **Validação:** Testes unitários e de integração do ciclo `propose → approve → sync`.

- [ ] **`feat-58-07` — Importador de Pacotes de Governança (`manifest.yaml`), Migração (`update`), Documentação e Release `v3.0.0-beta`**
  - **Escopo:**
    1. Suportar a importação idempotente de pacotes de governança externos compatíveis com o formato `manifest.yaml` do `planning-brain-governanca` (modos `bloco`, `substituir`, `criar_se_ausente`, `mesclar`, `criar_uma_vez` + registro em `docs/governanca/VERSAO`).
    2. Atualizar `forge-sdd update` para migrar projetos existentes; atualizar `sdd/FLOW.md` e `README.md` (Regra 12); marcar `discovery-53` e `discovery-56` como `superseded by discovery-58`.
    3. Preparar e publicar a **`v3.0.0-beta`** (Regras 11 e 12): registrar os destaques reais em `sdd/releases/history.md` (com prévias de comandos CLI e identificação de agentes), atualizar versão em `sdd/.sddrc`, abrir o PR mantendo-o aberto sem mergear em `main` e criar a tag `v3.0.0-beta` diretamente na branch `feat/58-orquestracao-autonoma-governanca-skills`.
  - **Validação:** `go test ./...`, `go vet ./...`, `forge-sdd gate` e `forge-sdd doctor` passando 100%; PR aberto e tag `v3.0.0-beta` publicada na branch da feature.

---

## Janelas de Paralelismo (para a Fila do Builder)

```
feat-58-01 (Constitution IDs + Rules + Templates) ──┬──► feat-58-03 (Concepção + Reuse Check + Plan)
feat-58-02 (Papéis 5 Estações + Prompt Camada B)   ─┘
                                                    ├──► feat-58-04 (Fila Go + Camada A Gate) ──► feat-58-05 (Camada B JSON + Painel Líder)
                                                    └──► feat-58-06 (Motor Auto-Skills DUP-05) ─┴──► feat-58-07 (Importador manifest.yaml + Update + Docs)
```

---

## Mapeamento de Rules e IDs da Constituição por Task (Preview do Planejador)

| Task | Arquivos `.agents/rules/` carregados pelo Builder/Governança | IDs Principais Verificados |
|---|---|---|
| `feat-58-01` | `sdd-arquitetura.md`, `sdd-reuso-e-duplicacao.md` | `ARQ-01..06`, `DUP-01`, `DUP-10`, `GOV-01`, `GOV-04` |
| `feat-58-02` | `sdd-arquitetura.md`, `sdd-seguranca.md`, `sdd-reuso-e-duplicacao.md` | `ARQ-03`, `SEG-04`, `DUP-10`, `FLX-08` |
| `feat-58-03` | `sdd-reuso-e-duplicacao.md`, `sdd-fluxo-e-portao.md` | `DUP-02`, `SEG-07`, `FLX-01`, `ARQ-11` |
| `feat-58-04` | `sdd-arquitetura.md`, `sdd-codigo-e-testes.md`, `sdd-fluxo-e-portao.md` | `P-01..P-23`, `COD-02`, `TST-01..05`, `GOV-02`, `GOV-04` |
| `feat-58-05` | `sdd-seguranca.md`, `sdd-codigo-e-testes.md`, `sdd-fluxo-e-portao.md` | `S-01..S-07`, `FLX-03`, `FLX-08`, `SEG-04..05` |
| `feat-58-06` | `sdd-arquitetura.md`, `sdd-reuso-e-duplicacao.md`, `sdd-codigo-e-testes.md` | `DUP-05`, `DUP-10`, `GOV-03`, `TST-01` |
| `feat-58-07` | `sdd-arquitetura.md`, `sdd-fluxo-e-portao.md` | `GOV-01`, `GOV-04`, `FLX-06..08`, Regra 12 |
