# ADR-0001: Remoção do Forge Studio

- **Data:** 2026-11-13
- **Status:** `aprovado`
- **Regras afetadas (IDs da constitution):** `ARQ-06`, `DUP-01`, `COD-04`
- **Autor / Responsável:** Líder (decisão do dono do produto no discovery 69)

## Contexto

O Forge Studio (`internal/studio`, cerca de 5,2 mil linhas Go mais `three.min.js` embutido) acompanhava as 5 estações por interface web. O dono do produto não o utiliza, e ele aumentava o contexto lido pelos agentes (seção 8 do `FLOW.md`, README, rule ACP, papel do Líder) e a manutenção (feats 59, 63, 66 e 67). O único chamador de `StartHeartbeat` e da higienização de atividades órfãs era o servidor do Studio. Referência: `sdd/discovery/69-otimizacao-estrutura-tokens-e-paralelismo/` e `feat-69-01`.

## Opções Consideradas

1. **Opt-in fora do padrão:** manter o código e esconder do fluxo. Reversível, mas mantém 5 mil linhas, testes e assets sem uso.
2. **Remover o Studio e manter o ACP:** apaga a interface; o protocolo e as sessões físicas (feat 68) seguem.
3. **Remover Studio e ACP:** volta ao fluxo só por CLI e fila; inviabiliza a feat 68.

## Decisão

Opção 2. `internal/studio`, `cmd/forge-sdd/studio.go` e `docs/manual-forge-studio.md` foram removidos, junto com as citações em README, `FLOW.md`, rule ACP, papel do Líder, templates e goldens. O servidor `forge-sdd acp` passa a higienizar atividades órfãs e a manter o heartbeat do pool. As tasks 68-04 e 68-05, que existiam só para o Studio, foram canceladas. Nenhum ID da Constitution foi criado ou alterado; `ARQ-06` já não listava `studio`.

## Consequências

- **Positivas:** menos código e menos contexto; um conceito a menos no catálogo (`DUP-01`); o ACP deixa de depender de uma interface.
- **Negativas / Trade-offs:** perde-se a visão gráfica do fluxo; quem a usava passa a `forge-sdd queue stations` e, após a 69-04, a `forge-sdd status`. `acp.HTTPHandler` ficou sem consumidor dentro do repositório.
- **Ações de acompanhamento:** decidir na 69-07 se `acp.HTTPHandler` e o transporte HTTP/SSE continuam como contrato público ou saem (`COD-04`); `sdd/spec/concepts.md` já foi corrigido (`acp.HTTPHandler` no lugar de `acp.HTTPTransport`).
