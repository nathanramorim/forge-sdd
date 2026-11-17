# O que mudou na esteira: onda "esteira econômica" (iniciativa 90)

Este guia explica, em linguagem simples, as melhorias da iniciativa 90. Elas foram mescladas na branch `feat/88-modo-pipeline-avancado` em 17/11/2026, pelo [PR #88](https://github.com/nathanramorim/forge-sdd/pull/88). O detalhe técnico está indicado no fim.

## Em resumo

A onda teve duas ideias:

1. **Tirar o trabalho repetitivo das mãos dos agentes.** Abrir PR, arquivar e conferir dependência passaram para o programa `forge`, que faz sempre igual e não gasta tokens.
2. **Criar um medidor de verdade**, para saber quanto cada trabalho custa.

| O que | Antes | Agora |
|---|---|---|
| Medir o gasto | O número era estimado pelo tamanho dos arquivos | O `forge` lê o gasto real de cada job |
| Abrir o PR | O agente abria à mão e às vezes errava o destino | O `forge` abre, sempre no destino certo |
| Fix | Job montado à mão, fora da esteira | Um comando: `forge dispatch <iniciativa> --fix <id>` |
| Status lido pelos agentes | 11,4 KB a cada comando | 0,8 KB |
| Dependência entre tarefas | Dois leitores, um com defeito | Um leitor só, e conferência antes de construir |
| Arquivar uma iniciativa | Copiava a pasta e deixava duplicata | Move a pasta, uma vez só |
| Modelo de cada sessão | Sempre o padrão | Você escolhe modelo e esforço por sessão |

## Quatro palavras que aparecem neste guia

- **Job:** um pacote de trabalho que a esteira entrega ao Builder ou ao Reviewer.
- **Turno:** cada vez que o agente pensa e executa uma ação. Um job tem dezenas de turnos.
- **Tokens relidos:** a cada turno, o agente relê a conversa inteira até ali. É a maior parte do gasto.
- **Base do PR:** a branch que vai receber o PR.

## Por que fizemos

Medimos os logs de 12 jobs dos dias 16 e 17/11: 276 turnos e 14,7 milhões de tokens relidos. Apareceram três problemas:

- O painel mostrava um número que não batia com o real: 869 mil tokens no Builder, contra 13,6 milhões nos logs.
- Os agentes gastavam turnos com trabalho mecânico e erravam. O PR #87 abriu contra a `main`, com 3.908 linhas.
- Faltavam peças: fix não entrava na esteira, e o arquivamento deixava pastas duplicadas.

## As melhorias, uma por uma

### 1. Medidor de gasto real

- **Antes:** o `forge` estimava os tokens pelo tamanho da spec e do log.
- **Agora:** ao fim de cada job, o `forge` lê o resumo que o próprio `claude` escreve no log e grava tokens, tokens relidos, turnos, duração, modelo e custo de tabela.
- **Como usar:**

```bash
forge report --initiative 90-esteira-economica-e-autonoma --jobs
```

- **Bom saber:**
  - Só o `claude` informa o gasto real. Nas outras CLIs o número continua estimado, e o relatório mostra os dois em linhas separadas ("real" e "estimado").
  - "Custo de tabela" é o preço de lista da API. Quem usa assinatura não paga esse valor, mas ele serve para comparar um job com outro.
  - Jobs que rodaram antes desta versão continuam como "estimado".

### 2. O PR sai sempre no destino certo

- **Antes:** o Builder fazia o push e abria o PR à mão. O destino vinha de uma configuração fixa (`main`). Por isso o PR #87 abriu contra a `main`, e para despachar a própria iniciativa 90 foi preciso editar a configuração e desfazer logo depois.
- **Agora:** o Builder só faz o commit e chama `forge pipeline pr`. O `forge` roda o portão automático (`forge gate`), faz o push, abre o PR e registra. Se já existir um PR no destino errado, ele corrige o destino.
- **Como o destino é escolhido**, nesta ordem:
  1. o que for informado na hora (`forge queue enqueue --base <branch>`);
  2. o que estiver gravado na fila da iniciativa;
  3. a branch de onde a branch da iniciativa nasceu (o git guarda isso);
  4. a configuração geral (`base_branch` no `forge.json`).
- **Travas:** recusa PR que sai da `main` ou da própria base, recusa branch sem commit novo e recusa descrição fora do modelo do projeto. Nunca mescla e nunca aprova.
- **Conferido:** com a configuração geral ainda em `main`, a simulação do despacho da 90 respondeu `Base do PR: feat/88-modo-pipeline-avancado (fonte: origem da branch (git))`.

### 3. Fix entra na esteira

- **Antes:** `forge dispatch` só aceitava features. Para os fixes 84-03, 88-01 e 88-02 foi preciso montar o job à mão, sem a conferência antes do build.
- **Agora:**

```bash
forge dispatch 88-modo-pipeline-avancado --fix fix-88-02
```

- O fix passa pela mesma conferência, ganha a branch `fix/88-02-<nome>` e o PR vai contra a branch da iniciativa.

### 4. O status que os agentes leem ficou 15 vezes menor

- **Antes:** todo comando começava lendo `forge status`: 11,4 KB, com trechos de log ilegíveis no meio.
- **Agora:** quando a saída não vai para um terminal (é o caso dos agentes), sai um resumo de 0,8 KB, só com as iniciativas em andamento.
- **Para você:** no terminal continua o painel completo. `forge status --full` mostra tudo em qualquer situação, e a parte "Ao vivo" agora vem em texto legível.

### 5. Dependências lidas do mesmo jeito, e conferência antes de construir

- **Antes:** havia dois leitores da linha "Depende de" das specs. Um deles lia "nenhuma (usa …)" como se cada palavra fosse uma dependência, e a tarefa ficava presa na fila. O outro ignorava fixes.
- **Agora:** existe um leitor só. E antes de gastar tokens, o despacho confere cada tarefa. Se ela depende de algo que não existe, ou se o critério de aceitação não tem um comando para rodar, o despacho para e lista de uma vez tudo o que falta.

### 6. Arquivar agora move a pasta

- **Antes:** ao aprovar, o `forge` copiava a pasta para o arquivo e deixava a original. As iniciativas 84, 88, 89 e 90 ficaram duplicadas e foi preciso limpar à mão. Ele também repetia linhas no histórico.
- **Agora:** `forge archive` move a pasta, uma vez só. Rodar de novo não faz nada. Ele recusa, e diz o motivo, se você não estiver na branch da iniciativa ou se houver tarefa ou fix pendente.
- **Quando acontece:**
  - No `/toca-ai`: sozinho, ao aprovar a última etapa.
  - Nas outras iniciativas: você recebe o aviso "pronta para arquivar" e escolhe a hora, porque é comum abrir fix depois da aprovação.
- **Como usar:**

```bash
forge archive --concluidas --dry-run    # só lista as candidatas, não move nada
forge archive --initiative <nome>       # move a pasta
```

- **Bom saber:** o comando move a pasta, mas não faz commit. A mudança aparece no `git status` para ser commitada. Hoje a lista traz 9 candidatas: 69, 70, 71, 72, 75, 77, 84, 89 e 90.

### 7. Modelo e esforço por sessão

- **Antes:** Discovery, Builder e Reviewer usavam sempre o modelo padrão da CLI.
- **Agora:** em **Settings**, cada sessão tem duas linhas novas, "Modelo do …" e "Esforço do …". O padrão é "padrão da CLI": se você não mexer, nada muda.
- **Opções com o `claude`:** modelo `haiku`, `sonnet` ou `opus`; esforço `low`, `medium`, `high`, `xhigh` ou `max`. O menu marca como "(econômico)" as opções mais leves para Builder e Reviewer.
- **Na primeira vez:** em projeto que já tinha o pipeline configurado (como este), o menu pergunta se pode "incluir o marcador". Responda sim. Sem isso, a escolha fica gravada mas não é aplicada.
- **Outras CLIs:** `agy` aceita modelo e esforço; `gemini`, só modelo; `codex` e `copilot` ainda não, porque não estavam instalados para conferir.
- **Dica:** mude uma sessão por vez e compare no `forge report`.

## O que você precisa fazer para usar

1. **Publicar e instalar a nova beta.** A versão instalada (4.9.1-beta) ainda não tem nada disso. As melhorias estão só na branch `feat/88-modo-pipeline-avancado`.
2. **Nos outros projetos** (`reel-studio`, `marketing`): depois de instalar, rode `forge update --upgrade`. Ele atualiza as instruções dos agentes, e é assim que o Builder passa a abrir o PR pelo comando novo.
3. **Opcional:** escolher modelo e esforço em Settings.

## O que ainda não sabemos e o que faltou

### A economia ainda não foi medida

A própria iniciativa 90 foi construída pela versão antiga da esteira. Por isso o relatório novo só tem registros "estimado" para ela. A primeira comparação de verdade será a próxima iniciativa construída com a versão nova.

O que os logs mostram dos três últimos jobs do Builder:

| Iniciativa | Tarefas | Linhas escritas | Turnos | Tokens relidos | Custo de tabela |
|---|---|---|---|---|---|
| 88 | 8 | 1.348 | 81 | 8,3 milhões | US$ 2,06 |
| 89 | 2 | 1.037 | 43 | 3,2 milhões | US$ 1,56 |
| 90 | 8 | 2.809 | 132 | 21,5 milhões | US$ 5,75 |

A 90 escreveu o dobro de código da 88 e custou quase o triplo. A causa provável: um job só carregou as 8 tarefas, e a cada turno o agente releu tudo o que já estava na conversa (163 mil tokens por turno, contra 102 mil na 88). Esta onda não ataca esse ponto. Ele é o primeiro item da próxima.

### O que foi e o que não foi conferido

- **A revisão automática foi por amostragem.** O Reviewer leu por inteiro só o arquivo que abre o PR.
- **Conferido depois do merge:** `go build`, `go vet` e todos os testes passam na branch. Os comandos de leitura foram testados neste repositório: `status`, `report`, a simulação do `dispatch` (feature e fix) e a lista do `archive`.
- **Ainda não testado à mão:** abrir um PR de verdade com `forge pipeline pr`, mover uma pasta com o `archive`, o menu Settings e um job inteiro com o medidor novo.

### O que faltou

- **As instruções do Líder não foram atualizadas.** `/novo-fix`, `/implementar` e `/split-features` ainda não mandam usar `--fix` nem `--base`. Os comandos funcionam, mas uma sessão nova não vai usá-los sozinha. Esse ajuste de texto não estava previsto na spec.
- **O commit continua na mão do Builder.** Só o push e o PR passaram para o `forge`.

### Ficou para a próxima onda

- Um job por tarefa, com conversa nova a cada tarefa.
- Um mapa do código, para o agente ler menos arquivos.
- Lições tiradas automaticamente dos pareceres do Reviewer.
- Um comando para publicar a beta.

## Onde está o detalhe

- **Decisão técnica:** `sdd/spec/adr/adr-0014-esteira-economica-e-autonoma.md`
- **Referência dos comandos:** `docs/pipeline.md`, seção "Esteira econômica"
- **Specs e critérios:** `sdd/discovery/90-esteira-economica-e-autonoma/`
