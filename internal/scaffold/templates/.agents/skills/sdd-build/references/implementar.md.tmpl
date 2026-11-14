# Comando: implementar

**Uso:** `/implementar` (executa a fila da iniciativa da branch atual) ou `/implementar <nome-da-feature>`.

## Modo
- **Pipeline** (existe `.forge/config/forge.json`): `/implementar <feature>` (alias `/implement`) **não constrói nesta sessão**. Consulte o card da feature pelo MCP do gerenciador de projeto, grave `card-evidence.json` na pasta da iniciativa (`provider`, `id`, `url`, `status`, `feature` e `checked_at` em RFC3339) e rode `forge-sdd dispatch <feature>`. Se o gate listar pendências, repasse-as ao usuário e pare. Se passar, mova o card para `pm_mcp.status_on_implement` pelo MCP e confirme o job enfileirado; o Builder e o Reviewer rodam nas outras sessões (`forge` ou `forge-sdd up`). Nesta sessão, nunca edite `forge.json` nem o código da feature; grave `card-evidence.json` só conforme o passo acima, e só com `pm_mcp.require_existing_card: true` (com `false` não há card nem evidência).
- **Sessão única** (sem `forge.json`): o fluxo abaixo, o da 4.0.0-beta.

## 1. Resolver o alvo e reportar o plano
- **Sem argumento:** use a iniciativa da branch atual (`forge-sdd status`).
- **Com argumento:** procure `sdd/discovery/*<nome>*`. Se houver mais de uma correspondência, liste e pergunte. Use `forge-sdd status --all-branches` para achar iniciativas que vivem em outra branch.
- Reporte iniciativa, branch e tasks na fila (por situação) e aguarde confirmação do usuário, a menos que `--yes` tenha sido pedido.

## 2. Branch (Regra 15, `FLX-02`)
Uma única branch cobre toda a pasta da iniciativa. Se a branch já existir, pergunte se retoma; se não, pergunte a branch de partida (padrão `main`) e crie `feat/<id>-<nome>`.

## 3. Laço da fila
Para cada task em ordem de dependência:
1. `forge-sdd queue next` e implemente conforme `SKILL.md` do `sdd-build`.
2. Rode os critérios da spec e `forge-sdd gate`; `forge-sdd queue complete <task-id>`.
3. Camada B: acione o Analista de Governança (skill `sdd-governanca`); registre com `forge-sdd queue review`. Em `violacao`, corrija e repita a task; em `conforme`, siga.
4. Pare ao esgotar a fila ou quando o portão falhar de forma que exija decisão humana.

## 4. Fechamento
- Atualize `status.md` da iniciativa (até 1 KB) e escreva a release note no topo de `sdd/releases/history.md` (seção `## Entregas`).
- `git push origin <branch>` (sem `gh`, informe o link da branch); consulte `VCS / Work Item System` na Constitution: `github` usa `gh pr create --fill` (versões `-beta` não mesclam em `main`), `azure-devops` usa `az repos pr create`, `nenhum` só informa a branch pronta.
- Telemetria por task: `forge-sdd session record --feature "<spec>" --phase "<id>" --agent-path "lider,builder,governance" --outcome approved|rejected|blocked --criterio-atendido=true|false --tokens-input <n> --tokens-output <n> --turns <n> --duration-seconds <n> --model "<modelo>" --files-touched "<arquivos>"`. Estime os tokens de forma realista (1 token ≈ 4 caracteres) e nunca zerados se houve interação. Sessão interrompida, cancelada ou com timeout: `--outcome blocked --criterio-atendido=false`.
