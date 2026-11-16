# Plano — Iniciativa 88 (de baixo para cima)

## Onda 1: Fundação
- **88-01 Config `autonomy`** — `ForgeConfig.Autonomy`, validação, default `normal`, `advanced` ⇒ `permissions: full`, `isolated_confirmed`. Rules: ARQ-01, SEG-01, COD-03. Deps: —. CA: `go test ./internal/config -run TestAutonomy`
- **88-02 Gesto no menu** — tecla →/← em "Iniciar esteira (up)", confirmação de ambiente isolado, linha `avançado`; fallback `forge pipeline autonomy <modo>`. Rules: ARQ-01, ARQ-06, DUP-01. Deps: 88-01. CA: `go test ./cmd/forge-sdd -run 'TestMenu_(Autonomy|UpGesto)'`
- **88-03 ADR-0013** — Reviewer como revisor da spec no modo avançado (altera a leitura de FLX-01 só nesse modo). Rules: FLX-05, GOV-01. Deps: —. CA: arquivo presente + `forge gate`.

## Onda 2: Portões
- **88-04 Job `spec_review` e parecer** — tipo de job, `ApplySpecVerdict`, `review-spec.md` escrito pelo Go, retorno à Discovery com limite. Rules: ARQ-01, SEG-06, DUP-01, TST-01. Deps: 88-01. CA: `go test ./internal/bus ./internal/pipeline -run 'TestSpecReview|TestApplySpecVerdict'`
- **88-05 Checklist do Reviewer (spec e entrega)** — prompt e skill com Constitution, rules, ADRs vigentes, `O que já existe`, critérios; exige `forge gate` verde. Rules: FLX-06, FLX-08, GOV-04. Deps: 88-04. CA: `go test ./internal/pipeline -run TestReviewPrompt`
- **88-06 `blocked` com continuação** — fila segue nas independentes; inbox com motivo. Rules: ARQ-01, COD-03. Deps: 88-04. CA: `go test ./internal/queue -run TestBlocked`

## Onda 3: Segurança, docs e Settings
- **88-07 Salvaguardas e docs** — tetos por job/dia, STOP, `doctor`, LEIA-ME e `FLOW.md`. Rules: SEG-05, COD-01. Deps: 88-02, 88-06. CA: `go test ./internal/pipeline -run TestSalvaguardas && go test ./...`
- **88-08 Settings e default do card** — `RequireExistingCard` default `false`; pergunta no `pipeline init`; menu "Settings" com chave liga/desliga (card); remove o item de `configMenu`; atualiza `docs/pipeline.md` e os golden. Rules: ARQ-01, ARQ-06, DUP-01, COD-03, TST-01. Deps: — (independente; pode ir primeiro). CA: `go test ./internal/config ./cmd/forge-sdd -run 'TestDefaultCard|TestPipelineInit_PerguntaCard|TestMenu_Settings'`

Paralelismo: 88-08 é independente das demais e pode ser a primeira entrega; 88-01 e 88-03 em paralelo; 88-02 e 88-04 em paralelo após 88-01; 88-05 e 88-06 em paralelo após 88-04.
