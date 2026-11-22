# fix-106-14 — O `forge init` e o `forge update` instalam o hook do pr_guard sem entregar o script, e o Claude Code fica bloqueado

**Branch da spec:** `fix/106-14-spec-hook-sem-script`; **da correção:** `fix/106-14-impl-hook-script`, a partir da principal
**Origem:** Iniciativa 106, teste do upgrade 5.0 para a beta em projeto descartável, 22/11/2026, depois da pergunta do dono "qual comando reconfigura para a nova estrutura?".
**Tamanho:** `M`
**Status:** `spec`

## Sintoma

Em projeto novo (`forge init --agent claude`) e em projeto atualizado (`forge update --upgrade`, 5.0 para 5.2.0-beta), o `.claude/settings.json` ganha o hook `PreToolUse` com `uv run scripts/pr_guard.py --hook` (matcher `Bash`), **mas não existe `scripts/pr_guard.py`** no projeto. Rodando o hook à mão: `error: Failed to spawn: scripts/pr_guard.py ... No such file or directory`, **código de saída 2**. No Claude Code, código 2 num hook `PreToolUse` **bloqueia a chamada da ferramenta**: todo comando Bash do agente seria barrado, sem o agente poder consertar. (Reproduzido em `/tmp/up-test` e `/tmp/init-test`.)

## Feature Correlacionada

- **Arquivo de origem:** `internal/scaffold/hooks.go` (`prGuardCommand`, `MergeClaudePreToolUseHRook`) chamado por `internal/scaffold/scaffold.go` (`installClaudeInboxHook`), feat-106-14/15.
- **Causa-raiz:** o hook é instalado pelo Go, mas o script que ele chama só existe no **repositório do forge-sdd** (`scripts/pr_guard.py`); nenhum template o embute e o `/upgrade-sdd` (que menciona `scripts/`) roda no agente, que já estaria bloqueado.
- **Evidência:** `find internal/scaffold/templates -name 'pr_guard*'` vazio; `ls scripts` em projeto recém-criado: inexistente; `forge doctor` diz "Hook PreToolUse instalado" sem checar o script.

## Modelo do módulo

**Completo** (`ARQ-11`): embed do script (e do que ele importa), instalador no `init` e no `update`, checagem do `doctor`.

## Dependências novas

Nenhuma (`SEG-07`): `uv` já é exigido e o `doctor` já avisa quando falta.

## Rules e IDs da Constituição Aplicáveis

`DUP-01` (uma fonte: o `scripts/pr_guard.py` do repositório é copiado para o template por teste de igualdade, não duplicado à mão), `SEG-02` (gravação atômica e sem sobrescrever script do usuário), `SEG-01` (caminho confinado ao projeto), `TST-02` e `TST-03` (golden e caso negativo), `FLX-07` (beta).

## Decisões do projeto aplicadas

- A barreira é segurança do fluxo: nunca instalar o hook sem o script, e nunca deixar o agente sem poder agir.
- Falha do guard por **ausência do script** deve **liberar** a chamada com aviso (fail-open) e nunca bloquear; bloquear só por violação real da política.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Script do guard | `scripts/pr_guard.py` (só no repo) | sem template | **embutir** em `internal/scaffold/templates/scripts/pr_guard.py` e instalar no `init`/`update` |
| Instalação do hook | `hooks.go` | instala sempre | **condicionar**: só instala o hook depois de garantir o script |
| Comando do hook | `prGuardCommand` | `uv run scripts/pr_guard.py --hook` | **tornar seguro**: `[ -f scripts/pr_guard.py ] && uv run scripts/pr_guard.py --hook || exit 0` (equivalente no Windows via `python`) |
| Diagnóstico | `doctor.go` `reportPRGuardCoverage` | só procura o texto no settings | **conferir** que o arquivo existe e roda (`--selftest`) |
| Atualização | fix-106-09 | `shouldPreserve` | **não sobrescrever** `scripts/pr_guard.py` editado pelo usuário (marcador de gerado) |

## Comportamento esperado

1. `forge init` e `forge update` gravam `scripts/pr_guard.py` (e `scripts/atestado.py` se for dependência) **antes** de mesclar o hook; se a gravação falhar, o hook **não** é instalado e o comando avisa.
2. O comando do hook é fail-open: sem o script, sai com 0 e uma linha de aviso em stderr.
3. `forge doctor` marca erro de configuração (não só "instalado") quando o hook existe e o script não, e diz como corrigir (`forge update`).
4. Projeto já atualizado (hook sem script): `forge update` repara gravando o script, sem duplicar o hook.
5. Script editado pelo usuário (sem o marcador de gerado) permanece intocado e o `update --dry-run` o lista.

## Critérios de Aceitação Executáveis

Vermelhos **até o Builder**. Os de projeto usam `t.TempDir()` e `uv` falso quando necessário.

```bash
set -o pipefail
# 1. init e update entregam o script antes do hook
go test ./internal/scaffold -run 'TestInitEntregaPRGuardAntesDoHook|TestUpdateEntregaPRGuardAntesDoHook' -count=1
# 2. hook fail-open: sem o script o comando sai com 0
go test ./internal/scaffold -run 'TestHookPRGuardSemScriptNaoBloqueia' -count=1
# 3. doctor acusa hook sem script e update repara
go test ./cmd/forge-sdd -run 'TestDoctorAcusaHookSemScript|TestUpdateReparaHookSemScript' -count=1
# 4. script do template igual ao scripts/pr_guard.py do repositório (DUP-01)
go test ./internal/scaffold -run 'TestPRGuardTemplateIgualAoDoRepositorio' -count=1
# 5. existência dos 6
[ "$(go test ./internal/scaffold ./cmd/forge-sdd -list 'PRGuard|HookSemScript' | grep -c '^Test')" -ge 6 ]
go test ./internal/scaffold ./cmd/forge-sdd ./scripts -count=1
```

6. **Reprodução ponta a ponta (manual do Builder):** projeto 5.0 em `/tmp`, `forge update --upgrade`, e `echo '{"tool_name":"Bash","tool_input":{"command":"ls"}}' | <comando do hook>` devolve código 0.

## Fora do escopo

- Entregar o `ci.yml`, o `ci-local.sh` e os demais scripts de política (continuam no `/upgrade-sdd`).
- A propagação de rules editadas (fix-106-09, furo 1).
