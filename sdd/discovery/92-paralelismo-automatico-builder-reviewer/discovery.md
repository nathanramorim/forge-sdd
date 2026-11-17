# Discovery 92 — Builder e Reviewer em paralelo, decidido pelo Go

Nasceu de um `/nova-feature`, mas mexe em barramento, worker e revisão ao mesmo tempo; por isso veio com discovery, critérios e plano (como a 89).

## Problema
Conferido em 17/11 no código e nos logs da esteira.

1. **O paralelismo existe só no papel.** O `forge.json` nasce com `builder.max_parallel_tasks: 3` e `use_git_worktrees: true`. O `forge bus plan` calcula as ondas e o prompt manda o Builder "isolar cada task com `forge pipeline worktree add`" e "paralelizar com subagentes". Nos 7 jobs medidos do Builder: nenhum subagente e nenhum worktree. Nos três jobs de iniciativa ele chamou o `forge bus plan` e, mesmo assim, fez tudo em fila, numa conversa só.
2. **Ele nem poderia.** A definição do agente Builder não libera a ferramenta de subagente. E dentro de um worktree os comandos do `forge` não acham o projeto: `.forge/` fica fora do git e o caminho é sempre relativo à pasta atual. Testado num repositório de teste: o mesmo `forge worker --once` roda na pasta principal e, na cópia, responde "modo pipeline não ativo".
3. **Tempo perdido.** A iniciativa 90 levou 28 minutos em fila. Pelas horas dos commits e pelas dependências, com 3 vagas ela caberia em algo entre 16 e 20 minutos, conforme quantas tasks mexem no mesmo arquivo. É estimativa: cada task em conversa nova tem um custo de partida que ela não conta.
4. **Quem decide é o modelo.** Dividir, isolar, integrar e limpar estão em texto de skill. `RemoveWorktree` usa `--force` e apaga a branch auxiliar: um descuido perde trabalho não integrado.
5. **Dependência não segura a fila.** `bus.Claim` pega o próximo job por prioridade e ordem, sem olhar `**Depende de:**`. Com um job por task (91-07), duas vagas pegariam tasks dependentes ao mesmo tempo.
6. **Ninguém olha conflito de arquivo.** Duas tasks independentes que mexem no mesmo arquivo cairiam na mesma onda.
7. **O Reviewer revisa em bloco.** É um job por PR inteiro: na 90, 3.719 linhas em 36 segundos, por amostragem. A ADR-0004 já registrava que a revisão deveria ser por task e em paralelo.
8. **Tudo acontece na sua pasta.** Builder e Reviewer trabalham no checkout em que você conversa com o Líder. Foi assim que um arquivo seu, fora do git, entrou no PR #88 (`git add -A .agents`). E nos logs o Reviewer, que é somente leitura, rodou `git stash` e `git checkout` nessa pasta para conseguir uma cópia limpa do código; numa revisão ele mesmo criou um worktree para rodar os testes.

## Para quem
Dono do projeto que deixa a esteira rodando e quer a entrega mais cedo, com revisão de verdade em cada task e sem agente mexendo na pasta em que ele trabalha.

## Jornada
1. No `/split-features`, o Líder declara em cada task o que ela toca (`**Toca:**`), ao lado de `**Depende de:**`.
2. O `forge dispatch` mostra o plano: o que roda junto, o que espera e por quê.
3. O worker do Builder abre até N agentes. Cada um trabalha numa cópia própria (worktree) da branch da iniciativa: implementa, testa e commita ali.
4. Ao terminar, o agente chama o mesmo `forge pipeline pr`. O worker integra um de cada vez na branch da iniciativa, roda a Camada A, faz o push e atualiza o PR único.
5. Cada task integrada vira uma revisão própria. O Reviewer roda várias ao mesmo tempo, cada uma numa cópia somente leitura do commit integrado.
6. Quando a última task é aprovada, chega um aviso só: "PR pronto".
7. O painel mostra quem está em cada vaga, e o relatório compara o tempo de relógio com a soma dos jobs.

## Regras de negócio
- **RN-1 Quem paraleliza é o Go**, não o modelo: vagas, ordem, isolamento, integração e limpeza são código.
- **RN-2 Unidade:** um job por task (91-07). Sem isso, tudo roda como hoje.
- **RN-3 Pode rodar junto** a task que não tem dependência pendente, não toca arquivo de task em execução e encontra vaga livre. A dependência conta como pronta quando foi integrada. Task sem `**Toca:**` roda sozinha.
- **RN-4 Cada job roda numa cópia própria**, criada pelo Go. Nenhum agente trabalha no checkout da pessoa.
- **RN-5 Só o worker integra, um por vez.** Conflito nunca é resolvido à força: a integração é desfeita e a task volta ao Builder sobre a base nova, com a lista dos arquivos em conflito.
- **RN-6 Nada se perde:** cópia com commit ainda não integrado nunca é apagada; entrega com arquivo fora do commit é recusada.
- **RN-7 Uma branch e um PR por iniciativa** (Regra 15). As branches auxiliares `forge/…` são locais, temporárias e nunca vão ao remoto.
- **RN-8 A revisão é por task**, sobre o diff daquela task, e o veredito vale só para ela. "PR pronto" é avisado uma vez, quando todas estão aprovadas e a Camada A passa na branch integrada.
- **RN-9 Freio automático:** limite de uso da conta ou dois conflitos seguidos na mesma iniciativa reduzem para uma vaga, até normalizar.
- **RN-10 Nada muda para quem não ligar:** sem "um job por task", vale o fluxo de hoje; com `max_parallel_tasks: 1`, os jobs rodam em fila; o modo `live` e o `/toca-ai` seguem com uma vaga.
- **RN-11 Tetos e `STOP`** (88) valem para cada job, e o Reviewer continua somente leitura (`SEG-06`).

## Escopo
**Dentro:** campo `**Toca:**` e plano de execução; reivindicação que respeita dependência e arquivo; cópia por job criada e limpa pelo Go; comandos do `forge` funcionando de dentro da cópia; integração serializada pelo worker, com retrabalho em caso de conflito; vagas do Builder e do Reviewer; revisão por task; aviso único de PR pronto; freio; painel, Settings e relatório; instruções do Builder e do Reviewer; ADR-0017.

**Fora:** a esteira integrar numa cópia própria (a branch da iniciativa continua no seu checkout, que precisa estar nela); paralelo no modo `live` e no `/toca-ai` (etapas que geram arquivo fora do git se perderiam numa cópia); várias iniciativas ao mesmo tempo; o Go commitar a fila e os pareceres; subagentes dentro de um job; revisão final do PR inteiro por agente; comando de preparo da cópia (por exemplo, instalar dependências).

## Riscos
| Risco | Mitigação |
|---|---|
| Conflitos de integração gastarem mais do que o paralelo economiza | `**Toca:**` evita o previsível; freio (RN-9); o relatório conta os conflitos; começa desligado |
| Paralelo consumir o limite de uso mais rápido | Freio por limite de uso; vagas configuráveis; o total de tokens não aumenta |
| Seu checkout em outra branch ou com alteração local no caminho | A integração é adiada e o inbox diz o motivo; o trabalho fica guardado na cópia |
| Merge limpo que quebra o build | Camada A depois de cada integração; a revisão da task roda os testes no commit integrado; CI |
| `**Toca:**` declarado errado | É dica de agenda, não fronteira: o git pega o conflito real, e o relatório lista quem tocou fora |
| Corrida entre vagas | Um único processo integra, com exclusão mútua; testes com `-race` |
| Projeto com dependências instaladas na pasta (`node_modules`) pagar a instalação em cada cópia | Cópia de job retomado é reaproveitada; comando de preparo fica para depois de medir |
| Revisão por task perder problema entre tasks | Camada A final, CI e a aprovação humana do PR inteiro |
