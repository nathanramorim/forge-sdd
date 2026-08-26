# Clarify — demo

Lógica única, referenciada por `/nova-feature`, `/novo-fix` e `/discovery` (nos três agentes) para decidir quando esclarecer uma dúvida sobre a descrição recebida antes de agir — evita a mesma heurística duplicada em 3+ prompts, e evita tanto o silêncio total (agir sobre suposição não confirmada) quanto a pergunta forçada em pedidos já claros.

## Quando perguntar (sinais objetivos)

Antes de iniciar o passo seguinte do comando (criação de branch em `/nova-feature`/`/novo-fix`, ou produção dos três artefatos em `/discovery`), avalie a descrição recebida do usuário contra estes sinais:

1. **Critério de aceitação ausente ou implícito** — não dá para saber, a partir da descrição, como validar que a tarefa terminou.
2. **Escopo com mais de uma interpretação plausível** — a descrição admite pelo menos duas leituras razoavelmente diferentes de "o que construir".
3. **Dependência externa não mencionada** — a tarefa parece depender de outro sistema, decisão ou artefato (ex: outra feature, uma API externa, uma escolha de arquitetura) que não foi citado.

## Como agir

- **Nenhum sinal detectado:** siga direto, sem pergunta forçada. Descrições já claras e completas não geram fricção.
- **Algum sinal detectado:** faça uma rodada objetiva de perguntas ao usuário (curta, focada nos sinais encontrados — não um questionário genérico) antes de prosseguir. Registre a resposta recebida no artefato criado (spec de feature/fix, ou discovery), para não perder o contexto que motivou a pergunta.
- Este passo é **condicional, não obrigatório a cada invocação** — não é um checklist a marcar sempre, é um filtro que só ativa quando há lacuna real.
