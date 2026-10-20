// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package http

import (
	"bufio"
	"bytes"
	"io"
	"io/fs"
	"net/http"
	"os"
	filePath "path"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

type webHandler struct {
	log     zerolog.Logger
	embedFS fs.FS
}

func newWebHandler(log zerolog.Logger, embedFS fs.FS) *webHandler {
	return &webHandler{
		log:     log.With().Str("module", "web-assets").Logger(),
		embedFS: embedFS,
	}
}

// registerAssets walks the FS Dist dir and registers each file as a route
func (h *webHandler) registerAssets(r *chi.Mux) {
	err := fs.WalkDir(h.embedFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		h.log.Trace().Str("path", path).Msg("web asset found")

		if path == "index.html" || path == ".gitkeep" {
			return nil
		}

		// use old path.Join to not be os specific
		FileFS(r, filePath.Join("/", path), path, h.embedFS)

		return nil
	})

	if err != nil {
		return
	}
}

func (h *webHandler) RegisterRoutes(r *chi.Mux) {
	h.registerAssets(r)

	assets, err := fs.Sub(h.embedFS, "assets")
	if err != nil {
		h.log.Error().Err(err).Msg("could not load assets sub dir")
	}

	StaticFS(r, "/assets", assets)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fsFile(w, r, "index.html", h.embedFS)
	})

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		file := r.URL.Path

		if validWebRoute(file) {
			fsFile(w, r, "index.html", h.embedFS)
			return
		}

		fsFile(w, r, strings.TrimPrefix(file, "/"), h.embedFS)
	})
}

type defaultFS struct {
	prefix string
	fs     fs.FS
}

func (fs defaultFS) Open(name string) (fs.File, error) {
	if fs.fs == nil {
		return os.Open(name)
	}
	return fs.fs.Open(name)
}

// FileFS registers a new route with path to serve a file from the provided file system.
func FileFS(r *chi.Mux, path, file string, filesystem fs.FS) {
	r.Get(path, StaticFileHandler(file, filesystem))
}

// StaticFileHandler creates a handler function to serve a file from the provided file system.
func StaticFileHandler(file string, filesystem fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fsFile(w, r, file, filesystem)
	}
}

// StaticFS registers a new route with path prefix to serve static files from the provided file system.
func StaticFS(r *chi.Mux, pathPrefix string, filesystem fs.FS) {
	r.Handle(pathPrefix+"*", http.StripPrefix(pathPrefix, http.FileServer(http.FS(filesystem))))
}

// fsFile is a helper function to serve a file from the provided file system.
func fsFile(w http.ResponseWriter, r *http.Request, file string, filesystem fs.FS) {
	f, err := filesystem.Open(file)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	data, err := io.ReadAll(bufio.NewReader(f))
	if err != nil {
		http.Error(w, "Failed to read the file", http.StatusInternalServerError)
		return
	}

	reader := bytes.NewReader(data)
	http.ServeContent(w, r, file, stat.ModTime(), reader)
}

var validWebRoutes = []string{"filters", "releases", "settings", "logs", "notifications"}

func validWebRoute(route string) bool {
	if route == "" || route == "/" {
		return true
	}
	for _, valid := range validWebRoutes {
		if strings.HasPrefix(route, valid) || strings.HasPrefix(route, "/"+valid) {
			return true
		}
	}

	return false
}
