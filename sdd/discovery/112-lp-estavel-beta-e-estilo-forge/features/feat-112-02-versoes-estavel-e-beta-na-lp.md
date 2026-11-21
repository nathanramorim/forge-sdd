# feat-112-02 — Versões estável e beta na LP, lidas do npm ao vivo

**Iniciativa:** 112-lp-estavel-beta-e-estilo-forge
**Branch:** feat/112-lp-estavel-beta-e-estilo-forge
**Status:** spec pronta, aguarda Gate 1 (21/11/2026)
**Depende de:** feat-112-01 (as duas editam `docs/index.html`; a 01 apaga os blocos ~1099-1103 e ~1774-1776, a 02 mexe em ~1133, ~1253, ~1317, ~1576 e ~2085; sem a ordem há conflito de merge)

## Descrição de negócio
- **Problema:** a LP só fala da 5.0.0. Quem quer testar a 5.1.0-beta não sabe que ela existe nem como instalar.
- **Para quem:** quem visita a LP para instalar; o dono, que não quer editar o HTML a cada release.
- **Valor:** a LP mostra a estável e a beta sozinha, com os dois comandos de instalação e o aviso de qual usar.

## Problema
`docs/index.html` tem `v5.0.0` escrito à mão (chip ~1133, terminal ~1253 e ~2085, rodapé ~1576) e um único comando de instalação (~1317). A página não faz `fetch` hoje.

## Comportamento
1. **Endpoint (confirmado em 21/11/2026 com `curl`):** duas leituras GET, `https://registry.npmjs.org/@nathanramorim/forge-sdd/latest` e `.../beta`, ambas com `access-control-allow-origin: *` e campo `version`. Não usar `.../-/package/@nathanramorim/forge-sdd/dist-tags`: responde 200, mas **sem** cabeçalho CORS (o navegador bloquearia). Não usar CDN de terceiros (jsDelivr): seria um segundo destino de privacidade.
2. **Fallback estático obrigatório:** o HTML já nasce com `5.0.0` e `5.1.0-beta` dentro de elementos com `data-versao="latest"` e `data-versao="beta"`. Sem JavaScript, ou se a chamada falhar, passar de **3 s** (`AbortController`) ou vier fora do formato, nada muda na tela.
3. **Validação estrita:** só entra `version` que casar com `^\d+\.\d+\.\d+$` (estável) ou `^\d+\.\d+\.\d+-beta(\.\d+)?$` (beta); qualquer outro valor é descartado e o campo mantém o estático. Inserção sempre com `textContent`, nunca `innerHTML`. Cada campo é validado sozinho (uma tag ruim não derruba a outra).
4. **Privacidade:** `fetch(url, { credentials: 'omit', referrerPolicy: 'no-referrer', cache: 'default', signal })`; sem dado do usuário, sem cookie nosso, sem telemetria. O Cloudflare do registro pode setar o cookie dele (`__cf_bm`); como a chamada é `credentials: 'omit'`, o navegador não o guarda para esta origem. Frase discreta no rodapé: "As versões vêm do registro público do npm; se ele não responder, mostramos as últimas conhecidas."
5. **Cache e desempenho:** uma chamada por visita, sem polling, sem `setInterval`, sem `localStorage`; roda após `DOMContentLoaded`, sem bloquear a renderização. `prefers-reduced-motion` não é tocado.
6. **Onde aparece:** chips "Estável X" e "Beta Y" no topo (substituem o chip único ~1133); rodapé ~1576 mostra a estável; terminal de exemplo (~1253 e ~2085) segue a estável (D-4); seção de instalação com dois comandos: `npm install -g @nathanramorim/forge-sdd` (estável, recomendado) e `npm install -g @nathanramorim/forge-sdd@beta` com o aviso "Para testar novidades. A beta não é a versão recomendada para produção." O comparativo "Antes (v4.x) vs Agora (v5.0)" (~1495-1524) **não muda**.
7. **CSP:** não existe em `docs/`, `site/wrangler.jsonc` nem `site/public/_headers`; nenhum cabeçalho muda. Se um dia houver CSP, `connect-src https://registry.npmjs.org` é obrigatório (item de segurança do dono; não feito aqui).
8. **Estrutura testável:** a lógica pura (`versaoValida`, `escolherVersoes`) fica entre `/* versoes:inicio */` e `/* versoes:fim */`, sem tocar o DOM; o resto (fetch e `textContent`) usa essas funções.
9. **Wiki:** fora do escopo (D-5); seguimento.
10. **Publicação:** só vai ao ar com o deploy da Cloudflare corrigido (token sem Workers Scripts/Routes em Editar); é dependência de publicação, não de construção.
11. **Nota de release:** o checklist ganha "conferir a LP"; a LP lê ao vivo, então só o fallback (`5.0.0`, `5.1.0-beta`) acompanha a release. Registrar em `sdd/releases/history.md`/checklist na implementação, com o dono.

## Modelo do módulo
**Enxuto** (`ARQ-11`): HTML/JS estático e um arquivo de teste Go; nenhum pacote novo.

## Decisões do projeto aplicadas
`SEG-01` (entrada externa validada na borda), `SEG-05` (nada do usuário na chamada), `SEG-07` (sem dependência), `COD-01`.

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Chamada de rede na LP | `docs/index.html` | Nenhum `fetch`; three.js e fontes externos | **Criar** a única chamada | Pedido do dono |
| Versões na LP | `docs/index.html` ~1133, 1253, 1576, 2085 | `v5.0.0` fixo | **Estender** com `data-versao` | Mesmo lugar, mesmo texto |
| Instalação | `docs/index.html` ~1317 | Um comando global | **Estender** com o beta | Sem seção nova |
| Testes da LP | `internal/scaffold/pipeline_scaffold_test.go` | Leitura de `docs/` como texto; sem teste Node | **Criar** `lp_versoes_test.go` | `t.Skip` sem node |
| Origem das versões | `npm view ... dist-tags` | `latest` 5.0.0, `beta` 5.1.0-beta | **Reusar** como fonte única | Sem arquivo de versões paralelo |

## Dependências novas
Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis
`SEG-01`, `SEG-05`, `SEG-07`, `COD-01`, `TST-01`, `TST-05`. Rules: `sdd-codigo-e-testes`.

## Critérios de Aceitação Executáveis
Testes planejados em `internal/scaffold/lp_versoes_test.go`: `TestLP_VersoesFallbackEstatico` (leitura estática), `TestLP_VersoesChamadaSegura` (leitura estática do bloco), `TestLP_VersoesFuncaoPura` (`node -e` sobre o texto entre os marcadores; `t.Skip` se `node` não existir).
```bash
set -o pipefail
for t in TestLP_VersoesFallbackEstatico TestLP_VersoesChamadaSegura TestLP_VersoesFuncaoPura; do
  go test ./internal/scaffold -list "$t" | grep -F "$t" >/dev/null
done
go test ./internal/scaffold -run 'TestLP_Versoes'
# negativo: fallback estático presente nos dois campos e nos dois comandos de instalação
grep -q 'data-versao="latest"[^>]*>5\.0\.0<' docs/index.html
grep -q 'data-versao="beta"[^>]*>5\.1\.0-beta<' docs/index.html
grep -qF 'npm install -g @nathanramorim/forge-sdd@beta' docs/index.html
# negativo: o bloco das versões não usa innerHTML nem polling
bloco="$(sed -n '/versoes:inicio/,/versoes:fim/p' docs/index.html)"
test -n "$bloco"
! printf '%s' "$bloco" | grep -E 'innerHTML|setInterval|localStorage' || exit 1
printf '%s' "$bloco" | grep -qF "credentials: 'omit'"
printf '%s' "$bloco" | grep -qF "referrerPolicy: 'no-referrer'"
printf '%s' "$bloco" | grep -qF 'AbortController'
# negativo: o endpoint sem CORS não é usado
! grep -F -- '/-/package/' docs/index.html || exit 1
# negativo: o comparativo histórico segue igual
grep -qF 'Agora (v5.0) · 3 Sessões em Pipeline' docs/index.html
go vet ./... && go test ./...
```
- Teste negativo (`TestLP_VersoesFuncaoPura`): entradas `5.0.0<img src=x onerror=1>`, `javascript:alert(1)`, `5.0`, `5.1.0-beta` no campo `latest`, `5.0.0` no campo `beta`, `null` e objeto vazio devolvem o fallback; `5.0.0` e `5.1.0-beta` passam.
- Prova de borda (manual, só leitura): `curl -s -D - -o /dev/null -H 'Origin: https://forge-sdd.org' https://registry.npmjs.org/@nathanramorim/forge-sdd/beta | grep -i access-control-allow-origin` retorna `*`.

## A confirmar no Gate 1
D-4 (terminal de exemplo segue a estável) e D-5 (wiki depois). Se o dono preferir esconder o chip Beta quando o npm não tiver a tag, dizer agora (padrão da spec: manter o valor estático).
