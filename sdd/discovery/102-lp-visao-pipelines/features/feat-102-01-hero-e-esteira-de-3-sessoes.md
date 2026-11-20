# feat-102-01 — Título, hero e esteira de 3 sessões

**Iniciativa:** 102-lp-visao-pipelines
**Status:** ready
**Tamanho:** M

## Objetivo

Trocar a visão de 5 estações pela de 3 sessões de pipeline no topo da LP: título da página, hero e a esteira interativa (seção 1), com os mascotes 3D das 3 sessões e o fechamento "você aprova".

## Modelo do módulo

**Enxuto** (ARQ-11): só `docs/index.html`; nenhuma dependência e nenhum Go.

## O que muda

| Onde | Antes | Depois |
|---|---|---|
| `<title>` | "Orquestração Autônoma em 5 Estações & Governança IA" | "Pipeline de 3 sessões: Discovery, Builder e Reviewer" |
| Hero | "5 especialistas em esteira" | "3 sessões que planejam, constroem e conferem"; você conversa e aprova, o Builder constrói em paralelo, o Reviewer confere |
| Menu | "Fluxo Processual" | "Pipeline" (mesma âncora `#estacoes`) |
| Seção 1 | 5 cartões e 5 etapas | 3 cartões e 3 etapas (Discovery, Builder, Reviewer), com o fechamento "você aprova" |
| Dados | `STATIONS` com 5 objetos | `SESSIONS` com 3 objetos; o rótulo da última etapa vira "↺ 02" (o Reviewer devolve ao Builder) |

Texto de cada sessão (todo extraído do README e de `docs/pipeline.md`):

| Sessão | O que faz | Recebe | Entrega | Portão |
|---|---|---|---|---|
| Discovery | Você conversa e planeja; o agente faz o Reuse Check, escreve a spec e os critérios | Sua ideia e o contexto do projeto (`forge brief`) | Spec, critérios e tasks na fila | FLX-01: sua aprovação da spec |
| Builder | Pega o job da fila e constrói em worktrees isolados, em paralelo quando não há conflito | Job da fila, spec e rules do escopo | Código, testes e PR aberto | Camada A: `go test`, `go vet` e `forge gate` |
| Reviewer | Confere o PR contra spec, ADRs e rules, sem poder editar código; devolve com comentários, até 3 vezes | PR, spec, ADRs e rules | Parecer: conforme ou devolvido | Camada B; conforme, você aprova e faz o merge |

## Fora do escopo

Emulador, guia de comandos, governança, comparativo e vídeo (fases 2 e 3); `docs/edicao/`; ids das seções.

## Decisões do projeto aplicadas

- **Discovery 102** e a decisão do dono de 19/11/2026: a LP mostra só o pipeline.
- **Critérios 1, 2, 4 e 5** da iniciativa.
- **FLX-02:** branch única da iniciativa, `feat/102-lp-visao-pipelines`.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Motor da esteira | `docs/index.html`: `STATIONS`, stepper, cartões, barra de handoff | Guiado por dados, com ids fixos | estender: trocar os dados e o que está fixo em 5 |
| Mascotes 3D | `createForgeMascot3D` | 5 marcas, uma por estação | reusar com 3 marcas (`diamond`, `wave`, `tri`) |
| Descrição do pipeline | `README.md`, `docs/pipeline.md` | 3 sessões, papéis e o ciclo de revisão | reusar como fonte do texto |
| Camada da Edição Missões | `docs/edicao/` | Mapeia cartões por ordem | não tocar; ajuste fica na 101 |

## Dependências novas

Nenhuma.

## Rules e IDs da Constituição Aplicáveis

- **ARQ-11**, **DUP-02**, **COD-01**, **COD-04**, **COD-05**, **SEG-07**, **TST-01**, **TST-05**, **FLX-01**, **FLX-02**.
- Rules: `.agents/rules/sdd-reuso-e-duplicacao.md`, `.agents/rules/sdd-codigo-e-testes.md`, `.agents/rules/sdd-fluxo-e-portao.md`.

## Critérios de Aceitação Executáveis

```bash
# 1. Topo sem a visão antiga
test "$(sed -n '1,1200p' docs/index.html | grep -c -i -E '5 estações|5 especialistas|O Líder|Planejador|PM &amp; Arquiteto|PM & Arquiteto')" = "0"
grep -q 'Pipeline de 3 sessões' docs/index.html

# 2. Dados e esteira com 3 sessões
test "$(grep -c 'const STATIONS' docs/index.html)" = "0"
grep -q 'const SESSIONS' docs/index.html
test "$(awk '/const SESSIONS = \[/,/^\];/' docs/index.html | grep -c "step: 'Sessão 0")" = "3"
for s in Discovery Builder Reviewer; do awk '/const SESSIONS = \[/,/^\];/' docs/index.html | grep -q "name: '$s'" || { echo "falta $s"; exit 1; }; done

# 3. Âncoras e ids preservados
for id in estacoes emulador instalacao governanca evolucao video stepperTrack agentsGrid handoffBar; do grep -q "id=\"$id\"" docs/index.html || { echo "falta id $id"; exit 1; }; done

# 4. A camada da edição não foi tocada
git diff --quiet origin/main -- docs/edicao

# 5. O script da página continua sintaticamente válido
node -e '
const fs = require("fs");
const html = fs.readFileSync("docs/index.html", "utf8");
const scripts = [...html.matchAll(/<script>([\s\S]*?)<\/script>/g)].map(m => m[1]);
if (!scripts.length) { console.error("sem script inline"); process.exit(1); }
for (const s of scripts) new Function(s);
console.log("ok");'

# 6. Portão do projeto
go test ./... && go vet ./...
```

## Conferência visual (Gate 2)

LP no navegador: hero, esteira com 3 etapas navegável (setas, cliques e ciclo automático) e 3 cartões com mascote. Texto revisado pelo dono.

## Status

- [x] Implementado em feat-102-01 (19/11/2026): critérios 1 a 6 com saída 0 e `forge gate` em bloqueio sem pendências.
