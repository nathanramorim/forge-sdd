# ADR-0004: Pipeline Multi-Sessão (Discovery, Builder e Reviewer)

- **Data:** 2026-11-13
- **Status:** `aprovado`
- **Regras afetadas (IDs da constitution):** `ARQ-06` (novos subcomandos `pipeline`, `dispatch`, `bus`, `worker`, `inbox` e `up`), `DUP-01`, `FLX-06`, `SEG-06`, `GOV-01` (este ADR; nenhum ID foi criado ou removido)
- **Autor / Responsável:** Líder (decisão do dono do produto no discovery 70)

## Contexto

Uma sessão só concebia, implementava e revisava: contexto misturado, revisão enviesada, nada de paralelismo entre conversar e construir, e nenhum portão que exigisse um card no gerenciador de projeto nem um ciclo automático de correção quando o PR violava uma ADR ou rule. Referência: `sdd/discovery/70-pipeline-multi-sessao-orquestrada/` e `feat-70-01` a `feat-70-07`.

## Opções Consideradas

1. **Sessões interativas esperando a fila:** uma CLI interativa não consegue consultar uma fila sozinha.
2. **Worker em Go que invoca o agente de forma não interativa por job (escolhida).**
3. **Barramento por serviço ou fila externa:** adiciona infraestrutura e dependências que o produto não tem (`ARQ-04`).

## Decisão

- **D-1 Worker em Go.** `forge-sdd worker --role builder|reviewer` reivindica um job e invoca o agente com o perfil da CLI de `forge.json` (`claude` é a referência; os argumentos do `agy` são configuráveis e **a confirmar**), com argv explícito e timeout, e renova o lease durante a execução.
- **D-2 Gate em duas metades.** O Go valida o que é determinístico (spec, tasks, ADRs, rules, IDs); o agente consulta o card pelo MCP e grava `card-evidence.json`, que o Go valida (esquema estrito, provider, feature e frescura). Sem evidência não há despacho.
- **D-3 Máquina de estados do Reviewer no Go.** O modelo emite o veredito; `forge-sdd bus review` aplica a regra: aprova e notifica, devolve como `pr_rework` (prioridade alta, tentativa + 1) até `max_rework_attempts`, escala no limite e em dúvida.
- **D-4 Barramento em arquivos.** Um job por arquivo em `.forge/runtime/queues/<papel>/<estado>/`; a reivindicação é `os.Rename` atômico com lease; o nome do arquivo ordena por prioridade e FIFO; `recover` devolve jobs de lease vencido. `internal/bus` é um conceito novo, separado da fila de tasks (`internal/queue`).
- **D-5 Contrato do job** com os campos da especificação mais `schema_version`, `initiative`, `task_ids`, `lease` e `events`.
- **D-6 Modo por `forge.json`.** Sem `.forge/config/forge.json`, nada muda em relação à 4.0.0-beta; o scaffold entrega só arquivos `.example`.
- **D-7 Dados não confiáveis.** Diff, corpo do PR e `rework_feedback` são dados (`FLX-08`); o Builder só aplica o que mapeia para regra ou ADR citada e está no escopo. Segredos de e-mail e webhook só por variável de ambiente, nunca em arquivo, job, log ou mensagem de erro (`SEG-04`, `SEG-05`).
- **D-8 Launcher em Go.** A lógica de abrir os terminais por sistema e CLI vive em `forge-sdd up` (testável por geração de comandos); `forge-up.sh` e `forge-up.ps1` são envoltórios finos.
- **Agente `reviewer`:** definição própria somente leitura sobre o código (`SEG-06`), com `gh` apenas para comentar no PR quando configurado.
- **Notificações:** o canal padrão grava no inbox da sessão Discovery (com toast do sistema, se ligado) e um hook do Claude (`UserPromptSubmit`) mais o comando `/inbox` exibem o link do PR; e-mail, webhook e `pm_comment` são opcionais, e `pm_comment` vira instrução ao agente porque o MCP só é alcançável por ele.
- **`ARQ-06`** passa a listar os novos subcomandos; nenhum ID da Constitution mudou.

## Consequências

- **Positivas:** conversa e construção deixam de competir pelo mesmo contexto; toda feature despachada tem card; o ciclo de correção é automático e limitado; os seis cenários de aceitação rodam em testes sem MCP e sem modelo.
- **Negativas / Trade-offs:** cada job é uma execução headless com custo de tokens próprio (mitigado por `forge-sdd context`); os flags headless do `agy` não estão confirmados; a abertura real dos terminais por sistema só é verificável manualmente (os testes cobrem a geração dos comandos); a primeira versão não reaproveita o `ProcessRunner` da feat 68 (ainda não mesclada).
- **Ações de acompanhamento:** confirmar os flags do `agy`; convergir o runner com o da feat 68 quando ela entrar; avaliar se a escalada por `duvida` precisa de um canal próprio.
