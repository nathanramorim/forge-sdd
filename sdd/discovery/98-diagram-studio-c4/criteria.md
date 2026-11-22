# Critérios de Arquitetura — 98-diagram-studio-c4

## Modelo do módulo

**Enxuto** (ARQ-11): sem pacote Go novo. Apenas arquivo de skill do repositório (`.agents/skills/diagramas-c4-mermaid/SKILL.md`).

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Motor de render atual | `.agents/skills/diagramas-c4-mermaid/SKILL.md` | `renderizar-mermaid.py` via mermaid-cli | Reusar como opção A | Renderiza no GitHub nativamente; sem alteração |
| diagram-studio | `github.com/nathanramorim/diagram-studio` | Plugin Claude Code com `scripts/diagrama.py` e skill própria | Estender SKILL.md com opção B | Output visual superior: SVG/PNG dark-mode com assinatura |
| Script de renderização | `.agents/skills/diagramas-c4-mermaid/scripts/renderizar-mermaid.py` | Python funcional, sem bugs conhecidos | Reusar sem alteração | Fora do escopo desta feature |

## Decisões do projeto aplicadas

- **ARQ-11**: spec declara modelo enxuto.
- **DUP-01**: os dois fluxos são independentes; nenhuma lógica duplicada.
- **COD-01**: responsabilidade única por fluxo dentro do SKILL.md.
- **ADR-0003**: skills por estação — a skill vive em `.agents/skills/diagramas-c4-mermaid/SKILL.md`, no repositório.

## Dependências novas

Nenhuma dependência Go nova. O fluxo diagram-studio depende do plugin `nathanramorim/diagram-studio` externo, verificado em runtime; instalar é opcional por skill, limitado ao projeto aberto, e pergunta hub (master) ou subprojeto quando houver forge em subprojetos.
