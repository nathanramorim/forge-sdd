# Comando: implementar

**Uso:** `/implementar` (executa a fila da iniciativa da branch atual) ou `/implementar <nome-da-feature>`.

## Modo
- **Pipeline** (existe `.forge/config/forge.json`): `/implementar <feature>` (alias `/implement`) **não constrói nesta sessão**. Consulte o card da feature pelo MCP do gerenciador de projeto, grave `card-evidence.json` na pasta da iniciativa (`provider`, `id`, `url`, `status`, `feature` e `checked_at` em RFC3339) e rode `forge dispatch <feature>`. Se o gate listar pendências, repasse-as ao usuário e pare. Se passar, mova o card para `pm_mcp.status_on_implement` pelo MCP e confirme o job enfileirado; o Builder e o Reviewer rodam nas outras sessões (`forge` ou `forge up`). Nesta sessão, nunca edite `forge.json` nem o código da feature; grave `card-evidence.json` só conforme o passo acima, e só com `pm_mcp.require_existing_card: true` (com `false` não há card nem evidência).
- **Sessão única** (sem `forge.json`): o fluxo abaixo, o da 4.0.0-beta.

## 1. Resolver o alvo e reportar o plano
- **Sem argumento:** use a iniciativa da branch atual (`forge status`).
- **Com argumento:** procure `sdd/discovery/*<nome>*`. Se houver mais de uma correspondência, liste e pergunte. Use `forge status --all-branches` para achar iniciativas que vivem em outra branch.
- Reporte iniciativa, branch e tasks na fila (por situação) e aguarde confirmação do usuário, a menos que `--yes` tenha sido pedido.

## 2. Uma branch e um PR por feature (Regra 15, `FLX-02`, ADR-0022)

**Não existe mais "uma branch cobre toda a iniciativa".** A spec já voltou à principal; a construção
parte dela:

- **Uma branch por feature do split:** `feat/<id>-<nome>-<sid>` (ou `fix/...` quando a task é fix),
  sempre **aberta a partir da principal** — nunca da branch da spec, que já foi mesclada e sumiu.
- **Um PR em draft por feature**, com a base na principal. Nenhum PR tem base em outro PR.
- Se a branch da feature já existir, pergunte se retoma; se não, crie a partir de `main`.
- A ordem dentro da onda respeita o plano: uma feature só começa com as dependências **mescladas**
  (feat-106-24c), não apenas abertas.

## 3. Laço da fila
Para cada task em ordem de dependência:
1. `forge queue next` e implemente conforme `SKILL.md` do `sdd-build`.
2. Rode os critérios da spec e `forge gate`; `forge queue complete <task-id>`.
3. Camada B: acione o Analista de Governança (skill `sdd-governanca`); registre com `forge queue review`. Em `violacao`, corrija e repita a task; em `conforme`, siga.
4. Pare ao esgotar a fila ou quando o portão falhar de forma que exija decisão humana.

## 4. Fechamento
- Atualize `status.md` da iniciativa (até 1 KB) e escreva a release note no topo de `sdd/releases/history.md` (seção `## Entregas`).
- **PR da feature em draft, pelo adaptador:** consulte `VCS / Work Item System` na Constitution.
  - `github`: `gh pr create --draft --base main --head <branch> --fill`. O `--draft` é **obrigatório**
    — um PR que nasce pronto fura a Conferência 2 e o CI roda caro antes da revisão (`RN-7`).
  - `azure-devops`: `az repos pr create --draft true`.
  - `nenhum`: só informe a branch pronta; o merge é local (`RN-25`).
  **Nunca** use `gh pr create --fill` sozinho: sem `--draft` o PR nasce pronto.
- **Marcar pronto é outra etapa:** só depois da Conferência 2 e do `[Continuar]`, com parcel
  `conforme` para o **HEAD atual** (feat-106-21). `gh pr ready` antes disso é recusado.
- Telemetria por task: `forge session record --feature "<spec>" --phase "<id>" --agent-path "lider,builder,governance" --outcome approved|rejected|blocked --criterio-atendido=true|false --tokens-input <n> --tokens-output <n> --turns <n> --duration-seconds <n> --model "<modelo>" --files-touched "<arquivos>"`. Estime os tokens de forma realista (1 token ≈ 4 caracteres) e nunca zerados se houve interação. Sessão interrompida, cancelada ou com timeout: `--outcome blocked --criterio-atendido=false`.
