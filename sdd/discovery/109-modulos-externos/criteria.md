# Critérios — Iniciativa 109 (módulos externos do forge-sdd)

Legenda dos diagramas: **HUMANO** (decisões e aprovações do dono), **IA** (agentes) e **AUTOMÁTICO** (código do forge, sem LLM).

## Restrições (Constitution)
- `ARQ-01`, `ARQ-02` (lógica em `internal/modulo`; `cmd/` só adapta flags; `internal/modulo` não importa `internal/pipeline`), `ARQ-04` (só stdlib: `crypto/sha256`, `os/exec`, `encoding/json`), `ARQ-05` (`forge modulo instalar --dry-run` sem efeito), `ARQ-06` (`modulo` entra na lista da regra pela 106/0.8), `ARQ-10`, `ARQ-11`.
- `SEG-01` (manifesto, evento e saída do módulo validados na borda), `SEG-02` (caminhos dentro do projeto e da pasta do usuário, sem link simbólico para fora, gravação atômica do registro), `SEG-03` (argumentos explícitos; **quebra deliberada da lista fixa `git`/`gh`**, por ADR), `SEG-04`, `SEG-05` (ambiente por lista, nada de segredo em log), `SEG-08` (saída do módulo saneada).
- `DUP-01` (dono do conceito: `internal/modulo`), `DUP-02`, `DUP-03`, `DUP-04` (versão do contrato definida uma vez), `DUP-06`.
- `FLX-01`, `FLX-05` (ADR do contrato e da confiança), `GOV-01` (campo da Constituição no PR único da 106/0.8).
- `TST-01`, `TST-02` (goldens), `TST-03` (casos conformes e negativos), `TST-04` (sem rede, `t.TempDir()`), `COD-05`.

## Modelo do módulo
**Completo** (`ARQ-11`): pacote novo `internal/modulo` (manifesto, contrato, confiança, execução, eventos, instantâneo), comando `cmd/forge-sdd/modulo.go` (oculto) e extensão do `doctor` e do `--help`. Justificativa: tem estado (registro de confiança), borda externa (processo de terceiros) e fronteira de segurança.

## Decisões do projeto aplicadas
- **ADR-0004:** o Go decide o fluxo; o módulo só reage. Nenhuma decisão da esteira passa por ele.
- **ADR-0012:** o Go grava o veredito; o evento `veredito.gravado` só o informa.
- **ADR-0017** (MCP somente leitura com saída saneada) e **ADR-0018** (gravação segura): a saída do módulo é dado não confiável e o registro de confiança usa gravação segura.
- **ADR-0003/P-24:** texto novo em comandos e skills entra por referência; a Constituição fica em 9.173 de 9.216 bytes.
- **ADR do 109 (task 109-12, onda 4, antes da 109-05; a 109-11 só o texto mínimo; número no merge; 0021 e 0022 já ocupados ou reservados):** contrato v1, modelo de confiança, quebra da lista fixa de comandos.
- **Memória do dono:** handoff é menu (autorização do módulo é menu).

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| LookPath injetável | `internal/pipeline/preflight.go:35-45`, `launcher.go:49`, `internal/queue/pr.go:14` | Padrão de injeção para teste | **Reusar o padrão** | sem importar `pipeline` (`ARQ-02`) |
| Execução com timeout | `internal/pipeline/runner.go:133-146` | `CommandContext`, `WaitDelay`, ambiente herdado inteiro | **Criar** com ambiente por lista | herdar tudo vazaria segredo |
| Caminho seguro | `internal/scaffold/hooks.go:181,231`; `pipeline/worktree.go:56` | `checkSafeTarget`, `writeFileSafe`, `EvalSymlinks` | **Reusar a técnica**; extrair função comum na 3ª repetição | `DUP-05` |
| `--version` de agente | busca em `internal/` e `cmd/` | nenhuma | **Criar** o aperto de mão | |
| Verificação de integridade | busca por `sha256` | só IDs | **Criar** | |
| `doctor` | `internal/doctor/harness.go:66-80` | lista fixa de 7 checagens, `ExitFalha=3` | **Estender** condicionalmente | byte a byte sem módulo |
| Orçamento da Constituição | `internal/gate/gate.go:789` | 9.216 bytes | **Reusar** | |
| Eventos de domínio | `internal/notify`, `internal/bus` | nenhum canal de eventos de domínio | **Criar** em `internal/modulo` | `DUP-01` |

## Dependências novas
Nenhuma dependência de módulo Go (`ARQ-04`, `SEG-07` não se aplica). O repositório do módulo é outro projeto, com a stdlib como regra dele.

## Diagramas (C4 em Mermaid)
Contexto e sequência estão em `diagramas/` (PNG e JSON do diagram-studio): `forge-109-fronteira.png` (core, contrato, módulo) e `forge-109-sequencia.png` (instalar, verificar, evento, módulo age). Contêineres: `forge` (core) lê a Constituição, o registro de confiança e o manifesto; chama o binário do módulo por processo; o módulo chama `gh`.

## Critérios de Aceitação Executáveis
Os nomes de teste são os que as tasks devem criar; a verificação estática do portão (`--criterios`) confere forma, não execução.

```bash
set -o pipefail
go build ./... && go vet ./...
# contrato v1: manifesto, faixa de versão e recusa clara
go test ./internal/modulo/... -run 'TestManifestoValidoAceito|TestContratoForaDaFaixaRecusaComMensagem|TestEnvelopeDoEventoV1'
go test ./internal/modulo/... -run 'TestCampoDesconhecidoNoEventoEIgnorado|TestRemoverCampoExigeNovoContrato'
# confiança: pasta do usuário, hash, autorização por versão
go test ./internal/modulo/... -run 'TestModuloSoDaPastaDoUsuario|TestHashDoBinarioConferidoAntesDeExecutar|TestAutorizacaoPorVersao'
go test ./internal/modulo/... -run 'TestModoSemHumanoNaoAutorizaModulo|TestAgenteNaoSeAutorizaPeloProjeto'
go test ./internal/modulo/... -run 'TestAmbienteSoComListaPermitida|TestSaidaLimitadaESaneada|TestTempoLimiteViraAviso'
# eventos e instantaneo (semânticos, sem caminho de arquivo)
go test ./internal/modulo/... -run 'TestEventosSaemDoFeatureEstadoMudou|TestEventoSpecAprovadaNoGate1|TestEventoPRAbertoComSHA|TestInstantaneoSomenteLeitura'
go test ./internal/modulo/... -run 'TestPayloadNaoTemCaminhoDeArquivo'
# Constituição (PR único da 106/0.8): campo, orçamento, rule
test "$(wc -c < sdd/memory/constitution.md)" -le 9216
grep -F '| Módulos permitidos | nenhum |' sdd/memory/constitution.md
grep -F '| Início automático do Builder |' sdd/memory/constitution.md
go test ./internal/gate/... -run 'TestCheckP24_RepositorioDoForgeSddPassaNoProprioPortao'
# byte a byte sem módulo (RN-1, requisito inegociável)
# os testes existem de fato (P-25): -run sozinho passa sem rodar nada
set -o pipefail
go test ./cmd/... -list 'TestSemModuloHelpIdentico' | grep -F 'TestSemModuloHelpIdentico'
go test ./cmd/... -list 'TestSemModuloDoctorIdentico' | grep -F 'TestSemModuloDoctorIdentico'
go test ./cmd/... -list 'TestSemModuloStatusIdentico' | grep -F 'TestSemModuloStatusIdentico'
go test ./cmd/... -list 'TestSemModuloComandoDesconhecidoIdentico' | grep -F 'TestSemModuloComandoDesconhecidoIdentico'
go test ./cmd/... -list 'TestSemModuloCodigosDeSaidaIdenticos' | grep -F 'TestSemModuloCodigosDeSaidaIdenticos'
go test ./internal/pipeline/... -list 'TestEsteiraCompletaIdenticaSemModulo' | grep -F 'TestEsteiraCompletaIdenticaSemModulo'
go test ./internal/modulo/... -list 'TestCampoNenhumNaoTocaDiscoNemProcesso' | grep -F 'TestCampoNenhumNaoTocaDiscoNemProcesso'
go test ./internal/modulo/... -list 'TestCampoAusenteVaiNenhumSemAviso' | grep -F 'TestCampoAusenteVaiNenhumSemAviso'
go test ./internal/scaffold/... -list 'TestInitNaoGeraCampoDeModulos' | grep -F 'TestInitNaoGeraCampoDeModulos'
go test ./cmd/... -run 'TestSemModuloHelpIdentico|TestSemModuloDoctorIdentico|TestSemModuloStatusIdentico'
go test ./cmd/... -run 'TestSemModuloComandoDesconhecidoIdentico|TestSemModuloCodigosDeSaidaIdenticos'
go test ./internal/pipeline/... -run 'TestEsteiraCompletaIdenticaSemModulo'
go test ./internal/modulo/... -run 'TestCampoNenhumNaoTocaDiscoNemProcesso|TestCampoAusenteVaiNenhumSemAviso'
go test ./internal/scaffold/... -run 'TestInitNaoGeraCampoDeModulos'
# módulo ausente, quebrado ou incompatível nunca derruba o core
go test ./internal/modulo/... -run 'TestModuloAusenteIgnoradoSemSaida|TestModuloQuebradoIgnoradoEMensagemSoNoDoctor'
go test ./internal/modulo/... -run 'TestModuloIncompativelIgnoradoEMensagemSoNoDoctor|TestModuloLentoNaoBloqueiaAFila'
go test ./internal/doctor/... -run 'TestDoctorModuloProblemaViraAvisoNuncaFalha|TestDoctorListaModuloAutorizado'
# P-25 (109-12): ADR da execução existe e cita FLX-05 e SEG-03 (número definido no merge)
test -n "$(ls sdd/spec/adr/adr-*-execucao-de-modulos-externos.md 2>/dev/null)"
grep -F 'FLX-05' sdd/spec/adr/adr-*-execucao-de-modulos-externos.md
grep -F 'SEG-03' sdd/spec/adr/adr-*-execucao-de-modulos-externos.md
# negativo: cada caso abaixo deve falhar, e o teste existe de fato
go test ./internal/modulo/... -list 'TestBinarioAlteradoERecusado' | grep -F 'TestBinarioAlteradoERecusado'
go test ./internal/modulo/... -list 'TestCaminhoForaDoUsuarioRecusado' | grep -F 'TestCaminhoForaDoUsuarioRecusado'
go test ./internal/modulo/... -list 'TestManifestoSemContratoRecusado' | grep -F 'TestManifestoSemContratoRecusado'
go test ./internal/modulo/... -list 'TestModuloNaoListadoNaConstituicaoNaoRoda' | grep -F 'TestModuloNaoListadoNaConstituicaoNaoRoda'
go test ./internal/modulo/... -list 'TestPATHNaoEPesquisado' | grep -F 'TestPATHNaoEPesquisado'
go test ./internal/modulo/... -list 'TestSegredoNaoDeclaradoNaoChega' | grep -F 'TestSegredoNaoDeclaradoNaoChega'
go test ./internal/modulo/... -list 'TestModuloNaoPodeEscreverEstadoDoCore' | grep -F 'TestModuloNaoPodeEscreverEstadoDoCore'
go test ./internal/modulo/... -run 'TestBinarioAlteradoERecusado|TestCaminhoForaDoUsuarioRecusado|TestManifestoSemContratoRecusado|TestModuloNaoListadoNaConstituicaoNaoRoda|TestPATHNaoEPesquisado|TestSegredoNaoDeclaradoNaoChega|TestModuloNaoPodeEscreverEstadoDoCore'
```
