# feat-98-01 — skill diagramas-c4-mermaid: escolha entre Mermaid e diagram-studio

**Iniciativa:** 98-diagram-studio-c4
**Status:** draft
**Tamanho:** S

## Objetivo

Atualizar a skill `diagramas-c4-mermaid` (`~/.claude/skills/diagramas-c4-mermaid/SKILL.md`) para oferecer ao agente a **escolha entre dois motores de renderização** ao produzir diagramas:

- **Mermaid** (comportamento atual): blocos ` ```mermaid ` que renderizam nativamente no GitHub, VS Code e Obsidian.
- **diagram-studio** (`github.com/nathanramorim/diagram-studio`): JSON → SVG/PNG dark-mode com assinatura, via `uv run scripts/diagrama.py`. Resultado visual superior; exige embed de imagem no markdown.

A escolha é feita via `AskUserQuestion` no início de cada uso da skill, antes de qualquer esboço.

## Modelo do módulo

**Enxuto** (ARQ-11): sem código Go novo. Modifica apenas arquivos de skill fora do repositório (`~/.claude/skills/diagramas-c4-mermaid/`) e o `renderizar-mermaid.py` existente permanece inalterado. Nenhum pacote Go é criado.

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Motor atual | `~/.claude/skills/diagramas-c4-mermaid/SKILL.md` | `renderizar-mermaid.py` via `mermaid-cli` + Chrome | Reusar | Mantém o fluxo Mermaid que renderiza nativamente no GitHub, VS Code e Obsidian. |
| diagram-studio | `github.com/nathanramorim/diagram-studio` | plugin Claude Code com `scripts/diagrama.py` e `skills/diagrama-studio/SKILL.md` | Estender | A skill existente ganha o fluxo alternativo sem duplicar a renderização Mermaid. |
| Skill existente | `~/.claude/skills/diagramas-c4-mermaid/scripts/renderizar-mermaid.py` | Script Python funcional | Reusar | O script atende à opção Mermaid e fica inalterado conforme o escopo. |

## Decisões do projeto aplicadas

- **D1 / ADR-0003**: Skills por estação — a alteração é na skill de usuário, fora do repo. Sem impacto em `.agents/skills/`.
- **ARQ-11**: Spec declara modelo enxuto (sem pacote novo).
- **DUP-01**: O bloco "diagram-studio" não duplica a lógica Mermaid; são fluxos independentes dentro da mesma skill.

## Conteúdo da alteração

### 1. `.agents/skills/diagramas-c4-mermaid/SKILL.md` — reescrever (a skill vive no repositório, nunca em `~/.claude/`)

#### Novo `description` (frontmatter)

```
description: Cria diagramas de arquitetura C4 (contexto, container, componente, dinâmico, implantação). Pergunta ao agente qual motor usar — Mermaid (renderiza no GitHub) ou diagram-studio (SVG/PNG dark-mode com assinatura). Ative sempre que o pedido for diagrama, C4, desenho de arquitetura, fluxo ou proposta visual.
```

#### Novo passo 0 — escolha do motor

Antes de qualquer esboço, o agente usa `AskUserQuestion` com:

```
"Qual motor de renderização usar?"
  A) Mermaid  — blocos ```mermaid```, renderiza no GitHub/VS Code nativamente
  B) diagram-studio — SVG/PNG dark-mode com assinatura (nathanramorim/diagram-studio)
```

- Se **A (Mermaid)**: segue o fluxo atual (seções 1–6 do SKILL.md existente, sem alteração).
- Se **B (diagram-studio)**: segue o novo fluxo descrito na seção 2 abaixo.

#### Novo fluxo para diagram-studio (seção paralela no SKILL.md)

1. **Verificar instalação:** `ls ~/.claude/plugins/diagram-studio/scripts/diagrama.py 2>/dev/null || echo "AUSENTE"`. Se ausente, instruir:  
   ```
   /plugin marketplace add nathanramorim/diagram-studio
   /plugin install diagram-studio@diagram-studio
   ```
2. **Escolher modelo** (igual ao `SKILL.md` do diagram-studio): Fases, Personas, Fluxo ou Camadas.
3. **Montar JSON** em `docs/diagramas/<nome>.json` usando a referência do JSON do diagram-studio.
4. **Mapeamento C4 → diagram-studio**:

   | Nível C4 | Elemento JSON |
   |---|---|
   | Pessoa | `caixa` com `estilo: "contorno"` ou `"pill"` |
   | Sistema / Contêiner | `caixa` com `estilo: "normal"` ou `"destaque"` |
   | Sistema externo | `caixa` com `estilo: "apagado"` ou `"tracejado"` |
   | Fronteira / Boundary | `caixa` com `estilo: "grupo"` |
   | Relacionamento | `seta` com `forma: "horizontal"` ou `"vertical"` |

5. **Renderizar:** `uv run ~/.claude/plugins/diagram-studio/scripts/diagrama.py docs/diagramas/<nome>.json`
6. **Conferir PNG** (abrir imagem): texto apertado, setas cruzando caixas, blocos desalinhados.
7. **Entregar:** SVG para README/docs (`![](diagrama.svg)`), PNG para slides e chat.

### 2. Instalação opcional e restrita ao projeto aberto (decisão do dono)

Toda instalação de skill ou plugin feita pelo forge obedece a três regras:

1. **Opcional por skill:** a skill `diagramas-c4-mermaid` e o plugin `diagram-studio` só são instalados se o usuário escolher; o fluxo Mermaid funciona sem o plugin.
2. **Limitada ao projeto aberto:** a instalação grava só dentro do projeto em que o forge foi aberto (`.agents/skills/`, `.claude/`), nunca em `~/.claude/` nem em outro projeto.
3. **Hub ou subprojeto:** se existirem subprojetos com forge sob o projeto aberto, o forge pergunta onde instalar: no master (hub) ou em um dos subprojetos listados. Sem subprojetos, instala no projeto aberto sem perguntar.

O caminho do script do plugin no fluxo passa a ser resolvido a partir do destino escolhido, não de `~/.claude/plugins/`.

### 3. Sem alteração

- `scripts/renderizar-mermaid.py` do fluxo Mermaid — inalterado.

## Dependências novas

Nenhuma dependência Go. O fluxo diagram-studio requer o plugin `nathanramorim/diagram-studio` (verificado em runtime; orientação de instalação incluída no fluxo da skill).

## Rules e IDs da Constituição aplicáveis

`ARQ-11`, `DUP-01`, `COD-01`

## Critérios de Aceitação Executáveis

```bash
# 1. A skill vive no repositório, com frontmatter válido
head -5 .agents/skills/diagramas-c4-mermaid/SKILL.md | grep "name: diagramas-c4-mermaid"

# 2. Fluxo Mermaid preservado e integração diagram-studio referenciada
grep -c "renderizar-mermaid.py" .agents/skills/diagramas-c4-mermaid/SKILL.md   # >= 1
grep -c "diagram-studio" .agents/skills/diagramas-c4-mermaid/SKILL.md          # >= 3

# 3. Nada é gravado fora do projeto aberto, e hub/subprojeto é perguntado
go test ./internal/scaffold -run 'TestInstalaSkillSoNoProjetoAberto|TestInstalaPerguntaHubOuSubprojeto|TestInstalaOpcionalPorSkill'

# 4. Build e suíte
go build ./... && go test ./...
```
