# Critérios de Aceitação — Discovery 67: Otimização de UX, Animação de Handoff e Notificações

## Critérios de Aceitação (Gherkin)

### Cenário 1: Notificação Superior de Aprovação Pendente
- **Dado** que uma especificação técnica ou artefato requer aprovação humana
- **Quando** o Forge Studio carregar ou receber o evento via SSE
- **Então** um ícone de notificação (sino) na barra superior deve exibir um indicador visual de pendência
- **E** ao clicar no ícone, deve abrir diretamente a janela de revisão com resumo em linguagem clara e botões de ação ("Aprovar e Avançar" ou "Ajustar").

### Cenário 2: Animação Visual de Handoff entre Mascotes 3D
- **Dado** que uma estação de trabalho conclui sua etapa e realiza um handoff para a próxima estação
- **Quando** o evento de handoff for processado na interface 3D
- **Então** os dois mascotes envolvidos devem executar uma animação visual de interação (arco de conexão de partículas ou orientação mútua)
- **E** um indicador visual sutil de transferência de tarefa deve fluir da estação de origem para a de destino.

### Cenário 3: Linguagem Acessível por Padrão sem Siglas Crípticas
- **Dado** que o usuário está visualizando a tela principal, gavetas de estações ou relatórios de governança
- **Quando** inspecionar as informações de conformidade e status
- **Então** todos os pilares devem ser descritos em linguagem natural e direta (ex: "Estrutura do Sistema", "Segurança da Informação", "Qualidade do Código")
- **E** as siglas técnicas e códigos de regras só devem ser exibidos se o usuário clicar no botão "Ver Detalhes Técnicos".

### Cenário 4: Inicialização Imediata com Sessões Conectadas
- **Dado** que o usuário inicia o Forge Studio via comando `forge-sdd studio`
- **Quando** a aplicação carrega a interface no navegador
- **Então** todas as estações já devem reportar suas sessões instanciadas e prontas no painel
- **E** nenhum agente deve figurar como órfão ou desconectado.

### Cenário 5: Unificação de Concepção e Planejamento no Fluxo Visual
- **Dado** que o usuário acompanha o fluxo visual de trabalho
- **Quando** uma demanda é concebida pelo PM & Arquiteto
- **Então** a quebra e fatiamento em tarefas determinísticas deve ser apresentada como uma subetapa integrada da Concepção, sem exigir transição desconexa de estação.

## Regras Constitucionais Aplicáveis
- `ARQ-03`, `ARQ-04`: 100% Go stdlib pura, templates embutidos via `embed.FS`, zero dependências npm em produção.
- `SEG-01`: Sanitização estrita contra injeção de HTML/XSS em todos os painéis e notificações.
- `COD-01`, `COD-05`: Nível de linguagem padrão em pt-BR e responsabilidade única de componentes.
- `TST-04`: Testes automatizados determinísticos passando com `go test -count=1 -race ./...`.
