# Discovery 122 — Builder por task e esteira adaptativa

**Origem:** Nathan, 24/11/2026, no chat: "a gente já consegue fazer a quebra por tasks, só não consegue colocar um agente de builder para implementar cada task. Quando a gente vai fazendo a fila e coloca o mesmo agente, o contexto fica muito grande. [...] escalar os agentes, até mesmo o revisor, gastar menos token quando quebrar para implementação [...] e não travar tanto quando a coisa é simples, mas quando a fila já é maior, não tem como."

Retoma o item que a 90 deixou para a próxima onda: "Builder por tarefa com contexto novo" (`90/discovery.md`, Fora).

## Problema
1. **Um contexto para a feature inteira.** O barramento cria um job `feature_build` por feature e o worker roda um único `claude -p` para todas as tasks dela (`internal/pipeline/worker.go`, `processJob`). Cada task nova relê o histórico das anteriores: o custo por turno cresce com o tamanho da feature. Linha de base da 90: o Builder da 88 usou 81 turnos e 8,3 mi de tokens relidos, cerca de 102 mil por turno.
2. **Paralelismo só como sugestão.** `PlanWaves` já calcula as ondas por "Depende de" e `max_parallel_tasks` (padrão 3), mas o resultado vira só uma dica no prompt (`BuilderHints`). Paralelizar fica a critério do modelo.
3. **Peças prontas sem quem as acione.** O leque (`leque.go`: worktree por task, junção em ordem, conflito vira bloqueio), a continuação por onda (`onda.go`: `PodeComecar`, `FimDaOnda`) e o laço por task (`laco.go`: estados gravados pelo Go) existem e têm teste, mas nenhum comando os chama.
4. **Revisão cara e uniforme.** O parecer LLM roda no mesmo ritmo do Builder, mesmo quando a Camada A (determinística) já reprovaria.
5. **Ritmo fixo.** Fila de uma task e fila de quinze seguem o mesmo caminho. Não há regra para rodar em sequência quando é simples, nem para subir o paralelo quando a fila cresce, nem para segurar quando a janela de uso está no fim.

## Para quem
Quem deixa a esteira rodando depois do Gate 1 e quer entregar mais rápido, com menos tokens e sem perder o controle do que cada agente está fazendo.

## Jornada alvo
1. O split aprovado gera as tasks com "Depende de" e tier (115).
2. `forge dispatch <feature>` enfileira **uma unidade de trabalho por task**, com o pacote mínimo dela (`forge context <task>`).
3. O despachante calcula as ondas e decide o paralelo: 1 worker quando a onda tem 1 ou 2 tasks simples; até `max_parallel_tasks` quando a onda é larga; menos quando a janela de uso passou do limite configurado.
4. Cada task roda num **Builder novo**, no seu worktree, e termina com a Camada A da task. Falha da Camada A volta para a mesma task, sem gastar parecer LLM.
5. Ao fim da onda, o leque junta as tasks na branch da feature, na ordem do plano. Conflito vira bloqueio com menu, nunca resolução automática.
6. Com a feature junta, o PR em draft sai e o Reviewer emite **um** parecer LLM para a feature.
7. Cada task grava sua telemetria (`sdd/.metrics`, `medicao: real` quando houver evento `result`), e o `forge status` mostra tokens por task.

## Regras de negócio
- **RN-1** Uma task, um processo de agente, um contexto. Nenhum Builder recebe o histórico de outra task.
- **RN-2** O pacote de entrada é o de `forge context <task>` mais a spec da task. Nada de `status.md` inteiro, logs ou diffs de outras tasks.
- **RN-3** A ordem é a do plano: uma task só começa com as dependências integradas (`PodeComecar`).
- **RN-4** Paralelo efetivo = mínimo entre largura da onda, `max_parallel_tasks` e o teto pela janela de uso. Fila simples (até 2 tasks, tier `fast`) roda em sequência num worker.
- **RN-5** Janela de uso acima do limite configurado (padrão 85%) reduz o paralelo para 1. O adiamento por limite de uso que já existe (`OutcomeDeferred`) continua valendo.
- **RN-6** Camada A por task, sem modelo. Parecer LLM só por feature, depois da junção. O teto de `max_rework_attempts` vale por task.
- **RN-7** Conflito na junção nunca é resolvido pelo modelo sem decisão humana (mesmo contrato do leque, RN-9 da 104).
- **RN-8** O modo atual (um job por feature) continua disponível por configuração enquanto a medição não comprovar ganho.
- **RN-9** O modelo de cada task vem do tier da 115; esta iniciativa não cria outra classificação.

## Escopo
**Dentro:** job por task no barramento; despachante de ondas com paralelo adaptativo; leque ligado ao worker no modo `sessions`; Camada A por task e parecer por feature; telemetria por task; chave de configuração para voltar ao modo por feature; medição antes e depois.

**Fora:** classificação de tier (é da 115); modo `chat` em subagentes (é da 104, que poderá reusar o despachante depois); orquestrador de sessões da 107; painel visual (é da 123); qualquer mudança no fluxo de aprovação humana.

## Riscos
| Risco | Mitigação |
|---|---|
| Conflito de merge entre tasks paralelas | Ondas por "Depende de", junção em ordem, bloqueio com menu (leque) |
| Task sem contexto suficiente | Pacote de `forge context` validado na Camada A; falha volta à mesma task |
| Paralelo esgota a janela de uso mais rápido | RN-5 e adiamento por limite já existente |
| Mais processos, mais custo fixo de arranque por task | Medição antes e depois (RN-8); fila simples segue em sequência |
| Dependência da 115-03 (tier no job) | Sem tier, usa o modelo do papel (fallback da ADR-0024) |

## Esforço estimado
Entre 4 e 5 dias de esteira, sem histórico de velocidade para calibrar. O maior risco é a junção paralela (onda 2 do plano).
