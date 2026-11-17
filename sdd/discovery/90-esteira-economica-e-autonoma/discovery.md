# Discovery 90 — Esteira econômica e autônoma

Absorve a iniciativa 85 (modelo e esforço por sessão), que parou no Gate 1 em 15/11 com base desatualizada (arquivada em `sdd/archive/discovery/85-modelos-por-sessao-e-roteamento-economico/`).

## Problema
**Linha de base** (logs em `.forge/runtime/logs`, eventos `result` do `claude -p`, jobs de 16 e 17/11): 12 jobs somaram 276 turnos, 14,7 milhões de tokens relidos do cache, 632 mil gravados no cache e 146 mil de saída (US$ 5,85 a preço de tabela). Só o Builder da 88 usou 81 turnos e 8,3 milhões relidos, cerca de 102 mil por turno. O gasto cresce com o número de turnos vezes o tamanho do contexto.

1. **Não dá para medir.** O worker grava tokens estimados pelo tamanho da spec e do log (`estimateJobTokens`, `cmd/forge-sdd/worker.go`), sem cache, turnos, duração nem modelo, e com a versão fixa em `4.5.0-beta` (`internal/config/config.go`). `forge report` sai quase zerado; o painel mostra 869 mil tokens no Builder, contra cerca de 13,6 milhões relidos nos logs.
2. **Contexto gasto com ruído.** 87,5% do que volta das ferramentas ao Builder é saída de shell, inclusive `git status` com arquivos que a própria esteira deixa soltos. `forge status`, lido no início de todo comando, tem 12,8 KB; `forge status --compact`, que promete menos de 200 tokens, entrega 6,2 KB porque o "Ao Vivo" despeja o JSON cru do log.
3. **Passo mecânico na mão do modelo.** Commit, push, PR e registro do PR custam turnos e erram: a base vem do `base_branch` global (`main`), então o PR #87 abriu contra `main` (+3.908 linhas) e o #86 contra a branch de outro fix.
4. **Fix fora da esteira.** `forge dispatch` só lista `features/feat-*.md`; os fixes 84-03, 88-01 e 88-02 exigiram montar o job à mão, sem portão pré-build e com `task_ids` vazio.
5. **Dois leitores de dependência.** `queue.parseDependsOnContent` quebra "nenhuma (usa …)" em palavras, e `forge queue next` nunca liberaria a 89-01; `pipeline.readDependsOn` lê certo, mas só aceita `feat-…` (`DUP-01`).
6. **Archive que não arquiva.** `AutoArchiveInitiative` copia em vez de mover, repete a linha do progress-log a cada aprovação, grava telemetria zerada e roda na branch que estiver aberta.
7. **Sem escolha de modelo e esforço** (85): Builder e Reviewer usam `model: inherit`; não há como pôr um modelo ou esforço menor nos papéis mecânicos.

## Para quem
Dono do projeto que deixa a esteira rodando (modo normal ou avançado), quer gastar menos, não ser chamado depois do Gate 1 e ter números para comparar.

## Jornada
1. Aprova a discovery (ou o Reviewer aprova, no modo avançado).
2. A esteira despacha features e fixes pelo mesmo comando; o Go abre o PR na base certa.
3. Cada job grava o consumo real; `forge report` mostra por job e por iniciativa.
4. No fechamento, o archive move a pasta uma única vez, com os totais reais.
5. No Settings, escolhe modelo e esforço por sessão e vê o efeito no relatório.

## Regras de negócio
- **RN-1 Medir antes de otimizar:** cada melhoria mostra o número de antes (a linha de base acima) e o de depois.
- **RN-2 O que é mecânico vai para o Go**, não para mais texto nas skills (lição em `lessons.md`; orçamento da ADR-0003).
- **RN-3 Nada muda para quem não configurar:** sem modelo escolhido, o comando da CLI é idêntico byte a byte; `forge.json` antigo segue válido; `forge pipeline pr-opened` continua para quem abre o PR à mão.
- **RN-4 O Go nunca faz push para `main` nem para a base**, nunca mescla e nunca aprova PR (`FLX-02`, ADR-0004).
- **RN-5 A base do PR é a da iniciativa**, registrada na fila; o `base_branch` global vira só o padrão. Fix de iniciativa ainda aberta vai contra a branch da iniciativa.
- **RN-6 O archive move e é idempotente;** só roda na branch da iniciativa e sem task ou fix pendente. Aprovar o último job arquiva sozinho só em iniciativa do `/toca-ai` (promessa da 89, que roda sem ninguém); nas demais apenas avisa, porque é comum abrir fix depois da aprovação (84-01 a 84-03, 88-01 e 88-02).
- **RN-7 Consumo é dado:** US$ é preço de tabela e aparece rotulado; quem usa assinatura olha tokens e turnos. Medida estimada nunca se mistura com a real.
- **RN-8 Telemetria sem dado sensível** (`SEG-05`): só contadores, modelo, duração e IDs.

## Escopo
**Dentro:** telemetria real e relatório; `forge status` curto para agentes; base da iniciativa e PR aberto pelo Go; fix no `dispatch`; um leitor de dependências e checagem das specs antes do build; archive que move; modelo e esforço por sessão; ADR-0014.

**Fora (próxima onda, depois de medir):** Builder por tarefa com contexto novo; mapa de código no `forge context`; lições automáticas a partir dos pareceres; `forge release beta`; merge automático de PR de fix; promover as ADRs 0009 a 0013; arquivar em lote as iniciativas 66 a 84 (o comando lista, a pessoa decide).

**Cortado depois da medição:** resumo dos testes e limpeza total do `git status`. No Builder, `go test` foi 3% e `git status` 2,4% da saída das ferramentas; a leitura de código (`sed`, `cat`, `grep`, `Read`) é o grosso e fica para o mapa de código. A 90-06 já tira o ruído que a esteira cria (cópias do archive e registros repetidos); a telemetria por job continua em `sdd/.metrics/` (`FLX-04`) e entra no commit do fechamento.

## Riscos
| Risco | Mitigação |
|---|---|
| CLI sem evento de resultado (agy, gemini, codex, copilot) | Registro marcado `estimado`, separado do real no relatório |
| `gh` ausente ou sem login | O comando falha com motivo claro no inbox; `pr-opened` segue como caminho manual |
| Status curto esconder algo de quem usa terminal | Curto só sem terminal; `--full` mostra tudo; menu e painel vivo não mudam |
| Archive mover pasta com fix a caminho | Só move sem task ou fix pendente e na branch da iniciativa; aprovar o último job apenas avisa |
| Modelo menor piorar a entrega | O padrão continua o da CLI; o relatório mostra turnos e retrabalho por modelo |
