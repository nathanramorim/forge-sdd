# fix-104-14 — README conta a esteira duas vezes

**Branch:** `fix/104-readme-esteira-duplicada`
**Origem:** Iniciativa 104 (PR #118, `3c8a8d5`, `docs/readme-esteira-agentes`)
**Tamanho:** `S`
**Status:** `spec`

## Problema

O `README.md` tem duas seções que explicam a esteira: `## Como funciona` (linha 36, com tabela de 3 sessões e diagrama Mermaid) e `## Como a esteira funciona` (linha 77, com o SVG `docs/img/esteira-3-sessoes.svg` e a lista de passos). O dono pediu manter a estrutura mais recente (a segunda, mais `## Quem faz o quê` e `## O que é obsoleto (não use)`) e remover a primeira, sem duplicar informação e sem perder o que só ela diz.

## Evidência

- `grep -n '^## ' README.md` na `main` (`8bd3a0a`): 30 `O que é`, 36 `Como funciona`, 77 `Como a esteira funciona`, 90 `Quem faz o quê`, 108 `O que é obsoleto (não use)`, 128 `Por que usar`.
- `git log -S'## Como a esteira funciona' -- README.md`: só o `3c8a8d5` (#118). `git log -S'## Como funciona' -- README.md`: `f389f5d` (#79, README comercial), anterior.

## Feature Correlacionada

- Origem: `README.md`, seção `## Como a esteira funciona`, adicionada pelo `3c8a8d5` (#118) sem remover a `## Como funciona` do `f389f5d` (#79).
- Causa-raiz: o #118 acrescentou a explicação nova ao lado da antiga em vez de substituí-la; nenhum critério checava título duplicado de seção.
- Iniciativa de origem: 104. O #118 saiu dentro dela (a spec aponta a 104 em `sdd/discovery/104-pipeline-em-subagentes/`; o texto novo descreve o modo `chat` em subagentes). A 101 (tema) só tocou o cabeçalho entre `edicao:inicio` e `edicao:fim`, que não muda aqui.

## Decisões do projeto aplicadas

- DUP-01 e DUP-02: uma única explicação da esteira; o que só existia na seção antiga migra para a nova, sem repetir.
- ADR-0004 (3 sessões) e ADR-0010 (comando `forge`): o texto mantido já os respeita.
- Marcadores `<!-- edicao:inicio -->` (linha 1) e `<!-- edicao:fim -->` (linha 19) ficam intactos: `.github/scripts/encerrar-edicao/encerrar.sh` e `TestScaffold_DocsEdicao` (`internal/scaffold/pipeline_scaffold_test.go`) dependem deles.
- Decisão do dono: manter o `sequenceDiagram` do ciclo da feature em `## Como usar`, corrigir a legenda abaixo dele (sem citar o flowchart removido) e relaxar o critério de "zero blocos mermaid" para "o flowchart antigo não volta".
- Fora de escopo: `npm/README.md` tem só um `## Como funciona` (linha 36, sem a seção nova); não há duplicata ali, então não muda.
- Segurança: nada de segredo nem caminho de usuário.

## Correção

Tabela de decisão por informação (A = seção antiga, N = nova):

| Informação | Só na antiga | Só na nova | Nas duas | Decisão |
|---|---|---|---|---|
| 3 sessões Discovery, Builder, Reviewer | | | sim | Manter só na nova (bullets) |
| Reviewer devolve até 3 vezes | | detalhe `max_rework_attempts` | sim | Manter a nova |
| Fluxo em diagrama | Mermaid | SVG `esteira-3-sessoes.svg` | sim | Descartar o Mermaid (o SVG cobre o mesmo fluxo) |
| "Quando o PR está conforme, você é avisado na sessão de Discovery e faz a aprovação final" | sim | | | Mover: ao bullet "Você aprova o PR" |
| Builder "constrói em paralelo" e "worktree isolado" | sim | | | Mover: ao bullet "O Builder" |
| Reviewer "somente leitura" | | sim (parecer JSON que não edita) | sim | Manter a nova |
| Agentes por sessão (Claude, Gemini, Copilot) | rótulo no Mermaid | | | Descartar: já em "Por que usar" (linhas 134 e 140) e no selo |
| Gate 1 humano, `forge gate`, parecer JSON | | sim | | Manter |
| Modos `sessions` e `chat`, `forge pipeline init`, `forge up`, ADR-0015 | | sim | | Manter |
| Quem faz o quê; o que é obsoleto | | sim | | Manter sem mudança |

Passos:

1. Remover do `README.md` as linhas de `## Como funciona` até a linha em branco antes de `## Como a esteira funciona` (tabela, frase de aviso e bloco Mermaid).
2. Na seção nova, ajustar o bullet do Builder para "implementa (em paralelo, em worktree isolado) e abre o PR" e o do PR para "Quando o PR está conforme, você é avisado na sessão de Discovery e faz a aprovação final e o merge".
3. Não tocar nas linhas 1 a 19 (bloco da Edição), no `npm/README.md` nem em outro arquivo.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Seção da esteira | `README.md` | Duas (linhas 36 e 77) | Manter a 77, remover a 36 | Pedido do dono; a nova é a estrutura mais recente |
| Diagrama da esteira | `docs/img/esteira-3-sessoes.svg` | Existe, do #118 | Reusar | Substitui o Mermaid sem perder fluxo |
| Links para `#como-funciona` | `grep -rIn 'como-funciona'` em `*.md`, `*.html`, `*.go`, `*.yml`, `*.sh` | Nenhum | Nada a corrigir | Sem link quebrado |
| Texto "Como funciona" fora do README | `npm/README.md:36`, `docs/video-pipeline/video.html:55`, `site/src/components/SeoHead.astro`, specs 101 | Outros contextos, sem dependência do README | Não mexer | Não duplicam a esteira do README |
| Testes que leem o README | `internal/scaffold/pipeline_scaffold_test.go` (`TestScaffold_DocsEdicao`, `validarMarcadoresEdicao`) | Só conferem os marcadores da edição | Manter o teste verde | Os marcadores não mudam |
| Scripts da edição | `.github/scripts/encerrar-edicao/encerrar.sh:57` | Troca só o cabeçalho entre marcadores | Não tocar | Fora do trecho removido |
| Template ou golden que copie o README | `internal/scaffold` | Nenhum (só `templates/.agents/rules/README.md.tmpl`, outro arquivo) | Nada a fazer | Sem golden vermelho |

## Dependências novas

Nenhuma.

## Rules e IDs da Constituição Aplicáveis

DUP-01, DUP-02, GOV-01, SEG-07 (sem segredo), P-03, P-21, P-25.

## Descrição de negócio (release notes)

- **README sem repetição:** a explicação da esteira aparece uma vez só, com o diagrama das 3 sessões, quem faz o quê e o que é obsoleto.

## Critérios de Aceitação Executáveis

```bash
set -o pipefail
# a seção antiga some e a nova fica uma vez só
test "$(grep -n '^## Como funciona$' README.md | wc -l)" -eq 0
test "$(grep -n '^## Como a esteira funciona$' README.md | wc -l)" -eq 1
# negativo: o flowchart antigo e a tabela de sessões não voltam
test "$(grep -n 'flowchart LR' README.md | wc -l)" -eq 0
test "$(grep -n 'classDef pessoa' README.md | wc -l)" -eq 0
# o sequenceDiagram do ciclo da feature (em "Como usar") continua, uma vez só
test "$(grep -n '^sequenceDiagram' README.md | wc -l)" -eq 1
test "$(grep -nF 'É onde você conversa e planeja a ideia' README.md | wc -l)" -eq 0
# o que só existia na antiga foi movido para a nova
grep -n 'avisado na sessão de Discovery' README.md
grep -n 'worktree isolado' README.md
# negativo: cada marcador da edição aparece exatamente 1 vez
test "$(grep -n '<!-- edicao:inicio -->' README.md | wc -l)" -eq 1
test "$(grep -n '<!-- edicao:fim -->' README.md | wc -l)" -eq 1
# negativo: nenhum link para a âncora removida
test "$(grep -rIn '#como-funciona' README.md npm/README.md docs site/src sdd/memory .agents 2>/dev/null | wc -l)" -eq 0
# seções vizinhas intactas
test "$(grep -n '^## Quem faz o quê$' README.md | wc -l)" -eq 1
test "$(grep -n '^## O que é obsoleto (não use)$' README.md | wc -l)" -eq 1
# a imagem referenciada existe
test -f docs/img/esteira-3-sessoes.svg
go test -list 'TestScaffold_DocsEdicao' ./internal/scaffold | grep -F 'TestScaffold_DocsEdicao' >/dev/null
go test ./internal/scaffold -run 'TestScaffold_DocsEdicao' -count=1
```
