# Critérios de Arquitetura — 98-diagram-studio-c4

## Modelo do módulo

**Enxuto** (ARQ-11): sem pacote Go novo. Apenas arquivo de skill do usuário (`~/.claude/skills/diagramas-c4-mermaid/SKILL.md`).

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Motor de render atual | `~/.claude/skills/diagramas-c4-mermaid/SKILL.md` | `renderizar-mermaid.py` via mermaid-cli | Reusar como opção A | Renderiza no GitHub nativamente; sem alteração |
| diagram-studio | `github.com/nathanramorim/diagram-studio` | Plugin Claude Code com `scripts/diagrama.py` e skill própria | Estender SKILL.md com opção B | Output visual superior: SVG/PNG dark-mode com assinatura |
| Script de renderização | `~/.claude/skills/diagramas-c4-mermaid/scripts/renderizar-mermaid.py` | Python funcional, sem bugs conhecidos | Reusar sem alteração | Fora do escopo desta feature |

## Decisões do projeto aplicadas

- **ARQ-11**: spec declara modelo enxuto.
- **DUP-01**: os dois fluxos são independentes; nenhuma lógica duplicada.
- **COD-01**: responsabilidade única por fluxo dentro do SKILL.md.
- **ADR-0003**: skills por estação — a alteração é na skill de usuário, fora do repo.

## Dependências novas

Nenhuma dependência Go nova. O fluxo diagram-studio depende do plugin `nathanramorim/diagram-studio` instalado via `/plugin install diagram-studio@diagram-studio` (externo, verificado em runtime).
