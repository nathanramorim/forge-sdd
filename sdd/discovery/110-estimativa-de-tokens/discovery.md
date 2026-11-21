# Discovery — 110 Régua de estimativa de tokens (previsibilidade de consumo por iniciativa e por feature)

## Pedido do dono (20/11/2026, chat)
Um plugin/módulo que, **já na Discovery**, estime quantos tokens custa executar a iniciativa e cada feature que sai do `/split-features`, com base na **média** das iniciativas já feitas, dizendo **qual agente, qual modelo e qual ferramenta**. Nasce como **módulo separado da 109** (repositório novo, privado por padrão, decisão aberta), com README (readme-studio, plano mostrado antes de gravar), diagramas (diagram-studio) e o forge-sdd estável mais recente instalado e configurado (`forge-sdd init`). **Nada foi criado**: nem repositório, nem código.

## ID
Conferido em 20/11/2026: `ls sdd/discovery | grep -E '^11'` vazio; `git branch -r | grep -E '/(feat|fix)/1[01][0-9]'` mostra 105 a 109 e nenhuma 110; `git worktree list` sem outra 110. O `#110` do GitHub é um PR (bump da LP), não iniciativa. Depois de criada a branch, `forge next-id` devolveu 111, porque conta o nome da própria branch; é esperado.

## Clarify
Sem pergunta antes de escrever: o pedido traz forma, fronteira e restrições. As lacunas viraram `## Decisões para o Gate 1`.

## Descrição de negócio (obrigatória em toda spec nova)
Hoje o dono só descobre quanto uma iniciativa consumiu depois de gasto. A régua mostra, antes do Gate 1 e antes de cada feature começar, uma **faixa** de tokens (e, se o dono mantiver uma tabela de preços, de dinheiro), com o agente, o modelo e a ferramenta que a história sugere. Ganho: o dono decide escopo e modelo com previsibilidade, e a esteira passa a medir o próprio erro e a melhorar. Regra de honestidade: a régua nunca finge precisão; mostra o tamanho da amostra e diz "sem estimativa" quando não há dado.

## Problema
O forge-sdd grava telemetria, mas ela é pouco usada para prever: tokens vazios, modelo vazio, nenhuma sessão de Discovery medida e nenhum campo de ferramenta (números abaixo). Sem dado, uma "média de todas as iniciativas" seria inventada.

## Para quem
Nathan (dono: orçamento, escolha de modelo, Gate 1) e, depois, qualquer usuário do forge-sdd que instale o módulo.

## Evidência: que dados de consumo existem (medido em `origin/main` 8bd3a0a e no checkout principal)

**1. `sdd/.metrics/session-*.json`** (esquema `forge-sdd/metrics/1.0`, `internal/monitor/session_record.go`, gravado por `forge session record`, `cmd/forge-sdd/session.go`). Campos: `feature`, `phase`, `agent_path`, `tokens_input/output`, `turns`, `duration_seconds`, `model`, `files_touched`, `outcome`, `sdd_version`, opcionais `tokens_cache_read/creation`, `cost_usd_tabela`, `medicao` (`real|estimado|resumo`), `complexidade` (reservado).

| Medida em 97 arquivos | Valor |
|---|---|
| arquivos de sessão | 97 |
| com `tokens_input` e `tokens_output` zero | 88 |
| com tokens diferentes de zero | 9 |
| `medicao: real` | 0 (1 `estimado`; 96 sem o campo, que `SumUsage` conta como estimado) |
| com `model` vazio | 63 (modelos nomeados: 6 grafias, 34 arquivos) |
| com `agent_path` vazio | 35 |
| com tokens de cache | 0 |
| sessões cuja `feature` é `discovery.md` | 11; **com tokens: 0** |
| de versão 4.9.x ou posterior | 1 |

Os 9 valores diferentes de zero (por exemplo 6.800/2.400 e 48.000/9.500) parecem arredondados e **possivelmente declarados pelo agente, não verificado**. Soma por sessão (entrada+saída), 8 sem `medicao`: mín. 3.800, mediana 11.200, máx. 70.900, coeficiente de variação 0,92.

**2. Logs de job do pipeline** (`.forge/runtime/logs/*.log`, fora do git, só no checkout principal; `ParseJobUsage` em `internal/monitor/usage.go` soma os eventos `result` do stream-json do Claude): 46 logs, **25 com evento `result` real** (12 builder, 13 reviewer, 0 discovery). Entrada+saída por job: builder mín. 4.039, mediana 19.113, máx. 115.055; reviewer mín. 1.609, mediana 3.201, máx. 8.373. **Com cache** (leitura+criação) a ordem de grandeza muda: builder mediana 1.752.174, máx. 21.921.898; reviewer mediana 202.326. Modelos nos logs: `claude-sonnet-5-5` (14), `claude-sonnet-4-6` (10), `claude-opus-5-5` (6). A régua precisa **dizer qual definição de token usa** (entrada+saída ou com cache).

**3. Transcrições do Claude Code** (`~/.claude/projects/<slug>/*.jsonl`, fora do forge): 38 arquivos neste projeto; em 1 arquivo lido (só as chaves) cada mensagem traz `usage` (`input_tokens`, `output_tokens`, `cache_read_input_tokens`, `cache_creation_input_tokens`) e `model`. Cobrem sessões **interativas**, como a Discovery. **Formato interno do Claude Code, sem contrato: não verificado se é estável.**

**4. Variáveis explicativas (já existem, do lado da entrada):** 26 pastas de iniciativa na main, 21 com `queue.json` (90 itens: 70 `approved`, 14 `completed`, 3 `queued`, 3 `review_pending`), 109 specs de feature/fix; `discovery.md` de 0,5 a 7,9 KB nas iniciativas lidas; nº de critérios e de tasks no `plan.md`.

**5. Fora da main:** o bloco `spec` do fix-97-01 (`internal/monitor/spec_metrics.go`: tipo, bytes, seções, features_count) **existe só na branch de PR #131, aberto**; é dependência, não código existente.

**Conclusão do dado:** o lado "tamanho da spec" tem ~109 pontos; o lado "tokens gastos" tem **25 jobs reais (só builder e reviewer) e 0 de Discovery**, e eles não estão ligados à feature por chave estável (o nome do log leva o id do job). Uma estimativa **por papel** de builder e reviewer, com faixa larga e aviso, já é possível; **estimativa de Discovery e por tipo/feature não é defensável hoje** (n = 0 e n ≤ 12 por papel, variação de 28 vezes no builder).

**Ferramentas.** Perfis em `internal/config/forge_profiles.go`: Claude (`--output-format stream-json`), Gemini/`agy` (`-o stream-json`, **a confirmar**), Codex (`--json`, **a confirmar**), Copilot (sem saída estruturada). `ParseJobUsage` só entende o formato `result` do Claude; para as outras três não há leitor: **tokens reais só para Claude; as demais "não verificado"** e caem na estimativa por bytes (`EstimateTokensFromBytes`, ~4 bytes por token).

## Modelo da estimativa (começar simples)
- **Unidade:** tokens por sessão, por papel (`pm-architect`/Discovery, `planner`, `builder`, `reviewer`/governança), por tipo de feature (`codigo`, `doc`, `adr`, `rule`) e por iniciativa (soma).
- **Variáveis:** bytes da spec, nº de critérios, nº de tasks, tipo, linguagem, uso de subagentes, modelo, ferramenta.
- **Método:** 1) mediana e faixa p25–p75 por papel; 2) por papel e tipo quando houver amostra; 3) regressão só com dados suficientes (decisão futura). Média aritmética não é usada como única medida (cauda longa; CV 0,92).
- **Incerteza (proposta, a confirmar):** n < 5 por célula: "sem estimativa" e mostra o observado; 5 ≤ n < 15: "amostra pequena", mediana e mín–máx; n ≥ 15: mediana e p25–p75. O n aparece sempre.
- **Saída:** `forge estimar` imprime e grava `sdd/discovery/<id>/estimativa.md`: uma linha por feature (faixa, agente e modelo recomendados, ferramenta, n) e total da iniciativa.
- **Recomendação de modelo/agente:** começa descritiva ("o mais usado com `approved` nesse papel"); sem n suficiente diz "sem recomendação".
- **Dinheiro (opção):** o módulo **não embute preço**: lê `precos.json` editável pelo dono, com `atualizado_em`; sem o arquivo, só tokens. Precedente: `cost_usd_tabela` ("preço de tabela, não é fatura").

## Coleta que falta
| Lacuna | Existe hoje? | Proposta |
|---|---|---|
| tokens de Discovery | não (0 de 11) | ler a transcrição do Claude (módulo, "não verificado") ou registrar `real` no fim da sessão |
| modelo | campo existe, 63/97 vazio | preencher a partir do `result`/transcrição (módulo) |
| agente | `agent_path` existe, 35 vazio | idem |
| **ferramenta (claude/codex/gemini/copilot)** | **não existe** | campo novo (ver abaixo) |
| ligação job a feature | só por nome de log | usar `feature` + `phase` da sessão |
| tokens por feature | não | amostra futura via `feature` |

**Mínimo no core (se algo), sem mudar o core sem módulo (RN-1 da 109, byte a byte):** nada na v0; a v1 acrescenta **só aditivo**: campo opcional `tool` no JSON da sessão gerado pelo pipeline (que já conhece o perfil) e, no contrato de módulos, o evento `sessao.registrada`. Mudar as flags de `forge session record` altera `--help` e o golden: **decisão do dono**. O resto fica no módulo.

## Encaixe no contrato da 109 (sem editar a 109)
Lido de `origin/feat/109-modulos-externos` (PR #154, ainda sem Gate 1). O contrato v1 tem eventos `feature.estado_mudou`, `spec.aprovada`, `pr.aberto`, `pr.pronto`, `pr.mesclado`, `veredito.gravado`, `iniciativa.arquivada`, o instantâneo `forge modulo contexto --json` e `doctor`/`pendentes`. Para a régua **não bastam**:
- não há evento de fim de sessão nem de split; `spec.aprovada` sai **depois** do menu, e a estimativa deve aparecer **antes** do Gate 1;
- o instantâneo não traz métricas e o módulo não lê arquivo do core;
- o core não ganha linha nos templates dos comandos.

Extensões **aditivas** propostas (dentro do v1, RN-6): evento `spec.criada` (Discovery terminou, com `iniciativa`, `features_count`, bytes) e `sessao.registrada` (papel, modelo, agente, ferramenta, tokens, `medicao`); ou campo `metricas` no instantâneo. Sem elas, o módulo roda como **comando de leitura** (`forge estimar`, invocado pelo dono ou por skill) lendo o esquema público `sdd/.metrics/schema.json`. Isso pede decisão: o contrato diz que o módulo não deve ler arquivo do core.

**Onde exibir:** `sdd/discovery/<id>/estimativa.md` (principal); no menu final da Discovery como opção "ver estimativa" (**handoff ao usuário é menu**, preferência do dono; não é a RN-9 da 109, que trata de testes sem rede); corpo do PR (opcional); linha em `forge status` só com módulo.

## Plugin do Claude Code versus módulo do forge
| Forma | Prós | Contras |
|---|---|---|
| **A. Módulo da 109** (binário `forge-estimar`) | determinístico, testável sem rede, multi-agente, mesma esteira e confiança por hash | depende da 109 (sem Gate 1); sem hook na Discovery |
| B. Plugin do Claude Code (skill/hook) | roda dentro da Discovery, acha a transcrição | só Claude; hook é do agente, não da esteira; não testável de forma determinística |
| C. Módulo com skill fina | motor do A, gatilho do B | duas peças para manter |

**Recomendação: A como motor, com C depois** (skill que só chama `forge estimar --json`); a transcrição do Claude entra como **fonte** do módulo, não como hospedeiro. Funciona em outros agentes porque o motor lê arquivos, não o agente.

## Privacidade e segurança
Tudo local; **nenhuma telemetria sai da máquina** e o módulo não precisa de rede (`permissoes.rede = false`). Lê só contadores (SEG-05), nunca o conteúdo das conversas. Se compartilhado, o relatório exportado anonimiza iniciativa e caminho; transcrições não são copiadas.

## Acurácia e honestidade
A cada iniciativa concluída o módulo grava estimativa versus real e o erro em `.forge/modulos/estimar/historico.jsonl`. **META a confirmar pelo dono, não fato:** erro mediano ≤ 40% com ≥ 15 iniciativas avaliadas, no papel builder; abaixo disso a saída carrega "calibrando". Recalibra a cada nova medição.

## Regras de negócio
- **RN-1** Sem o módulo, o forge-sdd se comporta byte a byte como hoje (herda a 109).
- **RN-2** Sem amostra suficiente, a saída diz "sem estimativa"; nunca número inventado.
- **RN-3** Toda faixa mostra n, a definição de token e a data do dado.
- **RN-4** Preço nunca embutido: só da tabela do dono, com data.
- **RN-5** O módulo só lê; não grava estado do core.
- **RN-6** Dado local, sem envio externo.
- **RN-7** Todo handoff ao usuário é menu.
- **RN-8** O repositório nasce dentro da esteira SDD, com o forge-sdd estável mais recente (P-00).

## Escopo
**Dentro:** leitor da amostra, estatística, `forge estimar` (somente leitura), `estimativa.md`, tabela de preços opcional, histórico de erro, bootstrap do repositório, README e diagramas. **Fora:** criar o repositório; alterar a 109 ou o core agora; regressão; leitor de tokens de Codex, Gemini e Copilot; preço embutido.

## Riscos
| Risco | Mitigação |
|---|---|
| Amostra pequena produz falsa precisão | rótulos por n, "sem estimativa" |
| Formato da transcrição do Claude mudar | adaptador isolado, marcado não verificado, cai na estimativa por bytes |
| Token com ou sem cache confunde | a definição aparece sempre |
| Módulo sem evento de Discovery | primeira entrega é comando de leitura; evento vira extensão aditiva |
| Depender de 109 e 107 | onda 1 não depende deles |

## Decisões para o Gate 1 (recomendação de cada uma; a decisão é do dono)
1. **Forma:** módulo-motor (A) e skill fina depois (C) **versus** só plugin do Claude. Recomendo A e C.
2. **Ler arquivo do core na v0** (`sdd/.metrics`, esquema público) **versus** esperar o evento/instantâneo da 109. Recomendo ler na v0, só leitura, e propor `sessao.registrada` à 109 como extensão aditiva.
3. **Campo `tool` e flags novas no core** (muda `--help` e golden) **versus** inferir no módulo. Recomendo o campo no JSON gerado pelo pipeline, sem flag nova, depois da 109.
4. **Metas de aceite** (n mínimo 5/15, erro ≤ 40% com ≥ 15 iniciativas). Recomendo adotá-las como metas revisáveis.
5. **Visibilidade e licença do repositório** (herda a decisão 3 da 109). Recomendo privado ao começar.
