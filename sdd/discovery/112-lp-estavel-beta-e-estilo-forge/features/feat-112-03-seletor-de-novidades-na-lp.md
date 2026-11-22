# feat-112-03 — Seletor de novidades da versão na LP, alimentado pelo `history.md`

**Iniciativa:** 112-lp-estavel-beta-e-estilo-forge
**Branch:** feat/112-lp-estavel-beta-e-estilo-forge
**Status:** spec pronta, aguarda Gate 1 (21/11/2026)
**Depende de:** feat-112-01 e feat-112-02 (as três editam `docs/index.html`; esta mexe no rodapé e cria uma seção antes dele, a 02 já mexeu nos chips, no terminal e na instalação, e a 01 removeu os blocos da Edição Missões)

## Descrição de negócio
- **Problema:** a LP mostra as versões estável e beta no topo, mas não conta **o que mudou** em cada uma. Quem chega pelo site tem de sair da página e abrir `sdd/releases/history.md` no GitHub; o rodapé só diz "veja as novidades" apontando para fora. E cada publicação exige editar a LP à mão — nada liga a release notes ao que o site exibe.
- **Para quem:** quem visita a LP para decidir se instala ou atualiza; o dono, que hoje repete o trabalho de resumir a release em cada entrega.
- **Valor:** o site passa a contar as novidades da versão sozinho, com as mesmas duas versões que o topo mostra, e a lista acompanha toda publicação sem ninguém editar HTML.

## Problema
`docs/index.html` exibe as versões (`data-versao="latest"` / `data-versao="beta"`, blocos `/* versoes:inicio */ … /* versoes:fim */`), mas não exibe as novidades. O único consumidor do `sdd/releases/history.md` hoje é `.github/workflows/release.yml`, que o copia para a wiki — `docs/` nunca recebe esse conteúdo.

## Comportamento
1. **Fonte única:** `sdd/releases/history.md`, no formato já vigente (`### Versão X — título (Estável|Beta)` seguido de bullets `*   **Destaque:** explicação`). Nenhum texto novo de release é escrito à mão na LP (`DUP-01`).
2. **Gerador `scripts/novidades_lp.py`** (Python, só stdlib, via `uv run`, mesmo padrão de `scripts/pr_guard.py` e `scripts/licoes.py`):
   - Lê `sdd/releases/history.md` e escreve `docs/data/releases.json`.
   - Formato: `{"gerado_em": "<ISO8601>", "versao_historico": "<sha curto>", "versoes": [{"versao": "5.1.0", "rotulo": "5.1.0-beta", "canal": "beta|stable", "titulo": "<título da seção>", "itens": [{"titulo": "<parte em negrito>", "texto": "<resto>", "link": "<href opcional>"}]}]}`.
   - Só entram as **duas do topo**: a versão estável mais recente e a beta mais recente (`latest` e `beta` do npm); as demais ficam fora do JSON. Se o npm não responder, o script usa a estável e a beta mais recentes do próprio histórico.
   - Idempotente: sem mudança no `history.md`, o JSON sai idêntico (o `gerado_em` só muda quando o conteúdo muda), para o diff não sujar o CI.
   - Recusa `..`, caracteres de controle e seção sem bullets (`SEG-01`, `SEG-02`); falha com mensagem clara se o `history.md` tiver seção malformada.
3. **Seção `#novidades` na LP**, imediatamente antes do `footer`, entre marcadores `<!-- novidades:inicio -->` e `<!-- novidades:fim -->`:
   - **Coluna esquerda:** duas abas compactas, no visual de `.gov-tab` (mesma borda, mesmo estado ativo, mesma tipografia), rotuladas "Estável · <versão>" e "Beta · <versão>", reaproveitando `data-versao` para não ter duas fontes de versão.
   - **Coluna direita:** lista compacta da versão selecionada — o destaque (parte em negrito do bullet) à esquerda e a explicação à direita, uma linha por item, sem painel grande nem personagens (o inspetor da Governança não é reusado, só o visual das abas).
   - **Rodapé da seção:** link "Ver release notes completas" → `https://github.com/nathanramorim/forge-sdd/blob/main/sdd/releases/history.md`, ancorado na versão selecionada.
   - **Estados:** carregando (esqueleto discreto), vazio ("As novidades desta versão ainda não foram publicadas") e erro (mantém a última lista carregada na sessão).
4. **Consumo na LP:** função `buscarNovidades()` isolada entre `/* novidades:inicio */` e `/* novidades:fim */`, com `AbortController` e timeout de 3 s, no mesmo espírito de `buscarCanalNpm`. Ordem de tentativa: `docs/data/releases.json` (mesma origem) → raw do `history.md` no GitHub → última lista conhecida → vazio. Todo texto entra por `textContent`/`createElement`, **nunca** `innerHTML` (`SEG-01`); `link` só é aceito com `https://` e domínio do próprio GitHub do projeto.
5. **Automatizado em toda publicação** — duas emendas, sem criar caminho de release novo (`FLX-07`):
   - `scripts/lp-deploy.sh` (chamado também pelo `deploy-lp.yml`, só manual desde 22/11/2026, fix-106-11): roda `scripts/novidades_lp.py` e o `--check` antes do deploy; o site nunca sai sem a seção.
   - `.github/workflows/release.yml`: depois de publicar a tag, roda o gerador e commita `docs/data/releases.json` na `main` (com `[skip ci]` ou caminho que não dispare release), para o arquivo versionado não divergir do histórico.
6. **Privacidade:** nenhuma chamada nova a terceiros — o JSON é da própria origem, e o fallback vai ao GitHub, que já é link do rodapé. Sem cookie, sem telemetria, `credentials: 'omit'`, `referrerPolicy: 'no-referrer'`.
7. **CSP:** não existe hoje em `docs/`, `site/wrangler.jsonc` nem `site/public/_headers`; nenhum cabeçalho muda. Se um dia houver, `connect-src 'self' https://raw.githubusercontent.com` passa a ser obrigatório (registrado aqui, não feito).
8. **Fora do escopo:** wiki (segue o D-5 da 112) e a LP em inglês.

## Modelo do módulo
**Enxuto** (`ARQ-11`): um script Python de stdlib, um JSON gerado, HTML/JS estático e um arquivo de teste Go. Nenhum pacote Go novo e nada novo no binário (`ARQ-04`).

## Decisões do projeto aplicadas
- `COD-01` — o gerador só converte; a LP só exibe.
- `DUP-01` / `DUP-02` — o conceito "release notes" continua com uma casa só (`sdd/releases/history.md`); a LP ganha visualização, não uma segunda cópia editável. O padrão visual de abas vem de `.gov-tab`, não de um componente novo.
- `DUP-10` — a seção respeita o orçamento da página; lista compacta, sem painel grande.
- `SEG-01`, `SEG-02` — JSON e `history.md` tratados como dado não confiável na borda; texto escapado.
- `SEG-05`, `SEG-07` — nada do usuário na chamada; nenhuma dependência nova.
- `TST-01`, `TST-05` — o contrato da seção vale no portão local e no CI.
- `FLX-07` — a automação se pendura nos workflows de release e de deploy que já existem; nenhum fluxo paralelo.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Lista de novidades na LP | `docs/index.html` | Ausente; rodapé só linka para fora | **Criar** seção `#novidades` antes do `footer` | Pedido do dono |
| Abas com painel lateral | `docs/index.html` `#governanca` (`.gov-tabs`, `.gov-tab`) | Existe, acoplada a simulação e avatares | **Reusar o visual das abas**; não reusar o inspetor | `DUP-01` |
| Leitura de versão ao vivo | bloco `/* versoes:inicio */ … /* versoes:fim */` | `buscarCanalNpm`, `escolherVersoes`, `aplicarVersoesNoDOM` com `AbortController` e fallback | **Reusar o padrão** (timeout, validação, `textContent`) | Mesmo problema, mesma solução |
| Fonte das notas | `sdd/releases/history.md` | `### Versão X — título (Estável/Beta)` + bullets `*   **Destaque:** texto` | **Reusar** como fonte única | `DUP-01` |
| Conversor de notas | `.github/workflows/release.yml`, passo `Extract Release Notes` | Node inline que extrai **uma** seção para o GitHub Release | **Adaptar** para todas as seções relevantes, em `scripts/`, sem duplicar a lógica | `DUP-02` |
| Teste de contrato da LP | `internal/scaffold/lp_versoes_test.go` | Valida marcadores, fallback e função pura via `node -e`, com `t.Skip` sem node | **Reusar o estilo** em `lp_novidades_test.go` | `TST-01` |
| Publicação da LP | `.github/workflows/deploy-lp.yml` | `wrangler deploy` de `docs/` | **Estender** com a geração antes do deploy | `FLX-07` |
| Caminho de release | `.github/workflows/release.yml` | Tag `v*` → notas → goreleaser → npm → wiki | **Estender** com a geração e commit do JSON | `FLX-07` |

## Dependências novas
`Nenhuma` (`SEG-07`). Python com stdlib apenas, no mesmo padrão dos scripts já entregues.

## Rules e IDs da Constituição Aplicáveis
`ARQ-01`, `ARQ-04`, `COD-01`, `COD-03`, `DUP-01`, `DUP-02`, `DUP-10`, `FLX-07`, `SEG-01`, `SEG-02`, `SEG-04`, `SEG-05`, `SEG-07`, `TST-01`, `TST-05`.
Rules: `sdd-codigo-e-testes`, `sdd-fluxo-e-portao`.

## Critérios de Aceitação Executáveis
Testes em `internal/scaffold/lp_novidades_test.go`: `TestLP_NovidadesMarcadores`, `TestLP_NovidadesJsonVersionado`, `TestLP_NovidadesFuncaoPura` (`node -e` sobre o bloco, `t.Skip` sem node), `TestNovidadesGerador` (roda `uv run scripts/novidades_lp.py --check` e compara com o JSON versionado).
```bash
set -o pipefail
uv run scripts/novidades_lp.py --check          # JSON versionado em dia com o history.md
python3 -c "import json;d=json.load(open('docs/data/releases.json'));assert d['versoes'],'sem versoes';assert all(v['itens'] for v in d['versoes'])"
grep -q 'novidades:inicio' docs/index.html
grep -q 'novidades:fim' docs/index.html
grep -qF 'releases/history.md' docs/index.html     # link para as notes completas
grep -qF 'data-versao="latest"' docs/index.html    # as abas usam as mesmas versões do topo
go test ./internal/scaffold -run 'TestLP_Novidades|TestNovidadesGerador'
# negativo: a seção não usa innerHTML nem polling
bloco="$(sed -n '/novidades:inicio/,/novidades:fim/p' docs/index.html)"
test -n "$bloco"
! printf '%s' "$bloco" | grep -E 'innerHTML|setInterval|localStorage' || exit 1
printf '%s' "$bloco" | grep -qF 'AbortController'
# negativo: nenhuma dependência nova e nenhum endpoint de terceiro
! grep -F -- '/-/package/' docs/index.html || exit 1
go vet ./... && go test ./...
```
- Teste negativo do gerador: `history.md` de teste com seção sem bullets e com caractere de controle → o script sai com erro e **não** regrava o JSON.
- Teste negativo da função pura (`TestLP_NovidadesFuncaoPura`): item com `<img src=x onerror=1>`, `javascript:alert(1)` em `link` e objeto sem `itens` são descartados; item válido passa.
- Prova de borda (manual): abrir a LP com o `docs/data/releases.json` removido e conferir que a lista cai no raw do GitHub e, sem rede, mostra a última lista conhecida ou o vazio — nunca erro na tela.

## A confirmar no Gate 1
- Quantidade de itens por aba: a spec mostra **todos** os bullets da versão; se a lista ficar longa demais (ex.: 5.0.0 tem 9), o corte para os 5 primeiros com "ver as demais nas release notes" fica para a implementação, a seu critério.
- O commit automático do JSON no `release.yml` empurra direto na `main`. Se você preferir PR automático em vez de push, dizer agora (padrão da spec: push direto com `[skip ci]`, porque é artefato gerado).