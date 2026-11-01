# Guia de Integração IDE com o Protocolo ACP (`forge-sdd acp`)

O **Protocolo ACP (Agent Client Protocol)** padroniza a comunicação bidirecional em tempo real entre IDEs modernas (como **Zed Editor**, **Cursor** e **VS Code**), o **Forge Studio** e a esteira de 5 estações autônomas do **forge-sdd**.

Baseado no padrão **JSON-RPC 2.0**, o ACP permite que ambientes de desenvolvimento externos interajam diretamente com o ciclo de vida do projeto: inicializar workspaces, enviar prompts e demandas para as filas das estações, receber streaming de eventos assíncronos e homologar portões de qualidade (`FLX-01`).

---

## 🚀 1. Pré-requisitos

Certifique-se de que o binário `forge-sdd` está compilado e acessível no seu `$PATH`:

```bash
# Se estiver trabalhando no repositório forge-sdd:
go install ./cmd/forge-sdd

# Ou via pacote npm global:
npm install -g @nathanramorim/forge-sdd@beta

# Verifique se o subcomando acp está disponível:
forge-sdd acp --help
```

---

## ⚡ 2. Configuração no Zed Editor

O **Zed Editor** suporta nativamente servidores de agente de IA através do protocolo ACP configurados no arquivo de preferências do usuário ou do projeto.

### 📝 Passo a Passo de Configuração

1. Abra o arquivo de configurações do Zed (`~/.config/zed/settings.json` ou `.zed/settings.json` na raiz do seu projeto).
2. Adicione ou edite o bloco `agent_servers`:

```json
{
  "agent_servers": {
    "forge-sdd": {
      "command": "forge-sdd",
      "args": ["acp"],
      "env": {}
    }
  }
}
```

> **Dica de Workspace Local:** Caso deseje restringir a execução ao diretório do projeto aberto no Zed, você pode especificar o argumento `--dir`:
> ```json
> {
>   "agent_servers": {
>     "forge-sdd": {
>       "command": "forge-sdd",
>       "args": ["acp", "--dir", "."]
>     }
>   }
> }
> ```

3. Reinicie o Zed ou recarregue a janela.
4. No painel do assistente de IA do Zed, selecione `forge-sdd` como o servidor de agente ativo. O handshake `initialize` e a criação de sessão via `session/new` ocorrerão automaticamente através do pipe stdio.

---

## 💻 3. Configuração no Cursor e VS Code

No **Cursor** e no **VS Code**, a integração pode ser realizada via extensões clientes ACP ou por meio de bridge de processo sobre stdio.

### Configuração com Extensões ACP (stdio bridge)

Se você utiliza uma extensão cliente compatível com ACP no VS Code ou Cursor, adicione a seguinte definição ao seu `settings.json`:

```json
{
  "acp.servers": {
    "forge-sdd": {
      "transport": "stdio",
      "command": "forge-sdd",
      "args": ["acp"],
      "rootPath": "${workspaceFolder}"
    }
  }
}
```

### Configuração via VS Code Task (Serviço em Background)

Caso prefira rodar o servidor em background integrado ao terminal da IDE:

`.vscode/tasks.json`:
```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Forge-SDD: Servidor ACP (stdio)",
      "type": "shell",
      "command": "forge-sdd acp",
      "isBackground": true,
      "problemMatcher": [],
      "presentation": {
        "reveal": "silent",
        "panel": "dedicated"
      }
    }
  ]
}
```

---

## 🌐 4. Modos de Transporte Suportados

O protocolo ACP no forge-sdd suporta dois modos de transporte:

| Transporte | Canal de Comunicação | Cenário de Uso | Comando / Endpoint |
|---|---|---|---|
| **stdio** (Padrão) | `os.Stdin` / `os.Stdout` via pipes | IDEs locais (Zed, Cursor, VS Code) e scripts CLI | `forge-sdd acp` |
| **HTTP + SSE** | HTTP POST + Server-Sent Events | Forge Studio, dashboards web e serviços desacoplados | POST `/acp`<br/>GET `/acp/events` |

---

## 📜 5. Exemplos de Mensagens JSON-RPC 2.0

### 5.1. Handshake Inicial (`initialize`)

Enviado pelo cliente assim que o processo stdio é iniciado:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "1.0.0",
    "clientInfo": {
      "name": "zed-editor",
      "version": "0.140.0"
    },
    "capabilities": {
      "terminal": true,
      "prompts": true
    }
  }
}
```

**Resposta do Servidor:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "1.0.0",
    "serverInfo": {
      "name": "forge-sdd-acp",
      "version": "3.2.1-beta"
    },
    "capabilities": {
      "sessions": true,
      "gates": true,
      "prompts": true
    }
  }
}
```

---

### 5.2. Criar Sessão de Trabalho (`session/new`)

Cria uma sessão vinculada ao diretório do workspace (`workspaceDir` com validação de path traversal):

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "session/new",
  "params": {
    "workspaceDir": "/caminho/do/projeto",
    "feature": "feat-64-protocolo-acp",
    "branch": "feat/64-protocolo-acp"
  }
}
```

**Resposta do Servidor:**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "sessionId": "session-1730419200000000000",
    "status": "active",
    "workspaceDir": "/caminho/do/projeto"
  }
}
```

---

### 5.3. Enviar Demanda para uma Estação (`session/prompt`)

Enfileira uma tarefa na fila determinística da estação indicada (ex: `concepcao`, `planejador`, `builder`, `governanca`):

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "session/prompt",
  "params": {
    "sessionId": "session-1730419200000000000",
    "station": "concepcao",
    "prompt": "Elaborar especificação técnica da feature 64",
    "operation": "spec",
    "metadata": {
      "autor": "zed-user"
    }
  }
}
```

**Resposta do Servidor:**
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "sessionId": "session-1730419200000000000",
    "station": "concepcao",
    "status": "queued",
    "item": {
      "id": "prompt-1730419200100000000",
      "title": "Elaborar especificação técnica da feature 64",
      "station": "concepcao",
      "operation": "spec",
      "status": "queued"
    },
    "idempotentHit": false
  }
}
```

---

### 5.4. Notificação Assíncrona de Streaming (`session/notification`)

Emitida automaticamente pelo servidor sem requisição direta (`id` omitido):

```json
{
  "jsonrpc": "2.0",
  "method": "session/notification",
  "params": {
    "sessionId": "session-1730419200000000000",
    "type": "prompt_enqueued",
    "station": "concepcao",
    "message": "Item prompt-1730419200100000000 enfileirado na estação concepcao",
    "timestamp": "2026-11-01T00:00:00Z"
  }
}
```

---

### 5.5. Homologação do Portão FLX-01 (`gate/approve`)

Valida os critérios canônicos de governança (`P-01`, `P-02`, `P-03`) e realiza o handoff automático para a estação seguinte:

```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "gate/approve",
  "params": {
    "sessionId": "session-1730419200000000000",
    "specPath": "sdd/features/feat-64-protocolo-acp/feat-64-04-exemplos-ide-docs-e2e.md",
    "branch": "feat/64-protocolo-acp",
    "fromStation": "concepcao",
    "toStation": "planejador"
  }
}
```

**Resposta em caso de Sucesso:**
```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "result": {
    "featureId": "feat-64-04-exemplos-ide-docs-e2e",
    "specPath": "sdd/features/feat-64-protocolo-acp/feat-64-04-exemplos-ide-docs-e2e.md",
    "gate": "FLX-01",
    "status": "approved",
    "fromStation": "concepcao",
    "toStation": "planejador",
    "handoffItem": {
      "id": "feat-64-04-exemplos-ide-docs-e2e",
      "title": "Feature feat-64-04-exemplos-ide-docs-e2e",
      "station": "planejador",
      "status": "queued"
    }
  }
}
```

**Resposta em caso de Violação (Código `-32001`):**
```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "error": {
    "code": -32001,
    "message": "Portão FLX-01 rejeitado: [P-03/DUP-02]: tabela '## O que já existe' ausente na Spec",
    "data": {
      "gate": "FLX-01"
    }
  }
}
```

---

### 5.6. Encerrar Sessão (`session/cancel`)

```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "method": "session/cancel",
  "params": {
    "sessionId": "session-1730419200000000000",
    "reason": "Feature concluída com sucesso"
  }
}
```

---

## 🔒 6. Boas Práticas e Segurança (`SEG-01`, `SEG-02`, `SEG-04`)

1. **Isolamento de Workspaces:** Cada sessão opera estritamente no diretório validado. Tentativas de acessar diretórios superiores via `..` ou strings malformadas com caracteres nulos (`\0`) são sumariamente rejeitadas com erro `-32602`.
2. **Proteção Anti-DoS:** O buffer de linha do transporte stdio possui teto de 2 MB (`DefaultMaxMessageSize`), prevenindo estouro de memória em mensagens desproporcionais.
3. **Ausência de Segredos:** Nunca passe chaves privadas, senhas ou tokens de API em metadados de sessão ou prompts ACP (`SEG-04`).
