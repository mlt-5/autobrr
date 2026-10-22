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

    let id: Int
    let client: EngineClient

    @Published private(set) var values: [String: JSONValue] = [:]
    @Published private(set) var original: [String: JSONValue] = [:]
    @Published private(set) var indexerOptions: [JSONValue] = []
    @Published private(set) var duplicateProfiles: [Option] = []
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
            let filter = try JSONDecoder().decode([String: JSONValue].self, from: data)

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
    }

    func save() async -> Bool {
        isSaving = true
        defer { isSaving = false }

        do {
            try await client.request(.put, "filters/\(id)", body: JSONEncoder().encode(values))
            original = values
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
}
