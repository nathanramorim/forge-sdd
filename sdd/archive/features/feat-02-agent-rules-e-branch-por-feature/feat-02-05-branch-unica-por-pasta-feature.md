# Feature 02-05 — Branch Única por Pasta de Feature Quebrada

Formaliza como regra de lifecycle o que hoje depende de inferência do modelo: quando uma feature é quebrada em subpastas (`sdd/features/feat-XX-nome/*.md`), a pasta inteira é a unidade de execução, com uma única branch agrupando todas as subtarefas. Depende de feat-02-03 (corpo canônico em `.agent/commands/`, propagando a mudança automaticamente aos três agentes em vez de exigir edição tripla).

## Critérios de Aceitação Executáveis

1. `/nova-feature`, `/proxima-feature` e `/novo-fix` (corpo canônico em `.agent/commands/`) passam a detectar quando a feature/fix corrente é uma subpasta (`sdd/features/feat-XX-nome/*.md`, não um arquivo único) e tratam a pasta inteira como unidade, usando uma única branch (`feat/XX-nome`) que agrupa todas as subtarefas.
2. Comportamento de branch para features de arquivo único (sem subpasta, formato atual) permanece inalterado — sem regressão nos fluxos já cobertos por golden fixtures/documentação.
3. Nova regra registrada em `sdd/memory/constitution.md` (revisar contagem atual de regras antes de adicionar — cabeçalho declara limite de 12, hoje já em 14).
4. Validado manualmente executando o fluxo contra `sdd/features/feat-02-agent-rules-e-branch-por-feature/` (esta própria pasta) como caso real de feature quebrada.

## Status: done

`.agent/commands/nova-feature.md.tmpl` e `proxima-feature.md.tmpl` (corpo canônico, propagado automaticamente aos três agentes via feat-02-03) ganharam a instrução explícita: se a feature/fix corrente for uma subpasta (`sdd/features/<prefixo>-ID-<nome>/*.md`), a pasta inteira é a unidade de execução, com uma única branch agrupando todas as subtarefas — nunca uma branch por subtarefa. Comportamento de arquivo único (sem subpasta) não mudou. Nova Regra 15 registrada em `sdd/memory/constitution.md` (header ajustado de "máx. 12" para "máx. 15", refletindo a contagem real após a adição — já estava em 14 antes desta feature) e replicada no template genérico `constitution.md.tmpl` (rule 7) para novos projetos.
