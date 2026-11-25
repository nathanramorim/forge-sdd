# ADR-0029: Builder por task, worktree efêmero e esteira adaptativa

- **Data:** 2026-11-25
- **Status:** `proposto` (aprovação do dono no Gate 1 da iniciativa 122)
- **Regras afetadas (IDs da constitution):** `FLX-05`, `GOV-01`, `ARQ-04`, `SEG-02`, `SEG-03`, `FLX-06`
- **Autor / Responsável:** Builder (feat-122-01), dono do projeto

## Contexto

O barramento cria um job `feature_build` por feature e um único `claude -p` implementa todas as tasks, relendo o histórico das anteriores (88: 81 turnos, 8,3 mi de tokens relidos). `PlanWaves`, o leque e `PodeComecar` existem, mas nenhum comando os aciona. O worktree só sai na junção da onda e o paralelo é um número fixo, sem olhar o disco da máquina. Um único Reviewer atende a fila.

## Opções Consideradas

1. **Manter um job por feature** e só melhorar o prompt: o contexto continua crescendo.
2. **Um job por task, worktree efêmero, pares Builder e Reviewer e teto por disco:** contexto novo por task, vaga de disco devolvida no commit, revisão acompanha a produção.
3. **Subagentes dentro do mesmo job:** o contexto do orquestrador continua crescendo e o Go não controla disco.

## Decisão

Opção 2, atrás da chave `builder.granularidade` (`feature`, padrão, ou `task`).

- **Job por task:** novo tipo `task_build` com `task_id` e `feature_job` (id do `feature_build` pai). `task_id` precisa estar em `task_ids`, o plano da feature. Um Builder novo por task; o pacote é o de `forge context <task>` mais a spec da task (RN-1, RN-2).
- **Worktree efêmero:** nasce para a task, é apagado no commit; a branch da task continua com o commit. Só fica se o commit falhar ou houver trabalho não commitado (órfão vira menu, contrato do leque).
- **Pares:** cada commit de task gera um job de review com o diff `base...task`, sem worktree, tier econômico; até um Reviewer ativo por Builder ativo.
- **Junção contínua:** task aprovada entra na branch da feature por merge sob lock, em série, na ordem do plano; conflito bloqueia a task e preserva a branch. Um PR em draft por feature.
- **Teto por disco:** `pares = min(vagas de disco, largura da onda, teto da janela)`, com vagas = (`builder.limite_disco`% da capacidade − uso) ÷ tamanho do worktree; `max_parallel_tasks` vira teto máximo. `builder.limite_disco` fica entre 50 e 95 (padrão 80). Medição só com a biblioteca padrão (`ARQ-04`), recalculada a cada worktree apagado.
- **Contrato compacto entre agentes:** entrada em seções de tags XML com chave e valor e listas em tabela de cabeçalho único, sem moldura nem emoji, com teto de bytes; saída lida pelo Go em JSON validado. Para o humano, texto em tópicos gerado pelo Go a partir dos dados estruturados, nunca a saída crua do agente (`SEG-08`). O formato final da entrada é decidido por `forge eval compare`.
- **Reversão:** voltar `builder.granularidade` a `feature`; nenhum dado migra.

## Consequências

- **Positivas:** contexto limpo por task, menos tokens relidos, disco devolvido na hora, revisão acompanhando a produção, teto adaptado a cada máquina.
- **Negativas / Trade-offs:** mais jobs e mais processos; dobra o número de revisões; junção contínua exige cuidado com conflito. Várias máquinas na mesma fila continuam fora (exigem barramento compartilhado).
- **Ações de acompanhamento:** telemetria por task e papel (feat-122-04), worktree efêmero (feat-122-03), medição nos dois modos antes de mudar o padrão (feat-122-08, RN-12).
