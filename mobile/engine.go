// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

// Package mobile embeds autobrr in a host app through gomobile. The host
// calls the API through Engine.Serve, so no network listener is opened.
package mobile

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/autobrr/autobrr/internal/application"
	"github.com/autobrr/autobrr/pkg/errors"
)

// running keeps a second engine off the same config, database and the
// process-wide viper and logger state.
var running atomic.Bool

// Engine is a running autobrr instance.
type Engine struct {
	app    *application.App
	cancel context.CancelFunc
	done   chan error

	stopOnce sync.Once
	stopErr  error
}

// Start runs autobrr with its config, database and logs in configDir and
// returns once it is ready to serve requests. Only one engine runs at a time.
func Start(configDir string) (e *Engine, err error) {
	if configDir == "" {
		return nil, errors.New("config dir is required")
	}

	if !running.CompareAndSwap(false, true) {
		return nil, errors.New("autobrr is already running")
	}

	defer func() {
		// an unrecovered panic in a gomobile library aborts the host app
		errors.RecoverPanic(recover(), &err)

		if err != nil {
			running.Store(false)
		}
	}()

	e, err = start(configDir)
	if err != nil {
		return nil, errors.Wrap(err, "could not start autobrr")
	}

	return e, nil
}

func start(configDir string) (*Engine, error) {
	app, err := application.New(configDir)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	e := &Engine{
		app:    app,
		cancel: cancel,
		done:   make(chan error, 1),
	}

	go func() {
		var err error

		defer func() {
			errors.RecoverPanic(recover(), &err)
			e.done <- err
		}()

		err = e.app.Run(ctx)
	}()

	select {
	case <-e.app.Ready():
		return e, nil
	case err := <-e.done:
		cancel()
		if err == nil {
			err = errors.New("autobrr stopped while starting")
		}

		return nil, err
	}
}

// Stop shuts autobrr down and waits for it to finish. Requests still being
// served are cancelled.
func (e *Engine) Stop() error {
	e.stopOnce.Do(func() {
		e.cancel()
		e.stopErr = <-e.done
		running.Store(false)
	})

	return e.stopErr
}
