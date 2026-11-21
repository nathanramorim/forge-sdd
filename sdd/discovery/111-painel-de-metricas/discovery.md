# Discovery — 111 Painel de métricas e aprendizado (módulo da 109)

## Pedido do dono (20/11/2026, chat)
Módulo instalável, no mesmo modelo da 109 (repositório separado, privado por padrão, instalação por contrato), que acompanha o trabalho e as métricas. MVP: um **painel em HTML** (HTML semântico, CSS e JavaScript em `assets/`) que lê dados e relatórios e mostra indicadores em listas com paginação, no visual do design system do Forge SDD. Indicadores: (a) comandos **depreciados** ainda em uso; (b) **erros e retornos** (itens devolvidos ao Builder, voltas à Discovery, nova-feature, novo-fix, vereditos do Reviewer e da Governance); (c) **consumo de tokens** por papel, agente, modelo, ferramenta e iniciativa; (d) o **log de achados de conformidade** (105-08). Norte de longo prazo: entender o que os agentes erram e **ensiná-los a não repetir**, com o dono aprovando. Nada foi criado: nem repositório, nem código.

## ID
Conferido em 20/11/2026: `ls sdd/discovery | grep -E '^1[01]'` na main traz 100, 101, 102 e 104; `git branch -r` mostra 105 a 109 e **110** (`feat/110-estimativa-de-tokens`, PR #155, em escrita); nenhuma 111 em branch, worktree ou PR. O ID é **111**.

## Descrição de negócio (obrigatória em toda spec nova)
Hoje o dono só sabe onde a esteira erra lendo arquivos um a um. O painel junta, num lugar só e sem sair da máquina, o que foi devolvido, por qual regra, quantas vezes a mesma regra falhou, quais comandos depreciados ainda são usados e quanto se gastou. O ganho é duplo: **menos retrabalho e menos token** (o erro que se repete vira lição aprovada pelo dono, e o comando depreciado com substituto mais barato é achado e trocado) e **previsibilidade** (o dono vê a tendência, com o tamanho da amostra sempre à mostra). Regra de honestidade: o painel nunca finge precisão; mostra `n`, a origem do dado e o rótulo `real` ou `estimado`.

## Clarify
Sem pergunta antes de escrever: o pedido traz forma, fronteira e restrições. As lacunas viraram `## Decisões para o Gate 1`.

## 1. Que dados existem hoje (medido em `origin/main` 8bd3a0a, neste worktree; nada executado além de leitura)

| Fonte | Caminho e formato | Quem escreve | Volume REAL agora | Estado |
|---|---|---|---|---|
| Telemetria de sessão | `sdd/.metrics/session-*.json`, esquema `forge-sdd/metrics/1.0` (`internal/monitor/session_record.go`): `feature`, `phase`, `agent_path`, `tokens_input/output`, `turns`, `duration_seconds`, `model`, `outcome`, `rework_lines`, `files_touched`, `sdd_version`, opcionais `tokens_cache_*`, `cost_usd_tabela`, `medicao` | `forge session record` (agente chama no fim da sessão) | **97 sessões**; 9 com tokens diferentes de zero (soma 162.530 de entrada e 50.788 de saída); `model` vazio em 63; `agent_path` vazio em 35; `outcome`: 87 `approved`, 9 `blocked`, 1 `rejected`; `medicao` presente em 1 (`estimado`) | na main |
| Vereditos | `sdd/discovery/*/verdicts/*.json`: `veredito`, `achados[]`, `tentativa`, `head_sha`, `revisado_em`, `task_id`, `candidate_skill` | pipeline (`ApplyVerdict`), ADR-0012 | **17 arquivos** (3 de `rework`): 12 `conforme`, 3 `violacao`, 2 `duvida`; **19 achados**; regra mais citada `TST-01` (5), depois `DUP-04`, `COD-02`, `ARQ-11`, `DUP-01` (2 cada) | na main |
| Fila | `sdd/discovery/*/queue.json`: `items[]` com `status`, `attempts`, `governance_verdict`, `rules`, `skills_used` | `forge queue` | **21 arquivos, 90 itens**: 70 `approved`, 14 `completed`, 3 `review_pending`, 3 `queued`; `attempts`: 52 com 1, 38 com 0. **Nenhum item em `changes_requested` agora** (o estado é transitório; o histórico dele não é guardado) | na main; a 106/0.7 tira o status do `queue.json` |
| Specs e correções | `sdd/discovery/*/features/*.md`, `fixes/*.md`, `status.md` | Discovery e Líder | **125 specs de feature, 15 de fix, 22 `status.md`** | na main |
| Lições | `sdd/memory/lessons.md` (comando `forge lessons add`) | Líder, por fix aprovado | **2 lições** | na main; a 106 (RN-8, decisão v) migra para `sdd/memory/licoes/<id>.md` |
| Uso por papel | `internal/monitor/tokens.go` (`TokenMetricsPerRole`, `NormalizeRole`), `usage.go` (`ParseJobUsage` lê o stream-json do Claude), `forge report` | monitor | logs de job em `.forge/runtime/logs`, **fora do git** (a 110 mediu 46 logs, 25 com `result` real: 12 builder, 13 reviewer, 0 discovery; não refiz) | na main |
| Depreciados | `sdd/memory/logs/depreciados.csv` (`data_utc, comando, contexto, agente, substituto, desde, versao_forge`), catálogo em `internal/deprecation` | `avisoDepreciado` no `PreRun` | **0 linhas na main**: o pacote nem existe na main, só na branch `feat/105-07-registro-de-depreciados`. O único arquivo que achei é artefato de teste em pasta temporária; **não conta** | só especificado (105-07) |
| Achados de conformidade | `sdd/memory/logs/achados.csv` (`data_utc, ponto, iniciativa, task, job, rule_id, severidade, classificacao, desfecho, tentativa, arquivo, resumo, versao_forge`), `internal/csvlog`, `achados.Registrar` | pipeline e `forge queue review` | **0 linhas e o pacote não existe**: só no `feat-105-08` (branch da 105) | só especificado (105-08) |
| Estado por feature | `sdd/discovery/<id>/state/<sid>.json` | `queue.Claim/Review` (106/0.7) | **0** | só especificado (106); o esquema nem está definido |
| Bloco `spec` da sessão da Discovery | `internal/monitor/spec_metrics.go` (tipo, bytes, seções, `features_count`) | `session record --sessao spec` | **0** na main | só na branch do PR #131 (aberto), segundo a 110 |

**Já na main:** sessões, vereditos, fila, specs, lições, uso por papel. **Só especificado:** depreciados (105-07), achados (105-08), estado por feature (106/0.7), bloco `spec`.

### Lacunas (o que não é medido)
| Pergunta do dono | Por que não responde hoje |
|---|---|
| Devolvidos ao Builder, **por motivo** | só `attempts` e o veredito; sem `achados.csv` não há regra por devolução, e `changes_requested` não deixa histórico |
| Voltas à Discovery, `nova-feature`, `novo-fix` | **nenhum registro**: só se infere de commits e arquivos `fixes/*.md` (15), de forma frágil |
| Tokens por papel, modelo, ferramenta | 88 de 97 sessões sem token; modelo vazio em 63; **não existe campo de ferramenta** (a 110 propõe `tool`); 0 sessões de Discovery com token |
| Média de tokens | com n = 9 sessões e variação de 28 vezes no builder (medida pela 110), a média é **estatisticamente vazia**: o painel mostra mediana, faixa e `n`, nunca só a média |
| Economia dos depreciados | sem uso registrado (0 linhas) não há o que cruzar; o método existe, o dado ainda não |

## Reuse Check (DUP-02)
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| "painel" | `sdd/spec/adr/adr-0011`, `internal/pipeline/status_live.go`, `survey.QueueDrilldownTUI`, discovery 84 | painel **de terminal** (`forge status --live`) já tem dono | **Não usar o nome `forge painel`**; comando do módulo é `forge metricas` | `DUP-01`: um conceito, um dono |
| Relatório de tokens | `cmd/forge-sdd/report.go` (`forge report`, `--initiative`, `--jobs`), `internal/monitor/tokens.go` | agrega `session-*.json` por item no terminal | **Reusar o esquema e a semântica de papel** (`NormalizeRole`: builder, reviewer, discovery), **não o código** | Go proíbe importar `internal/` de outro repositório |
| Lições | `cmd/forge-sdd/lessons.go` (`forge lessons add`), 106 RN-8 (`licoes/<id>.md`) | gravação determinística e orçada | **O módulo propõe; o core grava** | RN-7 da 109: o módulo não escreve arquivo do core |
| Repetição por regra | feat-105-08 (`achados.Consultar`, `forge context --rule`) | só especificado; repetição **recontada na leitura**, N = 3 proposto | **Reusar a mesma regra de contagem** | evita dois critérios de "reincidência" |
| Estimativa de tokens | discovery 110 (`forge estimar`, mediana e faixa, histórico estimativa x real) | PR #155 | **Fronteira abaixo**; sem duplicar a estatística | |
| Contrato de módulos | discovery 109 (manifesto, eventos, `doctor`, instantâneo) | PR #154 | **Reusar inteiro, sem editar** | |
| Tokens de design | `docs/index.html:14-36` (`:root`), `docs/wiki.css:3-14` | o Forge SDD **não tem arquivo de tokens próprio**: `docs/design/` só contém a Edição Missões (temporária, **proibida aqui**) | **Copiar os valores de `docs/index.html:14-36` para o `tokens.css` do módulo** | `wiki.css` diverge (`--bg3` `#1a1f2e`, `--text` no lugar de `--fg`); vale a LP |
| Teste que lê HTML | `internal/scaffold/pipeline_scaffold_test.go:315-333` lê `docs/index.html` e `docs/wiki.html` | o projeto testa HTML por **leitura estática em Go**; a CI roda só `go test ./...` e não tem Node | **Mesmo padrão** no módulo | não inventar ferramenta |
| Contagem de tokens a partir do veredito | nada | — | **Criar no módulo** | |

## 2. Contrato de dados (v1, proposto)
**Princípio:** o painel não lê formato bruto espalhado; um **leitor versionado** (`internal/fonte` do módulo) converte cada fonte em registros de um esquema estável, e o `forge metricas gerar` grava `data/*.json` com `"esquema": "forge-metricas/1"`.

| Registro (JSON Lines no histórico do módulo; JSON no `data/`) | Campos mínimos |
|---|---|
| `sessao` | `quando`, `iniciativa`, `feature`, `papel`, `agente`, `modelo`, `ferramenta`, `tokens_in`, `tokens_out`, `medicao` (`real`, `estimado`, `desconhecido`), `resultado` |
| `retorno` | `quando`, `iniciativa`, `feature`, `ponto` (`spec` ou `entrega`), `tentativa`, `desfecho` (`rework`, `ajustar`, `escalado`, `registrado`), `regra_id`, `severidade` |
| `volta` | `quando`, `iniciativa`, `tipo` (`discovery`, `nova-feature`, `novo-fix`), `origem` (`evento` ou `inferido`) |
| `depreciado` | `quando`, `comando`, `agente`, `substituto`, `desde` (as colunas do 105-07) |
| `licao` | `id`, `regra_id`, `texto`, `versao`, `estado` (`proposta`, `aprovada`, `recusada`, `revertida`), `desde` |

**Compatibilidade:** só aditivo dentro de `forge-metricas/1`; remover ou renomear exige `/2`; campo desconhecido é ignorado nos dois sentidos; **linha ruim é contada e mostrada** ("3 linhas ilegíveis"), nunca descartada em silêncio. O leitor de cada fonte do core é um adaptador com **versão do layout** (o `queue.json` de hoje e o `state/<sid>.json` da 106/0.7), rotulado frágil.

**Core (mínimo, aditivo, só com módulo ativo; byte a byte sem módulo, RN-1 da 109):** nada na v0. Para a v1: o evento `sessao.registrada` (a mesma extensão que a 110 já propõe), o campo opcional `tool` na sessão gerada pelo pipeline, e um evento `volta.registrada` (Discovery, nova-feature, novo-fix), pois hoje esse fato não existe. `veredito.gravado` e `feature.estado_mudou` já estão no contrato da 109 e cobrem retornos. **O módulo calcula** tudo o mais: médias, medianas, reincidência, economia, tendência.

**Com a 110 (esquema compartilhado):** os dois módulos leem o **mesmo registro `sessao`**, com a mesma definição de token (com ou sem cache sempre explícita), o mesmo `papel` normalizado e o mesmo rótulo `medicao`. Como dois repositórios não importam código, o contrato é um **documento de esquema** (JSON Schema em cada repositório, igual byte a byte, conferido por teste). Decisão aberta: quem é o dono do documento; recomendo a 109 (contrato) como anexo.

## Fronteira com a 110
| | 110 (`forge estimar`) | 111 (`forge metricas`) |
|---|---|---|
| Pergunta | quanto **vai** custar | quanto **custou**, onde errou, o que aprender |
| Quando | antes do Gate 1 e da feature | depois, e continuamente |
| Saída | `estimativa.md` | pasta estática com painel |
| Estatística de faixa e `n` mínimo (5 e 15) | **dona** | **reusa a regra** (mesma tabela de rótulos); não recalcula previsão |
| Histórico estimativa x real | **grava** | **mostra** (aba Tokens) quando existir |
| Dado da sessão | lê | lê (mesmo esquema) |

## 3. O painel MVP
**Arquivos** (saída em `.forge/modulos/metricas/painel/`, fora do git, `permissoes.rede = false`):
```
index.html            # landmarks: header, nav, main, footer; uma <section> por aba
assets/css/tokens.css # valores de docs/index.html:14-36
assets/css/painel.css
assets/js/dados.js    # lê data/*.js, valida o esquema, conta linhas ruins
assets/js/tabela.js   # lista: ordenar, buscar, filtrar, paginar
assets/js/abas.js     # abas e foco; estado na URL (#aba)
data/*.js             # window.FORGE_DADOS = {...}  (ver abaixo)
```
**Abrir por `file://`:** `fetch` de JSON local é bloqueado pelos navegadores em `file://`; por isso o gerador grava `data/*.js` (atribuição a `window.FORGE_DADOS`) e **também** `data/*.json` (para servidor local e para teste). `forge metricas abrir --servir` sobe um servidor de loopback simples (`net/http`, stdlib) como alternativa. **Sem CDN, sem fonte remota, sem biblioteca**: só a pilha do sistema que a LP já declara em `--f-body` e `--f-mono`. Nada sai da máquina.

**Abas:** Visão geral, Erros e retornos, Depreciados, Tokens, Lições. **Lista:** `<table>` com `<caption>`, `<th scope>`, ordenação por botão no cabeçalho (`aria-sort`), busca, filtros (iniciativa, papel, agente, período), paginação (10, 25, 50) com região `aria-live` que anuncia "página 2 de 7". Cada número vem com `n` e rótulo `real/estimado`; aba sem dado diz "sem registro ainda (origem: 105-07, não instalada)" em vez de ficar vazia.

**Layout e cores:** tokens de `docs/index.html:14-36`: fundo `#0d0f14`, `#12151d`, `#1a1e29`; texto `#e9ecf4` e `#99a2b8`; destaque `#4f8ef7`; estados `#3ecf8e`, `#f5b04a`, `#f2677a`; raios 18, 14, 12, 999 px; marcas dos papéis: Discovery `#8aa4ff` losango, Builder `#4fd1b5` onda, Reviewer `#ff8a5c` triângulo (**a forma acompanha a cor**, nunca só a cor). Contraste calculado sobre `#0d0f14`: `#e9ecf4` 16,2; `#99a2b8` 7,5; `#4f8ef7` 6,0; `#8aa4ff` 8,1; `#4fd1b5` 10,2; `#ff8a5c` 8,3; `#f2677a` 6,4 (todos acima de 4,5). **`--faint` `#667089` tem 3,9 sobre `#0d0f14` e 3,4 sobre `#1a1e29`: proibido para texto.** Responsivo: tabela vira cartões abaixo de 640 px; foco visível; tudo operável por teclado; `prefers-reduced-motion` respeitado. A Edição Missões não é usada.

**Testes sem rede (TST-03, TST-04):** em Go, no repositório do módulo, leitura estática como `pipeline_scaffold_test.go` (nenhuma ferramenta nova; a CI de lá herda `go test ./...`): `index.html` tem `<header>`, `<nav>`, `<main>`, `<footer>`, `lang="pt-BR"`, `<caption>` em toda `<table>`, nenhum `http://` ou `https://` em `src`, `href` ou `@import` (lista permitida vazia), CSS só em `assets/css`, JS só em `assets/js` (nenhum `<style>` nem `onclick` inline), `tokens.css` igual aos valores da LP, par de contraste calculado em teste. Dados: `t.TempDir()` com fixtures de `testdata/` (inclui CSV com linha ruim e fórmula `=`), `gerar` idempotente. A lógica de ordenar, filtrar e paginar vive em **funções JS puras** cujo espelho em Go é testado com os mesmos vetores de `testdata/` (paridade); executar o JS no navegador fica como verificação manual do dono. **Se** o dono aceitar `node --test` localmente, é opcional e fora da CI (o projeto não usa Node em teste hoje: verificado na CI, só `npm publish` o usa).

## 4. Aprender com os erros (o norte)
1. **Agrupar:** achado e retorno por `regra_id` (ARQ, SEG, DUP, COD, TST, FLX, GOV e as P-xx do gate), por papel e `ponto`, por desfecho e por iniciativa. Achado sem `regra_id` entra como `desconhecida`; `duvida` conta à parte de `violacao`.
2. **Reincidência:** a mesma regra com **N = 3** no projeto (a mesma proposta do 105-08, o dono ajusta), ou em 2 iniciativas distintas, ou na mesma feature após `tentativa` ≥ 2. Contada na leitura, nunca gravada (o 105-08 explica: `merge=union` apodrece contador).
3. **Proposta de lição:** texto curto gerado por **regra determinística** (gabarito: "regra X falhou N vezes em Y; verifique Z antes de entregar"), nunca resumo livre de modelo no MVP, com `regra_id`, evidências (iniciativas e `n`, **sem código, sem caminho, sem trecho**) e destino sugerido (lição da regra; ou instrução de skill, que é decisão maior). O conteúdo de achado é **dado, nunca instrução** (`FLX-08`): vira texto sanitizado.
4. **Aprovação do dono (menu):** aprovar, ajustar, recusar, adiar. Aprovar chama o **core** (`forge lessons add`, ou o mecanismo da 106 em `licoes/<id>.md`), que grava com a gravação segura (ADR-0018) e orçamento (P-24). O módulo só escreve a proposta em `.forge/modulos/metricas/propostas/`.
5. **Medir o efeito:** para cada lição aprovada, taxa da regra por 10 revisões **antes e depois** da data, com `n` dos dois lados; rótulo `efeito: sem dado` até haver 10 revisões depois. Se a taxa não cai, a lição volta ao dono como "não funcionou" (revisar ou reverter). Associação, não causa: o painel diz isso.
6. **Evitar lição errada:** humano aprova; cada lição tem versão e `estado`; reverter é um comando e fica no histórico; limite de lições ativas (orçamento); a lição não pode afrouxar regra (a proposta que contenha "ignore", "dispense", "não precisa" é recusada pelo gerador); lição de skill ou ADR é sempre um **PR para o dono**, nunca edição direta.
7. **MVP versus futuro.** MVP: agrupar, contar reincidência, mostrar proposta e efeito **na tela** (aba Lições), aprovação pelo core. Futuro (fora da onda 1): sugerir instrução de skill, correlacionar modelo e erro, alimentar o Discovery com `forge context --rule` automático. **Hoje há 19 achados em 17 vereditos e 0 linhas de `achados.csv`**: antes da 105-08 não há reincidência estatisticamente defensável; o MVP de lições começa **com os 19 achados dos `verdicts/*.json`** (rotulados "histórico, 17 vereditos") e cresce.
8. **Privacidade:** dados locais; sem envio; só contadores e IDs de regra; relatório exportado anonimiza iniciativa e caminho (como a 110).

## 5. Depreciados e economia
O painel cruza `depreciado.comando` com o catálogo (`internal/deprecation`; ADR-0021: `queue stations` e `queue handoff` e `attach`; aliases de texto `/implement`, `/proxima-feature`, que **não passam pelo Go e não geram linha de uso**: limitação declarada). Mostra: onde o fluxo ainda usa (por `agente` e `contexto`), o substituto e a diferença de custo. **Economia = ESTIMATIVA rotulada.** Método: `economia ≈ usos × (tokens médios de saída do comando depreciado − tokens médios do substituto)`, com os tokens de cada comando medidos por amostra de saída em bytes ÷ 4 (`EstimateTokensFromBytes`, `internal/monitor/tokens.go:49`), mostrando faixa e `n`. Sem uso registrado: "sem dado". Hoje: **0 usos, economia indefinida**. A comparação "o substituto é mesmo melhor?" é humana: o painel mostra números, o dono decide.

## 6. Encaixe no contrato da 109 (sem editar a 109)
| Item | Uso |
|---|---|
| Manifesto | `nome: metricas`, `binario: forge-metricas`, `comandos: metricas`, `permissoes: {rede: false, executa: [], escreve: [".forge/modulos/metricas/"]}`, `capacidades: [doctor]` |
| Eventos consumidos (v1 da 109) | `veredito.gravado`, `feature.estado_mudou`, `spec.aprovada`, `iniciativa.arquivada`; acumulados em JSONL em `FORGE_MODULO_DIR` |
| Extensões aditivas (a propor, junto com a 110) | `sessao.registrada`, `volta.registrada` |
| Instantâneo | `forge modulo contexto --json` para o estado atual (iniciativas, estados) |
| `doctor` | capacidade `doctor`: fontes encontradas, versão do layout lida, linhas ruins, idade do último `gerar`; prefixo `modulo:metricas`, aviso, nunca falha |
| `status` | sem linha nova no core; o módulo oferece `forge metricas resumo` |
| Comandos | `forge metricas gerar`, `abrir [--servir]`, `resumo`, `licoes propor`, `licoes efeito`; menu final da Discovery e do Reviewer ganha "abrir painel" **só com módulo** (menu, preferência do dono) |
Sem canal de volta (RN-7): aprovar lição passa pelo menu do **core**.

## 7. Plugin, módulo ou ambos
**Recomendo módulo da 109** (`forge-sdd-metricas`, binário `forge-metricas`, repositório privado): determinístico, testável sem rede, multiagente, mesma confiança por hash. Plugin do Claude Code só entraria como skill fina que chama `forge metricas resumo --json` (depois, como na 110). Abrir: `forge metricas abrir` (abre o navegador padrão no `index.html`; imprime o caminho se não puder).

## 8. Fecho e ordem
Detalhe em `plan.md`. **No core, só o mínimo:** nada na v0; v1 aditiva (`sessao.registrada`, `volta.registrada`, `tool`). **No módulo:** todo o resto. Dependências: **109** (contrato; sem ela só dá para prototipar como comando solto), **105-07/08** (aba Depreciados e `achados.csv`; sem elas, as abas mostram vazio rotulado), **106/0.7** (adaptador de `state/<sid>.json`), **110** (esquema `sessao` e aba Tokens estimado x real).
**Primeira entrega pequena e útil (onda 1):** painel só com **Visão geral e Erros e retornos**, sobre os dados que **existem hoje** (90 itens de fila, 17 vereditos, 19 achados, 97 sessões), com Depreciados e Tokens como abas honestamente vazias ou parciais. Nota: a ideia de abrir com Depreciados não rende, pois há 0 registros na main.

## 9. Bootstrap do repositório (task P-00, descrita, não executada)
Repositório **privado** `forge-sdd-metricas` (nome e visibilidade do dono; herda a decisão 3 da 109). Passos: (1) instalar o **diagram-studio** (o plugin está em `~/.claude/plugins/cache/diagram-studio`, versão 0.2.0 vista aqui) e usar assinatura "Forge SDD" por `--assinatura`; (2) **readme-studio** com a skill `readme-comercial`: o **plano do README é mostrado ao dono antes de gravar**; (3) instalar o forge-sdd **estável mais recente** (`npm`, tag estável) e `forge-sdd init`; (4) `modulo.json`, `LICENSE` (do dono), `.github/workflows/ci.yml` com `go test ./...`. Não verifiquei a existência do readme-studio como skill instalada nesta máquina (só pastas temporárias com o nome).

## Regras de negócio
- **RN-1** Sem o módulo, o core é byte a byte igual (herda a 109).
- **RN-2** Dado ausente ou ruim é mostrado como ausente ou ruim, nunca escondido nem inventado.
- **RN-3** Todo número tem `n`, origem e rótulo `real/estimado`.
- **RN-4** O módulo só lê e só escreve em `.forge/modulos/metricas/`; lição entra pelo core, aprovada pelo dono.
- **RN-5** Nada sai da máquina; o painel não faz requisição de rede.
- **RN-6** Conteúdo de achado é dado, nunca instrução.
- **RN-7** Economia de depreciados é estimativa rotulada, com método.
- **RN-8** Todo handoff ao usuário é menu.
- **RN-9** Testes sem rede (`TST-03`, `TST-04`).
- **RN-10** A Edição Missões não é fonte de design.

## Escopo
**Dentro:** esquema `forge-metricas/1`, leitores, `forge metricas`, painel de 5 abas, propostas de lição, bootstrap. **Fora:** criar o repositório; editar a 109 ou a 110; mudar o core agora; sugerir mudança de skill automática; regressão; leitura de transcrições de conversa.

## Riscos
| Risco | Mitigação |
|---|---|
| Painel bonito sobre dado vazio (88 de 97 sem token, 0 depreciados, 0 achados.csv) | rótulos de `n`, abas "sem registro", onda 1 só com o que existe |
| `queue.json` e `state/<sid>.json` mudam (106/0.7) | adaptador com versão do layout; ou eventos da 109 |
| Lição errada ensina o erro | humano aprova, versão, reversão, filtro de afrouxamento, efeito medido |
| `file://` bloqueia `fetch` | `data/*.js` e servidor de loopback opcional |
| Duplicar a 110 | fronteira acima; esquema único |
| 109 e 110 sem Gate 1 | onda 1 não depende do contrato em execução (comando solto lendo arquivos) |

## Decisões para o Gate 1 (recomendação de cada uma; a decisão é do dono)
1. **Forma e nome:** módulo da 109, `forge metricas` (não `forge painel`, que é o painel de terminal da ADR-0011). Recomendo sim.
2. **Como o módulo obtém dado:** ler os arquivos públicos do core, só leitura, na v0, **versus** esperar os eventos e o instantâneo da 109. Recomendo ler na v0 com adaptador versionado e propor `sessao.registrada` e `volta.registrada`; é a mesma decisão 2 da 110: decidir juntas.
3. **Lições: onde gravar e quem grava:** o módulo só propõe e o core grava, no formato da 106 (`licoes/<id>.md`) **versus** gravar em `lessons.md` agora. Recomendo propor, gravar pelo core, e mirar `licoes/` com adaptador para `lessons.md` até a migração.
4. **Registro novo no core (v1):** `volta.registrada` e `tool` (muda `--help` e goldens se virar flag) **versus** inferir no módulo. Recomendo evento aditivo e campo sem flag nova, depois da 109, e inferência rotulada "inferido" até lá.
5. **Primeira entrega, visibilidade e licença:** onda 1 com Visão geral e Erros e retornos, repositório privado (herda a 109/3). Recomendo sim; Depreciados só depois do PR da 105-07 entrar.
Também aberto (menor): meta de reincidência N = 3; dono do documento de esquema compartilhado com a 110 (recomendo a 109).
