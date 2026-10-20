// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package http

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/autobrr/autobrr/internal/config"
	"github.com/autobrr/autobrr/internal/domain"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type configJson struct {
	Application   string `json:"application"`
	ConfigDir     string `json:"config_dir"`
	LogLevel      string `json:"log_level"`
	LogPath       string `json:"log_path"`
	LogMaxSize    int    `json:"log_max_size"`
	LogMaxBackups int    `json:"log_max_backups"`
	Version       string `json:"version"`
	Commit        string `json:"commit"`
	Date          string `json:"date"`
}

type buildInfo struct {
	version string
	commit  string
	date    string
}

type configHandler struct {
	encoder encoder

	cfg         *config.AppConfig
	versionInfo buildInfo
}

func newConfigHandler(encoder encoder, versionInfo buildInfo, cfg *config.AppConfig) *configHandler {
	return &configHandler{
		encoder:     encoder,
		cfg:         cfg,
		versionInfo: versionInfo,
	}
}

func (h configHandler) Routes(r chi.Router) {
	r.Get("/", h.getConfig)
	r.Patch("/", h.updateConfig)
}

func (h configHandler) getConfig(w http.ResponseWriter, r *http.Request) {
	conf := configJson{
		ConfigDir:     h.cfg.Config.ConfigPath,
		LogLevel:      h.cfg.Config.LogLevel,
		LogPath:       h.cfg.Config.LogPath,
		LogMaxSize:    h.cfg.Config.LogMaxSize,
		LogMaxBackups: h.cfg.Config.LogMaxBackups,
		Version:       h.versionInfo.version,
		Commit:        h.versionInfo.commit,
		Date:          h.versionInfo.date,
	}

	ex, err := os.Executable()
	if err != nil {
		h.encoder.Error(w, err)
		return
	}

	conf.Application = ex

	absPath, err := filepath.Abs(h.cfg.Config.ConfigPath)
	if err != nil {
		h.encoder.Error(w, err)
		return
	}

	conf.ConfigDir = absPath

	render.JSON(w, r, conf)
}

func (h configHandler) updateConfig(w http.ResponseWriter, r *http.Request) {
	var data domain.ConfigUpdate

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		h.encoder.Error(w, err)
		return
	}

	if data.LogLevel != nil {
		h.cfg.Config.LogLevel = *data.LogLevel
	}

	if data.LogPath != nil {
		h.cfg.Config.LogPath = *data.LogPath
	}

	if err := h.cfg.UpdateConfig(); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, errorResponse{
			Message: err.Error(),
			Status:  http.StatusInternalServerError,
		})
		return
	}

	render.NoContent(w, r)
}
