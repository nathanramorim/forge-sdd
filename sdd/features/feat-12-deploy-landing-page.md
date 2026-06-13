# feat/npm-release-and-deploy

**Branch:** `feat/npm-release-and-deploy`
**Fase:** 12
**Depende de:** `feat-10-landing-page`, `feat-11-unify-agent-structures`
**Status:** `todo`

## Objetivo

Consolidar o processo de release e deploy do projeto, incluindo a publicação automática no NPM com a nova versão e o deploy final da Landing Page.

## Critério de conclusão

```bash
# 1. Verificar se a versão no NPM condiz com a última tag
npm view @nathanramorim/forge-sdd version

# 2. Validar deploy da landing page
curl -I https://forge-sdd.vercel.app | grep "HTTP/2 200"

# 3. Validar CI/CD
# Verificar se o workflow .github/workflows/npm-publish.yml passou
```

## Tarefas

- [ ] **12-1** Atualizar `npm/package.json` para a versão estável atual (v1.3.1 ou superior)
- [ ] **12-2** Revisar `.github/workflows/npm-publish.yml` para garantir que o deploy ocorra apenas em tags
- [ ] **12-3** Finalizar o conteúdo da landing page em `site/` (SEO, links para npm e brew)
- [ ] **12-4** Configurar o deploy na Vercel para monitorar apenas o diretório `site/` (se aplicável)
- [ ] **12-5** Realizar o push da tag e monitorar os deploys (NPM + Vercel)

## Arquivos gerados/modificados

```
npm/package.json
.github/workflows/npm-publish.yml
site/
```

## Skills relevantes

- `orquestrador.chatmode.md` (Release flow)
- `builder.chatmode.md` (NPM/Vercel)
