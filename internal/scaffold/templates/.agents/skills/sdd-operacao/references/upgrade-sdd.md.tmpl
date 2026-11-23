# Comando: upgrade-sdd e Migrator

**Uso:** `/upgrade-sdd <versão>` atualiza a estrutura do projeto preservando o domínio.

1. Leia `sdd/.sdd-version` e confirme a versão alvo com o usuário.
2. `which forge-sdd`: se estiver em cache do npm, oriente `rm -rf ~/.cache/forge-sdd` e `npx @nathanramorim/forge-sdd@latest`.
3. Rode `forge update --dry-run`, apresente o que seria criado, movido ou removido e só então `forge update`.
4. **Preservação de domínio:** nunca sobrescreva `sdd/memory/constitution.md` fora de blocos demarcados, iniciativas em `sdd/discovery/`, `sdd/spec/` ou `.agents/rules/` customizadas.
5. Liste arquivos criados, atualizados e removidos e sugira `/doctor`.

## O que o upgrade entrega (e o que fica com você)

O `/upgrade-sdd` é o **canal de entrega** da política de PR (ADR-0022, ADR-0023). O forge só promete
o que o código entrega:

| Peça | O que o forge faz | Onde |
|---|---|---|
| **Hook `PreToolUse`** | `forge init` e `forge update` mesclam o hook (`forge hook pr-guard`), que bloqueia `gh pr ready` e `gh pr merge` sem parecer `conforme` para o HEAD | `.claude/settings.json` |
| **Script `pr_guard.py`** | `forge init` e `forge update` gravam o script; sem o marcador `# forge-sdd:generated` ele é seu e fica intocado | `scripts/pr_guard.py` |
| **Campos de modo** | entregues pela feat-118-03: `Modo do atestado`, `Modo de revisão humana`, `WIP humano`, `Módulos permitidos`, `Início automático do Builder`, só se ainda não existirem | `sdd/memory/constitution.md` |
| **Workflow `forge-pronto.yml`** | **só com `--ci`**: `forge update --ci` (ou `forge init --ci`) cria o arquivo novo com o job `pronto?`, que recusa PR em draft; sem `--ci` nada é criado | `.github/workflows/forge-pronto.yml` |

### O que o forge NÃO faz

1. **Não lê nem altera o seu `ci.yml`.** O que mudar nele é decisão sua: se quiser a recusa de PR em
   draft no seu CI, copie o job `pronto?` do `forge-pronto.yml` ou use o arquivo como está.
2. **Não entrega `ci-politica.py` nem `atestado.py`.** Se o seu CI usa scripts assim, eles são seus.
3. **Um `forge-pronto.yml` seu** (sem o marcador `# forge-sdd:generated`) é preservado, com aviso.
4. **O `settings.json` é mesclado, nunca substituído.** JSON inválido não é tocado: o upgrade avisa e
   pede a adição manual.
5. **Edição nas skills `sdd-*` é sobrescrita** a cada upgrade; personalize em `.agents/rules/` ou
   em skills próprias, que o forge preserva.

Os valores dos campos de modo já escolhidos pelo dono são preservados; mudar um valor exige ADR (`GOV-01`).
A prévia é obrigatória: rode `forge update --dry-run` (ou `forge update --ci --dry-run`) e mostre a
lista; o upgrade só grava depois do OK do dono.

### Ao terminar

Liste o que mudou por categoria (hook, script, workflow, Constituição) e rode `forge doctor`: a
seção **Barreira de PR** informa se o hook está instalado, se o `uv` está no PATH e em quais
agentes a barreira vale (só o Claude Code executa hooks; nos outros, vale o job `pronto?` se você
criou o workflow com `--ci`).
