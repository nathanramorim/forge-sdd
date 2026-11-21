# feat-112-01 — Encerrar a Edição Missões antes do prazo

**Iniciativa:** 112-lp-estavel-beta-e-estilo-forge
**Branch:** feat/112-lp-estavel-beta-e-estilo-forge
**Status:** spec pronta, aguarda Gate 1 (21/11/2026)
**Depende de:** nenhuma (a feat-112-02 depende desta)

## Descrição de negócio
- **Problema:** a LP, o README e o npm/README ainda usam o tema da Edição Missões (neon, palmeiras, "missões"), que só sairia sozinho em 23/11/2026.
- **Para quem:** o dono e quem visita o projeto.
- **Valor:** a cara do Forge volta agora, sem esperar o relógio e sem copiar nada à mão.

## Problema
O fim automático depende da data em `docs/edicao/edicao.js` (`fim: 2026-11-23T00:00:00-03:00`) e do cron do workflow. Antecipar mexendo na data ou disparando o workflow não funciona: o workflow só age quando `vigente(Date.now())` é falso.

## Comportamento
1. **Como antecipar (D-1):** a partir de `main` atualizada, na branch `feat/112-lp-estavel-beta-e-estilo-forge`, rodar `bash .github/scripts/encerrar-edicao/encerrar.sh .`; commitar; o PR da iniciativa leva a mudança (nada em `main`, nada de data alterada, nada de `workflow_dispatch`). O mesmo script é o que o workflow roda, então o resultado é idêntico ao do dia 23.
2. **O que sai (o script):** `docs/edicao/` inteira (`edicao.js`, `edicao.css`, `edicao-mascotes.js`, `equipe.svg`, `hero-arte.svg`, `personagens/`); em `docs/index.html` o bloco de `<link>/<script>` (linhas ~1099-1103) e a chamada `vestirMascote` (~1774-1776); em `docs/wiki.html` o bloco ~11-14; o cabeçalho temático (linhas 1-19) de `README.md` e `npm/README.md`, trocado pelo padrão (banner e selos, "Versão atual: 5.0.0 (estável)"); `.github/scripts/encerrar-edicao/`.
3. **O que sai além do script:** `.github/workflows/encerrar-edicao.yml` (D-3).
4. **O que PERMANECE:** o design system base da LP (tokens e cores do próprio `index.html`, fontes, three.js e os mascotes 3D dos bots, que são da página); `docs/design/edicao-missoes/` e `docs/img/banner-v5-missoes.*` (D-2, recomendado manter, sem uso pela LP); a linha 5.0.0 das tabelas de versões do README/npm/README e o histórico em `sdd/releases/history.md` e `sdd/discovery/101-*` (história, não tema).
5. **Teste:** `TestScaffold_DocsEdicao` já tem o ramo "pasta ausente" que exige zero marcadores em `docs/index.html`, `docs/wiki.html`, `README.md` e `npm/README.md`; com a pasta removida ele passa sem mudar. Ajuste mínimo: apagar o ramo "edição vigente" (código morto depois do encerramento, `COD-04`) e exigir a pasta ausente; manter as três asserções de `validarMarcadoresEdicao`. Novo `TestLP_SemTemaDaEdicao` confere que `index.html` e `wiki.html` não citam `forgeEdicao`, `data-edicao` nem `edicao/`.
6. **Prova de que nada sobrou** (simulada numa cópia: só sobram as exceções abaixo; "permissões" casa com "missões" e "medicao" com "edicao", então a busca usa `\b` e `[^a-z]`; resultado vazio após o script):
   exceções justificadas: tabelas de versão (`| **5.0.0** |`) de `README.md` e `npm/README.md`; `docs/design/edicao-missoes/**`; `docs/img/banner-v5-missoes.*`; `sdd/releases/history.md`.

## Modelo do módulo
**Enxuto** (`ARQ-11`): remoção de conteúdo e um teste; nenhum pacote novo.

## Decisões do projeto aplicadas
Regra 1 (branch e PR, nada em `main`), `COD-04` (sai o que ficou inerte), `DUP-01` (reusa o `encerrar.sh`).

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Script de encerramento | `.github/scripts/encerrar-edicao/encerrar.sh` | Faz tudo da lista acima; rc=0 numa cópia | **Reusar** | `DUP-01` |
| Cabeçalhos padrão | `README.cabecalho.md`, `npm.README.cabecalho.md` | Banner, selos, "5.0.0 (estável)" | **Reusar** | Já escritos para isto |
| Teste do encerramento | `TestScaffold_DocsEdicao` | Ramo pós-encerramento existe | **Estender** | `COD-04` |
| Workflow | `encerrar-edicao.yml` | Inerte após o merge | **Remover** | Não faz mais nada |

## Dependências novas
Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`COD-04`, `DUP-01`, `TST-01`, `TST-05`. Rules: `sdd-codigo-e-testes`.

## Critérios de Aceitação Executáveis
```bash
set -o pipefail
go test ./internal/scaffold -list 'TestScaffold_DocsEdicao|TestLP_SemTemaDaEdicao' | grep -F 'TestLP_SemTemaDaEdicao' >/dev/null
go test ./internal/scaffold -run 'TestScaffold_DocsEdicao|TestLP_SemTemaDaEdicao'
test ! -e docs/edicao || exit 1
test ! -e .github/scripts/encerrar-edicao || exit 1
test ! -e .github/workflows/encerrar-edicao.yml || exit 1
# negativo: nenhum marcador residual
! grep -rIl 'edicao:inicio\|edicao:fim' docs README.md npm/README.md || exit 1
# negativo: nada do tema sobrou (exceções: tabela de versões, design system e banner mantidos, ver D-2)
sobra="$(grep -rInE -i '(^|[^a-z])edicao|forgeedicao|\bmissões\b|\bmissoes\b|\bgta\b' docs README.md npm/README.md --exclude-dir=edicao-missoes --exclude='banner-v5-missoes.*' | grep -vE '^(README|npm/README)\.md:[0-9]+:\| \*\*5\.0\.0\*\* \|' || true)"
test -z "$sobra"
# positivo: os mascotes 3D da página continuam
grep -q 'new THREE' docs/index.html
go vet ./... && go test ./...
```
- Teste negativo: `validarMarcadoresEdicao` recusa marcador residual pós-encerramento (já coberto no teste atual).

## A confirmar no Gate 1
D-1 (rodar o script local), D-2 (manter design system e banner) e D-3 (apagar o workflow). A publicação depende do deploy da Cloudflare.
