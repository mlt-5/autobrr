// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import SwiftUI

struct FeedsView: View {
    let client: EngineClient

    private static let kind = RecordKind(
        title: "Feeds",
        noun: "feed",
        path: "feeds",
        empty: "No feeds. Adding a Torznab or RSS indexer creates its feed.",
        template: [:],
        canCreate: false,
        canTest: true,
        actions: [
            RecordAction(title: "Run now", suffix: "forcerun", done: "The feed is being refreshed."),
            RecordAction(title: "Clear cache", method: .delete, suffix: "cache", done: "The feed cache was cleared."),
        ],
        summary: { feed in
            let interval = feed["interval"]?.intValue ?? 0
            return "\(feed["type"]?.stringValue ?? "") · every \(interval) min"
        }
    )

    var body: some View {
        RecordListView(kind: Self.kind, client: client) { draft in
            let type = draft.values["type"]?.stringValue ?? ""

            Section {
                TextRow(label: "Name", text: draft.string("name"))
                Toggle("Enabled", isOn: draft.bool("enabled"))
                LabeledContent("Type", value: type)
            }

            Section("Source") {
                TextRow(label: "URL", text: draft.string("url"), prompt: "https://")
                Picker("Download", selection: draft.string(in: "settings", "download_type")) {
                    Text("Torrent").tag("TORRENT")
                    Text("Magnet").tag("MAGNET")
                }

                if type == "TORZNAB" {
                    SecretRow(label: "API key", text: draft.string("api_key"))
                } else {
                    SecretRow(label: "Cookie", text: draft.string("cookie"))
                    TextRow(label: "User agent", text: draft.string("user_agent"))
                }

                Toggle("Skip TLS verification", isOn: draft.bool("tls_skip_verify"))
            }

            Section {
                NumberRow(label: "Interval (minutes)", value: draft.int("interval"))
                NumberRow(label: "Timeout (seconds)", value: draft.int("timeout"))
                NumberRow(label: "Max age (seconds)", value: draft.int("max_age"))
                NumberRow(label: "Cache (days)", value: draft.int(in: "settings", "cache_ttl_days"))
            } header: {
                Text("Refresh")
            } footer: {
                Text("Releases older than the max age are skipped; 0 turns it off. Feeds refresh only while the app is open.")
            }
        }
    }
}

struct ListsView: View {
    let client: EngineClient

    private static let types = ["TRAKT": "Trakt", "ANILIST": "AniList", "PLAINTEXT": "Plain text"]

    private static let kind = RecordKind(
        title: "Lists",
        noun: "list",
        path: "lists",
        empty: "No lists. A list fills the shows, movies or albums of the filters it is linked to.",
        template: [
            "enabled": .bool(true),
            "type": .string("PLAINTEXT"),
            "name": .string(""),
            "url": .string(""),
            "headers": .array([]),
            "api_key": .string(""),
            "filters": .array([]),
            "match_release": .bool(false),
            "skip_clean_sanitize": .bool(false),
        ],
        actions: [RecordAction(title: "Refresh now", suffix: "refresh", done: "The list was refreshed.")],
        summary: { list in
            let filters = list["filters"]?.arrayValue?.count ?? 0
            let status = list["last_refresh_status"]?.stringValue ?? ""
            return [ListsView.types[list["type"]?.stringValue ?? ""] ?? "", "\(filters) filters", status.capitalized]
                .filter { !$0.isEmpty }
                .joined(separator: " · ")
        }
    )

    var body: some View {
        RecordListView(kind: Self.kind, client: client) { draft in
            let type = draft.values["type"]?.stringValue ?? ""

            Section {
                TextRow(label: "Name", text: draft.string("name"))
                Toggle("Enabled", isOn: draft.bool("enabled"))
                if draft.id == nil {
                    Picker("Type", selection: draft.string("type")) {
                        ForEach(["TRAKT", "ANILIST", "PLAINTEXT"], id: \.self) { Text(Self.types[$0] ?? $0).tag($0) }
                    }
                } else {
                    LabeledContent("Type", value: Self.types[type] ?? type)
                }
            }

            Section("Source") {
                TextRow(label: "URL", text: draft.string("url"), prompt: "https://")
                if type == "TRAKT" && !(draft.values["url"]?.stringValue ?? "").hasPrefix("https://api.autobrr.com/") {
                    SecretRow(label: "API key", text: draft.string("api_key"))
                }
                Toggle("Match release names", isOn: draft.bool("match_release"))
                if type == "PLAINTEXT" {
                    Toggle("Keep names as written", isOn: draft.bool("skip_clean_sanitize"))
                }
            }

            Section {
                FilterLinksRow(client: client, selection: Binding {
                    draft.values["filters"]?.arrayValue ?? []
                } set: {
                    draft.set("filters", .array($0))
                })
            } footer: {
                Text("Matching titles from the list replace the shows, movies or albums of these filters on every refresh.")
            }
        }
    }
}

/// FilterLinksRow picks the filters a list fills, stored as {id, name} objects.
private struct FilterLinksRow: View {
    let client: EngineClient
    @Binding var selection: [JSONValue]

    @State private var filters: [FilterSummary] = []

    var body: some View {
        NavigationLink {
            List(filters) { filter in
                Button {
                    toggle(filter)
                } label: {
                    HStack {
                        Text(filter.name).foregroundStyle(Color.primary)
                        Spacer()
                        if selectedIDs.contains(Int64(filter.id)) {
                            Image(systemName: "checkmark").foregroundStyle(Color.accentColor)
                        }
                    }
                }
            }
            .navigationTitle("Filters")
            .navigationBarTitleDisplayMode(.inline)
            .task { filters = (try? await client.get("filters", query: [URLQueryItem(name: "sort", value: "name-asc")])) ?? [] }
        } label: {
            LabeledContent("Filters") {
                Text(names.isEmpty ? "None" : names.joined(separator: ", ")).lineLimit(1)
            }
        }
    }

    private var selectedIDs: Set<Int64> {
        Set(selection.compactMap { $0.objectValue?["id"]?.intValue })
    }

    private var names: [String] {
        selection.compactMap { $0.objectValue?["name"]?.stringValue }
    }

    private func toggle(_ filter: FilterSummary) {
        let id = Int64(filter.id)

        if selectedIDs.contains(id) {
            selection.removeAll { $0.objectValue?["id"]?.intValue == id }
        } else {
            selection.append(.object(["id": .int(id), "name": .string(filter.name)]))
        }
    }
}
