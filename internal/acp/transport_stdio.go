package acp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

const (
	// DefaultMaxMessageSize limita o tamanho de uma linha de mensagem para 2 MB (SEG-01 proteção anti-DoS).
	DefaultMaxMessageSize = 2 * 1024 * 1024
)

// StdioTransport gerencia a comunicação bidirecional JSON-RPC 2.0 via streams de I/O (stdin/stdout ou pipes).
type StdioTransport struct {
	in             io.Reader
	out            io.Writer
	dispatcher     *RPCDispatcher
	writeMu        sync.Mutex
	maxMessageSize int
	concurrent     bool
	wg             sync.WaitGroup
	subChan        chan Notification
	stopChan       chan struct{}
	stopOnce       sync.Once
}

// NewStdioTransport cria uma instância de transporte conectando in e out ao despachador fornecido.
func NewStdioTransport(dispatcher *RPCDispatcher, in io.Reader, out io.Writer) *StdioTransport {
	if dispatcher == nil {
		dispatcher = NewDispatcher()
	}
	t := &StdioTransport{
		in:             in,
		out:            out,
		dispatcher:     dispatcher,
		maxMessageSize: DefaultMaxMessageSize,
		subChan:        make(chan Notification, 64),
		stopChan:       make(chan struct{}),
	}
	dispatcher.SubscribeNotifications(t.subChan)
	return t
}

// NewDefaultStdioTransport instancia o transporte utilizando os.Stdin e os.Stdout do sistema operacional.
func NewDefaultStdioTransport(dispatcher *RPCDispatcher) *StdioTransport {
	return NewStdioTransport(dispatcher, os.Stdin, os.Stdout)
}

// SetMaxMessageSize configura a capacidade máxima da linha recebida.
func (t *StdioTransport) SetMaxMessageSize(size int) {
	if size > 0 {
		t.maxMessageSize = size
	}
}

// SetConcurrent ativa ou desativa o processamento concorrente das requisições recebidas.
func (t *StdioTransport) SetConcurrent(v bool) {
	t.concurrent = v
}

// WriteLine escreve bytes finalizados com quebra de linha de forma thread-safe (COD-01).
func (t *StdioTransport) WriteLine(data []byte) error {
	t.writeMu.Lock()
	defer t.writeMu.Unlock()

	if _, err := t.out.Write(data); err != nil {
		return fmt.Errorf("transport write error: %w", err)
	}
	if _, err := t.out.Write([]byte("\n")); err != nil {
		return fmt.Errorf("transport newline write error: %w", err)
	}
	return nil
}

// WriteResponse serializa e escreve uma resposta JSON-RPC 2.0.
func (t *StdioTransport) WriteResponse(resp *Response) error {
	if resp == nil {
		return nil
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("failed to marshal rpc response: %w", err)
	}
	return t.WriteLine(data)
}

// WriteNotification serializa e escreve uma notificação JSON-RPC 2.0.
func (t *StdioTransport) WriteNotification(notif Notification) error {
	data, err := json.Marshal(notif)
	if err != nil {
		return fmt.Errorf("failed to marshal rpc notification: %w", err)
	}
	return t.WriteLine(data)
}

// Close encerra a escuta de notificações e aguarda tarefas em segundo plano.
func (t *StdioTransport) Close() error {
	t.stopOnce.Do(func() {
		close(t.stopChan)
		if t.dispatcher != nil && t.subChan != nil {
			t.dispatcher.UnsubscribeNotifications(t.subChan)
		}
	})
	t.wg.Wait()
	return nil
}

// Run executa o loop de leitura de mensagens até EOF no reader ou cancelamento do contexto.
func (t *StdioTransport) Run(ctx context.Context) error {
	defer func() {
		_ = t.Close()
	}()

	// Goroutine de retransmissão de notificações do dispatcher para a saída
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.stopChan:
				return
			case notif, ok := <-t.subChan:
				if !ok {
					return
				}
				_ = t.WriteNotification(notif)
			}
		}
	}()

	scanner := bufio.NewScanner(t.in)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, t.maxMessageSize)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.stopChan:
			return nil
		default:
		}

		if !scanner.Scan() {
			break
		}

		line := scanner.Bytes()
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 {
			continue
		}

		lineCopy := make([]byte, len(trimmed))
		copy(lineCopy, trimmed)

		if t.concurrent {
			t.wg.Add(1)
			go func(data []byte) {
				defer t.wg.Done()
				resp := t.dispatcher.Dispatch(ctx, data)
				if resp != nil {
					_ = t.WriteResponse(resp)
				}
			}(lineCopy)
		} else {
			resp := t.dispatcher.Dispatch(ctx, lineCopy)
			if resp != nil {
				if err := t.WriteResponse(resp); err != nil {
					return err
				}
			}
		}
	}

	if err := scanner.Err(); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("scanner error: %w", err)
	}

	return nil
}
