# Decisões — forge-sdd

## Resolvidas

| # | Decisão | Resolução | Motivo |
|---|---------|-----------|--------|
| D1 | Linguagem do CLI | Go | Binário estático, sem runtime externo |
| D2 | Framework CLI | cobra v1.9.1 | Padrão Go, flags + help automático |
| D3 | Prompts interativos | huh v0.3 (charmbracelet) | Sem CGO, TUI moderno, sem deps pesadas |
| D4 | Render de templates | text/template (stdlib) | Zero dependência, suficiente para o caso de uso |
| D5 | Embed de templates | embed.FS (stdlib) | Binário único, sem assets externos em runtime |
| D6 | Não sobrescrever arquivos | Erro com lista de conflitos | Evitar perda de dados silenciosa |
| D7 | Subcomandos de runtime | Nenhum (chatmodes Copilot) | CLI mínimo, sem manutenção dupla |
| D8 | Idioma padrão | pt-BR | Público-alvo primário; `--lang en` para inglês |
| D9 | Distribuição | goreleaser v2 + GitHub Releases | Multi-OS automatizado, zero manual |
| D13 | Configurações de MCP | Separadas por agente (`.gemini/`, `.vscode/`, etc.) | Cada agente possui formas distintas de gerenciar MCPs e habilidades. |
| D14 | Pipeline de trabalho | 3 sessões (Discovery, Builder, Reviewer) ligadas por barramento de jobs | Contexto separado, revisão independente e paralelismo (ADR-0004) |
| D15 | Nome do comando | `forge` oficial; `forge-sdd` mantido como alias no npm | Um nome só para pessoas e agentes (ADR-0010) |
| D16 | Reset de contexto | Por job, no mesmo painel, com cartão de handoff | Previsível e sem fechar o tmux (ADR-0009) |
| D17 | Estado da fila após aprovação | O Go propaga `approved` às tasks do job | O agente não precisa lembrar; o painel fica correto (ADR-0012) |
| D18 | Modo de execução do pipeline | `sessions` (padrão) ou `chat` em subagentes; o Go é dono do laço | ADR-0015 |
| D19 | Exclusão mútua por iniciativa | Lock em arquivo com lease em `.forge/runtime/chat/locks` | ADR-0016 |
| D20 | Superfície MCP | Só `forge_doctor`, somente leitura e saneada | ADR-0017 |
| D21 | Escrita em arquivo do usuário | Validar alvo e gravar por temporário e rename | ADR-0018 |
| D22 | Edição temporária da LP | Porteiro por data em `docs/edicao/edicao.js` | ADR-0019 |
| D23 | Hospedagem da LP | Cloudflare Workers por CI | ADR-0020 |

## Abertas

| # | Questão |
|---|---------|
| D10 | [ ] Suportar `--lang en` com templates em inglês? Duplicar templates ou sistema de i18n? |
| D11 | [ ] `forge-sdd init` em diretório não-vazio: bloquear tudo ou listar conflitos e continuar? |
| D12 | Embutir versão da metodologia no binário via `ldflags` ou constante em `internal/config`? |

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
| 0015 | Pipeline em Subagentes no Mesmo Chat (modo `chat`) | `proposto` |
| 0016 | Lock por Iniciativa e Estado do Laço em `.forge/runtime/chat` | `aceito` (20/11/2026) |
| 0017 | Servidor MCP Somente Leitura com Saída Saneada | `aceito` (20/11/2026) |
| 0018 | Gravação Segura de Arquivos do Usuário | `aceito` (20/11/2026) |
| 0019 | Edição Temporária do Site com Porteiro por Data | `aceito` (20/11/2026) |
| 0020 | Deploy da LP por Cloudflare Workers | `aceito` (20/11/2026) |
| 0021 | Obsolescência das 5 Estações como Visão do Fluxo | `aceito` (20/11/2026) |
