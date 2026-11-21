# Discovery — 112-lp-estavel-beta-e-estilo-forge

Pedido do dono (21/11/2026): "aproveita a versão beta para colocar na LP a estável e a beta, e muda o estilo de volta para o do Forge, sem ser do GTA".

## Descrição de negócio
- **Problema:** a LP ainda veste a Edição Missões (tema temporário que acaba sozinho em 23/11/2026 00:00 -03:00) e mostra só a v5.0.0, embora a 5.1.0-beta já esteja no npm (canal `beta`, desde 21/11). Quem chega não sabe que existe beta nem como instalá-la.
- **Para quem:** quem visita forge-sdd.org para instalar; e o dono, que hoje precisa editar o HTML a cada release.
- **Valor:** a LP volta à cara do Forge 2 dias antes do prazo e passa a mostrar sozinha as versões estável e beta.

## Decisões do dono (menu, 21/11/2026)
1. **Estilo:** antecipar o encerramento oficial da Edição Missões com o mecanismo que já existe. Resultado: LP, README e npm/README no estilo do Forge, sem o tema.
2. **Versões:** ler o npm ao vivo no navegador (automático), com fallback estático `5.0.0` e `5.1.0-beta`. O dono sabe que isso é uma chamada externa nova.

## Fatos conferidos (21/11/2026)
- `npm view @nathanramorim/forge-sdd dist-tags --json` -> `{"beta":"5.1.0-beta","latest":"5.0.0"}`.
- O endpoint `.../-/package/<pkg>/dist-tags` responde 200 mas **sem `access-control-allow-origin`** (testado com `Origin`): o navegador bloquearia. Os endpoints `https://registry.npmjs.org/@nathanramorim/forge-sdd/latest` e `/beta` trazem `access-control-allow-origin: *`; o corpo tem `version` (1,6 KB cada).
- A LP **já** faz chamadas externas: three.js em `cdnjs.cloudflare.com` (`index.html:10`) e fontes do Google. Não há CSP em `docs/` nem no `site/wrangler.jsonc`; `site/public/_headers` não traz CSP (e o Worker serve `../docs`, então esse arquivo nem é o ativo). Nenhum `connect-src` precisa mudar hoje.
- Encerramento hoje: workflow `encerrar-edicao.yml` (cron diário 03:10 UTC) só age se `e.vigente(Date.now())` for falso (`docs/edicao/edicao.js`, `fim`); então roda `encerrar.sh`, cria `chore/101-encerrar-edicao`, PR via `GITHUB_TOKEN` (cai em issue se o PR falhar). Não commita em `main`. Simulado numa cópia: `encerrar.sh` roda com rc=0.
- Não existe nenhum teste Node na LP; os testes Go leem `docs/` como texto (`TestScaffold_DocsEdicao`).

## Decisões abertas para o dono
- **D-1 Como antecipar:** recomendado rodar `encerrar.sh` localmente numa branch e abrir PR (não alterar a data do `edicao.js` nem disparar o workflow, que só age com a vigência vencida e usa o nome `chore/101-...`).
- **D-2 Design system e banner da edição** (`docs/design/edicao-missoes/`, `docs/img/banner-v5-missoes.*`): o script os mantém. Recomendado manter (o histórico da 5.0.0 os cita); a LP não os usa mais. Se o dono quiser remover, o `grep` de prova perde as exceções.
- **D-3 Workflow inerte:** apagar `.github/workflows/encerrar-edicao.yml` no mesmo PR (o commit local pode; só o token do Actions não pode). Recomendado apagar.
- **D-4 Terminal de exemplo** (`v5.0.0` em ~1253 e ~2085): recomendado seguir a estável (preenchido pelo JS); o beta não entra no exemplo.
- **D-5 Wiki:** só a LP agora; a wiki (`docs/wiki.html`) como seguimento.
- **Segurança:** nenhum cabeçalho muda. Se o dono pedir CSP no futuro, `connect-src https://registry.npmjs.org` é obrigatório; fica registrado, não feito.

## Riscos
- Deploy da Cloudflare falha (`No access to the specified resource`): a LP nova só vai ao ar depois do token corrigido (publicação, não construção).
- Registro fora do ar ou lento: o fallback estático cobre (limite de 3 s).
- Resposta adulterada/inesperada: regex ancorada e `textContent`; o resto é descartado.
- `beta` ausente no npm no futuro: o chip Beta fica com o valor estático; avaliar esconder (ver feature 02).
- O README/npm/README mostram "Versão atual: 5.0.0" no cabeçalho padrão restaurado: fora do escopo (só a LP lê ao vivo).

## Nota para o checklist de release
Incluir "conferir a LP": ela lê o npm ao vivo, então só os valores de fallback (`5.0.0`, `5.1.0-beta`) precisam acompanhar a release. Não editado aqui (`sdd/releases/history.md` e o checklist ficam para a implementação/dono).
