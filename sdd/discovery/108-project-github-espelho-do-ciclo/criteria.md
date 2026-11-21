# Critérios — Iniciativa 108 (GitHub Project como espelho do ciclo)

## Restrições (Constitution)
`ARQ-01` (`cmd/` só adapta flags do `forge project`; o mapa e a regra ficam em `internal/pipeline`, a chamada ao `gh` em `internal/vcs`), `ARQ-02` (sem ciclo: `vcs` não importa `pipeline`; `pipeline` define a interface que consome), `ARQ-04` (só stdlib em runtime; `gh` é binário externo opcional já listado em `allow_commands`), `ARQ-05` (`--dry-run` e a leitura do quadro não têm efeito em disco nem chamam escrita no GitHub), `ARQ-06` (o comando entra sob `forge project`, subcomando novo documentado em `--help`; ADR se o dono preferir raiz), `ARQ-11` (modelo do módulo declarado), `COD-01`, `COD-03`, `DUP-01` (um dono do mapa de estados), `DUP-03` (nenhum pacote paralelo ao `vcs` da 106), `FLX-01` (a Discovery só escreve spec; criar o Project é ato do dono), `FLX-06`, `GOV-01` (nenhuma regra nova sem ADR), `SEG-01` (validar entrada), `SEG-03` (só `gh` em `allow_commands`), `SEG-04` (token nunca em arquivo nem saída), `SEG-07` (sem dependência nova), `SEG-08` (saneamento de texto externo) e `TST-03` (teste negativo por regra de recusa).

## Modelo do módulo
**Enxuto** (`ARQ-11`): sem pacote novo além do que a 106 já prevê. (a) O **mapa estado do arquivo para etapa** e a **lista fechada de 7 etapas** vivem em `internal/pipeline` (já dona de `queue`, `status` e do despacho; `DUP-01`). (b) O **adaptador do quadro** (ler etapa, mover cartão, vincular PR) é uma interface mínima `Quadro` declarada no consumidor (`internal/pipeline`) e implementada em `internal/vcs` pelo adaptador `github` (`gh project`), junto da interface `Adaptador` da 106 task 0.2; `azure-devops` ganha o equivalente Azure Boards **a verificar**; `nenhum` devolve "sem quadro". (c) `forge project` (`cmd/forge-sdd/project.go`) só adapta flags. Fronteira: enquanto `internal/vcs` não existir em `main` (a 106 ainda não foi implementada), a task 1.x cria **só o necessário** do pacote (interface `Quadro` e o adaptador `github`) sem duplicar o que a 106 vai definir, e **sequencia com a 106 task 0.2** (a verificar quem entra primeiro). Justificativa: um pacote `internal/project` novo duplicaria o conceito de adaptador de VCS (`DUP-03`).

## Decisões do projeto aplicadas
- **ADR-0004:** o Go decide o que é determinístico (mapa de estados, validação da lista fechada, idempotência); o modelo não decide etapa.
- **ADR-0012:** a aprovação de um job atualiza a fila; o cartão é consequência da fila, nunca o contrário sem confirmação.
- **ADR-0013:** Gate 1 humano; arrastar o cartão não substitui o menu do Gate 1.
- **ADR-0015 e ADR-0016:** o chat é o coordenador; o estado em arquivo tem lock por iniciativa; o espelho lê o estado, não o disputa.
- **ADR-0017:** saída somente leitura e saneada; o `forge project` lê o quadro com saída saneada (texto de issue nunca vai ao agente).
- **ADR-0018:** gravação segura de arquivos do usuário (`status.md` só pelo caminho seguro).
- **ADR-0003:** skill até 2.048 bytes e agente até 420 bytes; a rotina nova vai numa `references/` de `sdd-operacao` e os comandos só a apontam.
- **Fluxos:** `sdd/spec/flows.md` (pipeline e modo chat); o espelho entra depois de cada transição de estado, sem mudar a transição.
- **Decisão de VCS (106):** campo `VCS / Work Item System` (`sdd/memory/constitution.md:25`); `github`, `azure-devops`, `nenhum`. O quadro é adaptador do mesmo campo.
- **106 (coordenação, sem editá-la):** sinais `spec-para-revisao` e `pronto-para-revisao`, veredito com `head_sha`, atestado `desligado|local|vcs`, split pelo Builder. A 108 lê esses sinais; não os redefine.
- **107 (coordenação, sem editá-la):** o orquestrador lê a etapa e conta os cartões em Validation para o WIP humano; entrada de sessão par é dado.
- **105 (coordenação, sem editá-la):** `forge status` e menus; a coluna do cartão entra no `status` só depois que a 105/2.6 sair.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Ligação com gerenciador de projeto | `internal/config/forge.go:91-98,183-190,337-346` (`PMConfig`, `pm_mcp`) | Provedores `clickup`, `github-projects`, `azure-devops`, `jira`; `RequireExistingCard`; `StatusOnImplement` "In Progress", `StatusOnPRApproved` "Ready for Review" | **Reusar** `provider` para escolher o quadro; **estender** com o nome do campo e as 7 etapas; **não** criar outra config | `DUP-06`: um dono de `forge.json`. Os dois status atuais não casam com as 7 etapas (a mapear na task 1.4) |
| Uso de card no despacho e no veredito | `internal/pipeline/dispatch.go:200-262,431-482`, `review.go:107` | Só evidência (`card-evidence.json`) e uma frase `CardInstruction`; nenhuma chamada de API, nenhuma leitura de coluna | **Reusar** a evidência (provider, id, url) como vínculo cartão-iniciativa; **criar** a leitura e o movimento reais | O código hoje delega ao agente via MCP, que não está registrado em `mcps.md` |
| Estados da fila | `internal/queue/queue.go:16-23`, `Complete` `:464`, `Review` `:488` | `queued`, `building`, `review_pending`, `approved`, `changes_requested`, `blocked`; sem `in_progress` nem `completed` | **Reusar**; o mapa lê esses estados, não cria estado novo | `DUP-01` |
| Jobs do barramento | `internal/bus/job.go:33-37` | `feature_build`, `pr_rework`, `pr_review`, `spec_review` | **Reusar** como evento: `pr_rework` volta o cartão a Building | `DUP-01` |
| Adaptador de VCS | 106 (`plan.md` task 0.2, `internal/vcs`) | Só desenhado; **não existe em `main`** | **Depende de**; declarar `Quadro` no consumidor e implementar junto do adaptador `github` | `DUP-03`: evitar pacote paralelo |
| Atestado, veredito com SHA, sinais no `status.md` | 106 (`discovery.md` RN-4, RN-6, RN-13, RN-18; `plan.md` 1.1, 1.6, 3.2) | Só desenhados | **Depende de**: o arrastar do humano só vale com o atestado conforme o modo | Evita um segundo canal de aprovação |
| Orquestrador e WIP humano | 107 (`discovery.md`, decisões do Gate 1) | Só desenhados | **Depende de** para contar Validation; a 108 só expõe o número | `DUP-01` |
| Menus e `forge status` | 105 (2.2, 2.6) | Só desenhados/PRs | **Sequenciar com**: a coluna em `status` só depois | Evita conflito no mesmo arquivo |
| Comando `gh project` | `gh project --help` (somente leitura) | `create`, `edit`, `field-create`, `field-list`, `item-add`, `item-edit`, `item-list`, `link` existem; **não** há `field-edit` (as opções do `Status` mudam por GraphQL `updateProjectV2Field`) | **Reusar** o `gh` (já em `allow_commands`); o campo da etapa é o `Status` (opções trocadas pela API GraphQL, verificado em 20/11/2026); não existe campo `Etapa` | `SEG-07`: nenhuma dependência nova |
| Saneamento de texto externo | `internal/notify/notify.go:122` (`Sanitize`), `internal/gate` (`MaskSecrets`) | Saneadores prontos | **Reusar** para título e corpo de issue lidos do quadro | `SEG-08` |
| Reference de rotina por agente | `.agents/skills/sdd-operacao/references/` (`archive`, `doctor`, `pipeline`, `status`...) | Sem `project.md` | **Criar** `project.md` (ponto único); `SKILL.md` de `sdd-concepcao` não muda (2045 de 2048 bytes) | `DUP-01`, P-24 |
| Conceito "quadro do ciclo" | `sdd/spec/concepts.md` | Sem linha | **Adicionar** linha (dono: `internal/pipeline`, adaptador em `internal/vcs`) | `DUP-01` |

## Dependências novas
Nenhuma (`SEG-07`). O `gh` (CLI do GitHub) já é usado e está em `allow_commands` (`internal/config/forge.go:181`); JSON do `gh` lido com a stdlib (`encoding/json`, `DisallowUnknownFields`). Para `azure-devops`, `az boards` seria binário externo opcional, **a verificar**.

## Rules e IDs da Constituição Aplicáveis
`.agents/rules/sdd-arquitetura.md` (ARQ-01, 02, 04, 05, 06, 11), `sdd-codigo-e-testes.md` (COD-01, 03, TST-01 a 03), `sdd-fluxo-e-portao.md` (FLX-01, 06, GOV-01), `sdd-seguranca.md` (SEG-01, 03, 04, 07, 08), `sdd-reuso-e-duplicacao.md` (DUP-01, 03, 06). `forge context <task>` resolve o recorte de cada task.

## Tabela de mapeamento (estado do arquivo, coluna, evento)
| Coluna | Estado no arquivo | Evento que move o cartão | Quem move |
|---|---|---|---|
| Backlog | Ideia registrada sem pasta de spec (item de backlog da iniciativa ou issue) | Alguém registra a ideia | humano ou pm-architect |
| Spec | Pasta `sdd/discovery/<id>-<nome>/` existe; `status.md` sem `spec-para-revisao`; `queue.json` ausente | `/discovery`, `/nova-feature` ou `/novo-fix` abrem a spec | pm-architect (IA) |
| Spec-validation | `status.md` com `spec-para-revisao` (106 RN-4); PR da spec em draft | Spec entregue e script local verde | pm-architect (IA) |
| Todo | Gate 1 aprovado (marcador `Gate 1: aprovado`, 106 task 2.2); `queue.json` com itens `queued` ou, pela 106, split ainda por fazer (a cargo do Builder) | Menu do Gate 1 ([Seguir]) | forge, depois do menu (HUMANO) |
| Building | Item `building` ou `changes_requested`; PR da feature em draft | Builder pega a task (`queue next`); reviewer devolve (`pr_rework`) | forge, a partir da fila (IA) |
| Validation | Item `review_pending` ou `approved`; `pronto-para-revisao` (106); PR pronto para a conferência | `queue complete`, ou veredito conforme | forge, a partir da fila (AUTOMÁTICO) |
| Done | PR da feature mesclado | Merge feito pelo dono ou pelo adaptador após [Continuar] | forge (HUMANO decide) |
| (campo `Bloqueado`) | Item `blocked` | `blocked_reason` preenchido | forge; cartão fica na etapa de origem |
Recusa: qualquer valor fora dos 7 nomes exatos é descartado e vira divergência. Sem `status.md` ou `queue.json` legíveis o cartão não se move (falha fechada).

## Sinais da 106, 107 e 105 (como se relacionam)
- **106:** `spec-para-revisao` leva a Spec-validation; `Gate 1: aprovado` leva a Todo; `pronto-para-revisao` leva a Validation; veredito com `head_sha` diferente do HEAD volta o cartão a Building (commit novo invalida). O atestado (`desligado`, `local`, `vcs`) define se arrastar Validation para Done exige prova externa.
- **107:** o orquestrador lê a etapa, nunca o texto do cartão; o WIP humano conta cartões em Validation e em Spec-validation (a decisão é do dono).
- **105:** `forge status` ganha a etapa depois da 2.6; a 2.7 (descrição de negócio) é a fonte do texto comercial das issues.

## Espelho em duas vias (fonte da verdade e conflito)
| Dado | Fonte da verdade | Como o outro lado acompanha |
|---|---|---|
| Estado da spec, da fila e do PR | Arquivo (`status.md`, `queue.json`) e o PR | O forge move o cartão depois de cada transição (idempotente) |
| Etapa vista no quadro | O campo `Status` do Project (lido em `.status` no `item-list`) | O forge lê por polling (`gh project item-list --format json`) a cada comando e em `doctor`; o orquestrador lê o campo |
| Aprovação do Gate 1 e [Continuar] | O menu do forge e o atestado (106) | Arrastar o cartão **só gera o menu**; nunca grava aprovação |
| Conflito (quadro diz X, arquivo diz Y) | Arquivo vence para estado; o humano vence só pelo menu | `forge doctor` lista a divergência com o ator e a hora; nada é corrigido sem confirmação |
Percepção do arrastar: **polling** pelo CLI (sem webhook nem Action; o CI está bloqueado por cobrança e webhooks de Projects v2 de usuário são **a verificar**). Se `item-list` não expuser o ator do movimento (**a verificar**), o arrastar é só um pedido e o forge confirma por menu.

## Diagramas (C4, Mermaid; fontes em `diagramas/*.mmd`, PNGs renderizados com `mmdc`)
Contexto, contêineres e sequência abaixo; o ciclo das 7 colunas está em `diagramas/forge-108-ciclo.png` (diagram-studio, tema escuro, assinatura do projeto).

### Contexto
```mermaid
flowchart LR
  dono(["Dono do projeto<br/>decide e arrasta cartão"])
  ag(["Agentes<br/>pm-architect, builder, reviewer"])
  forge["forge CLI<br/>mapeia estado em etapa"]
  arq[("Arquivos no repositório<br/>spec, queue.json, status.md")]
  proj[/"GitHub Project<br/>7 etapas, gestão à vista"/]
  gh["GitHub<br/>issues e PRs"]
  dono -->|"arrasta o cartão"| proj
  dono -->|"decide nos menus"| forge
  ag -->|"escrevem o estado"| arq
  ag -->|"chamam"| forge
  forge -->|"lê e grava"| arq
  forge -->|"lê e move cartão via gh"| proj
  proj -->|"vincula"| gh
```

### Contêineres
```mermaid
flowchart TB
  cli["cmd/forge-sdd<br/>forge project (só adapta flags)"]
  mapa["internal/pipeline<br/>estado de arquivo para etapa"]
  vcs["internal/vcs<br/>Adaptador: github, azure-devops, nenhum"]
  fila["internal/queue<br/>queue.json e status.md"]
  doc["internal/doctor<br/>divergência arquivo x quadro"]
  ghcli["gh project<br/>item-list, item-edit, field-list"]
  quadro[/"GitHub Project<br/>campo Status, 7 opções"/]
  cli --> mapa
  mapa --> fila
  cli --> vcs
  vcs --> ghcli
  ghcli --> quadro
  doc --> mapa
  doc --> vcs
```

### Sequência
```mermaid
sequenceDiagram
  participant H as Dono
  participant P as Project (Status)
  participant F as forge project
  participant A as Arquivos (status.md, queue.json)
  Note over F,A: arquivo vira quadro
  F->>A: lê estado e calcula a etapa
  F->>P: move o cartão se a etapa difere
  Note over H,A: quadro vira arquivo
  H->>P: arrasta o cartão para outra etapa
  F->>P: polling lê Status e ator do evento
  F->>F: valida a etapa na lista fechada de 7
  F->>H: menu: confirmar a etapa nova?
  H->>F: confirma (atestado da 106)
  F->>A: grava o estado e registra o ator
```

## Critérios de Aceitação Executáveis
Os nomes de teste abaixo são o **contrato**: as features os criam. Antes da implementação, `forge gate --criterios` (estático) passa, mas `--executar-criterios` falha de propósito nas linhas `go test -list ... | grep -F`, que só passam quando o teste existe.
```bash
set -o pipefail
go build ./... && go vet ./... && go test ./...
# mapa: as 7 etapas cobrem todos os estados da fila e os sinais da 106
go test -list 'TestEtapaDoEstadoCobreAsSeteColunas' ./internal/pipeline/ | grep -F 'TestEtapaDoEstadoCobreAsSeteColunas'
go test -run 'TestEtapaDoEstadoCobreAsSeteColunas' ./internal/pipeline/
# negativo: valor fora da lista fechada de 7 (inclusive texto livre e instrução embutida) é recusado
go test -list 'TestEtapaRecusaValorForaDaListaFechada' ./internal/pipeline/ | grep -F 'TestEtapaRecusaValorForaDaListaFechada'
go test -run 'TestEtapaRecusaValorForaDaListaFechada' ./internal/pipeline/
# negativo: item blocked mantém a etapa de origem e marca Bloqueado, sem coluna nova
go test -list 'TestEtapaBloqueadoMantemEtapaDeOrigem' ./internal/pipeline/ | grep -F 'TestEtapaBloqueadoMantemEtapaDeOrigem'
go test -run 'TestEtapaBloqueadoMantemEtapaDeOrigem' ./internal/pipeline/
# negativo: título, corpo e comentário de issue nunca viram ordem nem saem sem saneamento
go test -list 'TestQuadroTextoDeIssueEhDadoENaoOrdem' ./internal/vcs/ | grep -F 'TestQuadroTextoDeIssueEhDadoENaoOrdem'
go test -run 'TestQuadroTextoDeIssueEhDadoENaoOrdem' ./internal/vcs/
# leitura somente leitura: lê só o campo Status e nunca chama escrita no GitHub
go test -list 'TestQuadroLeituraNaoChamaEscrita' ./internal/vcs/ | grep -F 'TestQuadroLeituraNaoChamaEscrita'
go test -run 'TestQuadroLeituraNaoChamaEscrita' ./internal/vcs/
# negativo: mover o mesmo cartão duas vezes não repete a chamada (idempotente)
go test -list 'TestMoverCartaoEhIdempotente' ./internal/vcs/ | grep -F 'TestMoverCartaoEhIdempotente'
go test -run 'TestMoverCartaoEhIdempotente' ./internal/vcs/
# negativo: --dry-run não chama escrita nem grava arquivo
go test -list 'TestMoverCartaoDryRunNaoEscreve' ./cmd/forge-sdd/ | grep -F 'TestMoverCartaoDryRunNaoEscreve'
go test -run 'TestMoverCartaoDryRunNaoEscreve' ./cmd/forge-sdd/
# negativo: sem VCS (nenhum) não há quadro e o gh nunca é chamado
go test -list 'TestQuadroSemVCSNaoChamaGh' ./internal/vcs/ | grep -F 'TestQuadroSemVCSNaoChamaGh'
go test -run 'TestQuadroSemVCSNaoChamaGh' ./internal/vcs/
# negativo: arrastar Spec-validation para Todo não grava o Gate 1 sem o menu de confirmação
go test -list 'TestArrastarNaoAprovaGate1SemMenu' ./internal/pipeline/ | grep -F 'TestArrastarNaoAprovaGate1SemMenu'
go test -run 'TestArrastarNaoAprovaGate1SemMenu' ./internal/pipeline/
# negativo: arrastar Validation para Done sem PR mesclado ou sem atestado conforme é recusado
go test -list 'TestArrastarNaoFechaDoneSemPRMescladoEAtestado' ./internal/pipeline/ | grep -F 'TestArrastarNaoFechaDoneSemPRMescladoEAtestado'
go test -run 'TestArrastarNaoFechaDoneSemPRMescladoEAtestado' ./internal/pipeline/
# negativo: movimento feito por ator desconhecido ou sem permissão de escrita nunca vale como aprovação
go test -list 'TestAtorDesconhecidoNaoAprova' ./internal/pipeline/ | grep -F 'TestAtorDesconhecidoNaoAprova'
go test -run 'TestAtorDesconhecidoNaoAprova' ./internal/pipeline/
# doctor: divergência arquivo versus quadro é listada e nunca corrigida em silêncio
go test -list 'TestDoctorDetectaDivergenciaQuadroArquivo' ./internal/doctor/ | grep -F 'TestDoctorDetectaDivergenciaQuadroArquivo'
go test -run 'TestDoctorDetectaDivergenciaQuadroArquivo' ./internal/doctor/
# negativo: a saída do forge project nunca imprime token nem valor de variável de ambiente
go test -list 'TestProjectSaidaNaoVazaToken' ./cmd/forge-sdd/ | grep -F 'TestProjectSaidaNaoVazaToken'
go test -run 'TestProjectSaidaNaoVazaToken' ./cmd/forge-sdd/
# script de criação (rascunho): sintaxe válida e dry-run sem escrita
bash -n sdd/discovery/108-project-github-espelho-do-ciclo/criar-project.sh
bash sdd/discovery/108-project-github-espelho-do-ciclo/criar-project.sh --dry-run | grep -F 'DRY-RUN'
# negativo: flag desconhecida é recusada e nada é criado
if bash sdd/discovery/108-project-github-espelho-do-ciclo/criar-project.sh --inventada; then exit 1; fi
```
