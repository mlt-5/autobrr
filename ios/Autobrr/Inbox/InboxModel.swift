// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import Foundation

/// InboxMessage is a notification the built-in notification service stored for the app.
struct InboxMessage: Decodable, Identifiable {
    let id: Int64
    let event: String
    let title: String
    let message: String
    var releaseName: String?
    var indexer: String?
    var filterName: String?
    var filterId: Int?
    var action: String?
    var actionClient: String?
    var rejections: [String]?
    var url: String?
    var readAt: Date?
    let createdAt: Date

    var isUnread: Bool { readAt == nil }
}

@MainActor
final class InboxModel: ObservableObject {
    private struct Page: Decodable {
        let data: [InboxMessage]?
        let count: Int
        let allCount: Int
        let unreadCount: Int
    }

    private static let pageSize = 50

    private let client: EngineClient

    @Published private(set) var messages: [InboxMessage] = []
    @Published private(set) var total = 0
    @Published private(set) var allCount = 0
    @Published private(set) var unreadCount = 0
    @Published private(set) var hasLoaded = false
    @Published var error: String?
    @Published var unreadOnly = false {
        didSet {
            messages = []
            Task { await load() }
        }
    }

    init(client: EngineClient) {
        self.client = client
    }

    /// load refreshes every page already shown, so an update does not drop what the list scrolled past.
    func load() async {
        do {
            let page = try await fetch(offset: 0, limit: max(Self.pageSize, messages.count))
            messages = page.data ?? []
            apply(page)
        } catch is CancellationError {
            return
        } catch {
            self.error = error.localizedDescription
        }

        hasLoaded = true
    }

    func loadMore() async {
        guard messages.count < total else { return }

        do {
            let page = try await fetch(offset: messages.count, limit: Self.pageSize)
            let known = Set(messages.map(\.id))
            messages.append(contentsOf: (page.data ?? []).filter { !known.contains($0.id) })
            apply(page)
        } catch {
            self.error = error.localizedDescription
        }
    }

    /// markRead marks the given messages read, or every message when ids is empty.
    func markRead(_ ids: [Int64]) async {
        await run { try await self.client.send(.post, "notification/inbox/read", body: ["ids": ids]) }
    }

    func delete(_ ids: [Int64]) async {
        await run { try await self.client.send(.post, "notification/inbox/delete", body: ["ids": ids]) }
    }

    func deleteAll() async {
        await run { _ = try await self.client.request(.delete, "notification/inbox") }
    }

    /// watch reloads whenever the engine reports an inbox change, reconnecting with a growing delay, until cancelled.
    func watch() async {
        var attempt = 0

        while !Task.isCancelled {
            do {
                var parser = EventStreamParser()

                for try await chunk in client.stream("events", query: [URLQueryItem(name: "stream", value: "notifications")]) {
                    attempt = 0

                    if !parser.append(chunk).isEmpty {
                        await load()
                    }
                }
            } catch {}

            guard !Task.isCancelled else { return }

            try? await Task.sleep(nanoseconds: UInt64(min(30, 1 << min(attempt, 5))) * 1_000_000_000)
            attempt += 1
        }
    }

    private func fetch(offset: Int, limit: Int) async throws -> Page {
        var query = [
            URLQueryItem(name: "limit", value: String(limit)),
            URLQueryItem(name: "offset", value: String(offset)),
        ]
        if unreadOnly {
            query.append(URLQueryItem(name: "unread", value: "true"))
        }

        return try await client.get("notification/inbox", query: query)
    }

    private func apply(_ page: Page) {
        total = page.count
        allCount = page.allCount
        unreadCount = page.unreadCount
        error = nil
    }

    private func run(_ operation: @escaping () async throws -> Void) async {
        do {
            try await operation()
        } catch {
            self.error = error.localizedDescription
        }

        await load()
    }
}
