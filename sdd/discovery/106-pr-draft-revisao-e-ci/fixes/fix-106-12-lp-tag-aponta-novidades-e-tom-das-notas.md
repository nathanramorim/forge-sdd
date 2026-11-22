# fix-106-12 — A tag de versão da LP leva às novidades, e as notas de release ganham emoção

**Branch:** `fix/106-12-lp-tag-e-tom-das-notas`
**Origem:** Iniciativa 106, revisão da LP local pelo dono em 22/11/2026: "clicou na tag lá em cima, vai direto para o que mudou em cada versão; dá para dar mais emoção à descrição do release notes; deixa nas rules para não esquecer nas próximas".
**Tamanho:** `S`
**Status:** `spec`

## Sintoma

1. No topo da LP, o chip `Beta 5.2.0-beta` (`docs/index.html:1183`, `data-versao="beta"`) é só texto, e fica **dentro** de `<a href="#" class="brand">` (linha 1180): clicar não leva a lugar nenhum.
2. `#novidades` (`docs/index.html:1628`) é uma tela de **2 abas** (`.nv-tab`, `data-nv=stable|beta`) com **uma lista** (`#nvLista`) que o JS reescreve ao trocar de aba (`nvRenderLista`); não existe cartão por versão.
3. As notas de `sdd/releases/history.md` descrevem o que foi feito em tom de changelog (identificadores de PR e fix **dentro do título em negrito**, crases, nomes de teste) e pouco o que a pessoa ganha.

## Feature Correlacionada

- **Arquivo de origem:** `docs/index.html` (chip e seção de novidades), `scripts/novidades_lp.py` (gera `docs/data/releases.json`, chave real `versoes`, 2 entradas) e a regra de release em `.agents/rules/sdd-fluxo-e-portao.md` (bullet "Versões `-beta` e `-rc`", linha 40).
- **Causa-raiz:** o chip foi feito como selo informativo, sem destino; e a regra de release exige "destaques reais" sem dizer **como escrevê-los**.
- **Evidência:** `docs/index.html:1180-1183`, `1636-1658`, `2593-2782`; notas da 5.2.0 em `history.md` linhas 11-21.

## Modelo do módulo

**Enxuto.** HTML/CSS/JS da LP, uma regra com o espelho (template e golden) e testes. Sem Go novo além dos testes.

## Dependências novas

Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis

`FLX-07` (release), `DUP-01` (a regra mora só em `sdd-fluxo-e-portao.md`; template e golden são o espelho), `SEG-07`. Não altera a Constituição, então não pede ADR.

## Decisões do projeto aplicadas

- Regra do dono: a nota mostra o ganho da pessoa, com emoção sem exagero, sem marca de terceiros e sem promessa que o código não cumpre.
- `docs/data/releases.json` é gerado: edita-se `history.md` e roda-se `python3 scripts/novidades_lp.py`. **Não há campo `ancora`** (cortado: o destino é fixo).
- O link do chip não fixa número de versão: o texto visível é o que o JS já preenche em `data-versao`.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Chip da versão | `docs/index.html:1180-1183` | `<span class="chip">` dentro de `<a class="brand">` | **mover** o chip para fora do `<a class="brand">` (irmão) e trocá-lo por `<a href="#novidades">`, evitando `<a>` aninhado |
| Abas das novidades | `nvAtivarAbas`, `nvRenderLista(versoes,'beta')` | seleção de aba por clique | **reusar**: o chip ativa a aba da sua versão (`beta`) |
| Testes da LP | `lp_novidades_test.go`, `lp_versoes_test.go` | exigem `data-versao="beta"` com o texto `5.2.0-beta` | **manter** o atributo e o texto dentro do novo `<a>` |
| Regra de release | `sdd-fluxo-e-portao.md:40` | exige destaques reais | **estender** o mesmo bullet (sem bloco novo) |
| Espelho da regra | `internal/scaffold/templates/.agents/rules/sdd-fluxo-e-portao.md.tmpl` e `testdata/golden/.agents/rules/sdd-fluxo-e-portao.md` | cópias da regra | **atualizar** os dois, regenerando com `-update -run TestGoldenInit` |

## Comportamento esperado

1. **Chip clicável:** o chip Beta deixa o `<a class="brand">` e vira `<a class="chip" href="#novidades" data-nv-alvo="beta">` com o mesmo texto visível e `title="Ver o que mudou nesta versão"`. Sem `aria-label` com número (o nome acessível é o texto visível). Regra CSS `.chip[href]:focus-visible{outline:2px solid var(--accent);outline-offset:2px}`; o ajuste do CSS que hoje oculta `.chip` no brand fica coberto.
2. **Destino:** o clique rola suavemente até `#novidades` e, por JS (dentro do bloco `novidades:inicio/fim`, depois de `nvAtivarAbas`), ativa a aba `beta` e renderiza a lista. Sem JavaScript, cai em `#novidades` com a aba estável ativa.
3. O chip estável (se existir) leva à aba `stable` do mesmo jeito.
4. **Regra de redação** no bullet "Versões `-beta` e `-rc`" de `sdd-fluxo-e-portao.md` (+ template e golden), em até 4 linhas: o **título em negrito** de cada item diz o ganho da pessoa em linguagem direta ("Agora você…", "Acabou…"), sem identificador de PR/fix, arquivo ou flag; o detalhe técnico vai no fim do texto, entre parênteses, fora do negrito; sem exagero, sem marca de terceiros.
5. As notas da **5.2.0** são reescritas nesse tom (mesmo conteúdo, nova redação); versões anteriores ficam como estão. Rodar `python3 scripts/novidades_lp.py` e versionar o `releases.json` (muda `chave_conteudo`).

## Critérios de Aceitação Executáveis

Os de testes novos ficam **vermelhos até o Builder** criar os testes e a edição (esperado).

- **Chip:** `go test ./internal/scaffold -run 'TestLPChipVersaoLeva' -count=1` (teste novo): o HTML tem `<a ... href="#novidades" ...>` contendo `data-versao="beta"`, e **nenhum** `<a>` dentro de `<a class="brand">`.
- **Regra e notas:** `go test ./internal/scaffold -run 'TestRegraRedacaoRelease|TestNotas520SemTecnicoNoTitulo' -count=1`: a regra existe na rule, no template e no golden idênticos; nenhum **título em negrito** da seção 5.2.0 de `history.md` contém `(#`, `fix-`, `.go`, `.md`, `.py`, `.sh` ou `--`.
- **Existência dos testes:** `[ "$(go test ./internal/scaffold -list 'TestLPChipVersaoLeva|TestRegraRedacaoRelease|TestNotas520SemTecnicoNoTitulo' | grep -c '^Test')" -eq 3 ]` (só confirma existência; a aprovação é a execução acima).
- **Releases:** `python3 -I scripts/novidades_lp.py --check --quiet` sai com 0 depois de regerar.
- **Orçamento e golden:** `go test ./internal/scaffold -run 'ContextBudget|Golden' -count=1` passa (a regra cabe no orçamento de leitura e o golden segue idêntico ao template).

## Fora do escopo

- Reescrever notas de versões antigas.
- Campo `ancora` por versão e rolagem a item específico.
- O deploy da LP (token da Cloudflare, `pendencias-humanas.md`).
