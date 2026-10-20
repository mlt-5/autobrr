// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package events

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The caller field is stamped with a hardcoded frame skip, so it silently
// starts pointing at the bus itself if the Emit<Event>/On<Event> wrappers gain
// or lose a level of indirection.
func TestEventBus_CallerPointsAtCallSite(t *testing.T) {
	var buf bytes.Buffer

	bus := NewEventBus(zerolog.New(&buf).Level(zerolog.TraceLevel))

	unregister := bus.OnIRC(func(ctx context.Context, event IRCEvent) error {
		return nil
	})
	defer unregister()

	bus.EmitIRC(t.Context(), IRCEvent{Type: IRCDisconnected})

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	require.NotEmpty(t, lines)

	for _, line := range lines {
		var entry struct {
			Caller  string `json:"caller"`
			Message string `json:"message"`
		}
		require.NoError(t, json.Unmarshal([]byte(line), &entry))

		assert.Containsf(t, entry.Caller, "eventbus_test.go", "%q logged caller %q", entry.Message, entry.Caller)
	}
}

// signals' TryEmit bails on a cancelled context, so the bus detaches
// cancellation before dispatch: an aborted HTTP request must not skip the
// subscribers that reconcile what the request already committed.
func TestEventBus_EmitSurvivesCallerCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	bus := NewEventBus(zerolog.Nop())

	var received int
	unregister := bus.OnIRC(func(ctx context.Context, event IRCEvent) error {
		received++
		return nil
	})
	defer unregister()

	bus.EmitIRC(ctx, IRCEvent{Type: IRCDisconnected})

	assert.Equal(t, 1, received)
}
