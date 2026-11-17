# Discovery 91 — Modelo certo para cada trabalho e respostas por menu

## Problema
Dois incômodos do dia a dia, conferidos em 17/11 no código e nos logs da esteira.

**A. Perguntas que a pessoa responde digitando.**
1. **Pergunta sem menu.** A definição do Líder (`.claude/agents/lider.md`) lista seis ferramentas e nenhuma é a de pergunta. As skills dizem "pergunte" em 10 pontos, e o passo PLAN das instruções raiz também, sem dizer como. Nesta sessão isso virou pelo menos três rodadas de opções numeradas para digitar a resposta.
2. **O Líder não consegue delegar.** O PLAN manda perguntar "vai para um subagente?", mas a ferramenta de subagente não está na lista do Líder.

**B. Modelo escolhido sem critério.**
3. **O modelo da esteira é o que estiver salvo no chat.** O `forge.json` não tem `models`, então Builder e Reviewer usam o padrão da CLI. O `/model` do chat grava esse padrão (`~/.claude/settings.json` está com `model: fable`). Trocar o modelo na conversa muda, sem aviso, o modelo e o custo dos próximos jobs.
4. **Um modelo para tudo.** Os 14 jobs medidos (88, 89, 90 e fixes) rodaram em `claude-sonnet-5-5`, da ADR de uma página ao comando que abre PR.
5. **Um job leva todas as tasks.** O `forge dispatch` cria um job com as 8 tasks: mesmo modelo para todas e conversa que só cresce (163 mil tokens relidos por turno na 90, contra 102 mil na 88). O motor do `/toca-ai` já faz um job por etapa, mas com uma branch por etapa (`feat/<iniciativa>-<etapa>`), o que a Regra 15 não permite. Com essa branch, o archive automático do `/toca-ai` recusa: a guarda de branch da 90-06 não reconhece a branch de etapa como a da iniciativa.
6. **O menu de modelo só serve ao `claude`.** Ele oferece `haiku`, `sonnet` e `opus` (falta `fable`, que a CLI aceita). Para `agy` (18 modelos em `agy models`) e `gemini` (`-m`) só aparece "Padrão da CLI", e não há como digitar outro.
7. **Modelo preso à sessão, não ao agente.** `models.builder` é um texto só: trocar o agente da sessão mantém o nome de modelo do agente anterior.
8. **Troca manual por comando.** Nesta sessão o modelo foi trocado à mão três vezes: um mais forte para o `/discovery`, um mais leve para o `/split-features` e outro para a documentação.

## Para quem
Dono do projeto que conversa com o Líder (inclusive pelo `/rc`) e deixa a esteira rodando. Quer responder com um toque e não escolher modelo a cada passo.

## Jornada
1. Em **Settings**, escolhe para cada agente (claude, agy, gemini…) o modelo de três níveis, **leve**, **médio** e **forte**, a partir da lista do próprio agente, e liga a seleção automática.
2. No `/split-features`, o Líder dá um nível a cada task por uma regra objetiva; o despacho confere.
3. A esteira roda cada task no modelo do seu nível. Se o Reviewer devolver, o retrabalho sobe um nível.
4. No chat, toda decisão chega como menu. O `/discovery` roda no forte e o `/status` no leve, sem `/model`.
5. `forge report --por-modelo` mostra como cada nível e modelo se saiu.

## Regras de negócio
- **RN-1 Um vocabulário:** três níveis, `leve`, `medio` e `forte`, usados por tasks, comandos e agentes.
- **RN-2 Tabela por agente:** cada agente (CLI) tem a sua tabela nível → modelo e esforço; nome de modelo não vale de um agente para outro.
- **RN-3 Nada muda para quem não ligar:** a seleção `manual` é o padrão; comandos da CLI, adaptadores e definições de agente ficam idênticos.
- **RN-4 Ordem de escolha do modelo de um job:** o que estiver gravado no job; o nível (se `auto`); a escolha da sessão; o padrão da CLI.
- **RN-5 Quem decide o nível é a spec** (o Líder, por rubrica); o Go só confere e aplica. Task sem nível vale `medio`, com aviso.
- **RN-6 Retrabalho sobe um nível**, até `forte`; nunca desce.
- **RN-7 O Reviewer fica fora do roteamento:** usa a escolha da própria sessão.
- **RN-8 Pergunta só por menu:** uma decisão por pergunta, a opção recomendada primeiro; nunca pedir para digitar número ou letra.
- **RN-9 Sem menu disponível**, o agente segue a opção recomendada e registra em "A confirmar"; pergunta em frase só para ação irreversível ou externa (merge em `main`, apagar dados, gasto, segredo).
- **RN-10 Medir:** cada job grava nível, modelo e tentativa; a sugestão do relatório nunca muda a tabela sozinha.
- **RN-11 Segurança:** modelo e nível validados na borda (`SEG-01`) e passados em argv, sem shell (`SEG-03`).

## Escopo
**Dentro:** regra única de pergunta por menu e ferramentas do Líder; lista de modelos por agente no Settings, com "Outro (digitar)"; níveis e tabela por agente; nível na spec com rubrica; roteamento por job no worker; um job por task no `dispatch` (opcional), sempre na branch e no PR da iniciativa, regra que o motor do `/toca-ai` passa a seguir; escalada no retrabalho; relatório por nível e modelo; modelo por comando e por papel no chat; correção das skills do Líder que ficou da 90 (`--fix`, `--base`); ADR-0016.

**Fora:** trocar de modelo a cada job no modo `live` (vale o da sessão); roteamento do Reviewer; o Go classificar o nível sozinho pelos sinais da spec; a tabela se ajustar sozinha pelo histórico; menu próprio do Forge por popup do tmux (não aparece no `/rc`); o Líder delegar por conta própria (ele ganha a ferramenta, mas a decisão continua da pessoa).

## Riscos
| Risco | Mitigação |
|---|---|
| Nome da ferramenta de pergunta ou do campo `model:` de comando ser outro na versão instalada | Conferir na documentação oficial no início da task; sem suporte, vale a RN-9 e a entrega fica só por agente |
| Modelo leve reprovar mais e o retrabalho comer a economia | Escalada (RN-6), `manual` como padrão e relatório com a taxa de aprovação de primeira |
| Um job por task multiplicar as revisões | Opcional; o relatório soma Builder e Reviewer por iniciativa |
| Com um job por task, a task seguinte ser construída antes do parecer da anterior (o worker não espera) | O retrabalho tem prioridade alta e entra antes das próximas; o relatório mostra o retrabalho em cascata antes de se decidir travar por dependência |
| Nível mal classificado | Rubrica objetiva; o relatório lista as tasks `leve` que precisaram de retrabalho |
| `agy models` lento ou fora do ar | Limite de tempo e lista de reserva |
| Menus em excesso | Só pergunta quando há lacuna real (`clarify.md`), uma decisão por menu |
