# ADR-0024 — Orquestração por iniciativa e tier de modelo

- **Status:** Proposta
- **Data:** 2026-11-22

## Contexto

O barramento é compartilhado no projeto, enquanto status e demanda são por iniciativa. Workers persistentes podem atender jobs alheios ou permanecer vivos após a fila visível esvaziar. Além disso, o modelo fixo por papel (ADR-0014) não expressa a complexidade de cada task.

## Decisões propostas

1. Cada worker recebe uma iniciativa, reivindica somente jobs dela e encerra após drenar Builder, Reviewer e rework dessa iniciativa. Jobs ativos concluem normalmente.
2. PM-Architect registra complexidade, justificativa e tier lógico por task. O job transporta o tier; o worker resolve o modelo usando configuração específica da CLI, com o modelo estático por papel como fallback para jobs legados.
3. Tiers são `fast`, `balanced` e `reasoning`. Tasks claras usam tiers econômicos; raciocínio avançado é reservado ao Advisor quando a incerteza justificar. Não há fallback entre provedores.
4. O Builder recebe a spec aprovada; Reviewer recebe código somente após Camada A e registro do PR.
5. Tmux é uma opção de interface e não requisito para continuidade do pipeline.

## Consequências

O runtime precisa transportar identidade da iniciativa e tier opcional do job. A configuração por CLI valida mapeamentos/argumentos suportados. As sessões em andamento não mudam; a política vale para jobs futuros.

## Alternativas consideradas

- Manter workers globais persistentes: rejeitado porque diverge do escopo exibido por iniciativa.
- Usar sempre modelo de raciocínio: rejeitado por custo e por não refletir tasks já bem delimitadas.
- Remover tmux como pré-requisito: rejeitado; sua retirada é opcional e pode ser entregue separadamente.
