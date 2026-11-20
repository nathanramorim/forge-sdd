# feat-102-03 — Governança, comparativo e vídeo na visão de pipelines

**Iniciativa:** 102-lp-visao-pipelines
**Status:** ready
**Tamanho:** P
**Depende de:** feat-102-01

## Objetivo

Fechar a troca de visão nas seções 4 a 6: quem audita é o Reviewer (não o "Analista de Governança" de uma estação), o painel final é o seu (não o "Painel do Líder"), o comparativo mostra as 3 sessões e o texto do vídeo deixa de falar em 5 estações.

## Modelo do módulo

**Enxuto** (ARQ-11): só `docs/index.html`.

## O que muda

| Onde | Antes | Depois |
|---|---|---|
| Governança, abas e inspetor | "Analista de Governança", "Auditor", "Retorna 05 ↺ 04", "Avança 05 → 01", "Painel do Líder" | "Reviewer", "Retorna ao Builder", "Avança para você", "seu painel (`forge status`)" |
| Governança, cenários (`GOV_SCENARIOS`) | "retorna à Estação 04", "(05 ↺ 04)", "Painel Executivo do Líder" | "volta ao Builder com comentários", "liberado para a sua aprovação" |
| Comparativo, lado "Agora" | "5 Estações em Esteira", "Handoff explícito entre 5 especialistas" e a cadeia Líder → PM/Arquiteto → Planejador → Builders → Governança | "3 sessões em pipeline": Discovery planeja com você, Builder constrói em paralelo e abre o PR, Reviewer confere em modo somente leitura e devolve até 3 vezes |
| Vídeo, texto | "esteira autônoma em 5 estações" | o pipeline de 3 sessões; o vídeo em si não muda |

As 3 camadas (portão automático, revisão somente leitura e painel) continuam, porque ainda valem.

## Fora do escopo

O arquivo de vídeo, as seções 1 a 3, `docs/edicao/`.

## Decisões do projeto aplicadas

- **Discovery 102** e critérios 1, 4 e 5 da iniciativa.
- Fonte do texto: `README.md` (Reviewer somente leitura, até 3 devoluções, aviso na Discovery) e `docs/pipeline.md`.
- **FLX-02:** branch `feat/102-lp-visao-pipelines`.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Simulador de governança | `docs/index.html`: abas `.gov-tab`, `GOV_SCENARIOS`, inspetor | Texto preso às estações 04 e 05 e ao Líder | estender: trocar o texto, sem mudar a mecânica |
| Comparativo | `docs/index.html`, `#evolucao` | Lado "Agora" com 5 estações | estender |
| Vídeo | `docs/index.html`, `#video` | Texto citando 5 estações | estender só o texto |

## Dependências novas

Nenhuma.

## Rules e IDs da Constituição Aplicáveis

- **ARQ-11**, **DUP-02**, **COD-01**, **COD-04**, **COD-05**, **SEG-07**, **TST-01**, **TST-05**, **FLX-01**, **FLX-02**.
- Rules: `.agents/rules/sdd-reuso-e-duplicacao.md`, `.agents/rules/sdd-codigo-e-testes.md`, `.agents/rules/sdd-fluxo-e-portao.md`.

## Critérios de Aceitação Executáveis

```bash
# 1. Seções 4 a 6 e os cenários sem a visão antiga
sec=$(awk '/<!-- SEÇÃO 4/,/<\/video>/' docs/index.html)
test "$(printf '%s' "$sec" | grep -c -i -E 'Estação|5 estações|5 especialistas|Analista de Governança|Líder|Planejador|05 ↺ 04|05 → 01')" = "0"
test "$(awk '/const GOV_SCENARIOS = \[/,/^\];/' docs/index.html | grep -c -i -E 'Estação|Líder|05 ↺ 04|05 → 01|Analista')" = "0"
printf '%s' "$sec" | grep -q 'Reviewer'

# 2. A página inteira não vende mais as 5 estações
test "$(grep -c -i -E '5 estações|5 especialistas|O Líder \(Orquestrador\)|Fila de Builders' docs/index.html)" = "0"

# 3. Ids preservados, camada da edição intacta e scripts válidos
for id in governanca evolucao video govTabs govInspector govTerminal; do grep -q "id=\"$id\"" docs/index.html || { echo "falta id $id"; exit 1; }; done
git diff --quiet origin/main -- docs/edicao
node -e 'const h=require("fs").readFileSync("docs/index.html","utf8"); for (const m of h.matchAll(/<script>([\s\S]*?)<\/script>/g)) new Function(m[1]); console.log("ok");'

# 4. Portão do projeto
go test ./... && go vet ./...
```

## Conferência visual (Gate 2)

Simulador de governança navegável nos 4 cenários; comparativo e texto do vídeo legíveis. Texto revisado pelo dono.

## Status

- [x] Implementado em feat-102-03 (19/11/2026): critérios 1 a 4 com saída 0.
