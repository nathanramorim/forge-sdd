# Comando: split-features

**Uso:** `/split-features [iniciativa]`.

1. **Brief primeiro:** rode `forge brief` no início; se o estado for `rascunho`, proponha `/constitution` antes de fatiar tarefas.
2. **Gate (`FLX-01`):** leia `plan.md` e `criteria.md` da iniciativa e confirme a revisão humana da Spec.
3. **ADRs e rules:** carregue as ADRs e as rules aplicáveis antes de fatiar; nenhuma task pode violá-las, e cada task cita as que a governam na seção `## Decisões do projeto aplicadas`.
4. **Critérios de quebra:** fatie quando a feature cruzar mais de um contexto delimitado ou tiver mais de 7 tarefas. Sem dependência circular; cada task testável sozinha; ordem de baixo para cima (infraestrutura, domínio, aplicação, interface).
5. **Sem branch nova aqui (`FLX-01`, ADR-0022):** o split **não** cria branch. A spec já voltou à principal pelo PR dela; cada feature terá **a sua própria branch** (`feat/<id>-<nome>-<sid>`), aberta pelo Builder a partir da principal, com o seu PR em draft. Só cria specs em `sdd/discovery/<id>-<nome>/features/`, na branch atual.
6. **Tasks:** crie `sdd/discovery/<id>-<nome>/features/feat-<id>-<nn>-<nome>.md` com `Modelo do módulo`, `Decisões do projeto aplicadas`, `O que já existe`, `Dependências novas`, `Rules e IDs da Constituição Aplicáveis` e `Critérios de Aceitação Executáveis`.
7. **Estado:** atualize `status.md` da iniciativa; não escreva índice nem progresso globais.

**Visualizar artefatos (antes do Gate 1)**

Use AskUserQuestion com as opções:
- "Abrir pasta no Finder" → `open sdd/discovery/<id>-<nome>/`
- "Abrir no VS Code" → verifique `command -v code && code sdd/discovery/<id>-<nome>/` (fallback: Finder se VS Code ausente)
- "Ver lista dos arquivos" → liste os arquivos em `sdd/discovery/<id>-<nome>/`
- "Pular"

Execute a ação escolhida. Consulte `sdd/memory/clarify.md` → seção "Visualização pós-spec" para comportamento detalhado (macOS-only, fallbacks).

8. **Fila:** com a confirmação do humano, `forge queue enqueue --task <spec> --task-id <id> --rules <IDs>` para cada task, na ordem de dependência.
9. **Pipeline (existe `.forge/config/forge.json`):** depois da confirmação humana que enfileirou as tasks, rode `forge dispatch <iniciativa>` (gate e job para o Builder), informe "entregue ao Builder" e **pare**: o Builder e o Reviewer rodam nas outras sessões. Esta sessão não implementa, não edita `forge.json`, `card-evidence.json` nem o código da feature; pendência do gate vai para a pessoa. Com `pm_mcp.require_existing_card: false` o despacho não pede card.
10. **Telemetria:** `forge session record --feature "sdd/discovery/<id>-<nome>/features/" --tokens-input <N> --tokens-output <N> --outcome approved --criterio-atendido=true --files-touched <lista>`. Nunca passe zeros: se não souber os tokens, omita as flags (o CLI vai estimar).

Use `/nova-feature` apenas para adições manuais e isoladas.
