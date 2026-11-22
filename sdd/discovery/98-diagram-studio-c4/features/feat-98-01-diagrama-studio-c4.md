# feat-98-01 — skill diagramas-c4-mermaid: escolha entre Mermaid e diagram-studio

**Iniciativa:** 98-diagram-studio-c4
**Status:** draft
**Tamanho:** S

## Objetivo

Atualizar a skill `diagramas-c4-mermaid` (`.agents/skills/diagramas-c4-mermaid/SKILL.md`, movida do `~/.claude/skills/` para o repositório) para oferecer ao agente a **escolha entre dois motores de renderização** ao produzir diagramas:

- **Mermaid** (comportamento atual): blocos ` ```mermaid ` que renderizam nativamente no GitHub, VS Code e Obsidian.
- **diagram-studio** (`github.com/nathanramorim/diagram-studio`): JSON → SVG/PNG dark-mode com assinatura, via `uv run scripts/diagrama.py`. Resultado visual superior; exige embed de imagem no markdown.

A escolha é feita via `AskUserQuestion` no início de cada uso da skill, antes de qualquer esboço.

## Modelo do módulo

**Enxuto** (ARQ-11): esta feature só move e reescreve arquivos de skill dentro do repositório (`.agents/skills/diagramas-c4-mermaid/`). Não cria código Go nem pacote novo; a justificativa vale para esta feature, e não para a política de instalação, que vive na feat-98-02 (modelo completo, Go).

## O que já existe

| Procurei | Onde | Achei | Decisão | Justificativa |
|---|---|---|---|---|
| Skill de diagramas no repo | `.agents/skills/` | só `novo-subcomando-cli-sdd` e `sdd-*`; `diagramas-c4-mermaid` ausente | Criar no repo | A skill precisa viver no repositório (decisão do dono) |
| Motor atual | `~/.claude/skills/diagramas-c4-mermaid/` (somente leitura) | `SKILL.md` e `scripts/renderizar-mermaid.py` | Reusar o conteúdo | Mantém o fluxo Mermaid; script copiado sem alteração |
| diagram-studio | `~/.claude/plugins/marketplaces/diagram-studio/` | plugin 0.2.0 com `scripts/diagrama.py` | Referenciar como opção B | Output visual superior; instalação é tratada na feat-98-02 |

## Decisões do projeto aplicadas

- **ADR-0003**: skills por estação vivem em `.agents/skills/`, no repositório.
- **ARQ-11**: modelo enxuto declarado e justificado acima.
- **DUP-01**: o fluxo diagram-studio não duplica a lógica Mermaid.

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

1. **Verificar instalação:** `ls .agents/plugins/diagram-studio/scripts/diagrama.py 2>/dev/null || echo "AUSENTE"` (caminho único, relativo ao destino de instalação: o projeto aberto, ou o hub/subprojeto escolhido). Se ausente, oferecer a instalação restrita ao projeto da feat-98-02 (`forge skills instalar plugin diagram-studio`, que grava só em `<destino>/.agents/plugins/diagram-studio/`). **Não usar `/plugin marketplace add` nem `/plugin install`**, pois gravam em `~/.claude/`.
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

5. **Renderizar:** `uv run .agents/plugins/diagram-studio/scripts/diagrama.py docs/diagramas/<nome>.json --saida docs/diagramas/<nome>` (rodar a partir do destino resolvido no passo 1; mesmo caminho do passo 1)
6. **Conferir PNG** (abrir imagem): texto apertado, setas cruzando caixas, blocos desalinhados.
7. **Entregar:** SVG para README/docs (`![](diagrama.svg)`), PNG para slides e chat.

### 2. Instalação

A skill só descreve o fluxo e aciona o comando de instalação restrita ao projeto. A política (opcional por skill, limitada ao projeto aberto, hub ou subprojeto) é implementada e testada na **feat-98-02**, que depende desta.

### 3. Sem alteração (copiado do original, sem editar)

- `scripts/renderizar-mermaid.py` do fluxo Mermaid — inalterado.

## Dependências novas

Nenhuma dependência Go. Plugin `diagram-studio` e `uv` são tratados na feat-98-02 (SEG-07).

## Rules e IDs da Constituição aplicáveis

`ARQ-11`, `DUP-01`, `COD-01`

## Critérios de Aceitação Executáveis

```bash
# 1. A skill vive no repositório, com frontmatter válido
head -5 .agents/skills/diagramas-c4-mermaid/SKILL.md | grep "name: diagramas-c4-mermaid"

# 2. Fluxo Mermaid preservado e diagram-studio referenciado
test -f .agents/skills/diagramas-c4-mermaid/scripts/renderizar-mermaid.py
grep -c "renderizar-mermaid.py" .agents/skills/diagramas-c4-mermaid/SKILL.md   # >= 1
grep -c "diagram-studio" .agents/skills/diagramas-c4-mermaid/SKILL.md          # >= 3

# 3. Nenhuma referência a ~/.claude/ dentro da skill nem da spec executável
! grep -rn "~/.claude" .agents/skills/diagramas-c4-mermaid/

# 3b. Nenhum /plugin install na skill
! grep -rn "/plugin install\|/plugin marketplace" .agents/skills/diagramas-c4-mermaid/

# 4. Build e suíte
go build ./... && go test ./...
```
