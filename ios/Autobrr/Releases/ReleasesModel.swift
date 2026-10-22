// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import Foundation

@MainActor
final class ReleasesModel: ObservableObject {
    struct Query: Equatable {
        var search = ""
        var status: PushStatus?
        var indexer: String?
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

    func loadIndexers() async {
        indexers = (try? await client.get("release/indexers")) ?? []
    }

    private func fetch(cursor: Int64) async throws -> ReleasePage {
        var items = [URLQueryItem(name: "limit", value: String(Self.pageSize))]

        if cursor > 0 {
            items.append(URLQueryItem(name: "cursor", value: String(cursor)))
        }

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

        return try await client.get("release", query: items)
    }
}
