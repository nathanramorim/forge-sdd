# Plano de Execução — Discovery 67: Otimização de UX, Animação de Handoff e Notificações

## Fatiamento das Subtarefas / Features

O plano é dividido em 4 entregas ordenadas e independentes, mantendo o funcionamento atual 100% íntegro a cada passo:

### Tarefa 1: `feat-67-01-linguagem-clara-e-toggle-tecnico`
- **Escopo:**
  - Atualizar os templates de cartões, pilares constitucionais e gaveta de estações (`internal/studio/assets/index.html` e `studio.js`).
  - Substituir rótulos crípticos por títulos em linguagem natural acessível.
  - Adicionar botão alternador "Ver Detalhes Técnicos" (oculto por padrão).
- **Critério Executável:** Testes em `internal/studio/assets_test.go` validando os novos textos didáticos e o toggle técnico.

### Tarefa 2: `feat-67-02-central-notificacao-aprovacao-topo`
- **Escopo:**
  - Adicionar o componente visual de sino de notificações (`#headerNotificationBell`) no cabeçalho do Studio.
  - Sincronizar via SSE e estado do snapshot: quando houver item com `status: pending_review`, acionar pulso e contador.
  - Clicar no sino abre diretamente o modal de aprovação humana com resumo leigo e botões de ação imediata.
- **Critério Executável:** Testes unitários de DOM e contrato de ativação da notificação ao receber eventos de pendência.

### Tarefa 3: `feat-67-03-animacao-handoff-threejs-mascotes`
- **Escopo:**
  - Estender o módulo Three.js em `internal/studio/assets/studio.js` para criar a rotina de animação de handoff.
  - Quando ocorrer a transição de estado/handoff entre duas estações, animar uma partícula/curva luminosa viajando do mascote de origem até o de destino, acompanhado de rotação sutil e balão de status temporário.
- **Critério Executável:** Testes de renderização Three.js sem exceções de console e sem vazamento de geometria ou materiais.

### Tarefa 4: `feat-67-04-boot-sessoes-prewarming-e-unificacao-concepcao`
- **Escopo:**
  - Garantir no boot do Studio (`server.go`) que o SessionPool execute o pre-warming imediato de todas as sessões.
  - Na apresentação visual da esteira, integrar as etapas de concepção e fatiamento sob a estação unificada de Concepção & Planejamento, simplificando o fluxo para o usuário.
  - Validação end-to-end com `go test -race ./...` e auditoria de Governança.
- **Critério Executável:** Suite completa de testes passando 100% limpa com concorrência.
