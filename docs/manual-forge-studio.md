# Manual Rápido & Protótipo do Forge Studio (`v3.2.0-beta`)

O **Forge Studio** é o módulo visual plugável do Forge-SDD que projeta no navegador o fluxo de trabalho autônomo entre as 5 estações de agentes de IA, com sincronização em tempo real via Server-Sent Events (SSE), filas individuais atômicas e inspetor de governança.

---

## ⚡ 1. Como Iniciar o Modo Visual

Você pode iniciar o Forge Studio diretamente via comando global `forge` ou através do `npx`:

### 🚀 Inicialização Rápida (Abre direto no navegador padrão)
```bash
# Comando global oficial (após npm install -g @nathanramorim/forge-sdd@beta):
forge studio --open

# Ou via npx (execução direta sem instalar globalmente):
npx @nathanramorim/forge-sdd@beta studio --open
```

### ⚙️ Opções e Flags Disponíveis
```bash
# Especificar porta personalizada (padrão: 4321)
forge studio --port 8080

# Especificar host de bind (padrão: 127.0.0.1 por segurança)
forge studio --host 127.0.0.1 --port 4321

# Apontar para outro diretório de projeto Forge-SDD
forge studio --dir ./meu-projeto --open
```

O servidor iniciará em `http://127.0.0.1:4321` exibindo no terminal:
```text
🌐 Forge Studio iniciado em http://127.0.0.1:4321
   Modo: Estritamente Read-Only (SEG-01 / SEG-06)
   Eventos SSE: http://127.0.0.1:4321/api/events
   Pressione Ctrl+C para encerrar.
```

---

## 🏗️ 2. Como Funciona por Baixo dos Panos

```mermaid
flowchart TD
    subgraph Repo["📁 Repositório Local (Sistema de Arquivos)"]
        Q["sdd/.queue/queue.json\n(Fila Central do Builder)"]
        QS["sdd/.queue/stations/*.json\n(Filas Individuais por Estação)"]
        C["sdd/memory/constitution.md\n(Pilares Constitucionais)"]
    end

    subgraph Core["⚙️ Forge Studio Backend (100% Go Stdlib)"]
        AG["Aggregator\n(Lê e cruza arquivos de estado)"]
        SH["SSE Hub\n(Detecta alterações de Hash/Versão)"]
        SRV["HTTP Server (127.0.0.1:4321)\nMiddleware Read-Only: Rejeita POST/PUT/DELETE com 405"]
        FS["embed.FS\n(Assets Web HTML/CSS/JS embutidos no binário)"]
    end

    subgraph Browser["🖥️ Navegador Web (Frontend Studio)"]
        UI["Interface Visual (index.html)"]
        M5["5 Mascotes Vetoriais SVG (CLS = 0)"]
        TH["Trilho de Handoff Animado"]
        IG["Inspetor de Governança (Camada A & B)"]
        SW["Chaveador LIVE (SSE) ⇄ DEMO (Simulação)"]
    end

    Repo -->|Leitura segura sem lock concorrente| AG
    AG --> SH
    SH -->|event: state (SSE)| Browser
    FS -->|GET /, /studio.css, /studio.js| Browser
    Browser -->|GET /api/state| SRV
```

### Princípios Arquiteturais e Garantias:
1. **Zero Dependências Go Externas (`ARQ-03` e `ARQ-04`):** O backend utiliza estritamente `net/http`, `embed.FS` e bibliotecas nativas de Go.
2. **Segurança Read-Only Absoluta (`SEG-01` e `SEG-06`):** O servidor atua apenas como observador. Qualquer tentativa de requisição mutante (`POST`, `PUT`, `DELETE`, `PATCH`) é bloqueada instantaneamente com código HTTP `405 Method Not Allowed`.
3. **Escrita Atômica e Idempotência nas Filas:** Cada estação possui sua fila (`sdd/.queue/stations/<station>.json`). Atualizações usam arquivos temporários + `os.Rename` atômico e cada pacote possui uma `IdempotencyKey` calculada via SHA-256.

---

## 🎨 3. Protótipo da Interface Visual

Abaixo está o layout estruturado projetado no navegador:

```text
+-------------------------------------------------------------------------------------------------------------------+
|  ● forge-sdd studio                [Projeto: meu-app]  [v3.2.0-beta]  [● LIVE SSE]  [ Modo: ( LIVE ) [ DEMO ] ]   |
+-------------------------------------------------------------------------------------------------------------------+
|                                                                                                                   |
|  ┌─ TRILHO DE HANDOFF ATIVO ───────────────────────────────────────────────────────────────────────────────────┐  |
|  │  [ 04. builder ] ───( ===[Pacote em Trânsito]====> )───> [ 05. governance ]                                │  |
|  │  Transição & Contrato: Artefatos Implementados & Suite de Testes Locais                [ ● Portão Validado ]│  |
|  └─────────────────────────────────────────────────────────────────────────────────────────────────────────────┘  |
|                                                                                                                   |
|  ┌─ GRADE DAS 5 ESTAÇÕES ──────────────────────────────────────────────────────────────────────────────────────┐  |
|  │ ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐         │  |
|  │ │  (O) 01. Líder  │ │ < > 02. PM/Arq  │ │ [=] 03. Planner │ │ ~~~ 04. Builder │ │ /_\ 05. Governan│         │  |
|  │ │     (Anel)      │ │    (Diamante)   │ │    (Barras)     │ │     (Ondas)     │ │   (Triângulo)   │         │  |
|  │ │ Status: [idle]  │ │ Status: [idle]  │ │ Status: [idle]  │ │ Status: [work]  │ │ Status: [approv]│         │  |
|  │ ├─────────────────┤ ├─────────────────┤ ├─────────────────┤ ├─────────────────┤ ├─────────────────┤         │  |
|  │ │Recebe: Demanda  │ │Recebe: Visão    │ │Recebe: Spec     │ │Recebe: Task     │ │Recebe: Diff     │         │  |
|  │ │Entrega: Alinhado│ │Entrega: Spec C4 │ │Entrega: Tasks   │ │Entrega: Código  │ │Entrega: Laudo   │         │  |
|  │ ├─────────────────┤ ├─────────────────┤ ├─────────────────┤ ├─────────────────┤ ├─────────────────┤         │  |
|  │ │Fila: 0 itens    │ │Fila: 0 itens    │ │Fila: 0 itens    │ │Fila: 2 itens    │ │Fila: 1 item     │         │  |
|  │ │(v1)             │ │(v1)             │ │(v1)             │ │(v4)             │ │(v3)             │         │  |
|  │ └─────────────────┘ └─────────────────┘ └─────────────────┘ └─────────────────┘ └─────────────────┘         │  |
|  └─────────────────────────────────────────────────────────────────────────────────────────────────────────────┘  |
|                                                                                                                   |
|  ┌─ INSPETOR DE GOVERNANÇA ──────────────────────────────────────────┬─ PILARES CONSTITUCIONAIS ───────────────┐  |
|  │ [✓] Parecer da Camada B: CONFORME                                 │  [ ARQ (1..4) OK ]  [ SEG (1..6) OK ]   │  |
|  │ Todas as regras de arquitetura, testes e ausência de segredos     │  [ DUP (1..5) OK ]  [ COD (1..5) OK ]   │  |
|  │ validadas rigorosamente de acordo com a Constituição.             │  [ TST (1..5) OK ]  [ FLX (1..4) OK ]   │  |
|  │                                                                   │  [ GOV (1..5) OK ]                      │  |
|  └───────────────────────────────────────────────────────────────────┴─────────────────────────────────────────┘  |
+-------------------------------------------------------------------------------------------------------------------+
|  forge-sdd studio • Runtime Go 100% Stdlib • Porta 4321                         Última atualização: 18:24:49      |
+-------------------------------------------------------------------------------------------------------------------+
```

---

## 🕹️ 4. Como Interagir e Testar na Prática

### 1. Testando com o Modo DEMO (Simulação Automática)
Clique no botão **`DEMO`** no canto superior direito da tela:
- O Studio ativa um motor de simulação que percorre as 5 estações sequencialmente a cada 2,2 segundos.
- Cada mascote altera suas expressões faciais (olhos e cor da antena) conforme o estado (`working`, `finished`, `approval`).
- O trilho de handoff demonstra a transição de pacotes entre as estações.
- Para voltar à escuta real do seu repositório, basta clicar no botão **`LIVE`**.

### 2. Disparando Handoffs Reais via CLI (Modo LIVE)
Com o Studio aberto em modo `LIVE`, abra outro terminal e execute transições na esteira:

```bash
# Enviar pacote do Builder para Governança
forge queue handoff \
  --from builder \
  --to governance \
  --task feat-01-exemplo \
  --gate "Testes Locais Conforme"

# Simular um loop de retrabalho da Governança de volta para o Builder
forge queue handoff \
  --from governance \
  --to builder \
  --task feat-01-exemplo \
  --gate "Ajuste de Assinatura TST-02" \
  --loop

# Inspecionar o estado de todas as estações pelo terminal
forge queue stations
```

A interface no navegador atualizará instantaneamente via Server-Sent Events, refletindo a nova rota e movimentação do pacote.
