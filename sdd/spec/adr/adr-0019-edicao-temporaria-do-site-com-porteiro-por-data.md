# ADR-0019: Edição Temporária do Site com Porteiro por Data

- **Data:** 2026-11-20
- **Status:** `aceito` (20/11/2026, aprovado pelo dono no chat)
- **Regras afetadas (IDs da constitution):** `GOV-01` (esta ADR). Nenhum ID é criado ou removido; fora do binário Go.
- **Autor / Responsável:** Concepção; a iniciativa 101 (Edição Missões) já está no `main`

## Relação com ADRs vigentes
Nenhuma. Não toca o binário nem `ARQ-04`.

## Contexto
A LP (`docs/`) ganhou uma camada visual temporária (Edição Missões, tema da v5). Ela precisa desligar sozinha em data certa, sem deixar marca de terceiros e sem mexer na LP base.

## Opções Consideradas
1. **A. Camada isolada com porteiro por data** em um único arquivo. 2. **B. Branch ou deploy separado.** 3. **C. Remoção manual.**

## Decisão
- **D-1:** adotar **A**: `docs/edicao/edicao.js` guarda `id` e `fim` (hoje `missoes`, `2026-11-23T00:00:00-03:00`); só ativa se `fim` for RFC 3339 válida e `agora < fim`; `?edicao=on|off` força para teste.
- **D-2:** ativa, marca `data-edicao` no `<html>`; ativos em `docs/edicao/` (CSS, mascotes, SVGs) e tokens em `docs/design/edicao-missoes/`.
- **D-3:** `.github/workflows/encerrar-edicao.yml` roda diariamente (cron) e confere a vigência pelo mesmo parâmetro; se vencida, encerra por PR (a remoção em si: não verificado).

## Consequências
- **Positivas:** fim determinístico; LP base intacta. **Negativas:** o relógio é o do navegador; a data em dois lugares (JS e workflow) leria a mesma fonte, mas o workflow não foi executado nesta sessão.
