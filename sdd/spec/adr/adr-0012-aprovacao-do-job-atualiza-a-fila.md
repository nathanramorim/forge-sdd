# ADR-0012: A Aprovação do Job Atualiza a Fila da Iniciativa

- **Data:** 2026-11-15
- **Status:** `aceito` (2026-11-20, aprovado pelo dono no Gate 1 da 105)
- **Regras afetadas (IDs da constitution):** `ARQ-01`, `SEG-01`, `SEG-06`, `DUP-01`, `GOV-01` (este ADR; nenhum ID foi criado ou removido)
- **Autor / Responsável:** Líder (`fix-84-02`)

## Contexto

Na iniciativa `55e0-credenciais-convites-midia-google-calendar`, o Reviewer aprovou o job e a notificação saiu, mas `forge status` e `forge queue list` seguiram em 0/6 e as tasks continuaram `queued`. O ramo `conforme` de `ApplyVerdict` só notificava e completava o job no barramento; ninguém atualizava o `queue.json`. Além disso, `queue.Load/Save` só enxergam a iniciativa da **branch atual**, e Builder e Reviewer podem estar em outra. O auto-archive (ADR-0009) nunca disparava, pois exige todas as tasks `approved`. Referência: `sdd/discovery/84-…/fixes/fix-84-02-fila-defasada-apos-aprovacao-por-job.md`.

## Opções Consideradas

1. **O agente atualiza cada task com `forge queue review`:** depende de o modelo lembrar e contradiz o princípio de que o Go decide o fluxo; foi o que falhou.
2. **Derivar o status na leitura a partir dos jobs:** a fila ficaria sempre certa, mas cada painel e comando teria de cruzar jobs e tasks.
3. **O Go propaga o resultado no momento da decisão:** uma escrita, no ponto onde o veredito já é aplicado.

## Decisão

- **D-1:** `ApplyVerdict` (`conforme`) marca as tasks de `job.TaskIDs`, da iniciativa `job.Initiative`, como `approved` **antes** do `bus.Complete`, para que uma falha seja reaplicável (a notificação é idempotente).
- **D-2:** `forge pipeline pr-opened` marca as tasks do job como `review_pending`.
- **D-3:** `queue.SetTaskStatuses` abre o `queue.json` da iniciativa **pelo nome** (validado por `layout.Initiative`, `SEG-01`), é idempotente, nunca regride uma task `approved` e trata fila ausente ou iniciativa inválida como "nada a fazer".
- **D-4:** o Go guarda uma cópia do veredito em `sdd/discovery/<id>/verdicts/<job>.json`, para o registro não ficar só em `/tmp`. A cópia é do Go, não do agente, para o Reviewer continuar somente leitura (`SEG-06`).

## Consequências

- **Positivas:** fila, painel e auto-archive acompanham a aprovação; o Reviewer segue sem escrever no repositório.
- **Negativas / Trade-offs:** violação e escalação não alteram o status das tasks (comportamento mantido); o `status.md` da iniciativa continua sendo atualizado pelo Líder, não pelo Go.
- **Ações de acompanhamento:** avaliar se violação deve marcar `changes_requested` automaticamente; tasks já defasadas em iniciativas antigas precisam de ajuste manual.
