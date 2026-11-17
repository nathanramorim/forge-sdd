# Plano — Iniciativa 93 (de baixo para cima)

## Onda 1: Medir e decidir
- **93-01 ADR-0018** — coleta de contexto: medida real, meta e teto, ganchos do Claude Code, cartão de retomada e teto nos jobs; estende as ADRs 0009 (selo e handoff) e 0014 (uso real). Rules: FLX-05, GOV-01. Deps: —. CA: arquivo presente + `forge gate`.
- **93-02 Medidor de contexto real** — leitor do transcrito (só contadores, pelo fim do arquivo) e do log do job (pico e coletas) em `internal/monitor`; campos na telemetria e colunas em `forge report --jobs`. Rules: ARQ-01, SEG-01, SEG-05, DUP-01, TST-04. Deps: —. CA: CA-01, CA-02.
- **93-03 Decisão de coleta e `forge session gc`** — função pura (meta, teto, ponto seguro, limite de avisos), marca de fechamento no `forge session record`, modos `--hook` e `--resume`, cartão de retomada. Rules: ARQ-01, ARQ-06, DUP-03, COD-03, TST-03. Deps: 93-02. CA: CA-03, CA-04, CA-06.

## Onda 2: Ligar no Claude Code, nos jobs e no menu
- **93-04 Ganchos, teto e instruções no Claude Code** — começa conferindo na documentação oficial quais ajustes são públicos (teto em tokens ou percentual, compactação em ociosidade); mescla ganchos e ajuste em `.claude/settings.json`; instruções de compactação na instrução raiz; goldens. Rules: SEG-02, SEG-07, DUP-10, TST-02. Deps: 93-03. CA: CA-05, CA-07.
- **93-05 Teto nos jobs** — o worker passa o teto no ambiente do agente quando a CLI aceita; sem mudar o argv; modo `live` segue com o reset por job. Rules: ARQ-01, SEG-03, COD-03, TST-03. Deps: 93-02. CA: CA-08.
- **93-06 Settings, selo e documentação** — chave "Coleta de contexto", meta e teto; selo do painel com o contexto real; `docs/pipeline.md` com a coleta e a seção "Gancho ou habilidade?". Rules: ARQ-01, SEG-01, TST-02. Deps: 93-03, 93-04. CA: CA-09 e a parte de docs da CA-11.
- **93-07 Subagente por padrão no trabalho pesado do Líder** — skills de concepção e de planejamento, só com a coleta ligada. Rules: DUP-10, COD-01, TST-02. Deps: 93-04 e a 91-02 (ferramenta de subagente do Líder). CA: CA-10.

## Fechamento (Líder)
CA-11; uma sessão real de várias horas com a coleta ligada e a comparação no `status.md`.

Paralelismo: na onda 1, 93-01 e 93-02 são independentes. Na onda 2, 93-04 e 93-05 tocam pacotes diferentes (`internal/scaffold` e `internal/pipeline`) e podem ir juntas; 93-06 vem depois da 93-04; 93-07 espera a 91-02.

Ordem de entrega recomendada: esta iniciativa não depende da 91 nem da 92, a não ser na 93-07. Pode ir logo depois da beta com a 90.
