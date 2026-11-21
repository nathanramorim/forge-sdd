# Critérios — Iniciativa 112

## Restrições (Constitution)
`ARQ-11` (modelo enxuto: sem pacote novo; só `docs/`, scripts já existentes e testes em `internal/scaffold`), `COD-01`, `COD-04` (o ramo "edição vigente" do teste e o workflow inerte saem), `DUP-01`, `DUP-02`, `SEG-01` (entrada do registro validada na borda), `SEG-05` (sem dado do usuário na chamada), `TST-01`, `TST-05`, `FLX-01` (Gate 1 humano).

## Modelo do módulo
**Enxuto** (`ARQ-11`): não há regra de domínio nova; muda conteúdo estático (`docs/`), remove arquivos e acrescenta testes de leitura estática. Justificativa: nenhum pacote Go novo.

## Decisões do projeto aplicadas
Nenhuma ADR muda. Regra 1 (nunca commitar em `main`): o encerramento vai por branch e PR. Regra 4 (segredos): a chamada ao registro é pública e anônima.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Encerramento da edição | `.github/scripts/encerrar-edicao/encerrar.sh` | Remove blocos `edicao:inicio/fim` de `index.html` e `wiki.html`, troca o cabeçalho de `README.md` e `npm/README.md`, apaga `docs/edicao` e a própria pasta do script | **Reusar** (rodar localmente) | `DUP-01`: um só mecanismo |
| Gatilho automático | `.github/workflows/encerrar-edicao.yml` | Cron diário; só age com vigência vencida; abre PR `chore/101-encerrar-edicao` | **Não usar**; apagar depois (D-3) | Antecipar pelo workflow exigiria mexer na data |
| Teste do encerramento | `internal/scaffold/pipeline_scaffold_test.go` `TestScaffold_DocsEdicao` | Já tem o ramo "pasta `docs/edicao` ausente" que exige 0 marcadores nos 4 arquivos | **Estender** (remover o ramo vigente) | `COD-04` |
| Chamada externa na LP | `docs/index.html` | Só three.js (cdnjs) e fontes; nenhum `fetch` | **Criar** uma chamada, a única | Pedido do dono |
| Teste Node/JS da LP | repositório | Nenhum (`LookPath("node")` não aparece) | **Criar** com `t.Skip` sem node | Sem precedente; função pura entre marcadores |
| CSP | `docs/`, `site/wrangler.jsonc`, `site/public/_headers` | Nenhuma | **Nada a mudar** | Registrado em `discovery.md` |

## Dependências novas
Nenhuma (`SEG-07`): JavaScript puro no navegador, sem biblioteca; teste usa `node` só se existir.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go build ./... && go vet ./... && go test ./...
# negativo: a LP não pode ter CSP nova nem endpoint sem CORS (nenhum cabeçalho de segurança muda nesta iniciativa)
! grep -rIl 'Content-Security-Policy' docs site/wrangler.jsonc || exit 1
! grep -F -- '/-/package/' docs/index.html || exit 1
```
Os critérios por feature estão nos arquivos de `features/`.
