# ADR-0021: Obsolescência das 5 Estações como Visão do Fluxo

- **Data:** 2026-11-20
- **Status:** `aceito` (20/11/2026, Gate 1 da iniciativa 105, aprovado pelo dono)
- **Regras afetadas (IDs da constitution):** `ARQ-06` (citada; a lista de subcomandos não muda aqui), `FLX-05`, `GOV-01` (esta ADR). Nenhum ID é criado ou removido.
- **Autor / Responsável:** Concepção; decisão do dono no Gate 1 de 2026-11-20

## Relação com ADRs vigentes
Supersede **parcialmente** a ADR-0003 (a visão de "skills por estação" como fluxo). Mantém ADR-0004 (3 sessões), ADR-0005 e ADR-0002.

## Contexto
O produto virou 3 sessões (Discovery, Builder, Reviewer) mais aprovação humana (ADR-0004), mas a ADR-0003 ainda descreve o fluxo em 5 estações e o CLI expõe `queue stations`, `attach` e `queue handoff`. `ARQ-06` exige ADR para depreciar subcomando. Fonte: `sdd/discovery/105-experiencia-cli-e-harness/obsoletos.md` (O2 a O7).

## Opções Consideradas
1. **Remover tudo já:** limpa o código, mas quebra `forge up` (o `lider` é o agente da sessão Discovery) e usuários de ACP.
2. **Só avisar e manter:** sem quebra; remoção futura fica em ADR própria. Custo: dívida temporária.

## Decisão
Opção 2. Nada é removido nesta ADR.
- (a) As 5 estações deixam de ser o fluxo. As skills `sdd-*` ficam como apoio das 3 sessões.
- (b) Depreciar (só avisar, com mensagem): `queue stations`, `attach`, `queue handoff`, `/implement` e `/proxima-feature`. O13 (pareceres antigos da 97) também só avisa.
- (c) O agente `lider` fica; só muda a descrição (para Discovery/3 sessões). Renomear exige teste e ADR.
- (d) Remover qualquer item de (b) exige ADR própria; remover subcomando altera a lista de `ARQ-06`.

## Consequências
- **Positivas:** o texto e o CLI passam a refletir as 3 sessões sem risco de quebra.
- **Negativas / Trade-offs:** código e comandos obsoletos seguem até a ADR de remoção.
- **Ações de acompanhamento:** tasks 1.2, 1.3 e 1.6 da 105 aplicam os avisos e o texto; anotar a supersessão parcial no cabeçalho da ADR-0003.
