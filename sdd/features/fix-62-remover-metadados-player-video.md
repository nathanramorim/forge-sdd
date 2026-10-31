# Fix 62 — Remover Nome de Arquivo e Menções de Geração do Player de Vídeo

## Modelo do módulo
- **Camada:** Apresentação Web (`site/src/components/VideoSection.astro`, `docs/index.html` e `docs/showcase-v3-lp.html`).
- **Problema:** Na Seção 06 (`#video`), a barra `.video-meta-bar` abaixo do player exibia o nome do arquivo (`forge-sdd-v3-demo.mp4`) e a atribuição `"Apresentação gerada com BRAG"`. A diretriz de design exige manter apenas o player de vídeo limpo, sem metadados textuais sobre o arquivo ou a ferramenta geradora.
- **Solução:** Remover o bloco `.video-meta-bar` (e classes auxiliares não utilizadas) de `site/src/components/VideoSection.astro`, `docs/index.html` e `docs/showcase-v3-lp.html`, preservando o container responsivo, cantos arredondados, poster e controles nativos HTML5.

## O que já existe
- `site/src/components/VideoSection.astro` — Componente Astro da seção de vídeo.
- `docs/index.html` e `docs/showcase-v3-lp.html` — Páginas estáticas espelhadas.
- `site/tests/verify-build.mjs` — Suite determinística de build do site.

## Dependências novas
Nenhuma.

## Rules e IDs da Constituição Aplicáveis
- `COD-01` / `COD-02` — Simplicidade e código limpo.
- `DUP-01` — Manter `site/`, `docs/index.html` e `docs/showcase-v3-lp.html` sincronizados.
- `TST-01` / `TST-05` — Testes automatizados passando.

## Critérios de Aceitação Executáveis
- [x] **CA-01 (Remoção dos Metadados Textuais do Vídeo):** Nenhuma menção a `forge-sdd-v3-demo.mp4` ou `BRAG` / `gerada com BRAG` é exibida no DOM da Seção 06.
- [x] **CA-02 (Player de Vídeo Limpo e Funcional):** O elemento `<video>` permanece 100% funcional com `controls`, `poster`, `playsinline` e container estilizado com cantos arredondados.
- [x] **CA-03 (Build, Testes e Gate Conformes):** `npm --prefix site test`, `go test ./...` e `forge-sdd gate` executam com sucesso.
