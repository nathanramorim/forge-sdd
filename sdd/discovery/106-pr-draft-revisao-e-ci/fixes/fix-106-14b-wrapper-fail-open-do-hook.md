# fix-106-14b — O hook do pr_guard tolera script ausente por um wrapper em Go e migra o comando antigo

**Branch da spec:** `fix/106-14-spec-hook-sem-script`; **da correção:** `fix/106-14b-impl-wrapper`, a partir da principal, **depois do 14a**
**Origem:** Iniciativa 106, mesmo teste de upgrade; metade 2 de 2. **Depende de:** fix-106-14a mesclado.
**Tamanho:** `M`
**Status:** `spec`

## Sintoma

Em projeto novo (`forge init --agent claude`) e em projeto atualizado (`forge update --upgrade`, 5.0 para a 5.2.0-beta), o `.claude/settings.json` ganha o hook `PreToolUse` (matcher `Bash`) com `uv run scripts/pr_guard.py --hook`, **mas não existe `scripts/pr_guard.py`** no projeto. O hook sai com `error: Failed to spawn ... No such file or directory` e **código 2**. Pela convenção de hooks (declarada na docstring do próprio `pr_guard.py`: "0 permitido; 2 bloqueado"), código 2 em `PreToolUse` **bloqueia a chamada**, e o agente não consegue consertar. Reproduzido por dois caminhos independentes (`/tmp/up-test`, `/tmp/init-test` e a conferência do Reviewer). **Fonte oficial do "código 2 bloqueia": não verificada nesta spec**; o Builder a cita da documentação do Claude Code (context7) no PR.

## Feature Correlacionada

- **Arquivo de origem:** `internal/scaffold/hooks.go` (`prGuardCommand`) e o `main` do `forge`.
- **Causa-raiz do que falta depois do 14a:** o comando do hook depende de o script existir; qualquer projeto em que ele sumir (apagado, branch antiga, checkout parcial) volta ao bloqueio. Um comando de shell (`[ -f ... ] && ... || exit 0`) **não serve**: é sintaxe POSIX e, por `A && B || C`, também sai 0 quando o guard devolve 2, mascarando todo bloqueio legítimo.

## Modelo do módulo

**Enxuto** (`ARQ-11`): um subcomando oculto e a troca da string do hook. **`ARQ-06`:** `forge hook pr-guard` é subcomando novo da CLI; vai com **ADR curta** (GOV-01 se alterar a lista) no mesmo PR, e o Builder confere se a ADR-0025 já cobre `hook`.

## Dependências novas

Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis

`ARQ-06`, `GOV-01`, `SEG-01`, `SEG-02`, `TST-02`, `TST-03`, `DUP-01`.

## Decisões do projeto aplicadas

- O hook passa a chamar `forge hook pr-guard` (o `forge` já está no PATH: o doctor e o inbox usam `forge ...` nos hooks). Portável em Windows, macOS e Linux.
- Semântica do wrapper: **sem** `scripts/pr_guard.py`, imprime um aviso em stderr e sai **0**; **com** o script, executa `uv run scripts/pr_guard.py --hook`, repassa `stdin` e devolve **o mesmo código de saída** (o bloqueio real, código 2, permanece).
- **Migração:** `MergeClaudePreToolUseHRook` passa a **substituir** a entrada antiga (`uv run scripts/pr_guard.py --hook`) pela nova, sem duplicar e sem tocar hooks do usuário nem JSON inválido; o `forge doctor` reconhece os dois comandos.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Hooks que chamam o `forge` | `doctorHookCommand`, inbox | `forge doctor`, `forge inbox ...` | **seguir o padrão** |
| Merge de hook | `mergeClaudeHookWithMatcher` | retorna "já instalado" se achar o marcador | **estender** com a substituição do comando antigo |

## Comportamento esperado

1. `forge hook pr-guard` (oculto) com o script presente devolve exatamente o código e o stderr do guard.
2. Sem o script: aviso e código 0.
3. `init`/`update` gravam o comando novo; projeto com o comando antigo é migrado pelo `update`.
4. O `doctor` mostra o estado correto nos dois casos.

## Critérios de Aceitação Executáveis

```bash
set -o pipefail
go test ./cmd/forge-sdd -run '^(TestHookPRGuardSemScriptSaiZero|TestHookPRGuardRepassaCodigo2|TestHookPRGuardRepassaStdin)$' -count=1
go test ./internal/scaffold -run '^(TestHookPRGuardMigraComandoAntigo|TestHookPRGuardNaoDuplicaNemTocaHooksDoUsuario)$' -count=1
[ "$(go test ./cmd/forge-sdd ./internal/scaffold -list '^(TestHookPRGuardSemScriptSaiZero|TestHookPRGuardRepassaCodigo2|TestHookPRGuardRepassaStdin|TestHookPRGuardMigraComandoAntigo|TestHookPRGuardNaoDuplicaNemTocaHooksDoUsuario)$' | grep -c '^Test')" -eq 5 ]
go test ./cmd/forge-sdd ./internal/scaffold -count=1
```

**Manual do Builder:** com `forge-beta`, projeto sem o script: `echo '{...}' | forge hook pr-guard; echo $?` imprime 0; com o script e um `gh pr merge` sem parecer, devolve 2.

## Fora do escopo

- Entregar o script (fix-106-14a).
