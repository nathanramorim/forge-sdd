# Plano de Discovery 59 — Módulo Visual Plugável (`forge-sdd studio`) e Filas Idempotentes por Estação

Roadmap técnico e quebra em subtarefas para o `/split-features` dentro da subpasta única `sdd/features/feat-59-modulo-visual-estacoes-personagens/` (Regra 13), operando sobre uma **única branch** (`feat/59-modulo-visual-estacoes-personagens`, conforme Regra 15 e `FLX-02`).

Respeita a **Rule of 7** (4 subtarefas ordenadas de baixo para cima), incorpora as decisões aprovadas no Portão `FLX-01` (Opção A1: subcomando opcional `forge-sdd studio` via `embed.FS` + `SSE`; Opção A2: filas dedicadas por estação `sdd/.queue/stations/<station>.json` com `idempotency_key` determinístico e escrita atômica) e reutiliza a skill canônica `.agents/skills/novo-subcomando-cli-sdd/SKILL.md` (`DUP-05`).

---

## 1. Quebra de Tasks / Subfeatures (`P / M / G`)

### Fase 1 — Fundação de Domínio: Filas por Estação e Motor de Idempotência Estrita

- [ ] **`feat-59-01` — Motor de Filas por Estação (`sdd/.queue/stations/*.json`), Idempotência Estrita (`IdempotencyKey`) e Escrita Atômica** · **Tamanho: `M`**
  - **Depende de:** —
  - **Escopo:**
    1. Estender `internal/queue` (`DUP-01` — dono do conceito de Fila em `sdd/spec/concepts.md`) para gerenciar as **5 filas canônicas por estação**:
       - `sdd/.queue/stations/lider.json` (`01. O Líder`)
       - `sdd/.queue/stations/pm-architect.json` (`02. PM & Arquiteto`)
       - `sdd/.queue/stations/planner.json` (`03. O Planejador`)
       - `sdd/.queue/stations/builder.json` (`04. Fila de Builders`)
       - `sdd/.queue/stations/governance.json` (`05. Analista de Governança`)
    2. Implementar o motor de **Idempotência Estrita**:
       - Cálculo determinístico de `idempotency_key` via `crypto/sha256` (`station + ":" + item_id + ":" + operation + ":" + artifact_path + ":" + attempt`) quando não informada explicitamente.
       - Janela de `processed_keys` e verificação de estado-alvo: chamadas repetidas de `Enqueue`, `Claim`, `Complete`, `Review` ou `Handoff` com a mesma chave retornam `IdempotentHit = true` sem duplicar itens e sem incrementar `version`.
    3. Implementar **escrita atômica (`writeFileAtomic`)** via arquivo temporário no mesmo diretório + `os.Rename` e incremento monotônico de `version`.
    4. Garantir **sincronia bidirecional transparente** entre `sdd/.queue/queue.json` e `sdd/.queue/stations/builder.json` + `governance.json` para manter 100% de retrocompatibilidade com a Feature 58.
    5. Expor subcomandos na CLI `cmd/forge-sdd/queue.go` (`forge-sdd queue handoff`, `forge-sdd queue stations`) com flag opcional `--idempotency-key`.
  - **Skills Reutilizadas:** `novo-subcomando-cli-sdd`
  - **Critério Executável:** `go test -race ./internal/queue/... ./cmd/forge-sdd/... && go vet ./...` (cobrindo `CA-02`, `CA-03`, `CA-04`).

---

### Fase 2 — Backend Observador Passivo (`internal/studio`) e Stream SSE Read-Only

- [ ] **`feat-59-02` — Agregador Read-Only de Estado (`internal/studio`) e Servidor HTTP/SSE (`forge-sdd studio`)** · **Tamanho: `M`**
  - **Depende de:** `feat-59-01`
  - **Escopo:**
    1. Criar o pacote `internal/studio` (modelo `enxuto` — `ARQ-11`, estritamente Read-Only `SEG-02`/`SEG-06`) que lê passivamente `sdd/.queue/stations/*.json`, `sdd/.queue/queue.json`, `sdd/memory/progress.md`, `sdd/discovery/*` e `.agents/skills/*`.
    2. Consolidar o DTO determinístico `StudioSnapshot` (`GET /api/state`) com as 5 estações (`id`, `role`, `color`, `mark`, `state`, `badge_label`, `current_focus`, `input_contract`, `output_contract`, `items[]`, `version`), o `active_station`, o último `handoff` (incluindo detecção do loop `05 ↺ 04` em `changes_requested`), o `governance_verdict` e as `skills[]`.
    3. Implementar o endpoint `GET /api/events` (**Server-Sent Events — SSE**) que monitora o vetor de `version` das 5 filas (com resiliência a leituras concorrentes e zero escrita em disco) e emite eventos `event: state` apenas quando o hash/versão mudar (evitando emissões duplicadas).
    4. Aplicar bloqueio de segurança (`SEG-01`): bind padrão restrito a `127.0.0.1` e rejeição com `405 Method Not Allowed` para qualquer método HTTP diferente de `GET`/`HEAD`.
    5. Criar o adaptador CLI `cmd/forge-sdd/studio.go` (`forge-sdd studio --port 4321 --host 127.0.0.1 --dir .`) seguindo a skill `novo-subcomando-cli-sdd`.
  - **Skills Reutilizadas:** `novo-subcomando-cli-sdd`
  - **Critério Executável:** `go test -race ./internal/studio/... ./cmd/forge-sdd/... && go vet ./...` (cobrindo `CA-01`, `CA-07`, `CA-09`).

---

### Fase 3 — Interface Visual 3D Embarcada (`embed.FS`): 5 Personagens `dock-ia` + Filas + Governança

- [ ] **`feat-59-03` — Interface Visual Embarcada (`embed.FS`) com os 5 Personagens 3D (`dock-ia` + `Three.js`), Trilho de Handoff, Filas Individuais e Inspetor de Governança** · **Tamanho: `M`**
  - **Depende de:** `feat-59-02`
  - **Escopo:**
    1. Embutir em `internal/studio/assets/index.html` (via `embed.FS` — `ARQ-03`) a interface visual evoluída a partir de `docs/showcase-v3-lp.html` com os tokens oficiais do `dock-ia` (`--bg: #0d0f14`, `--bg2: #12151d`, `--bg3: #1a1e29`, `Montserrat` + `JetBrains Mono`).
    2. Renderizar os **5 personagens 3D compactos em Three.js** (`buildCompactBot` com fallback SVG 2D automático caso WebGL não esteja disponível):
       - `01. O Líder` (`#f2a65a`, marca `ring`)
       - `02. PM & Arquiteto` (`#8aa4ff`, marca `diamond`)
       - `03. O Planejador` (`#c9a7ff`, marca `bar`)
       - `04. Fila de Builders` (`#4fd1b5`, marca `wave`)
       - `05. Analista de Governança` (`#ff8a5c`, marca `tri`)
    3. Exibir em cada um dos 5 cards:
       - O mascote 3D reagindo ao `mascot_state` (`idle`, `working`, `thinking`, `searching`, `approval`, `finished`, `error`);
       - Os itens reais da **fila individual daquela estação** (`sdd/.queue/stations/<station>.json`) com indicador de `version` e `idempotency_key` curta;
       - Os contratos explícitos **`Recebe (Entrada)`** e **`Entrega (Saída)`**.
    4. Conectar o cliente `EventSource('/api/events')` e fallback `/api/state` para atualização em tempo real (`LIVE`), além de manter o seletor de **Modo Simulação (`DEMO`)** para demonstração interativa quando o repositório estiver parado.
    5. Renderizar o **Inspetor de Governança (Camadas A, B e C)** exibindo os achados semânticos (`S-01..S-07`), evidências e Auto-Skills propostas.
  - **Skills Reutilizadas:** —
  - **Critério Executável:** `go test ./internal/studio/... && go vet ./...` validando presença de todos os 5 personagens, marcas (`ring`, `diamond`, `bar`, `wave`, `tri`), contratos de entrada/saída, conexão `/api/events` e ativos embutidos via `embed.FS` (`CA-05`, `CA-06`, `CA-08`).

---

### Fase 4 — Integração nos Papéis/Comandos, Catálogo de Conceitos e Documentação

- [ ] **`feat-59-04` — Integração de Handoff nas 5 Estações (`.agents/`), Catálogo de Conceitos (`concepts.md`) e Documentação (`FLOW.md` / `README.md` / `history.md`)** · **Tamanho: `P`**
  - **Depende de:** `feat-59-01`, `feat-59-03`
  - **Escopo:**
    1. Atualizar `sdd/spec/concepts.md` (`DUP-01`) registrando a extensão de `internal/queue` (Filas por Estação com Idempotência) e o novo conceito `Módulo Visual Plugável (Forge Studio)` em `internal/studio`.
    2. Atualizar as instruções de handoff em `.agents/roles/` e `.agents/commands/` (e seus espelhos em `internal/scaffold/templates/`) para documentar as filas por estação (`sdd/.queue/stations/<station>.json`) e o comando opcional `forge-sdd studio`.
    3. Atualizar `sdd/FLOW.md`, `README.md` e `sdd/releases/history.md` conforme a Regra 12 (`FLX-07`) com prévias de `forge-sdd studio` e `forge-sdd queue handoff`.
    4. Executar validação completa da Camada A (`forge-sdd gate` e `forge-sdd doctor`).
  - **Skills Reutilizadas:** `novo-subcomando-cli-sdd`
  - **Critério Executável:** `go test ./... && go vet ./... && go run ./cmd/forge-sdd gate` com 100% de aprovação.

---

## 2. Janelas de Paralelismo e Dependências (para a Fila do Builder)

```
feat-59-01 (5 Filas por Estação + IdempotencyKey + Escrita Atômica)
    │
    ▼
feat-59-02 (Agregador Read-Only internal/studio + Servidor HTTP/SSE)
    │
    ▼
feat-59-03 (SPA 3D Embarcada embed.FS: 5 Personagens dock-ia + Filas + Inspetor)
    │
    ▼
feat-59-04 (Catálogo concepts.md + Templates .agents/ + Docs Regra 12 + Gate)
```

*(Nota de paralelismo: `feat-59-01` é a base das 5 filas; assim que `feat-59-02` definir o contrato JSON `/api/state`, `feat-59-03` e a preparação documental de `feat-59-04` podem avançar rapidamente em sequência limpa sem conflito de arquivos).*

---

## 3. Mapeamento de Rules, IDs da Constituição e Skills por Task

| Task | Tamanho | Arquivos `.agents/rules/` carregados pelo Builder e Governança | IDs da Constituição Verificados | Skills Reutilizadas (`.agents/skills/`) |
| --- | --- | --- | --- | --- |
| **`feat-59-01`** | `M` | `sdd-arquitetura.md`, `sdd-seguranca.md`, `sdd-reuso-e-duplicacao.md`, `sdd-codigo-e-testes.md` | `ARQ-01`, `ARQ-02`, `ARQ-06`, `ARQ-10`, `ARQ-11`, `SEG-01`, `SEG-02`, `DUP-01..04`, `COD-01..03`, `TST-01..05` | `novo-subcomando-cli-sdd` |
| **`feat-59-02`** | `M` | `sdd-arquitetura.md`, `sdd-seguranca.md`, `sdd-codigo-e-testes.md` | `ARQ-01`, `ARQ-03`, `ARQ-04`, `ARQ-06`, `ARQ-11`, `SEG-01`, `SEG-02`, `SEG-05`, `SEG-06`, `COD-01..03`, `TST-01..05` | `novo-subcomando-cli-sdd` |
| **`feat-59-03`** | `M` | `sdd-arquitetura.md`, `sdd-reuso-e-duplicacao.md`, `sdd-codigo-e-testes.md` | `ARQ-03`, `ARQ-04`, `DUP-02`, `COD-01`, `COD-05`, `TST-01`, `TST-04`, `TST-05` | — |
| **`feat-59-04`** | `P` | `sdd-reuso-e-duplicacao.md`, `sdd-codigo-e-testes.md`, `sdd-fluxo-e-portao.md` | `DUP-01`, `DUP-10`, `TST-02`, `TST-05`, `FLX-01..07`, `GOV-04`, Regra 12 | `novo-subcomando-cli-sdd` |

---

## 4. Handoff para o Líder (`/split-features`)

O plano técnico (`plan-59`) está completo, alinhado às decisões do usuário no Portão `FLX-01` e mapeado contra a Constituição e `.agents/rules/`.
Próxima ação do **Líder**: executar `/split-features` para gerar os arquivos de especificação em `sdd/features/feat-59-modulo-visual-estacoes-personagens/feat-59-01..04.md` e enfileirá-los para os Builders.
