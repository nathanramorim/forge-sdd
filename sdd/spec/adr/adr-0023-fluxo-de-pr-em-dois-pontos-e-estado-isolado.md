# ADR-0023: Fluxo de PR em dois pontos, branch por feature e estado isolado

- **Data:** 2026-11-21
- **Status:** `aceito` (complementa a ADR-0022; decisões do dono nos Gates 1 de 20 e 21/11/2026)
- **Regras afetadas (IDs da constitution):** `ARQ-06`, `FLX-01`, `FLX-02`, `FLX-03`, `FLX-06`, `FLX-07`, `GOV-03`, `GOV-04`
- **Autor / Responsável:** Concepção e Builder da iniciativa 106, aprovado pelo dono

## Contexto

A ADR-0022 fixou a política (PR em draft, modos do projeto, spec que volta à principal, branch por feature) e concentrou na Constituição só o essencial. Faltava a ADR que descreve **como o fluxo se comporta em execução**, para que os comandos, skills e workflows tenham uma referência única de comportamento — e não seis descrições divergentes espalhadas.

A implementação das 20 tasks da 106 (ondas 0 a 3) mostrou três pontos que precisavam de decisão registrada:

1. **Quando o CI roda.** Rodar em todo push de branch consumiu runner com trabalho ainda não conferido, e o GitHub chegou a bloquear a conta por cobrança.
2. **Quem pode tirar o PR do draft.** Nada conferia isso; o `gh pr ready` e o `gh pr merge` passavam sem parecer.
3. **Onde vive o estado da fila.** Um `queue.json` compartilhado gera conflito de merge quando duas features evoluem em paralelo.

## Opções Consideradas

1. **CI em todo push + estado compartilhado em `queue.json` + marcar pronto liberado:** simples de implementar, mas repete o custo de runner, permite PR sem revisão e trava o paralelismo por conflito de merge.
2. **Fluxo em dois pontos com sinal determinístico, CI só fora do draft e estado por feature:** cada um dos três problemas vira uma regra verificável, com uma prova executável por regra.

## Decisão

Escolhida a **Opção 2**:

### 1. Dois pontos de PR, nunca contínuo
- **PR da spec:** nasce em draft na concepção. Sai do draft depois da **Conferência 1** (point `spec`) e do **`[Seguir]`** do humano; então é mesclado na principal. A branch da spec some.
- **PR de cada feature:** nasce em draft quando o Builder abre a branch, a partir da principal. Sai do draft depois da **Conferência 2** (ponto `entrega`) e do **`[Continuar]`**. Com checks verdes, a decisão de merge é humana.
- **Nenhum PR vive em draft "até o fim" de toda a iniciativa.** O sinal (`spec-para-revisao`, `pronto-para-revisao`) é determinístico e fica no `status.md`; a conferência só se repete quando o sinal do ponto se repetir.

### 2. CI no ponto certo
- Job `pronto?` é obrigatório na principal e recusa PR em draft.
- Teste e lint rodam em `push` na principal e em PR **fora de draft**; em draft os jobs ficam `skipped`.
- `concurrency` por PR cancela a execução anterior ao subir um commit novo.
- Verificação do atestado (`scripts/atestado.py`) roda só no PR de feature, no modo `local`, em segundos.

### 3. Estado isolado por feature
- Cada feature tem `sdd/discovery/<id>/state/<sid>.json` com `schema`, `task_id`, `status`, `attempt`, `head_sha` e o parecer.
- `queue.json` passa a ser **plano**, não estado: nada nele muda por execução, então não conflita.

### 4. Prova antes de pronto e antes de merge
- `gh pr ready` e `gh pr merge` exigem parecer **`conforme` com `head_sha` igual ao HEAD atual**. Parecer de outro commit não vale (o código mudou depois da revisão).
- A barreira vale no Claude Code (hook `PreToolUse`); nos outros agentes, o job `pronto?` do CI é quem segura.
- O modo do atestado vem da Constituição: `desligado` diz em voz alta que nada foi conferido; `vcs` é recusado como adiado.

## Consequências

- **Positivas:** o custo de runner acompanha o valor entregue; PR não sai de draft sem revisão válida para aquele commit; duas features paralelas não conflitam em estado; cada regra do fluxo tem um teste executável.
- **Negativas / Trade-offs:** mais peças para manter (hook, job, script, sinal). O hook só vale no Claude Code; Gemini e Copilot dependem do CI.
- **Ações de acompanhamento:** as tasks 3.5 e 3.7 da 106 (prova por VCS e primeira leva de repositórios) ficaram fora do escopo reduzido; quando voltarem, esta ADR é a referência.
- **Relação com a ADR-0022:** a 0022 fixa a política e os campos da Constituição; esta registra o comportamento em execução. Nenhuma das duas edita a Constituição por si — a edição veio na task 0.8, num PR único.