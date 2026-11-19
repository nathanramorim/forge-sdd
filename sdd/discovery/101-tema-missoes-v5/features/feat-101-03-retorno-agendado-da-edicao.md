# feat-101-03 — Retorno agendado: o PR que encerra a edição

**Iniciativa:** 101-tema-missoes-v5
**Status:** ready
**Tamanho:** P
**Depende de:** fix-101-01 (marcadores nos READMEs) e feat-101-01 (camada e data de fim)

## Objetivo

Na virada da edição, um workflow abre sozinho o PR que devolve o README ao layout padrão e remove a camada da edição do site. O Nathan só mescla.

O visual do site não depende deste PR: a troca acontece na hora, pelo porteiro de data da feat-101-01. Este PR faz a limpeza e cuida do README, que é estático.

## Decisões registradas (Nathan, 19/11/2026)

- "Remover na segunda-feira dia 23/11/2026 0h, de forma automática voltar ao layout anterior."
- O README volta por **PR agendado**, e não por commit automático: a regra de não commitar direto no `main` continua valendo.

## Modelo do módulo

**Enxuto** (ARQ-11): um workflow, um script de encerramento e dois modelos de cabeçalho. Nenhum pacote Go novo.

## Como funciona

1. **Script de encerramento**, `.github/scripts/encerrar-edicao/encerrar.sh`. Recebe a pasta do repositório e:
   - apaga as linhas entre `edicao:inicio` e `edicao:fim` em `docs/index.html` e `docs/wiki.html`;
   - nos dois READMEs, troca o bloco marcado da edição pelo cabeçalho padrão, guardado em `README.cabecalho.md` e `npm.README.cabecalho.md`, ao lado do script;
   - apaga `docs/edicao/` e a própria pasta `.github/scripts/encerrar-edicao/`;
   - rodar duas vezes dá o mesmo resultado da primeira.
2. **Workflow**, `.github/workflows/encerrar-edicao.yml`: roda todo dia às 3h10 UTC (0h10 em Brasília) e também por disparo manual.
   - Lê a data de fim em `docs/edicao/edicao.js`. Não existe segunda cópia da data.
   - Sem esse arquivo, ou antes da data: termina sem fazer nada.
   - Depois da data: roda o script, envia a branch `chore/101-encerrar-edicao` e abre o PR. Se o PR já existir, não abre outro.
3. **O que fica depois do encerramento:** o design system (`docs/design/edicao-missoes/`), o banner da edição e a entrada da 5.0.0 no histórico, como registro. O arquivo do workflow também fica, inerte: o token do Actions não pode apagar workflows, e o corpo do PR avisa que ele pode ser removido à mão.
4. **Publicação.** O merge do PR muda `docs/` e dispara o `Deploy LP`, que publica o site já sem a camada.

O agendamento do GitHub pode atrasar alguns minutos. Isso só afeta a hora em que o PR aparece.

### Como o PR é aberto

Hoje a opção "Allow GitHub Actions to create and approve pull requests" está desligada no repositório (`can_approve_pull_request_reviews=false`). Com ela desligada, o Actions não consegue abrir PR.

**Decidido no Gate 1 (19/11/2026):** o workflow não depende dessa opção.

1. Envia a branch `chore/101-encerrar-edicao` e tenta abrir o PR com `gh pr create`.
2. Se o repositório recusar, abre uma issue com `gh issue create`, com o link que cria o PR em um clique.

Ligar a opção em Settings → Actions → General é opcional: com ela, o passo 1 já resolve.

## Fora do escopo

- A troca visual na virada (feat-101-01).
- Commit direto no `main`.
- Apagar o design system ou reescrever o histórico da versão.

## Decisões do projeto aplicadas

- **Regra 1 e FLX-02:** nada entra no `main` sem PR; o workflow só cria branch e PR.
- **Critério 6 da iniciativa** (`criteria.md`): a edição termina em 23/11/2026, 0h de Brasília.
- **COD-04:** a camada e o material de encerramento saem no mesmo PR.
- **Hospedagem:** a publicação continua pelo `Deploy LP`; nenhuma mudança no Worker.
- **Release:** `sdd/releases/history.md` não é alterado pelo encerramento.

## O que já existe

| Procurei | Onde | Achei | Decisão |
|---|---|---|---|
| Marcadores da edição | `README.md`, `npm/README.md` (fix-101-01), `docs/index.html`, `docs/wiki.html` (feat-101-01 e feat-101-02) | `edicao:inicio` e `edicao:fim` | reusar |
| Data de fim | `docs/edicao/edicao.js` (feat-101-01) | Parâmetro único | reusar |
| Cabeçalho padrão do README | `README.md` e `npm/README.md` antes do #113 (commit `24c1b51`), com `docs/img/banner.svg` e os selos azuis | Versão anterior do cabeçalho | reusar o texto, já com "Versão atual: 5.0.0" |
| Workflow agendado | `.github/workflows/` | `ci.yml`, `npm-publish.yml`, `release.yml`, `deploy-lp.yml`; nenhum com `schedule` | criar |
| Scripts de apoio a workflow | `.github/` | Não existem | criar em `.github/scripts/encerrar-edicao/` |
| Publicação depois do merge | `.github/workflows/deploy-lp.yml` | Dispara em mudanças de `docs/**` | reusar |

## Dependências novas

Nenhuma biblioteca. O workflow usa só `actions/checkout`, `git`, `node` e o `gh` que já vêm no runner.

## Rules e IDs da Constituição Aplicáveis

- **ARQ-11:** modelo enxuto, declarado acima.
- **DUP-02:** a data e os marcadores são reusados; nada é duplicado.
- **COD-01:** o script encerra a edição; o workflow só decide quando rodar.
- **COD-04:** camada, modelos e script saem no PR de encerramento.
- **COD-05:** textos em pt-BR, nível padrão.
- **SEG-01:** a data lida é validada; data inválida faz o workflow terminar sem agir.
- **SEG-04:** só o token do próprio Actions; nenhum segredo novo no repositório.
- **SEG-07:** nenhuma dependência nova.
- **TST-01 e TST-05:** `go test ./...` e `go vet ./...` no portão.
- **FLX-01 e FLX-02:** spec antes do código; nada direto no `main`.
- Rules: `.agents/rules/sdd-fluxo-e-portao.md`, `.agents/rules/sdd-seguranca.md`, `.agents/rules/sdd-codigo-e-testes.md`.

## Critérios de Aceitação Executáveis

```bash
base=$(git merge-base HEAD origin/main)
copia=$(mktemp -d)
git worktree add -q --detach "$copia" HEAD

# 1. Depois do script, não sobra nada da edição
bash .github/scripts/encerrar-edicao/encerrar.sh "$copia"
test ! -d "$copia/docs/edicao" && test ! -d "$copia/.github/scripts/encerrar-edicao"
test "$(grep -rl -E 'edicao:inicio|edicao:fim|data-edicao' "$copia/docs/index.html" "$copia/docs/wiki.html" "$copia/README.md" "$copia/npm/README.md" | wc -l | tr -d ' ')" = "0"

# 2. LP e wiki voltam a ser as de antes da edição
diff "$copia/docs/index.html" <(git show "$base:docs/index.html") >/dev/null
diff "$copia/docs/wiki.html" <(git show "$base:docs/wiki.html") >/dev/null

# 3. Os READMEs voltam ao cabeçalho padrão, na versão atual
grep -q 'docs/img/banner.svg' "$copia/README.md"
grep -q 'Versão atual: 5.0.0' "$copia/README.md"
grep -q 'docs/img/banner.png' "$copia/npm/README.md"
test "$(grep -c 'banner-v5-missoes' "$copia/README.md" "$copia/npm/README.md" | awk -F: '{s+=$2} END{print s}')" = "0"

# 4. O que deve ficar, fica
test -d "$copia/docs/design/edicao-missoes"
grep -q '^### Versão 5.0.0' "$copia/sdd/releases/history.md"

# 5. Rodar de novo na mesma cópia não muda mais nada
antes=$(git -C "$copia" status --porcelain | sort)
bash .github/scripts/encerrar-edicao/encerrar.sh "$copia"
test "$antes" = "$(git -C "$copia" status --porcelain | sort)"
git worktree remove --force "$copia"

# 6. Workflow: agenda diária, disparo manual, PR em vez de push no main, data vinda do parâmetro único
w=.github/workflows/encerrar-edicao.yml
grep -q 'schedule:' "$w" && grep -q 'workflow_dispatch' "$w" && grep -q 'gh pr create' "$w" && grep -q 'gh issue create' "$w"
grep -q 'docs/edicao/edicao.js' "$w"
test "$(grep -c -E 'push origin (HEAD:)?main' "$w")" = "0"
test "$(grep -c '2026-11-23' "$w" .github/scripts/encerrar-edicao/encerrar.sh | awk -F: '{s+=$2} END{print s}')" = "0"

# 7. Portão do projeto
go test ./... && go vet ./...
```

## Conferência (Gate 2)

- Disparo manual do workflow antes da data: termina sem alterar nada.
- Ensaio em um fork ou com a data adiantada em uma branch de teste: o PR abre com o resultado dos critérios 1 a 4.
- O corpo do PR explica o que sai, o que fica e que o arquivo do workflow pode ser apagado à mão.

## Status

- [x] Implementado em `feat-101-03` (19/11/2026): critérios 1 a 7 (incluindo ensaio em worktree temporário descartável e idempotência), validação YAML/bash do workflow, `go test ./...`, `go vet ./...` e `forge gate --mode bloqueio` aprovados.
