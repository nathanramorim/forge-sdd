package acp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mcpCall(t *testing.T, d *RPCDispatcher, msg string) map[string]any {
	t.Helper()
	resp := d.Dispatch(context.Background(), []byte(msg))
	require.NotNil(t, resp)
	data, err := json.Marshal(resp)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	return m
}

func TestMCPForgeDoctor(t *testing.T) {
	d := NewMCPDispatcher("1.0", func() ([]byte, error) { return []byte(`{"versao":1}`), nil })

	lista := mcpCall(t, d, `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
	assert.Contains(t, mustJSON(lista), MCPDoctorTool)

	ok := mcpCall(t, d, `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"forge_doctor"}}`)
	assert.Nil(t, ok["error"])
	assert.Contains(t, mustJSON(ok), `versao`)

	// negativo: ferramenta desconhecida (com ESC e bidi) devolve erro JSON-RPC sem eco cru
	ruim := mcpCall(t, d, "{\"jsonrpc\":\"2.0\",\"id\":3,\"method\":\"tools/call\",\"params\":{\"name\":\"x\\u001b[31m\\u202e\\u0000\"}}")
	e, _ := ruim["error"].(map[string]any)
	require.NotNil(t, e)
	assert.EqualValues(t, CodeInvalidParams, e["code"])
	raw := mustJSON(ruim)
	assert.NotContains(t, raw, `\u001b`)
	assert.NotContains(t, raw, `‮`)

	// negativo: argumentos hostis são recusados e nada é executado
	chamado := false
	d2 := NewMCPDispatcher("1.0", func() ([]byte, error) { chamado = true; return []byte(`{}`), nil })
	arg := mcpCall(t, d2, `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"forge_doctor","arguments":{"dir":"/etc"}}}`)
	assert.NotNil(t, arg["error"])
	assert.False(t, chamado)

	// negativo: método ACP não existe neste servidor
	acp := mcpCall(t, d, `{"jsonrpc":"2.0","id":5,"method":"session/new"}`)
	assert.NotNil(t, acp["error"])

	// falha do diagnóstico vira isError sanitizado; saída que não é JSON vira erro
	d3 := NewMCPDispatcher("1.0", func() ([]byte, error) { return nil, errors.New("boom\x1b[2J" + strings.Repeat("a", 1000)) })
	f := mustJSON(mcpCall(t, d3, `{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"forge_doctor"}}`))
	assert.Contains(t, f, "isError")
	assert.NotContains(t, f, `\u001b`)
	d4 := NewMCPDispatcher("1.0", func() ([]byte, error) { return []byte("não json"), nil })
	assert.NotNil(t, mcpCall(t, d4, `{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"forge_doctor"}}`)["error"])
}

func TestMCPStdioPontaAPonta(t *testing.T) {
	d := NewMCPDispatcher("1.0", func() ([]byte, error) { return []byte(`{"versao":1}`), nil })
	in := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n" +
		`{"jsonrpc":"2.0","method":"notifications/initialized"}` + "\n" +
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"forge_doctor"}}` + "\n")
	var out bytes.Buffer
	require.NoError(t, NewStdioTransport(d, in, &out).Run(context.Background()))
	assert.Equal(t, 2, strings.Count(strings.TrimSpace(out.String()), "\n")+1)
}

func mustJSON(v any) string { b, _ := json.Marshal(v); return string(b) }

func TestMCPIdInvalidoRetornaErro(t *testing.T) {
	d := NewMCPDispatcher("1.0", nil)
	m := mcpCall(t, d, `{"jsonrpc":"2.0","id":1e999,"method":"ping"}`)
	e, _ := m["error"].(map[string]any)
	require.NotNil(t, e)
	assert.EqualValues(t, CodeInvalidRequest, e["code"])
	assert.Nil(t, m["id"])
}
