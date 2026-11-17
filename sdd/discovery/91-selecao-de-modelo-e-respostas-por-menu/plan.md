# Plano — Iniciativa 91 (de baixo para cima)

## Onda 1: Menu para responder e lista de modelos por agente
- **91-01 ADR-0016** — níveis e tabela por agente, roteamento e escalada, um job por task, pergunta por menu e ferramentas dos agentes; estende as ADRs 0003, 0004, 0005 e 0014. Rules: FLX-05, GOV-01. Deps: —. CA: arquivo presente + `forge gate`.
- **91-02 Pergunta por menu e skills do Líder** — seção "Como perguntar" em `clarify.md`; ponteiro nas instruções raiz; ferramenta de pergunta em `lider`, `pm-architect` e `planner` e a de subagente no `lider` (nomes conferidos na documentação oficial); skills do Líder citam `--fix`, `--base` e a rubrica do nível; goldens. Rules: DUP-06, DUP-10, SEG-06, TST-02. Deps: —. CA: CA-01, CA-02.
- **91-03 Catálogo de modelos por agente e Settings** — catálogo por CLI (`claude` com `fable`, `agy models` com limite de tempo, `gemini`, reserva), "Outro (digitar)" com `Input` no menu, limpeza ao trocar o agente da sessão, aviso de sessão sem modelo fixo. Rules: SEG-01, SEG-03, ARQ-01, DUP-01, TST-04. Deps: —. CA: CA-03, CA-04.
- **91-04 Níveis e tabela por agente** — `model_selection` e `tiers` no `forge.json`, validação, "Seleção de modelo" e passo a passo dos três níveis no Settings. Rules: SEG-01, ARQ-01, DUP-04, TST-02. Deps: 91-03. CA: CA-05.

## Onda 2: Modelo por task na esteira
- **91-05 Nível na spec** — `**Nível:**` no leitor do cabeçalho, checagem antes do build, `forge context` com nível e modelo. Rules: FLX-01, DUP-01, DUP-06, TST-03. Deps: 91-01. CA: CA-06.
- **91-06 Roteamento por job** — resolvedor único (job, nível, sessão, CLI), campos opcionais no job, worker `headless` aplica, aviso no modo `live` e em job com várias tasks. Rules: ARQ-01, DUP-01, SEG-03, COD-03, TST-03. Deps: 91-04, 91-05. CA: CA-07.
- **91-07 Um job por task no `dispatch`** — `builder.job_per_task` e chave no Settings, reusando o motor do `/toca-ai`; todos os jobs na branch e no PR da iniciativa, inclusive no `/toca-ai`, que hoje cria uma branch por etapa. Rules: FLX-02, DUP-03, ARQ-01, TST-03. Deps: 91-01. CA: CA-08.
- **91-08 Escalada no retrabalho** — `pr_rework` sobe um nível em `auto`, com registro no evento. Rules: ARQ-01, COD-03, TST-03. Deps: 91-06. CA: CA-09.

## Onda 3: Medir e levar ao chat
- **91-09 Relatório por nível e modelo** — `nivel` e `tentativa` na telemetria, `forge report --por-modelo`, sugestão com amostra mínima. Rules: ARQ-01, SEG-05, FLX-04, TST-04. Deps: 91-06. CA: CA-10.
- **91-10 Modelo por comando e por papel no chat** — `model:` nos adaptadores de comando e nas definições de agente do Claude, renderizados da tabela; começa conferindo o suporte na documentação oficial e, sem ele, entrega só por papel. Rules: DUP-10, ARQ-03, TST-02. Deps: 91-02, 91-04. CA: CA-11.

## Fechamento (Líder)
CA-12; validação em sessão nova (menus) e comparação com a 90 no `status.md`.

Paralelismo: na onda 1, 91-01, 91-02 e 91-03 tocam arquivos diferentes (docs, templates, `internal/config`). Na onda 2, 91-05 e 91-07 podem ir juntas; 91-06 espera as duas bases (níveis e nível na spec). Na onda 3, as duas são independentes.

Ordem de entrega recomendada: publicar a beta com a iniciativa 90 antes do build desta, para a base do PR sair certa e a 91 ser a primeira medida com números reais.
