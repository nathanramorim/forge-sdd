# Discovery — 113 Criação automática de issue/card na concepção (`/discovery`, `/nova-feature`, `/novo-fix`)

## Pedido do dono (21/11/2026)
> "crie uma novo discovery que permita todo /discovery, /novo-fix, /nova-feature já crie a issue no gestor configurado, no caso aqui usamos o project do github, oque não configurou ainda ou não quer deixa sem criar mesmo, mas considere a feature que fizemos agora a apouco recente sobre essa gestao do project que é opcional e na verdade é até uma extensao do forge, continuar funcionando, acredito que a escala disso aqui será definir o mcp de conexão com o painel de companhamento, pra github usar gh CLI, veja o que ja tem e não gere impacto, isso aqui deve ser uma feature incremental."

## Por que o ID é 113
Conferido em `origin/main` (commit `332071b`): `sdd/discovery/` possui iniciativas até `112-lp-estavel-beta-e-estilo-forge`. O próximo ID livre é `113`.

## Clarify (`sdd/memory/clarify.md`)
Sem pergunta forçada antes de produzir os artefatos: o pedido define claramente o comportamento esperado nos três comandos de concepção (`/discovery`, `/nova-feature`, `/novo-fix`), a condição de opcionalidade ("o que não configurou ainda ou não quer deixa sem criar mesmo"), a tecnologia por provedor (`gh` CLI para GitHub, MCP para outros painéis de acompanhamento) e a restrição inegociável de não gerar impacto na extensão opcional do Project (`108` / `109`). As escolhas finas de governança ficam consolidadas em `## Decisões para o Gate 1` com recomendação explícita.

## Descrição de negócio
Ao abrir uma concepção com `/discovery`, `/nova-feature` ou `/novo-fix`, o `forge-sdd` passa a criar automaticamente a issue ou o card correspondente no gestor/painel de acompanhamento configurado pelo projeto — e somente quando ele estiver configurado. Projetos que não configuraram gestor ou preferem trabalhar apenas em arquivos continuam com **custo zero** (nenhuma issue criada, nenhuma chamada externa, nenhum erro ou bloqueio). Quando o gestor é o GitHub (com ou sem o módulo opcional `forge-sdd-project`), a criação usa o `gh` CLI, grava o vínculo local (`card-evidence.json`) de forma idempotente e alimenta o módulo externo do Project sem duplicar cartões; quando o gestor é outro painel (`clickup`, `jira`, `azure-devops`), a criação escala pelo contrato padronizado de conexão via MCP (`sdd/memory/mcps.md` + `pm_mcp`), gravando o mesmo vínculo para que toda a esteira posterior (Builder, PR e Reviewer) já receba a referência pronta.

- **Problema:** Hoje, `internal/config/forge.go` (`PMConfig` / `pm_mcp`) e `internal/pipeline/dispatch.go` (`CardEvidence` / `card-evidence.json`) só sabem cobrar ou repassar um card que alguém já tenha criado manualmente antes do despacho (`require_existing_card`). Ao mesmo tempo, a jornada do quadro do ciclo (Iniciativas 108 e 109) prevê que `/discovery`, `/nova-feature` e `/novo-fix` coloquem a issue na coluna **Spec** assim que a concepção nasce, mas os três comandos de concepção (`.agents/skills/sdd-concepcao/references/{discovery,nova-feature,novo-fix}.md`) ainda não têm o passo nem o mecanismo incremental para criar a issue no GitHub via `gh` CLI ou o card via MCP nos demais gestores.
- **Para quem:** O dono do projeto (que quer gestão à vista desde o nascimento da spec, sem criar issue à mão antes de rodar `/discovery`, `/nova-feature` ou `/novo-fix`), equipes que usam painéis externos via MCP (ClickUp, Jira, Azure Boards) e equipes que não usam gestor algum (que não podem sofrer qualquer atrito ou impacto).
- **Valor:** Elimina o trabalho manual de abrir issue/card antes da concepção, fecha o elo entre o início da spec e o módulo externo `forge-sdd-project` (coluna **Spec**), padroniza a escala para outros gestores via MCP e mantém o core 100% neutro e silencioso quando nenhum gestor está ativo.

## Evidência do que já existe no código (`Reuse Check`, `DUP-02`)
Verificado em `origin/main` (`332071b`):
1. **`internal/config/forge.go` (`PMConfig`):** já existe a estrutura `pm_mcp` em `.forge/config/forge.json` com `provider` (`auto`, `clickup`, `github-projects`, `azure-devops`, `jira`), `supported`, `require_existing_card` (`false` por padrão), `status_on_implement`, `status_on_pr_approved` e `evidence_max_age_minutes`.
2. **`internal/pipeline/dispatch.go` (`CardEvidence`):** já existe o contrato `card-evidence.json` na pasta da iniciativa (`sdd/discovery/<id>-<nome>/card-evidence.json`) com os campos `provider`, `id`, `url`, `status`, `feature` e `checked_at` (RFC3339), além de `bus.Card{Provider, ID, URL}` no job do Builder/Reviewer. Hoje nenhum código grava esse arquivo automaticamente na concepção.
3. **`internal/vcs/vcs.go` e `github.go` (entregue na 106-02):** já existe o pacote `internal/vcs` com `Adaptador`, `LeVCS(dir)` (lendo `VCS / Work Item System` da Constituição) e `RunGitOrGh` (execução segura de `git` e `gh` com argv explícito, sem shell, `SEG-03`). Hoje `Adaptador` opera apenas PRs (`CriarPRDraft`, `ProcurarPRAberto`, `CorrigirBase`, `MarcarPronto`, `Mesclar`), ainda sem método para criar ou localizar issue.
4. **`sdd/memory/mcps.md` e `cmd/forge-sdd/mcp.go`:** `mcps.md` registra os servidores MCP disponíveis no projeto e seu status (`ativo`/`indisponível`), enquanto `cmd/forge-sdd/mcp.go` expõe o servidor somente leitura do próprio forge (`forge_doctor`, ADR-0017).
5. **Iniciativas 108 e 109 (`sdd/discovery/109-modulos-externos/` e repositório `forge-sdd-project`):** o quadro do GitHub Project virou o módulo externo opcional `project` (`forge-project`), autorizado pelo dono via `Módulos permitidos` na Constituição (106/0.8) e configurado em `sdd/modulos/project.json`. O core se comunica com módulos externos por eventos semânticos v1 (`internal/modulo`).

## Como funciona (Jornada incremental nos 3 cenários)

### Cenário A — Gestor não configurado ou desligado (Padrão: Custo Zero)
Quando o projeto não configurou gestor de issues/cards:
- `Módulos permitidos` não contém `project` (ou não há `sdd/modulos/project.json`) **e** `pm_mcp.provider` não aponta para um gestor ativo (por exemplo, não há `.forge/config/forge.json`, ou `pm_mcp` está desligado/`nenhum`, ou `provider: "auto"` sem MCP de gestão `ativo` em `sdd/memory/mcps.md`).
- **Comportamento:** `/discovery`, `/nova-feature` e `/novo-fix` seguem **exatamente como hoje**. Nenhuma chamada a `gh`, nenhuma chamada a MCP, nenhum arquivo `card-evidence.json` fabricado, nenhum aviso de erro. Custo zero absoluto (`RN-1` da 109 e `RN-16` da 108).

### Cenário B — GitHub configurado (`gh` CLI + convivência com o módulo `forge-sdd-project`)
Quando o gestor configurado para acompanhamento é o GitHub — seja porque o módulo `project` (109) está ativo (`sdd/modulos/project.json` / `Módulos permitidos`), seja porque o projeto optou pela criação automática de issue no GitHub (`VCS / Work Item System: github` com gestor `github-projects`/`github` ativo):
1. **Via `gh` CLI (`internal/vcs`):** um subcomando determinístico do core (ou método do adaptador `internal/vcs`) cria ou reaproveita a issue usando `gh issue list` / `gh issue create` via `RunGitOrGh` (argv explícito, sem shell, `SEG-03`).
2. **Idempotência (sem duplicatas):** antes de criar, verifica se já existe `card-evidence.json` para o alvo (iniciativa em `/discovery`, feature em `/nova-feature`, fix em `/novo-fix`) ou issue aberta correspondente no GitHub. Se já existir, apenas reaproveita o número e a URL.
3. **Descrição comercial e higienizada (`RN-8` da 108):** o corpo da issue criada automaticamente usa a `## Descrição de negócio` (sem jargão interno, sem caminhos locais absolutos, sem dados sensíveis, `SEG-05`, `SEG-08`) e aponta o link relativo para a pasta/arquivo da spec (`sdd/discovery/<id>-<nome>/...`).
4. **Gravação do vínculo (`card-evidence.json`):** grava atomicamente (` WriteFileSafe` / `WriteJSONAtomic`, `SEG-02`) o vínculo (`provider: "github-projects"`, `id: "<numero>"`, `url: "<html_url>"`, `status: "Spec"`, `feature: "<alvo>"`, `checked_at: "<RFC3339>"`). Assim, `internal/pipeline/dispatch.go` e a task `4.1` / `P-06` do módulo (que coloca `Closes #N` no PR que mira a `main`) já encontram o vínculo pronto sem heurística frágil.
5. **Convivência limpa com o módulo externo `forge-sdd-project` (108/109):**
   - Se o módulo externo `project` (109) estiver instalado e autorizado, o core emite o evento semântico aditivo `spec.aberta` (ou aciona o contrato v1 de `internal/modulo`) passando nos `dados` a `iniciativa`, o `alvo` (`discovery`, `feat-<id>-<nn>` ou `fix-<id>-<nn>`) e a `issue` (`numero`, `url`, `titulo`).
   - O módulo `forge-project` recebe a issue já criada (ou cria/vincula se ainda não estiver no Project), aplica as etiquetas (`ClassificaEtiquetas`, incluindo `spec`), aplica o prefixo no título (`PrefixoDoTitulo`) e posiciona o cartão na coluna **Spec** do GitHub Project.
   - Se o módulo `project` **não** estiver instalado, mas a criação de issue no GitHub estiver habilitada no projeto, a issue é criada normalmente no repositório via `gh issue create` e registrada no `card-evidence.json`, sem tentar falar com o binário `forge-project`.

### Cenário C — Outros painéis de acompanhamento via MCP (`clickup`, `jira`, `azure-devops`)
Quando o projeto usa outro gestor configurado em `pm_mcp.provider` (ou detectado via `auto`) e registrado como `ativo` em `sdd/memory/mcps.md`:
1. **Conexão via MCP:** o `forge` não embute SDK nem cliente HTTP de ClickUp, Jira ou Azure Boards no binário Go (`ARQ-04`). A escala para outros gestores ocorre pelo **MCP do painel de acompanhamento** declarado em `sdd/memory/mcps.md` e configurado no ambiente do agente (`.mcp.json` / `.gemini/mcp.json`).
2. **Contrato de criação e registro:** durante `/discovery`, `/nova-feature` ou `/novo-fix`, a estação de Concepção consulta o status do gestor (`forge card status` / `pm_mcp` + `sdd/memory/mcps.md`). Estando o provedor MCP ativo:
   - Se `card-evidence.json` do alvo ainda não existir, o agente invoca a ferramenta MCP de criação de card/task do provedor configurado (com título e descrição comercial saneados, `SEG-08`).
   - Em seguida, registra o retorno (`id`, `url`, `provider`, `status`) por meio do comando determinístico do core (`forge card registrar --provider <p> --id <id> --url <url> --feature <alvo>`), que valida os campos na borda (`SEG-01`) e grava `card-evidence.json` de forma segura (`SEG-02`).
3. **Sem MCP ativo ou não configurado:** se o provedor MCP estiver `indisponível` ou não configurado, não cria nada, emite no máximo um aviso informativo e segue a concepção sem bloquear.

## Jornada por comando de concepção

| Comando | Alvo criado | Quando gestor desligado / não configurado | Quando GitHub (`gh` CLI) ativo | Quando gestor MCP (`clickup`, `jira`, `azure-devops`) ativo |
|---|---|---|---|---|
| `/discovery` | Issue-mãe / card da iniciativa (`<id>-<nome>`) | Nada é criado (custo zero) | `gh issue create` com a Descrição de negócio + link da pasta; grava `card-evidence.json`; se módulo `project` ativo, entra na coluna **Spec** | Cria card via ferramenta MCP do provedor ativo e grava `card-evidence.json` via CLI |
| `/nova-feature` | Issue / card da feature (`feat-<id>-<nn>`) ou da iniciativa enxuta | Nada é criado (custo zero) | `gh issue create` para a feature; grava evidência vinculada à feature; se módulo `project` ativo, entra na coluna **Spec** | Cria card da feature via MCP e grava evidência vinculada à feature |
| `/novo-fix` | Issue / card do fix (`fix-<id>-<nn>`) correlacionado à feature de origem | Nada é criado (custo zero) | `gh issue create` para o fix citando a feature correlacionada; grava evidência do fix; se módulo `project` ativo, entra na coluna **Spec** | Cria card do fix via MCP e grava evidência vinculada ao fix |

## Regras de negócio
- **RN-1 (Opcionalidade e custo zero quando não configurado):** Se nenhum gestor de acompanhamento estiver configurado ou se estiver explícito como desligado/`nenhum`, `/discovery`, `/nova-feature` e `/novo-fix` não criam issue/card, não chamam `gh`, não exigem MCP, não gravam `card-evidence.json` e não mudam a saída padrão (`RN-1` da 109).
- **RN-2 (Zero impacto nas Iniciativas 106, 108 e 109):** O contrato de módulos externos da 109 (`internal/modulo`), a política de PR em draft da 106 (`internal/vcs`) e o desenho do módulo `forge-sdd-project` (108/109) permanecem intactos. Qualquer adição ao contrato de eventos v1 da 109 (como o evento `spec.aberta`) é **estritamente aditiva** (`RN-6` da 109: mudanças aditivas dentro do contrato v1 são compatíveis e ignoradas por quem não as usa).
- **RN-3 (GitHub via `gh` CLI; demais gestores via MCP):** Para o GitHub, a criação e consulta de issue usam exclusivamente `gh` CLI com argumentos explícitos (`RunGitOrGh`, `SEG-03`). Para `clickup`, `jira`, `azure-devops` e futuros painéis, a criação ocorre via servidor MCP registrado em `sdd/memory/mcps.md`, mantendo o binário Go livre de dependências externas (`ARQ-04`).
- **RN-4 (Idempotência por iniciativa, feature ou fix):** Rodar novamente o passo de criação para a mesma iniciativa, feature ou fix **nunca duplica** a issue ou o card. O vínculo gravado em `card-evidence.json` (ou a busca por título base no GitHub) é verificado antes; se já existir, o ID e a URL existentes são preservados.
- **RN-5 (Melhor esforço: falha externa nunca trava a concepção):** Se o `gh` estiver sem rede/autenticação ou se o servidor MCP falhar no momento da concepção, o problema vira **aviso claro** na saída (e no `forge doctor`), mas **nunca impede** a criação e o commit dos artefatos da spec (`discovery.md`, `criteria.md`, `plan.md`, `status.md`, `feat-*.md`, `fix-*.md`).
- **RN-6 (Higienização e descrição comercial, `RN-8` da 108, `SEG-05`, `SEG-08`):** O título e o corpo enviados ao `gh issue create` ou ao MCP passam por sanitização (sem caracteres de controle, sequências ANSI ou tokens/segredos mascarados por `gate.MaskSecrets`) e usam a descrição comercial da spec + caminho relativo no repositório, nunca caminhos absolutos da máquina local.
- **RN-7 (`--dry-run` sem efeito colateral, `ARQ-05`):** O comando determinístico de criação/vinculação de issue suporta `--dry-run`, exibindo o que seria criado no `gh` ou registrado no `card-evidence.json` sem executar chamada de escrita nem tocar o disco.
- **RN-8 (Compatibilidade com `dispatch.go` e `card-evidence.json`):** O vínculo gravado na concepção segue o esquema estrito de `CardEvidence` (`internal/pipeline/dispatch.go:42-49`), permitindo que `forge dispatch` preencha `job.Card` mesmo quando `require_existing_card` for `false` (quando o arquivo existir e for válido) e renovando `checked_at` quando consultado.

## Escopo
### Dentro do escopo
- Pacote/extensão no core para resolver o gestor de acompanhamento ativo (custo zero quando desligado; GitHub via `gh` CLI; provedor MCP via `pm_mcp` + `sdd/memory/mcps.md`).
- Extensão incremental de `internal/vcs` para criar e procurar issue no adaptador `github` (`CriarIssue`, `ProcurarIssueAberta` via `gh issue create` e `gh issue list --json`, `SEG-03`) e no-op/erro semântico no adaptador `nenhum`.
- Gravação e leitura idempotente do vínculo em `card-evidence.json` (por iniciativa e por feature/fix) com validação na borda (`SEG-01`, `SEG-02`) e suporte a `--dry-run` (`ARQ-05`).
- Emissão aditiva do evento `spec.aberta` para o módulo externo `project` (quando a infraestrutura de `internal/modulo` da 109 estiver presente e autorizada), permitindo que o `forge-project` posicione automaticamente a issue criada na coluna **Spec**.
- Reference curta e enxuta em `.agents/skills/sdd-concepcao/references/` (respeitando o orçamento P-24) acionada por `/discovery`, `/nova-feature` e `/novo-fix` para criar/registrar a issue ou card quando o gestor estiver configurado.
- Atualização de `internal/pipeline/dispatch.go` para carregar `card-evidence.json` em `job.Card` sempre que o arquivo já tiver sido criado pela concepção (mesmo com `require_existing_card: false`).

### Fora do escopo
- Reimplementar no core a inteligência das 7 colunas, etiquetas ou sincronia bidirecional do GitHub Project (isso pertence ao módulo externo `forge-sdd-project`, conforme decisão da 109).
- Embutir clientes HTTP/REST de ClickUp, Jira ou Azure DevOps dentro do binário Go do `forge-sdd` (a integração com esses painéis é exclusivamente via MCP).
- Alterar ou atrasar as ondas já aprovadas das Iniciativas 105, 106, 107 ou 109.
- Tornar obrigatória a criação de issue/card para quem não configurou gestor.

## Riscos e Mitigações
| Risco | Impacto | Mitigação |
|---|---|---|
| Criar issue sem querer em projeto que não usa gestor (ex.: porque `VCS / Work Item System` é `github` por padrão) | Alto (polui o repositório do usuário) | Só criar issue no GitHub quando o acompanhamento estiver explicitamente habilitado (módulo `project` ativo, `sdd/modulos/project.json` presente, ou `pm_mcp.provider` configurado como `github-projects` / `create_on_conception: true`). Com o padrão de fábrica sem configuração, **nada é criado**. |
| Duplicar issue entre o comando de concepção e o módulo `forge-sdd-project` (109) | Médio | O vínculo (`numero` e `url`) é gravado em `card-evidence.json` e enviado no payload do evento `spec.aberta`; além disso, a busca por título base (`ProcurarIssueAberta`) garante idempotência (`RN-4`). |
| Estourar o orçamento de contexto (`P-24`) em `SKILL.md` ou nos adaptadores | Alto (quebra o gate `P-24`) | `sdd-concepcao/SKILL.md` tem 2.045/2.048 bytes e não será editado (ou receberá no máximo ajuste neutro); o passo entra apenas nas references `discovery.md`, `nova-feature.md`, `novo-fix.md` e em reference auxiliar dedicada. |
| Falha de rede, falta de autenticação no `gh` ou MCP fora do ar travar `/discovery` | Médio | Regra de melhor esforço (`RN-5`): falha externa na criação da issue devolve aviso e código tratável, nunca bloqueia a escrita nem o commit da spec. |
| Vazamento de caminho local ou segredo no corpo da issue em repositório público | Alto (`SEG-05`, `RN-8` da 108) | O montador do corpo da issue extrai apenas o título e a `## Descrição de negócio`, aplica `gate.MaskSecrets` e sanitização de controle (`SEG-08`), e usa apenas caminhos relativos ao repositório. |

## Decisões para o Gate 1 (com recomendação)

1. **Qual interruptor ativa a criação automática de issue no GitHub?**
   - *Contexto:* Em `sdd/memory/constitution.md`, `VCS / Work Item System` já vem como `github` por padrão para permitir PRs. Porém, o dono pediu: *"o que não configurou ainda ou não quer deixa sem criar mesmo"*.
   - **Opção A (Recomendada):** Ativar automaticamente quando **(1)** o módulo `project` estiver ativo/configurado (`Módulos permitidos` contém `project` ou existe `sdd/modulos/project.json`) **OU (2)** `.forge/config/forge.json` (ou configuração do projeto) declarar explicitamente o gestor ativo (`pm_mcp.provider: "github-projects"` ou campo booleano opcional `pm_mcp.create_on_conception: true`). Se nenhum dos dois estiver configurado, fica desligado (custo zero).
   - **Opção B:** Criar issue sempre que `VCS / Work Item System` for `github`, a menos que o usuário desligue. *(Não recomendada: criaria issues em repositórios que só querem usar PRs e não configuraram gestão de issues/Project).*

2. **Granularidade das issues em `/discovery`, `/nova-feature` e `/novo-fix`:**
   - **Opção A (Recomendada, alinhada à 108/109):**
     - `/discovery` cria **1 issue-mãe** da iniciativa (entra na coluna `Spec`).
     - `/nova-feature` cria **1 issue da feature** (ou da iniciativa enxuta).
     - `/novo-fix` cria **1 issue do fix** (referenciando a feature correlacionada).
     - Quando uma iniciativa completa passa pelo split de features na construção (106/108), cada feature recebe seu vínculo/issue conforme o plano.
   - **Opção B:** Criar issue apenas no `/discovery` e nunca criar issue própria para `/nova-feature` ou `/novo-fix` dentro de uma iniciativa existente. *(Não recomendada: contraria o pedido literal "permita todo /discovery, /novo-fix, /nova-feature já crie a issue no gestor configurado").*

3. **Como escalar para outros gestores (`clickup`, `jira`, `azure-devops`) sem inflar o binário Go (`ARQ-04`):**
   - **Opção A (Recomendada — Híbrida CLI + MCP):** O binário Go (`forge`) resolve a configuração, executa diretamente o caso GitHub via `gh` CLI (`internal/vcs`) e oferece o subcomando determinístico (`forge card ...`) para validar e gravar `card-evidence.json`. Para gestores baseados em MCP (`clickup`, `jira`, `azure-devops` ativos em `sdd/memory/mcps.md`), a reference da skill `sdd-concepcao` instrui o agente a chamar a ferramenta MCP de criação do provedor ativo e passar o `id` e `url` retornados para `forge card registrar`, mantendo um único formato de evidência.
   - **Opção B:** Criar um módulo externo separado (padrão 109) para cada gestor (`forge-sdd-clickup`, `forge-sdd-jira`). *(Pode coexistir no futuro, mas para a concepção onde o agente já tem o MCP conectado, a Opção A entrega valor imediato com zero binário extra).*
