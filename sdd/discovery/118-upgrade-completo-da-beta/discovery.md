# 118 — Upgrade completo da 5.0 para a 5.2.0-beta

## Problema

Quem tem o forge-sdd 5.0 instalado e passa para a 5.2.0-beta recebe só **parte** do que a versão promete, e uma parte do que recebe **atrapalha**. Duas auditorias (inventário do código e teste empírico em 7 projetos descartáveis, 22/11/2026) mostram que o caminho `npm install -g ...@beta` + `forge update --upgrade` + `/upgrade-sdd` tem lacunas que o texto das notas e do `/upgrade-sdd` não admite.

## Para quem

Usuários do 5.0 que atualizam projetos existentes e usuários de projetos novos criados com a beta; Claude Code é o mais afetado (único agente com hook).

## Achados (verificados), por severidade

| # | Achado | Severidade |
|---|---|---|
| 1 | O hook `PreToolUse` (`uv run scripts/pr_guard.py --hook`) é instalado sem o script (no `init` e no `update`); sai com código 2 e **bloqueia o Bash do Claude Code**. O `/upgrade-sdd`, que deveria entregar os scripts, roda no agente já bloqueado | bloqueia |
| 2 | `update --upgrade` apaga a stack: o `.sddrc` do 5.0 não a guarda, e `CLAUDE.md`, `GEMINI.md`, `AGENTS.md` e `system_instructions.md` voltam com "Stack:" e "Banco:" em branco | perde dado |
| 3 | `CLAUDE.md` é reescrito por inteiro a cada `update`, inclusive na segunda execução: edição do usuário some | perde dado |
| 4 | O `/upgrade-sdd` promete job `pronto?` no `ci.yml`, `scripts/ci-politica.py`, `scripts/atestado.py` e campos de modo na Constituição; **nenhum código** os entrega. O `ci.yml` do usuário fica intacto (o CI antigo segue rodando em todo push, sem `pronto?`) | confunde |
| 5 | Campos de modo da Constituição (atestado, revisão humana, WIP, módulos, início automático) não existem no template: nem projeto novo os tem; só valem por padrão | confunde |
| 6 | Rules existentes e `sdd/.metrics/schema.json` não são atualizados (criar-se-não-existe): o projeto fica sem "Liberado para o CI", sem modos e sem o objeto `spec` do schema | confunde |
| 7 | O `forge doctor` marca "Hook instalado" e "Cobertura: claude" **sem checar o script**; e sem Claude avisa "rode forge update --upgrade", que nunca instala o hook (aviso circular); cita o job `pronto?`, que não existe | confunde |
| 8 | `update --dry-run` não lista o hook `PreToolUse`, nem os arquivos que serão reescritos; o log real também não confirma o hook | confunde |
| 9 | JSON inválido em `settings.json`/`.mcp.json` não é tocado (bom), mas o upgrade termina com código 0, grava a versão nova e deixa o doctor mandando rodar o mesmo comando; aviso do MCP duplicado | confunde |
| 10 | Skills `sdd-*`, prompts e comandos por agente são sempre regenerados; o `/upgrade-sdd` diz preservar o que é seu mas não avisa que edições nas `sdd-*` somem | confunde |
| 11 | `settings.json` é reserializado (chaves ordenadas): diff grande, sem perda | cosmético |

**Sem problema (verificado):** `update --upgrade` é idempotente; `settings.json` e `.mcp.json` do usuário (permissões, hooks próprios, servidores) são preservados e não duplicam; não ficam arquivos obsoletos; commands e agents batem com um `init` novo; o `ci.yml` do usuário não é alterado.

## Regras de negócio

- RN-1: nenhum comando do forge instala um hook cujo script não exista.
- RN-2: `update` nunca perde dado do usuário: stack e banco persistem; arquivo editado pelo usuário é preservado ou avisado.
- RN-3: tudo que a nota ou o `/upgrade-sdd` promete deve ser entregue por código ou **deixar de ser prometido**.
- RN-4: o `ci.yml` do usuário **nunca** é reescrito; o forge entrega um workflow **próprio** (arquivo novo) e documenta o que o usuário muda no dele.
- RN-5: `forge doctor` só diz "ok" depois de conferir o arquivo.

## Escopo

**Dentro:** os achados 1 a 10 e, como verificação na 118-07, os itens (a) a (e) abaixo. **Fora:** o achado 11 (cosmético), o `Deploy LP`, hooks para Gemini/Copilot/Codex (não têm hook), reescrever o `/upgrade-sdd` além do texto.

**Itens que a 118-07 só verifica (sem achado, por ora):** (a) `.forge/config/forge.json` e a versão de esquema dele no `update`; (b) `.gemini/` e os prompts e chatmodes do Copilot: reescritos como o `CLAUDE.md` (cobertos pela 118-02); (c) `sdd/.sdd-version` e o campo `version` do `.sddrc` mudam juntos (observado 5.0.0 para 5.2.0-beta; a 118-06 impede gravar a versão se o hook falhar); (d) migração de layout: o dry-run diz "nenhuma necessária" no 5.0 real, e a fixture da 118-07 prova; (e) `update` sem `--yes` em terminal não interativo dá erro de TTY: a 118-06 faz o comando orientar `--yes` em vez de falhar com mensagem do terminal.

## Riscos

- Propagar rules/Constituição mexe em conteúdo do usuário: só **acréscimo** com marcador, nunca reescrita.
- Entregar um workflow novo pode duplicar CI em quem já tem o seu: opt-in com `--ci`.
- A 14b emenda a `ARQ-06` e traz ADR-0027: exige aprovação do dono. **Esse aceite bloqueia as ondas seguintes que dependem do wrapper**; as features 118-01 a 118-05 não dependem dela e podem andar antes.
