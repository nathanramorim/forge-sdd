# feat-105-09 — Descrição de negócio (release notes) em toda spec nova de feature e de fix

**Iniciativa:** 105-experiencia-cli-e-harness (onda 2, task 2.7)
**Branch:** feat/105-experiencia-cli-e-harness (única; `FLX-02`)
**Depende de:** feat-105-02 (texto das references e dos templates já corrigido, para não regravar duas vezes) e, na parte de ligação (item 4), da 106 task 2.6 mesclada.
**Paralelo com:** 2.2 a 2.5 (sem arquivos em comum).
**Sequenciar com:** 106 task 2.6 (reescreve os passos de `/discovery`, `/nova-feature`, `/novo-fix` e do split nos mesmos arquivos) e 107 (tabela de roteamento): série com rebase; nada da 106 nem da 107 é editado aqui.

## Problema
O dono decidiu: toda spec **nova** de feature e de fix passa a trazer a seção `## Descrição de negócio (release notes)`, com um item pronto para `sdd/releases/history.md`. Hoje nenhum template nem skill a exige; só a spec da 105-08 a tem, escrita à mão. A Regra 12 da Constituição pede "destaques reais" em `history.md` por versão; sem a seção, o texto do release é escrito no fim, de memória.

## Onde a spec nasce hoje (evidência)
| Gerador | Arquivo | O que diz |
|---|---|---|
| `/nova-feature` | `.agents/skills/sdd-concepcao/references/nova-feature.md` passo 6 | "as seis seções obrigatórias da skill `sdd-concepcao`" |
| `/novo-fix` | `.../references/novo-fix.md` passo 6 | mesmas seis seções mais `## Feature Correlacionada` |
| `/discovery` | `.../references/discovery.md` passo 6 | seções do `criteria.md`; não gera spec de task |
| `/split-features` | `.agents/skills/sdd-planejamento/references/split-features.md`; `SKILL.md` da planejamento diz "as cinco seções obrigatórias" | contagem diverge da concepção (seis); a spec de task é escrita aqui |
| Lista das seções | `.agents/skills/sdd-concepcao/SKILL.md`, regra 4 (2045 de 2048 bytes, `MaxSkillBytes`) | sem folga: não dá para acrescentar item à lista |
| Templates do scaffold | `internal/scaffold/templates/.agents/skills/sdd-*/references/*.md.tmpl` e `SKILL.md.tmpl` | espelham os arquivos acima; goldens (`TST-02`) |
Não existe template de spec com a seção: confirmado por busca em `internal/scaffold/templates` e `.agents` (a verificar de novo no Builder).

## Comportamento
1. **Texto num ponto único.** Cria-se um arquivo novo `.agents/skills/sdd-concepcao/references/descricao-de-negocio.md` (e o `.tmpl` espelho), que define a seção. As references de `nova-feature`, `novo-fix` e `split-features` apenas **apontam** para ele numa frase curta ("inclua `## Descrição de negócio (release notes)` conforme `references/descricao-de-negocio.md`"). Motivo: o `SKILL.md` está sem folga e as references são reescritas pela 106 task 2.6; o conteúdo vive em um arquivo que a 106 não toca.
2. **Texto mínimo da seção** (o que o arquivo manda escrever, em pt-BR e linguagem de produto):
   - Um único item pronto para `sdd/releases/history.md`, no tom das entradas existentes: `*   **Nome do benefício (`comando`):** o que muda para quem usa, em frase corrida. (Iniciativa NN).`
   - Conta o valor para quem usa, cita o comando ou o artefato, e fecha com "Iniciativa NN".
   - Honesto sobre o que entrega agora e sobre o que ainda não faz.
   - Sem número que não foi medido; sem jargão de código nem nome de função.
   - Spec de fix: uma frase do que deixa de dar errado para o usuário (ou "sem mudança visível", dito às claras).
3. **Como Builder e Reviewer conferem:** o Builder confere que a spec que implementa tem a seção e ajusta o texto se a entrega divergiu (nunca promete o que não entregou). O Reviewer confere a seção contra a entrega (promessa sem entrega é achado) e contra a regra de honestidade acima. Sem gate novo nesta task (item 6).
4. **Ligação nos passos:** as três frases de apontamento entram nas references **depois** da 106 task 2.6 (rebase e regeração dos goldens, cada arquivo escrito uma vez por task). Se esta task vier antes, ela entrega só o arquivo novo (item 1) e a ligação fica como pendência registrada no `status.md` até a 2.6 da 106 mesclar. Não editar em paralelo os mesmos arquivos.
5. **Só specs novas.** Specs existentes não são alteradas retroativamente.
6. **Proposta opcional, decisão do dono, fora desta task:** novo `P-NN` no `internal/gate` que reprova spec de feature ou de fix sem a seção. Custo: muda `internal/gate` (verificação nova e modo `observacao` antes de `bloqueio`), testes positivos e negativos, e a lista de IDs de P-xx documentada; specs antigas precisariam de isenção (data de corte). Recomendação: **não** implementar agora; observar por uma onda se Builder e Reviewer conferem sem o gate, e só então propor (`GOV-01`).
7. **Consolidação no release (Regra 12):** quem fecha a versão (hoje o checklist de release) coleta a seção das specs entregues desde a versão anterior, junta os itens em `history.md` na ordem de relevância e ajusta o texto ao que foi de fato mesclado. A Constituição não muda; a Regra 12 já exige destaques reais, a seção só garante que o rascunho existe.
8. **Orçamento (P-24):** Constituição (9189 de 9216 bytes) e `SKILL.md` (2045 de 2048) não crescem. O arquivo novo e as três frases ficam nas references (não há grupo de P-24 para elas, mas há o teste de conjunto de leitura `TestContextBudget_ConjuntoDeLeituraPorComando`: manter o teto). Goldens regravados no mesmo PR (`TST-02`).

## Modelo do módulo
**Sem módulo Go novo.** Texto de references e templates do scaffold, mais goldens. Se o Builder precisar listar o arquivo novo no scaffold, só registra o caminho na lista de templates existente (`ARQ-01`).

## Decisões do projeto aplicadas
ADR-0003 (skills e references enxutas, orçamento de contexto), ADR-0021 (3 sessões; vem da 1.1), ADR-0002 (estado em `sdd/`). Nenhuma decisão nova; a decisão do dono (seção em toda spec nova) está no `status.md`.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Template com a seção | `internal/scaffold/templates`, `.agents` | Nenhum | **Criar** `descricao-de-negocio.md` | pedido do dono |
| Lista de seções obrigatórias | `sdd-concepcao/SKILL.md` regra 4 | 2045 de 2048 bytes | **Não tocar** | `MaxSkillBytes` |
| Contagem "cinco" vs "seis" | `sdd-planejamento/SKILL.md` e references de concepção | Divergem | **Registrar**; corrigir fica com a 106 2.6 ou item à parte | `DUP-06` |
| Passos dos comandos | references de `nova-feature`, `novo-fix`, `split-features` | Reescritos pela 106 2.6 | **Só apontar**, depois da 2.6 | evita colisão |
| Release notes | `sdd/releases/history.md`, Regra 12 | Formato de entradas em negrito, comando e iniciativa | **Reusar** o tom | Regra 12 |
| Spec com a seção | `feat-105-08` | Exemplo manual | **Reusar** como modelo | `DUP-01` |
| Checagem por gate | `internal/gate` | Sem P-NN para seção de release | **Não criar agora** (item 6) | `GOV-01` |
| Orçamento de leitura | `internal/scaffold/budget_test.go` | Teto do conjunto de leitura por comando | **Respeitar** | `TST-02` |

## Dependências novas
Nenhuma (`SEG-07`).

## Riscos
- Colisão de arquivos com a 106 2.6 e a 107: mitigada pelo ponto único e pela ordem do item 4.
- Seção copiada sem revisão vira promessa falsa: mitigada pela conferência do Reviewer.
- Aumento do conjunto de leitura dos comandos: o arquivo novo só é lido por quem escreve spec (a verificar no teste de orçamento).

## Rules e IDs da Constituição Aplicáveis
`COD-05`, `DUP-01`, `DUP-06`, `GOV-01`, `GOV-03`, `TST-02`. Rules: `sdd-codigo-e-testes`, `sdd-reuso-e-duplicacao`, `sdd-fluxo-e-portao`.

## Critérios de Aceitação Executáveis
Testes usam `t.TempDir()`; não dependem da identidade do git nem do `forge` do PATH.
```bash
set -o pipefail
go build ./... && go vet ./...
go test -list 'TestReferenciaDescricaoDeNegocioRenderiza' ./internal/scaffold | grep -F 'TestReferenciaDescricaoDeNegocioRenderiza'
go test ./internal/scaffold -run 'TestReferenciaDescricaoDeNegocioRenderiza'
# negativo: SKILL.md de concepção e Constituição não ultrapassam o orçamento
go test -list 'TestSkillEConstituicaoDentroDoOrcamento' ./internal/scaffold | grep -F 'TestSkillEConstituicaoDentroDoOrcamento'
go test ./internal/scaffold -run 'TestSkillEConstituicaoDentroDoOrcamento'
# negativo: o conjunto de leitura de /discovery e do Builder continua dentro do teto
go test -list 'TestContextBudget_ConjuntoDeLeituraPorComando' ./internal/scaffold | grep -F 'TestContextBudget_ConjuntoDeLeituraPorComando'
go test ./internal/scaffold -run 'TestContextBudget_ConjuntoDeLeituraPorComando'
# negativo: specs existentes não são alteradas (só o arquivo novo e as references apontam para a seção)
go test -list 'TestSpecsExistentesIntactas' ./internal/scaffold | grep -F 'TestSpecsExistentesIntactas'
go test ./internal/scaffold -run 'TestSpecsExistentesIntactas'
go test ./...
```
- Goldens regravados no mesmo PR (`TST-02`).
- O Builder registra no `status.md` a ligação pendente, se a 106 2.6 ainda não tiver mesclado.
