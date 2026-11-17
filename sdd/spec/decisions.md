# Decisões — forge-sdd

## Resolvidas

| # | Decisão | Resolução | Motivo |
|---|---------|-----------|--------|
| D1 | Linguagem do CLI | Go | Binário estático, sem runtime externo |
| D2 | Framework CLI | cobra v1.9 (`go.mod`: v1.9.1) | Padrão Go, flags + help automático |
| D3 | Prompts interativos | huh v0.3 (charmbracelet) | Sem CGO, TUI moderno, sem deps pesadas |
| D4 | Render de templates | text/template (stdlib) | Zero dependência, suficiente para o caso de uso |
| D5 | Embed de templates | embed.FS (stdlib) | Binário único, sem assets externos em runtime |
| D6 | Não sobrescrever arquivos | Erro com lista de conflitos | Evitar perda de dados silenciosa |
| D7 | Subcomandos de runtime | Substituída por D14: o pipeline tem subcomandos de runtime (`worker`, `bus`, `dispatch`) | Era "nenhum"; a ADR-0004 os introduziu |
| D8 | Idioma padrão | pt-BR | Público-alvo primário; `--lang en` para inglês |
| D9 | Distribuição | goreleaser v2 + GitHub Releases | Multi-OS automatizado, zero manual |
| D12 | Versão no binário | Injetada por `ldflags` (`-X main.version`) e lida só por `config.BinaryVersion` | Fonte única da versão em telemetria, ACP e defaults (`DUP-04`, ADR-0014) |
| D13 | Configurações de MCP | Separadas por agente (`.gemini/`, `.vscode/`, etc.) | Cada agente possui formas distintas de gerenciar MCPs e habilidades. |
| D14 | Pipeline de trabalho | 3 sessões (Discovery, Builder, Reviewer) ligadas por barramento de jobs | Contexto separado, revisão independente e paralelismo (ADR-0004) |
| D15 | Nome do comando | `forge` oficial; `forge-sdd` mantido como alias no npm | Um nome só para pessoas e agentes (ADR-0010) |
| D16 | Reset de contexto | Por job, no mesmo painel, com cartão de handoff | Previsível e sem fechar o tmux (ADR-0009) |
| D17 | Estado da fila após aprovação | O Go propaga `approved` às tasks do job | O agente não precisa lembrar; o painel fica correto (ADR-0012) |
| D18 | Modo avançado | `autonomy: advanced` em `forge.json`: o Reviewer revisa a spec no lugar do Gate 1 humano; tetos, `STOP` e `blocked` | Esteira que roda sem acompanhamento, em ambiente isolado; experimental (ADR-0013) |
| D19 | `/toca-ai` | Motor em Go (`forge pipeline toca-ai`): pré-voo com consentimentos só da pessoa, um job por etapa; archive automático ao fim só com a branch da iniciativa em uso (a recusa é silenciosa, limitação conhecida) | O Líder planeja e delega; o contexto e a ordem ficam com o Go (iniciativa 89) |
| D20 | Passos mecânicos no Go | `forge pipeline pr` abre o PR na base da iniciativa; o worker lê o uso real do `result`; `forge status` curto sem terminal; `forge archive` move a pasta | Menos turnos do agente, número real em vez de estimativa (ADR-0014) |
| D21 | Modelo e esforço por sessão | `models` e `efforts` por papel em `forge.json`, aplicados pelos marcadores `{model}` e `{effort}` do perfil | Sem escolha, o argv não muda; só flags confirmadas por CLI (ADR-0014 D-5) |

## Abertas

| # | Questão |
|---|---------|
| D10 | [ ] Suportar `--lang en` com templates em inglês? Duplicar templates ou sistema de i18n? |
| D11 | [ ] `forge init` em diretório não-vazio: bloquear tudo ou listar conflitos e continuar? |

## Índice de ADRs (`sdd/spec/adr/`)

| ADR | Título | Status |
|-----|--------|--------|
| 0001 | Remoção do Forge Studio | `aprovado` |
| 0002 | Layout e Estado por Iniciativa | `aprovado` |
| 0003 | Skills por Estação e Orçamento de Contexto | `aprovado` |
| 0004 | Pipeline Multi-Sessão (Discovery, Builder e Reviewer) | `aprovado` |
| 0005 | Menu Navegável e Agente por Sessão do Pipeline | `aprovado` |
| 0006 | Visibilidade, tmux e Entrega ao Builder | `aprovado` |
| 0007 | Contexto do Projeto na Discovery, Briefing e Contingência de Terminal | `aprovado` |
| 0008 | Sessões com CLI Aberta no Tmux (Modo Live e Despachante de Entrega) | `aprovado` |
| 0009 | Tokens por Janela, Reset Suave de Contexto e Auto-Archive | `proposto` |
| 0010 | `forge` como Único Nome de Comando | `proposto` |
| 0011 | Painel Interativo, Stream de Chat nos Workers e Fila com Drill-Down | `proposto` |
| 0012 | A Aprovação do Job Atualiza a Fila da Iniciativa | `proposto` |
| 0013 | Reviewer Revisa a Spec no Modo Avançado | `proposto` |
| 0014 | Esteira Econômica e Autônoma | `proposto` |
| 0015 | Alinhamento da Constituição e da Especificação Arquitetural com o Código | `proposto` |
