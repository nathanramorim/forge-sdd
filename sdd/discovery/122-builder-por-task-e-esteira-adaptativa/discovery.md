# Discovery 122 — Builder por task e esteira adaptativa

**Origem:** Nathan, 24/11/2026, no chat: "a gente já consegue fazer a quebra por tasks, só não consegue colocar um agente de builder para implementar cada task. Quando a gente vai fazendo a fila e coloca o mesmo agente, o contexto fica muito grande. [...] escalar os agentes, até mesmo o revisor, gastar menos token quando quebrar para implementação [...] e não travar tanto quando a coisa é simples, mas quando a fila já é maior, não tem como."

Ajuste do dono no mesmo dia: "acabei de implementar no worktree, fiz o meu commit, já apago o worktree [...] todos podem trabalhar na mesma branch da feature [...] squash depois [...] não tenho espaço de armazenamento para rodar vários worktrees e eu quero ter escala [...] até 80% da capacidade de armazenamento é o limite de paralelizar [...] para cada builder tem que ter um revisor."

Retoma o item que a 90 deixou para a próxima onda: "Builder por tarefa com contexto novo" (`90/discovery.md`, Fora).

## Problema
1. **Um contexto para a feature inteira.** O barramento cria um job `feature_build` por feature e o worker roda um único `claude -p` para todas as tasks dela (`internal/pipeline/worker.go`, `processJob`). Cada task nova relê o histórico das anteriores. Linha de base da 90: o Builder da 88 usou 81 turnos e 8,3 mi de tokens relidos, cerca de 102 mil por turno.
2. **Paralelismo só como sugestão.** `PlanWaves` calcula as ondas por "Depende de", mas o resultado vira só uma dica no prompt (`BuilderHints`).
3. **Peças prontas sem quem as acione.** O leque (`leque.go`: worktree e branch por task, merge na branch base, remoção do worktree), a continuação por onda (`onda.go`) e o laço por task (`laco.go`) existem e têm teste, mas nenhum comando os chama.
4. **Worktree vive demais e ocupa disco.** No leque, o worktree só sai na junção, no fim da onda inteira. Com muitas tasks, o disco enche antes de a onda acabar.
5. **Paralelo sem noção da máquina.** `max_parallel_tasks` é um número fixo (1 a 8). Não olha disco livre, e cada máquina tem uma capacidade.
6. **Revisão desacoplada do Builder.** Um único Reviewer atende a fila. Com vários Builders, a revisão vira gargalo.

## Para quem
Quem deixa a esteira rodando depois do Gate 1, em uma ou mais máquinas, e quer entregar mais rápido, com menos tokens, sem travar o computador.

## Jornada alvo
1. O split aprovado gera as tasks com "Depende de" e tier (115).
2. `forge dispatch <feature>` enfileira **uma unidade de trabalho por task** na branch da feature, com o pacote mínimo dela (`forge context <task>`).
3. O despachante calcula quantos **pares Builder e Reviewer** cabem nesta máquina: o disco ocupado não pode passar de 80% contando os worktrees novos, e a janela de uso não pode estar no limite.
4. Cada Builder novo trabalha no **seu worktree efêmero**, roda a Camada A, faz o commit na branch da task e o worktree é **apagado na hora**. A vaga de disco volta para a fila.
5. O Reviewer do par recebe o commit pelo git (sem worktree) e emite o parecer da task. Pediu ajuste: um worktree novo nasce da branch da task, o Builder corrige, apaga de novo.
6. Aprovada, a task entra na **branch da feature** por merge sob lock, na ordem do plano, e a branch da task é apagada.
7. Quando várias vagas liberam de uma vez, o despachante sobe vários pares de uma vez, até o teto.
8. Fechadas as tasks, sai **um PR em draft da feature**, com squash no merge final quando o dono preferir.
9. Cada task grava telemetria real do Builder e do Reviewer em `sdd/.metrics`.

## Regras de negócio
- **RN-1** Uma task, um Builder novo, um contexto. Nenhum Builder recebe o histórico de outra task.
- **RN-2** O pacote de entrada é o de `forge context <task>` mais a spec da task. Nada de `status.md` inteiro, logs ou diffs de outras tasks.
- **RN-3** Uma task só começa com as dependências já na branch da feature (`PodeComecar`).
- **RN-4** Worktree é efêmero: nasce para a task, morre no commit. Só fica se o commit falhar ou houver trabalho não commitado (órfão vira menu, contrato do leque).
- **RN-5** Sem teto fixo: o limite é o que a máquina aguenta sem travar. `pares = min(vagas de disco, vagas de memória, largura da onda, teto da janela)`. Vagas de disco = (80% da capacidade do volume dos worktrees − uso atual) ÷ tamanho de um worktree (`builder.limite_disco`). Vagas de memória = (memória disponível − reserva de 2 GB) ÷ memória medida de um par Builder e Reviewer (`builder.reserva_memoria_mb`). `max_parallel_tasks` vira teto opcional (0 = só recurso). Decisão do dono em 24/11/2026.
- **RN-6** O tamanho de um worktree é medido no primeiro checkout do projeto e guardado no estado local da máquina; reavaliado a cada 20 tasks.
- **RN-7** O cálculo roda de novo a cada worktree apagado. Se várias vagas abrem juntas, várias tasks começam juntas.
- **RN-8** Um Reviewer por Builder ativo. O Reviewer lê só o diff da task (`git diff base...task`) e a spec da task, com tier econômico; não precisa de worktree.
- **RN-9** Junção na branch da feature em série, sob lock, na ordem do plano. Conflito nunca é resolvido pelo modelo sem decisão humana (contrato do leque).
- **RN-10** Janela de uso acima do limite (padrão 85%) reduz o teto para 1 par. O adiamento por limite de uso que já existe continua valendo.
- **RN-11** Cada máquina calcula o próprio teto. Várias máquinas na mesma fila exigem barramento compartilhado, que fica fora desta iniciativa.
- **RN-12** O modo atual (um job por feature) continua disponível por configuração (`builder.granularidade`) até a medição comprovar ganho.
- **RN-13** O modelo de cada task vem do tier da 115; esta iniciativa não cria outra classificação.
- **RN-14** Dois registros de comunicação. **Entre agentes**, compacto e sem moldura, emoji ou repetição, com teto por campo: (a) **entrada do agente** (prompt do job, handoff): seções delimitadas por tags XML, como a Anthropic recomenda, com chave e valor dentro; listas uniformes (tasks, arquivos, achados anteriores) em tabela com cabeçalho único, no estilo TOON, gerada por um codificador próprio sem dependência; (b) **saída do agente lida pelo Go** (retorno do Builder, parecer): JSON validado, porque é o formato que os modelos geram com mais acerto e o Go lê só com a biblioteca padrão. **Para o humano** (status, notificação, corpo do PR, menu): texto assertivo em tópicos, gerado pelo Go a partir dos dados estruturados, nunca a saída crua do agente. A escolha final de formato da entrada é decidida por `forge eval compare` com dados reais (tokens e acerto), não por benchmark de terceiros.
- **RN-15** O agente devolve só o contrato estruturado. Prosa livre do agente fica no log, não segue para o próximo agente.

## Escopo
**Dentro:** job por task; worktree efêmero apagado no commit; despachante com teto por disco e por janela; pares Builder e Reviewer; junção contínua na branch da feature; contrato compacto entre agentes e saída legível para o humano; telemetria por task; chave para voltar ao modo por feature; medição antes e depois.

**Fora:** barramento compartilhado entre máquinas; classificação de tier (115); modo `chat` em subagentes (104, que poderá reusar o despachante); orquestrador de sessões (107); painel visual (123); mudança no fluxo de aprovação humana.

## Riscos
| Risco | Mitigação |
|---|---|
| Conflito de merge entre tasks paralelas | Ondas por "Depende de", junção em série na ordem do plano, bloqueio com menu |
| Disco medido errado trava a máquina | Limite conservador (80%), tamanho remedido a cada 20 tasks, piso de 1 par |
| Build e testes também ocupam disco e memória | O tamanho medido inclui o primeiro build; memória com reserva de 2 GB e consumo por par medido |
| Dobrar revisores dobra tokens de revisão | Reviewer só com o diff da task e tier econômico; medição na onda 3 |
| Retrabalho recria worktree e gasta disco | Recria só quando o Reviewer pede ajuste; mesma regra de vagas |
| Formato compacto perde informação que o próximo agente precisa | Contrato com campos obrigatórios validados na borda; prosa completa preservada no log |
| Dependência da 115-03 (tier no job) | Sem tier, usa o modelo do papel (fallback da ADR-0024) |

## Esforço estimado
Entre 4 e 5 dias de esteira, sem histórico de velocidade para calibrar. O maior risco é a junção contínua com pares em paralelo (onda 2 do plano).
