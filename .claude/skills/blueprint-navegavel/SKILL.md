---
name: blueprint-navegavel
description: Gera o BLUEPRINT NAVEGÁVEL da aplicação inteira, um único HTML autocontido (CSS e JS embutidos) onde o usuário clica e vê como o produto funciona, com os diagramas Mermaid de apoio à parte. Use sempre que pedirem blueprint, protótipo navegável, wireframe clicável, "ver a aplicação inteira", "como consigo ver isso?" ou maquete do produto, em discovery ou feature de UI. Não use para diagrama isolado (use diagramas-c4-mermaid) nem para mapa de backend em páginas (use mapa-de-arquitetura-em-paginas).
---

# Blueprint navegável

Modelo de referência: `forge-on-desktop/sdd/spec/wireframes/escritorio.html` (discovery 19e7), aprovado pelo usuário em duas rodadas.

## Princípios (correções que o usuário já fez)
1. **UM só navegável.** Nunca vários blueprints soltos. É a ÚNICA aplicação: todas as telas, menus e fluxos numa tela só.
2. **Janelas sobre uma cena**, não páginas. Barra mínima lateral e um seletor único de contexto (ex.: escritório) abrem janelas que empilham e fecham com `Esc`.
3. **Cada janela responde uma pergunta** do usuário (o que está acontecendo, quem faz o quê, em que etapa está). Escreva a pergunta no título ou subtítulo.
4. **Navegação em no máximo 2 passos, sem beco sem saída.** Toda tela tem caminho de volta.
5. **Chat natural** em vez de painéis cheios de botões; vidro escuro translúcido, texto claro legível.
6. **Só ícones SVG de traço apagado. Nunca emoji.** Botões pequenos só com ícone e dica (`title`/tooltip).
7. **Seleção única e clara** do contexto ativo; nomes de domínio, não técnicos (ex.: Projetos → Escritórios).
8. **Portas, não acoplamentos:** o que pode ser trocado no futuro (ex.: provedor de IA) aparece como interface (`AssistantProvider`), não fixo na tela.
9. Tipografia **Montserrat**; cores só por variáveis CSS (`--bg`, `--panel`, `--primary`, ...), com tema claro e escuro.

## Processo
1. **READ-MIN:** `sdd/memory/progress.md`, `constitution.md`, o discovery e os critérios da iniciativa, `sdd/spec/decisions.md` e as telas/menus já existentes. Leia nessa ordem e liste as decisões que o blueprint precisa respeitar.
2. **PLAN (≤ 500 tokens):** liste janelas, a pergunta de cada uma, o grafo de navegação e a cena. Aguarde confirmação (CLAUDE.md).
3. **Gere por script a partir de um template** (ex.: `sdd/spec/wireframes/escritorio.html` gerado de um template), para poder regerar. Salve em `sdd/spec/wireframes/`; imagens em `wireframes/img/`.
4. **Diagramas Mermaid à parte** em `sdd/discovery/diagramas-<id>/` (`.mmd` + `.png`): container, fluxo clique→resultado, módulos e navegação. Use a skill `diagramas-c4-mermaid` e confira o PNG renderizado.
5. **Verifique:**
   - Abra no navegador (Chrome) e capture telas das janelas principais em `evidencias-<id>/`.
   - Teste de fidelidade (ex.: `tests/fidelity.test.ts`): ícones usados existem no conjunto, toda janela é alcançável em ≤ 2 passos, sem classe CSS duplicada.
6. **Mostre ao usuário** (abra o HTML ou envie o arquivo) e itere. Registre decisões novas em `sdd/spec/decisions.md`.

## Armadilhas já ocorridas
- **Ordem de inicialização do JS** quebrou na geração: declare `ICONS` e o estado antes de qualquer render, e teste com o console aberto (sem erros).
- **Colisão de classe CSS** com uma tela antiga: prefixe as classes do blueprint (ex.: `bp-`) ou isole num escopo.
- **Ícones não apareciam** porque o mapa `ICONS` não tinha todas as chaves (ex.: grid, link). A checagem de fidelidade deve cobrir isso.
- Excesso de botões no chat e transparência que prejudica a leitura.

## Entrega
Responda curto: caminho do HTML, caminho dos diagramas, evidências e o que o usuário deve conferir. Não cole o conteúdo do HTML na conversa.
