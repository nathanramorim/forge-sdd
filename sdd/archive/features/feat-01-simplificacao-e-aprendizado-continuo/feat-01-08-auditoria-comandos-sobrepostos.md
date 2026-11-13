# Feature 01-08 — Auditoria de Comandos Sobrepostos (Recomendação)

Produz uma lista de candidatos a fusão/aposentadoria entre os 12 comandos existentes, para decisão explícita do usuário em uma rodada futura — não remove nenhum comando nesta feature.

## Critérios de Aceitação Executáveis

1. Documento de auditoria mapeando, para cada um dos 12 comandos de chat, seu propósito, sobreposição com outros comandos (se houver) e uma recomendação objetiva (manter / fundir com X / candidato a aposentadoria).
2. Nenhuma remoção ou fusão é executada nesta feature — a saída é a recomendação documentada, análoga ao spike de feat-5ae2-06.
3. Métrica de referência: número de comandos/conceitos que um novo usuário precisa aprender no onboarding (hoje 17, entre CLI e chat).

## Status: done (auditoria concluída, sem remoção nesta rodada)

### Tabela de Auditoria (12 comandos de chat)

| Comando | Propósito | Sobreposição | Recomendação |
|---|---|---|---|
| `/constitution` | Alinha `constitution.md`/`stack.md` com o codebase real; agora também MCPs/VCS (feat-01-05) | Nenhuma | Manter |
| `/discovery` | Discovery de produto + técnico; gera 3 arquivos (`discovery`/`criteria`/`plan`) | Também produz C4 (como `/c4-architecture`) | Manter |
| `/split-features` | Quebra um `plan-XX.md` de discovery em várias features numa subpasta | Cria `feat-XX.md`, como `/nova-feature` | Manter — já tem regra explícita ("use `/nova-feature` só para adições manuais e isoladas") que evita ambiguidade |
| `/nova-feature` | Cria uma feature (ou fix, via prefixo `fix:`) isolada, fora do fluxo de discovery | **Sobreposição real com `/novo-fix`** — já aceita `"/nova-feature fix: <descrição>"` fazendo exatamente o mesmo que `/novo-fix` | **Candidato a fusão**: `/novo-fix` poderia ser removido e documentado como um atalho/alias de `/nova-feature fix:` |
| `/novo-fix` | Cria um fix isolado | Ver acima — funcionalmente redundante com `/nova-feature fix:` | **Candidato a fusão** com `/nova-feature` |
| `/proxima-feature` | Executa a próxima feature `todo`: branch, implementação, PR, telemetria | Nenhuma — comando central do fluxo | Manter |
| `/revisar` | Revisa código da feature ativa contra critério de conclusão | Nenhuma | Manter |
| `/status` | Relatório de progresso (concluído/ativo/próximo/bloqueios) | Parcial com `/doctor` (ambos "relatam estado"), mas propósitos distintos (progresso de features vs. saúde estrutural) | Manter separado |
| `/doctor` | Diagnóstico de saúde estrutural (budgets, arquivos, MCPs, convenção de nomenclatura, telemetria) | Ver acima | Manter separado |
| `/archive` | Compacta `progress.md` em `progress-log.md` | Nenhuma | Manter |
| `/upgrade-sdd` | Atualiza a estrutura do projeto para nova versão da metodologia | **Sobreposição real com `forge-sdd update --upgrade`** (comando Go já implementado, determinístico, testado — ver `cmd/forge-sdd/main.go`) — o prompt de chat é um wrapper vago de uma linha para algo que o CLI já faz de forma confiável | **Candidato a aposentadoria**: documentar `forge-sdd update --upgrade` como o caminho oficial e remover o prompt de chat equivalente |
| `/c4-architecture` | Gera diagramas C4 (Mermaid) ad-hoc para qualquer descrição técnica | Sobreposição parcial com o C4 já gerado dentro de `/discovery` | Manter — útil para gerar/atualizar diagramas fora do fluxo de discovery (ex: documentar arquitetura existente) |

### Métrica de referência

Onboarding hoje: 17 comandos (11-12 de chat × paridade entre agentes + comandos CLI `init`/`update`/`doctor`/`destroy`/`version`). Se as duas fusões acima forem executadas numa rodada futura, a contagem de **comandos de chat** cai de 12 para 10 (~17%) sem perda de capacidade — `/novo-fix` vira documentação de um atalho de `/nova-feature`, e `/upgrade-sdd` é substituído pelo comando CLI já existente.

### Escopo desta feature

Nenhuma remoção ou fusão foi executada — esta feature entrega apenas a auditoria e a recomendação, para decisão explícita do usuário antes de qualquer remoção real de comando (que afeta muscle memory de usuários existentes e requer nota de breaking change nas release notes).
