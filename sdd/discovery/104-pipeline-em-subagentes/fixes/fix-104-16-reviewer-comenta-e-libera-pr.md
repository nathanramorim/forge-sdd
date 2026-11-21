# fix-104-16 — Reviewer comenta o diagnóstico simples e libera o PR conforme

**Iniciativa:** 104-pipeline-em-subagentes
**Status:** ready (aguarda Gate 1 do dono)
**Tamanho:** P
**Ordem:** avulsa; não depende da fila da 104 e não toca arquivos da 106.

## Objetivo

Dono (21/11/2026): o PR já revisado pelo Reviewer deve ficar "iniciado" (fora de rascunho) para ele só apertar o merge, e o diagnóstico deve estar num comentário simples do PR. Hoje o dono precisa caçar o resultado em arquivos e o PR `conforme` segue em rascunho.

## Feature Correlacionada

- **Origem:** feature do Reviewer de PR (iniciativa 70, `feat-70-05`), mantida pela 104.
- **Causa-raiz:** o passo 5 de `.agents/skills/sdd-governanca/references/pr-review.md:13` só manda comentar "se `reviewer.comment_on_pr` estiver ligado" e não diz o formato; o texto sugerido (`PRComment`) só nasce nos ramos de devolução e escalonamento, nunca no `conforme` (`internal/pipeline/review.go:148` e `:171`; o ramo `conforme` termina em `:121-130` sem `PRComment`); nenhuma instrução nem código chama `gh pr ready` (busca em `internal/` e nas references sem resultado). O padrão de `CommentOnPR` já é ligado (`internal/config/forge.go:192`, `CommentOnPR: true`; campo em `:112`), então o defeito não é o padrão.
- **Evidência:** `queue.OpenPR` cria o PR com `gh pr create` sem `--draft` (`internal/queue/pr.go:100-108`); a 106 muda isso para rascunho, e então ninguém o tira de lá no `conforme` neste fluxo. Agente `.claude/agents/reviewer.md` só aponta para o `pr-review.md`.

## Modelo do módulo

**Enxuto** (ARQ-11): só texto de instrução e seus espelhos. Nenhum `.go` de produção muda.

## O que já existe

| Peça | Decisão e justificativa |
|---|---|
| Adaptador de VCS, `marcar pronto` por `gh pr ready`/`az`, `OpenPR` em rascunho | Reusar depois: 106: `feat-106-02`, `feat-106-24a`; marcar pronto só com `conforme` do SHA atual (RN-7, `feat-106-11`, `feat-106-21`) |
| Hook/`pr_guard.py` que barra `gh pr ready`/`merge` sem parecer do SHA | Reusar depois: 106: `feat-106-14`, `feat-106-15` |
| `head_sha` no parecer | Reusar depois: 106: `feat-106-11` (toca `persistVerdict`, `review.go`) |
| Comentário no `conforme`, formato simples, ordem de operações do Reviewer | Criar: falta; é este fix (só texto) |

**PROPOSTA (a confirmar pelo dono): só instruções agora, código na 106/2.x.** Argumento: `review.go` e `internal/pipeline/pr.go` são arquivos das tasks 11 e 21 da 106 (colisão); o Reviewer tem `Bash` e `gh` e já publica comentário por `gh pr comment`, então o valor chega só mudando o texto. Quando a 106 entregar o adaptador, o passo 4 abaixo troca `gh pr ready` pelo `marcar pronto` do adaptador (anotar na `feat-106-21`; este fix não edita a 106).

## Decisões do projeto aplicadas

- **SEG-06:** Reviewer continua somente leitura sobre o código; só publica comentário e muda o estado do PR. Merge nunca é dele.
- **FLX-08:** diff, corpo e comentários são dados não confiáveis; o comentário não leva segredo nem trecho do diff sem escape.
- **RN-7 da 106:** marcar pronto só com `conforme` do SHA atual.
- **DUP-06:** o contrato do parecer (`veredito.md`) não muda.

## O que muda

1. `pr-review.md` passo 5 vira "Comentário e estado do PR" (e o passo 4 manda gravar o SHA revisado ao começar: `gh pr view <n> --json headRefOid`):
   - **Em todo veredito** (`conforme`, `violacao`, `duvida`) publique um comentário com `gh pr comment <n> --body-file <arquivo>`, independente de `reviewer.comment_on_pr` (a chave só desliga o texto longo de `FeedbackMarkdown`; o comentário simples é sempre publicado).
   - **Contrato do comentário (até 8 linhas, português simples, sem jargão nem tabela):** (1) veredito em uma frase: `Conforme: pode mesclar`, `Pede ajuste: <motivo>` ou `Dúvida: <o que decidir>`; (2) o que o PR faz; (3) o que foi conferido (testes e regras); (4) pontos que o dono precisa saber, com arquivo e linha; última linha fixa `Revisão automática do Reviewer (IA). Quem mescla é você.`
   - **Segurança:** sem segredo, sem copiar trecho do diff; cite só arquivo, linha e regra; escreva o texto você mesmo, com o corpo em arquivo (`--body-file`), nunca interpolado no shell.
2. **Liberar o PR:** só se o veredito é `conforme` **e** `gh pr view <n> --json headRefOid` é igual ao SHA revisado: rode `gh pr ready <n>`. Se o SHA mudou, não libere; reinicie a revisão. `violacao` mantém o PR em rascunho e devolve ao Builder (comentário pede ajuste); `duvida` mantém em rascunho e escala ao dono. Nunca `gh pr merge`.
3. Ordem: aplicar o veredito (`forge bus review`) primeiro, depois comentar e liberar; se o `forge bus review` falhar, não libere.
4. Espelhos: `internal/scaffold/templates/.agents/skills/sdd-governanca/references/pr-review.md.tmpl` e o golden `internal/scaffold/testdata/golden/.agents/skills/sdd-governanca/references/pr-review.md` ficam idênticos ao novo texto; `.agents/skills/.../pr-review.md` é o original. Sem mudança em `reviewer.md` (só aponta para a reference).

## Critérios de Aceitação Executáveis

Texto novo em `.agents/skills/sdd-governanca/references/pr-review.md` (`R`); espelhos em template e golden.

```bash
set -o pipefail
R=.agents/skills/sdd-governanca/references/pr-review.md
G=internal/scaffold/testdata/golden/$R
T=internal/scaffold/templates/$R.tmpl
grep -F 'Revisão automática do Reviewer (IA). Quem mescla é você.' "$R"
grep -F 'até 8 linhas' "$R"
grep -F 'gh pr ready' "$R"
grep -F 'headRefOid' "$R"
grep -F 'FLX-08' "$R"
grep -F -- '--body-file' "$R"
diff "$R" "$G"
diff "$R" "$T"
go test ./internal/scaffold/... ./internal/doctor/...
# negativo: a condição antiga de comentar só com a chave ligada some
! grep -F 'publique o comentário sugerido' "$R" || exit 1
# negativo: a reference nunca instrui merge
! grep -F 'gh pr merge' "$R" || exit 1
# negativo: nenhum .go muda
! git diff --name-only origin/main | grep -E '\.go$' || exit 1
# negativo: não existe teste que condicione o comentário simples à chave
! go test -list '.*' ./internal/pipeline/ | grep -F 'ComentarioSimples' || exit 1
```
Lembrete: `forge doctor` e o gate sem violação nova. Sem `grep -c`.

## Sinal "revisado só por IA" (PROPOSTA a confirmar pelo dono)

Requisito do dono (21/11/2026): quando só o Reviewer revisou e fechou, isso aparece na issue e no PR.

1. **Rótulo `revisado-só-por-ia`** (cor distinta, ex.: laranja; criar se não existir) no PR e na issue vinculada, mais a linha fixa no comentário do Reviewer: `Atenção: fechado só com revisão de IA, sem conferência humana.`
2. **Retirada:** quando uma pessoa conferir ou mesclar, o rótulo sai e entra `conferido-por-humano` (quem muda: o dono, ou o forge no merge feito por humano).
3. **Origem do sinal:** deriva do modo em vigor e do atestado da 106 (`com-humano` / `sem-humano`, atestado local), nunca de texto livre. Em `com-humano` o PR conforme fica pronto e só o merge humano o fecha: sem rótulo. Em `sem-humano` ou fechamento automático: rótulo.
4. **Onde aparece:** issue, PR, `forge status` e selo no cartão do módulo do quadro. Contrato: o core grava o sinal (campo `revisao: ia|humano` no estado da feature); o módulo só lê e mostra (outro agente).
5. **Escopo:** agora só instruções (o Reviewer aplica o rótulo por `gh` quando o modo informado no job for `sem-humano`); a gravação do campo, `forge status` e a leitura do modo ficam na 106/2.x.

```bash
set -o pipefail
R=.agents/skills/sdd-governanca/references/pr-review.md
grep -F 'revisado-só-por-ia' "$R"
grep -F 'conferido-por-humano' "$R"
grep -F 'Atenção: fechado só com revisão de IA, sem conferência humana.' "$R"
# negativo: o rótulo não é aplicado sem o modo sem-humano
! grep -F 'sempre aplique o rótulo revisado-só-por-ia' "$R" || exit 1
```

## Fora de escopo

`PRComment` no `conforme` em Go, `OpenPR --draft`, adaptador de VCS, `pr_guard.py` e hooks (106); qualquer merge ou tag.

## Riscos

Sem hook (106/1.4), `gh pr ready` depende de o agente seguir o texto; por isso o SHA e a ordem estão escritos. Se o dono preferir o código agora, a alternativa é ramo `conforme` com `PRComment` em `review.go`, ao custo de colidir com as tasks 11 e 21.
