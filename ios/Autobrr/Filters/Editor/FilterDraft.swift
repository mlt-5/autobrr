// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import SwiftUI

/// FilterDraft holds a filter as the engine returns it and saves it back whole,
/// so fields the screens do not show survive an edit.
@MainActor
final class FilterDraft: ObservableObject, JSONDocument {
    struct Option: Decodable, Identifiable, Hashable {
        let id: Int64
        let name: String
    }

    struct NotificationOption: Decodable, Identifiable {
        let id: Int64
        let name: String
        let type: String
        let enabled: Bool
    }

    // the engine omits empty lists, but saving needs them: the filter table keeps them NOT NULL
    private static let listKeys = [
        "announce_types", "resolutions", "sources", "codecs", "containers", "match_hdr", "except_hdr",
        "match_other", "except_other", "match_language", "except_language", "formats", "quality", "media",
        "match_release_types", "origins", "except_origins", "indexers", "actions", "external", "notifications",
    ]

    let id: Int
    let client: EngineClient

    @Published var values: [String: JSONValue] = [:]
    @Published private(set) var original: [String: JSONValue] = [:]
    @Published private(set) var indexerOptions: [JSONValue] = []
    @Published private(set) var duplicateProfiles: [Option] = []
    @Published private(set) var notificationOptions: [NotificationOption] = []
    @Published private(set) var hasLoaded = false
    @Published private(set) var isSaving = false
    @Published var error: String?

    init(id: Int, client: EngineClient) {
        self.id = id
        self.client = client
    }

    var hasChanges: Bool { values != original }

    var name: String { values["name"]?.stringValue ?? "Filter" }

    func load() async {
        do {
            let data = try await client.request(.get, "filters/\(id)")
            var filter = try JSONDecoder().decode([String: JSONValue].self, from: data)

            for key in Self.listKeys where filter[key] == nil || filter[key] == .null {
                filter[key] = .array([])
            }
            if filter["max_downloads_window_type"]?.stringValue?.isEmpty != false {
                filter["max_downloads_window_type"] = .string("FIXED")
            }
            if (filter["max_downloads_period"]?.intValue ?? 0) == 0 {
                filter["max_downloads_period"] = .int(1)
            }

            values = filter
            original = filter
            error = nil
        } catch is CancellationError {
            return
        } catch {
            self.error = error.localizedDescription
        }

        hasLoaded = true

        let indexers = try? await client.request(.get, "indexer/options")
        indexerOptions = indexers.flatMap { try? JSONDecoder().decode([JSONValue].self, from: $0) } ?? []
        duplicateProfiles = (try? await client.get("release/profiles/duplicate")) ?? []
        notificationOptions = (try? await client.get("notification")) ?? []
    }

    func save() async -> Bool {
        isSaving = true
        defer { isSaving = false }

        do {
            try await client.request(.put, "filters/\(id)", body: JSONEncoder().encode(values))
            // new actions and external filters get their ids on save, so take the stored filter back
            await load()
            return true
        } catch {
            self.error = error.localizedDescription
            return false
        }
    }

    func discard() {
        values = original
    }
}
