# Plano — Iniciativa 105

3 ondas, ordem de baixo para cima (a base de cada onda primeiro). Máximo de 7 tasks por onda. Cada task vira `features/feat-105-NN-*.md` no `/split-features`, com critério executável e teste negativo (P-25). Rules citadas por ID da Constitution.

## Onda 1 — Obsoletos e limpeza (independente de tema e da 104)
| Task | O que | Depende de | Paralela | Rules/IDs |
|---|---|---|---|---|
| 1.1 | ADR de obsolescência: 5 estações (parcial sobre ADR-0003), depreciação de `queue stations`/`attach`/`queue handoff`, decisão sobre `lider`; próxima ADR livre após a 0020 | decisão do dono (`A confirmar`) | — | `FLX-05`, `GOV-01`, `ARQ-06` |
| 1.2 | Corrigir texto: `Long` do raiz, `docs/pipeline.md` e `flows.md` (versão), "5 estações" nos templates/goldens, `docs/acp-ide-setup.md`, `sdd-acp.md` | 1.1 | com 1.3 e 1.4 | `COD-05`, `TST-02`, `DUP-06` |
| 1.3 | Avisar (O3, O4, O5, O7, O13; cobra `Deprecated`) em `queue stations`, `attach`, `queue handoff`, `/implement`, `/proxima-feature`; nada removido | 1.1 | com 1.2 | `ARQ-06`, `COD-04`, `TST-03` |
| 1.4 | `doctor` compara `.sdd-version` com o binário; lista de features só com `--verbose` | — | com 1.2 | `ARQ-01`, `SEG-01`, `TST-03` |
| 1.5 | `status` esconde iniciativa com fila `completed`/`approved` (ADR-0002); `--all` mostra | — | com 1.4 | `ARQ-10`, `DUP-06`, `TST-03` |
| 1.6 | `lider`: trocar a descrição e o texto (sem renomear nem remover); testes de `launcher`, `tmux`, `preflight`, `up_sessions` verdes | 1.1 | — | `COD-04`, `TST-02`, `FLX-02` |

## Onda 2 — Navegação e tema do CLI (depende da 1.2 só para não regravar texto duas vezes)
| Task | O que | Depende de | Paralela | Rules/IDs |
|---|---|---|---|---|
| 2.1 | `internal/ui`: paleta do Forge, estilos, `NO_COLOR`, TTY e largura; verificar `huh`/`lipgloss` no context7; promover `lipgloss` a direto | — | — | `ARQ-02`, `DUP-01`, `DUP-04`, `SEG-07`, `TST-04` |
| 2.2 | Tema nos menus `huh` (`internal/survey`) e no menu de `menu.go` (opção recomendada marcada) | 2.1 | com 2.3 | `ARQ-01`, `DUP-03`, `TST-04` |
| 2.3 | Erros: grupos sem ação e subcomando desconhecido saem !=0, uma mensagem, sugestão em pt-BR; `--version` | 2.1 | com 2.2 | `COD-03`, `SEG-05`, `ARQ-05`, `TST-03` |
| 2.4 | `--help` agrupado (Começar, Pipeline, Qualidade, Avançado) em pt-BR | 2.3 | com 2.5 | `ARQ-06`, `COD-05`, `TST-02` |
| 2.5 | `doctor` e `status` temáticos e curtos; golden sem cor, sem TTY e a 40 colunas | 2.1 | com 2.4 | `TST-02`, `TST-04`, `SEG-05` |
| 2.6 | Menu raiz com "próximo passo" (estado: sem `sdd/`, constituição rascunho, sem `forge.json`, fila pronta, PR pronto) menu ao fim do `init` e menu de agente no `init` (Claude Code, Copilot, Gemini, Codex; o detectado no PATH vem recomendado, sem padrão fixo; sem TTY exige `--agent` e falha se faltar) | 2.2, 2.5 | — | `FLX-01`, `ARQ-05`, `SEG-01`, `SEG-03` |

## Onda 3 — Medição de proficiência (depende da 104 na `main`; justificativa: usa `internal/doctor`, `eval`, `harness`)
| Task | O que | Depende de | Paralela | Rules/IDs |
|---|---|---|---|---|
| 3.1 | Corrigir origem da telemetria (tokens zero) e gravar `init_at` | 104 mesclada | com 3.2 | `FLX-04`, `SEG-01`, `TST-04` |
| 3.2 | Agregador H1, H2, H5, H6 a partir de vereditos, jobs e rules; "a medir" sem dado | 104 mesclada | com 3.1 | `ARQ-01`, `DUP-01`, `FLX-08`, `TST-03` |
| 3.3 | `forge report --proficiencia [--json]`, somente leitura, saída saneada e tema da onda 2 | 3.1, 3.2, 2.1 | — | `ARQ-05`, `SEG-05`, `TST-02` |
| 3.4 | Limiares aprovados pelo dono viram checagem de `doctor --harness` (saída 3 se violar) | 3.3, aprovação | — | `ARQ-06`, `GOV-01`, `FLX-06` |
| 3.5 | Doc `harness` e consolidação de lições da onda | 3.4 | — | `GOV-03`, `DUP-06` |

Paralelismo: na onda 1, 1.2/1.3/1.4/1.5 em paralelo depois da 1.1 (1.4 e 1.5 não precisam dela); na onda 2, 2.2 e 2.3 depois da 2.1; na onda 3, 3.1 e 3.2. Ondas 1 e 2 podem rodar em paralelo exceto 1.2 antes da 2.4. Sequência com a 106 (mexe em templates, goldens, `split-features.md`, `hooks.go`, `doctor` e menus): as tasks 1.2, 1.4 e 2.6 desta e as 1.7, 2.5 e 2.6 da 106 tocam os mesmos arquivos e entram em série, ordem definida no `/split-features`. A 3.1 inclui o conserto dos tokens zerados (aprovado). Arquivar as iniciativas 79 a 88 com `/archive` só depois da 1.5. Uma branch por iniciativa (`feat/105-experiencia-cli-e-harness`), nunca por task (`FLX-02`).
