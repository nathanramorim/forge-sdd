# fix-88-01 — Teto diário de jobs derruba a esteira avançada ("Esteira parada")

**Branch:** `fix/88-01-teto-diario-nao-deve-parar-a-esteira` (a partir de `feat/88-modo-pipeline-avancado`; PR contra ela)
**Origem:** Iniciativa 88, `feat-88-07` (`internal/pipeline/safeguards.go`, `internal/pipeline/worker.go`, `internal/config/forge.go`); visto na 4.9.0-beta
**Tamanho:** `M`
**Status:** `ready`

## Sintoma
No projeto `reel-studio` (4.9.0-beta, modo avançado), chega a notificação "Esteira parada: teto diário de 30 jobs atingido (40 hoje). Remova a causa (ou o arquivo STOP) e rode forge up de novo." O modo existe para rodar por dias sem intervenção, e a parada exige que a pessoa volte e rode `forge up`.

## Feature Correlacionada
- **Arquivo de origem:** `feat-88-07-salvaguardas-e-docs.md` (RN-6 da discovery: "teto de jobs por dia").
- **Causa-raiz (verificada no código):**
  1. **O teto nasce ligado.** `DefaultForgeConfig` grava `Safeguards.MaxJobsPerDay: 30`; o modo avançado herda a parada sem a pessoa ter pedido.
  2. **A contagem não é de jobs.** `jobsStartedToday` soma cada evento `processing` de todos os jobs de hoje, então re-claim após `Recover`/`Requeue`, rework e o `pr_review` de cada PR inflam o número ("40 hoje" para menos de 40 jobs distintos).
  3. **A parada é definitiva.** Em `Run` (`worker.go`), `HaltReason` faz `return nil`: o worker encerra e só volta com `forge up`, mesmo que o dia seguinte já tenha começado. O "dia" é em UTC, não no fuso local.
  4. **Só o worker headless confere.** O despachante do modo live (`deliver.go`) não chama `HaltReason`; a contagem de 40 acima do teto de 30 sugere checagem tardia ou concorrente (a confirmar com teste).
- **Evidência:** imagem da notificação enviada pelo usuário (30 / 40 hoje) e o código acima.

## Modelo do módulo
**Enxuto** (ARQ-11): ajusta `internal/pipeline/safeguards.go`, `worker.go` e `internal/config`; sem módulo novo.

## Decisões do projeto aplicadas
ADR-0013 (modo avançado e salvaguardas), ADR-0009 (limite de uso e reset de contexto). Rules: `sdd-arquitetura`, `sdd-codigo-e-testes`.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Contagem diária | `pipeline.jobsStartedToday` | Conta eventos `processing` | **corrigir**: contar IDs distintos reivindicados hoje | O nome e a spec dizem "jobs" |
| Parada por teto | `pipeline.HaltReason`, `Run` | Encerra o worker | **estender**: o teto de jobs por dia **pausa** até o próximo dia (fuso local) e retoma sozinho; o `STOP` continua encerrando | Modo de dias sem intervenção |
| Padrão do teto | `config.DefaultForgeConfig().Safeguards` | `MaxJobsPerDay: 30` | **mudar** para `0` (desligado) | Salvaguarda opcional; quem quer teto liga |
| Chave liga/desliga | menu **Settings** (feat-88-08) | Só tem a chave do card | **estender** com "Teto diário de jobs" | RN-9: novas chaves entram no Settings |
| Aviso | `notifyHalt` | Notificação "Esteira parada" | **reusar** com texto "Esteira em pausa até <data>" | Evita notificação enganosa de parada definitiva |

## Dependências novas
Nenhuma.

## Rules e IDs da Constituição Aplicáveis
`ARQ-01`, `COD-01`, `COD-03`, `DUP-01`, `TST-01`.

## Premissas (a confirmar no Gate 1)
- Teto desligado por padrão (`max_jobs_per_day: 0`); quem ligar passa a ter pausa até o dia seguinte, não parada.
- `STOP` e os tetos por job (tokens e tempo) ficam como estão.
- `forge.json` existente com `max_jobs_per_day` gravado mantém o valor.

## Escopo
1. `DefaultForgeConfig`: `MaxJobsPerDay` = 0.
2. `jobsStartedToday`: contar IDs distintos; dia no fuso local.
3. `Run`: ao atingir o teto, aguardar até o início do próximo dia e continuar, sem encerrar; uma notificação por dia ("em pausa até …").
4. Despachante live: aplicar a mesma checagem antes de entregar o próximo job.
5. Settings: chave "Teto diário de jobs" (desativado / valor).
6. Testes de regressão e atualização de `docs/pipeline.md` e do LEIA-ME.

## Fora do escopo
Mudar `STOP`, os tetos por job ou o modo normal.

## Critérios de Aceitação Executáveis
```bash
go test -v -race ./internal/pipeline -run 'TestSalvaguardas|TestHaltReason|TestRun_PausaNoTeto'
go test -v -race ./internal/config ./cmd/forge-sdd -run 'TestDefaultSafeguards|TestMenu_Settings'
go vet ./... && go test ./...
```
- Padrão de uma esteira nova: sem teto diário.
- 1 job reclamado 3 vezes conta como 1.
- No teto, o worker não encerra: pausa, notifica uma vez e retoma no dia seguinte.
- `STOP` continua encerrando de forma limpa.
