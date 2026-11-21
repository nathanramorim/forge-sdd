# feat-105-08 — Log CSV dos achados de conformidade (repetição por rule)

**Iniciativa:** 105-experiencia-cli-e-harness (onda 3, task 3.6)
**Branch:** feat/105-experiencia-cli-e-harness (única; `FLX-02`)
**Depende de:** feat-105-07 (PR #129: `.gitattributes` com `merge=union` e o mecanismo de CSV seguro), 3.1 (telemetria; `FLX-04`), 106 task 1.1 (veredito com `ponto`; até lá, `ponto` vem do tipo do job). Não precisa da 107.
**Paralelo com:** 3.2 (não compartilham arquivos).
**Sequenciar com:** 106 (task 1.2 lições e task 2.6 passos dos comandos) e 107: série com rebase; nada da 106/107 é editado aqui.

## Problema
Hoje o achado do reviewer só vai do reviewer para o Builder (rework) e some: o veredito fica em `verdicts/<job>.json` e a mesma rule falha de novo em outra spec ou PR sem ninguém ver o padrão. Pedido do dono: registrar o erro de conformidade que se repete, no ponto Discovery para reviewer (gera ajuste da spec) e no ponto Builder para reviewer (gera rework/fix), para ver onde erra e ser previsível já ao gerar a spec.

## Evidência de onde o veredito passa hoje
| Caminho | Onde | O que acontece | Observação |
|---|---|---|---|
| Entrega (job `pr_review`) | `internal/pipeline/review.go` `ApplyVerdict` | valida, `bus.CheckOwner`, chama `persistVerdict` (l. 302), depois conforme aprova, `violacao` com `Attempt < MaxReworkAttempts` enfileira `pr_rework`, senão escala | `AlreadyApplied` retorna antes de `persistVerdict` |
| Spec (job `spec_review`) | `internal/pipeline/spec.go` `ApplySpecVerdict` | mesmo esqueleto; chama `persistVerdict` e `writeSpecReview` | só no modo avançado (`!cfg.IsAdvanced()` é erro): o ponto `spec` só tem dado nesse modo |
| Despacho | `cmd/forge-sdd/review.go` (`forge bus review`) | escolhe `ApplySpecVerdict` ou `ApplyVerdict` pelo tipo do job | decodificação estrita com `DisallowUnknownFields` |
| Fila (modo simples) | `queue.Review` (`internal/queue/queue.go:488`), chamado por `forge queue review` (`cmd/forge-sdd/queue.go:162`) | `harness.DecodeParecer`, grava `GovernanceVerdict` e muda o status | **não** passa por `persistVerdict` |
| Laço de chat | `internal/pipeline/laco.go` (`p.Agentes.Review`, `sanitizarAchados`) | achados ficam no estado do laço | **fora de escopo** desta task (a verificar numa futura) |
| Contrato | `internal/harness/parecer.go` `Parecer`, `Retorno` | `Regra`, `Severidade`, `Classificacao`, `Arquivo`, `Evidencia`, `Explicacao`; violação exige regra, arquivo, linha e evidência | achado `duvida` pode vir sem `Regra` |

Fato que limita o desenho (provado com `go list`): `internal/notify` importa `internal/queue`, e `internal/scaffold` importa `gate`, `notify` e `queue`. Logo nada que use `notify.Sanitize` ou `scaffold.WriteFileSafe` pode ser chamado de dentro de `internal/queue` nem de `internal/gate` sem ciclo.

## Comportamento
1. **Função única de captura:** `achados.Registrar(dir, Contexto, harness.Parecer)`. Dois sítios de chamada, porque os dois caminhos não compartilham código hoje: (a) `persistVerdict` em `internal/pipeline` (cobre `spec_review` e `pr_review`; `ponto` vem de `job.Type` até a 106 task 1.1 trazer `ponto` no `Parecer`; depois o campo do veredito vence) e (b) `queueReviewCmd` em `cmd/forge-sdd/queue.go` (o `queue.Review` não pode chamar por causa do ciclo acima; `ponto` = `entrega`). O revisor manual da spec no modo simples não passa por nenhum dos dois: **limitação declarada**, a 106 pode ligá-lo depois via `ponto`. Cobertura real: spec só no modo avançado; entrega nos dois modos.
2. **O que vira linha:** um achado por linha, só os de `classificacao` `violacao` ou `duvida`. Veredito `conforme` não grava nada. Veredito sem achados não grava.
3. **Arquivo:** `sdd/memory/logs/achados.csv`. O `.gitattributes` da 1.7 tem `sdd/memory/logs/*.csv merge=union` e já cobre o novo arquivo (confirmado no código da branch da 1.7; vale quando o PR #129 entrar). Decisão do dono: versionar com `merge=union`.
4. **Colunas** (cabeçalho na 1ª linha, UTF-8):
   | Coluna | Conteúdo | Justificativa |
   |---|---|---|
   | `data_utc` | RFC 3339 UTC | ordenar |
   | `ponto` | `spec` ou `entrega` | pedido do dono; origem e destino ficam implícitos (spec = Discovery para reviewer; entrega = Builder para reviewer), por isso não há coluna `origem` (`DUP-06`) |
   | `iniciativa` | id da iniciativa do job; vazio se não houver | agrupar |
   | `task` | `TaskIDs` do job ligados por `+` (limite de 100 caracteres), ou o `task-id` da fila | localizar |
   | `job` | id do job (vazio na fila) | idempotência (item 6) |
   | `rule_id` | ID da Constituição que casa com `^(ARQ\|SEG\|DUP\|COD\|TST\|FLX\|GOV)-\d{2}$` na `Regra`; senão `desconhecida` | `Regra` é texto livre do modelo; agrupar exige chave validada |
   | `severidade` | `Severidade` do achado, só se for um dos valores vistos no contrato, senão vazio (a verificar o conjunto no Builder) | priorizar |
   | `classificacao` | `violacao` ou `duvida` | separar erro de incerteza |
   | `desfecho` | `rework` (devolvido ao Builder), `ajustar` (spec devolvida), `escalado` (teto ou dúvida) ou `registrado` (fila, sem fluxo) | em vez de `gerou_fix`: nenhum código Go cria "novo-fix" hoje (é a task 2.5 da 106); `rework` e `ajustar` são o que gera trabalho de correção |
   | `tentativa` | `Attempt` do job; vazio na fila | ver se corrige de primeira |
   | `arquivo` | caminho relativo ao projeto; vazio se absoluto, com `..`, com `~` ou acima de 120 caracteres | "ver onde tem errado" sem vazar caminho do usuário |
   | `resumo` | `Verificacao` (nome curto da checagem) sanitizado, 100 caracteres | curto e sem código |
   | `versao_forge` | versão do binário | correlacionar |
   A coluna `repeticao` **não é gravada**: com `merge=union` dois PRs paralelos gravariam o mesmo contador, que apodrece. A repetição é recontada na leitura (item 7). `Evidencia`, `Explicacao` e `Correcao` **nunca** entram (podem ter código, caminho ou segredo).
5. **Escrita segura, reaproveitando a 1.7 sem duplicar.** O `deprecation.Registrar` da 1.7 não é genérico (cabeçalho, `RelPath`, `ErrLimite` e `Celula` com 200 caracteres fixos). Esta task extrai um pacote de apoio `internal/csvlog` (anexar linha com cabeçalho, limite de bytes, recusa de link simbólico, escrita atômica por `scaffold.WriteFileSafe`, `Celula` com neutralização de fórmula) e **refatora `internal/deprecation` para chamá-lo** (os testes da 1.7 seguem verdes sem edição de comportamento). Limite do arquivo: 256 KiB, igual à 1.7; acima disso não grava mais e avisa uma vez. Cada célula passa por `notify.Sanitize` e, se iniciar com `=`, `+`, `-` ou `@`, ganha prefixo `'`. Como `gate` não pode importar `scaffold` nem `achados` (ciclo), a escrita fica em `achados`/`csvlog` e a leitura (item 7) é chamada só do `cmd`.
6. **Robustez:** valida na borda (`SEG-01`): o `Parecer` já vem de `DecodeParecer`/`validateVerdict`; a captura roda depois da validação e da posse (`CheckOwner`). Falha de escrita NUNCA quebra a revisão: só uma linha de aviso no stderr (`COD-03`). Idempotência: como `persistVerdict` roda antes de `Complete`/`Enqueue`/`Dispatch`, que podem falhar e permitir reaplicar o veredito, a linha só é gravada se não existir outra com o mesmo `job`, `rule_id`, `arquivo` e `tentativa` (a captura já lê o arquivo para anexar). Sem `job` (fila), sem dedupe. `--dry-run` não grava (`ARQ-05`). Concorrência: ler e regravar pode perder uma linha; aceito e documentado, como na 1.7. O conteúdo do achado é DADO (`FLX-08`): nunca é executado nem obedecido; só vira texto sanitizado numa célula.
7. **Consulta barata por rule, sem novo subcomando.** A Constituição está com 9189 de 9216 bytes (`MaxConstitutionSize`, P-24): sem folga para tocar `ARQ-06`. Logo: (a) função de leitura `achados.Consultar(dir, ruleID, iniciativa)` que devolve até 5 linhas mais recentes da rule, o total na iniciativa e o total no projeto (recontados das linhas; linha malformada é ignorada com contagem e aviso, nunca falha); (b) `forge context <task|spec>` (já listado em `ARQ-06`) acrescenta ao fim da saída a seção "Repetições de conformidade" só para os IDs que a spec cita (`ContextPack.IDs`); a chamada fica em `cmd/forge-sdd/context.go`, pois `gate` não pode importar `achados`; (c) flag nova `forge context --rule <ID>` (sem argumento posicional) para o Discovery, que ainda não tem spec. Contrato de `context` muda de forma aditiva: validar `--rule` com a regex de ID, ID inexistente na Constituição é erro (como hoje). Saída com teto de linhas e de bytes e sempre sanitizada. Acima de **N = 3** repetições no projeto (proposta, o dono aprova) a seção imprime o aviso fixo "a rule repete; leia a lição da rule e confirme se ela está funcionando".
8. **Passos dos comandos (texto), uma vez, em série com a 106 task 2.6:** em `.agents/skills/sdd-concepcao/references/{discovery,nova-feature,novo-fix}.md` e `.agents/skills/sdd-build/references/{implementar,worker-job}.md`, uma linha: "antes de escrever/implementar, rode `forge context --rule <ID>` só para a rule que vai tocar (o Builder já recebe a seção em `forge context <task>`)". Não entra em `.claude/commands/*` (adaptador de ~130 bytes, `MaxAdapterBytes` 300), nem em `.claude/agents/*` (`MaxRoleAdapterBytes` 420), nem na Constituição (P-24). Os mesmos arquivos são templates do scaffold: goldens regravados no mesmo PR (`TST-02`). O limite de `SKILL.md` (2048) não muda, pois nada entra nele (a verificar com `forge gate` no Builder se alguma reference cai em outro grupo de P-24).
9. **Pareceres antigos da 97 que não decodificam (5 de 17, `harness.md`):** o log não faz retroativo. Só o veredito que passa agora por `DecodeParecer` gera linha. A leitura de vereditos antigos para métricas (3.2) conta o arquivo ilegível e segue, sem falhar.

## Descrição de negócio (release notes)
Item pronto para `sdd/releases/history.md`:

*   **O forge lembra onde mais se erra (`sdd/memory/logs/achados.csv`):** a cada revisão, da spec e da entrega, o forge registra os achados de conformidade, uma linha por achado, com a rule violada, o desfecho (devolvido, ajustar ou escalado) e a tentativa. Ele conta quantas vezes cada rule já falhou na iniciativa e no projeto. Antes de escrever a spec ou de implementar, `forge context` (e `forge context --rule <ID>`, para o Discovery) mostra as repetições só das rules que você vai tocar, com um aviso quando a mesma rule repete. As specs saem mais previsíveis e o mesmo erro deixa de se repetir sem ninguém ver. Nesta entrega o forge registra, conta e mostra ao Discovery e ao Builder; ele ainda **não** atualiza ADRs, rules nem agentes sozinho. O registro vira planilha, não guarda código nem caminho seu e nunca atrapalha a revisão (Iniciativa 105).

## Habilita (fora desta task)
"Correção inteligente: a base se atualiza com os erros cometidos". Com o log, achados que se repetem permitem que o forge e os agentes **proponham** atualizar a lição (106), a ADR, as rules e as instruções dos agentes. Nada disso entra na 3.6. Salvaguardas que isso exigirá (sem decidir aqui):
- Toda alteração em ADR, rule, Constituição ou instrução de agente é PROPOSTA com aprovação humana (`GOV-01`, ADR e Gate), nunca edição automática.
- O conteúdo de um achado é DADO, nunca instrução (`FLX-08`): risco de prompt injection e de um agente "ensinar" a si mesmo a afrouxar uma regra.
- Orçamento de contexto (P-24) limita o que cada proposta pode acrescentar.
- Medir se a lição reduziu a repetição: o log dá esse número.

## Fronteira com a 106 (lições) — sem duplicar (`DUP-01`)
- **Lição** (106 RN-8, task 1.2, `sdd/memory/licoes/<id>.md`, `licoes.py --rule`): orientação pequena por rule, "como evitar". Dono do conceito: a 106.
- **Log** (esta task): ocorrências e contagem de repetição. Não guarda orientação; não escreve lição.
- **Ordem:** o reviewer grava a linha do log no momento do veredito (esta task); a lição nasce do mesmo achado pela 106. A leitura da rule mostra primeiro a repetição (`forge context --rule`) e a 106 mostra a lição. Acima de N, o aviso do item 7 sinaliza que a lição não está funcionando. Esta task **não lê** `sdd/memory/licoes/` (acoplamento evitado; o aviso é texto fixo). Se a 106 não existir quando esta entrar, nada quebra.
- Registrar o novo conceito "log de achados" em `sdd/spec/concepts.md` com `internal/achados` como dono (linha nova, sem tocar nas linhas da 106).

## Relação com a proficiência (`harness.md`)
O log tem uma linha por achado, sem denominador de jobs: não calcula taxa sozinho.
- **Ganha dado:** falhas por rule (novo); H2 parcial (`tentativa` e `desfecho` dos achados); H9 parcial (`desfecho` `escalado`).
- **H1** (gate verde de primeira) continua exigindo `verdicts/*.json` mais `Job.Attempt` (denominador): segue "a medir". H3 a H8 não mudam.
- A 3.2 (agregador) e a 3.3 (`forge report --proficiencia`) podem ler este CSV; ajusta-se só a coluna "Depende de" da 3.2 no `plan.md`, sem reescrevê-la.

## Modelo do módulo
**Novo** `internal/achados` (modelo da linha, validação do `rule_id`, redução do achado a células, `Registrar`, `Consultar`) e **novo** `internal/csvlog` (anexo seguro genérico, extraído da 1.7). `cmd/forge-sdd` só liga: chamada em `queueReviewCmd`, flag e seção em `context.go` (`ARQ-01`, `ARQ-10`). `internal/pipeline` chama `achados.Registrar` em `persistVerdict`. Dependências: `achados` importa `harness` (tipos), `notify` e `csvlog`; `gate` e `queue` não importam nenhum dos dois (ciclo provado). A verificar no Builder com `go list`: se `csvlog` pode importar `scaffold`; se houver ciclo, mover `WriteFileSafe` para um pacote folha.

## Decisões do projeto aplicadas
ADR-0004 (o Go decide o fluxo; o log não altera o fluxo), ADR-0002 (estado em `sdd/`), ADR-0021 (3 sessões; criada na 1.1), ADR-0013 (modo avançado: `spec_review` só nele). Nenhuma decisão nova; sem ADR porque `ARQ-06` não muda (flag aditiva em `context`).

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Ponto de captura | `internal/pipeline/review.go:302` `persistVerdict`, chamado por `ApplyVerdict` e `ApplySpecVerdict` | Grava cópia do veredito, melhor esforço | **Estender** | `DUP-01` |
| Fila manual | `internal/queue/queue.go:488` `Review`; `cmd/forge-sdd/queue.go:162` | Sem job; importar `notify` daria ciclo | **Chamar do `cmd`** | `ARQ-01` |
| Contrato do achado | `internal/harness/parecer.go` `Parecer`, `Retorno` | Campos de regra, severidade, arquivo | **Reusar** | `DUP-06` |
| Sanitização do achado | `internal/pipeline/laco.go` `sanitizarAchados` | Limita e sanitiza para o laço | **Reusar o padrão** (`notify.Sanitize`); a função não é exportada | `DUP-01` |
| CSV seguro | `internal/deprecation/deprecation.go` (branch da 1.7, PR #129): `Celula`, `Registrar`, `MaxBytes` | Específico de depreciados | **Extrair** para `csvlog` e refatorar | `DUP-01` |
| Escrita segura | `internal/scaffold` `WriteFileSafe` | Atômica, recusa link simbólico | **Reusar** | `SEG-02` |
| Merge de CSV | `.gitattributes` da 1.7 | `sdd/memory/logs/*.csv merge=union` | **Reusar** | decisão do dono |
| ID de rule | `internal/gate/context.go` `constitutionIDLine` | Regex dos IDs | **Reusar o padrão** | `DUP-06` |
| Leitura mínima por rule | `forge context` | Já traz IDs citados e rules | **Estender** (seção e flag) | `ARQ-06` |
| Lições por rule | `origin/feat/106-pr-draft-revisao-e-ci` (RN-8, task 1.2) | Previsto, não implementado | **Não tocar**; fronteira acima | `DUP-01` |
| Pasta de logs | `sdd/memory/logs/` | Criada pela 1.7 | **Reusar** | — |

## Dependências novas
Nenhuma (`SEG-07`); só stdlib (`encoding/csv`).

## Riscos
- Cobertura parcial (spec só no modo avançado; fila manual de spec fora; laço de chat fora): dito no item 1; o log não prova ausência de erro.
- `Regra` do modelo fora do padrão cai em `desconhecida`; muitas linhas assim indicam prompt do reviewer a ajustar.
- Perda de linha em concorrência; limite de 256 KiB trava o registro (aviso uma vez).
- Refatorar `deprecation` mexe em código da 1.7 (PR #129): fazer só depois dela mesclada.
- N = 3 é proposta sem calibração.

## Rules e IDs da Constituição Aplicáveis
`ARQ-01`, `ARQ-05`, `ARQ-06`, `ARQ-10`, `COD-03`, `COD-05`, `DUP-01`, `DUP-06`, `FLX-08`, `GOV-03`, `SEG-01`, `SEG-02`, `SEG-04`, `SEG-05`, `TST-02`, `TST-03`. Rules: `sdd-arquitetura`, `sdd-codigo-e-testes`, `sdd-seguranca`, `sdd-reuso-e-duplicacao`.

## Critérios de Aceitação Executáveis
Testes usam `t.TempDir()` e `t.Setenv`; não dependem da identidade do git nem do `forge` do PATH.
```bash
set -o pipefail
go build ./... && go vet ./...
go test -list 'TestRegistrarGravaUmaLinhaPorAchado' ./internal/achados | grep -F 'TestRegistrarGravaUmaLinhaPorAchado'
go test ./internal/achados -run 'TestRegistrarGravaUmaLinhaPorAchado'
go test -list 'TestPontoVemDoTipoDoJob' ./internal/pipeline | grep -F 'TestPontoVemDoTipoDoJob'
go test ./internal/pipeline -run 'TestPontoVemDoTipoDoJob'
go test -list 'TestConsultarContaRepeticaoPorRule' ./internal/achados | grep -F 'TestConsultarContaRepeticaoPorRule'
go test ./internal/achados -run 'TestConsultarContaRepeticaoPorRule'
go test -list 'TestContextMostraRepeticoesDasRulesCitadas' ./cmd/forge-sdd | grep -F 'TestContextMostraRepeticoesDasRulesCitadas'
go test ./cmd/forge-sdd -run 'TestContextMostraRepeticoesDasRulesCitadas'
go test -list 'TestDeprecationUsaCsvlog' ./internal/deprecation | grep -F 'TestDeprecationUsaCsvlog'
go test ./internal/deprecation -run 'TestDeprecationUsaCsvlog'
# negativo: conforme não grava; evidência, explicação e correção nunca entram no CSV
go test -list 'TestConformeNaoGravaEEvidenciaNuncaEntra' ./internal/achados | grep -F 'TestConformeNaoGravaEEvidenciaNuncaEntra'
go test ./internal/achados -run 'TestConformeNaoGravaEEvidenciaNuncaEntra'
# negativo: regra fora do padrão vira desconhecida; arquivo absoluto, com .. ou ~ vira vazio
go test -list 'TestRuleIDEArquivoSaoValidados' ./internal/achados | grep -F 'TestRuleIDEArquivoSaoValidados'
go test ./internal/achados -run 'TestRuleIDEArquivoSaoValidados'
# negativo: célula iniciada por = + - @ é neutralizada; quebra de linha e ANSI saneados
go test -list 'TestCelulaNeutralizaFormulaESaneia' ./internal/csvlog | grep -F 'TestCelulaNeutralizaFormulaESaneia'
go test ./internal/csvlog -run 'TestCelulaNeutralizaFormulaESaneia'
# negativo: link simbólico no destino é recusado e a revisão segue
go test -list 'TestRegistroRecusaLinkSimbolico' ./internal/csvlog | grep -F 'TestRegistroRecusaLinkSimbolico'
go test ./internal/csvlog -run 'TestRegistroRecusaLinkSimbolico'
# negativo: falha de escrita não quebra o veredito (aviso no stderr, ação aplicada)
go test -list 'TestFalhaDeEscritaNaoQuebraRevisao' ./internal/pipeline | grep -F 'TestFalhaDeEscritaNaoQuebraRevisao'
go test ./internal/pipeline -run 'TestFalhaDeEscritaNaoQuebraRevisao'
# negativo: reaplicar o mesmo veredito não duplica linhas
go test -list 'TestReaplicarVereditoNaoDuplicaLinha' ./internal/pipeline | grep -F 'TestReaplicarVereditoNaoDuplicaLinha'
go test ./internal/pipeline -run 'TestReaplicarVereditoNaoDuplicaLinha'
# negativo: arquivo acima do limite não cresce
go test -list 'TestRegistroRespeitaTamanhoMaximo' ./internal/csvlog | grep -F 'TestRegistroRespeitaTamanhoMaximo'
go test ./internal/csvlog -run 'TestRegistroRespeitaTamanhoMaximo'
# negativo: --dry-run não cria o CSV (ARQ-05)
go test -list 'TestDryRunNaoGravaAchados' ./cmd/forge-sdd | grep -F 'TestDryRunNaoGravaAchados'
go test ./cmd/forge-sdd -run 'TestDryRunNaoGravaAchados'
# negativo: linha malformada e parecer antigo ilegível são ignorados com aviso, sem falhar
go test -list 'TestLinhaMalformadaNaoFalhaConsulta' ./internal/achados | grep -F 'TestLinhaMalformadaNaoFalhaConsulta'
go test ./internal/achados -run 'TestLinhaMalformadaNaoFalhaConsulta'
# negativo: --rule com ID inválido ou inexistente na Constituição é erro
go test -list 'TestContextRuleInvalidaEErro' ./cmd/forge-sdd | grep -F 'TestContextRuleInvalidaEErro'
go test ./cmd/forge-sdd -run 'TestContextRuleInvalidaEErro'
go test ./...
```
- Textos novos só nas `references` (item 8); `forge gate` mantém P-24 verde (Constituição não cresce) e os goldens são regravados no mesmo PR.
- O Builder registra no `status.md` o que ficou "a verificar" (conjunto de `severidade`, ciclo de import de `csvlog`, grupo de P-24 das references).
