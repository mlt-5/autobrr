// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

// Package mobile embeds autobrr in a host app through gomobile. The host
// passes each request from its web view to Engine.Serve, so no network
// listener is opened.
package mobile

import (
	"context"
	"sync"

	"github.com/autobrr/autobrr/internal/application"
	"github.com/autobrr/autobrr/pkg/errors"
)

// Engine is a running autobrr instance.
type Engine struct {
	app    *application.App
	cancel context.CancelFunc
	done   chan error

	stopOnce sync.Once
	stopErr  error
}

// Start runs autobrr with its config, database and logs in configDir and
// returns once it is ready to serve requests.
func Start(configDir string) (*Engine, error) {
	if configDir == "" {
		return nil, errors.New("config dir is required")
	}

	ctx, cancel := context.WithCancel(context.Background())

	e := &Engine{
		app:    application.NewEmbedded(configDir),
		cancel: cancel,
		done:   make(chan error, 1),
	}

	go func() {
		e.done <- e.app.Run(ctx)
	}()

	select {
	case <-e.app.Ready():
		return e, nil
	case err := <-e.done:
		cancel()
		if err == nil {
			err = errors.New("autobrr stopped while starting")
		}

		return nil, errors.Wrap(err, "could not start autobrr")
	}
}

// Stop shuts autobrr down and waits for it to finish. Requests still being
// served are cancelled.
func (e *Engine) Stop() error {
	e.stopOnce.Do(func() {
		e.cancel()
		e.stopErr = <-e.done
	})

	return e.stopErr
}
