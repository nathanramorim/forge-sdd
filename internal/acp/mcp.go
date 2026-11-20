package acp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// MCPDoctorTool é o nome da única ferramenta exposta pelo servidor MCP.
const MCPDoctorTool = "forge_doctor"

const (
	mcpProtocolVersion = "2024-11-05"
	mcpMaxNome         = 64
	mcpMaxSaida        = 64 * 1024
)

// MCPDiagnoser produz o relatório (JSON já validado) do doctor. Quem chama fixa o diretório: o
// cliente MCP nunca escolhe caminho, comando ou argumento (somente leitura, SEG-01).
type MCPDiagnoser func() ([]byte, error)

type mcpToolCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// NewMCPDispatcher cria um despachador JSON-RPC mínimo do protocolo MCP (initialize, ping,
// tools/list, tools/call) com a ferramenta forge_doctor. Não registra nenhum método ACP.
func NewMCPDispatcher(version string, diagnose MCPDiagnoser) *RPCDispatcher {
	d := &RPCDispatcher{
		handlers:                make(map[string]HandlerFunc),
		notificationHandlers:    make(map[string]NotificationHandlerFunc),
		notificationSubscribers: make([]chan Notification, 0),
		serverVersion:           version,
		sessionManager:          NewSessionManager(),
	}
	d.RegisterHandler("initialize", func(context.Context, json.RawMessage) (any, *RPCError) {
		return map[string]any{
			"protocolVersion": mcpProtocolVersion,
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]any{"name": "forge-sdd", "version": mcpSaneia(version, 32)},
		}, nil
	})
	d.RegisterHandler("ping", func(context.Context, json.RawMessage) (any, *RPCError) {
		return map[string]any{}, nil
	})
	d.RegisterHandler("tools/list", func(context.Context, json.RawMessage) (any, *RPCError) {
		return map[string]any{"tools": []any{map[string]any{
			"name":        MCPDoctorTool,
			"description": "Diagnostica o harness do forge-sdd (somente leitura): binário, skills, contratos, worktrees e processos.",
			"inputSchema": map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false},
		}}}, nil
	})
	d.RegisterHandler("tools/call", func(_ context.Context, params json.RawMessage) (any, *RPCError) {
		var p mcpToolCall
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, NewRPCError(CodeInvalidParams, "parâmetros de tools/call inválidos")
		}
		if p.Name != MCPDoctorTool {
			return nil, NewRPCError(CodeInvalidParams, fmt.Sprintf("ferramenta desconhecida: %q", mcpSaneia(p.Name, mcpMaxNome)))
		}
		if args := strings.TrimSpace(string(p.Arguments)); args != "" && args != "null" && args != "{}" {
			return nil, NewRPCError(CodeInvalidParams, "forge_doctor não aceita argumentos")
		}
		if diagnose == nil {
			return nil, NewRPCError(CodeInternalError, "diagnóstico indisponível")
		}
		out, err := diagnose()
		if err != nil {
			return map[string]any{
				"isError": true,
				"content": []any{map[string]any{"type": "text", "text": mcpSaneia("falha no diagnóstico: "+err.Error(), 300)}},
			}, nil
		}
		if len(out) > mcpMaxSaida || !json.Valid(out) {
			return nil, NewRPCError(CodeInternalError, "relatório do doctor inválido")
		}
		return map[string]any{"content": []any{map[string]any{"type": "text", "text": string(out)}}}, nil
	})
	return d
}

// mcpSaneia remove controles, ESC e caracteres de formatação/bidi e limita o tamanho.
func mcpSaneia(s string, max int) string {
	s = strings.ToValidUTF8(s, "")
	s = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) || r == ' ' || r == ' ' {
			return ' '
		}
		return r
	}, s)
	s = strings.Join(strings.Fields(s), " ")
	if utf8.RuneCountInString(s) > max {
		s = string([]rune(s)[:max])
	}
	return s
}
