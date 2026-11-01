# feat-65-03 — Subcomando CLI `forge-sdd attach` (`cmd/forge-sdd/attach.go`) e Multiplexador de Terminal (`internal/acp/attach.go`)

**Branch:** `feat/65-prewarming-pool-sessoes-acp-estacoes`
**Fase:** 65-03
**Tamanho:** `M`
**Depende de:** `feat-65-01`, `feat-65-02`
**Status:** `ready`

## Contexto & Motivação

Com a orquestração autônoma do pool pre-warmed de sessões ACP, a necessidade de manter 5 terminais abertos manualmente é eliminada. Contudo, o desenvolvedor ainda precisa inspecionar em tempo real o que uma estação está fazendo (ex: logs de raciocínio, ferramentas sendo executadas pelo Builder, validações de Governança) ou intervir diretamente com comandos e instruções pontuais.

Esta subfeature implementa o subcomando CLI `forge-sdd attach [estacao]` e o multiplexador de terminal `AttachBridge`:
1. Implementar o `AttachBridge` em [`internal/acp/attach.go`](file:///Users/nathanramorim/git/forge-sdd/internal/acp/attach.go), oferecendo streaming pub/sub thread-safe para que múltiplos clientes possam observar a mesma sessão simultaneamente sem concorrência espúria (`COD-01`, `TST-01`).
2. Implementar o método RPC `session/attach`, permitindo conexão tanto via CLI quanto via IDEs compatíveis com ACP (Zed, VS Code).
3. Criar o subcomando Cobra [`cmd/forge-sdd/attach.go`](file:///Users/nathanramorim/git/forge-sdd/cmd/forge-sdd/attach.go) seguindo estritamente a convenção da skill `novo-subcomando-cli-sdd` (`DUP-05`), mantendo separação estrita de fronteiras (`ARQ-01`, `ARQ-06`).
4. Suportar desconexão graciosa: quando o desenvolvedor encerra a visualização (`Ctrl+C` ou comando `exit`), o cliente é desanexado (`Detach`) sem que a sessão subjacente no pool seja abortada ou alterada de estado.
5. Validar o argumento de estação contra `queue.CanonicalStations()` (`SEG-01`), rejeitando nomes inválidos com mensagens explicativas.
6. Testes unitários e de comando CLI com concorrência segura (`go test -race`) em [`cmd/forge-sdd/commands_test.go`](file:///Users/nathanramorim/git/forge-sdd/cmd/forge-sdd/commands_test.go) e [`internal/acp/attach_test.go`](file:///Users/nathanramorim/git/forge-sdd/internal/acp/attach_test.go).

## Modelo do módulo

- **Modelo:** `enxuto` dentro de [`internal/acp`](file:///Users/nathanramorim/git/forge-sdd/internal/acp) e [`cmd/forge-sdd`](file:///Users/nathanramorim/git/forge-sdd/cmd/forge-sdd).
- **Justificativa:** O comando CLI atua unicamente como adaptador de entrada/saída do terminal (I/O bridge), conectando `os.Stdin` e `os.Stdout` ao canal de eventos de `internal/acp`. Não contém lógica de negócio própria (`ARQ-01`, `ARQ-06`).

## O que já existe (Reuse Check — `DUP-02`)

| Procurei | Onde | Achei | Decisão (`reusar`, `estender`, `criar`) | Justificativa |
|---|---|---|---|---|
| Comandos da CLI Cobra | [`cmd/forge-sdd/`](file:///Users/nathanramorim/git/forge-sdd/cmd/forge-sdd) | Subcomandos `acp.go`, `studio.go`, `gate.go`, `queue.go` | `estender` | Adicionar `cmd/forge-sdd/attach.go` e registrar em `cmd/forge-sdd/main.go` conforme a skill `novo-subcomando-cli-sdd` (`DUP-05`). |
| Lista de Estações Canônicas | [`internal/queue/stations.go`](file:///Users/nathanramorim/git/forge-sdd/internal/queue/stations.go) | `CanonicalStations()`, `IsValidStation()` | `reusar` | O comando valida o parâmetro de entrada exclusivamente contra as 5 estações canônicas de `internal/queue` (`DUP-01`). |
| RPC Dispatcher e Sessões | [`internal/acp/dispatcher.go`](file:///Users/nathanramorim/git/forge-sdd/internal/acp/dispatcher.go) | `RPCDispatcher` | `estender` | Registrar o método `session/attach` no despachante RPC. |
| Multiplexador e Bridge de Terminal | `internal/acp` | Inexistente | `criar` | Criar `internal/acp/attach.go`, `internal/acp/attach_test.go` e `cmd/forge-sdd/attach.go`. |

## Contrato de Entrada e Saída

### Contrato de Entrada CLI
```bash
forge-sdd attach <estacao> [--dir <path>]
```
- `<estacao>`: Obrigatório. Deve ser um dos valores: `lider`, `pm-architect`, `planner`, `builder`, `governance`.
- `--dir`: Opcional (padrão: `"."`).

### Contrato de Entrada RPC (`session/attach`)
```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "method": "session/attach",
  "params": {
    "station": "builder"
  }
}
```

### Contrato de Saída RPC
```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "result": {
    "sessionId": "sess-04-builder",
    "station": "builder",
    "state": "processing",
    "activeItemId": "feat-65-01-nucleo-pool",
    "history": [
      {"timestamp": "2026-11-01T02:00:00Z", "type": "info", "message": "Sessão inicializada em standby"},
      {"timestamp": "2026-11-01T02:05:00Z", "type": "tool_call", "message": "Executando go test -v -race"}
    ]
  }
}
```

## Regras Constitucionais e IDs Aplicáveis

- **Rules:** `.agents/rules/sdd-arquitetura.md`, `.agents/rules/sdd-ergonomia.md`, `.agents/rules/sdd-seguranca.md`, `.agents/rules/sdd-codigo-e-testes.md`.
- **IDs:** `ARQ-01`, `ARQ-06`, `SEG-01`, `SEG-04`, `DUP-02`, `DUP-05`, `COD-01`, `COD-02`, `COD-03`, `TST-01`, `TST-02`, `TST-03`, `TST-04`, `TST-05`.
- **Skills Reutilizadas:** `novo-subcomando-cli-sdd`.

## Critérios de Aceite Gherkin

```gherkin
Cenário: Anexação interativa via terminal CLI a uma estação específica
  Dado que o pool de sessões ACP está em execução no workspace
  Quando o usuário executar o comando:
    """
    forge-sdd attach builder
    """
  Então o comando deve exibir um banner de cabeçalho informando o ID da sessão e estado atual
  E deve exibir os eventos e logs emitidos pela estação em tempo real
  E deve permitir ao usuário digitar comandos interativos

Cenário: Desconexão graciosa sem encerrar a sessão
  Dado que o usuário está com o terminal anexado à estação "builder"
  Quando o usuário pressionar Ctrl+C ou digitar "exit"
  Então o comando attach deve finalizar com código de saída 0
  E a sessão "sess-04-builder" no pool deve permanecer ativa e com seu estado inalterado

Cenário: Rejeição de estação inválida com mensagem explicativa
  Dado que o usuário executa "forge-sdd attach devops"
  Quando o comando validar o parâmetro de entrada
  Então deve falhar com erro amigável informando que a estação é inválida
  E deve listar as 5 estações canônicas válidas: lider, pm-architect, planner, builder, governance
```

## Plano de Implementação Passo a Passo

1. **Implementação de `AttachBridge` em `internal/acp/attach.go`:**
   - Criar `AttachSubscriber` contendo canal de notificações `chan Notification`.
   - Adicionar métodos a `PooledSession`: `AttachSubscriber() *AttachSubscriber` e `DetachSubscriber(sub *AttachSubscriber)`.
   - Implementar método `AttachBridge.Run(ctx context.Context, in io.Reader, out io.Writer)` para streaming contínuo.
2. **Registro do Método RPC `session/attach`:**
   - Adicionar o handler de `session/attach` no `RPCDispatcher` retornando metadados e histórico da sessão.
3. **Criação do Subcomando Cobra `cmd/forge-sdd/attach.go`:**
   - Definir `attachCmd` com suporte ao argumento posicional `estacao`.
   - Validar estação contra `queue.IsValidStation`.
   - Conectar os streams `os.Stdin` e `os.Stdout` ao bridge.
   - Tratar sinais de interrupção (`signal.Notify` para `os.Interrupt`) para desconexão graciosa.
4. **Registro em `cmd/forge-sdd/main.go`:**
   - Adicionar `rootCmd.AddCommand(attachCmd)` conforme a convenção da skill `novo-subcomando-cli-sdd`.
5. **Testes Automatizados:**
   - Criar `internal/acp/attach_test.go` testando múltiplos assinantes concorrentes e desconexão via context.
   - Criar teste do subcomando em `cmd/forge-sdd/commands_test.go` verificando validação de flags e rejeição de entradas inválidas.

## Comandos de Verificação Determinísticos

```bash
go test -v -race -run "TestAttach" ./cmd/forge-sdd/... ./internal/acp/... && go vet ./...
```
