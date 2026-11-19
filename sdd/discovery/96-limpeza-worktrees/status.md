# Status — Iniciativa 96

**Iniciativa:** 96-limpeza-worktrees
**Branch:** feat/96-limpeza-worktrees
**Status:** spec_pronta — aguardando Gate 1 (aprovação humana)
**Próximo passo:** Líder revisa feat-96-01; se aprovada, enfileirar com `forge queue enqueue`.

## Features

| ID | Nome | Status |
|---|---|---|
| feat-96-01 | Detecção e limpeza de worktrees ociosos na inicialização da sessão | ready |

## A confirmar (decisões que só o usuário pode tomar)

- [ ] **Gate 1:** aprovação humana da spec (feat-96-01).
- [ ] **Jobs `failed`:** o usuário quer poder limpar worktrees de jobs que falharam (podem ser
      retentados), ou prefere sempre manter? A spec atual os mantém como `PROTECTED`.
- [ ] **Worktrees orphan** (sem job no barramento): limpar automaticamente se escolhidos, ou
      listá-los apenas para informação sem opção de remoção?
- [ ] **Remoção determinística no fim do job:** a spec cobre apenas a varredura na inicialização.
      A primeira frase do pedido ("remover imediatamente ao finalizar") pede também remoção
      automática quando o Builder encerra com sucesso — isso seria feat-96-02 ou está coberto
      por `worker-job.md` (instrução ao agente)? Decisão do Líder.
- [ ] **Quais sessões recebem o prompt:** hoje o hook está apenas em `forge up`. O usuário quer
      o mesmo aviso em `forge worker`, `forge inbox`, ou só no `forge up`?

## Notas técnicas (não bloqueiam Gate 1, registradas para o Builder)

- Importação cruzada `internal/survey` ↔ `internal/pipeline`: verificar antes de implementar
  `ChooseWorktreesToClean`; se houver ciclo, mover a função para `internal/pipeline/gc_prompt.go`
  usando `huh` diretamente (já é dependência do módulo; `huh.NewMultiSelect` confirmado em v0.3).
- `queue.LoadForInitiative` é helper novo em `internal/queue`; lê o queue.json da iniciativa
  pelo ID sem depender da branch git ativa.
