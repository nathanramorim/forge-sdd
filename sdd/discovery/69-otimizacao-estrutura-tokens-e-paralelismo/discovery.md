# Discovery 69 — Otimização de Estrutura, Tokens e Paralelismo

Status: `FLX-01` aprovado em 2026-11-13; fatiado em 7 features. Nenhum código alterado.

## Problema
O forge-sdd cresceu em três direções que hoje custam mais do que entregam:

1. **Interface que não será usada.** O Studio (`internal/studio`, 5.262 linhas, mais `three.min.js`) acompanha as estações visualmente. O dono do produto não quer usá-lo, mas ele segue no binário, no `FLOW.md`, nos templates e nas tasks da feat 68.
2. **Contexto fixo caro.** Um `/discovery` obriga a ler cerca de 65 KB antes de produzir qualquer coisa; 27 KB são o `sdd/features/index.md`, quase todo histórico `done`. Comandos, papéis e rules formam três camadas de indireção (adaptador → comando → papel → rules) replicadas em quatro árvores.
3. **Estado global em branch de feature.** `progress.md` (154 commits), `index.md` (131) e `sdd/.queue/*.json` (96 KB) são arquivos únicos editados por toda branch. Duas frentes paralelas sempre conflitam neles.

Além disso, a estrutura de pastas está pela metade: só as iniciativas 66, 67 e 68 usam `sdd/discovery/<id>-<nome>/`. Restam 14 discoveries soltos (42 arquivos) e 115 arquivos em `sdd/features/`.

## Para quem
- **Dono do produto / Líder:** quer rodar duas ou mais iniciativas em paralelo sem resolver conflito de arquivo de estado.
- **Agentes das estações:** precisam de um conjunto de leitura pequeno e previsível por comando.
- **Projetos que usam o forge-sdd:** recebem a mesma estrutura via `init`/`update`; a mudança não pode destruir o domínio deles (`SEG-02`).

## Decisões já tomadas (Clarify, 2026-11-13)
| Tema | Decisão do usuário |
|---|---|
| Studio | Remover do produto. O ACP permanece, sem interface. |
| Skills | Comandos e papéis são consolidados em skills por estação. |
| Paralelismo | Estado por iniciativa: somem `index.md` e `progress.md` globais. |
| Execução | Discovery escrito na sessão principal, sem subagente. |

## Proposta (o "como" macro)
Cinco frentes, uma iniciativa:

- **A. Remover o Studio.** Apagar pacote, subcomando, manual e citações. Revisar o escopo da feat 68 (tasks 68-04 e 68-05 deixam de existir).
- **B. Uma pasta por iniciativa.** Tudo que pertence a uma demanda vive em `sdd/discovery/<id>-<nome>/`: `discovery.md`, `criteria.md`, `plan.md`, `status.md`, `queue.json`, `features/`, `fixes/`. `sdd/features/` deixa de existir como raiz ativa.
- **C. Estado local à iniciativa.** `status.md` e `queue.json` ficam dentro da pasta. A visão geral é montada na leitura por `forge-sdd status`, inclusive lendo outras branches. Branches diferentes nunca tocam o mesmo arquivo.
- **D. Skills por estação.** 13 comandos e 7 papéis viram 5 skills com carregamento sob demanda. As rules continuam em `.agents/rules/` (o gate `P-23` depende desse caminho) e são citadas pelas skills.
- **E. Orçamento de contexto verificável.** Limites em bytes por conjunto de leitura, checados pelo gate e por teste.

## Escopo
**Dentro:** código Go do CLI, templates embutidos, golden files, artefatos de agente deste repositório, migração do legado para `sdd/archive/`, ADRs e ajuste da Constitution.

**Fora:** reescrever o conteúdo de discoveries e features antigos; alterar o protocolo ACP; o site (`site/`); mudar o formato de `sdd/.metrics/`.

## Jornada alvo
1. Líder roda `/discovery`; o ID vem de `forge-sdd next-id`, que consulta todas as refs e não colide com outra branch.
2. A pasta `sdd/discovery/<id>-<nome>/` nasce com os três artefatos e o `status.md`.
3. Em outra branch, uma segunda iniciativa faz o mesmo em outra pasta. Nenhum arquivo em comum.
4. `forge-sdd status --all-branches` mostra as duas frentes sem merge.
5. Cada agente lê só o `status.md` da iniciativa da branch atual e a skill da sua estação.

## Riscos
| Risco | Mitigação |
|---|---|
| Projetos em v3 com estrutura antiga | Migração no `update` com `--dry-run`, sem sobrescrever (`ARQ-05`, `SEG-02`) |
| Feat 68 em andamento edita o Studio | Remover as tasks de Studio da 68 antes de iniciar a frente A |
| Perda da visão geral sem `index.md` | `forge-sdd status` gera a visão; o histórico fica em `sdd/archive/` |
| Ferramentas sem skills nativas (Copilot) | Um adaptador fino por skill (`DUP-10`) |
| Mudança de regra constitucional sem registro | ADR no mesmo PR (`GOV-01`, `FLX-05`) |

## Decisões em aberto para o Gate 1
1. **Legado:** mover os 14 discoveries soltos e `sdd/features/` para `sdd/archive/` sem reescrever (proposta) ou migrar para o novo formato.
2. **Features e fixes sem discovery:** passam a criar uma pasta de iniciativa enxuta (só `status.md` e `features/`), eliminando `sdd/features/` (proposta), ou mantêm raiz própria.
3. **Branch base da 69:** *resolvido no Gate 1:* tip da `feat/68` em worktree separado (`release/v3.0.1` não contém o ACP).
4. **Versão:** a mudança de layout e a remoção do `studio` quebram contrato; proposta `4.0.0-beta`.
5. **Nome da pasta:** mantido `<id>-<nome>` (padrão de 66 a 68). Confirmar se o sufixo literal `-discovery` é desejado.
