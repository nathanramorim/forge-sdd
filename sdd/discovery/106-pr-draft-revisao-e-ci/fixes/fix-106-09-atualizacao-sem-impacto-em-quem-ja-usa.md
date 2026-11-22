# fix-106-09 — Atualizar quem já tem o forge instalado sem impacto no que está rodando

**Branch:** `docs/106-specs-lugar-certo-e-compat` (PR único das specs 08 e 09)
**Origem:** Iniciativa 106, pergunta do dono em 22/11/2026: "alguém já tem uma versão instalada anterior; como estruturar para os novos, sem gerar impacto ao que está rodando?"
**Tamanho:** `L`
**Status:** `spec` (Gate 1 aprovado em 22/11/2026, `[Seguir]`)

## Sintoma

Esta sessão acumulou mudanças que chegam a projetos com versão anterior instalada: comando renomeado (`/status` para `/forge-status`), subcomandos novos (`forge view`, `forge session mode`), campos novos de modo, estados novos em `.forge/runtime`, rules reescritas. Não existe uma garantia única de que **um projeto antigo atualiza sem perder nada e sem quebrar o que está rodando**.

## Como a atualização funciona hoje (verificado no código)

| Peça | Comportamento | Onde |
|---|---|---|
| Domínio do usuário | `sdd/**` **preservado**, exceto `.sdd-version` e `.sddrc` | `shouldPreserve`, `internal/scaffold/scaffold.go` |
| Skills `sdd-*` e configs de agente | **sobrescritas** a cada `init` e `update` | idem |
| Rules e skills do usuário | `.agents/rules/**` e skills que não são `sdd-*` **nunca sobrescritas** depois de criadas | idem |
| `AGENTS.md` | só atualizado se tem o marcador `<!-- forge-sdd:generated -->` | `IsGeneratedAgentsMD` |
| Renomeação ou remoção | lista **fixa no código** (`cleanObsoleteFiles`), com precedente `.prompt.md` para `.md` | `scaffold.go` |
| Simulação | `forge update --dry-run` mostra migração de layout e limpeza de obsoletos, sem gravar | `cmd/forge-sdd/update_plan.go` |
| Versão | `forge doctor` compara projeto e binário (`ok`, `projeto-antigo`, `binario-antigo`) e manda rodar `forge update` | `internal/doctor/versao.go` |
| Campo novo da Constituição | ausente vira valor padrão seguro (`modosPadrao()`: atestado `local`, revisão `com-humano`, WIP `3`, módulos `nenhum`, início automático `desligado`) | `internal/vcs/modos.go` |
| Job do barramento | tem `schema_version` e **recusa** versão diferente (`internal/bus/job.go:142`) | `internal/bus` |

## Furos (cada um com evidência)

1. **Rule alterada não chega ao projeto antigo.** `shouldPreserve` devolve verdadeiro para toda rule que já existe. Rule **nova** é criada; **edição** de rule existente (como as da 106) nunca propaga. Projeto antigo fica com regra velha e instrução nova.
2. **Renomeação depende de lembrar de editar a lista fixa.** Cada rename precisa de uma linha em `cleanObsoleteFiles`, sem teste que cobre "todo rename tem a linha".
3. **`queue.json` e os estados novos não têm versão de esquema.** O job do barramento tem, o `queue.json` não; `state/<sid>.json` e `sessao/<id>.json` (fix-106-06) nascem sem. Um worker antigo pode ler arquivo que não entende e errar em silêncio. **Correção de uma afirmação do chat:** o job **tem** `schema_version`; o furo é nos outros arquivos.
4. **A atualização não confere se há job rodando.** Não achei nenhuma checagem de job em andamento no `update`. Trocar skills e prompts no meio de uma execução muda o que o agente vê no meio do caminho.
5. **Comando novo em binário antigo.** Prompt atualizado chama `forge view`; se o binário do usuário é anterior, o subcomando não existe. O `doctor` avisa versão, mas nada **impede** o prompt novo de chegar antes do binário.
6. **Não há projeto-fixture antigo nos testes.** `internal/scaffold/testdata` só tem `golden` (a versão atual). Nada prova que "a versão N-1 atualiza para N".

## Feature Correlacionada

- **Arquivo de origem:** `internal/scaffold/scaffold.go` (`shouldPreserve`, `cleanObsoleteFiles`) e `cmd/forge-sdd/update_plan.go`.
- **Causa-raiz:** a compatibilidade foi tratada **caso a caso** (um rename aqui, um valor padrão ali), sem um **contrato** de atualização com teste de ponta a ponta. Cada mudança acerta isoladamente; ninguém testa a combinação.

## Modelo do módulo

**Completo** (`ARQ-11`): um manifesto embutido (`embed.FS`), seu leitor em `internal/scaffold`, uma checagem em `internal/pipeline`/`bus` para job ativo, e fixtures de teste. Sem subcomando novo (usa `update`, `update --dry-run` e `doctor`).

## Decisões do projeto aplicadas

- `ARQ-03` — o manifesto entra por `embed.FS`.
- `ARQ-05` — `--dry-run` nunca grava; imprime o plano completo.
- `DUP-01` — **um** manifesto de migração substitui a lista fixa; `update`, `doctor` e a nota de release leem dele.
- `SEG-01` — caminho do manifesto normalizado; nada fora do projeto; sem `..`.
- `SEG-02` — escrita de estado atômica.
- `GOV-01` — mudar a regra de preservação de rules é decisão de arquitetura: **ADR no mesmo PR**.
- `FLX-07` / RN-4 da 115 — toda mudança visível ao usuário entra primeiro numa **beta**.
- Regra do dono: pergunta e decisão por componente de ask (`fix-106-07`).

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Preservação e sobrescrita | `shouldPreserve` | regras por prefixo de pasta | **estender**: rule gerada e **não editada** pode ser atualizada |
| Marcador de arquivo gerado | `IsGeneratedAgentsMD` (`<!-- forge-sdd:generated -->`) | padrão já usado no `AGENTS.md` | **reusar** o padrão nas rules geradas |
| Limpeza de obsoletos | `cleanObsoleteFiles` | lista fixa | **substituir** pelo manifesto; manter o comportamento |
| Plano sem gravar | `update_plan.go` e `--dry-run` | existe | **estender** com renomeados, rules a atualizar e rules em conflito |
| Comparação de versão | `internal/doctor/versao.go` | existe | **reusar**; acrescentar "capacidade do binário" |
| Versão de esquema | `internal/bus/job.go` | só no job | **replicar o padrão** em `queue.json` e nos estados novos |
| Detecção de job ativo | `internal/bus`, `forge ps` | estados por pasta em `.forge/runtime/<papel>/<estado>` | **reusar** para recusar o `update` |
| Fixture de projeto antigo | `internal/scaffold/testdata` | só `golden` | **criar** `legado-N-1/` |

## Dependências novas (`SEG-07`)

`Nenhuma`.

## Rules e IDs da Constituição Aplicáveis

`ARQ-01`, `ARQ-03`, `ARQ-05`, `ARQ-11`, `COD-03`, `COD-04`, `DUP-01`, `FLX-05`, `FLX-07`, `GOV-01`, `SEG-01`, `SEG-02`, `TST-01`, `TST-03`, `TST-05`.

Rules: `sdd-arquitetura`, `sdd-codigo-e-testes`, `sdd-fluxo-e-portao`.

## Escopo

**Dentro:**
1. **Manifesto de migração** (`embed.FS`): lista declarativa de `renomeados`, `removidos`, `rules a atualizar`, `campos novos com valor padrão` e `estados novos`, cada item com a versão em que entrou. Lido por `update`, `update --dry-run`, `doctor` e pela geração da nota de release.
2. **Rules com dono:** a rule gerada leva o marcador e a **versão do conteúdo**. No `update`: se o arquivo está **igual ao que o forge entregou** (hash gravado em `.sddrc`), é atualizado; se o usuário **editou**, é **deixado como está** e o `doctor` avisa "rule X tem versão nova; revise" com o caminho do arquivo novo ao lado (`<nome>.novo.md`). Nunca sobrescreve edição do usuário.
3. **Versão de esquema** em `queue.json` e nos estados novos, com **leitor tolerante**: versão desconhecida **não é interpretada**; vira erro claro ("projeto mais novo que o binário") e nunca é reescrita.
4. **`update` recusa com job rodando:** se há job em estado `running` no barramento, termina com mensagem e código diferente de zero, sem tocar em nada. `--dry-run` continua permitido.
5. **Comando novo só depois do binário:** o manifesto marca a **versão mínima do binário** de cada comando novo. O `update` não instala o prompt que chama `forge view` se o binário em execução é anterior; avisa qual versão instalar.
6. **Fixture de projeto antigo** (`legado-N-1/`, com rules editadas, skills antigas, `queue.json` sem esquema, `/status` antigo) e teste de ponta a ponta: `update` N-1 para N preserva o domínio, aplica os renames, atualiza só rule não editada e não toca o que está rodando.
7. **Nota de release gerada do manifesto**, com "o que muda para você" e "o que fazer".

**Fora:**
- Atualizar sozinho sem o usuário rodar `forge update`.
- Migrar a Constituição do usuário (campos ausentes continuam valendo pelos padrões).
- Desfazer uma atualização (rollback). O `--dry-run` e a regra "nunca sobrescreve edição" são a proteção.

## Critérios de Aceitação Executáveis

```bash
set -o pipefail
# 1. O manifesto cobre tudo o que a lista fixa cobria (sem perda).
go test ./internal/scaffold/... -run 'TestManifestoCobreCleanObsoleteFilesAtual'
# 2. Todo rename declarado no manifesto é aplicado e testado.
go test ./internal/scaffold/... -run 'TestManifestoRenomeadosAplicados'
# 3. Rule não editada atualiza; rule editada é preservada e avisada.
go test ./internal/scaffold/... -run 'TestRuleNaoEditadaAtualiza|TestRuleEditadaPreservadaEAvisada'
# 4. Estado sem esquema ou com esquema desconhecido não é reescrito.
go test ./internal/queue/... ./internal/bus/... -run 'TestEsquemaDesconhecidoNaoEReescrito|TestQueueSemEsquemaLidoComoV1'
# 5. update recusa com job rodando; dry-run não.
go test ./cmd/forge-sdd/... -run 'TestUpdateRecusaComJobRodando|TestUpdateDryRunPermitidoComJobRodando'
# 6. Comando novo não chega antes do binário.
go test ./internal/scaffold/... -run 'TestComandoNovoSoComBinarioMinimo'
# 7. Ponta a ponta: projeto N-1 atualiza para N sem perder nada.
go test ./internal/scaffold/... -run 'TestAtualizaProjetoLegadoSemPerda'
go vet ./... && go test ./...
```

- **Teste negativo obrigatório (`TST-03`):** no fixture, uma rule **editada pelo usuário** continua byte a byte igual depois do `update`; um job em `running` faz o `update` terminar sem alterar **nenhum** arquivo (comparação de hash da árvore antes e depois); um binário anterior à versão mínima **não** recebe o prompt do comando novo.
- **Prova de que a lacuna era real:** o critério 7 falha hoje no fixture, porque a rule editada seria preservada mas **sem aviso** e a rule não editada **não** seria atualizada.

## Riscos

- **R1 — Hash de "não editado" falha** (fim de linha, espaço). Mitigação: normalizar quebra de linha antes de comparar; na dúvida, tratar como editado (o lado seguro).
- **R2 — `update` recusando demais** (job `running` esquecido). Mitigação: a mensagem diz qual job e como encerrar; `--dry-run` sempre disponível.
- **R3 — O manifesto vira mais um arquivo para esquecer.** Mitigação: teste que falha se um arquivo gerado some ou muda de nome sem entrada no manifesto.
- **R4 — Escopo grande (L).** Mitigação: o PR de implementação pode fatiar por item (1 e 7 primeiro, 2 depois), cada fatia com seu teste.
