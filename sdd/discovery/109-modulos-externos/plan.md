# Plano — 109 (módulos externos; o quadro do GitHub Project é o primeiro módulo)

Ordem de baixo para cima: primeiro prender o comportamento atual (byte a byte), depois contrato, confiança e leitura do campo, depois execução, eventos e comandos, por último a suíte final e a documentação. Até 7 tasks por onda. Cada task traz teste negativo (`TST-03`) e passa em `forge gate --criterios`. Nenhuma task toca a esteira (106/107): só emite eventos nos pontos que já existem.

## Parte A — Core (este repositório)

### Onda 1 — Baseline e contrato
| # | Task | Depende de | Paralelo | Rules |
|---|---|---|---|---|
| 109-01 | **Baseline byte a byte, antes de qualquer mudança:** goldens de `forge --help`, `forge doctor`, `forge status`, `forge project` (comando desconhecido do cobra), códigos de saída e uma execução completa da esteira com fixtures; os testes passam hoje e passam depois | — | com 109-02 | `TST-02`, `TST-03`, `TST-04` |
| 109-02 | `internal/modulo`: tipos do manifesto e do envelope v1, versão do contrato (constante única), validação na borda, recusa com a mensagem "o módulo X exige o contrato N; este forge fala M", campo desconhecido ignorado | — | com 109-01 | `ARQ-01`, `ARQ-02`, `DUP-01`, `DUP-04`, `SEG-01`, `TST-03` |

### Onda 2 — Confiança, campo e eventos
| # | Task | Depende de | Paralelo | Rules |
|---|---|---|---|---|
| 109-03 | Registro de confiança (`confianca.json` na pasta do usuário, gravação segura), `sha256` do binário e do manifesto, autorização por versão, recusa se mudou; caminhos validados (sem link para fora) | 109-02 | com 109-04, 109-06 | `SEG-01`, `SEG-02`, `SEG-04`, `TST-03` |
| 109-04 | Leitor de `Módulos permitidos` **pelo leitor único da tabela** (106/0.5, branch base): `nenhum` e ausente = nada acontece, nenhuma chamada de disco do usuário nem de processo; valor inválido recusado só no `doctor` | **106/0.5, 106/0.8**, 109-02 | com 109-03 | `DUP-01`, `SEG-01`, `GOV-01`, `TST-03` |
| 109-06 | Emissão de eventos v1 nos pontos que **já existem na main** (`queue next`, `queue complete`, `queue review`, marcador do Gate 1, PR aberto e mesclado), melhor esforço, sem bloquear; payload semântico, sem caminho; se não há módulo ativo, não faz nada | 109-02 | com 109-03 | `ARQ-02`, `FLX-06`, `SEG-05`, `TST-03` |

### Onda 3 — Execução
| # | Task | Depende de | Paralelo | Rules |
|---|---|---|---|---|
| 109-05 | Carregador e execução: `LookPath` injetável, pasta do usuário, ambiente por lista, tempo limite, saída limitada e saneada (`SEG-08`), códigos 0 a 5 mapeados a aviso; **quebra deliberada da lista fixa de comandos** documentada | 109-03, 109-04, 109-12 | com 109-07 | `SEG-03`, `SEG-05`, `SEG-08`, `ARQ-02`, `TST-03` |
| 109-07 | Instantâneo `forge modulo contexto --json` (somente leitura): iniciativas, estados, PRs, `wip_humano`, `modo`, `max_parallel_tasks`; sem ler arquivo que o módulo possa mudar | 109-06 | com 109-05 | `ARQ-01`, `SEG-05`, `TST-03` |

### Onda 4 — Comandos, ajuda e doctor
| # | Task | Depende de | Paralelo | Rules |
|---|---|---|---|---|
| 109-08 | `forge modulo` **oculto** (`listar`, `instalar --de --sha256 [--dry-run]`, `aprovar`, `remover`) com menu de autorização; comandos dinâmicos dos módulos autorizados (`DisableFlagParsing`, terminal herdado) | 109-03, 109-05 | com 109-09 | `ARQ-01`, `ARQ-05`, `SEG-02`, `FLX-01`, `TST-03` |
| 109-09 | `--help` com a seção "Módulos" só se houver módulo ok; checagem `modulos` do `doctor` só se o campo ≠ `nenhum` (aviso, nunca falha, `ExitFalha` intacto); capacidade `doctor` do módulo com prefixo | 109-05 | com 109-08 | `ARQ-01`, `TST-02`, `TST-03` |

### Onda 5 — Suíte final e documentação
| # | Task | Depende de | Paralelo | Rules |
|---|---|---|---|---|
| 109-10 | Suíte de byte-identidade completa (módulo ausente, quebrado, lento, incompatível, hash mudado) com um módulo falso de teste (`TestMain` que reexecuta o binário de teste); mede o custo zero por contagem de chamadas | 109-01, 109-08, 109-09, 109-06, 109-07 | com 109-11 | `TST-03`, `TST-04`, `COD-03` |
| 109-11 | **Sem ADR** (é da 109-12): linha em `sdd/spec/concepts.md` (dono: `internal/modulo`), texto mínimo "Módulos externos (opcionais)" na doc, **sem** linha nos comandos nem nos templates | 109-08, 109-09 | com 109-10 | `GOV-01`, `DUP-01` |

### Onda 4 do core — ADR antecipada
| # | Task | Depende de | Paralelo | Rules |
|---|---|---|---|---|
| 109-12 | ADR da execução de módulos externos: só o PR da ADR, texto curto (contrato, confiança, quebra da lista fixa `git`/`gh`); número reservado no merge (0021 é da 105, 0022 da 106/0.8) | nenhuma | com 109-04 e 109-07 | `FLX-05`, `SEG-03`, `GOV-01` |

**Alteração em task de outra iniciativa (sem PR novo de Constituição):** a **106/0.8** troca a linha `Quadro de acompanhamento` por `| Módulos permitidos | nenhum |`, troca `project` por `modulo` na `ARQ-06` e reescreve a seção 5 da rule `sdd-modos-de-projeto.md` (texto em `discovery.md`); 9.173 de 9.216 bytes (medido; **não editei a Constituição**).

**Task nova na 107 (criada em 20/11/2026 por decisão do dono; PR #122, aguarda aprovação formal da 107):** `107-13` início automático do Builder local (Todo para Building: spec na principal, WIP humano 3, no máximo 3 tentativas, só com a sessão aberta, comando sob `forge pipeline`), no core, antes na 108/3.7. Depende de 107-12, 106/0.5, 106/0.8, 106/0.7 e 106/2.4. Entra na onda 8 abaixo.

## Novo fecho de dependências (o que muda; ordem de merge refeita na seção `## Ordem das ondas do core`)
Quase nada some por demérito: a esteira não mudou. Muda o **motivo** (deixa de ser "pré-requisito da 108").

| Task externa | Continua necessária por si? | Pré-requisito do módulo? |
|---|---|---|
| PR único da Constituição (106/0.8) | **sim** (modos do atestado, WIP, e agora `Módulos permitidos`) | **sim** (campo) |
| 106/0.5 (leitor único da tabela) | **sim** (atestado, 107-12) | **sim** (109-04) |
| 106/0.2 (`internal/vcs`) | sim (PR em draft) | não |
| 106/0.7 (estado por feature) | sim (RN-23) | não; o emissor 109-06 muda de fonte e o evento fica igual |
| 106/1.1, 1.3, 1.6, 0.6, 2.1, 2.2, 2.3, 2.4, 3.2 (local) | sim, todas são a esteira | não |
| 107-01, 107-02, 107-12 | sim (laço, WIP humano) | não; o módulo só **lê** `wip_humano` no instantâneo |

O dono decide se a "prioridade 0 por causa da 108" cai (ver decisão 6 do `discovery.md`); eu **não** reduzi nenhum escopo.

**Fecho recomendado antigo: SUBSTITUÍDO** pela seção `## Ordem das ondas do core (substitui o fecho da prioridade 0)` abaixo (decisão 6 do dono). Em particular, a coluna "Onda" antiga (106/0.8, 109-01, 109-02 na onda 1; 109-04 na 3) e a linha "cada onda do fecho tem até 3 tasks" não valem mais.

## Parte B — Repositório do módulo (`forge-sdd-project`; só descrito, nada criado)
Ondas do repositório do módulo, rodadas na esteira SDD dele.

| Onda | # | Task | Depende de |
|---|---|---|---|
| 1 | P-00 | **Bootstrap (abaixo). FUTURA:** só começa depois de o dono criar o repositório **privado** (visibilidade decidida em 20/11/2026; licença em aberto) | o dono criar o repositório (fora desta tarefa) |
| 2 | P-01 | Cliente do contrato v1 (manifesto, envelope, saída, códigos), `gh` falso e fixtures | P-00 |
| 2 | P-02 | `Etapa` (7 valores), `EtapaDoEstado` sobre o evento (108/1.1) e config do Project (108/1.4) | P-00 |
| 2 | P-03 | `ClassificaEtiquetas`, `PrefixoDoTitulo`, via rápida (108/1.7, 2.7) | P-00 |
| 3 | P-04 | Adaptador `gh project` leitura e `forge project status` (108/1.2, 1.3, 1.5) | P-01, P-02 |
| 3 | P-05 | `Mover`, `Vincular`, `forge project move` (108/2.1, 2.2) | P-04 |
| 4 | P-06 | Reação a `feature.estado_mudou`, `spec.aprovada`, `pr.aberto` (108/2.3, 2.6, 4.1) | P-05, P-03 |
| 4 | P-07 | Sincronia e divergência sem canal de volta (108/3.1, 3.2) e capacidade `doctor` (108/3.3) | P-06 |
| 5 | P-08 | Capacidade `pendentes` (108/3.4), **só quando a interface `FonteDePendencias` existir na 107** | 107, P-07 |
| 5 | P-09 | `forge project rotina` (108/3.5), docs migradas (108/4.2), `criar-project.sh` (108/2.5, **dono**) | P-07 |
| 6 | P-10 | Prova ponta a ponta com Project privado (108/3.6, **dono**) | P-09 |

**P-00 Bootstrap do repositório do módulo (só descrição):**
1. **Versão do forge-sdd:** a estável mais recente. Em 20/11/2026: npm `dist-tags.latest = 5.0.0` (comando `npm view @nathanramorim/forge-sdd dist-tags`), release `v5.0.0` marcada como `Latest` no GitHub (`gh release list`), `sdd/releases/history.md` "Versão 5.0.0 (Estável)"; o `beta` é `4.9.1-beta` e **não** serve. Regra para saber no dia: a tag `latest` do npm, conferida com a release `Latest` do GitHub e com o primeiro bloco `(Estável)` do `history.md`; divergência: parar e perguntar.
2. **Instalação verificável:** versão **exata** (`npm install -g @nathanramorim/forge-sdd@<versão>`), depois comparar `npm view @nathanramorim/forge-sdd@<versão> dist.integrity` com a instalada e conferir `forge-sdd version`. Se a release do GitHub trouxer lista de hashes, conferir também (**não verifiquei que existe**).
3. **Configurar:** `forge-sdd init --yes --agent claude --name forge-sdd-project --stack go --lang pt-BR` (flags confirmadas em `cmd/forge-sdd/main.go`), Constituição preenchida com a stack do módulo (Go, stdlib), os campos de modo e o **contrato mínimo do core** (`contrato.min = 1`); o manifesto repete o mínimo e o `doctor` do core recusa módulo fora da faixa.
4. **Diagramas:** reaproveitar os da 108 (`forge-108-kanban*`, `forge-108-bpmn-ciclo`, `-fechamento`, `-sincronia`, `forge-108-ciclo`) e gerar os que faltarem com o diagram-studio (assinatura Forge SDD).
5. **README:** com o plugin readme-studio (skill `readme-comercial`); **o plano do README é mostrado ao dono antes de gravar**.
6. **Licença e visibilidade:** **privado, instalação por convite** (decidido pelo dono em 20/11/2026); licença ainda em aberto. O repositório é criado pelo dono, **só depois** do Gate 1, e fora desta iniciativa; esta task fica **futura** até lá. Instalação do convidado: baixa o asset com o próprio acesso e usa `--de <arquivo> --sha256` (sem token no core).
7. **Critérios:** teste sem rede; `forge doctor` do projeto com o módulo falso e verde; versão mínima do contrato registrada no README, no manifesto e na Constituição do módulo.

## O que da 108 vira a spec do módulo e o que se descarta
- **Vira a spec do módulo (reaproveitada):** `discovery.md` e `criteria.md` da 108 (RN-1 a RN-16, mapeamento estado-coluna, etiquetas, prefixo, via rápida, diagramas Kanban e BPMN, issues, `criar-project.sh`), menos o que sobe ao core.
- **Sobe ao core (novo ou 107):** leitura do campo e custo zero (109-04), emissão de eventos (109-06), `FonteDePendencias` (interface, 107), início automático (107-13 proposta).
- **Descarta:** os campos `Quadro de acompanhamento` e a nota dele na Constituição (migram para `sdd/modulos/project.json`); as linhas de aviso nos comandos e templates (3.5, quebrariam o byte a byte); o canal "arrastar como pedido que abre menu do core" (3.2, adiado); a ADR da 108 no core (vai para o repositório do módulo); `ARQ-06` com `project`.

## Ordem das ondas do core (substitui o fecho da prioridade 0)
Decisão do dono (chat, 20/11/2026, decisão 6): a ordem é reaberta **sem cortar escopo**; a prioridade 0 da 108 deixa de ser o fecho; o quadro vira módulo e é o último a construir. Regras desta tabela: onda = lote de merge (até 7 tasks); Builders partem da `main` depois da onda anterior mesclada; **toda dependência aponta para onda anterior** (conferida nos planos reais, na branch de cada iniciativa; a 106 em `38f12e5`). "106/x" é task da 106; 105-NN é o arquivo `feat-105-NN` (105-01 a 07 = tasks 1.1 a 1.7; 105-08 = 3.6; 105-09 = 2.7). `*` = adicionada por ser pré-requisito de task do inventário (não estava nele).

| Onda | Tasks (ordem de merge) | Observação |
|---|---|---|
| 0 | Merges do dono: fixes #130, #131, #132; specs #120 (106), #121 (105), #122 (107, com 107-13), #154 (109, só após o Gate 1); #153 (108) **parada** | sem Builder |
| 1 | 105-01, 105-04, 105-05, 105-02, 105-03, 105-07, 105-06 | PRs #123 a #129 **já construídos**: só merge, na ordem do plano da 105 |
| 2 | 106/0.8, 106/0.2, 106/1.1, 106/0.1*, 107-01, 109-01, 109-02 | todas sem dependência; os Builders podem rodar em paralelo à onda 1 (arquivos disjuntos); merge depois dela |
| 3 | 106/0.5, 106/0.7, 106/1.3, 107-02, 109-03, 109-06 | **5.1.0-beta ao fim desta onda** (tag na branch, PR aberto, FLX-07) |
| 4 | 106/1.6, 106/0.6, 106/1.2, 107-12, 109-04, 109-07, 109-12 | 109-04 precisa de 106/0.5 e 0.8; 109-12 (só a ADR) antes da 109-05; onda com 7 tasks, nada movido |
| 5 | 106/2.1, 106/3.2 (local), 106/1.4, 107-04, 109-05 | 109-05 depende de 109-12 (onda 4) |
| 6 | 106/2.2, 106/1.5, 106/1.7*, 107-05, 107-07, 107-09, 109-08 | |
| 7 | 106/2.3, 106/2.4, 107-03, 107-06, 107-08, 109-09 | |
| 8 | 106/2.6, 107-13, 107-10, 106/3.1, 109-10, 109-11 | 107-13 só depois de 106/2.4; **5.1.0 estável depois desta onda** |
| 9 | 106/2.7, 106/3.3*, 105-09, 105 onda 2: 2.1, 105 onda 3: 3.1 e 3.2 | 105-09 só depois de 106/2.6 |
| 10 | 106/3.4, 106/3.6, 107-11, 105 onda 2: 2.2, 2.3, 2.5, 105-08 | |
| 11 | 105 onda 2: 2.4, 2.6; 105 onda 3: 3.3 | |
| 12 | 105 onda 3: 3.4 (aprovação do dono nos limiares), 3.5; 104-12 e 104-10 | decisão do dono pendente para 104-12/10 |

**Depois (fora do core; repositório do módulo, privado, futuro):** P-00 a P-10 (parte B), na esteira SDD do repositório do módulo. O quadro (a antiga 108) só entra aí.

**Opcionais do core (decisão do dono, 20/11/2026):** as tasks 106/0.4, 2.5 e 3.7 ficam no plano da 106 como opcionais do core, fora das ondas; decidir depois.

**Caminho crítico:** 106/0.2 (onda 2) → 106/0.7 (3) → 106/1.6 (4) → 106/2.1 (5) → 106/2.2 (6) → 106/2.4 (7) → 107-13 (8); e 106/2.6 (8) → 106/2.7 (9) → 106/3.6 (10). Do 109: 109-02 (2) → 109-03 (3) → 109-04 (4) e 109-12 (4) → 109-05 (5) → 109-08 e 109-09 (6 e 7) → 109-10 (8).

**Alerta de conteúdo da 5.1.0 (decisão do dono):** pela tabela, **depois da onda 5 o módulo ainda não é usável**: o comando `forge modulo` (109-08) e a checagem do `doctor` (109-09) vêm nas ondas 6 e 7, e a suíte de byte a byte (109-10, RN-1 do requisito inegociável) e a ADR (109-11) na onda 8. A 5.1.0 literal "depois da onda 5" levaria só a infraestrutura e o início de execução (109-05), sem a prova de RN-1. **Decidido pelo dono (20/11/2026):** 5.1.0 estável **depois da onda 8** (primeira com 107-13 e a suíte 109-10); a beta ao fim da onda 3 contém só 105 onda 1, base da 106 (0.8, 0.2, 1.1, 0.5, 0.7, 1.3), 107-01/02 e 109-01/02/03/06, **sem** efeito visível do módulo. A beta fica ao fim da onda 3.

**Alerta de governança (FLX-05, SEG-03):** a 109-05 (onda 5) quebra de propósito a lista fixa de comandos (`git`/`gh`) e isso exige ADR; **decidido pelo dono (20/11/2026):** a ADR sobe para a task 109-12 (onda 4, só o PR da ADR), antes da 109-05; a 109-11 cobre só o texto mínimo e a documentação.

**Colisões de arquivo e quem vai antes**
| Arquivo | Quem toca (onda) | Ordem |
|---|---|---|
| `cmd/forge-sdd/doctor.go`, `internal/doctor` | 105-04 (1); 106/1.5 (6); 109-09 (7); 105 onda 2 (2.5, onda 9); 106/0.4 e 104-03 já na main | 105-04, depois 106/1.5, depois 109-09, por último o tema da 105; a 105 onda 2 **regenera os goldens da 109-01** |
| `cmd/forge-sdd/status.go` | 105-05 (1); 106/1.6 (linhas `Modo:`/`Atestado:`, onda 4); 105 onda 2 (2.5, onda 9) | 105 primeiro; 106/1.6 reaproveita o filtro |
| `cmd/forge-sdd/session.go` | **não verificado**: nenhum plano citou o arquivo; só `internal/acp/session.go` (68) | conferir no `/split-features` de cada task |
| Templates e goldens (`internal/scaffold`) | 105-02, 105-03 (1); 109-01 (2); 107-04 (5); 106/2.6 (8); 105-09 e 105 onda 2 (9 em diante) | 109-11 não toca template (só doc); 106/2.6 depois da onda 1 da 105 (regra do plano da 106) |
| `internal/scaffold/hooks.go` | 106/1.4 (5), 106/1.5 (6) | a 105 não toca no hook; 106/1.4 antes da 1.5 |
| Constituição e rule `sdd-modos-de-projeto.md` | só 106/0.8 (2); 109-04 e 107-12/13 apenas leem | ninguém mais edita; 9.173 de 9.216 bytes |
| `cmd/forge-sdd/main.go` (`AddCommand`) | 105-03 (1), 109-08 (6), 107-13 (8, subcomando de `pipeline`) | sem colisão na mesma onda |
| `internal/pipeline` (`laco.go`, `leque.go`, `dispatch.go`) | 106/2.1, 2.3, 2.4 (5 a 7), 107-02/05/07/09 (3 a 6), 107-13 (8) | 106/2.3 e 2.4 são disjuntas (dispatch versus leque); 107-13 depois da 2.4 |

## Divergências entre o inventário do dono e os planos reais
1. **Mesma onda com dependência** (violava a regra "cada task só depende de onda anterior"): 0.5 (precisa 0.2 e 0.8), 0.7 (0.2) e 1.3 (1.1) na onda 2; 106/2.1 (1.6), 3.2 (0.6) e 107-12 (107-02) na onda 3; 106/2.3 e 2.4 (2.2) e 109-05 (109-04) na onda 4; 106/1.5 (1.4), 2.6 (1.5) e 2.7 (2.6) na onda 5; 107-06 (107-05) e a cadeia da 105 onda 2 na onda 6; 106/3.4 (3.1) e 107-10 (107-03) na onda 7. Resolvido deslocando as tasks para a onda seguinte à da dependência.
2. **Acima de 7 tasks:** as ondas 3 (7 antes dos 109), 5 (8), 6 (10) e 7 (9) do inventário passavam do teto; refeitas com até 7.
3. **107-13 na onda 3 não cabe:** depende da 106/2.4 (onda 7 na ordem refeita; onda 4 no inventário); entra na onda 8.
4. **Pré-requisitos ausentes do inventário:** 106/0.1 (de 3.1 e 3.3), 106/1.7 (de 2.6), 106/3.3 (de 3.4). Incluídas e marcadas `*`. **Fora** do inventário e **não incluídas** (sem task dependente na tabela): 106/0.4, 106/2.5, 106/3.7; adiadas pela própria 106: 0.3, 3.5. Decisão do dono.
5. **109-01** não depende de nada, mas foi para a onda 2 (depois da 105 onda 1) para travar os goldens **já com** os `doctor` e `status` novos da 105; a 105 onda 2 regenera esses goldens. **109-03 e 109-06** dependem só da 109-02 (onda 3); só a **109-04** precisa de 106/0.5 e 0.8.
6. **105 onda 1** já está construída (PRs #123 a #129): na onda 1 é só merge. **105-09** (2.7) só liga os passos depois de 106/2.6 (onda 9, não 5). A **105 onda 3** depende da 104 já mesclada (#117) e da 106/1.1 e 1.7 (105-07, PR #129).
7. **Onda 0:** "#121 + cadeia 105" lido como #121 e os PRs #123 a #129; **#154 não mescla antes do Gate 1** da 109.
8. **Marco da 5.1.0:** conteúdo e ADR, ver os dois alertas acima.
