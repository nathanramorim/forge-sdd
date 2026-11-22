# fix-106-06 — Ninguém pergunta, no início da sessão, se o dono quer tudo automático

**Branch:** `docs/106-specs-worktree-modo-ask` (PR único das três specs 05, 06 e 07)
**Origem:** Iniciativa 106 (ADR-0022, `feat-106-08`) e `internal/pipeline/modo_humano.go`
**Tamanho:** `M`
**Status:** `spec` (aguarda Gate 1)

## Sintoma

O dono relatou: "quando inicia a sessão não pergunta se quero deixar automático tudo". A rule `sdd-modos-de-projeto.md` (seção 3) diz que `sem-humano` "só vale se escolhido **no início da sessão**, depois do aviso dos riscos". Na prática o início da sessão não pergunta nada, então o projeto roda sempre no padrão `com-humano` **por omissão**, e quem quiser o modo automático não tem como pedir.

## Apuração

| Pergunta | Evidência | Conclusão |
|---|---|---|
| Existe código que recebe a escolha? | `internal/pipeline/modo_humano.go`: `EscolhaDeModo` ("o que o dono respondeu no início da sessão"), `ResolveModoDeRevisao`, `AvisoDoModoSemHumano`, `LeModoDaConstituicao` | **Sim, e com testes** |
| Alguém chama? | Contei os usos fora de testes e fora dos dois arquivos: `AutorizaVoltaDaSpec`, `VoltaSpecParaPrincipal`, `SinalAposVolta`, `AvisoDoModoSemHumano`, `ResolveModoDeRevisao`, `LeModoDaConstituicao` = **0 usos** | **Nenhum chamador**: código pronto sem uso |
| Existe um gancho de início de sessão? | `internal/scaffold/hooks.go`: `MergeClaudeSessionStartHook` instala `forge doctor --harness` em `SessionStart`, e "a saída vira contexto" | **Sim, mas só diagnostica** |
| A configuração foi alterada? | `Modo de revisão humana | com-humano` desde `7cbe3af` (#177), nunca mudou | **Não** |

**Correção de uma afirmação anterior:** o `fix-106-03` (#219) dizia que o modo de revisão humana "funciona; é o único campo de modo aplicado". Isso estava **errado**: o código existe e é testado, mas **nada o chama**. O `fix-106-03` é corrigido neste mesmo PR.

## Feature Correlacionada

- **Arquivo de origem:** `internal/pipeline/modo_humano.go` e `internal/pipeline/volta_da_spec.go` (entregues sem ligação ao fluxo), e `internal/scaffold/hooks.go` (gancho de sessão que não pergunta).
- **Causa-raiz:** as funções foram entregues como tasks isoladas, sem a task que **as liga** ao início da sessão. O critério de aceitação testava a função, não a jornada.
- **Evidência do defeito:** `grep -rn 'ResolveModoDeRevisao' cmd internal` fora de `_test.go` e de `modo_humano.go` não devolve nada.

## Modelo do módulo

**Completo** (`ARQ-11`): liga `internal/pipeline` (já existe), `internal/scaffold` (hook) e `cmd/forge-sdd`. Pode exigir um subcomando novo (ver decisão abaixo).

## Decisões do projeto aplicadas

- ADR-0022 e rule `sdd-modos-de-projeto.md` §3: `sem-humano` só depois do aviso de riscos aceito; padrão sempre o seguro.
- `FLX-01`: o humano aprova antes do trabalho avançar; `sem-humano` é a exceção explícita.
- `SEG-01`: valor inválido recusado na borda; sem aceite, volta para `com-humano`.
- `DUP-01`: **reusar** `ResolveModoDeRevisao` e `AvisoDoModoSemHumano`; nenhuma regra nova de aceite.
- **Regra do dono:** a pergunta é componente de ask, nunca texto solto (ver fix-106-07).
- `ARQ-06`: se houver subcomando novo, a Constituição muda e exige ADR (`GOV-01`, `FLX-05`).

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Decisão do modo e aviso | `internal/pipeline/modo_humano.go` | `ResolveModoDeRevisao`, `AvisoDoModoSemHumano`, `LeModoDaConstituicao` | **reusar**, ligando a um chamador |
| Autorizar a volta da spec | `internal/pipeline/volta_da_spec.go` | `AutorizaVoltaDaSpec`, `VoltaSpecParaPrincipal` | **ligar** ao `[Seguir]` do Gate 1 |
| Gancho de início de sessão | `internal/scaffold/hooks.go`, `forge doctor --harness` em `SessionStart` | só diagnóstico | **estender**: quando a escolha da sessão não existe, a saída acrescenta a instrução de abrir o menu |
| Estado da sessão | `.forge/runtime/` (fora do git) | sem arquivo de escolha | **criar** `.forge/runtime/sessao/<id>.json` (escrita atômica, `SEG-02`) |
| Registro no `status.md` | rule §4 ("registro no `status.md`") | texto, sem código | **gravar** a escolha `sem-humano` e o carimbo de aceite |
| Início automático do Builder | `Início automático do Builder | desligado` | o laço é da 107 e **não existe** | **fora do escopo**; o menu não oferece o que não existe |

## Dependências novas (`SEG-07`)

`Nenhuma`.

## Rules e IDs da Constituição Aplicáveis

`ARQ-01`, `ARQ-06`, `ARQ-11`, `COD-03`, `DUP-01`, `FLX-01`, `SEG-01`, `SEG-02`, `TST-01`, `TST-03`.

Rules: `sdd-arquitetura`, `sdd-codigo-e-testes`, `sdd-fluxo-e-portao`, `sdd-modos-de-projeto`.

## Escopo

**Dentro:**
1. **Pergunta uma vez por sessão.** O hook de `SessionStart` detecta que a sessão não tem escolha gravada e acrescenta ao contexto: "DECISÃO PENDENTE: abra o menu do modo da sessão". O agente abre `AskUserQuestion`:
   - **Com humano (padrão, recomendado)** · **Sem humano (automático)** · **Decidir depois (segue com humano)**.
2. **Sem humano exige aviso:** se escolher, o agente mostra o texto de `AvisoDoModoSemHumano()` em um segundo menu: **Aceito os riscos** · **Voltar**. Só `Aceito` vale; qualquer outra resposta fica `com-humano`.
3. **Gravar a escolha** em `.forge/runtime/sessao/<id>.json` e, quando `sem-humano`, no `status.md` da iniciativa, com o carimbo (`ResolveModoDeRevisao` → `Registrar`).
4. **Não perguntar de novo** em `resume` e `compact` da mesma sessão (o hook confere a origem do evento).
5. **Ligar a escolha ao fluxo:** `AutorizaVoltaDaSpec` passa a ser chamada no `[Seguir]` do Gate 1. No `sem-humano`, a spec volta à principal depois de veredito conforme sem esperar o dono; no `com-humano`, continua esperando.
6. `forge brief` e `forge status` mostram o modo efetivo da sessão.

**Decisão aberta para o Gate 1 — como gravar a escolha:**

| Opção | O que é | Custo |
|---|---|---|
| **A. Subcomando `forge session mode`** (recomendada) | O Go valida (`ResolveModoDeRevisao`) e grava; o agente só chama | muda `ARQ-06`, exige ADR curto no mesmo PR |
| **B. O agente grava o arquivo direto** | Sem subcomando novo | o agente passa a escrever estado de runtime; a validação só na leitura (`SEG-01` mais fraco) |

**Fora:**
- O início automático do Builder e o laço da 107.
- Alterar o campo `Modo de revisão humana` da Constituição (a escolha de sessão **não** o sobrescreve; vale só na sessão).
- Perguntar o modo a cada prompt.

## Critérios de Aceitação Executáveis

```bash
set -o pipefail
# 1. O modo deixou de ser código sem chamador.
test "$(grep -rn --include='*.go' 'ResolveModoDeRevisao' cmd internal | grep -v '_test.go' | grep -v 'modo_humano.go' | wc -l | tr -d ' ')" -ge 1
# 2. Aviso obrigatório: sem-humano sem aceite volta para com-humano.
go test ./internal/pipeline/... -run 'TestSemHumanoSemAvisoRecusado|TestSemHumanoSemAceiteVoltaParaComHumano'
# 3. A escolha é gravada, uma vez por sessão.
go test ./internal/pipeline/... -run 'TestSessaoGravaModoEfetivo|TestNaoPerguntaDeNovoNaMesmaSessao'
# 4. O hook de abertura pede a decisão só quando falta escolha.
go test ./cmd/forge-sdd/... -run 'TestDoctorHarnessPedeModoNaAbertura|TestDoctorHarnessNaoPedeSeJaEscolhido'
# 5. A volta da spec respeita o modo.
go test ./internal/pipeline/... -run 'TestVoltaDaSpecComHumanoEspera|TestVoltaDaSpecSemHumanoDepoisDoConforme'
go vet ./... && go test ./...
```

- **Teste negativo obrigatório (`TST-03`):** modo inválido (`automatico`) é recusado com o erro de `ErrModoRevisaoInvalido`; `sem-humano` com `Aceito=false` ou `AvisoApresentado=false` devolve `com-humano` e o erro `ErrSemHumanoSemAviso`.
- **Prova de que deixou de ser órfão:** o critério 1 falha enquanto nenhum código de produção chamar `ResolveModoDeRevisao`.
- **A conferir na implementação:** os campos que o Claude Code entrega ao hook de `SessionStart` (origem do evento e identificador de sessão) na documentação da versão em uso.

## Riscos

- **R1 — `sem-humano` aprova erro do revisor.** Mitigação: o aviso de riscos obrigatório e aceito; o registro no `status.md`; vale só na sessão, não vira padrão do projeto.
- **R2 — A pergunta vira ruído** em toda abertura. Mitigação: uma vez por sessão; "Decidir depois" segue com o padrão seguro.
- **R3 — Outros agentes (Gemini, Copilot) não têm hook de sessão.** Mitigação: neles a pergunta fica no primeiro comando do forge, por lista numerada, como o `clarify.md` já prevê.
