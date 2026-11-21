# Discovery — 109 Módulos externos do forge-sdd (o quadro do GitHub Project vira o primeiro módulo)

## Pedido do dono (20/11/2026; decisões já tomadas, não reabertas)
A inteligência de operar o quadro do GitHub Project (espelho do ciclo, iniciativa 108) sai do core e vira um **módulo instalável**, em **repositório separado** (ciclo de lançamento e licença próprios). Mecanismo: **subcomando externo com contrato fixo** (`forge project ...` executa um programa separado, por exemplo `forge-project`). A construção da 108 está **pausada** até esta discovery ser aprovada; nenhum Builder rodou.

**Refinamento do dono (mesmo dia):** o core mantém a **esteira inteira, intacta** (spec, validador, OK do dono, Builder, validador da entrega, aprovação, main), com fila, estado por feature, orquestrador e laço da 107 (WIP humano, início automático do Builder, freios), PR por feature e atestado da 106, Constituição e gates. **Nada disso pode depender do módulo.** Vai para o módulo só a "inteligência de como operar o quadro": espelho, colunas e mapeamento estado-coluna, etiquetas e prefixo de título, via rápida, issues comerciais, sincronia nas duas vias, rotina dos agentes no quadro, `forge project ...`, checagem do `doctor` dele, documentação, BPMN e Kanban. O dono considera isso expertise de operação (quase mentoria) e não quer entregá-la no core.

## Por que o ID é 109
Conferido em 20/11/2026: `ls sdd/discovery` na main termina em 104; `git branch -r` mostra 105 a 108 e nenhuma 109. O 109 estava livre.

## Clarify
Sem pergunta antes de escrever: o pedido traz decisão, fronteira e requisito inegociável. As lacunas reais viraram `## Decisões para o Gate 1`.

## Descrição de negócio (obrigatória em toda spec nova)
O forge-sdd passa a aceitar **módulos opcionais**: pacotes separados que o dono instala, autoriza uma vez por versão e que reagem ao que a esteira faz (por exemplo, andar um cartão num quadro do GitHub). Quem não instala nada continua com o forge-sdd exatamente como hoje, sem custo e sem aviso novo. Quem instala ganha a operação do quadro sem que o forge-sdd precise carregá-la. O ganho comercial: a expertise de operação vira um produto com ciclo, licença e distribuição próprios, e o core fica pequeno, neutro e seguro.

## Problema
A 108 pôs no core tudo do quadro (23 a 24 tasks, mais 16 tasks externas como pré-requisito). Isso engorda o core, amarra o lançamento do forge-sdd ao do quadro, expõe de graça uma expertise de operação e mistura duas coisas: a esteira (que todo projeto usa) e o jeito de operar um quadro (que só alguns usam). Falta um mecanismo seguro de extensão: hoje o forge só executa binários que ele mesmo escolhe.

## Para quem
Nathan (dono: autoriza módulos, decide licença e distribuição) e os agentes (que continuam na esteira sem saber do módulo). Depois, qualquer autor de módulo que cumpra o contrato.

## Evidência no código (Reuse Check, DUP-02)
Lido em `origin/main` (8bd3a0a) neste worktree; **nada foi executado**.

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Descoberta de executável | `internal/pipeline/runner.go:133`, `preflight.go:35-45`, `launcher.go:49`, `queue/pr.go:14` | `exec.LookPath` direto ou injetado (`PreflightEnv.LookPath`, `prLookPath`) para agentes, `tmux`, `gh` | **Estender o padrão** (LookPath injetável) em pacote novo | `ARQ-02`: `internal/modulo` não importa `internal/pipeline`; o teste de custo zero conta chamadas ao `LookPath` injetado |
| Execução de caminho externo | `runner.go:142-146` | `CommandContext`, `cmd.Env = os.Environ()+FORGE_WORKER`, `WaitDelay` 10 s | **Criar** com ambiente por lista permitida | herdar todo o ambiente vazaria segredos a um binário de terceiros (`SEG-04`, `SEG-05`) |
| Allowlist de comandos | `internal/pipeline/pr.go:30-34`, `queue/pr.go:19-24`, `config/forge.go:181` (`AllowCommands`: `git`, `gh`, `forge-sdd`) | Só `git` e `gh` fixos no código (`SEG-03`) | **Estender com ADR** | executar um módulo quebra de propósito essa lista fixa; exige ADR (`FLX-05`) e é o ponto mais perigoso do desenho |
| Validação de caminho | `launcher.go` (`safeBinRe`), `scaffold/hooks.go:181` (`checkSafeTarget`), `pipeline/worktree.go:56` e `leque.go:115` (`EvalSymlinks`) | Recusa de caminho e de link simbólico para fora do projeto | **Reusar** a técnica | `SEG-02`, `DUP-03` (extrair para função comum só na 3ª repetição, `DUP-05`) |
| `--version` dos agentes | `grep '"--version"'` em `internal/` e `cmd/` | **Nenhuma sondagem**; só `config.BinaryVersion()` (`config/version.go:14`) | **Criar** o aperto de mão do contrato (`contrato`) | não há o que reusar |
| Verificação de binário | `grep sha256` | `crypto/sha256` só para IDs (`layout.go:104`, `notify.go:163`, `stations.go:89`); **nenhuma verificação de integridade** | **Criar** com `crypto/sha256` (stdlib, `ARQ-04`) | |
| Hooks | `internal/scaffold/hooks.go:21-35` | Hook `UserPromptSubmit` e `SessionStart` do Claude Code | **Não reusar** | hooks são do agente; módulo é do forge |
| `internal/vcs` | não existe na main (`ls internal`) | proposto só em 106/0.2 | **Não depender** | Go proíbe importar `internal/` de outro repositório; o módulo nunca importa o core |
| Eventos e notificações | `internal/notify/notify.go` (`Notification`), `internal/bus/job.go`, `sdd/spec/concepts.md:16-18` | Barramento de jobs entre sessões e inbox; **nenhum contrato de eventos de domínio** | **Criar** `internal/modulo` (dono único do conceito, `DUP-01`) | `bus` é fila entre sessões, `notify` é aviso ao humano; não são o canal do módulo |
| `doctor` | `internal/doctor/harness.go:66-80` (`Diagnose`, lista de `Checagem`), `ExitFalha = 3` (`:35`) | Lista fechada de 7 checagens | **Estender** com a 8ª só quando o campo da Constituição pedir módulo | `doctor` sem módulo permanece idêntico |
| Registro de comandos | `cmd/forge-sdd/main.go:401-423` (`AddCommand`) | Lista estática | **Estender** (comando oculto + comandos dinâmicos dos módulos autorizados) | |
| Orçamento da Constituição | `internal/gate/gate.go:789,823` | `MaxConstitutionSize = 9216` | **Reusar** | ver "Constituição" |
| Sessão de diagrama e entrega | `internal/gate/criteria.go:15-16,64-130` | P-25 confere forma estática (bash -n, `# negativo:`, nomes de `-run` em `-list`) | **Reusar** nos critérios | |

## Contrato de módulos (v1, mínimo e estável)

**Descoberta (decisão recomendada):** o core **não procura no PATH nem na pasta do projeto**. Um módulo só roda se (1) o nome está em `Módulos permitidos` na Constituição da **branch base**, (2) está instalado na pasta do **usuário** (`<os.UserConfigDir>/forge-sdd/modulos/<nome>/<versão>/`), (3) o `sha256` do binário e do manifesto bate com o registro de confiança e (4) o contrato é compatível. Repositório clonado não planta binário: o agente que escreve no projeto não consegue autorizar um módulo sozinho (duas travas: a lista versionada e o registro fora do repositório).

**Manifesto** (`modulo.json`, arquivo ao lado do binário, lido **sem executar nada**):
```json
{"contrato": {"min": 1, "max": 1}, "nome": "project", "versao": "0.1.0",
 "binario": "forge-project", "comandos": [{"nome": "project", "resumo": "Opera o quadro do ciclo"}],
 "eventos": ["feature.estado_mudou", "spec.aprovada", "pr.aberto", "pr.mesclado"],
 "capacidades": ["doctor", "pendentes"],
 "permissoes": {"rede": true, "executa": ["gh"], "escreve": [".forge/modulos/project/"]}}
```

**Invocação.** `forge project <args...>` repassa os argumentos ao binário com terminal herdado (flags não interpretadas pelo cobra). Eventos: `forge-project evento` com JSON no stdin e resposta no stdout. `forge-project doctor` devolve checagens. Todos recebem: diretório de trabalho = raiz do projeto; ambiente **somente** `PATH`, `HOME`, `LANG`, `TERM`, `GH_TOKEN` quando declarado, mais `FORGE_CONTRATO=1`, `FORGE_DIR`, `FORGE_MODULO_DIR=.forge/modulos/<nome>` e `FORGE_MODULO_CONFIG=sdd/modulos/<nome>.json` (arquivo versionado e **opaco ao core**).

**Envelope do evento (v1):**
```json
{"contrato": 1, "evento": "pr.aberto", "id": "uuid", "em": "2026-11-20T12:00:00Z",
 "projeto": {"raiz": "/abs", "vcs": "github", "base": "main"},
 "dados": {"iniciativa": "108-x", "feature": "feat-108-01", "estado": "review_pending", "pr": {"numero": 12, "url": "https://...", "tipo": "feature", "draft": true, "head_sha": "abc"}}}
```
Resposta: `{"contrato": 1, "ok": true, "avisos": ["texto"]}`. Campos desconhecidos são ignorados nos dois sentidos; só se **remove ou renomeia** com novo `contrato` (mudança só aditiva dentro do v1).

**Eventos v1 (lista fechada, semântica, sem caminho de arquivo):** `feature.estado_mudou` (`de`, `para` da lista fechada de `internal/queue`: `queued`, `building`, `review_pending`, `approved`, `changes_requested`, `blocked`), `spec.aprovada` (Gate 1), `pr.aberto`, `pr.pronto`, `pr.mesclado`, `veredito.gravado` (`ponto`, `conforme`, `head_sha`), `iniciativa.arquivada`. **Semântico de propósito:** o payload não diz onde o estado mora (`queue.json` hoje, `state/<sid>.json` depois da 106/0.7); o módulo v1 funciona com a main de hoje e continua funcionando depois da 106.

**Entrega:** melhor esforço, no máximo uma vez, **nunca bloqueia** a fila nem a esteira (já era a regra 108/2.3); tempo limite 10 s (manifesto pode pedir até 30 s); o módulo deve ser **idempotente** (RN-4 da 108). Instantâneo de leitura: `forge modulo contexto --json` (somente leitura) entrega iniciativas, estados por feature, PRs, `wip_humano`, `modo` e `max_parallel_tasks`, para o módulo reconciliar sem ler arquivo do core.

**Códigos de saída do módulo:** 0 ok; 2 uso inválido; 3 contrato incompatível; 4 recusado por configuração; 5 falha externa (por exemplo `gh`). Para o core, **qualquer código diferente de 0 vira aviso**, nunca falha da esteira. Saída: stdout limitado a 64 KB, stderr saneado e limitado (`SEG-08`).

**Versão e recusa.** `contrato.min` e `contrato.max` no manifesto; o core fala os contratos que conhece (v1). Fora da faixa: recusa com "o módulo X exige o contrato N; este forge fala o contrato M; atualize o forge ou o módulo", **só visível no `doctor`** (ver requisito de byte-a-byte). O módulo declara também o **contrato mínimo do core** (`contrato.min`) e a task de bootstrap o registra.

**Sem canal de volta em v1.** O módulo **não aprova, não move estado do core e não escreve em arquivo do core**. Arrastar um cartão no quadro é, no módulo, uma divergência a mostrar (e o menu de aprovação continua no core). Um canal de "pedidos" fica para depois (ver adiado).

**`forge --help` e `forge doctor`.** `--help` lista uma seção "Módulos" **só** para módulos autorizados e ok. O `doctor` ganha a checagem `modulos` **só** quando `Módulos permitidos` não for `nenhum`: lista nome, versão, contrato, autorização e hash; módulo ausente ou não autorizado: aviso; hash mudou ou contrato incompatível: aviso com recusa (nunca `falha`, para o código de saída 3 não mudar). A capacidade `doctor` do módulo acrescenta as checagens dele, com prefixo `modulo:<nome>`.

## Modelo de confiança (mínimo viável)
1. **Origem:** só a pasta do usuário acima, preenchida por `forge modulo instalar` (a partir de um arquivo local ou de um endereço `https` de release, **sempre com `--sha256` esperado**, e de um registro fechado de módulos oficiais com 1 entrada, decisão 4). Nunca PATH, nunca a pasta do projeto, nunca caminho relativo (`exec.ErrDot` mantido).
2. **Verificação:** `sha256` do binário e do manifesto; instalação grava o par no registro `confianca.json` (pasta do usuário, **fora do git**).
3. **Autorização:** o dono, por **menu**, uma vez **por versão**, vendo nome, versão, hash e as `permissoes` declaradas. Sem resposta ou "não": não executa. Modo sem humano da 106 **não** autoriza módulo (decisão de segurança: instalar é sempre humano).
4. **Binário mudou:** hash diferente do registrado = recusa e nova autorização; versão nova = nova autorização.
5. **O que o módulo pode ler e escrever:** o core **não consegue** confinar um processo com a stdlib (`ARQ-04`, nenhum sandbox portável). Fica explícito: o controle é (a) autorização do dono com `permissoes` declaradas e mostradas, (b) diretório de trabalho = raiz, (c) ambiente por lista, (d) tempo e tamanho limitados, (e) o core nunca passa segredo que o módulo não declarou. **Rede:** o primeiro módulo precisa de `gh`, então "sem rede por padrão" quebraria o caso real; a rede é uma **permissão declarada e exibida**, não imposta. Sandbox do SO, assinatura (`crypto/ed25519` é stdlib, o difícil é distribuir a chave) e permissão de escrita imposta ficam **adiados**.
6. **Honestidade:** nada disto impede um módulo autorizado de agir mal; protege contra binário trocado, contra repositório hostil e contra agente que se autoriza.

## A fronteira (task a task)
Critério do dono: **esteira = core; jeito de operar o quadro = módulo.** Ver `plan.md` para o fecho. A 108 tem **24 tasks** no `plan.md` (0.1, 0.2, 1.1, 1.2, 1.3, 1.4, 1.5, 1.7, 2.1 a 2.7, 3.1 a 3.7, 4.1, 4.2); a tabela do fecho dela lista 23 e põe a 0.2 como proposta de onda 2. Classifico as 24.

| 108 | O quê | Destino | Observação |
|---|---|---|---|
| 0.1 | Ler config e custo zero | **divide** | core: leitor de `Módulos permitidos` e custo zero (task C do 109); módulo: lê o próprio `sdd/modulos/project.json` |
| 0.2 | ADR da 108 e `concepts.md` | **divide** | ADR de fonte da verdade e do "arrastar" vai ao repositório do módulo; o core ganha a ADR do 109 |
| 1.1 | `Etapa` e `EtapaDoEstado` | módulo | consome o evento e o instantâneo |
| 1.2 | `Quadro` e adaptador `gh project` leitura | módulo | |
| 1.3 | `forge project status` | módulo | |
| 1.4 | config do Project e mapa | módulo | o `card` do PM MCP em `config/forge.go:185` **não muda** |
| 1.5 | teste de injeção | módulo | |
| 1.7 | `ClassificaEtiquetas`, `PrefixoDoTitulo` | módulo | |
| 2.1, 2.2 | `Mover`, `Vincular`, `forge project move` | módulo | |
| 2.3 | gatilhos depois de `queue next/complete/review` e Gate 1 | **divide** | core emite os eventos; módulo reage |
| 2.4 | sem VCS e sem Project: nada é chamado | **divide** | core: teste de custo zero sem módulo; módulo: o próprio |
| 2.5 | `criar-project.sh` | módulo | tarefa do dono |
| 2.6 | etiquetas e prefixo no PR e na issue | módulo | reage a `pr.aberto` |
| 2.7 | via rápida | módulo | regra do quadro |
| 3.1 | polling e divergência | módulo | |
| 3.2 | arrastar como pedido | módulo | sem canal de volta em v1: só mostra a divergência |
| 3.3 | checagem do `doctor` | módulo | via capacidade `doctor`; o core só carrega |
| 3.4 | `FonteDePendencias` | **divide** | **interface no core (107)**; o módulo implementa pela capacidade `pendentes` |
| 3.5 | reference `project.md` e linhas nos comandos | módulo | **o core não ganha linha nos comandos** (mudaria templates e goldens e violaria o byte-a-byte); o módulo traz `forge project rotina` |
| 3.6 | prova ponta a ponta | módulo | tarefa do dono |
| 3.7 | início automático (`observar`) | **divide** | **regra de início automático = laço da 107 (core)**, hoje sem task na 107; o módulo só soma a fonte de pendências |
| 4.1 | corpo do PR com palavra de fechamento | módulo | reage a `pr.aberto`; o core não altera o corpo que a 106 gera |
| 4.2 | guia e wiki do quadro | módulo | ver migração de documentos |

**Estado por feature e `internal/vcs` ficam no core**, por três argumentos: (1) Go não permite importar `internal/` de outro repositório; (2) o esquema do `state/<sid>.json` **nem está definido na 106** (a própria 108 anota isso), logo acoplar um módulo de ciclo próprio a ele quebraria a cada ajuste; (3) o módulo precisa só dos **fatos** (estado, PR, SHA), que o evento e o instantâneo entregam. O módulo faz as próprias chamadas `gh project` e, para PR, só lê o que veio no evento.

**Documentos que migram (a confirmar contra a branch #153, que não li):** `docs/quadro-do-ciclo.md`, as imagens `docs/img/quadro-*`, a seção "Quadro de acompanhamento (opcional)" de `docs/wiki.html`, `criar-project.sh`, as pastas `issues/` da 108 e os diagramas Kanban e BPMN (`forge-108-*`). **Texto mínimo que o core mantém:** uma seção "Módulos externos (opcionais)" na wiki e na doc do CLI: o que é um módulo, o campo `Módulos permitidos`, `forge modulo`, a autorização, sem nada da operação do quadro.

## Constituição (sem editar; medido sobre a proposta da 106/0.8)
Base: a proposta de `constitution.proposta.md` na branch da 106 tem **9.182 bytes** (main: 9.189; teto `MaxConstitutionSize` 9.216, `gate.go:789`).

| Variante | Bytes | Folga |
|---|---|---|
| 106/0.8 como está (com `Quadro de acompanhamento`) | 9.182 | 34 |
| **Recomendada:** troca a linha `Quadro de acompanhamento` por `Módulos permitidos`, e `project` sai de `ARQ-06` em favor de `modulo` | **9.173** | **43** |
| Troca a linha, `project` fica no `ARQ-06` | 9.174 | 42 |
| Troca a linha e `project` sai sem citar `modulo` | 9.163 | 53 |
| Acrescenta `Módulos permitidos` sem tirar `Quadro` | 9.215 | 1 (inviável na prática) |

**Decisão recomendada:** o campo `Quadro de acompanhamento` **migra** para a configuração do módulo (`sdd/modulos/project.json`: `quadro`, número do Project, nome do campo); `Início automático do Builder` e `WIP humano` **ficam** na Constituição (laço da 107, core). Linha nova: `| Módulos permitidos | nenhum |` (valor `nenhum` ou lista separada por vírgula). Campo **ausente** vale `nenhum`, sem aviso, e **não entra no template de `init`** na v1, para o scaffold e os goldens não mudarem. Tudo entra **no PR único da 106/0.8**, não em um segundo PR de Constituição (ADR-0022, `GOV-01`).

**Nota para a rule `sdd-modos-de-projeto.md` (proposta; a seção 5 da proposta da 106 é reescrita):**
> ### 5. Módulos permitidos e Início automático (`ARQ-06`, `SEG-03`)
> - `Módulos permitidos`: `nenhum` (padrão; campo ausente vale o mesmo) ou lista de nomes. Só roda módulo listado aqui, instalado na pasta do usuário, com hash conferido e autorizado pelo dono para aquela versão. `nenhum` não lê disco do usuário nem chama processo.
> - A configuração de cada módulo fica em `sdd/modulos/<nome>.json`, opaca ao core.
> - `Início automático do Builder`: `desligado` (padrão) ou `local`; é do laço da 107.

## Regras de negócio
- **RN-1** O core sem módulo se comporta **byte a byte como hoje**: mesma saída, mesmos códigos, nenhum aviso novo, custo zero. Inegociável e testável.
- **RN-2** Módulo ausente, não autorizado, quebrado, lento ou incompatível **nunca derruba o core**: é ignorado; a mensagem só aparece no `doctor` (e só quando há módulo permitido).
- **RN-3** O módulo é **aditivo**: o core emite eventos e oferece o contrato; nenhuma função da esteira lê o módulo.
- **RN-4** Executar módulo exige quatro coisas juntas: nome na Constituição da base, binário na pasta do usuário, hash igual ao autorizado, contrato compatível.
- **RN-5** Quem autoriza é o dono, por menu, uma vez por versão; modo sem humano não autoriza; agente não autoriza.
- **RN-6** O contrato v1 é só aditivo; remover ou renomear exige novo número de contrato.
- **RN-7** O módulo não aprova, não escreve estado do core e não recebe segredo não declarado.
- **RN-8** Eventos são semânticos (sem caminho de arquivo) e entregues no máximo uma vez, sem bloquear; o módulo é idempotente.
- **RN-9** Testes (do core e do módulo) rodam sem rede (`TST-03`, `TST-04`): binário falso de teste (`TestMain` que reexecuta o próprio teste), `t.TempDir()`, `LookPath` e `CommandRunner` injetados, `gh` falso.
- **RN-10** O repositório do módulo nasce **dentro da esteira SDD**, com o forge-sdd estável mais recente instalado e configurado (task de bootstrap).
- **RN-11** Visibilidade e licença do módulo são **decisão do dono**; esta discovery só descreve as opções.

## Repositório e distribuição do módulo (só descrito; nada foi criado)
- **Nome:** `forge-sdd-project` (binário `forge-project`). Convenção `forge-sdd-<nome>`/`forge-<nome>`.
- **Layout mínimo:** `modulo.json`, `cmd/forge-project/`, `internal/{contrato,etapa,quadro,etiqueta,evento}/`, `docs/`, `sdd/` (esteira SDD própria), `LICENSE`, `README.md`, `.github/workflows/release.yml`, `testdata/gh-falso/`.
- **Instalar (um comando):** `forge modulo instalar project` (registro fechado de módulos oficiais com uma entrada: fica como decisão, porque coloca um nome de módulo no core); alternativa neutra: `forge modulo instalar --de <arquivo|https> --sha256 <hash>`. Os dois mostram o menu de autorização.
- **Atualizar:** nova versão = nova pasta, nova autorização; a anterior continua até `remover`. **Remover:** `forge modulo remover <nome> [versão]` apaga a pasta e o registro; tirar o nome da Constituição já desliga sem apagar.
- **Teste sem rede:** `gh` falso no PATH do teste (executável de teste que imprime JSON de `testdata/`), servidor `httptest` de loopback para o instalador (rede local do próprio teste, não externa), relógio e diretórios injetados.

## Outros candidatos a módulo (só lista; sem desenho)
| Candidato | Hoje | Por que sai do core | Usa o contrato? |
|---|---|---|---|
| Quadro do GitHub Project | 108 | inteligência de operação | sim (eventos, `doctor`, `pendentes`) |
| Azure DevOps (adaptador de VCS e Boards) | 106/0.3 e 108 | adiado, "não verificado", depende de conta real | eventos + `doctor` |
| Medição de economia de CI | 106/3.3 (`ci-politica.py`) | leitura de cobrança, token próprio | `doctor`, comando próprio |
| Orquestrador da 107 | 107 | **não sai**: é o laço da esteira (decisão do dono) | só `FonteDePendencias` |
| Notificações externas (e-mail, webhook, toast) | `internal/notify` | integração opcional | eventos |
| Telemetria/relatórios de custo | `internal/monitor`, `report` | opcional | eventos |

Serve a mais de um módulo porque o contrato tem só: manifesto, evento, `doctor`, `pendentes`, instantâneo. Nada é específico de quadro.

## Escopo
**Dentro (core):** `internal/modulo` (manifesto, contrato v1, registro de confiança, execução, eventos, instantâneo), comando oculto `forge modulo`, comandos dinâmicos dos módulos autorizados, checagem do `doctor`, texto mínimo da doc, ADR, suíte de byte-identidade, e as alterações do PR único da Constituição.
**Dentro (módulo, outro repositório):** tudo da tabela marcada "módulo", bootstrap e docs.
**Fora:** criar o repositório do módulo; sandbox, assinatura e canal de volta; Azure; mexer na esteira; decidir visibilidade e licença.

## Riscos
| Risco | Mitigação |
|---|---|
| Binário de terceiros com acesso ao projeto e a `gh` | autorização por versão com hash e permissões visíveis; ambiente por lista; honestidade de que não é sandbox |
| `forge --help` ou `doctor` mudarem sem módulo | comando `modulo` oculto, seção só com módulo ok, checagem só com campo ≠ `nenhum`; goldens |
| Contrato congelar cedo demais | só aditivo, faixa `min`/`max`, um só módulo real para validar (o do quadro) |
| Módulo lento travar a esteira | tempo limite, melhor esforço, nunca bloqueia |
| Constituição estourar (43 bytes de folga) | uma só rodada de edição, no PR único; teste de P-24 |
| 106/107 mudam o esquema do estado | payload semântico; só o emissor do core muda |
| Dependência circular de merge (PR único da Constituição) | o campo entra na 106/0.8; a task de leitura do campo depende da 106/0.5 |
| Dois repositórios com ciclos diferentes | contrato versionado e `contrato.min` no módulo |

## Decisões para o Gate 1 (recomendação de cada uma; a decisão é do dono)
1. **Onde o módulo vive e como o core o acha:** pasta do usuário + lista na Constituição + hash autorizado (recomendado) **versus** PATH (o brief original). PATH aceita qualquer `forge-project` plantado; recomendo a pasta.
2. **Comando visível ou oculto:** `forge modulo` oculto (recomendado, preserva `--help` byte a byte) **versus** visível (mais descobrível, muda a ajuda de todos).
3. **Visibilidade e licença do repositório do módulo** (do dono; não decido). Opções:
   - **Privado, instalação por convite** (token ou colaborador): máxima proteção da expertise; custo: atrito, não há comunidade, o instalador precisa de autenticação. *Recomendação: começar aqui.*
   - **Público com licença própria** (fonte disponível, uso comercial restrito; por exemplo BUSL/PolyForm): transparência e confiança; custo: a expertise fica legível, difícil de fazer valer a licença.
   - **Distribuição paga ou por mentoria** (binário assinado entregue ao cliente): monetiza; custo: exige canal de entrega, suporte e licenciamento; pede a assinatura que fica adiada.
   - Observação: o contrato é aberto em qualquer cenário (o core é público); só o módulo é protegido.
4. **Registro de módulos oficiais no core** (uma entrada) **versus** só `--de <arquivo|https> --sha256`. Recomendo só `--de` na v1 (core 100% neutro); o comando de um passo vira alias documentado no README do módulo.
5. **Início automático (3.7) e regra de WIP** ficam no laço da 107: criar a task nova na 107 (número do dono). Recomendo: sim, task `107-13` proposta, fora do fecho.
6. **Prioridade 0 da 108 e o fecho:** com a 108 fora do core, as 16 tasks externas continuam úteis por mérito próprio, mas **deixam de ser pré-requisito de um lançamento do quadro**. Recomendo reabrir só a ordem (ver `plan.md`), sem reduzir escopo.
7. **Sem canal de volta (pedidos) em v1** (recomendado) **versus** já aceitar `pedidos` validados pelo core.
