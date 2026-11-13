# feat-58-05 — Estação do Analista de Governança (Camada B JSON) e Painel Executivo do Líder (Camada C)

**Branch:** `feat/58-orquestracao-autonoma-governanca-skills`
**Fase:** 58-05
**Depende de:** `feat-58-02`, `feat-58-03`, `feat-58-04`
**Status:** `done`

## Objetivo

Integrar a **Camada B (Analista de Governança Semântico)** ao motor da fila (`forge-sdd queue review`), validando estritamente o schema JSON de parecer com evidências (`S-01..S-07`), e atualizar os comandos `/proxima-feature`, `/revisar` e `/status` (`forge-sdd status`) para que o **Líder** entregue ao humano o painel consolidado da **Camada C** (*O que já foi feito*, *Como está cada etapa* e *O que ainda não foi desenvolvido*).

## Modelo do módulo (`ARQ-11`)
- **Modelo:** `completo` (validador de schema do parecer semântico em `internal/queue` + painel executivo de status + atualização dos comandos canônicos de execução e revisão).
- **Justificativa:** Implementa a validação da saída da IA (resposta inválida ou violação sem evidência falha na verificação) e governa o ciclo de devolução ao Builder ou aprovação para o Líder.

## O que já existe (`DUP-02`)

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
| --- | --- | --- | --- | --- |
| Comandos `/proxima-feature`, `/revisar`, `/status` | `.agents/commands/{proxima-feature,revisar,status}.md` | Comandos atuais de implementação, revisão e status prescritivo | `estender` | Atualizar para consumir a fila (`forge-sdd queue`), rodar o Portão Camada A + Camada B e exibir o painel executivo do Líder |
| Schema JSON da Camada B | `busize-solutions/planning-brain-governanca` (`portao/prompt-portao-github-action.md` Parte B) | Formato `{veredito, achados[{verificacao, regra, classificacao, severidade, arquivo, linha, evidencia, explicacao, correcao}], resumo}` | `reusar` / `estender` | Validar em Go no comando `forge-sdd queue review`, acrescentando o campo opcional `candidate_skill` |

## Dependências novas (`SEG-07`)
Nenhuma.

## Rules e IDs da Constituição Aplicáveis
- **Rules:** `.agents/rules/sdd-seguranca.md`, `.agents/rules/sdd-codigo-e-testes.md`, `.agents/rules/sdd-fluxo-e-portao.md`
- **IDs:** `S-01..S-07`, `FLX-03`, `FLX-06`, `FLX-08`, `SEG-04..05`, Regra 14.

## Critérios de Aceitação Executáveis

1. `forge-sdd queue review <task-id> --verdict-file <json>` (ou `--verdict-json`):
   - Valida o JSON contra o schema da Camada B (`veredito` ∈ `{"conforme", "violacao", "duvida"}`; todo item em `achados` com `classificacao == "violacao"` exige `regra`, `arquivo`, `linha > 0` e `evidencia` não-vazia — caso contrário, rejeita o parecer como inválido).
   - Se `veredito == "conforme"`, move a task para `approved`. Se `"violacao"`, move a task para `changes_requested` (devolvendo para a fila do Builder com os achados anexados).
2. `.agents/commands/proxima-feature.md` e `.agents/commands/revisar.md` (e seus `.md.tmpl`) orquestram o ciclo completo: Builder consome `queue next` em contexto isolado → `forge-sdd gate` (Camada A) → Analista de Governança (Camada B) → `forge-sdd queue review` → retorno ao Líder.
3. `.agents/commands/status.md` (e a saída da fila) apresenta ao humano as 4 seções claras do Líder:
   - **1. Concluído e Aprovado pela Governança (`approved`);**
   - **2. Em Andamento / Em Revisão (`building`, `review_pending`, `changes_requested`);**
   - **3. O que ainda não foi desenvolvido (`queued` / tasks `todo`);**
   - **4. Exceções vigentes e Auto-Skills propostas.**
4. Testes unitários cobrindo validação de JSON válido, JSON malformado e violação sem evidência; golden tests atualizados e `go test ./...` passando.

## Handoff
Depende de `feat-58-02`, `feat-58-03` e `feat-58-04`. Conecta-se com `feat-58-06` e `feat-58-07`.
