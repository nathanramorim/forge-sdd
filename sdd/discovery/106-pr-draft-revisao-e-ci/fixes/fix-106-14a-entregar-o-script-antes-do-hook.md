# fix-106-14a — Entregar `scripts/pr_guard.py` no `init` e no `update` e só instalar o hook depois dele

**Branch da spec:** `fix/106-14-spec-hook-sem-script`; **da correção:** `fix/106-14a-impl-script-antes-do-hook`, a partir da principal
**Origem:** Iniciativa 106, teste do upgrade 5.0 para a beta em projeto descartável (22/11/2026), depois da pergunta do dono sobre o comando de reconfiguração. Metade 1 de 2 (a metade 2 é o fix-106-14b).
**Tamanho:** `L`
**Status:** `spec`

## Sintoma

Em projeto novo (`forge init --agent claude`) e em projeto atualizado (`forge update --upgrade`, 5.0 para a 5.2.0-beta), o `.claude/settings.json` ganha o hook `PreToolUse` (matcher `Bash`) com `uv run scripts/pr_guard.py --hook`, **mas não existe `scripts/pr_guard.py`** no projeto. O hook sai com `error: Failed to spawn ... No such file or directory` e **código 2**. Pela convenção de hooks (declarada na docstring do próprio `pr_guard.py`: "0 permitido; 2 bloqueado"), código 2 em `PreToolUse` **bloqueia a chamada**, e o agente não consegue consertar. Reproduzido por dois caminhos independentes (`/tmp/up-test`, `/tmp/init-test` e a conferência do Reviewer). **Fonte oficial do "código 2 bloqueia": não verificada nesta spec**; o Builder a cita da documentação do Claude Code (context7) no PR.

## Feature Correlacionada

- **Arquivo de origem:** `internal/scaffold/hooks.go` (`prGuardCommand`, `MergeClaudePreToolUseHRook`), chamado por `scaffold.go` (`installClaudeInboxHook`, no `init` e no `update`), feat-106-14/15.
- **Causa-raiz:** o hook é instalado pelo Go, mas o script que ele chama só existe no **repositório do forge-sdd**; nenhum template o embute, e o `/upgrade-sdd` roda no agente, que já estaria bloqueado.
- **Evidência:** `find internal/scaffold/templates -name 'pr_guard*'` vazio; `forge doctor` diz "Hook PreToolUse instalado" sem checar o arquivo (`doctor.go:363-373`).

## Modelo do módulo

**Completo** (`ARQ-11`): embed do script, instalador no `init`/`update`, checagem do `doctor`, ajuste do guard. Justificativa: toca scaffold, doctor e script do guard.

## Dependências novas

Nenhuma (`SEG-07`). O `pr_guard.py` usa só a biblioteca padrão e **não** depende de `atestado.py` nem de `vcs`; o `atestado.py` e o `ci-politica.py` continuam do `/upgrade-sdd`.

## Rules e IDs da Constituição Aplicáveis

`ARQ-11`, `SEG-02` (gravação atômica; não sobrescreve script do usuário), `SEG-01` (entrada externa validada na borda: caminho confinado), `TST-02` (template e golden no mesmo PR), `TST-03` (casos negativos), `DUP-02` (reusar o que já existe: a cópia do template é validada contra o `scripts/pr_guard.py` do repositório por teste de igualdade, em vez de duplicada à mão).

## Decisões do projeto aplicadas

- Nunca instalar o hook sem o script.
- O script embutido leva a linha `# forge-sdd:generated v<versão>` **entre o shebang e o bloco PEP 723** (`# /// script`), injetada **na gravação** pelo instalador (não está no template). Sem a linha = editado pelo usuário = **intocado**; com a linha e versão menor = atualizado. O teste de igualdade compara o template com `scripts/pr_guard.py`; o golden de `init` que incluir `scripts/pr_guard.py` leva o marcador ou fica fora da comparação. `shouldPreserve` ganha a regra do marcador **antes** do laço genérico (hoje não há regra para `scripts/`, e o template novo sobrescreveria o script do usuário).
- **Selo do `ci-local.sh`:** o `ci-local.sh` do repositório é específico do forge-sdd (Go) e **não** é entregue. O `pr_guard.py` passa a exigir o selo **só quando `scripts/ci-local.sh` existe no projeto**; sem ele o `gh pr ready` exige apenas o parecer. No repositório do forge-sdd nada muda.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Script do guard | `scripts/pr_guard.py` (só no repo) | sem template | **embutir** em `internal/scaffold/templates/scripts/pr_guard.py.tmpl` |
| Preservação | `shouldPreserve` (`scaffold.go`) | sem regra para `scripts/` | **estender** com a regra do marcador |
| Diagnóstico | `doctor.go` `reportPRGuardCoverage` | só procura o texto no settings | **checar** que `scripts/pr_guard.py` existe (`os.Stat`); sem `--selftest` |
| `/upgrade-sdd` | `internal/scaffold/templates/.agents/skills/sdd-operacao/references/upgrade-sdd.md.tmpl` (linhas 18 e 20) e o golden `testdata/golden/.agents/skills/sdd-operacao/references/upgrade-sdd.md` | trata o guard como cópia manual | **atualizar** o texto e o golden: o `forge` entrega o guard; `atestado.py` e `ci-politica.py` seguem manuais |

## Comportamento esperado

1. `init` e `update` gravam `scripts/pr_guard.py` (atomicamente) **antes** de mesclar o hook; se a gravação falhar, o hook **não** é instalado e o comando avisa.
2. `update` repara projeto com hook sem script (grava o script, não duplica o hook).
3. Script com o marcador e versão menor é atualizado; sem o marcador, permanece intocado e o `update --dry-run` o lista.
4. `forge doctor` acusa **erro de configuração** quando o hook existe e o script não, com a correção (`forge update`).
5. Sem `scripts/ci-local.sh` no projeto, `gh pr ready` não exige selo. O teste existente `TestPRGuardExigeSeloLocalNoReady` (`scripts/pr_guard_test.go`, ~262-276) **deve ser ajustado**: a raiz temporária ganha `scripts/ci-local.sh`, e fica um caso negativo com o script presente e o selo ausente (bloqueia) e outro com o script ausente (libera).

## Critérios de Aceitação Executáveis

Vermelhos **até o Builder**. `-list` só prova existência; a aprovação é o `-run`, com nomes exatos.

```bash
set -o pipefail
go test ./internal/scaffold -run '^(TestInitEntregaPRGuardAntesDoHook|TestUpdateEntregaPRGuardAntesDoHook|TestFalhaDeGravacaoNaoInstalaOHook|TestPRGuardScriptDoUsuarioFicaIntocado|TestPRGuardGeradoAntigoEAtualizado|TestPRGuardTemplateIgualAoDoRepositorio)$' -count=1
go test ./cmd/forge-sdd -run '^(TestDoctorAcusaHookSemScript|TestUpdateReparaHookSemScript)$' -count=1
go test ./scripts -run '^TestPRGuardSemCILocalNaoExigeSelo$' -count=1
[ "$(go test ./internal/scaffold ./cmd/forge-sdd ./scripts -list '^(TestInitEntregaPRGuardAntesDoHook|TestUpdateEntregaPRGuardAntesDoHook|TestFalhaDeGravacaoNaoInstalaOHook|TestPRGuardScriptDoUsuarioFicaIntocado|TestPRGuardGeradoAntigoEAtualizado|TestPRGuardTemplateIgualAoDoRepositorio|TestDoctorAcusaHookSemScript|TestUpdateReparaHookSemScript|TestPRGuardSemCILocalNaoExigeSelo)$' | grep -c '^Test')" -eq 9 ]
go test ./internal/scaffold ./cmd/forge-sdd ./scripts -count=1
```

**Manual do Builder (relatar no PR):** projeto 5.0 em `/tmp`, `forge update --upgrade` e `echo '{"tool_name":"Bash","tool_input":{"command":"ls"}}' | uv run scripts/pr_guard.py --hook` devolve código 0.

## Fora do escopo

- O comando fail-open do hook e a migração do hook antigo (fix-106-14b).
- Propagação de rules editadas (fix-106-09).
