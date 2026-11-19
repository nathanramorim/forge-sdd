# fix-88-04 — AskUserQuestion para confirmar merge após aprovação do Reviewer

**Branch:** `fix/88-04-ask-user-merge-apos-aprovacao`
**Origem:** Iniciativa 88, `.agents/skills/sdd-operacao/references/inbox.md`, fluxo do `forge inbox` no Líder
**Tamanho:** `S`
**Status:** `ready`

## Sintoma
Quando o Reviewer aprova um PR (veredito `conforme`), a notificação chega no inbox do Líder (sessão Discovery), mas o agente não apresenta uma escolha clara ao usuário. O usuário fica sem saber se pode mergear e precisa ir ao GitHub manualmente para verificar. O comportamento esperado: ao processar uma notificação de PR aprovado, o Líder usa `AskUserQuestion` com as opções "Ver o PR antes" / "Mergear agora" diretamente na sessão Discovery.

## Feature Correlacionada
- **Arquivo de origem:** `.agents/skills/sdd-operacao/references/inbox.md` — passo 2 do `/inbox` não contempla AskUserQuestion para notificações de PR aprovado.
- **Causa-raiz:** `inbox.md` descreve o fluxo de leitura de notificações mas não instrui o agente a perguntar ao usuário sobre o merge quando a notificação for de aprovação. O agente lê a notificação, mostra o texto e para — sem oferecer ação.
- **Evidência:**
  - `.agents/skills/sdd-operacao/references/inbox.md:3–9` — sem menção a `AskUserQuestion` nem a ação de merge.
  - `fix-88-03` (PR #91) adicionou a regra geral de usar `AskUserQuestion` para perguntas nas skills, mas `inbox.md` é uma referência de comando, não um SKILL.md, e não foi atualizado.

## Modelo do módulo
**Enxuto** (ARQ-11): atualização de `.agents/skills/sdd-operacao/references/inbox.md` e do template correspondente em `internal/scaffold/templates/`. Sem código Go novo.

## Decisões do projeto aplicadas
- ADR-0005 (Menu Navegável e Agente por Sessão), ADR-0004 (Pipeline Multi-Sessão), fix-88-03.
- Rules: `sdd-operacao`, `sdd-fluxo-e-portao` (FLX-01, FLX-08), `sdd-acp` (GOV-04).

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Fluxo de merge no inbox | `.agents/skills/sdd-operacao/references/inbox.md` | Leitura + mark-read; sem AskUserQuestion | **estender** o passo 2 com: se notificação for PR aprovado → AskUserQuestion("PR pronto para merge", ["Ver o PR antes", "Mergear agora"]) | fix-88-03 mandatou AskUserQuestion para perguntas; inbox é o ponto de decisão de merge |
| Template do inbox | `internal/scaffold/templates/.agents/skills/sdd-operacao/references/inbox.md.tmpl` | Mesmo conteúdo sem AskUserQuestion | **atualizar** em sincronia com a mudança acima | Novos projetos devem herdar o comportamento |
| Comentário de aprovação no PR | `.agents/skills/sdd-build/references/worker-job.md` ou similar | Reviewer posta comentário de bloqueio mas não de aprovação final | **fora do escopo deste fix** (gap separado; já tratado manualmente no PR #90) | Escopo mínimo |

## Dependências novas
Nenhuma. Depende de fix-88-03 (PR #91) já mergeado ou na mesma base.

## Rules e IDs da Constituição Aplicáveis
`COD-01`, `DUP-01`, `FLX-01`, `FLX-08`, `GOV-04`.

## Premissas (a confirmar no Gate 1)
- O Líder é o agente que processa o `/inbox` na sessão Discovery.
- A notificação de PR aprovado tem tipo distinguível (campo `channel` ou texto) no `forge inbox`.
- O `AskUserQuestion` com opção "Mergear agora" leva o Líder a executar `gh pr merge <número> --squash --delete-branch`; "Ver o PR antes" exibe o link e aguarda.

## Escopo
1. `.agents/skills/sdd-operacao/references/inbox.md`: adicionar passo 2b após "identifique o tipo da notificação":
   - Se PR aprovado (`conforme`): usar `AskUserQuestion` com título "✅ PR #<n> aprovado — o que fazer?" e opções `["Ver o PR antes — abre o link e aguarda", "Mergear agora — gh pr merge <n> --squash --delete-branch"]`.
   - Se aprovado e usuário escolher "Mergear agora": executar o merge e marcar a notificação como lida.
2. `internal/scaffold/templates/.agents/skills/sdd-operacao/references/inbox.md.tmpl`: mesma atualização.
3. Golden de teste `internal/scaffold/testdata/golden/.agents/skills/sdd-operacao/references/inbox.md`: atualizar.

## Fora do escopo
- Postar comentário de aprovação final no GitHub (gap separado — tratado manualmente até haver fix dedicado).
- Mudar o fluxo para notificações que não sejam de PR aprovado.
- Alterar `forge inbox` (Go) ou o bus de notificações.

## Critérios de Aceitação Executáveis
```bash
# Auditoria: inbox.md contém AskUserQuestion
grep -n "AskUserQuestion" .agents/skills/sdd-operacao/references/inbox.md
# Deve retornar pelo menos 1 linha.

grep -n "AskUserQuestion" internal/scaffold/templates/.agents/skills/sdd-operacao/references/inbox.md.tmpl
# Deve retornar pelo menos 1 linha.

go test ./internal/scaffold/... -run 'TestGolden'
go vet ./...
go test ./...
```
- Em uso real: após veredito `conforme` do Reviewer, o Líder apresenta `AskUserQuestion` com "Ver o PR antes" / "Mergear agora" na sessão Discovery.
- Escolha "Mergear agora" executa `gh pr merge` sem pedir nova confirmação.
