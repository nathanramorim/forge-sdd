# Fix 61 — Correção de Sobreposição no Handoff Bar da LP e Atualização do Link do Instagram (`@nathan.ramorim`)

## Modelo do módulo
- **Camada:** Apresentação Web / Landing Page (`docs/showcase-v3-lp.html` e `docs/index.html`) + Especificação do Site Estático (`sdd/discovery/*-60-site-estatico-seo-performance.md`).
- **Problema 1 (Layout / Responsividade):** Na seção `#estacoes` ("Visão Processual entre os Agentes"), a barra `.handoff-bar` utiliza `grid-template-columns: minmax(260px, 1.05fr) minmax(240px, 1.35fr) auto;`. Quando a transição chega à etapa `04. Builders ── 05. Governança Read-Only`, a 3ª coluna (`.handoff-gate` com largura `auto`) se expande para acomodar a frase longa `"Só avança para o Analista de Governança quando go test e P-13 passam"`, comprimindo a 1ª coluna para `260px`. Como `.handoff-route` possui `flex-wrap: nowrap` e largura real de `~344px`, a pílula `05. Governança Read-Only` transborda por cima da 2ª coluna (`Artefato em trânsito (Handoff)`).
- **Problema 2 (Link Social):** O link do Instagram no Header e no Footer aponta para `https://www.instagram.com/nathanramorim` em vez do perfil oficial com ponto: `https://www.instagram.com/nathan.ramorim` (`@nathan.ramorim`).

## O que já existe
- `docs/showcase-v3-lp.html` e `docs/index.html` — Landing Page oficial da v3.0 com seção `#estacoes` e links sociais `.ds-link`.
- `sdd/discovery/{discovery,criteria,plan}-60-site-estatico-seo-performance.md` — Especificação da migração 1:1 para `/site`.

## Dependências novas
Nenhuma.

## Rules e IDs da Constituição Aplicáveis
- `COD-01` / `COD-02` — Simplicidade, zero regressão visual e código limpo.
- `DUP-01` — Manter `docs/showcase-v3-lp.html` e `docs/index.html` 100% sincronizados.

## Critérios de Aceitação Executáveis
- [x] **CA-01 (Sem Sobreposição em `.handoff-bar`):** Em qualquer uma das 5 etapas do fluxo (`01 → 02`, `02 → 03`, `03 → 04`, `04 → 05`, `05 ↺ 01 / 04`) e em qualquer largura de viewport (`320px` a `1920px+`), `.handoff-route` nunca transborda sobre `.handoff-payload` (`Artefato em trânsito (Handoff)`).
- [x] **CA-02 (Link Oficial do Instagram):** Todos os links de Instagram no Header e no Footer de `docs/showcase-v3-lp.html`, `docs/index.html` e nos documentos do Discovery 60 apontam para `https://www.instagram.com/nathan.ramorim` (`@nathan.ramorim`).
