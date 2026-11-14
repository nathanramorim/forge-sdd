# Comando: split-features

**Uso:** `/split-features [iniciativa]`.

1. **Brief primeiro:** rode `forge-sdd brief` no início; se o estado for `rascunho`, proponha `/constitution` antes de fatiar tarefas.
2. **Gate (`FLX-01`):** leia `plan.md` e `criteria.md` da iniciativa e confirme a revisão humana da Spec.
3. **ADRs e rules:** carregue as ADRs e as rules aplicáveis antes de fatiar; nenhuma task pode violá-las, e cada task cita as que a governam na seção `## Decisões do projeto aplicadas`.
4. **Critérios de quebra:** fatie quando a feature cruzar mais de um contexto delimitado ou tiver mais de 7 tarefas. Sem dependência circular; cada task testável sozinha; ordem de baixo para cima (infraestrutura, domínio, aplicação, interface).
5. **Branch única (Regra 15):** `git branch --list feat/<id>-*`; pergunte a branch de partida (padrão `main`) e se retoma branch existente; `git checkout -b feat/<id>-<nome>` antes de criar as specs.
6. **Tasks:** crie `sdd/discovery/<id>-<nome>/features/feat-<id>-<nn>-<nome>.md` com `Modelo do módulo`, `Decisões do projeto aplicadas`, `O que já existe`, `Dependências novas`, `Rules e IDs da Constituição Aplicáveis` e `Critérios de Aceitação Executáveis`.
7. **Estado:** atualize `status.md` da iniciativa; não escreva índice nem progresso globais.
8. **Fila:** com a confirmação do humano, `forge-sdd queue enqueue --task <spec> --task-id <id> --rules <IDs>` para cada task, na ordem de dependência.
9. **Pipeline (existe `.forge/config/forge.json`):** depois da confirmação humana que enfileirou as tasks, rode `forge-sdd dispatch <iniciativa>` (gate e job para o Builder), informe "entregue ao Builder" e **pare**: o Builder e o Reviewer rodam nas outras sessões. Esta sessão não implementa, não edita `forge.json`, `card-evidence.json` nem o código da feature; pendência do gate vai para a pessoa. Com `pm_mcp.require_existing_card: false` o despacho não pede card.
10. **Telemetria:** `forge-sdd session record --feature "sdd/discovery/<id>-<nome>/features/" --outcome approved --criterio-atendido=true`.

Use `/nova-feature` apenas para adições manuais e isoladas.
