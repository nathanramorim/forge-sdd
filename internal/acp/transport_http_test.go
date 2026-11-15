package acp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/forge-sdd/cli/internal/acp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// nonFlusherWriter simula um http.ResponseWriter que não implementa http.Flusher
type nonFlusherWriter struct {
	header http.Header
	code   int
	buf    bytes.Buffer
}

func newNonFlusherWriter() *nonFlusherWriter {
	return &nonFlusherWriter{header: make(http.Header)}
}

func (w *nonFlusherWriter) Header() http.Header         { return w.header }
func (w *nonFlusherWriter) Write(b []byte) (int, error) { return w.buf.Write(b) }
func (w *nonFlusherWriter) WriteHeader(statusCode int)  { w.code = statusCode }

func TestACPHTTPHandler_Direct(t *testing.T) {
	dispatcher := acp.NewDispatcherWithVersion("3.2.1-beta")
	handler := acp.NewHTTPHandler(dispatcher)

	t.Run("Getters and Setters", func(t *testing.T) {
		assert.Equal(t, dispatcher, handler.Dispatcher())
		assert.Equal(t, int64(acp.DefaultMaxMessageSize), handler.MaxMessageSize())
		handler.SetMaxMessageSize(1024 * 1024)
		assert.Equal(t, int64(1024*1024), handler.MaxMessageSize())
		handler.SetMaxMessageSize(acp.DefaultMaxMessageSize)
	})

	t.Run("Method Not Allowed for RPC", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/acp/rpc", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
		assert.Equal(t, "POST", rec.Header().Get("Allow"))
	})

	t.Run("Empty Body Error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/acp/rpc", strings.NewReader("   "))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		var rpcResp acp.Response
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&rpcResp))
		assert.NotNil(t, rpcResp.Error)
		assert.Equal(t, acp.CodeParseError, rpcResp.Error.Code)
	})

	t.Run("Payload Limit Exceeded", func(t *testing.T) {
		customHandler := acp.NewHTTPHandler(dispatcher)
		customHandler.SetMaxMessageSize(50)

		payload := strings.Repeat("x", 100)
		req := httptest.NewRequest(http.MethodPost, "/api/acp/rpc", strings.NewReader(payload))
		rec := httptest.NewRecorder()
		customHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
	})

	t.Run("Notification Handled (No Content 204)", func(t *testing.T) {
		notifPayload := `{"jsonrpc":"2.0","method":"custom/notify","params":{}}`
		req := httptest.NewRequest(http.MethodPost, "/api/acp/rpc", strings.NewReader(notifPayload))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("EventsHandler Method Not Allowed", func(t *testing.T) {
		eventsH := handler.EventsHandler()
		req := httptest.NewRequest(http.MethodPost, "/api/acp/events", nil)
		rec := httptest.NewRecorder()
		eventsH.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
		assert.Equal(t, "GET, HEAD", rec.Header().Get("Allow"))
	})

	t.Run("EventsHandler Non Flusher", func(t *testing.T) {
		eventsH := handler.EventsHandler()
		req := httptest.NewRequest(http.MethodGet, "/api/acp/events", nil)
		w := newNonFlusherWriter()
		eventsH.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.code)
		assert.Contains(t, w.buf.String(), "Streaming unsupported")
	})

	t.Run("EventsHandler Streaming Context Cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		req := httptest.NewRequest(http.MethodGet, "/api/acp/events", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		done := make(chan struct{})
		go func() {
			handler.ServeEvents(rec, req)
			close(done)
		}()

		// Aguarda o início e cancela
		time.Sleep(20 * time.Millisecond)
		cancel()

		select {
		case <-done:
		case <-time.After(1 * time.Second):
			t.Fatal("handler SSE não encerrou após cancelamento de contexto")
		}

		assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))
	})
}
