# fix-106-14b — O hook do pr_guard tolera script ausente por um wrapper em Go e migra o comando antigo

**Branch da spec:** `fix/106-14-spec-hook-sem-script`; **da correção:** `fix/106-14b-impl-wrapper`, a partir da principal, **depois do 14a**
**Origem:** Iniciativa 106, mesmo teste de upgrade; metade 2 de 2. **Depende de:** fix-106-14a mesclado.
**Tamanho:** `L`
**Status:** `spec`

## Sintoma

Ver o sintoma completo em `fix-106-14a-entregar-o-script-antes-do-hook.md`: o hook do `pr_guard` instalado sem o script sai com código 2 e bloqueia o Bash do Claude Code. Esta metade trata do que resta **depois** do 14a: o hook continua dependendo de o arquivo existir. **Fonte oficial de "código 2 bloqueia": não verificada**; o Builder a cita da documentação do Claude Code (context7) no PR, e a decisão desta spec depende dela.

## Feature Correlacionada

- **Arquivo de origem:** `internal/scaffold/hooks.go` (`prGuardCommand`) e o `main` do `forge`.
- **Causa-raiz do que falta depois do 14a:** o comando do hook depende de o script existir; qualquer projeto em que ele sumir (apagado, branch antiga, checkout parcial) volta ao bloqueio. Um comando de shell (`[ -f ... ] && ... || exit 0`) **não serve**: é sintaxe POSIX e, por `A && B || C`, também sai 0 quando o guard devolve 2, mascarando todo bloqueio legítimo.

## Modelo do módulo

**Completo** (`ARQ-11`): subcomando, troca do comando do hook, migração no merge e doctor. **`ARQ-06` (definitivo):** `forge hook pr-guard` é subcomando novo, e a ADR-0025 cobre só `skills`. Este fix **emenda a `ARQ-06`** em `sdd/memory/constitution.md` (acrescenta `hook`) e traz a **ADR-0027** no mesmo PR (`GOV-01`). Isso é Constituição e ADR: o merge **exige aprovação explícita do dono**, fora do piloto automático.

## Dependências novas

Nenhuma (`SEG-07`).

## Rules e IDs da Constituição Aplicáveis

`ARQ-06`, `GOV-01`, `SEG-01`, `SEG-02`, `TST-02`, `TST-03`, `DUP-01`.

## Decisões do projeto aplicadas

- O hook passa a chamar `forge hook pr-guard` (o `forge` já está no PATH: o doctor e o inbox usam `forge ...` nos hooks). Portável em Windows, macOS e Linux.
- Subcomando oculto: `Hidden: true` do cobra (não há precedente no código; o doctor e o inbox são visíveis).
- Resolução do caminho: relativo à **raiz do projeto** (`CLAUDE_PROJECT_DIR`; se ausente, o `git rev-parse --show-toplevel`; se nenhum, o `cwd`), pois o `cwd` do hook pode não ser a raiz.
- Sem `uv` no PATH: aviso em stderr e **saída 0** (fail-open consciente; o job `pronto?` do CI é o apoio).
- Semântica do wrapper: **sem** `scripts/pr_guard.py`, imprime um aviso em stderr e sai **0**; **com** o script, executa `uv run scripts/pr_guard.py --hook`, repassa `stdin` e devolve **o mesmo código de saída** (o bloqueio real, código 2, permanece).
- **Detecção:** os marcadores de `MergeClaudePreToolUseHRook` (`hooks.go:46-47`) e o `reportPRGuardCoverage` (`doctor.go:363-373`) passam a reconhecer **os dois** comandos (`pr_guard.py` e `hook pr-guard`).
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
# os testes de repasse usam um `uv` FALSO (stub no PATH do teste), nunca o uv real: a CI roda em Linux e a validação local em macOS
go test ./cmd/forge-sdd -run '^(TestHookPRGuardSemScriptSaiZero|TestHookPRGuardRepassaCodigo2|TestHookPRGuardRepassaStdin)$' -count=1
go test ./cmd/forge-sdd -run '^(TestHookPRGuardSemUvSaiZero|TestDoctorReconheceOsDoisComandosDoHook)$' -count=1
go test ./internal/scaffold -run '^(TestHookPRGuardMigraComandoAntigo|TestHookPRGuardNaoDuplicaNemTocaHooksDoUsuario)$' -count=1
[ "$(go test ./cmd/forge-sdd ./internal/scaffold -list '^(TestHookPRGuardSemScriptSaiZero|TestHookPRGuardRepassaCodigo2|TestHookPRGuardRepassaStdin|TestHookPRGuardSemUvSaiZero|TestDoctorReconheceOsDoisComandosDoHook|TestHookPRGuardMigraComandoAntigo|TestHookPRGuardNaoDuplicaNemTocaHooksDoUsuario)$' | grep -c '^Test')" -eq 7 ]
go test ./cmd/forge-sdd ./internal/scaffold -count=1
```

**Manual do Builder:** com `forge-beta`, projeto sem o script: `echo '{...}' | forge hook pr-guard; echo $?` imprime 0; com o script e um `gh pr merge` sem parecer, devolve 2.

## Fora do escopo

- Entregar o script (fix-106-14a).
