// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import SwiftUI

/// FilterDraft holds a filter as the engine returns it and saves it back whole, like the web editor,
/// so fields the native screens do not show survive an edit.
@MainActor
final class FilterDraft: ObservableObject {
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

    @Published private(set) var values: [String: JSONValue] = [:]
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

    func set(_ key: String, _ value: JSONValue) {
        values[key] = value
    }

    func string(_ key: String) -> Binding<String> {
        Binding { self.values[key]?.stringValue ?? "" } set: { self.set(key, .string($0)) }
    }

    func bool(_ key: String) -> Binding<Bool> {
        Binding { self.values[key]?.boolValue ?? false } set: { self.set(key, .bool($0)) }
    }

    func int(_ key: String) -> Binding<Int> {
        Binding { Int(self.values[key]?.intValue ?? 0) } set: { self.set(key, .int(Int64($0))) }
    }

    func list(_ key: String) -> Binding<[String]> {
        Binding {
            self.values[key]?.arrayValue?.compactMap(\.stringValue) ?? []
        } set: {
            self.set(key, .array($0.map(JSONValue.string)))
        }
    }

    func items(_ key: String) -> [[String: JSONValue]] {
        values[key]?.arrayValue?.compactMap(\.objectValue) ?? []
    }

    func setItems(_ key: String, _ items: [[String: JSONValue]]) {
        set(key, .array(items.map(JSONValue.object)))
    }

    func string(_ key: String, _ index: Int, _ field: String) -> Binding<String> {
        Binding { self.field(key, index, field)?.stringValue ?? "" } set: { value in
            self.update(key, index) { $0[field] = .string(value) }
        }
    }

    func bool(_ key: String, _ index: Int, _ field: String) -> Binding<Bool> {
        Binding { self.field(key, index, field)?.boolValue ?? false } set: { value in
            self.update(key, index) { $0[field] = .bool(value) }
        }
    }

    func int(_ key: String, _ index: Int, _ field: String) -> Binding<Int> {
        Binding { Int(self.field(key, index, field)?.intValue ?? 0) } set: { value in
            self.update(key, index) { $0[field] = .int(Int64(value)) }
        }
    }

    func list(_ key: String, _ index: Int, _ field: String) -> Binding<[String]> {
        Binding {
            self.field(key, index, field)?.arrayValue?.compactMap(\.stringValue) ?? []
        } set: { value in
            self.update(key, index) { $0[field] = .array(value.map(JSONValue.string)) }
        }
    }

    private func field(_ key: String, _ index: Int, _ field: String) -> JSONValue? {
        let items = items(key)
        return items.indices.contains(index) ? items[index][field] : nil
    }

    private func update(_ key: String, _ index: Int, _ change: (inout [String: JSONValue]) -> Void) {
        var items = items(key)
        guard items.indices.contains(index) else { return }

        change(&items[index])
        setItems(key, items)
    }
}
