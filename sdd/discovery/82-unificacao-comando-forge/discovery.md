# Discovery 82 — Unificação e Padronização do Comando `forge`

## Problema
Atualmente há uma dualidade e inconsistência histórica de nomenclatura em todo o ecossistema Forge-SDD:
1. **Comandos de Usuário vs Comandos da Metodologia:** O usuário aprende a usar `forge` para o menu e comandos de entrada (`forge init`, `forge doctor`, `forge`), mas o próprio CLI exibe cabeçalhos como `Usage: forge-sdd status [flags]`, dicas como `[Dica: Use 'forge-sdd ps --kill-zombies']` e erros orientando `forge-sdd dispatch`.
2. **Prompts e Regras Internas:** Os prompts de agentes ([GEMINI.md](file:///Users/nathanramorim/git/forge-sdd/GEMINI.md), [CLAUDE.md](file:///Users/nathanramorim/git/forge-sdd/CLAUDE.md)), as regras (`.agents/rules/`) e as skills usam `forge-sdd brief`, `forge-sdd status`, `forge-sdd gate`, gerando atrito cognitivo e a sensação de que certos comandos só funcionam com um prefixo ou outro.
3. **Templates Scaffoldados:** Novos projetos scaffoldados recebem templates e scripts (`forge-up.sh`, `forge-up.ps1`, `LEIA-ME.md`) que continuam instruindo o uso de `forge-sdd`.

## Para quem
- **Desenvolvedores:** Que utilizam o CLI no terminal e precisam de uma interface concisa, padronizada e sem ambiguidade.
- **Agentes de IA (Discovery, Builder, Reviewer):** Que executam verificações, gates e transições de estado via CLI.

## Jornada do Usuário
O desenvolvedor instala o pacote (`npm i -g @nathanramorim/forge-sdd`) e utiliza **exclusivamente `forge`** para todas as interações:
- `forge` (menu)
- `forge init` / `forge update` / `forge doctor`
- `forge status` / `forge status --live`
- `forge ps` / `forge ps --kill-zombies`
- `forge gate` / `forge brief` / `forge dispatch`
Todas as telas de ajuda (`forge --help`), mensagens de erro, dicas e logs referenciam unicamente `forge`. Caso o usuário ainda digite `forge-sdd` por hábito antigo, o wrapper continuará funcionando de forma transparente (retrocompatibilidade 100%).

## Regras de Negócio
1. `forge` é o comando canônico universal em 100% da documentação, CLI, telas de ajuda, mensagens de erro, regras e prompts.
2. `forge-sdd` é mantido no wrapper npm (`npm/package.json`) como alias retrocompatível 1:1, mas nenhuma tela, mensagem ou doc deve promover ou exibir `forge-sdd`.
3. Todo scaffold (`forge init` / `forge update`) deve gerar artefatos, scripts e instruções contendo unicamente `forge`.

## Escopo
### Dentro do Escopo
1. **CLI Go (`cmd/` e `internal/`):**
   - Alteração de `rootCmd.Use = "forge"` e descrições no Cobra.
   - Varredura de strings fixas em mensagens de tela, dicas, avisos e saídas formatadas (ex: `ps.go`, `doctor.go`, `pipeline.go`, `queue.go`, `status.go`).
   - Atualização dos testes unitários em `cmd/` e `internal/` que validavam a saída antiga `forge-sdd`.
2. **Templates de Scaffold e Testes Golden:**
   - Varredura em `internal/scaffold/templates/` (`LEIA-ME.md.tmpl`, `forge-up.sh`, `forge-up.ps1`, `.github/`, `.gemini/`, `.claude/`).
   - Sincronização dos goldens em `internal/scaffold/testdata/golden/`.
3. **Regras, Skills e Prompts de IA:**
   - Padronização em `.agents/rules/`, `.agents/skills/`, `GEMINI.md`, `CLAUDE.md`, `.gemini/` e `.claude/`.
4. **Documentação Geral:**
   - Atualização de `docs/comandos.md`, `README.md` e `npm/README.md`.

### Fora do Escopo
- Remoção do binário alias `forge-sdd` em `npm/package.json` (permanece como alias transparente).
- Renomeação do nome do pacote no npm (`@nathanramorim/forge-sdd`) ou repositório GitHub.

## Riscos e Mitigações
- **Risco:** Quebra de asserções em testes existentes que procuravam por `"forge-sdd"`.
  - **Mitigação:** Atualização simultânea das asserções de teste correspondentes e validação com `go test ./...` e `go vet ./...`.
- **Risco:** Regressão nos golden files do scaffold.
  - **Mitigação:** Rodar `go test ./internal/scaffold/...` e validar que os goldens gerados estão 100% alinhados com os templates.
