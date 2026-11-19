# Discovery — 99-wiki-get-started

## Problema

A landing page do forge-sdd (`docs/index.html`) não tem uma rota `/wiki` para novos usuários. Quem chega no site não consegue seguir uma jornada guiada de instalação e uso. Além disso, a documentação é escrita só para humanos — não tem prompts prontos para IA que o usuário possa copiar e colar no seu agente.

## Para quem

**Usuário primário:** desenvolvedor solo ou pequeno time que acabou de descobrir o forge-sdd e quer começar agora.

**Usuário secundário:** IA (Claude Code, Gemini, Copilot) lendo a wiki para entender como instalar e configurar o forge-sdd num projeto, sem precisar que o humano re-explique.

## Jornada atual (problema)

1. Usuário acessa forge-sdd.org.
2. Lê a landing page, vê features e emulador.
3. CTA leva para `#instalacao` (âncora dentro da LP).
4. Não há progressão clara: "instalei, e agora?".
5. Não há página de hub mode (usar em vários repos).

## Jornada ideal (solução)

1. Usuário acessa forge-sdd.org.
2. CTA primário "Get Started" abre `/wiki`.
3. Wiki mostra: instalação → primeira feature → jornada completa.
4. Cada passo tem um bloco "Prompt para IA" para copiar e colar no agente.
5. Seção Hub: como transformar o forge-sdd num hub de vários repos.

## O que é um "Prompt para IA"

Bloco especial na wiki (estilo AWS/Microsoft):
```
╭─ Prompt para IA  [copiar] ─────────────────────────────────╮
│  Instale o forge-sdd neste projeto:                         │
│  npm install -g @nathanramorim/forge-sdd && forge init      │
│  Me diga quando estiver pronto.                             │
╰────────────────────────────────────────────────────────────╯
```
O usuário copia o bloco inteiro e cola no seu agente. A IA sabe exatamente o que fazer.

## Estado atual da LP

A landing page publicada em **forge-sdd.org** vem do repo `nathanramorim/lp-forge-sdd` (que será desativado). O arquivo `docs/index.html` neste repo está desatualizado. A LP nova (publicada) tem:
- Hero: "Você decide numa sessão. Outras duas constroem e revisam sozinhas."
- CTAs: "Ver o pipeline funcionando", "Como Instalar & Usar" (CTA primário), "Guia do pipeline"
- Pill de instalação com botão "Copiar"

A wiki nova ficará neste repo (forge-sdd) em `docs/wiki.html`. A LP de referência para o design system é a versão publicada do `lp-forge-sdd`.

## Regras de negócio

1. Wiki fica em `docs/wiki.html` neste repo (forge-sdd), acessível em `/wiki`.
2. Stack: HTML/CSS/JS puro, design system idêntico ao da LP publicada em forge-sdd.org.
3. Linguagem simples: sem jargão técnico no Get Started.
4. Cada seção tem versão pt-BR (padrão) e en-US.
5. Prompts para IA são copiáveis com um clique.
6. Hub mode: como um único `forge init` pode rastrear múltiplos repos.

## Escopo

**Dentro:**
- `docs/wiki.html` (nova página, design do lp-forge-sdd)
- `docs/wiki.css` (estilos da wiki)
- CTA "Get Started →" ou troca do CTA primário da LP publicada → `/wiki`
- Conteúdo: Get Started (3 seções) + Hub mode
- i18n: pt-BR + en-US

**Fora:**
- Migrar a LP completa de lp-forge-sdd (iniciativa separada)
- Backend ou SSR
- Autenticação ou busca
- Mais de 8 páginas no total
- Mobile app

## Riscos

- **Escopo cresce:** wiki "enxuta" vira doc completo. Mitigação: máximo de 8 páginas no lançamento.
- **i18n manual é onerosa:** textos em dois idiomas duplicam o trabalho. Mitigação: um objeto JS de strings por idioma, igual ao existente na lp-forge-sdd.
