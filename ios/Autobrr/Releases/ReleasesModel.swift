// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import Foundation

@MainActor
final class ReleasesModel: ObservableObject {
    struct Query: Equatable {
        var search = ""
        var status: PushStatus?
        var indexer: String?
        var favorites = false
    }

    private static let pageSize = 50

    let client: EngineClient

    @Published var query = Query()
    @Published private(set) var releases: [Release] = []
    @Published private(set) var total = 0
    @Published private(set) var indexers: [String] = []
    @Published private(set) var error: String?
    @Published private(set) var hasLoaded = false

    private var cursor: Int64?
    private var generation = 0

    init(client: EngineClient) {
        self.client = client
    }

    var hasMore: Bool { cursor != nil }

    func reload() async {
        generation += 1
        let current = generation

        do {
            let page = try await fetch(cursor: 0)
            guard current == generation else { return }

            releases = page.data.mergingActions()
            total = page.count
            cursor = page.data.isEmpty ? nil : page.nextCursor
            error = nil
            hasLoaded = true
        } catch is CancellationError {
        } catch {
            guard current == generation else { return }

            self.error = error.localizedDescription
            hasLoaded = true
        }
    }

    func loadMore() async {
        guard let cursor else { return }

        let current = generation

        do {
            let page = try await fetch(cursor: cursor)
            guard current == generation, cursor == self.cursor else { return }

            releases = (releases + page.data).mergingActions()
            self.cursor = page.data.isEmpty ? nil : page.nextCursor
        } catch is CancellationError {
        } catch {
            guard current == generation else { return }

            self.error = error.localizedDescription
            self.cursor = nil
        }
    }

    /// setFavorite marks or unmarks a release as a favourite and updates the list.
    func setFavorite(_ release: Release, _ favorite: Bool) async throws {
        try await client.request(favorite ? .post : .delete, "release/\(release.id)/favorite")
        favoriteChanged(id: release.id, favorite)
    }

    /// favoriteChanged updates the list after a release was marked or unmarked elsewhere.
    func favoriteChanged(id: Int64, _ favorite: Bool) {
        if query.favorites && !favorite {
            releases.removeAll { $0.id == id }
            total = max(total - 1, 0)
        } else if let index = releases.firstIndex(where: { $0.id == id }) {
            releases[index].favorite = favorite
        }
    }

    func loadIndexers() async {
        indexers = (try? await client.get("release/indexers")) ?? []
    }

    /// export writes the releases matching the query, every page of them, to an Excel file and returns its URL.
    func export() async throws -> URL {
        var items = filterItems()
        items.append(URLQueryItem(name: "tz", value: TimeZone.current.identifier))

        let data = try await client.request(.get, "release/export", query: items)

        let formatter = DateFormatter()
        formatter.locale = Locale(identifier: "en_US_POSIX")
        formatter.dateFormat = "yyyy-MM-dd"

        let url = FileManager.default.temporaryDirectory.appendingPathComponent("autobrr-releases-\(formatter.string(from: Date())).xlsx")
        try data.write(to: url, options: .atomic)

        return url
    }

    private func fetch(cursor: Int64) async throws -> ReleasePage {
        var items = [URLQueryItem(name: "limit", value: String(Self.pageSize))]

        if cursor > 0 {
            items.append(URLQueryItem(name: "cursor", value: String(cursor)))
        }

        return try await client.get("release", query: items + filterItems())
    }

    private func filterItems() -> [URLQueryItem] {
        var items: [URLQueryItem] = []

        let search = query.search.trimmingCharacters(in: .whitespaces)
        if !search.isEmpty {
            items.append(URLQueryItem(name: "q", value: search))
        }

        if let status = query.status {
            items.append(URLQueryItem(name: "push_status", value: status.rawValue))
        }

        if let indexer = query.indexer {
            items.append(URLQueryItem(name: "indexer", value: indexer))
        }

        if query.favorites {
            items.append(URLQueryItem(name: "favorite", value: "true"))
        }

        return items
    }
}
