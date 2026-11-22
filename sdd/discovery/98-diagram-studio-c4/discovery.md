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
- O fluxo diagram-studio requer o plugin; se ausente, a skill oferece instalar. Instalar skill/plugin (sem `/plugin install`, que grava em `~/.claude/`) é opcional por skill, limitado ao projeto aberto, e pergunta hub (master) ou subprojeto quando houver forge em subprojetos.

## Escopo

**feat-98-01 (enxuto, sem Go):** mover/reescrever a skill para `.agents/skills/diagramas-c4-mermaid/` no repositório.
**feat-98-02 (completo, Go em `internal/scaffold`):** política de instalação opcional, limitada ao projeto aberto, com pergunta hub ou subprojeto; depende da 98-01 e exige nova ADR (FLX-05/GOV-01), pois vale para todo o forge.
**Fora:** modificar `renderizar-mermaid.py`.

## Riscos

- Plugin `diagram-studio` pode não estar instalado: mitigado com verificação e orientação de instalação no fluxo.
