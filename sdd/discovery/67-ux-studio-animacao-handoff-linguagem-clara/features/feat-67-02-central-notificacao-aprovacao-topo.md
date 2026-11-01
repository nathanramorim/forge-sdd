# Tarefa feat-67-02 — Central de Notificação Superior de Aprovação Humana

## 1. Contexto e Objetivo
Implementar um componente de notificação em destaque no cabeçalho do Forge Studio (`#headerNotificationBell`). 

Quando qualquer estação tiver um artefato ou especificação aguardando intervenção/revisão humana (como o status `pending_review` ou `review_pending` da Concepção / PM & Arquiteto no Portão FLX-01):
1. O ícone de notificação no topo exibe um pulso luminoso discreto com um badge de contagem de pendências (ex: `1 pendente`).
2. Ao clicar no botão de notificação, abre imediatamente a central/modal de revisão com o resumo explicativo da spec em linguagem clara.
3. Permite ao usuário aprovar o avanço com 1 clique ("Aprovar Spec (FLX-01)") ou inspecionar o arquivo completo.

## 2. O que já existe (Reuse Check — DUP-02)
- Visualizador de Spec (`#specViewerModal`, `openSpecViewer`) em `internal/studio/assets/studio.js`.
- Endpoint de aprovação `/api/spec/approve` em `internal/studio/server.go`.
- Eventos de SSE `/api/events` com o snapshot contendo as filas e itens em `pending_review`.

## 3. Critérios de Aceitação (Gherkin)
### Cenário 1: Notificação Silenciosa quando não há pendências
- **Dado** que não há nenhuma tarefa com status `pending_review` ou `review_pending`
- **Quando** o Forge Studio for carregado
- **Então** o ícone de notificação deve permanecer neutro sem badge de alerta.

### Cenário 2: Alerta Ativo de Aprovação Pendente
- **Dado** que a estação PM & Arquiteto concebeu uma spec e está aguardando aprovação humana
- **Quando** o snapshot for recebido pelo navegador
- **Então** o sino superior deve exibir a quantidade de pendências e classe `.has-pending` com pulso suave
- **E** ao clicar no sino, o modal de revisão deve abrir imediatamente carregando a especificação pendente.

## 4. Arquivos Afetados
- `internal/studio/assets/index.html`
- `internal/studio/assets/studio.css`
- `internal/studio/assets/studio.js`
- `internal/studio/assets_test.go`
