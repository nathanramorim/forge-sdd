# ADR-0009: Tokens por Janela, Reset Suave de Contexto e Auto-Archive

- **Data:** 2026-11-14
- **Status:** `aceito` (2026-11-20, aprovado pelo dono no Gate 1 da 105)
- **Regras afetadas (IDs da constitution):** `ARQ-01`, `COD-01`, `DUP-01`, `GOV-01` (este ADR; nenhum ID foi criado ou removido)
- **Autor / Responsável:** Líder (discovery 81)

## Contexto

Os painéis vivos do Builder e do Reviewer acumulavam o histórico de várias tasks: custo e latência subiam a cada mensagem e a precisão caía. Não havia como ver quanto cada janela já tinha consumido, e limpar o contexto à mão fazia o agente perder a task atual, a branch e onde parou. Referência: `sdd/discovery/81-tokens-por-janela-e-reset-suave/`.

## Opções Consideradas

1. **Fechar e reabrir o painel:** simples, mas quebra a disposição do tmux e perde o estado.
2. **Reset por limite de tokens automático:** exige estimar tokens com precisão que os CLIs não expõem; risco de reset no meio de uma task.
3. **Reset por job, no mesmo painel, com cartão de handoff:** previsível, in-place, e o custo de tokens fica visível num badge.

## Decisão

- **D-1 Tokens por janela:** `monitor.TokenMetricsPerRole` estima os tokens de Discovery, Builder e Reviewer; `FormatTokenBadge` mostra o consumo no `status-right` do tmux.
- **D-2 Reset suave por job:** na entrega do job, `deliverContextReset` envia `profile.ResetCommand` (`/clear` ou `/new`) ao mesmo painel, sem fechá-lo. O modo é `live.context`: `clear` (padrão) ou `continuous`. Perfil sem `reset_command` mantém o contexto contínuo e avisa.
- **D-3 Handoff:** `BuildHandoffCard` reinjeta o contexto mínimo (task, branch e onde parou) depois do reset.
- **D-4 Auto-archive:** ao aprovar um job, `CheckAndAutoArchive` arquiva a iniciativa se todas as tasks estiverem `approved`.

## Consequências

- **Positivas:** contexto limpo por task, consumo visível, sem flashes no tmux, retomada sem perda.
- **Negativas / Trade-offs:** a contagem de tokens é uma estimativa (`EstimateTokensFromBytes`); o reset é por job, não por limite.
- **Ações de acompanhamento:** o auto-archive depende de a fila refletir a aprovação (ADR-0012).
