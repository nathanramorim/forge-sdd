# Discovery — feat-64: Integração do Protocolo ACP (Agent Client Protocol) no Forge-SDD e Forge Studio

## 1. Visão Geral & Problema

### 1.1 O que é o Protocolo ACP (Agent Client Protocol)
O **Agent Client Protocol (ACP)** é um padrão aberto de comunicação — concebido e impulsionado por ecossistemas modernos de desenvolvimento como a IDE Zed, Anthropic e JetBrains — projetado para padronizar e desacoplar a interação entre **editores de código / clientes de usuário (IDEs)** e **agentes autônomos de engenharia de software**.

Enquanto o **LSP (Language Server Protocol)** revolucionou o suporte a linguagens ao unificar autocompletion, diagnóstico e navegação de símbolos sob um contrato único entre IDEs e analisadores estáticos, o **ACP** estabelece a mesma revolução para a era dos agentes de IA:
- Um editor não precisa implementar adaptadores proprietários para cada agente;
- Um agente ou esteira de engenharia não precisa se reescrever para cada IDE;
- A comunicação é estruturada via **JSON-RPC 2.0**, com suporte nativo a transporte por `stdio` (processos locais spawned pela IDE) ou por `HTTP/SSE/WebSocket` (serviços remotos e dashboards web).

### 1.2 ACP vs. MCP (Model Context Protocol): O Papel de Cada Um
É fundamental distinguir o papel do ACP frente ao **MCP (Model Context Protocol)** da Anthropic:
- **MCP (Model Context Protocol):** Conecta o **Agente às Ferramentas e Fontes de Dados** (banco de dados, busca de documentação context7, APIs, leitura do filesystem). Responde à pergunta: *"Quais ferramentas e contextos o agente pode invocar?"*
- **ACP (Agent Client Protocol):** Conecta o **Cliente/IDE ao Agente/Esteira** (sessão de conversa, streaming de raciocínio, aprovação de permissões/portões, notificações de estado e controle de ciclo de vida). Responde à pergunta: *"Como a IDE orquestra e apresenta a esteira de agentes para o desenvolvedor?"*

Ambos são complementares: o editor interage com o Forge-SDD via **ACP**, e as estações do Forge-SDD utilizam **MCP** para interagir com o ecossistema externo de ferramentas.

### 1.3 O Problema no Forge-SDD Hoje
Atualmente, o Forge-SDD opera com um modelo híbrido eficiente, porém fragmentado para integração externa:
1. **Chatmodes e CLI:** Os agentes operam acionados via chatmodes específicos de ferramentas (GitHub Copilot, Claude CLI, Gemini CLI) que leem instruções estáticas em `.gemini/`, `.claude/` ou executam subcomandos CLI (`forge-sdd queue`, `forge-sdd gate`).
2. **Forge Studio Local:** O Forge Studio (`internal/studio`) provê um cockpit visual read-only com mascotes 3D e SSE para visualização em tempo real de filas, mas a interface conversa com o backend através de endpoints HTTP/SSE proprietários (`/api/state`, `/api/events`, `/api/spec/approve`).
3. **Ausência de Contrato Universal para IDEs Modernas:** IDEs como **Zed**, **Cursor**, **VS Code** (com extensões ACP) e **JetBrains** não conseguem reconhecer o Forge-SDD como um agente de engenharia plug-and-play. O desenvolvedor é forçado a alternar abas, copiar/colar comandos ou operar exclusivamente no terminal.

### 1.4 Por que Adotar ACP no Forge-SDD
Ao implementar um servidor ACP nativo em Go (stdlib pura):
- **Forge-SDD como Agente Nativo na IDE:** O Forge-SDD torna-se um **ACP Server** executável via `forge-sdd acp` (transporte `stdio`), permitindo que a IDE Zed ou extensões ACP no VS Code/Cursor adicionem o Forge-SDD diretamente na barra de agentes externos.
- **Unificação com o Forge Studio:** O Forge Studio reutiliza o transporte ACP (HTTP/SSE) como barramento universal de eventos, simplificando os contratos de comunicação entre a interface web e as 5 estações.
- **Interoperabilidade Total com Agentes Heterogêneos:** Agentes locais ou em nuvem (Claude 3.7 Sonnet, Gemini 2.5 Pro, Codex, GPT-4o) podem participar da esteira SDD obedecendo ao ciclo estrito de governança (`ARQ-*`, `SEG-*`, `DUP-*`, `FLX-*`).

---

## 2. Personas

### Persona 1: Desenvolvedor Usando IDE Moderna (Zed, VS Code, Cursor, JetBrains)
- **Perfil:** Engenheiro de software focado no código, utilizando Zed, Cursor ou VS Code como ferramenta primária.
- **Dores:**
  - Ter que sair do editor para inspecionar filas em `sdd/.queue/` ou abrir abas extras no terminal.
  - Falta de feedback de streaming sobre o que o agente do SDD está raciocinando ou executando no momento.
  - Dificuldade de revisar e aprovar especificações (Portão `FLX-01`) de forma integrada à sua árvore de arquivos e painel lateral.
- **Objetivo com ACP:** Configurar o Forge-SDD em seu `settings.json` da IDE como um agente ACP externo; despachar comandos (`/discovery`, `/nova-feature`), ver o progresso das 5 estações em tempo real na aba de agentes e aprovar portões com 1 clique no editor.

### Persona 2: Desenvolvedor Usando o Forge Studio (Web / Local Cockpit)
- **Perfil:** Desenvolvedor ou líder técnico que prefere o cockpit visual com os mascotes 3D das 5 estações, visão kanban de filas e trilho de handoff.
- **Dores:**
  - Endpoints HTTP ad-hoc e dependência de polling/SSE com esquemas customizados.
  - Impossibilidade de conectar agentes externos arbitrários ao dashboard do Studio sem reescrever o agregador de estado.
- **Objetivo com ACP:** O Studio web comunica-se com o motor do Forge-SDD via transporte ACP (HTTP/SSE), exibindo streams padronizados de pensamento, execução de ferramentas e handoffs entre estações de forma fluida e desacoplada.

### Persona 3: Agentes Autônomos Heterogêneos
- **Perfil:** LLMs e runtimes autônomos (Claude Code, Gemini CLI, Cursor Agent, agentes customizados) operando em sessões de background.
- **Dores:**
  - Falta de um protocolo formal de entrada e saída para receber a demanda da estação anterior e devolver o artefato com métricas determinísticas.
- **Objetivo com ACP:** Comunicar-se via chamadas padronizadas JSON-RPC 2.0 (`session/new`, `session/prompt`, `session/update`), recebendo regras da Constituição (`ARQ-*`, `SEG-*`) injetadas no contexto da sessão e respeitando os portões de qualidade.

---

## 3. Casos de Uso Principais

### UC-01: Inicialização e Handshake ACP (`initialize`)
- **Ator:** IDE Client (Zed / VS Code / Cursor) ou Forge Studio Web.
- **Fluxo:**
  1. O cliente inicia o subprocesso `forge-sdd acp` via `stdio` (ou conecta via HTTP em `http://127.0.0.1:4321/api/acp`).
  2. O cliente envia a mensagem JSON-RPC `initialize` contendo informações do editor e capacidades suportadas.
  3. O Forge-SDD responde com suas capacidades: suporte a sessões (`sessions`), streaming de tokens/notificações (`streaming`), catálogo de ferramentas das estações (`tools`), e suporte nativo a aprovação de portões (`gates`).

### UC-02: Criação e Carregamento de Sessão SDD (`session/new`, `session/load`)
- **Ator:** Desenvolvedor na IDE.
- **Fluxo:**
  1. O usuário abre uma nova conversa ou seleciona uma feature existente no painel da IDE.
  2. A IDE dispara `session/new` passando o caminho do workspace.
  3. O servidor ACP do Forge-SDD carrega o estado da esteira (`sdd/memory/progress.md`, filas ativas em `sdd/.queue/stations/`) e associa a sessão ao ciclo da feature atual.
  4. O servidor devolve a confirmação da sessão com o contexto resumido da esteira (estação ativa, mascotes, tarefas pendentes).

### UC-03: Streaming de Atividades e Raciocínio das 5 Estações
- **Ator:** IDE Client e Forge Studio.
- **Fluxo:**
  1. Enquanto uma estação processa uma tarefa (ex: Concepção gerando a Spec, ou Builder compilando testes Go), o Forge-SDD emite notificações JSON-RPC `session/notification` contendo:
     - `station`: Identificador da estação canônica (`lider`, `pm-architect`, `planner`, `builder`, `governance`).
     - `state`: Estado do mascote/estação (`thinking`, `working`, `searching`, `approval`).
     - `tool_call`: Ferramentas executadas em tempo real (`run_command`, `read_file`, `write_file`).
     - `content_delta`: Texto ou diffs gerados incrementalmente.
  2. A IDE exibe o stream contínuo no chat/painel lateral; o Forge Studio anima o mascote e atualiza a barra de progresso instantaneamente.

### UC-04: Resolução de Portão Interativo Humano (`FLX-01`) via ACP
- **Ator:** Desenvolvedor e Estação 02 (PM & Arquiteto).
- **Fluxo:**
  1. Ao concluir a elaboração dos artefatos de spec (`discovery-XX`, `criteria-XX`, `plan-XX`), o Forge-SDD suspende o avanço automático da fila e emite um evento de portão pendente `gate/approval_required` com os dados da Spec e checklist.
  2. A IDE exibe uma interface interativa de aprovação (botão "Aprovar Spec" ou comando ACP).
  3. O desenvolvedor revisa e clica em "Aprovar".
  4. A IDE envia a chamada RPC `gate/approve` para o servidor ACP.
  5. O servidor valida a integridade da spec (regras `P-02` e `P-03` do Portão Camada A), transiciona o item na fila para `approved`, aciona o handoff para o Planejador (Estação 03) e propaga a notificação de desbloqueio.

### UC-05: Bridge de Transporte Dual (Stdio & HTTP/SSE)
- **Ator:** IDEs locais e Aplicações Web/Remotas.
- **Fluxo:**
  1. No modo CLI local, `forge-sdd acp` escuta em `stdin` e responde em `stdout`, permitindo que editores locais como Zed e VS Code gerenciem o ciclo de vida do processo nativamente.
  2. No modo Studio, o mesmo motor de parsing e despacho ACP é acoplado aos handlers HTTP/SSE de `internal/studio`, permitindo que clientes web ou agentes via rede se conectem via streaming HTTP sem duplicar regras de negócio.

---

## 4. Escopo da Iniciativa

### 4.1 No Escopo (In-Scope)
- Implementação do servidor ACP em **Go stdlib pura** (`internal/acp`), sem bibliotecas externas (`ARQ-04`, `SEG-07`).
- Suporte aos métodos essenciais do padrão ACP: `initialize`, `session/new`, `session/load`, `session/prompt`, `session/cancel`, `gate/approve`.
- Notificações de streaming JSON-RPC 2.0 para atualização de estado das 5 estações e telemetria.
- Novo subcomando CLI determinístico: `forge-sdd acp` com transporte `stdio` (`ARQ-06`).
- Extensão do Forge Studio (`internal/studio`) com endpoint HTTP/SSE compatível com ACP (`/api/acp/rpc` e `/api/acp/events`).
- Integração bidirecional com o subsistema de filas (`internal/queue`) e portões (`internal/gate`).
- Documentação de setup e configuração para Zed (`settings.json`), Cursor e VS Code.

### 4.2 Fora do Escopo (Out-of-Scope)
- Desenvolvimento de extensões de IDE proprietárias em TypeScript ou Rust (o foco é fornecer o servidor ACP padrão para o qual qualquer extensão ou IDE compatível pode se conectar).
- Modificação dos modelos fundacionais de IA (o servidor atua como protocolo de intermediação e governança).
- Transmissão de áudio ou vídeo (apenas payloads de texto, eventos e JSON-RPC).
