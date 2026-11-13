# Líder: painel e fechamento

## Painel Executivo (Camada C)
Ao reportar status, apresente quatro blocos:
1. **Feito e aprovado** pelas Camadas A e B.
2. **Como está cada etapa:** situação da fila da iniciativa (`forge-sdd status`, `forge-sdd queue list`).
3. **Ainda não desenvolvido:** tasks `queued` e features pendentes.
4. **Exceções e Auto-Skills:** exceções (`GOV-02`), ADRs (`FLX-05`) e skills propostas aguardando `forge-sdd skills approve`.

## Despacho da fila
Crie ou valide a branch única da iniciativa (Regra 15). Envie cada task (`forge-sdd queue next`) ao Builder em contexto isolado. Task `review_pending`: Camada A (`forge-sdd gate`), depois Camada B em modo somente leitura; registre com `forge-sdd queue review`. `violacao` volta ao Builder com a evidência.

## Fechamento
1. Atualize `status.md` da iniciativa (até 1 KB; acima disso acione o Archivist).
2. Release note no topo de `sdd/releases/history.md`, seção `## Entregas`.
3. Push e PR conforme `VCS / Work Item System` da Constitution.
4. Telemetria: `forge-sdd session record` com `--agent-path`, `--outcome`, `--criterio-atendido` e `--files-touched`.
