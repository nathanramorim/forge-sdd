# feat-102-02 — Emulador e guia de comandos com o fluxo do pipeline

**Iniciativa:** 102-lp-visao-pipelines
**Status:** ready
**Tamanho:** M
**Depende de:** feat-102-01

## Objetivo

O emulador (seção 2) e o guia de comandos (seção 3) ainda contam o fluxo das 5 estações, e o passo 1 diz que o `forge init` instala "os papéis das 5 estações", pasta que não existe mais. Esta task troca os dois pelo fluxo real do pipeline.

## Modelo do módulo

**Enxuto** (ARQ-11): só `docs/index.html`.

## O que muda

**Emulador**, 5 passos (as abas e o contador continuam com 5):

| Passo | Aba | Sessão | O que mostra |
|---|---|---|---|
| 1 | `$ forge init` | Instalação | Regras em `.agents/rules/`, skills em `.agents/skills/sdd-*` e a Constituição por IDs (o que o scaffold gera hoje) |
| 2 | `$ forge up` | Pipeline | `forge pipeline init` e `forge up` abrindo `forge-discovery`, `forge-builder` e `forge-reviewer` (estes dois "aguardando jobs na fila") |
| 3 | `/discovery` | Discovery | Reuse Check, spec e critérios, `/split-features` em 3 tarefas, Portão FLX-01 e `/implementar` enviando o job (`forge dispatch`) |
| 4 | Builder | Builder | Tarefas 01 e 02 em paralelo em worktrees, a 03 depois, `forge gate` verde e o PR aberto |
| 5 | Reviewer | Reviewer | Conferência do PR contra spec, ADRs e rules, parecer "conforme", aviso na Discovery e a sua aprovação; o retorno ao Builder (até 3 vezes) é citado |

**Guia de comandos:** cada comando indica a sessão em que roda (Discovery, Builder, Reviewer ou qualquer sessão), no lugar de "Estação NN". `/proxima-feature` dá lugar a `/implementar <feature>` (README). O texto "Atualizar para v3.0" vira atualização de projetos existentes, sem versão antiga.

## Fora do escopo

Seções 4 a 6 (feat-102-03), seção 1, `docs/edicao/`, ids das seções.

## Decisões do projeto aplicadas

- **Discovery 102** e critérios 1, 3, 4 e 5 da iniciativa.
- Fonte do texto: `README.md` (tabela de comandos e ciclo de revisão) e `docs/pipeline.md` (sessões, `forge up`, `forge pipeline init`).
- **FLX-02:** branch `feat/102-lp-visao-pipelines`.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Motor do emulador | `docs/index.html`: `EMU_STEPS`, `renderEmuStep`, abas `.emu-cmd-btn` | Guiado por dados, 5 passos | estender: trocar os dados e os rótulos das abas |
| O que o `forge init` gera | `internal/scaffold/templates/` | `.agents/rules`, `.agents/skills/sdd-*`, `sdd/`; sem `.agents/roles` | reusar como fonte do passo 1 |
| Abertura das sessões | `docs/pipeline.md` | `forge pipeline init`, `forge up`, nomes das 3 sessões | reusar |
| Tabela de comandos | `README.md`, seção "Como usar" | `/discovery`, `/split-features`, `/implementar`, `/revisar`, `/status`, `/tutorial` | reusar |

## Dependências novas

Nenhuma.

## Rules e IDs da Constituição Aplicáveis

- **ARQ-11**, **DUP-02**, **COD-01**, **COD-04**, **COD-05**, **SEG-07**, **TST-01**, **TST-05**, **FLX-01**, **FLX-02**.
- Rules: `.agents/rules/sdd-reuso-e-duplicacao.md`, `.agents/rules/sdd-codigo-e-testes.md`, `.agents/rules/sdd-fluxo-e-portao.md`.

## Critérios de Aceitação Executáveis

```bash
# 1. Emulador sem a visão antiga e com as 3 sessões
test "$(awk '/const EMU_STEPS = \[/,/^\];/' docs/index.html | grep -c -i -E 'Estação|Líder|Planejador|PM &amp; Arquiteto|5 estações|agents/roles')" = "0"
for s in "Sessão 01 · Discovery" "Sessão 02 · Builder" "Sessão 03 · Reviewer" "forge up"; do awk '/const EMU_STEPS = \[/,/^\];/' docs/index.html | grep -q "$s" || { echo "falta: $s"; exit 1; }; done
test "$(awk '/const EMU_STEPS = \[/,/^\];/' docs/index.html | grep -c 'stationLabel:')" = "5"

# 2. Seções 2 e 3 sem "Estação", sem /proxima-feature e sem "v3.0"
sec=$(awk '/<section id="emulador">/,/<!-- SEÇÃO 4/' docs/index.html)
test "$(printf '%s' "$sec" | grep -c -i -E 'Estação|proxima-feature|v3\.0|Painel do Líder')" = "0"
printf '%s' "$sec" | grep -q '/implementar'

# 3. Ids preservados, camada da edição intacta e scripts válidos
for id in emulador instalacao emuLog emuTabs emuFileList emuNextStepBtn; do grep -q "id=\"$id\"" docs/index.html || { echo "falta id $id"; exit 1; }; done
git diff --quiet origin/main -- docs/edicao
node -e 'const h=require("fs").readFileSync("docs/index.html","utf8"); for (const m of h.matchAll(/<script>([\s\S]*?)<\/script>/g)) new Function(m[1]); console.log("ok");'

# 4. Portão do projeto
go test ./... && go vet ./...
```

## Conferência visual (Gate 2)

Os 5 passos do emulador navegáveis pelas abas, pelo botão de ação e por "Executar próximo passo"; guia de comandos legível. Texto revisado pelo dono.

## Status

- [x] Implementado em feat-102-02 (19/11/2026): critérios 1 a 4 com saída 0.
