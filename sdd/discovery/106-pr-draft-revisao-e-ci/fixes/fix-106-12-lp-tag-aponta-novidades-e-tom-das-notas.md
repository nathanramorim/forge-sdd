# fix-106-12 — A tag de versão da LP leva às novidades, e as notas de release ganham emoção

**Branch:** `fix/106-12-lp-tag-e-tom-das-notas`
**Origem:** Iniciativa 106, revisão da LP local pelo dono em 22/11/2026: "clicou na tag lá em cima, vai direto para o que mudou em cada versão; dá para dar mais emoção à descrição do release notes; deixa nas rules para não esquecer nas próximas".
**Tamanho:** `S`
**Status:** `spec`

## Sintoma

1. No topo da LP, o chip `Beta 5.2.0-beta` (`docs/index.html:1183`, `data-versao="beta"`) é só texto: clicar não faz nada. O que mudou em cada versão está na seção `#novidades` (`docs/index.html:1628`), longe do topo.
2. As notas de `sdd/releases/history.md` descrevem **o que foi feito** em tom técnico (nomes de arquivo, flags, números de PR) e pouco **o que a pessoa ganha**. A LP as lê e as publica como estão.

## Feature Correlacionada

- **Arquivo de origem:** `docs/index.html` (chip do hero), `scripts/novidades_lp.py` (gera `docs/data/releases.json` do `history.md`) e a regra de release em `.agents/rules/sdd-fluxo-e-portao.md` (bloco "Fim de Onda e Release Beta").
- **Causa-raiz:** o chip foi feito como selo informativo, sem destino; e a regra de release exige "destaques reais", mas não diz **como escrevê-los**, então cada nota sai no tom de changelog.
- **Evidência:** `docs/index.html:1183`; notas da 5.2.0 em `history.md` (itens como "`(#181)`" e "`TestNovidades…`").

## Modelo do módulo

**Enxuto.** HTML da LP, uma regra e seu espelho em template e golden. Sem Go novo além dos testes.

## Dependências novas

Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis

`GOV-01` (regra nova de redação de release: sem ADR, é detalhamento de regra existente), `FLX-07`, `DUP-01` (a regra mora em **um** lugar: `sdd-fluxo-e-portao.md`), `SEG-08` (a nota não expõe segredo nem caminho de máquina).

## Decisões do projeto aplicadas

- Regra do dono: a nota de release mostra o ganho da pessoa, com emoção sem exagero nem marca de terceiros (memória do lançamento v5).
- `docs/data/releases.json` é gerado: edita-se o `history.md` e roda-se `python3 scripts/novidades_lp.py`.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Seção das novidades | `docs/index.html` `#novidades`, `#novidadesTitulo` | existe, sem âncora por versão | **estender**: cada versão ganha `id="v-<versão>"` |
| Chip da versão | `docs/index.html:1183` | `<span class="chip">` | **trocar** por `<a href="#novidades">` (a versão da beta leva à âncora da própria versão) |
| Regra de release | `sdd-fluxo-e-portao.md` ("Fim de Onda e Release Beta") | exige destaques reais | **estender** com a redação |

## Comportamento esperado

1. O chip do topo vira link: clicar leva a `#novidades` rolando suavemente até a versão indicada (`#v-5-2-0-beta`, ponto trocado por hífen); sem JavaScript funciona como âncora simples. Foco visível e `aria-label` "Ver o que mudou na versão 5.2.0-beta".
2. Cada cartão de versão em `#novidades` recebe o `id` correspondente, gerado por `novidades_lp.py` no `releases.json` (campo `ancora`).
3. **Regra de redação das notas** em `.agents/rules/sdd-fluxo-e-portao.md` (+ template e golden), em até 4 linhas: cada item abre com o **ganho da pessoa** em linguagem direta ("Agora você…", "Acabou…"), um verbo de ação e, quando couber, o antes e depois; o detalhe técnico (arquivo, flag, PR) vai por último, entre parênteses; sem exagero, sem marca de terceiros, sem promessa que o código não cumpre.
4. As notas da **5.2.0** são reescritas nesse tom (mesmo conteúdo, nova redação); as versões anteriores ficam como estão (histórico).

## Critérios de Aceitação Executáveis

- `[ "$(grep -c 'href="#novidades' docs/index.html)" -ge 1 ]` e o chip com `data-versao="beta"` está dentro de um `<a>`: `python3 scripts/novidades_lp.py --check` e o teste do chip passam.
- `[ "$(go test ./internal/scaffold -list 'LPChipVersao|RegraRedacaoRelease' | grep -c '^Test')" -ge 2 ]`.
- `python3 scripts/novidades_lp.py --check` sai com 0 e o `releases.json` traz o campo `ancora` em toda versão (`python3 -c "import json;d=json.load(open('docs/data/releases.json'));assert all('ancora' in v for v in d['versoes'])"`, ajustando a chave ao formato real do arquivo).
- A regra existe em `.agents/rules/sdd-fluxo-e-portao.md`, no template e no golden, e `go test ./internal/scaffold` passa (P-24).
- Nenhum item da 5.2.0 em `history.md` começa por nome de arquivo, flag ou número de PR (verificação por teste com expressão simples sobre o início de cada item).

## Fora do escopo

- Reescrever as notas de versões antigas.
- O deploy da LP (token da Cloudflare, `pendencias-humanas.md`).
