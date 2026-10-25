// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

//go:build integration

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/autobrr/autobrr/internal/domain"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type vocabServiceMock struct {
	stored   *domain.VocabEntry
	deleted  int64
	removed  domain.VocabKey
	restored domain.VocabKey
}

func (m *vocabServiceMock) List(ctx context.Context) ([]domain.VocabEntry, error) {
	return []domain.VocabEntry{{Source: domain.VocabSourceList, Kind: "group", Tag: "SubsPlease", Aliases: []string{}}}, nil
}

func (m *vocabServiceMock) Store(ctx context.Context, entry *domain.VocabEntry) error {
	if entry.Tag == "" {
		return domain.VocabErrorf("a tag is required")
	}
	entry.ID = 7
	m.stored = entry
	return nil
}

func (m *vocabServiceMock) Delete(ctx context.Context, id int64) error {
	if id != 7 {
		return domain.ErrRecordNotFound
	}
	m.deleted = id
	return nil
}

func (m *vocabServiceMock) Remove(ctx context.Context, key domain.VocabKey) error {
	m.removed = key
	return nil
}

func (m *vocabServiceMock) Restore(ctx context.Context, key domain.VocabKey) error {
	m.restored = key
	return nil
}

func TestVocabHandler(t *testing.T) {
	t.Parallel()

	service := &vocabServiceMock{}
	r := chi.NewRouter()
	r.Route("/api/vocab", newVocabHandler(encoder{}, service).Routes)
	server := httptest.NewServer(r)
	defer server.Close()

	do := func(method, path string, body any) *http.Response {
		t.Helper()
		var buf bytes.Buffer
		if body != nil {
			require.NoError(t, json.NewEncoder(&buf).Encode(body))
		}
		req, err := http.NewRequest(method, server.URL+"/api/vocab"+path, &buf)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		t.Cleanup(func() { _ = resp.Body.Close() })
		return resp
	}

	resp := do(http.MethodGet, "/", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var list vocabResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&list))
	assert.Len(t, list.Entries, 1)
	assert.Contains(t, list.Kinds, "group")
	assert.Contains(t, list.ReleaseTypes, "movie")

	resp = do(http.MethodPost, "/", map[string]any{"kind": "group", "tag": "ZqFansubs", "aliases": []string{"Zq Subs"}})
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var created domain.VocabEntry
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	assert.Equal(t, int64(7), created.ID)
	assert.Equal(t, []string{"Zq Subs"}, service.stored.Aliases)

	assert.Equal(t, http.StatusBadRequest, do(http.MethodPost, "/", map[string]any{"kind": "group"}).StatusCode)

	assert.Equal(t, http.StatusNoContent, do(http.MethodDelete, "/7", nil).StatusCode)
	assert.Equal(t, http.StatusNotFound, do(http.MethodDelete, "/8", nil).StatusCode)

	assert.Equal(t, http.StatusNoContent, do(http.MethodPost, "/removed", domain.VocabKey{Kind: "group", Tag: "SubsPlease"}).StatusCode)
	assert.Equal(t, domain.VocabKey{Kind: "group", Tag: "SubsPlease"}, service.removed)

	assert.Equal(t, http.StatusNoContent, do(http.MethodDelete, "/removed?kind=group&tag=Sub%26Please", nil).StatusCode)
	assert.Equal(t, domain.VocabKey{Kind: "group", Tag: "Sub&Please"}, service.restored)
}
