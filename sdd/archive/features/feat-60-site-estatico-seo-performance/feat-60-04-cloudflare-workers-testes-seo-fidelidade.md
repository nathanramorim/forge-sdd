# feat-60-04 — Configuração Cloudflare Workers (`wrangler.jsonc` + `_headers`), Suite de Testes de Fidelidade/SEO (`npm test`) e Documentação

**Branch:** `feat/60-site-estatico-seo-performance`
**Fase:** 60-04
**Tamanho:** `P`
**Depende de:** `feat-60-03-hidratacao-interacoes-motor-3d-otimizado`
**Status:** `done`

## Objetivo

Configurar o deploy de borda na Cloudflare via **Cloudflare Workers com Static Assets** (`site/wrangler.jsonc` apontando para `./dist` e rotas `forge-sdd.org` / `www.forge-sdd.org`) e cabeçalhos HTTP de cache imutável e segurança (`site/public/_headers`), criar a suite de verificação automatizada determinística (`site/tests/verify-build.mjs` executada por `npm --prefix site test`) cobrindo todos os critérios `CA-01` a `CA-10` do `criteria-60`, e documentar o diretório `/site` em `sdd/spec/concepts.md` e `README.md`.

## Modelo do módulo
- **Modelo:** `enxuto` (configuração declarativa de borda em `site/wrangler.jsonc` e `site/public/_headers`, script de teste determinístico em `site/tests/verify-build.mjs` e atualização documental).
- **Justificativa:** Fecha o ciclo de entrega da aplicação estática garantindo cache imutável para os bundles hashados (`/_astro/*`), cabeçalhos de segurança HTTP e verificação automatizada executável antes do portão `forge-sdd gate`.

## O que já existe

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
| --- | --- | --- | --- | --- |
| Artefato de referência 1:1 da LP com Fix 61 (`.handoff-bar`) e link oficial do Instagram (`https://www.instagram.com/nathan.ramorim`) | `docs/showcase-v3-lp.html` e build gerado em `site/dist/` (`feat-60-01..03`) | Tokens CSS, textos das 5 seções, links sociais, SEO Head e bundles gerados pelo Astro | `reusar` | Validar programaticamente em `site/tests/verify-build.mjs` que `site/dist/index.html` preserva 100% dos invariantes visuais, textuais e de SEO de `docs/showcase-v3-lp.html` (`DUP-01`, `DUP-02`) |
| Catálogo de Conceitos e README principal | `sdd/spec/concepts.md` e `README.md` | Documentação da arquitetura v3.0 e estrutura do repositório | `estender` | Documentar a fronteira isolada `/site` (`https://forge-sdd.org`) sem alterar nenhuma regra do binário Go |
| Configuração Cloudflare Workers Static Assets e suite de verificação | `site/wrangler.jsonc`, `site/public/_headers`, `site/tests/verify-build.mjs` | Não existem no repositório | `criar` | Criar arquivos de configuração Cloudflare para `forge-sdd.org` e suite de teste executável via `npm --prefix site test` |

## Dependências novas
Nenhuma (`SEG-07` — `site/tests/verify-build.mjs` utiliza apenas `node:fs`, `node:path` e `node:assert` nativos do Node.js).

## Rules e IDs da Constituição Aplicáveis
- **Rules:** `.agents/rules/sdd-seguranca.md`, `.agents/rules/sdd-codigo-e-testes.md`, `.agents/rules/sdd-fluxo-e-portao.md`
- **IDs:** `SEG-04`, `DUP-01`, `TST-01`, `TST-04`, `TST-05`, `FLX-01..06`
- **Skills Reutilizadas:** —

## Arquivos Tocados
- `site/wrangler.jsonc`
- `site/public/_headers`
- `site/tests/verify-build.mjs`
- `site/package.json`
- `sdd/spec/concepts.md`
- `README.md`

## Critérios de Aceitação Executáveis

1. **CA-09 (Configuração Pronta para Cloudflare Workers `forge-sdd.org` e `_headers`):**
   ```bash
   test -f site/wrangler.jsonc && \
   grep -q '"forge-sdd.org"' site/wrangler.jsonc && \
   npm --prefix site run build && \
   test -f site/dist/_headers && \
   grep -q 'max-age=31536000, immutable' site/dist/_headers && \
   grep -q 'X-Content-Type-Options: nosniff' site/dist/_headers
   ```
2. **CA-10 (Suite Automatizada de SEO, De-Para 1:1, Performance e Portão Determinístico `forge-sdd gate`):**
   ```bash
   npm --prefix site test && \
   go test ./... && \
   go vet ./... && \
   go run ./cmd/forge-sdd gate
   ```

## Handoff
Conclui a onda de implementação da Feature 60 (`feat/60-site-estatico-seo-performance`) e habilita a auditoria final da Estação 05 (Analista de Governança).
