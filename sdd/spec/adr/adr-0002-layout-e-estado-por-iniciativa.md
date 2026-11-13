# ADR-0002: Layout e Estado por Iniciativa

- **Data:** 2026-11-13
- **Status:** `aprovado`
- **Regras afetadas (IDs da constitution):** `FLX-02` (Regras 13 e 15), `FLX-04`, `DUP-01`, `DUP-04`, `GOV-01` (este ADR; o texto das Regras 13 e 15 e de `FLX-02` na Constitution deste repositório é atualizado na feat 69-07)
- **Autor / Responsável:** Líder (decisão do dono do produto no discovery 69)

## Contexto

`progress.md` (154 commits), `index.md` (131) e `sdd/.queue/*.json` (96 KB) eram arquivos únicos tocados por toda branch, o que gerava conflito constante entre frentes paralelas. Além disso, só as iniciativas 66 a 68 usavam a pasta `sdd/discovery/<id>-<nome>/`; restavam 14 discoveries soltos (42 arquivos) e 115 arquivos em `sdd/features/`, e o contexto lido a cada comando incluía um `index.md` de 27 KB quase todo histórico. Referência: `sdd/discovery/69-otimizacao-estrutura-tokens-e-paralelismo/` e `feat-69-02`, `feat-69-03`, `feat-69-04` e `feat-69-05`.

## Opções Consideradas

1. **Branch dedicada de estado:** índice, progresso e fila numa branch separada, acessada por worktree. Tira o estado das branches de feature, mas torna o fluxo dependente de worktree.
2. **Manter globais com merge `union`:** menor mudança, mas não resolve o JSON da fila nem o crescimento do índice.
3. **Estado por iniciativa:** cada pasta de iniciativa guarda `status.md` e `queue.json`; a visão geral é calculada na leitura.

## Decisão

Opção 3.
- Uma pasta por iniciativa: `sdd/discovery/<id>-<nome>/` com `discovery.md`, `criteria.md`, `plan.md`, `status.md` (até 1 KB), `queue.json`, `features/` e `fixes/`. A branch `<prefixo>/<id>-<nome>` aponta para essa pasta (`internal/layout`, dono único dos caminhos).
- Estado de execução (atividade e filas por estação) vai para `sdd/.runtime/`, fora do git: o scaffold cria `sdd/.runtime/.gitignore` (ignora tudo menos ele mesmo), de modo que um projeto novo nasce sem commitar esse estado.
- A visão geral e o próximo ID são calculados: `forge-sdd status [--all-branches]` e `forge-sdd next-id` leem todas as branches sem checkout.
- `index.md` e `progress.md` globais saem do fluxo. O check `P-21` passa a medir o `status.md` da iniciativa ativa; sem iniciativa, cai no `progress.md` legado se existir.
- **Migração:** `forge-sdd update` move discoveries soltos e `sdd/features/` para `sdd/archive/` (com backup em `sdd/backups/`, sem sobrescrever, recusando links simbólicos) e remove `.agents/commands/` e `.agents/roles/` apenas quando idênticos ao template anterior, preservando o que o usuário editou. `forge-sdd update --dry-run` imprime o plano sem gravar.
- Nenhum ID da Constitution foi criado ou removido; mudam as Regras 13 e 15 e o texto de `FLX-02` (caminhos), registradas aqui conforme `GOV-01`.

## Consequências

- **Positivas:** branches paralelas não tocam o mesmo arquivo de estado; menos contexto por comando; projetos atualizados ganham o layout sem perder conteúdo.
- **Negativas / Trade-offs:** `queue.json` e `progress.md` ainda conflitam ao mesclar com branches anteriores a esta mudança (a versão da 69 deve vencer); `internal/layout` acumulou caminhos, leitura de refs e cálculo de ID, e pode ser dividido em pacote próprio de leitura de estado; o `P-21` em HEAD destacado (CI, worktree de revisão) mede o legado, não o `status.md`; a limpeza de obsoletos só reconhece arquivos idênticos ao template da versão imediatamente anterior.
- **Ações de acompanhamento:** registrar `internal/layout` em `sdd/spec/concepts.md` e atualizar as Regras 13 e 15 e `FLX-02` na Constitution (feat 69-07); tratar `internal/config/sddrc_io.go` da feat 68 ao mesclar, trocando caminhos manuais por `internal/layout`.
