# feat/upgrade-node-ci

**Branch:** `feat/upgrade-node-ci`
**Fase:** 27
**Depende de:** `feat-08-npx`
**Status:** `todo`

## Objetivo
Atualizar as configurações de workflows do GitHub Actions para mitigar os avisos e erros de depreciação do Node 20. Isso é feito atualizando o setup do Node.js de `'20'` para `'24'` no workflow de publicação NPM.

## Critério de conclusão
```bash
grep -q "node-version: '24'" .github/workflows/npm-publish.yml
# → O comando acima deve retornar status 0 (sucesso), garantindo que a versão foi alterada para Node 24.
```

## Tarefas
- [ ] **27-1** Modificar a versão do Node no setup de `.github/workflows/npm-publish.yml` de `'20'` para `'24'`.
- [ ] **27-2** Opcional: Adicionar a variável de ambiente `ACTIONS_ALLOW_USE_UNSECURE_NODE_VERSION: 'true'` se houver quebra persistente com pacotes internos obsoletos de terceiros.

## Arquivos gerados/modificados
```
.github/workflows/npm-publish.yml
```

## Skills relevantes
- `Specifier`
- `Builder`
