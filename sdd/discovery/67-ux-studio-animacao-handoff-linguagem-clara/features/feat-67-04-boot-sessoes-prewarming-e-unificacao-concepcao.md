# Tarefa feat-67-04 — Pre-warming no Boot do Studio e Unificação Visual da Concepção & Planejamento

## 1. Contexto e Objetivo
Para proporcionar uma experiência fluida e sem atritos ao usuário desde o momento em que a aplicação é iniciada:
1. **Pre-warming de Sessões ACP no Boot:** As 5 sessões canônicas persistentes (`sess-01-lider`, `sess-02-pm`, `sess-03-planner`, `sess-04-builder`, `sess-05-governance`) devem ser pré-aquecidas e mantidas em estado de prontidão (`SessionStateStandby`) imediatamente na inicialização do servidor Studio e na configuração do despachante (`SetDispatcher`), garantindo que nenhum agente fique "frio" ou indisponível ao receber comandos.
2. **Unificação Visual da Concepção & Planejamento:** Apresentar de maneira integrada as etapas de Discovery/Arquitetura (PM & Arquiteto) e de Fatiamento (Planejador) sob a fase unificada de "Concepção & Planejamento", esclarecendo que o planejamento e quebra de tarefas é o desdobramento natural imediato após a aprovação humana do Discovery, sem criar confusão sobre onde a jornada se inicia.

## 2. O que já existe (Reuse Check — DUP-02)
- `internal/acp/pool.go`: `SessionPool`, `CanonicalPoolSessions`, `Warmup(ctx, wsDir)`.
- `internal/studio/server.go`: `NewServer` com integração ao `SessionPool`.
- `internal/studio/state.go`: `StationDTO` com `Name`, `Role`, `InputContract`, `OutputContract`.
- `internal/studio/assets/studio.js`: Renderização de cartões, badges e gaveta lateral de inspeção.

## 3. Critérios de Aceitação (Gherkin)
### Cenário 1: Pre-warming de todas as sessões no boot
- **Dado** que o servidor Studio é instanciado via `NewServer` ou tem seu despachante configurado
- **Quando** o pool de sessões for consultado
- **Então** `pool.IsWarmed()` deve ser verdadeiro
- **E** todas as 5 sessões canônicas devem estar registradas em estado `SessionStateStandby` com workspace validado.

### Cenário 2: Apresentação Unificada de Concepção & Planejamento
- **Dado** o snapshot gerado pelo Studio
- **Quando** as estações 02 (PM & Arquiteto) e 03 (Planejador) forem renderizadas
- **Então** os papéis e badges devem destacar claramente a fase unificada "Concepção & Planejamento" (Fase 1: Discovery & Arquitetura / Fase 2: Fatiamento de Tarefas)
- **E** a compatibilidade com a esteira técnica e contratos de fila deve ser estritamente preservada (sem quebra de IDs nem de portas).

## 4. Arquivos Afetados
- `internal/studio/server.go`
- `internal/studio/state.go`
- `internal/studio/assets/studio.js`
- `internal/studio/server_test.go`
- `internal/studio/assets_test.go`
