# ADR-0028: Acréscimo seguro dos campos de modo na Constituição do usuário

- **Data:** 2026-11-22
- **Status:** `aprovado` (aprovação do dono em 22/11/2026 para o acréscimo seguro, sem reescrita)
- **Regras afetadas (IDs da constitution):** `SEG-02`, `FLX-05` (a `GOV-01` rege só a Constituição deste repositório)
- **Autor / Responsável:** Builder (feat-118-03), dono do projeto

## Contexto

Projetos 5.0 não têm os cinco campos de modo (atestado, revisão humana, WIP humano, módulos permitidos, início automático) na `## Ferramentas e Integrações`. O `shouldPreserve` preserva tudo sob `sdd/`, então o `update` nunca os acrescenta.

## Opções Consideradas

1. **Reescrever a Constituição no `update`:** perde texto do usuário; viola `SEG-02`.
2. **Não fazer nada:** valem os padrões de `LeModos`, mas o usuário não vê nem edita os campos.
3. **Passo à parte que só acrescenta linhas faltantes:** seguro e idempotente.

## Decisão

Opção 3. `shouldPreserve` não muda: `Run`/`renderDir` nunca tocam `sdd/memory/constitution.md`. Depois de `Run`, o `update` chama `scaffold.AcrescentaCamposDeModo`, que insere na tabela da seção apenas as linhas que faltam, com os padrões de `vcs.CamposDeModoPadrao` (fonte única, `DUP-06`), por gravação atômica (`writeFileSafe`). Nunca altera linha existente (mesmo com valor inválido). Segunda execução não grava. Seção sem tabela: não grava e avisa em stderr. Seção ausente: criada ao fim. Nenhuma regra `ID` é alterada.

## Consequências

- **Positivas:** o usuário passa a ver e editar os campos sem perder nada.
- **Negativas / Trade-offs:** o `update` passa a gravar um arquivo de domínio do usuário (única exceção, restrita a linhas faltantes da tabela).
- **Ações de acompanhamento:** template e golden atualizados no mesmo PR (`TST-02`).
