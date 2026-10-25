// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/autobrr/autobrr/internal/domain"
	"github.com/autobrr/autobrr/pkg/errors"

	"github.com/go-chi/chi/v5"
)

type vocabService interface {
	List(ctx context.Context) ([]domain.VocabEntry, error)
	Store(ctx context.Context, entry *domain.VocabEntry) error
	Delete(ctx context.Context, id int64) error
	Remove(ctx context.Context, key domain.VocabKey) error
	Restore(ctx context.Context, key domain.VocabKey) error
	Export(ctx context.Context, w io.Writer, kind, search string) error
}

type vocabHandler struct {
	encoder encoder
	service vocabService
}

func newVocabHandler(encoder encoder, service vocabService) *vocabHandler {
	return &vocabHandler{
		encoder: encoder,
		service: service,
	}
}

func (h vocabHandler) Routes(r chi.Router) {
	r.Get("/", h.list)
	r.Post("/", h.store)
	r.Get("/export", h.export)
	r.Delete("/{vocabID}", h.delete)

	r.Post("/removed", h.remove)
	r.Delete("/removed", h.restore)
}

type vocabResponse struct {
	Entries      []domain.VocabEntry `json:"entries"`
	Kinds        []string            `json:"kinds"`
	KindTitles   map[string]string   `json:"kind_titles"`
	ReleaseTypes []string            `json:"release_types"`
}

func (h vocabHandler) list(w http.ResponseWriter, r *http.Request) {
	entries, err := h.service.List(r.Context())
	if err != nil {
		h.encoder.Error(w, err)
		return
	}

	h.encoder.StatusResponse(w, http.StatusOK, vocabResponse{
		Entries:      entries,
		Kinds:        domain.VocabKinds,
		KindTitles:   domain.VocabKindTitles,
		ReleaseTypes: domain.VocabReleaseTypes,
	})
}

func (h vocabHandler) export(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	if err := h.service.Export(r.Context(), &buf, r.URL.Query().Get("kind"), r.URL.Query().Get("q")); err != nil {
		h.encoder.Error(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="autobrr-vocabulary-%s.xlsx"`, time.Now().Format("2006-01-02")))
	w.WriteHeader(http.StatusOK)
	_, _ = buf.WriteTo(w)
}

func (h vocabHandler) store(w http.ResponseWriter, r *http.Request) {
	var entry domain.VocabEntry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		h.encoder.BadRequestErr(w, err)
		return
	}

	if err := h.service.Store(r.Context(), &entry); err != nil {
		h.error(w, err)
		return
	}

	h.encoder.StatusCreatedData(w, entry)
}

func (h vocabHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseURLParamInt(r, "vocabID")
	if err != nil {
		h.encoder.BadRequestErr(w, err)
		return
	}

	if err := h.service.Delete(r.Context(), int64(id)); err != nil {
		h.error(w, err)
		return
	}

	h.encoder.NoContent(w)
}

func (h vocabHandler) remove(w http.ResponseWriter, r *http.Request) {
	var key domain.VocabKey
	if err := json.NewDecoder(r.Body).Decode(&key); err != nil {
		h.encoder.BadRequestErr(w, err)
		return
	}

	if err := h.service.Remove(r.Context(), key); err != nil {
		h.error(w, err)
		return
	}

	h.encoder.NoContent(w)
}

func (h vocabHandler) restore(w http.ResponseWriter, r *http.Request) {
	key := domain.VocabKey{Kind: r.URL.Query().Get("kind"), Tag: r.URL.Query().Get("tag")}

	if err := h.service.Restore(r.Context(), key); err != nil {
		h.error(w, err)
		return
	}

	h.encoder.NoContent(w)
}

func (h vocabHandler) error(w http.ResponseWriter, err error) {
	if _, ok := errors.AsType[*domain.VocabError](err); ok {
		h.encoder.BadRequestErr(w, err)
		return
	}

	if errors.Is(err, domain.ErrRecordNotFound) {
		h.encoder.NotFoundErr(w, errors.New("could not find vocabulary entry"))
		return
	}

	h.encoder.Error(w, err)
}
