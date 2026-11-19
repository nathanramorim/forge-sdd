# fix-88-03 — AskUserQuestion e seletor huh obrigatórios para perguntas

**Branch:** `fix/88-03-ask-user-e-seletor-obrigatorios`
**Origem:** Iniciativa 88 (menu interativo), `.agents/skills/sdd-*/SKILL.md`, `cmd/forge-sdd/menu.go`
**Tamanho:** `S`
**Status:** `ready`

## Sintoma
Quando um agente Claude (Discovery, Builder ou Reviewer) precisa fazer uma pergunta ao usuário, o componente interativo de seleção do Claude Code (`AskUserQuestion`) às vezes não aparece — a pergunta é feita como texto plano e o usuário precisa digitar a resposta. O mesmo vale para o forge CLI: em alguns fluxos o `huh` selector (setas + Enter) não é invocado e a pergunta é feita sem a UI visual. O comportamento deve ser consistente: **toda pergunta ao usuário usa o mecanismo interativo disponível** — `AskUserQuestion` para agentes Claude, seletor `huh` para o forge CLI.

## Feature Correlacionada
- **Arquivo de origem:** Iniciativas 71 (menu), 88 (modo pipeline avançado), skills em `.agents/skills/sdd-*/SKILL.md`.
- **Causa-raiz:**
  1. As skills dos agentes (Discovery, Builder, Reviewer, Líder) não têm instrução explícita para usar `AskUserQuestion` em toda pergunta ao usuário; a ferramenta é usada por heurística natural do modelo, não por regra declarada.
  2. No `sdd/memory/clarify.md` a lógica de "quando perguntar" está documentada, mas não diz qual ferramenta usar para a pergunta.
  3. No forge CLI, o `huhPrompter` em `cmd/forge-sdd/menu.go` é o mecanismo correto, mas não há verificação de que todo ponto de pergunta usa `env.p.Select`/`env.p.Confirm` (algumas podem estar usando `fmt.Scan` ou outputs diretos em fluxos não cobertos).
- **Evidência:**
  - `.agents/skills/sdd-*/SKILL.md` — sem menção a `AskUserQuestion`.
  - `sdd/memory/clarify.md` — define quando perguntar, mas não como.
  - `cmd/forge-sdd/menu.go:30–48` — `Prompter` interface com `Select`/`Confirm`/`Pause`; verificar se todos os pontos de pergunta usam a interface (não `fmt.Scan`).

## Modelo do módulo
**Enxuto** (ARQ-11): atualização de skills (`.agents/skills/`), `sdd/memory/clarify.md` e verificação de conformidade em `cmd/forge-sdd/menu.go`. Sem módulo novo.

## Decisões do projeto aplicadas
- D3 (huh v0.3), ADR-0003 (Skills por Estação), ADR-0005 (Menu Navegável e Agente por Sessão).
- Rules: `sdd-acp.md` (GOV-04), `sdd-fluxo-e-portao.md` (FLX-08), `sdd-codigo-e-testes.md` (COD-01).

## O que já existe
| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Instrução de `AskUserQuestion` nas skills | `.agents/skills/sdd-*/SKILL.md` | Ausente | **adicionar** regra explícita: toda pergunta ao usuário via `AskUserQuestion`; nunca plain text | Garante comportamento consistente e não depende de heurística do modelo |
| Definição de "como perguntar" no clarify.md | `sdd/memory/clarify.md` | Define quando, não como | **estender** com seção `## Como perguntar`: `AskUserQuestion` com 2–4 opções objetivas; plain text apenas se a pergunta exige resposta aberta sem alternativas fixas | Mantém a heurística já existente e adiciona a ferramenta correta |
| Conformidade no forge CLI | `cmd/forge-sdd/menu.go` | Interface `Prompter` com `Select`/`Confirm`/`Pause`; métodos cobertos por `huhPrompter` | **auditar** todos os `fmt.Scan`, `fmt.Scanf` e `bufio.NewReader(os.Stdin).ReadString` no pacote `cmd/`; substituir por `env.p.Select`/`env.p.Confirm` se encontrados | Fecha a lacuna sem alterar a interface existente |

## Dependências novas
Nenhuma.

## Rules e IDs da Constituição Aplicáveis
`COD-01`, `DUP-01`, `FLX-08`, `GOV-04`.

## Premissas (a confirmar no Gate 1)
- A ferramenta `AskUserQuestion` está disponível em todas as sessões do pipeline (Discovery, Builder, Reviewer) no Claude Code.
- Perguntas abertas (resposta livre) continuam como plain text; a regra se aplica a perguntas de múltipla escolha ou sim/não.
- A auditoria de `fmt.Scan` não encontrará usos fora do `huhPrompter`; se encontrar, o escopo aumenta.

## Escopo
1. `.agents/skills/sdd-*/SKILL.md` (todos os agentes): adicionar seção `## Como perguntar ao usuário` com regra: toda pergunta de múltipla escolha ou sim/não usa `AskUserQuestion` com 2–4 opções; nunca plain text com digitação livre.
2. `sdd/memory/clarify.md`: adicionar `## Como perguntar` ao final.
3. `cmd/forge-sdd/`: grep por `fmt.Scan`, `bufio.NewReader(os.Stdin)`, `os.Stdin.Read`; se encontrado, substituir por `env.p.Select`/`env.p.Confirm`.
4. Atualizar `fix-90-01` como dependência (o seletor huh só funciona consistentemente depois que o fix de Windows estiver resolvido).

## Fora do escopo
- Mudar quando perguntar (clarify.md heurística de quando).
- Adicionar novas opções ao menu do forge CLI.
- Alterar a interface `Prompter` ou o `huhPrompter`.

## Critérios de Aceitação Executáveis
```bash
# Auditoria: nenhum fmt.Scan fora de testes no pacote cmd/
grep -rn "fmt\.Scan\|bufio\.NewReader(os\.Stdin)" cmd/ --include="*.go" | grep -v "_test.go"
# Deve retornar vazio.

go vet ./...
go test ./...
```
- [x] Todos os agentes (Discovery, Builder, Reviewer, Líder) têm instrução `## Como perguntar ao usuário` nas suas SKILL.md.
- [x] `sdd/memory/clarify.md` tem seção `## Como perguntar`.
- [x] Nenhum `fmt.Scan` / `bufio.NewReader(os.Stdin)` fora de testes no pacote `cmd/`.
- Em uso real: perguntas de múltipla escolha do Discovery aparecem como `AskUserQuestion` dialog no Claude Code.
