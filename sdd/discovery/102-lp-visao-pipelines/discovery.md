# Discovery 102 — LP com a visão de pipelines

## Problema
A landing page (`forge-sdd.org`) ainda vende o produto como "5 especialistas em esteira" (Líder, PM e Arquiteto, Planejador, Builders, Governança), texto da v3. O produto atual, descrito no README e em `docs/pipeline.md`, são 3 sessões de pipeline: **Discovery** (você + agente), **Builder** (worker) e **Reviewer** (worker, somente leitura). Quem chega na LP vê uma visão que não existe mais e que contradiz o README.

## Decisão do dono (Nathan, 19/11/2026)
"Agora é só pipeline; aqui no site tá a visão completa." A LP passa a mostrar só o pipeline de 3 sessões, e a visão das 5 estações sai.

## Para quem
Desenvolvedor que acabou de chegar no forge-sdd e precisa entender, em minutos, como o trabalho flui: você planeja e aprova, uma IA constrói em paralelo e outra confere.

## Escopo
- Título, hero e a esteira interativa (seção 1), com os cartões e os mascotes 3D das 3 sessões.
- Emulador e guia de comandos (seções 2 e 3) com o fluxo e os comandos do pipeline.
- Governança, comparativo "antes e agora" e texto do vídeo (seções 4 a 6).

**Fora:** o tema da Edição Missões (iniciativa 101, temporária), a wiki, o README, o vídeo em si, o Go.

## Regras
- Conteúdo permanente: não pode depender da camada `docs/edicao/` nem sumir com ela.
- Cada afirmação do texto vem do README, de `docs/pipeline.md` ou do código; sem promessa nova.
- Os ids das seções (`#estacoes`, `#emulador`, `#instalacao`, `#governanca`, `#evolucao`, `#video`) e as âncoras do menu não mudam, para não quebrar links nem a camada da edição.

## Riscos
- A Edição Missões lê a ordem dos cartões da seção 1. Ao reduzir de 5 para 3 cartões, a camada da 101 precisa ser ajustada (feito na própria 101, depois que a 102 entrar no `main`).
- O texto novo é conteúdo de produto: precisa da revisão do dono.

## Dono
Nathan. Revisão do texto antes do merge.
