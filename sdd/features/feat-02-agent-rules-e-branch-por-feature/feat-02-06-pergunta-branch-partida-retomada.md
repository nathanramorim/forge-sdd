# Feature 02-06 — Pergunta Obrigatória de Branch de Partida e Retomada

Complementa feat-02-05: além de agrupar subtarefas numa única branch, o agente passa a perguntar explicitamente de onde partir e se deve retomar trabalho em progresso, em vez de assumir a branch atual do checkout. Depende de feat-02-05.

## Critérios de Aceitação Executáveis

1. Antes de criar/usar a branch da feature (arquivo único ou pasta agrupada), o agente pergunta obrigatoriamente qual branch usar como ponto de partida, com default sensato (`main`) que só exige resposta explícita quando há ambiguidade real.
2. O agente checa se já existe uma branch de feature anterior para aquela mesma pasta/arquivo (`git branch --list feat/XX-*`) e, se existir, pergunta se deve continuar a partir dela em vez de recriar do zero.
3. Replicado nos três agentes via corpo canônico em `.agent/commands/` (feat-02-03), sem editar 3 arquivos divergentes.
4. Não regride o fluxo hoje testado quando não há branch anterior nem ambiguidade (segue com o default sem fricção extra).

## Status: done

`.agent/commands/nova-feature.md.tmpl`, `novo-fix.md.tmpl` e `proxima-feature.md.tmpl` ganharam o passo obrigatório: perguntar a branch de partida (default `main`, só pula a pergunta se já indicada explicitamente na mesma solicitação) e checar `git branch --list <prefixo>/*` antes de criar a branch — se já existir uma branch da mesma feature/fix, perguntar se deve retomá-la em vez de recriar. Propagado aos três agentes via corpo canônico único (feat-02-03), sem editar adaptadores individualmente.
