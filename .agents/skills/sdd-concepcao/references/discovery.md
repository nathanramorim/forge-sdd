# Comando: discovery

**Uso:** `/discovery <descrição da demanda>`.

1. **Brief primeiro:** rode `forge brief` no início; se o estado for `rascunho`, proponha `/constitution` antes de qualquer spec. Inspecione `forge status`: se houver tasks em `queued` ou `review_pending`, recomende priorizar o handoff para a esteira (Builder/Reviewer).
2. **Clarify:** aplique `sdd/memory/clarify.md`.
3. **ID e branch da spec (`FLX-02`, Regra 15, ADR-0022):** rode `forge next-id` e crie a branch `feat/<id>-<nome>`, **a partir da principal** (pergunte a branch de partida, padrão `main`; se já existir branch da iniciativa, pergunte se retoma). Esta branch carrega **só a spec**; ela não é a branch da construção — o Builder abre uma branch por feature depois do split.
4. **ADRs e rules:** antes de propor qualquer coisa, leia as ADRs e as rules aplicáveis ao tema; cite as decisões aplicadas na spec; ADR depreciada ou substituída não vale como base.
5. **Produto (`discovery.md`):** problema, para quem, jornada, regras de negócio, escopo (dentro e fora) e riscos.
6. **Arquitetura (`criteria.md`):** restrições citando IDs da Constitution; seções `Modelo do módulo`, `Decisões do projeto aplicadas`, `O que já existe` (tabela `Procurei | Onde | Achei | Decisão | Justificativa`) e `Dependências novas`; diagramas C4 em Mermaid (contexto, contêineres e sequência) e critérios de aceitação executáveis.
7. **Plano (`plan.md`):** quebra ordenada de baixo para cima (limite de 7 tasks por onda), dependências, paralelismo e rules/IDs por task.
8. **Estado:** crie `status.md` (até 1 KB) com `Status`, `Próximo passo` e `A confirmar`.
9. Se `Nível de Linguagem` na Constitution for `iniciante`, explique C4 e critério de aceitação com exemplos simples, sem alterar os critérios.
10. **Telemetria:** `forge session record --sessao spec --feature "sdd/discovery/<id>-<nome>/discovery.md" --tokens-input <N> --tokens-output <N> --outcome approved --criterio-atendido=true --files-touched <lista>`. Nunca passe zeros: se não souber os tokens, omita as flags (o CLI vai estimar).

**Visualizar artefatos (antes do Gate 1)**

Use AskUserQuestion com as opções:
- "Abrir pasta no Finder" → `open sdd/discovery/<id>-<nome>/`
- "Abrir no VS Code" → verifique `command -v code && code sdd/discovery/<id>-<nome>/` (fallback: Finder se VS Code ausente)
- "Ver lista dos arquivos" → liste os arquivos em `sdd/discovery/<id>-<nome>/`
- "Pular"

Execute a ação escolhida. Consulte `sdd/memory/clarify.md` → seção "Visualização pós-spec" para comportamento detalhado (macOS-only, fallbacks).
11. **Gate 1 (`FLX-01`):** submeta ao humano; após aprovar, sugira `/split-features`. Em modo pipeline (existe `.forge/config/forge.json`) esta sessão nunca constrói: o `/split-features` termina entregando ao Builder.

## PR da spec em draft e a volta à principal (ADR-0022, RN-1)

A spec **não fica amarrada numa branch de iniciativa** até a implementação acabar. O ciclo dela é:

1. **Abra o PR da spec em draft**, logo depois de escrever os artefatos:
   ```bash
   git add sdd/discovery/<id>-<nome> && git commit -m "docs(<id>): spec da iniciativa <nome>"
   git push -u origin feat/<id>-<nome>
   gh pr create --draft --base main --head feat/<id>-<nome> --fill
   ```
   O PR nasce **sempre em draft** (`RN-7`); em draft o CI não roda.
2. **Grave o sinal** `spec-para-revisao` no `status.md` (o Reviewer só é acionado por ele) — `RN-4`.
3. **Conferência 1** (`RN-3`): o Reviewer confere **só o julgamento** — spec contra ADRs, rules, Constituição e padrão. O determinístico (`forge gate`) já rodou local.
4. **`[Seguir]` ou `[Ajustar]`** (`FLX-01`): o humano responde no Gate 1.
   - `[Seguir]`: grave `Gate 1: aprovado` no `status.md`, marque o PR como pronto (`gh pr ready`) e **mescle a spec na principal**. A branch da spec some.
   - `[Ajustar]`: volte ao passo 5 com o parecer; o sinal `spec-para-revisao` se repete quando a spec voltar a estar pronta.
5. **Depois do merge**, a construção começa da principal: o `/split-features` (do Builder) fatia o `plan.md` aprovado.

> **Por que dois PRs e não um:** o PR da spec é documentação e volta cedo; o PR de cada feature carrega código e só sai do draft depois da Conferência 2. Nenhum PR vive em draft "até o fim" — `RN-1`.

**Nunca** crie aqui a branch de construção (`feat/<id>-<nome>-<sid>`): ela é do Builder, uma por feature do split, cada uma com o seu PR em draft.
