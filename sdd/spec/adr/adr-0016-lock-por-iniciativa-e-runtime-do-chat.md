# ADR-0016: Lock por Iniciativa e Estado do Laço em `.forge/runtime/chat`

- **Data:** 2026-11-20
- **Status:** `proposto`
- **Regras afetadas (IDs da constitution):** `ARQ-01`, `SEG-01`, `FLX-04`, `GOV-01` (esta ADR). Nenhum ID é criado ou removido.
- **Autor / Responsável:** Concepção (alinhamento da Constitution); decisão do dono pendente

## Relação com ADRs vigentes
- **ADR-0015: detalha** a decisão D-2 (Go dono do laço, lock por iniciativa).
- **ADR-0004: convive.** O barramento em `.forge/runtime/<papel>/<estado>/` não muda.

## Contexto
No modo `chat`, mais de um coordenador (chat, subagente, processo) pode tentar conduzir a mesma iniciativa. Sem exclusão mútua, dois laços gravariam o estado e o `queue.json` ao mesmo tempo, e um processo morto deixaria a iniciativa travada.

## Opções Consideradas
1. **A. Lock em arquivo com lease por iniciativa**, criação atômica e recuperação após expirar.
2. **B. Sem lock**, confiando em quem chama.
3. **C. Lock do sistema (`flock`)**, que não é portável ao Windows, alvo da distribuição.

## Decisão
- **D-1:** adotar **A**: `harness.AdquirirLock` cria `<iniciativa>.lock` em `.forge/runtime/chat/locks` (`pipeline.dirLocks`), com dono, PID e `Expira`; criação atômica por `Link` sem sobrescrever; `Renovar` e `Liberar` conferem a posse (`ErrLockPerdido`).
- **D-2:** lock expirado só é tomado por `os.Rename` (um processo vence); nada é apagado antes da lease (15 min em `leaseLaco` e `leaseLeque`).
- **D-3:** o id da iniciativa e o dono são validados e o caminho nunca sai do diretório de locks (`SEG-01`).
- **D-4:** o estado do laço fica em `.forge/runtime/chat/`, gravado só pelo Go.

## Consequências
- **Positivas:** uma iniciativa, um coordenador; recuperação sem intervenção humana.
- **Negativas / Trade-offs:** coordenador lento que não renova perde a posse e o laço devolve `retornoLockPerdido`; depende de relógio local.
- **Ações de acompanhamento:** não verificado: comportamento em sistemas de arquivos de rede.
