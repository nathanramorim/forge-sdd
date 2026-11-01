# Discovery — feat-66: Telemetria Detalhada, Inspeção Passo a Passo e Reconciliação de Atividades Órfãs

## 1. Visão Geral & Problema

### 1.1 Contexto Atual da Telemetria no Forge-SDD
A partir das Features 59, 63, 64 e 65, o **Forge-SDD** e o **Forge Studio** passaram a contar com um cockpit visual unificado de observabilidade, integrando as 5 estações canônicas de trabalho (`01. O Líder`, `02. PM & Arquiteto`, `03. O Planejador`, `04. Fila de Builders`, `05. Analista de Governança`), seus respectivos mascotes 3D e um console em tempo real intitulado **"Telemetria das Estações & Uso de Tools em Tempo Real"**.

Nesse console, cada micro-ação executada pelas estações (ou por subagentes e ferramentas ACP/MCP) gera um registro em `sdd/.queue/activity.json` contendo campos básicos: timestamp, estação, nome da tool, ação simplificada, detalhes e status (`running`, `success`, `error`).

### 1.2 Os Problemas Identificados

#### Problema A: Falta de Profundidade na Inspeção (Caixa-Preta de Execução)
Atualmente, as linhas da tabela de telemetria no Forge Studio são estáticas e puramente informativas em uma única linha condensada:
1. Ao olhar para uma linha como `go test -race ./... (PASS)` ou `edit_file server.go`, o operador não consegue clicar para inspecionar os detalhes reais daquela micro-ação:
   - Quais foram os **argumentos completos** e flags passadas ao comando?
   - Qual foi o **caminho exato do arquivo** inspecionado ou editado?
   - Qual foi o **output (stdout/stderr)** retornado pelo comando de teste, linter ou ferramenta MCP?
   - Qual foi a **duração em milissegundos** daquela etapa específica?
   - Qual era o **estado do mascote/estação** no exato momento da execução?
2. Em caso de falha (`status: "error"`), o desenvolvedor vê apenas o resumo e precisa recorrer ao terminal bruto ou abrir arquivos de log externos para diagnosticar a causa raiz da reprovação do Portão A ou da Governança Camada B.

#### Problema B: Status Órfãos (`running` sem justificativa) e Dessincronização de Estado
No estado atual da fila de telemetria (`sdd/.queue/activity.json`), é comum encontrar múltiplos registros antigos marcados perpetuamente com `status: "running"`:
1. Quando uma ferramenta inicia, ela emite um registro com `status: "running"`. No entanto, se o comando é concluído de forma implícita, se a estação avança para a próxima etapa sem emitir explicitamente um evento com a mesma chave ou se ocorre um encerramento de sessão, o registro anterior permanece congelado em `running` para sempre.
2. Na interface do Forge Studio, o usuário visualiza que todos os 5 agentes estão em estado **🟢 Standby** (ociosos, sem tarefas em processamento), mas na tabela de telemetria figuram linhas amarelas piscando como `running` ocorridas minutos ou horas antes.
3. Isso viola a integridade da observabilidade: **se os agentes estão ociosos (`idle`/`standby`), é logicamente impossível que haja atividades em `running` em execução.**

### 1.3 A Solução Proposta

A **Feature 66** implementa uma solução integral de observabilidade cirúrgica e reconciliação determinística para o Forge-SDD:

1. **Modal/Drawer de Inspeção de Telemetria Passo a Passo (Forge Studio):**
   - Torna cada linha da tabela de telemetria clicável com feedback visual de hover (`cursor: pointer`).
   - Abre um modal/drawer rico contendo:
     - **Estação e Mascote:** Identificação visual da estação, avatar/mascote e seu estado.
     - **Metadados Temporais:** Timestamp exato de início, timestamp de conclusão e duração em milissegundos (`duration_ms`).
     - **Tool & Comando:** Nome da ferramenta invocada, comando completo executado, argumentos detalhados e caminhos de arquivo envolvidos (com botão de cópia com 1 clique).
     - **Output Completo & Diagnóstico:** Visualizador monoespaçado estilizado tipo terminal para `stdout`, laudos JSON, parecer da Governança e bloco vermelho destacado para `stderr`/erros.
     - **Navegação Contínua:** Botões "Anterior" e "Próximo" e atalhos de teclado (`←`, `→`, `ESC`) para navegar sequencialmente pelas micro-ações sem sair do modal.

2. **Reconciliação Automática e Eliminação de Status Órfãos em Go stdlib:**
   - **Camada de Transição de Estado:** Sempre que uma estação transiciona para `standby` ou `idle`, qualquer atividade pendente daquela estação em `running` é automaticamente reconciliada para `success` ou `done`.
   - **Camada de Registro Consecutivo:** Ao registrar uma nova atividade para uma estação em `RecordActivity`, atividades anteriores em `running` da mesma estação são automaticamente finalizadas com cálculo da duração decorrida.
   - **Camada de Timeout Determinístico:** Qualquer atividade que permaneça em `running` por mais de 60 segundos sem batimento cardíaco é reconciliada por timeout (`timeout` / `done`).
   - **Sanitização no Boot:** Na inicialização do Forge Studio e comandos CLI de fila, `sdd/.queue/activity.json` é percorrido e qualquer registro órfão histórico é sanitizado de forma atômica (`WriteJSONAtomic`).

---

## 2. Personas

### Persona 1: Desenvolvedor / Operador do Forge Studio
- **Perfil:** Engenheiro de software full-cycle utilizando o Forge Studio como centro de comando do desenvolvimento guiado por especificações.
- **Dores:**
  - Incerteza ao ver um indicador `running` antigo na tabela e não saber se um teste de background está travado ou se é apenas lixo de telemetria.
  - Frustração ao ter que alternar para o terminal para ver por que um comando falhou, em vez de ver o output diretamente na interface do Studio.
- **Objetivo com a Feat 66:** Clicar na linha da telemetria e inspecionar instantaneamente o log de saída, flags do comando e tempo de execução; ter a certeza de que a telemetria reflete com 100% de precisão o estado real dos agentes.

### Persona 2: Líder Técnico / Analista de Governança
- **Perfil:** Responsável por auditar a esteira de desenvolvimento, garantindo conformidade com a Constituição (`sdd/memory/constitution.md`).
- **Dores:**
  - Falta de rastreabilidade temporal fina das micro-ações executadas pelos agentes autônomos.
  - Impossibilidade de verificar se as ferramentas invocadas respeitaram o menor privilégio (`SEG-06`) e os caminhos autorizados de arquivo.
- **Objetivo com a Feat 66:** Ter acesso a uma trilha de auditoria detalhada passo a passo com argumentos, arquivos tocados e durações, auditável tanto via UI quanto via JSON determinístico.

### Persona 3: Agentes Autônomos e Sessões do Pool ACP
- **Perfil:** Sessões do `SessionPool` (`sess-01-lider`, `sess-02-pm`, `sess-03-planner`, `sess-04-builder`, `sess-05-governance`) despachando micro-ações e testes.
- **Dores:**
  - Risco de atividades ficarem órfãs em caso de falha de conexão ou timeout de subprocesso.
- **Objetivo com a Feat 66:** Garantia de que a troca de estado para `standby` limpa automaticamente o contexto ativo, mantendo a fila consistente sem intervenção manual.

---

## 3. Casos de Uso Principais

### UC-01: Inspeção Detalhada de Micro-Ação no Forge Studio
- **Ator:** Desenvolvedor / Operador no navegador web.
- **Pré-condições:** Forge Studio em execução (`forge-sdd studio`) com atividades registradas no console.
- **Fluxo Principal:**
  1. O usuário clica em qualquer linha da tabela "Telemetria das Estações & Uso de Tools em Tempo Real".
  2. A interface abre o Drawer/Modal de Inspeção com transição suave.
  3. O modal carrega os detalhes completos do item selecionado:
     - Estação executora com badge colorido e ícone de mascote.
     - Ferramenta utilizada (ex: `run_command`, `read_file`, `portao_a`, `audit`).
     - Ação e comando executado com flags completas.
     - Argumentos e caminho de arquivos envolvidos.
     - Bloco de saída monoespaçado com o stdout/resultado da operação.
     - Timestamps de início/fim e duração calculada em ms.
  4. O usuário clica no botão "Copiar Comando" ou "Copiar Caminho" e recebe feedback de sucesso.
  5. O usuário navega entre ações anteriores e posteriores usando as setas do modal ou as teclas de teclado `←` e `→`.
  6. O usuário fecha o modal pressionando a tecla `ESC` ou clicando no botão `✕`.

### UC-02: Reconciliação Automática na Transição para Standby
- **Ator:** `SessionPool` ou `Aggregator` do Forge-SDD.
- **Pré-condições:** Existência de pelo menos uma atividade com `status: "running"` em `sdd/.queue/activity.json`.
- **Fluxo Principal:**
  1. Uma estação (ex: `builder` ou `governance`) conclui sua tarefa ou transiciona seu estado para `standby`/`idle`.
  2. O reconciliador verifica as atividades registradas daquela estação.
  3. Identifica as atividades com status `running`.
  4. Atualiza o status dessas atividades para `success` (ou `done`), define `completed_at` com o timestamp atual e calcula `duration_ms`.
  5. Persiste a alteração atomicamente em `sdd/.queue/activity.json` e emite evento SSE para atualização em tempo real no Studio.

### UC-03: Reconciliação no Registro de Nova Atividade
- **Ator:** Sistema de filas (`RecordActivity`).
- **Pré-condições:** A estação `planner` possui uma atividade `running` anterior não encerrada explicitamente.
- **Fluxo Principal:**
  1. A estação `planner` invoca uma nova tool ou ação via `RecordActivity`.
  2. Antes de inserir o novo registro no topo da lista, o reconciliador busca atividades anteriores da estação `planner` que estejam em `running`.
  3. Reconcilia as atividades anteriores para `success`, fechando seu ciclo de vida.
  4. Adiciona a nova atividade no início da fila com seu identificador único e metadados.
  5. Salva o arquivo de forma atômica e segura contra concorrência (`activityMu`).

### UC-04: Sanitização de Inicialização de Atividades Órfãs Históricas
- **Ator:** Inicialização do Studio ou CLI (`SanitizeOrphanActivities`).
- **Pré-condições:** Arquivo `activity.json` contendo atividades antigas em `running` de execuções passadas.
- **Fluxo Principal:**
  1. Durante o bootstrap de `forge-sdd studio` ou execução de comando de fila, o reconciliador lê `activity.json`.
  2. Detecta todas as atividades com `status: "running"` cuja idade seja superior a 60 segundos ou cujos agentes estejam ociosos.
  3. Reconcilia os itens para `done`, anotando em detalhes que foram saneados pelo boot do sistema.
  4. Persiste a coleção saneada, garantindo que o Studio inicie com 0 falsos-positivos de processos em andamento.

---

## 4. Requisitos Não-Funcionais e Restrições de Engenharia

1. **Go Standard Library Pura (`ARQ-03`, `ARQ-04`):** Toda a lógica do reconciliador, cálculo de tempos, parsing e persistência atômica utiliza estritamente `sync`, `time`, `os`, `path/filepath`, `encoding/json` e `crypto/sha256`. Zero bibliotecas externas adicionadas ao `go.mod`.
2. **Concorrência e Thread-Safety (`TST-01`):** A leitura, atualização e escrita de atividades deve ser totalmente thread-safe, protegida por mutex (`sync.Mutex` / `sync.RWMutex`), sem condições de corrida detectáveis com `go test -race ./...`.
3. **Limite Rígido de Memória e Armazenamento (`Orçamentos`):** O histórico de atividades mantém retenção máxima de 50 itens recentes. Cada campo `Output` possui truncamento seguro de no máximo 8 KB por item, evitando crescimento desenfreado do JSON no disco.
4. **Segurança e Sanitização (`SEG-01`, `SEG-02`, `SEG-06`):**
   - No frontend do Studio, qualquer string de output, comando ou argumento é devidamente escapada contra XSS via `escapeHtml()`.
   - Nenhum segredo ou variável privada é registrada na telemetria (`SEG-04`, `SEG-05`).
5. **Acessibilidade e Usabilidade:** O modal deve suportar foco gerenciado, navegação total por teclado (`ESC` para fechar, setas para navegar), e conformidade com contraste de cores do tema escuro do Forge Studio.

---

## 5. Próximos Passos
A aprovação deste documento de Discovery pelo Líder / Usuário habilita os critérios técnicos de engenharia em `criteria-66-telemetria-detalhada-inspecao-e-reconciliacao.md` e o plano de fatiamento em `plan-66-telemetria-detalhada-inspecao-e-reconciliacao.md`.
