# Tarefa feat-67-01 — Linguagem Clara, Remoção de Siglas e Toggle de Detalhes Técnicos

## 1. Contexto e Objetivo
Substituir todas as siglas e jargões constitucionais (`ARQ`, `SEG`, `DUP`, `COD`, `TST`, `FLX`, `GOV`) expostos na tela principal, gavetas de estações (`.station-drawer`), cartões e modais do Forge Studio por rótulos amigáveis e autoexplicativos em linguagem natural (pt-BR).

Para usuários avançados ou desenvolvedores que precisem inspecionar regras específicas, introduzir um botão alternador discreto ("Ver Detalhes Técnicos" / "Ocultar Detalhes Técnicos") que exibe ou oculta os códigos de regras e metadados brutos.

## 2. O que já existe (Reuse Check — DUP-02)
- Cartão de Governança e Grid de Pilares em `internal/studio/assets/index.html`.
- Mapeamento `CONSTITUTIONAL_PILLARS` em `internal/studio/assets/studio.js`.
- Estilos de badges e pilares em `internal/studio/assets/studio.css`.

## 3. Critérios de Aceitação (Gherkin)
### Cenário 1: Exibição Acessível por Padrão
- **Dado** que o usuário acessa o Forge Studio ou abre o Station Drawer
- **Quando** o painel de conformidade ou inspeção for renderizado
- **Então** os títulos dos pilares devem exibir termos claros: "Estrutura do Sistema", "Segurança da Informação", "Prevenção de Duplicidade", "Qualidade do Código", "Testes Automatizados", "Processo e Portões", "Governança Independente"
- **E** nenhuma sigla críptica deve ser visível na visualização padrão.

### Cenário 2: Alternância de Detalhes Técnicos
- **Dado** que a seção de conformidade/governança possui o botão "Ver Detalhes Técnicos"
- **Quando** o usuário clica nesse botão
- **Então** os IDs canônicos (ex: `ARQ-01..11`, `SEG-01..07`) e chaves técnicas devem ser revelados
- **E** o texto do botão deve mudar para "Ocultar Detalhes Técnicos".

## 4. Arquivos Afetados
- `internal/studio/assets/index.html`
- `internal/studio/assets/studio.css`
- `internal/studio/assets/studio.js`
- `internal/studio/assets_test.go`
