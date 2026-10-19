// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package http

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type webhookHandler struct {
	encoder encoder
	listSvc listService
}

func newWebhookHandler(encoder encoder, listSvc listService) *webhookHandler {
	return &webhookHandler{encoder: encoder, listSvc: listSvc}
}

func (h *webhookHandler) Routes(r chi.Router) {
	r.Route("/lists", func(r chi.Router) {
		r.Post("/trigger", h.refreshAll)
		r.Post("/trigger/lists", h.refreshAll)

		r.Get("/trigger", h.refreshAll)
		r.Get("/trigger/lists", h.refreshAll)

		r.Post("/trigger/{listID}", h.refreshByID)
	})
}

func (h *webhookHandler) refreshAll(w http.ResponseWriter, r *http.Request) {
	go h.listSvc.RefreshAll(context.Background())

	h.encoder.NoContent(w)
}

func (h *webhookHandler) refreshByID(w http.ResponseWriter, r *http.Request) {
	listID, err := parseURLParamInt(r, "listID")
	if err != nil {
		h.encoder.BadRequestErr(w, err)
		return
	}

	if err := h.listSvc.RefreshList(context.Background(), int64(listID)); err != nil {
		h.encoder.Error(w, err)
		return
	}

	h.encoder.NoContent(w)
}
