# Plano — Iniciativa 90 (de baixo para cima)

## Onda 1: Medir e destravar (as quatro em paralelo)
- **90-01 ADR-0014** — medição real, PR aberto pelo Go, base da iniciativa, archive no fechamento, modelo e esforço por sessão; substitui D-1 e D-4 da ADR-0009 e estende as ADRs 0004 e 0005. Rules: FLX-05, GOV-01. Deps: —. CA: arquivo presente + `forge gate`.
- **90-02 Telemetria real por job** — lê o `result` no fim do job (`internal/monitor`), campos novos em `SessionMetrics`, versão do binário, `forge report --jobs` e `--initiative`, painel com o dado real; tira `estimateJobTokens` de `cmd/`. Rules: ARQ-01, DUP-04, SEG-05, COD-03, COD-04, TST-04. Deps: —. CA: CA-01 a CA-03.
- **90-03 Um leitor de dependências e checagem das specs** — `queue` vira o dono do leitor (aceita `feat-`, `fix-` e `88-01`), `waves` reusa, fila antiga se corrige na leitura; checagem por task no `dispatch` e no `/toca-ai`. Rules: DUP-01, DUP-06, TST-03. Deps: —. CA: CA-09, CA-10.
- **90-04 `forge status` curto para agentes** — "Ao Vivo" pelo `ChatFeedFormatter`; sem terminal, curto por padrão; `--full` para tudo; testes que esperam a saída completa passam a usar `--full`. Rules: ARQ-01, ARQ-10, COD-01, TST-02. Deps: —. CA: CA-04.

## Onda 2: Entrega, archive e modelo
- **90-05 Base da iniciativa e PR pelo Go** — `base_branch` na fila e resolvedor único (usado por `dispatch` e `/toca-ai`); `forge pipeline pr`; a skill do Builder e o `BuildPrompt` trocam o passo do PR pelo comando. Rules: FLX-02, FLX-03, FLX-07, SEG-01, SEG-03, DUP-01, ARQ-06, TST-04. Deps: 90-01. CA: CA-05 a CA-07.
- **90-06 Archive que move, no fechamento** — move idempotente, guarda de branch e de pendências, totais reais; ao aprovar o último job, arquiva sozinho só com `origem: toca-ai` na fila (gravada pelo motor da 89) e avisa "pronta para arquivar" nos demais casos; `--concluidas --dry-run`. Rules: DUP-01, SEG-02, FLX-04, COD-04, TST-03. Deps: 90-01, 90-02. CA: CA-11.
- **90-07 Modelo e esforço por sessão (absorve a 85)** — `models` e `efforts` no `forge.json`, marcadores `{model}` e `{effort}`, tabela por CLI (claude conferido; as demais confirmadas com `--help` no build), Settings (oferece incluir o marcador em perfil gravado sem ele), aviso no preflight. Rules: SEG-01, SEG-03, ARQ-01, DUP-01, TST-02. Deps: 90-02. CA: CA-12.

## Onda 3: Fix na esteira
- **90-08 Fix no `dispatch`** — `forge dispatch <iniciativa> --fix <id>`, com a checagem da 90-03 e a base da 90-05. Rules: FLX-01, FLX-02, DUP-03, TST-03. Deps: 90-03, 90-05. CA: CA-08.

## Fechamento (Líder)
CA-13; comparação com a linha de base no `status.md` (RN-1); release beta quando a pessoa pedir.

Paralelismo: na onda 1 as quatro tocam pacotes diferentes (docs, `monitor`, `queue`, status). Na onda 2, a 90-05 e a 90-07 mexem em arquivos diferentes de `internal/config` e podem ir juntas; a 90-06 espera a 90-02 para gravar os totais reais.
