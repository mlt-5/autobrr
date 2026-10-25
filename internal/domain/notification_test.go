// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package domain

import (
	"testing"
	"time"

	"github.com/autobrr/rls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWebhookEvent(t *testing.T) {
	now := time.Now()
	payload := NotificationPayload{
		Event:          NotificationEventReleaseNew,
		Timestamp:      now,
		ReleaseName:    "Test.Release-Group",
		Indexer:        "MockIndexer",
		Protocol:       ReleaseProtocolTorrent,
		Implementation: ReleaseImplementationIRC,
		Filter:         "TestFilter",
		FilterID:       1,
	}

	release := &Release{
		Type:            rls.Movie,
		TorrentName:     "Test.Release-Group",
		Title:           "Test Release",
		Resolution:      "1080p",
		Source:          "WEB-DL",
		Codec:           []string{"H.264"},
		Size:            1234567,
		Seeders:         10,
		Leechers:        5,
		Freeleech:       true,
		MediaProcessing: "Encode",
		Indexer:         IndexerMinimal{Identifier: "mock_indexer"},
	}

	// set release on payload
	payload.Release = release

	id := "test-uuid-123"
	result := NewWebhookEvent(payload.Event, payload, id)

	assert.Equal(t, WebhookEventReleaseNew, result.Event)
	assert.Equal(t, id, result.ID)
	assert.Equal(t, now, result.Timestamp)
	assert.Equal(t, "1.0", result.Version)

	// Verify Data
	require.NotNil(t, result.Data)

	// Release Data
	require.NotNil(t, result.Data.Release)
	assert.Equal(t, "Test.Release-Group", result.Data.Release.Name)
	assert.Equal(t, "Test Release", result.Data.Release.Title)
	assert.Equal(t, "1080p", result.Data.Release.Resolution)
	assert.Equal(t, uint64(1234567), result.Data.Release.Size)

	// Indexer Data
	require.NotNil(t, result.Data.Indexer)
	assert.Equal(t, "MockIndexer", result.Data.Indexer.Name)
	assert.Equal(t, "mock_indexer", result.Data.Indexer.Identifier)

	// Filter Data
	require.NotNil(t, result.Data.Filter)
	assert.Equal(t, "TestFilter", result.Data.Filter.Name)
	assert.Equal(t, 1, result.Data.Filter.ID)

	// Action Data should be nil for this event type
	assert.Nil(t, result.Data.Action)
}

func TestNewWebhookEvent_Action(t *testing.T) {
	now := time.Now()
	payload := NotificationPayload{
		Event:        NotificationEventPushApproved,
		Timestamp:    now,
		Action:       "TestAction",
		ActionType:   ActionTypeWebhook,
		ActionClient: "qBittorrent",
		Status:       ReleasePushStatusApproved,
	}

	id := "test-uuid-456"
	result := NewWebhookEvent(payload.Event, payload, id)

	assert.Equal(t, WebhookEventActionApproved, result.Event)
	require.NotNil(t, result.Data.Action)
	assert.Equal(t, "TestAction", result.Data.Action.Name)
	assert.Equal(t, "WEBHOOK", result.Data.Action.Type)

	require.NotNil(t, result.Data.Result)
	assert.Equal(t, "PUSH_APPROVED", result.Data.Result.Status)
}

func TestNewWebhookEvent_NilRelease(t *testing.T) {
	now := time.Now()
	payload := NotificationPayload{
		Event:     NotificationEventTest,
		Timestamp: now,
	}

	id := "test-uuid-789"
	result := NewWebhookEvent(payload.Event, payload, id)

	assert.Equal(t, WebhookEventTest, result.Event)
	assert.Equal(t, id, result.ID)
	assert.Nil(t, result.Data.Release)
}

func TestNotification_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		notification Notification
		wantErr      string
	}{
		{name: "webhook missing url", notification: Notification{Type: NotificationTypeWebhook}, wantErr: "missing webhook url"},
		{name: "webhook valid", notification: Notification{Type: NotificationTypeWebhook, Webhook: "https://example.com/hook"}},
		{name: "removed type", notification: Notification{Type: "DISCORD"}, wantErr: "unsupported notification type"},
		{name: "builtin", notification: Notification{Type: NotificationTypeBuiltin}},
		{name: "missing type", notification: Notification{}, wantErr: "unsupported notification type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.notification.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}

			assert.ErrorIs(t, err, ErrNotificationInvalid)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestNewWebhookEvent_ListRefresh(t *testing.T) {
	tests := []struct {
		name  string
		event NotificationEvent
		want  WebhookEventType
		err   string
	}{
		{name: "success", event: NotificationEventListRefreshSuccess, want: WebhookEventListRefreshSuccess},
		{name: "error", event: NotificationEventListRefreshError, want: WebhookEventListRefreshError, err: "unexpected status code: 401"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := NotificationPayload{
				Event:     tt.event,
				List:      "Trakt Popular",
				ListID:    3,
				ListType:  ListTypeTrakt,
				ListError: tt.err,
			}

			result := NewWebhookEvent(payload.Event, payload, "test-uuid")

			assert.Equal(t, tt.want, result.Event)
			require.NotNil(t, result.Data.List)
			assert.Equal(t, &WebhookList{ID: 3, Name: "Trakt Popular", Type: "TRAKT", Error: tt.err}, result.Data.List)
		})
	}
}

func TestNewWebhookEvent_FeedRefresh(t *testing.T) {
	tests := []struct {
		name  string
		event NotificationEvent
		want  WebhookEventType
		err   string
	}{
		{name: "success", event: NotificationEventFeedRefreshSuccess, want: WebhookEventFeedRefreshSuccess},
		{name: "error", event: NotificationEventFeedRefreshError, want: WebhookEventFeedRefreshError, err: "error fetching feed items: unexpected status code: 503"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := NotificationPayload{
				Event:     tt.event,
				Feed:      "Mock Indexer",
				FeedID:    4,
				FeedType:  string(FeedTypeTorznab),
				FeedError: tt.err,
				Indexer:   "Mock Indexer",
			}

			result := NewWebhookEvent(payload.Event, payload, "test-uuid")

			assert.Equal(t, tt.want, result.Event)
			require.NotNil(t, result.Data.Feed)
			assert.Equal(t, &WebhookFeed{ID: 4, Name: "Mock Indexer", Type: "TORZNAB", Error: tt.err}, result.Data.Feed)
			require.NotNil(t, result.Data.Indexer)
			assert.Equal(t, "Mock Indexer", result.Data.Indexer.Name)
		})
	}
}

func TestNewWebhookEvent_IRC(t *testing.T) {
	tests := []struct {
		name    string
		event   NotificationEvent
		want    WebhookEventType
		message string
	}{
		{name: "disconnected", event: NotificationEventIRCDisconnected, want: WebhookEventIRCDisconnected},
		{name: "reconnected", event: NotificationEventIRCReconnected, want: WebhookEventIRCReconnected},
		{name: "stopped", event: NotificationEventIRCDisconnected, want: WebhookEventIRCDisconnected, message: "Network: P2P-Network stopped after repeated short-lived connections"},
		{name: "unhealthy", event: NotificationEventIRCUnhealthy, want: WebhookEventIRCUnhealthy, message: "P2P-Network: #announce: InviteFailed"},
		{name: "healthy", event: NotificationEventIRCHealthy, want: WebhookEventIRCHealthy},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := NotificationPayload{
				Event:      tt.event,
				IRCNetwork: "P2P-Network",
				IRCMessage: tt.message,
			}

			result := NewWebhookEvent(payload.Event, payload, "test-uuid")

			assert.Equal(t, tt.want, result.Event)
			require.NotNil(t, result.Data.IRC)
			assert.Equal(t, &WebhookIRC{Network: "P2P-Network", Message: tt.message}, result.Data.IRC)
		})
	}
}
