# Discovery — 98-diagram-studio-c4

## Problema

A skill `diagramas-c4-mermaid` usa apenas Mermaid como motor de renderização. O Mermaid renderiza nativamente no GitHub, mas tem limitações de layout (setas cruzadas, posicionamento). O `diagram-studio` (`github.com/nathanramorim/diagram-studio`) produz SVG/PNG dark-mode com assinatura e controle total de layout, mas requer escolha explícita pelo agente.

## Para quem

Agentes de IA (Claude Code, Gemini, Copilot) que usam a skill para gerar diagramas C4 no contexto do forge-sdd. O usuário final é Nathan.

## Jornada

1. Agente recebe pedido de diagrama.
2. Agente carrega a skill `diagramas-c4-mermaid`.
3. **Hoje:** segue direto para Mermaid.
4. **Com esta feature:** pergunta qual motor usar antes de começar.
5. Dependendo da escolha, segue o fluxo Mermaid (atual) ou o fluxo diagram-studio.

## Regras de negócio

- A escolha é feita a cada uso (não persiste entre sessões).
- O fluxo Mermaid existente não é alterado.
- O fluxo diagram-studio requer que o plugin esteja instalado; se ausente, orienta a instalação.

## Escopo

**Dentro:** atualizar `~/.claude/skills/diagramas-c4-mermaid/SKILL.md`.
**Fora:** modificar `renderizar-mermaid.py`; criar código Go no forge-sdd; alterar `.agents/skills/`.

## Riscos

- Plugin `diagram-studio` pode não estar instalado: mitigado com verificação e orientação de instalação no fluxo.
