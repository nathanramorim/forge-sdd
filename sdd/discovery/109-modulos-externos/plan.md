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
| 109-05 | Carregador e execução: `LookPath` injetável, pasta do usuário, ambiente por lista, tempo limite, saída limitada e saneada (`SEG-08`), códigos 0 a 5 mapeados a aviso; **quebra deliberada da lista fixa de comandos** documentada | 109-03, 109-04 | com 109-07 | `SEG-03`, `SEG-05`, `SEG-08`, `ARQ-02`, `TST-03` |
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
| 109-11 | ADR do 109 (contrato, confiança, quebra da lista fixa; número no merge), linha em `sdd/spec/concepts.md` (dono: `internal/modulo`), texto mínimo "Módulos externos (opcionais)" na doc, **sem** linha nos comandos nem nos templates | 109-08, 109-09 | com 109-10 | `FLX-05`, `GOV-01`, `DUP-01` |

**Alteração em task de outra iniciativa (sem PR novo de Constituição):** a **106/0.8** troca a linha `Quadro de acompanhamento` por `| Módulos permitidos | nenhum |`, troca `project` por `modulo` na `ARQ-06` e reescreve a seção 5 da rule `sdd-modos-de-projeto.md` (texto em `discovery.md`); 9.173 de 9.216 bytes (medido; **não editei a Constituição**).

**Task nova na 107 (proposta; número do dono, fora do fecho):** `107-13` regra do início automático do Builder (Todo para Building: spec na principal, WIP humano, teto de 3 tentativas, modo e atestado), antes da 108/3.7. Sem ela o início automático não existe em lugar nenhum depois que a 108 sair do core.

## Novo fecho de dependências (o que muda)
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

**Fecho recomendado de ondas do core (ordem de merge; até 7 por onda):**
| Onda | Tasks |
|---|---|
| 1 | 106/0.8 (alterada), 109-01, 109-02 |
| 2 | 106/0.5, 109-03, 109-06 |
| 3 | 109-04, 109-05, 109-07 |
| 4 | 109-08, 109-09 |
| 5 | 109-10, 109-11 |
As demais tasks da 106 e da 107 seguem a ordem normal delas, sem prazo atrelado ao módulo. Cada onda do fecho tem até 3 tasks.

## Parte B — Repositório do módulo (`forge-sdd-project`; só descrito, nada criado)
Ondas do repositório do módulo, rodadas na esteira SDD dele.

| Onda | # | Task | Depende de |
|---|---|---|---|
| 1 | P-00 | **Bootstrap (abaixo)** | decisão do dono (visibilidade, licença) |
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
6. **Licença e visibilidade:** a que o dono decidir no Gate 1 (decisão 3). O repositório **só** é criado depois disso.
7. **Critérios:** teste sem rede; `forge doctor` do projeto com o módulo falso e verde; versão mínima do contrato registrada no README, no manifesto e na Constituição do módulo.

## O que da 108 vira a spec do módulo e o que se descarta
- **Vira a spec do módulo (reaproveitada):** `discovery.md` e `criteria.md` da 108 (RN-1 a RN-16, mapeamento estado-coluna, etiquetas, prefixo, via rápida, diagramas Kanban e BPMN, issues, `criar-project.sh`), menos o que sobe ao core.
- **Sobe ao core (novo ou 107):** leitura do campo e custo zero (109-04), emissão de eventos (109-06), `FonteDePendencias` (interface, 107), início automático (107-13 proposta).
- **Descarta:** os campos `Quadro de acompanhamento` e a nota dele na Constituição (migram para `sdd/modulos/project.json`); as linhas de aviso nos comandos e templates (3.5, quebrariam o byte a byte); o canal "arrastar como pedido que abre menu do core" (3.2, adiado); a ADR da 108 no core (vai para o repositório do módulo); `ARQ-06` com `project`.
