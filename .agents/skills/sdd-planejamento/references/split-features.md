# Comando: split-features

**Uso:** `/split-features [iniciativa]`.

1. **Gate (`FLX-01`):** leia `plan.md` e `criteria.md` da iniciativa e confirme a revisão humana da Spec.
2. **ADRs e rules:** carregue as ADRs e as rules aplicáveis antes de fatiar; nenhuma task pode violá-las, e cada task cita as que a governam.
3. **Critérios de quebra:** fatie quando a feature cruzar mais de um contexto delimitado ou tiver mais de 7 tarefas. Sem dependência circular; cada task testável sozinha; ordem de baixo para cima (infraestrutura, domínio, aplicação, interface).
4. **Branch única (Regra 15):** `git branch --list feat/<id>-*`; pergunte a branch de partida (padrão `main`) e se retoma branch existente; `git checkout -b feat/<id>-<nome>` antes de criar as specs.
5. **Tasks:** crie `sdd/discovery/<id>-<nome>/features/feat-<id>-<nn>-<nome>.md` com `Modelo do módulo`, `O que já existe`, `Dependências novas`, `Rules e IDs da Constituição Aplicáveis` e `Critérios de Aceitação Executáveis`.
6. **Estado:** atualize `status.md` da iniciativa; não escreva índice nem progresso globais.
7. **Fila:** com a confirmação do humano, `forge-sdd queue enqueue --task <spec> --task-id <id> --rules <IDs>` para cada task, na ordem de dependência.
8. **Telemetria:** `forge-sdd session record --feature "sdd/discovery/<id>-<nome>/features/" --outcome approved --criterio-atendido=true`.

Use `/nova-feature` apenas para adições manuais e isoladas.
