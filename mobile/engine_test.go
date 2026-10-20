// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package mobile

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recorder struct {
	status   chan int
	headers  chan map[string]string
	data     chan []byte
	finished chan struct{}
}

func newRecorder() *recorder {
	return &recorder{
		status:   make(chan int, 1),
		headers:  make(chan map[string]string, 1),
		data:     make(chan []byte, 64),
		finished: make(chan struct{}),
	}
}

func (r *recorder) WriteHeader(status int, headersJSON string) {
	var h map[string]string
	_ = json.Unmarshal([]byte(headersJSON), &h)
	r.status <- status
	r.headers <- h
}

func (r *recorder) Write(data []byte) {
	r.data <- data
}

func (r *recorder) Finish() {
	close(r.finished)
}

type response struct {
	status  int
	headers map[string]string
	body    string
}

func (r *recorder) wait(t *testing.T) response {
	t.Helper()

	select {
	case <-r.finished:
	case <-time.After(10 * time.Second):
		t.Fatal("request did not finish")
	}

	res := response{status: <-r.status, headers: <-r.headers}
	for {
		select {
		case b := <-r.data:
			res.body += string(b)
		default:
			return res
		}
	}
}

func startEngine(t *testing.T) *Engine {
	t.Helper()

	e, err := Start(t.TempDir())
	require.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, e.Stop())
	})

	return e
}

func serve(t *testing.T, e *Engine, method, url string, headers map[string]string, body string) response {
	t.Helper()

	h, err := json.Marshal(headers)
	require.NoError(t, err)

	rec := newRecorder()
	e.Serve(method, url, string(h), []byte(body), rec)

	return rec.wait(t)
}

func TestStartRequiresConfigDir(t *testing.T) {
	_, err := Start("")
	assert.Error(t, err)
}

func TestServe(t *testing.T) {
	e := startEngine(t)

	res := serve(t, e, http.MethodGet, "app://autobrr/api/healthz/liveness", nil, "")
	assert.Equal(t, http.StatusOK, res.status)
	assert.Equal(t, "OK", res.body)

	res = serve(t, e, http.MethodGet, "app://autobrr/api/config", nil, "")
	assert.Equal(t, http.StatusOK, res.status)
	assert.Equal(t, "application/json", res.headers["Content-Type"])
	assert.Contains(t, res.body, `"config_dir"`)
}

func TestServeStreamsEvents(t *testing.T) {
	e := startEngine(t)

	rec := newRecorder()
	req := e.Serve(http.MethodGet, "app://autobrr/api/events?stream=logs", "", nil, rec)

	select {
	case status := <-rec.status:
		assert.Equal(t, http.StatusOK, status)
	case <-time.After(10 * time.Second):
		t.Fatal("no response header")
	}

	assert.Equal(t, "text/event-stream", (<-rec.headers)["Content-Type"])

	// the logs stream replays startup logs, so data arrives before Finish
	select {
	case b := <-rec.data:
		assert.Contains(t, string(b), "data:")
	case <-time.After(10 * time.Second):
		t.Fatal("no event received")
	}

	select {
	case <-rec.finished:
		t.Fatal("stream finished before it was cancelled")
	default:
	}

	req.Cancel()

	select {
	case <-rec.finished:
	case <-time.After(10 * time.Second):
		t.Fatal("stream did not finish after cancel")
	}
}

func TestStartOneEngineAtATime(t *testing.T) {
	e, err := Start(t.TempDir())
	require.NoError(t, err)

	_, err = Start(t.TempDir())
	assert.Error(t, err)

	require.NoError(t, e.Stop())

	startEngine(t)
}

func TestStartInvalidConfig(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.toml"), []byte(`logMaxSize = "large"`), 0o644))

	_, err := Start(dir)
	assert.Error(t, err)

	startEngine(t)
}

func TestBridgeWriterAbort(t *testing.T) {
	rec := newRecorder()
	rw := newBridgeWriter(rec)
	rw.Header().Set("Content-Type", "application/json")

	rw.abort()
	rw.finish()

	res := rec.wait(t)
	assert.Equal(t, http.StatusInternalServerError, res.status)
	assert.Empty(t, res.headers)

	rec = newRecorder()
	rw = newBridgeWriter(rec)
	_, _ = rw.Write([]byte(`{"partial":`))

	rw.abort()
	rw.finish()

	res = rec.wait(t)
	assert.Equal(t, http.StatusOK, res.status, "a status already sent stays")
	assert.Empty(t, res.body, "unflushed output is dropped")
}
