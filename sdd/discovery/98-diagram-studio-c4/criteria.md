# Critérios de Arquitetura — 98-diagram-studio-c4

## Modelo do módulo

- **feat-98-01 — enxuto** (ARQ-11): só move/reescreve a skill em `.agents/skills/diagramas-c4-mermaid/`; sem Go.
- **feat-98-02 — completo** (ARQ-11): Go em `internal/scaffold` (`instalar_skill.go`: `InstalarSkill`, `InstalarPlugin`, `ListarDestinosInstalacao`), chamado pelos subcomandos `skills destinos` e `skills instalar` do grupo existente `skills`; reusando `WriteFileSafe` e `shouldPreserve`; sem pacote novo.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Motor de render atual | `.agents/skills/diagramas-c4-mermaid/SKILL.md` | `renderizar-mermaid.py` via mermaid-cli | Reusar como opção A | Renderiza no GitHub nativamente; sem alteração |
| diagram-studio | `github.com/nathanramorim/diagram-studio` | Plugin Claude Code com `scripts/diagrama.py` e skill própria | Estender SKILL.md com opção B | Output visual superior: SVG/PNG dark-mode com assinatura |
| Script de renderização | `.agents/skills/diagramas-c4-mermaid/scripts/renderizar-mermaid.py` | Python funcional, sem bugs conhecidos | Reusar sem alteração | Fora do escopo desta feature |

## Decisões do projeto aplicadas

- **ARQ-11**: 98-01 enxuto, 98-02 completo, cada um justificado na própria feature.
- **SEG-01/SEG-02/SEG-07, TST-01**: validação de destino, gravação atômica com preservação, versões exatas e testes (98-02).
- **FLX-05/GOV-01**: a política vale para todo o forge; follow-up declarado: nova ADR em `sdd/spec/adr/`.
- **DUP-01**: os dois fluxos são independentes; nenhuma lógica duplicada.
- **COD-01**: responsabilidade única por fluxo dentro do SKILL.md.
- **ADR-0003**: skills por estação — a skill vive em `.agents/skills/diagramas-c4-mermaid/SKILL.md`, no repositório.

## Dependências novas

Nenhuma dependência Go. Runtime do fluxo diagram-studio (SEG-07, consultado em 2026-11-22):

- plugin `diagram-studio` 0.2.0, commit `3996fd2e2bcdc2462e6f0d1edd7f6d18f649bc03` (`plugin.json` local; `git ls-remote` confirma o commit; sem tags).
- `uv` 0.9.7 (`uv --version`); docs.astral.sh/uv confirma que `uv run` respeita metadados inline do script. O context7 não estava disponível: sem consulta nele.
- Pendente: `resvg-py>=0.1` do `diagrama.py` é faixa aberta; o Builder fixa a versão exata. Detalhe na feat-98-02.

## Política de instalação

Skill **e plugin** ficam restritos ao projeto aberto (ou ao destino hub/subprojeto escolhido); `/plugin install` não é usado. ADR nova: `sdd/spec/adr/adr-0025-instalacao-de-skills-e-plugins-limitada-ao-projeto.md` (0024 existe em outra branch).

## Critérios executáveis resumidos

`go test ./internal/scaffold -run 'TestInstalaSkillSoNoProjetoAberto|TestInstalaPluginSoNoProjetoAberto|TestInstalaPerguntaHubOuSubprojeto|TestInstalaNaoRepeteQuandoHubTem|TestInstalaNoSubprojetoQuandoSemHub|TestInstalaOpcionalPorSkill'`; `ls sdd/spec/adr/adr-0025-instalacao-de-skills-e-plugins-limitada-ao-projeto.md`; `! grep -rn "/plugin install" .agents/skills/diagramas-c4-mermaid/`.
