# fix-102-01 — o boneco de erro treme só na entrada e respeita menos movimento

**Branch:** `fix/boneco-de-erro-treme-so-na-entrada`
**Origem:** Iniciativa 102 (LP da visão em pipelines, `docs/index.html`); relato do dono: "o personagem fica tremendo"
**Tamanho:** `S`
**Status:** `spec`

## Problema

Na página do produto, o mascote 3D em estado de erro (chip "Bloqueado" e cenários "Não conforme") balança sem parar, 28 vezes por segundo. A seção de Governança já nasce em erro, então o boneco fica tremendo o tempo todo, o que cansa o olhar e incomoda quem pediu menos movimento ao sistema.

**Decisão do dono (Gate 1, 21/11/2026):** o tremor dura cerca de meio segundo ao ENTRAR no estado de erro, com decaimento rápido; depois o boneco fica parado, mantendo o olho em X e a cor vermelha. O tremor NÃO acontece quando o sistema pede menos movimento (`prefers-reduced-motion: reduce`).

## Evidência

`origin/main` (`79e260a`), `docs/index.html`:

- `:1608` `createForgeMascot3D(mountEl, agentHex, initialState, markType)`.
- `:1758` `let currentState = initialState || 'working';` e `:1759` `let hopTimer = 0;`. Não existe registro de quando o erro começou.
- `:1761-1771` `setState(nextState, nextColorHex)`: troca `currentState`, orbe e olhos (`xGroup.visible` quando `eyes === 'x'`) e `hopTimer = 0.45`.
- `:1780` `triggerHop: () => { hopTimer = 0.55; }` (só o pulo; não mexe no erro).
- `:1782` `update: (t, lookX, lookY)`, `t` = segundos do `THREE.Clock` (`:2403`, `clock.getElapsedTime()`), chamado por `animateAll` (`:2401`) a cada frame.
- `:1807-1811` o defeito:
  `if (currentState === 'error') { botGroup.position.x = Math.sin(t * 28) * 0.025; } else { botGroup.position.x = 0; }`
- Outros estados: só `working` e `searching` têm varredura do olho (`:1791`, `scanX`), `approval`/`working` balançam a antena (`:1801`). Só `error` treme o corpo. O olhar que segue o ponteiro (`:1795-1797`, `lookX`/`lookY`) e o piscar (`:1799`) são independentes do erro.
- Mascotes da página (3 pontos de criação): `mascotControllers` (`:1901`, um por sessão do pipeline, `:1985`, estados do `SESSIONS`/`stateKeys`, sem `error` inicial), `emuMascot` (`:2194`, nasce `'finished'`, `setState(st.agentState, ...)` em `:2227`) e `govMascot` (`:2329`, nasce `'error'`; `setState(sc.state)` em `:2352`; `'finished'` em `:2373`). Os cenários `GOV_SCENARIOS` com `state: 'error'` estão em `:2249` e `:2289`. Os três são atualizados em `animateAll` (`:2407-2409`).
- Reaparição: cada `selectGov` chama `govMascot.setState(sc.state)` (`:2352`), então o erro "reentra" ao trocar de aba de cenário e o tremor de entrada voltaria a rodar (comportamento desejado).
- Wiki: `docs/wiki.html` e `docs/wiki.css` não têm mascote 3D nem animação do mascote (`grep` por `mascot`, `error`, `t * 28`: 0 ocorrências).
- Mesmo código em outros arquivos (fora do escopo, ver "A confirmar"): `site/src/scripts/mascot-3d.ts:230-231` (fonte Astro) e `docs/showcase-v3-lp.html:1797-1798` (cópia estática).
- Como a LP é testada hoje: Go lê o arquivo por texto, sem navegador (`internal/scaffold/pipeline_scaffold_test.go`: `TestScaffold_DocsEdicao`, `:309`, lê `docs/index.html` e valida marcadores `edicao:inicio/fim`). `site/tests/verify-build.mjs` (`npm test` em `site/`) exige `site/dist` pós-build do Astro e a CI (`.github/workflows/ci.yml`) só roda `go test ./... -v` e golangci-lint; não roda Node nem navegador. `docs/` não tem `package.json` nem teste JS.

## Feature Correlacionada

- Origem: `docs/index.html`, função `createForgeMascot3D`, trecho do estado `error` (introduzido com o motor de personagens da LP, reformulado na iniciativa 102).
- Causa-raiz: o balanço do estado de erro foi escrito como função contínua do tempo global (`Math.sin(t * 28)`), sem noção de "quando o erro começou" e sem consultar a preferência de movimento do sistema.
- Iniciativa de origem: 102. Numeração: `fix-102-01` não existe em `origin/main`, em nenhuma branch remota (`git branch -r`, incluindo `origin/feat/102-lp-visao-pipelines`) nem nos PRs abertos (#122, #153, #164). A pasta `fixes/` da 102 ainda não existia; esta spec a cria.

## Decisões do projeto aplicadas

- DUP-02: estender `createForgeMascot3D` existente (nada de novo mascote nem biblioteca).
- P-03: tabela "O que já existe" abaixo.
- P-25: critérios verificáveis, negativos com `# negativo:`, sem `grep -c`.
- Edição Missões é temporária: `docs/edicao/**` e os marcadores `edicao:inicio/fim` do `index.html` não são tocados; `TestScaffold_DocsEdicao` segue verde.
- Não inventar ferramenta: teste em Go (já é o padrão do projeto para ler a LP); Node só como apoio opcional, com `t.Skip` se ausente, porque a CI não instala Node.
- Segurança: nada de segredo nem caminho de usuário.

## Correção

Menor desenho (tudo marcado "a confirmar" é proposta):

1. Constantes nomeadas no escopo de `docs/index.html`, junto de `STATE_DEFS` (a confirmar os valores): `TREMOR_ERRO = { duracao: 0.5, amplitude: 0.025, frequencia: 28 }` (segundos, unidade de cena, Hz de ciclo; amplitude e frequência iguais às de hoje).
2. Função pura, entre os comentários `/* tremor:inicio */` e `/* tremor:fim */` (a confirmar o nome): `deslocamentoTremor(t, desde, reduzMovimento)`. Retorna `0` se `reduzMovimento` for verdadeiro, se `desde` for `null`, ou se `t - desde` estiver fora de `[0, duracao)`. Senão: `Math.sin((t - desde) * frequencia) * amplitude * (1 - (t - desde) / duracao)` (decaimento linear; a fase parte de 0 para não dar salto). Único lugar com o tremor.
3. Em `createForgeMascot3D`: `let errorSince = null;`. Em `setState`, depois de trocar o estado: `errorSince = null;` (reinicia a cada `setState`, inclusive `setState('error')` repetido e a criação inicial). No `update`, no ramo `currentState === 'error'`: `if (errorSince === null) errorSince = t;` (usa o mesmo relógio do `update`, sem acessar o `THREE.Clock`) e `botGroup.position.x = deslocamentoTremor(t, errorSince, reduzMovimento())`. Outros estados continuam `botGroup.position.x = 0`.
4. Menos movimento: `const mqMenosMovimento = window.matchMedia('(prefers-reduced-motion: reduce)');`, uma vez por página (não por mascote); `reduzMovimento = () => mqMenosMovimento.matches` é lido a cada frame, então reage à mudança sem ouvinte extra (a confirmar: basta ler `.matches`; um `addEventListener('change')` seria redundante). Fallback: se `window.matchMedia` não existir, `reduzMovimento` retorna `false`.
5. Não regressão (inalterados): olho em X e orbe vermelho (`STATE_DEFS.error`), pulo (`hopTimer`), bob, piscar, antena, varredura do olho, olhar que segue o ponteiro (`lookX`/`lookY`), os 5 outros estados, `setState`/`triggerHop` (assinatura e retorno), `docs/edicao/**`.
6. Testes (arquivo novo `internal/scaffold/lp_mascote_tremor_test.go`, a confirmar o nome; pacote `scaffold`, mesmo da leitura da LP):
   - `TestLP_MascoteTremorSoNaEntrada` (estático, sempre roda): lê `docs/index.html`; exige que `Math.sin(t * 28)` NÃO exista; que o bloco entre `/* tremor:inicio */` e `/* tremor:fim */` exista, contenha `Math.sin(` e `TREMOR_ERRO`; que `Math.sin(` com `frequencia` só apareça dentro do bloco; que `errorSince = null` apareça dentro de `setState` e `errorSince = t` no `update`; que `prefers-reduced-motion: reduce` apareça.
   - `TestLP_MascoteTremorComportamento` (comportamento, `t.Skip` se `node` não estiver no PATH): extrai o bloco `tremor:inicio..fim` e roda com `node -e` (sem rede, sem navegador). Casos: `t = desde` dá 0; `t = desde + 0.1` e `0.2` dão `|x| > 0` e `<= amplitude`; a envoltória decai (`|x|` máximo na janela [0,0.25) maior que na janela [0.25,0.5)); `t = desde + 0.5`, `+0.6` e `+10` dão exatamente 0; `reduzMovimento = true` dá 0 em qualquer `t`; `desde = null` dá 0.

## Seguimento (fora do escopo, para o dono decidir)

- Fonte Astro `site/src/scripts/mascot-3d.ts:230` e `docs/showcase-v3-lp.html:1797` têm o mesmo tremor contínuo. Esta spec só corrige `docs/index.html` (o que o dono relatou e confirmou).

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Mascote 3D | `docs/index.html:1608` `createForgeMascot3D` | Motor único com `setState`/`update` | Estender | DUP-02 |
| Registro de entrada no estado | `docs/index.html:1758-1771` | Só `currentState` e `hopTimer`; sem `errorSince` | Criar `errorSince` junto de `hopTimer` | Mesmo escopo de closure |
| Preferência de menos movimento | `docs/index.html`, `docs/wiki.css` | Nenhum `prefers-reduced-motion` | Criar uma leitura única | Não há o que reusar |
| Teste da LP | `internal/scaffold/pipeline_scaffold_test.go:309` | Go lê `docs/index.html` como texto | Mesmo padrão em teste novo | Ferramenta já usada |
| Teste JS da LP | `site/tests/verify-build.mjs` | Exige build do Astro; fora da CI | Não usar | CI só roda `go test` |
| Biblioteca de testes JS / navegador | `package.json`, `.github/workflows/ci.yml` | Nenhuma | Não introduzir | Sem dependência nova |

## Dependências novas

Nenhuma. O teste de comportamento usa `node` só se já existir no ambiente (`t.Skip` caso contrário).

## Rules e IDs da Constituição Aplicáveis

DUP-02, GOV-01, SEG-07 (sem segredo), P-03, P-25.

## Descrição de negócio (release notes)

- **O boneco de erro não fica mais tremendo:** na página do produto, o mascote em "Bloqueado"/"Não conforme" treme por cerca de meio segundo ao entrar no erro e depois fica parado, com o olho em X e a cor vermelha. Quem pede menos movimento no sistema não vê o tremor.

## A confirmar

- Valores: duração 0,5 s, amplitude 0,025, frequência 28 (os dois últimos são os de hoje); decaimento linear.
- Nomes: `TREMOR_ERRO`, `deslocamentoTremor`, `errorSince`, arquivo de teste.
- Ler `.matches` a cada frame (proposto) em vez de ouvinte `change`.
- Corrigir também `site/src/scripts/mascot-3d.ts` e `docs/showcase-v3-lp.html` em outra entrega?
- Com menos movimento, o bob e o pulo seguem como hoje (o dono só pediu o tremor).

## Critérios de Aceitação Executáveis

```bash
set -o pipefail
# o tremor contínuo sumiu e a regra mora só na função de decaimento
# negativo: Math.sin(t * 28) não pode existir mais no arquivo
if grep -F 'Math.sin(t * 28)' docs/index.html >/dev/null; then exit 1; fi
grep -F '/* tremor:inicio */' docs/index.html >/dev/null
grep -F '/* tremor:fim */' docs/index.html >/dev/null
grep -F 'TREMOR_ERRO' docs/index.html >/dev/null
grep -F 'errorSince' docs/index.html >/dev/null
# existe a verificação de menos movimento
grep -F 'prefers-reduced-motion: reduce' docs/index.html >/dev/null
# negativo: a amplitude fixa 0.025 não aparece mais fora das constantes (só no TREMOR_ERRO)
test "$(grep -F '* 0.025' docs/index.html | wc -l)" -eq 0
# o olho em X e a cor do erro seguem definidos
grep -F "error:    { label: 'Bloqueado',    orb: 0xf2677a, cssOrb: '#f2677a', eyes: 'x' }" docs/index.html >/dev/null
# negativo: Edição Missões, wiki e demais páginas intactas
test "$(git diff --name-only origin/main -- docs/edicao docs/wiki.html docs/wiki.css docs/showcase-v3-lp.html site | wc -l)" -eq 0
# testes Go: o teste existe antes de rodar (-list) e passa
go test ./internal/scaffold -list 'TestLP_MascoteTremorSoNaEntrada' | grep -F 'TestLP_MascoteTremorSoNaEntrada' >/dev/null
go test ./internal/scaffold -run 'TestLP_MascoteTremorSoNaEntrada' -count=1
go test ./internal/scaffold -list 'TestLP_MascoteTremorComportamento' | grep -F 'TestLP_MascoteTremorComportamento' >/dev/null
go test ./internal/scaffold -run 'TestLP_MascoteTremorComportamento' -count=1 -v
# a Edição Missões segue válida
go test ./internal/scaffold -list 'TestScaffold_DocsEdicao' | grep -F 'TestScaffold_DocsEdicao' >/dev/null
go test ./internal/scaffold -run 'TestScaffold_DocsEdicao' -count=1
# o gate da CI continua verde
go build ./... && go test ./... -count=1
```

Verificação visual (manual, a cargo do dono): abrir `docs/index.html`, ir à seção de Governança, ver o boneco tremer ~0,5 s ao entrar no erro e parar; trocar de aba de cenário e ver o tremor de entrada de novo; ligar "Reduzir movimento" no sistema e confirmar que não treme.
