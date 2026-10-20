// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package mobile

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"sync"
)

// ResponseWriter receives a response from Engine.Serve. The host implements
// it, typically by forwarding to a WKURLSchemeTask. WriteHeader is called
// once, then Write any number of times, then Finish once. Calls come from a
// background goroutine, never concurrently for the same request.
type ResponseWriter interface {
	WriteHeader(status int, headersJSON string)
	Write(data []byte)
	Finish()
}

// Request is a request being served. Cancel it when the host no longer wants
// the response, such as when the web view stops a scheme task.
type Request struct {
	cancel context.CancelFunc
}

// Cancel stops the request. Finish is still called on its ResponseWriter.
func (r *Request) Cancel() {
	r.cancel()
}

// Serve handles one request from the web view in the background. rawURL may
// use any scheme and host; only its path and query are routed. headersJSON is
// a JSON object of header names to values and may be empty.
func (e *Engine) Serve(method, rawURL, headersJSON string, body []byte, w ResponseWriter) *Request {
	ctx, cancel := context.WithCancel(context.Background())
	req := &Request{cancel: cancel}

	go func() {
		defer cancel()

		rw := newBridgeWriter(w)
		defer rw.finish()

		defer func() {
			p := recover()
			if p == nil {
				return
			}

			if p != http.ErrAbortHandler {
				e.app.Log().Error().Interface("panic", p).Bytes("stack", debug.Stack()).Str("url", rawURL).Msg("request panicked")
			}

			rw.abort()
		}()

		r, err := newHTTPRequest(ctx, method, rawURL, headersJSON, body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		e.app.Handler().ServeHTTP(rw, r)
	}()

	return req
}

func newHTTPRequest(ctx context.Context, method, rawURL, headersJSON string, body []byte) (*http.Request, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequestWithContext(ctx, method, u.RequestURI(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	if headersJSON != "" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(headersJSON), &headers); err != nil {
			return nil, err
		}

		for k, v := range headers {
			r.Header.Set(k, v)
		}
	}

	r.Host = u.Host
	r.RequestURI = u.RequestURI()
	r.RemoteAddr = "127.0.0.1:0"

	return r, nil
}

// bridgeWriter buffers the body and hands it to the host on Flush, so event
// streams arrive as they are written and other responses arrive in one piece.
type bridgeWriter struct {
	mu          sync.Mutex
	host        ResponseWriter
	header      http.Header
	buf         bytes.Buffer
	wroteHeader bool
}

func newBridgeWriter(host ResponseWriter) *bridgeWriter {
	return &bridgeWriter{host: host, header: make(http.Header)}
}

func (w *bridgeWriter) Header() http.Header {
	return w.header
}

func (w *bridgeWriter) WriteHeader(status int) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.writeHeader(status)
}

func (w *bridgeWriter) writeHeader(status int) {
	if w.wroteHeader {
		return
	}

	w.wroteHeader = true

	headers := make(map[string]string, len(w.header))
	for k, v := range w.header {
		headers[k] = strings.Join(v, ", ")
	}

	b, _ := json.Marshal(headers)

	w.host.WriteHeader(status, string(b))
}

func (w *bridgeWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.wroteHeader {
		if w.header.Get("Content-Type") == "" {
			w.header.Set("Content-Type", http.DetectContentType(p))
		}

		w.writeHeader(http.StatusOK)
	}

	return w.buf.Write(p)
}

func (w *bridgeWriter) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.flush()
}

func (w *bridgeWriter) flush() {
	w.writeHeader(http.StatusOK)

	if w.buf.Len() == 0 {
		return
	}

	w.host.Write(bytes.Clone(w.buf.Bytes()))
	w.buf.Reset()
}

// abort drops output not yet handed to the host and answers 500 when the
// status has not been sent.
func (w *bridgeWriter) abort() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.buf.Reset()

	if !w.wroteHeader {
		w.header = make(http.Header)
		w.writeHeader(http.StatusInternalServerError)
	}
}

func (w *bridgeWriter) finish() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.flush()
	w.host.Finish()
}
