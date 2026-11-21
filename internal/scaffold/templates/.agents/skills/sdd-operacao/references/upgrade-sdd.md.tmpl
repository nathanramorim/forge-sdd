# Comando: upgrade-sdd e Migrator

**Uso:** `/upgrade-sdd <versão>` atualiza a estrutura do projeto preservando o domínio.

1. Leia `sdd/.sdd-version` e confirme a versão alvo com o usuário.
2. `which forge-sdd`: se estiver em cache do npm, oriente `rm -rf ~/.cache/forge-sdd` e `npx @nathanramorim/forge-sdd@latest`.
3. Rode `forge update --dry-run`, apresente o que seria criado, movido ou removido e só então `forge update`.
4. **Preservação de domínio:** nunca sobrescreva `sdd/memory/constitution.md` fora de blocos demarcados, iniciativas em `sdd/discovery/`, `sdd/spec/` ou `.agents/rules/` customizadas.
5. Liste arquivos criados, atualizados e removidos e sugira `/doctor`.

## A política de PR em draft vai junto (feat-106-36)

O `/upgrade-sdd` é o **canal de entrega** da política de PR (ADR-0022, ADR-0023). Ao atualizar a
estrutura, ele leva as peças novas do fluxo para o projeto existente:

| Peça | O que instala | Onde |
|---|---|---|
| **Hook `PreToolUse`** | barreira de PR (`scripts/pr_guard.py`) que bloqueia `gh pr ready` e `gh pr merge` sem parecer `conforme` para o HEAD | `.claude/settings.json` |
| **Job `pronto?`** | recusa PR em draft na branch principal | `.github/workflows/ci.yml` |
| **Scripts de política** | `ci-politica.py` (custo do CI) e `atestado.py` (veredito no CI) | `scripts/` |
| **Campos de modo** | `Modo do atestado`, `Modo de revisão humana`, `WIP humano`, `Módulos permitidos`, `Início automático do Builder` | `sdd/memory/constitution.md` |

### Regras do upgrade para o CI e o hook

1. **O `ci.yml` do user é preservado.** O forge **não sobrescreve** o workflow existente: ele
   acrescenta só o job `pronto?` e, quando não houver nenhum `if` de draft, avisa o que falta e
   mostra o diff. Passos customizados do projeto nunca são perdidos.
2. **O `settings.json` do usuário é mesclado, nunca substituído.** O hook entra sem duplicar e sem
   remover permissões, modelos ou outros hooks. JSON inválido **não é tocado**: o upgrade avisa e
   pede a adição manual.
3. **Os campos de modo entram na Constituição só se ainda não existirem.** Valores já escolhidos
   pelo dono são preservados; uma mudança de valor exige ADR (`GOV-01`).
4. **A prévia é obrigatória.** Rode `forge update --dry-run` primeiro e mostre a lista; o upgrade
   só grava depois do OK do dono.

### Ao terminar

Liste o que mudou por categoria (hook, workflow, scripts, Constituição) e rode `forge doctor`: a
seção **Barreira de PR** informa se o hook está instalado, se o `uv` está no PATH e em quais
agentes a barreira vale (só o Claude Code executa hooks; nos outros, o job `pronto?` segura).