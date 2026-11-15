# Discovery 67 — Otimização de Experiência do Usuário (UX), Linguagem Clara, Animação de Handoff entre Agentes e Inicialização Imediata de Sessões no Forge Studio

## 1. Contexto e Oportunidade
O Forge Studio evoluiu para se tornar a cabine de comando (cockpit) de observabilidade visual em tempo real do framework Forge-SDD. Com a introdução dos mascotes 3D em Three.js, suporte ao protocolo ACP (Agent Communication Protocol) e telemetria viva, o fluxo operacional funciona com alta precisão e integridade determinística.

No entanto, o feedback direto de uso apontou pontos fundamentais de atrito na experiência do usuário (UX):
1. **Linguagem Técnica Críptica e Siglas Ocultando o Significado:** A interface ainda expõe siglas constitucionais (`ARQ`, `SEG`, `DUP`, `COD`, `TST`, `FLX`, `GOV`) e termos de baixo nível na visão primária, dificultando o entendimento por pessoas de produto, líderes ou desenvolvedores que desejam ver o progresso sem decifrar códigos.
2. **Passagem de Tarefas Invisível / Estática:** A transição de uma estação para a outra (Handoff) é refletida principalmente por textos e tabelas. Não há um feedback visual rico e dinâmico que transmita claramente a sensação de "um agente conversando com o outro e passando a tarefa adiante".
3. **Ausência de Central Visual de Notificação de Aprovação:** Quando uma especificação técnica ou decisão requer validação humana (Portão de Revisão da Spec), o usuário precisa deduzir onde clicar. Deve existir um sino/ícone de notificação destacado no topo que chame a atenção de forma intuitiva e abra direto a gaveta de aprovação simplificada.
4. **Partida a Frio de Sessões:** Ao iniciar o ambiente, as estações surgem marcadas como `idle` ou desconectadas até que um comando seja disparado. A aplicação deve carregar com o pool de sessões já ativo e conectado para cada um dos agentes.
5. **Atrito entre Concepção e Planejamento:** A jornada atual separa a concepção da spec (PM & Arquiteto) do fatiamento de tarefas (Planejador) como duas etapas estáticas distintas no menu principal. O fatiamento em tarefas é um desdobramento natural do Discovery aprovado, devendo ser unificado na etapa de Concepção & Fatiamento, eliminando trocas de contexto desnecessárias.

## 2. O que já existe (Reuse Check — DUP-02)
| Recurso / Componente | Onde reside | Status | Decisão de Reúso |
|---|---|---|---|
| Mascotes 3D Three.js | `internal/studio/assets/studio.js` | Ativo | **Estender** para adicionar animações dinâmicas de handoff (diálogo visual, conexões de partículas) |
| SessionPool ACP | `internal/acp/pool.go` | Ativo | **Reusar** para garantir pre-warming no boot do Studio |
| Gaveta da Estação (`.station-drawer`) | `internal/studio/assets/studio.js` | Ativo | **Reusar** e refinar para exibir linguagem clara com botão "Ver Detalhes Técnicos" |
| Visualizador de Spec & Modal | `internal/studio/assets/index.html` | Ativo | **Estender** para abrir via sino de notificações no topo |
| Endpoints de Telemetria e Fila | `internal/studio/server.go` | Ativo | **Preservar 100%** retrocompatibilidade e integridade de dados |

## 3. Objetivos Centrais
- **UX Descomplicada e Sem Siglas:** Todos os painéis primários apresentam termos acessíveis em português padrão: *Estrutura do Sistema*, *Segurança da Informação*, *Boas Práticas de Código*, *Testes Automatizados*, *Reaproveitamento*, etc. Detalhes brutos e códigos constitucionais ficam sob opção expansível *"Ver Detalhes Técnicos"*.
- **Animações de Interação e Handoff entre Mascotes:** Quando o Líder repassa demanda para Concepção, ou Concepção para Builder, ou Builder para Governança, a cena 3D executa uma coreografia de passagem de tarefa (ex: arco de luz/partículas conectando os mascotes e balão sutil de mensagem).
- **Central de Notificação de Aprovação no Topo:** Um ícone de sino pulsante com badge de contagem de pendências (ex: `1 pendente`) alerta o usuário. Ao clicar, abre o modal de aprovação direta com resumo executivo leigo e botão de aprovação em 1 clique.
- **Sessões Pre-Ativadas no Boot:** O Studio inicia o pool com todas as sessões instanciadas, exibindo cada agente pronto e conectado.
- **Concepção & Planejamento Unificados:** A interface consolida a concepção da spec e sua decomposição em um fluxo contínuo.

## 4. Fronteiras e Não-Objetivos
- **NÃO** alterar os contratos JSON de API de backend existentes nem remover campos que garantem a segurança do framework.
- **NÃO** adicionar bibliotecas externas de frontend (mantendo Three.js já embutido e JavaScript puro na stdlib Go).
- **NÃO** quebrar o funcionamento do CLI nem ferramentas de linha de comando.
