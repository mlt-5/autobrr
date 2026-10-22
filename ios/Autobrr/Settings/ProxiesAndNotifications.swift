// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import SwiftUI

struct ProxiesView: View {
    let client: EngineClient

    private static let kind = RecordKind(
        title: "Proxies",
        noun: "proxy",
        path: "proxy",
        empty: "No proxies. Indexers, IRC networks and feeds can route their traffic through one.",
        template: [
            "enabled": .bool(true),
            "name": .string("Proxy"),
            "type": .string("SOCKS5"),
            "addr": .string("socks5://ip:port"),
            "user": .string(""),
            "pass": .string(""),
        ],
        canTest: true,
        summary: { "\($0["type"]?.stringValue ?? "") · \($0["addr"]?.stringValue ?? "")" }
    )

    var body: some View {
        RecordListView(kind: Self.kind, client: client) { draft in
            Section {
                TextRow(label: "Name", text: draft.string("name"))
                Toggle("Enabled", isOn: draft.bool("enabled"))
                Picker("Type", selection: draft.string("type")) {
                    Text("SOCKS5").tag("SOCKS5")
                    Text("HTTP").tag("HTTP")
                }
            }

            Section {
                TextRow(label: "Address", text: draft.string("addr"), prompt: "socks5://ip:port")
                TextRow(label: "Username", text: draft.string("user"))
                SecretRow(label: "Password", text: draft.string("pass"))
            } footer: {
                Text("The address includes the scheme, such as socks5:// or http://.")
            }
        }
    }
}

struct NotificationsView: View {
    let client: EngineClient

    private static let events = [
        "PUSH_APPROVED": "Push approved",
        "PUSH_REJECTED": "Push rejected",
        "PUSH_ERROR": "Push error",
        "IRC_DISCONNECTED": "IRC disconnected",
        "IRC_RECONNECTED": "IRC reconnected",
        "IRC_UNHEALTHY": "IRC unhealthy",
        "IRC_HEALTHY": "IRC healthy",
        "LIST_REFRESH_SUCCESS": "List refreshed",
        "LIST_REFRESH_ERROR": "List refresh failed",
        "FEED_REFRESH_SUCCESS": "Feed refreshed",
        "FEED_REFRESH_ERROR": "Feed refresh failed",
        "RELEASE_NEW": "New release",
    ]

    private static let eventOrder = [
        "PUSH_APPROVED", "PUSH_REJECTED", "PUSH_ERROR", "IRC_DISCONNECTED", "IRC_RECONNECTED", "IRC_UNHEALTHY",
        "IRC_HEALTHY", "LIST_REFRESH_SUCCESS", "LIST_REFRESH_ERROR", "FEED_REFRESH_SUCCESS", "FEED_REFRESH_ERROR",
        "RELEASE_NEW",
    ]

    private static let kind = RecordKind(
        title: "Notifications",
        noun: "notification",
        path: "notification",
        empty: "No notification services. Add a webhook to be told about pushes, IRC health and more.",
        template: [
            "enabled": .bool(true),
            "type": .string("WEBHOOK"),
            "name": .string(""),
            "webhook": .string(""),
            "method": .string("POST"),
            "events": .array([]),
        ],
        canTest: true,
        canDelete: { $0["type"]?.stringValue != "BUILTIN" },
        summary: { record in
            let count = record["events"]?.arrayValue?.count ?? 0
            return record["type"]?.stringValue == "BUILTIN" ? "In-app inbox · \(count) events" : "Webhook · \(count) events"
        }
    )

    var body: some View {
        RecordListView(kind: Self.kind, client: client) { draft in
            let builtin = draft.values["type"]?.stringValue == "BUILTIN"

            Section {
                TextRow(label: "Name", text: draft.string("name"))
                Toggle("Enabled", isOn: draft.bool("enabled"))
            }

            if !builtin {
                Section("Webhook") {
                    TextRow(label: "URL", text: draft.string("webhook"), prompt: "https://example.com/hook")
                    Picker("Method", selection: draft.string("method")) {
                        ForEach(["GET", "POST", "PUT", "PATCH", "DELETE"], id: \.self) { Text($0).tag($0) }
                    }
                    TextRow(label: "Headers", text: draft.string("headers"), prompt: "Authorization=Bearer token")
                }
            }

            Section("Events") {
                MultiSelectRow(
                    label: "Events",
                    options: builtin ? Self.eventOrder.filter { $0 != "RELEASE_NEW" } : Self.eventOrder,
                    selection: draft.list("events"),
                    titles: Self.events
                )
            }
        }
    }
}

/// SecretRow edits a password. The engine sends stored secrets redacted and keeps the stored value when the redacted text comes back.
struct SecretRow: View {
    let label: String
    @Binding var text: String

    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            Text(label)
                .font(.caption)
                .foregroundStyle(.secondary)

            SecureField("", text: $text)
                .textInputAutocapitalization(.never)
                .autocorrectionDisabled()
        }
    }
}
