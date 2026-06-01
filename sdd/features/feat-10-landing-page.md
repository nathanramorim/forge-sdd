# feat/landing-page

**Branch:** `feat/landing-page`
**Fase:** 10
**Depende de:** `feat/multi-agent` (ou pode ser paralela — sem dependência de código Go)
**Status:** `todo`

## Objetivo
Criar uma página estática de apresentação do **forge-sdd**, publicada via **Vercel** com deploy automático a cada push em `main`.

A página é 100% estática (HTML + CSS + JS vanilla ou framework leve como Astro) — sem backend, sem banco.

## Critério de conclusão
```
https://forge-sdd.vercel.app  (ou domínio customizado)
→ página carrega em < 2s
→ deploy automático funcionando no push em main
→ lighthouse performance > 90
```

## Estrutura

```
site/
  index.html          (ou src/index.astro se usar Astro)
  public/
    og-image.png      → imagem Open Graph
  src/
    styles.css
    demo.js           → animação do terminal "forge-sdd init"
vercel.json           → configuração de deploy (output dir, build command)
```

## Conteúdo da página

### Seções
1. **Hero** — tagline + comando `npx @nathanramorim/forge-sdd@latest init` copiável + botão "Ver no GitHub"
2. **Demo animado** — terminal fake mostrando o `forge-sdd init --yes --dry-run` com os 32 arquivos aparecendo
3. **O que é gerado** — grid com os 4 blocos: `sdd/memory`, `sdd/spec`, `.github/chatmodes`, `.vscode/`
4. **Instalação** — 3 abas: npx / Homebrew / Download direto
5. **Agentes suportados** — GitHub Copilot (atual), Claude, Gemini (em breve — feat-09)
6. **Footer** — MIT License · GitHub · npm

### Stack recomendada
**Opção A — Zero build (mais simples):** HTML + CSS + JS vanilla em `site/`
**Opção B — Astro (recomendada):** componentes isolados, markdown para conteúdo, bundle otimizado

## Tarefas
- [ ] **10-1** Decidir stack (vanilla vs Astro) e inicializar estrutura em `site/`
- [ ] **10-2** Criar layout base: hero + navegação
- [ ] **10-3** Implementar terminal demo animado (CSS keyframes ou JS)
- [ ] **10-4** Criar seção "O que é gerado" com grid de arquivos
- [ ] **10-5** Criar seção de instalação com tabs (npx / brew / download)
- [ ] **10-6** Criar `vercel.json` + conectar repositório ao Vercel
- [ ] **10-7** Configurar deploy automático: `main` → produção, PRs → preview
- [ ] **10-8** Adicionar meta tags Open Graph e link canônico
- [ ] **10-9** Testar Lighthouse (performance, acessibilidade, SEO > 90)
- [ ] **10-10** Adicionar link da página no README principal

## Arquivos gerados / alterados
```
site/
  (estrutura conforme stack escolhida)
vercel.json
README.md        (link da página)
```

## Notas
- `site/` é ignorado pelo goreleaser e pelo wrapper npm — sem impacto nas releases Go
- O deploy na Vercel usa o repositório `nathanramorim/forge-sdd` (pode ser privado — Vercel tem acesso via OAuth)
- Para domínio customizado futuro: registrar `forge-sdd.dev` ou `forgesdd.dev`

## Skills relevantes
(consultar `skills/index.md`)
