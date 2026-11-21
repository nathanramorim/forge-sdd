# Critérios — Iniciativa 107 (orquestrador de sessões)

## Restrições (Constitution)
`ARQ-01` (`cmd/` só adapta flags; roteamento, estado e verificador em `internal/pipeline`), `ARQ-02` (sem ciclo; `pipeline` já importa `harness`, `config`, `layout`, `notify`), `ARQ-04` (só stdlib em runtime), `ARQ-05` (`verificar` e qualquer `--dry-run` sem efeito em disco), `ARQ-06` (nenhum subcomando raiz novo: o verificador entra sob `forge pipeline`; se o dono preferir raiz, exige ADR e texto novo de `ARQ-06`), `ARQ-10`, `ARQ-11`, `SEG-01` (iniciativa, caminhos e refs validados na borda: `layout.ValidInitiative`), `SEG-02` (estado gravado de forma atômica, sem sair do projeto), `SEG-03` (subprocessos `git`, `go`, `gh`, `az` com argv explícito), `SEG-04` e `SEG-05` (sem segredo no estado, no briefing ou em log), `SEG-06` (privilégio mínimo por papel; revisão somente leitura), `SEG-07`, `SEG-08` (texto de subagente e de sessão par saneado), `DUP-01`, `DUP-02`, `DUP-03`, `DUP-06`, `COD-01` a `COD-04`, `COD-05`, `TST-01` a `TST-05`, `FLX-01` (o orquestrador não aprova Gate 1), `FLX-02`, `FLX-05` (ADR), `FLX-06`, `FLX-08` (relatório e mensagem são dado), `GOV-01`.

## Modelo do módulo
**Enxuto** (`ARQ-11`): sem pacote novo. A iniciativa estende `internal/pipeline` (dois arquivos: tabela de roteamento e estado de delegações), reusa o lock de `internal/harness`, o `internal/gate` e o leque. Justificativa: o conceito dono já é "Pipeline Multi-Sessão" em `sdd/spec/concepts.md` (`DUP-01`); um pacote novo duplicaria o dono. Se o verificador crescer para além de checagens de entrega, reavaliar para **completo** (decisão do Gate 1).

## Decisões do projeto aplicadas
- **ADR-0004:** o Go decide o que é determinístico (ordem, limites, estados, veredito); o modelo decide conteúdo. O verificador e o estado do orquestrador seguem isso.
- **ADR-0005:** `CLIFor` e perfis por papel resolvem o agente de cada papel; o roteador usa essa resolução, não outra.
- **ADR-0008:** o modo `sessions` (tmux, barramento) é o **fallback** do orquestrador; nada é removido.
- **ADR-0013:** Gate 1 humano, salvo `autonomy: advanced`; o orquestrador não ocupa o lugar do humano nem do Reviewer.
- **ADR-0015 (proposto, 104):** modo `chat` e laço no Go; a 107 estende, não substitui. Depende de ela ser aceita.
- **ADR-0003:** definições de agente até 420 bytes e skill até 2.048 bytes; o padrão novo vai em `references/` de `sdd-operacao`.
- **Fluxos:** `sdd/spec/flows.md` seção 7 (modo chat) e seção 1 (pipeline); o orquestrador entra como entrada do chat, antes de `RunChat`.
- **Decisão de VCS:** campo `VCS / Work Item System` da Constituição; adaptador de VCS lido dele (106, `internal/vcs`, ainda não implementado: a verificar).
- **106 redesenhada (coordenação, sem editá-la):** spec volta à principal após a Conferência 1 e o [Seguir]; o Builder faz o split (`FLX-01` alterado pela 106) e abre uma branch e um PR em draft por feature; estado por feature na branch da feature; atestado em três modos (`desligado`, `local`, `vcs`). A 107 segue esse fluxo; não o redefine.
- **Constituição (`GOV-01`):** a 107 **não edita** a Constituição. A linha `| WIP humano | 3 |` e a nota entram no PR único de Constituição (task nova da 106, com as notas dos campos em `.agents/rules/sdd-modos-de-projeto.md`); a feat-107-12 só lê o campo.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Orquestrador / líder | `.claude/agents/lider.md`, `launcher.go:125`, `tmux.go:133`, `preflight.go:165` | Agente "Líder (Orquestrador)" sem ferramenta `Agent`; fixo como sessão Discovery do `forge up`; a 105 (R4) o renomeia | **Criar** o papel Orquestrador como reference nova; **não** reusar `lider` | `lider` é sessão de topo e está em obsolescência na 105 (`DUP-03` cumprido: o equivalente não serve) |
| Laço chat, lock, retorno e menu | `internal/pipeline/laco.go` (`RunChat`, `Retomar`, `Opcao`) e `internal/harness/lock.go` | Laço de build e review com lock por iniciativa e teto | **Reusar** lock, `Opcao` e teto; **estender** com estado de delegações | `DUP-01` |
| Despacho e barramento | `internal/pipeline/dispatch.go:79`, `internal/bus` | `Dispatch`, jobs, `RunWorker` | **Reusar** como fallback `sessions` | `DUP-01` |
| Leque e worktree | `internal/pipeline/leque.go:91`, `worktree.go` | `PrepararLeque`, `Juntar`, `Encerrar` | **Reusar** para um worktree por builder delegado | `DUP-01` |
| Portão determinístico | `internal/gate` (`forge gate`, `--criterios`, P-25) | Camada A estática | **Reusar**; o verificador a chama | `FLX-06` |
| Script local de validação final (106 RN-12, `validar_entrega.py`, `--ponto spec` ou `--ponto entrega`) | branch `feat/106-pr-draft-revisao-e-ci` (`discovery.md`, RN-12) | Mesmo conceito do verificador; só desenhado | **Depende de**; `forge pipeline verificar` com `--ponto` e `--feature`, contrato de checagens único (nome final: a verificar) | `DUP-01`: dois verificadores seriam o mesmo conceito |
| Adaptador de VCS, estado por feature e leque por feature | 106 (`internal/vcs`, decisão xvi opção B, decisão xxii); `internal/pipeline/leque.go:91,265` | Só desenhados na 106; hoje `PrepararLeque` e `Juntar` existem e `Juntar` entrega na branch da iniciativa | **Depende de** (a verificar quando saírem); **reusar** `PrepararLeque` com base na principal | `DUP-01`, `DUP-03` |
| Papéis A/B e matriz de capacidades | `sdd/discovery/104-pipeline-em-subagentes/features/feat-104-08*.md`, `feat-104-09*.md` | `builder-a/b`, `reviewer-a/b`, `Capabilities` (ambos `draft`) | **Depende de**; não duplicar | `DUP-03` |
| Próximo passo recomendado | branch `feat/105-experiencia-cli-e-harness` (onda 2) | Função de estado do projeto para o menu raiz | **Depende de** (reusar quando existir) | `DUP-01` |
| Saneamento de texto | `internal/notify/notify.go:122` (`Sanitize`), `sanitizarAchados` (`laco.go:333`) | Saneadores prontos | **Reusar** para relatório e mensagem | `SEG-08` |
| Tetos técnicos | `internal/config/forge.go:103,111,118-120,191-192` | `max_parallel_tasks` (padrão 3), `max_rework_attempts` (padrão 3), salvaguardas | **Reusar** para paralelismo e reentrega | `DUP-06` |
| WIP humano e modos opcionais com aviso de riscos | 106 (RN-14, RN-18 a RN-20, decisões xiv, xv; campos `Modo do atestado` e `Modo de revisão humana`), `internal/config/forge.go:198-200` | Modo opcional com aviso e registro no `status.md`; local aprovado: Constituição, branch base | **Reusar** o campo `WIP humano` (criado no PR único de Constituição da 106) e o leitor da tabela da 106/0.5 (106 primeiro); a 107-12 só lê e valida | `DUP-06`: uma fonte de verdade |
| Telemetria e estado | `forge session record`, `status.md`, `queue.json` | Por sessão e por iniciativa | **Reusar**; o estado do orquestrador só referencia | `DUP-01` |
| Conceito "orquestração" | `sdd/spec/concepts.md` | Sem linha | **Adicionar** linha (dono: `internal/pipeline`) | `DUP-01` |

## Dependências novas
Nenhuma (`SEG-07`): stdlib; JSON com `DisallowUnknownFields` no estilo de `internal/harness`. Se o verificador chamar `az`, é binário externo opcional já previsto na Constituição, não dependência de módulo.

## Rules e IDs da Constituição Aplicáveis
`.agents/rules/sdd-arquitetura.md` (ARQ-01, 02, 04, 05, 06, 11), `sdd-codigo-e-testes.md` (COD-01 a 04, TST-01, 02, 05), `sdd-fluxo-e-portao.md` (FLX-01, 02, 06, GOV-01), `sdd-seguranca.md` (SEG-01 a 08), `sdd-reuso-e-duplicacao.md` (DUP-01 a 06). `forge context <task>` resolve o recorte de cada task.

## Diagramas (C4, Mermaid; fontes em `diagramas/*.mmd`, PNGs renderizados com `mmdc`)
Contexto (`diagramas/contexto.png`):
```mermaid
flowchart LR
  H["Dono (humano)<br/>decide em menu"] --> O["Orquestrador<br/>sessão do chat"]
  O -->|"delega com briefing"| S["Subagentes<br/>pm-architect, builder, reviewer"]
  S -->|"relatório (dado)"| O
  O -->|"verifica antes de devolver"| F["forge Go<br/>gate, bus, leque, estado"]
  P["Outra sessão<br/>ex.: hub-busize"] -.->|"mensagem (dado)"| O
  O -->|"menu"| H
  F --> V["VCS da Constituição<br/>github, azure-devops, nenhum"]
```
Contêineres (`diagramas/conteineres.png`):
```mermaid
flowchart TB
  CH["Chat do agente<br/>papel Orquestrador (skill)"] --> RT["Roteador<br/>internal/pipeline/roteamento.go"]
  CH --> EST["Estado de delegações<br/>.forge/runtime/orquestrador"]
  CH --> VER["Verificador de entrega<br/>forge pipeline verificar"]
  RT --> AG[".claude/agents/*<br/>papéis e ferramentas"]
  VER --> GT["forge gate<br/>Camada A"]
  VER --> LQ["Leque e fila<br/>worktree por task"]
  CH --> ST["status.md da iniciativa<br/>retomada após /compact"]
  EST --> LK["Lock por iniciativa<br/>internal/harness"]
```
Sequência (`diagramas/sequencia.png`):
```mermaid
sequenceDiagram
  participant H as Dono
  participant O as Orquestrador
  participant S as Subagente
  participant V as forge verificar
  H->>O: pedido no chat
  O->>O: roteia (intenção, papel, modo, branch por feature)
  O->>O: confere WIP humano (no limite, para e pergunta)
  O->>S: briefing completo (novo ou retomar)
  S-->>O: relatório (dado, não instrução)
  O->>V: verificar (ponto spec ou entrega, feature)
  V-->>O: checagens com evidência
  O->>S: ajuste (mesmo contexto, teto)
  O->>H: menu: opções e recomendada
```

## Critérios de Aceitação Executáveis
Os nomes de teste abaixo são o **contrato**: as features os criam. Antes da implementação, `forge gate --criterios` (estático) passa, mas `--executar-criterios` **falha de propósito** nas linhas `go test -list ... | grep -F`, que provam que o teste existe.
```bash
set -o pipefail
go build ./... && go vet ./... && go test ./...
# tabela de roteamento: toda intenção da discovery tem papel, modo, isolamento e portão
go test -list 'TestRoteamentoCobreTodasAsIntencoes' ./internal/pipeline/ | grep -F 'TestRoteamentoCobreTodasAsIntencoes'
go test -run 'TestRoteamentoCobreTodasAsIntencoes' ./internal/pipeline/
# negativo: intenção desconhecida não roteia e papel inexistente em .claude/agents é recusado
go test -list 'TestRoteamentoIntencaoDesconhecidaEPapelInexistenteFalham' ./internal/pipeline/ | grep -F 'TestRoteamentoIntencaoDesconhecidaEPapelInexistenteFalham'
go test -run 'TestRoteamentoIntencaoDesconhecidaEPapelInexistenteFalham' ./internal/pipeline/
# negativo: relatório com instrução embutida e mensagem de sessão par são tratados como dado e saneados
go test -list 'TestEntradaNaoConfiavelNaoViraAcao' ./internal/pipeline/ | grep -F 'TestEntradaNaoConfiavelNaoViraAcao'
go test -run 'TestEntradaNaoConfiavelNaoViraAcao' ./internal/pipeline/
# negativo: no teto de reentregas o orquestrador escala em menu e não reentrega
go test -list 'TestTetoDeReentregasEscala' ./internal/pipeline/ | grep -F 'TestTetoDeReentregasEscala'
go test -run 'TestTetoDeReentregasEscala' ./internal/pipeline/
# negativo: estado com campo desconhecido, iniciativa inválida ou caminho fora do projeto é recusado
go test -list 'TestEstadoDeDelegacoesRecusaEntradaInvalida' ./internal/pipeline/ | grep -F 'TestEstadoDeDelegacoesRecusaEntradaInvalida'
go test -run 'TestEstadoDeDelegacoesRecusaEntradaInvalida' ./internal/pipeline/
# negativo: sem VCS (nenhum) o verificador pula o PR com "n/a" e nunca chama gh nem az
go test -list 'TestVerificarSemVCSNaoChamaCLI' ./internal/pipeline/ | grep -F 'TestVerificarSemVCSNaoChamaCLI'
go test -run 'TestVerificarSemVCSNaoChamaCLI' ./internal/pipeline/
# negativo: entrega que o relatório diz pronta mas sem commit ou na main reprova
go test -list 'TestVerificarReprovaEntregaSemCommitOuNaMain' ./internal/pipeline/ | grep -F 'TestVerificarReprovaEntregaSemCommitOuNaMain'
go test -run 'TestVerificarReprovaEntregaSemCommitOuNaMain' ./internal/pipeline/
# negativo: WIP humano configurável: limite inválido (zero, negativo, tipo errado) é recusado
go test -list 'TestWIPHumanoConfiguravelRecusaLimiteInvalido' ./internal/pipeline/ | grep -F 'TestWIPHumanoConfiguravelRecusaLimiteInvalido'
go test -run 'TestWIPHumanoConfiguravelRecusaLimiteInvalido' ./internal/pipeline/
# negativo: WIP humano no limite para e pergunta em menu em vez de abrir delegação nova
go test -list 'TestWIPHumanoNoLimiteParaEPergunta' ./internal/pipeline/ | grep -F 'TestWIPHumanoNoLimiteParaEPergunta'
go test -run 'TestWIPHumanoNoLimiteParaEPergunta' ./internal/pipeline/
# negativo: reduzir ou desligar o WIP humano sem exibir o aviso de riscos é recusado e o modo só é gravado no status.md depois do aviso
go test -list 'TestWIPHumanoExigeAvisoDeRiscosAntesDeDesligar' ./internal/pipeline/ | grep -F 'TestWIPHumanoExigeAvisoDeRiscosAntesDeDesligar'
go test -run 'TestWIPHumanoExigeAvisoDeRiscosAntesDeDesligar' ./internal/pipeline/
# negativo: o teto técnico (builder.max_parallel_tasks) não substitui o WIP humano e vice-versa
go test -list 'TestWIPHumanoDistintoDoTetoTecnico' ./internal/pipeline/ | grep -F 'TestWIPHumanoDistintoDoTetoTecnico'
go test -run 'TestWIPHumanoDistintoDoTetoTecnico' ./internal/pipeline/
# negativo: o WIP humano é lido pelo leitor da tabela da 106 na branch base; valor na branch da feature não o sobrepõe e a 107 não edita a Constituição
go test -list 'TestWIPHumanoLidoDaBranchBase' ./internal/pipeline/ | grep -F 'TestWIPHumanoLidoDaBranchBase'
go test -run 'TestWIPHumanoLidoDaBranchBase' ./internal/pipeline/
# negativo: PROPOSTA do coordenador (a confirmar pelo dono): o WIP humano recebe a contagem por FonteDePendencias; com dublê soma as fontes e fonte com erro recusa seguir
go test -list 'TestWIPHumanoContaPorFonteDePendencias' ./internal/pipeline/ | grep -F 'TestWIPHumanoContaPorFonteDePendencias'
go test -run 'TestWIPHumanoContaPorFonteDePendencias' ./internal/pipeline/
# negativo: PROPOSTA (a confirmar): a fonte padrão por arquivo conta itens review_pending e specs aguardando [Seguir] e ignora os demais estados
go test -list 'TestFonteDePendenciasPadraoPorArquivo' ./internal/pipeline/ | grep -F 'TestFonteDePendenciasPadraoPorArquivo'
go test -run 'TestFonteDePendenciasPadraoPorArquivo' ./internal/pipeline/
# negativo: a rota de split vai ao builder (nunca ao planner) e feature dependente não abre antes da predecessora mesclada
go test -list 'TestRoteamentoSplitPeloBuilderEDependenteEspera' ./internal/pipeline/ | grep -F 'TestRoteamentoSplitPeloBuilderEDependenteEspera'
go test -run 'TestRoteamentoSplitPeloBuilderEDependenteEspera' ./internal/pipeline/
# negativo: o orquestrador não mescla a spec sem o [Seguir] do dono; a rota de arquivar não depende da branch da iniciativa
go test -list 'TestRoteamentoSpecSoMesclaAposSeguirEArquivarDaPrincipal' ./internal/pipeline/ | grep -F 'TestRoteamentoSpecSoMesclaAposSeguirEArquivarDaPrincipal'
go test -run 'TestRoteamentoSpecSoMesclaAposSeguirEArquivarDaPrincipal' ./internal/pipeline/
# negativo: verificar sem --ponto, com ponto desconhecido ou sem --feature no ponto entrega é recusado
go test -list 'TestVerificarExigePontoEFeature' ./internal/pipeline/ | grep -F 'TestVerificarExigePontoEFeature'
go test -run 'TestVerificarExigePontoEFeature' ./internal/pipeline/
forge gate --dir . --mode bloqueio --branch feat/107-orquestrador-de-sessoes --spec sdd/discovery/107-orquestrador-de-sessoes/criteria.md --criterios
```
- Cada feature traz ao menos um teste negativo, checado por `forge gate --criterios` (P-25).
- Handoff com menu (RN-2): a saída do orquestrador oferece as opções e a recomendada; teste confere que não é texto solto.
- Convivência: o modo `sessions` e os testes de `laco_test.go` seguem verdes.
- Goldens e templates tocados atualizam em conjunto (`TST-02`).
