// Copyright (c) 2021-2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package notification

import (
	"github.com/autobrr/autobrr/internal/domain"
)

// BuildTitle constructs the title of the notification message.
func BuildTitle(event domain.NotificationEvent) string {
	titles := map[domain.NotificationEvent]string{
		domain.NotificationEventAppUpdateAvailable: "Autobrr update available",
		domain.NotificationEventPushApproved:       "Push Approved",
		domain.NotificationEventPushRejected:       "Push Rejected",
		domain.NotificationEventPushError:          "Push Error",
		domain.NotificationEventIRCDisconnected:    "IRC Disconnected",
		domain.NotificationEventIRCReconnected:     "IRC Reconnected",
		domain.NotificationEventIRCUnhealthy:       "IRC Unhealthy",
		domain.NotificationEventIRCHealthy:         "IRC Healthy",
		domain.NotificationEventListRefreshSuccess: "List Refresh Success",
		domain.NotificationEventListRefreshError:   "List Refresh Error",
		domain.NotificationEventFeedRefreshSuccess: "Feed Refresh Success",
		domain.NotificationEventFeedRefreshError:   "Feed Refresh Error",
		domain.NotificationEventReleaseNew:         "New Release",
		domain.NotificationEventTest:               "Test",
	}

	if title, ok := titles[event]; ok {
		return title
	}

	return "New Event"
}
