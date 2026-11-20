# ADR-0014: Esteira Econômica e Autônoma

- **Data:** 2026-11-17
- **Status:** `aceito` (2026-11-20, aprovado pelo dono no Gate 1 da 105)
- **Regras afetadas (IDs da constitution):** `FLX-05`, `GOV-01` (esta ADR; nenhum ID foi criado ou removido)
- **Autor / Responsável:** Builder (iniciativa `90-esteira-economica-e-autonoma`, `feat-90-01`)

## Contexto

A esteira gasta turnos e tokens com trabalho mecânico: o painel mostra 869 mil tokens no Builder enquanto os logs somam 13,6 milhões relidos; o Builder abre o PR à mão e erra a base (PR #87 contra `main`); o archive copia em vez de mover; `forge status` custa 6,2 KB a cada comando; Builder e Reviewer usam sempre o modelo padrão. A iniciativa 90 muda contratos estruturais, e `FLX-05` e `GOV-01` exigem ADR antes do código.

## Opções Consideradas

1. **Manter o fluxo e só ajustar prompts:** barato, mas depende de o agente lembrar cada passo.
2. **Mover cada passo mecânico para o Go e medir o uso real:** o Go decide e valida; o agente só implementa.

## Decisão

- **D-1 (uso real):** o uso do job vem do evento `result` do log `stream-json` (tokens, cache, turnos, duração, modelo, custo de tabela). Sem `result`, grava a estimativa por bytes com `medicao: estimado`. Substitui D-1 da ADR-0009.
- **D-2 (PR pelo Go):** `forge pipeline pr` roda a Camada A, faz push com refspec explícito, abre ou corrige o PR e o registra. Nunca mescla, aprova nem empurra para `main`. Estende a ADR-0004.
- **D-3 (base da iniciativa):** a base fica na fila (`base_branch`); ordem: flag, fila, origem da branch no git, `base_branch` do `forge.json`. Um resolvedor único.
- **D-4 (archive):** `forge archive` move a pasta (idempotente, só na branch da iniciativa e sem pendência). Ao aprovar o último job, arquiva sozinho só com `origem: toca-ai`; nas demais, avisa "pronta para arquivar". Substitui D-4 da ADR-0009; o parecer segue a ADR-0012.
- **D-5 (modelo e esforço):** `forge.json` ganha `models` e `efforts` por papel; os perfis aceitam `{model}` e `{effort}`, como `{permissions}`. Sem escolha, o argv não muda. Estende a ADR-0005.
- **D-6 (status curto):** sem terminal, `forge status` sai curto (até 1,5 KB); `--full` mostra tudo. Vale nos dois modos (ADR-0013).

## Consequências

- **Positivas:** número real em vez de estimativa; menos turnos do Builder; PR sempre contra a base certa; leitura inicial barata.
- **Negativas / Trade-offs:** só o `claude` emite `result`; as demais CLIs seguem estimadas. Flags de modelo de outras CLIs só entram quando confirmadas.
- **Ações de acompanhamento:** atualizar `docs/pipeline.md` e `.forge/LEIA-ME.md`; confirmar flags de agy, gemini, codex e copilot (feat-90-07). ADRs afetadas: 0004, 0005, 0009, 0012, 0013. Nenhum ID da Constituição criado ou removido.
