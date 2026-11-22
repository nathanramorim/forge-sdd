# Proficiência do harness — o que se mede e o que falta

"Harness" aqui = contratos entre etapas, rules por ID, gate (Camada A), Reviewer (Camada B), fila e telemetria. Regra: métrica sem dado é **a medir**, nunca 0 (RN-5). Limiares são **propostos**, a calibrar com a linha de base; quem aprova é o dono.

## O que já se mede (evidência)
| Medida | Onde | Estado |
|---|---|---|
| Saúde estrutural | `forge doctor` (`cmd/forge-sdd/doctor.go`) | Existe na `main`; compara `.sddrc` com `.sdd-version`, **não** com o binário (A4) |
| Saúde do harness | `forge doctor --harness --json` (`internal/doctor/harness.go`, só em `bbd26d6`) | 7 checagens rodadas aqui: `versao` (aviso: projeto 4.5.0-beta, binário 5.0.0), `forge-json`, `skills-rules`, `contrato-skill-binario`, `pareceres-gravados` (aviso: 5 de 17 não decodificam), `worktrees`, `workers` (aviso: 2 workers 4.9.1-beta) → 0 falhas, 3 avisos |
| Telemetria por sessão | `forge session record` → `sdd/.metrics/session-*.json` (`internal/monitor/session_record.go`: tokens, turnos, `rework_lines`, `outcome`, `medicao`) | 97 arquivos; `outcome`: 87 approved, 9 blocked, 1 rejected (medido, contagem de linhas em todo o diretório) |
| Relatório | `forge report [--initiative] [--jobs]` | Soma tokens, modelo, duração |
| Avaliação | `forge eval run/compare` (`internal/monitor/eval.go`, só em `bbd26d6`) | **Só simulador** (`origem: simulado`): taxa de aprovação, voltas, tokens, segundos, intervenções; não chama modelo |
| Retrabalho por job | `bus.Job.Attempt`, `ReworkFeedback`, `max_rework_attempts` (`internal/bus/job.go`); vereditos em `sdd/discovery/*/verdicts/*.json` | Dado existe: 17 vereditos na árvore; 3 com `rework` no nome |
| Cobertura de rules por ID | frontmatter `regras:` de `.agents/rules/*.md` contra IDs de `constitution.md` | **Medido agora:** 36 de 45 IDs (80%); faltam `ARQ-10 DUP-03 DUP-04 DUP-06 FLX-04 SEG-03 SEG-06 TST-03 TST-04` (conta feita com `grep` sobre os arquivos; não há comando) |

## Veredito
**Não dá para dizer "o harness está proficiente" com número hoje.** O harness *funciona e se diagnostica* (doctor, contratos, gate), mas não existe um indicador de resultado: a avaliação é simulada, e a telemetria de custo está quebrada na origem (`tokens_input: 0` em 88 de 97 arquivos de sessão; só 1 marcado `medicao: estimado`). Dá para afirmar, com dado, apenas: 0 falhas e 3 avisos no `doctor --harness`; 80% de cobertura de rules por ID; 5 de 17 pareceres ilegíveis (29%).

## Métricas propostas
| ID | Métrica | Fonte de dados | Hoje | Limiar proposto (calibrar) |
|---|---|---|---|---|
| H1 | Taxa de gate verde de primeira = jobs `pr_review` com veredito `conforme` na tentativa 1 ÷ jobs `pr_review` | `verdicts/*.json` + `Job.Attempt` | **a medir** (17 vereditos, 5 ilegíveis; base pequena e mista) | ≥ 70% após 20 jobs; abaixo, revisar rules |
| H2 | Retrabalhos por PR = média de `Attempt - 1` | `bus.Job` (`done/`) e vereditos | **a medir** | ≤ 1,0; escalada (limite atingido) < 10% |
| H3 | Tokens por task | `session-*.json` com `medicao: real` | **a medir**: 88/97 com zero; precisa corrigir a gravação antes (`forge session record` sem flags estima; ver lição "nunca passar zeros") | sem limiar até haver 20 sessões `real` |
| H4 | Tempo do `init` ao primeiro PR | não existe marca de `init` nem evento "PR aberto" ligados | **a medir**: exige gravar `init_at` em `sdd/.sdd-version` ou `.metrics` e ler `PR.Number` do job | ≤ 1 dia de calendário em projeto de exemplo (hipótese) |
| H5 | Cobertura de rules por ID | frontmatter das rules × `constitution.md` | **80% (36/45)** medido | 100% dos IDs com pelo menos uma rule, ou lista de IDs "só do gate" |
| H6 | Pareceres legíveis | `doctor --harness` (`pareceres-gravados`) | **71% (12/17)** | 100% dos novos; antigos arquivados |
| H7 | Contrato skill ↔ binário coincide | `doctor --harness` (`contrato-skill-binario`) | ok | sempre ok (falha bloqueia) |
| H8 | Workers e projeto na versão do binário | `doctor --harness` (`versao`, `workers`) | 3 avisos | 0 avisos |
| H9 | Intervenções humanas por feature | `eval` (simulado) e `Events` do job | **a medir** (simulado não vale) | tendência decrescente |

## Entrega proposta (onda 3)
`forge report --proficiencia [--json]`, somente leitura (`ARQ-05`), que agrega H1 a H9, imprime "a medir" quando a fonte é vazia ou `estimado`, e sai com 3 se algum limiar aprovado for violado. Reusa `doctor --harness` para H6 a H8 e `internal/monitor` para H1 a H4. Saída saneada (`SEG-05`); vereditos lidos com decodificação estrita e arquivo ilegível contado, nunca executado (`SEG-01`, `FLX-08`). Teste não depende de identidade do git.

## Pré-requisitos e ordem
1. Corrigir a origem da telemetria (H3): sem isso, H3 fica "a medir" para sempre.
2. Registrar `init_at` (H4).
3. Só então publicar limiares no `doctor`. **Bloqueado até a 104 chegar à `main`** (`internal/doctor`, `eval`).

## Não verificado
Se `Attempt` é gravado de forma consistente nos jobs reais; se `.forge/runtime` do projeto tem jobs `done` para calcular H2 (a pasta `.forge` não existe neste worktree); a causa dos zeros de `tokens_input`.
