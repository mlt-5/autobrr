// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import Foundation

struct FilterSummary: Decodable, Identifiable, Hashable {
    struct Indexer: Decodable, Hashable {
        let name: String
    }

    let id: Int
    let name: String
    var enabled: Bool
    let priority: Int
    let actionsCount: Int
    let actionsEnabledCount: Int
    let indexers: [Indexer]?
}

@MainActor
final class FiltersModel: ObservableObject {
    enum Sort: String, CaseIterable, Identifiable {
        case name = "name-asc"
        case priority = "priority-desc"
        case newest = "created_at-desc"
        case updated = "updated_at-desc"

        var id: String { rawValue }

        var label: String {
            switch self {
            case .name: "Name"
            case .priority: "Priority"
            case .newest: "Newest"
            case .updated: "Recently updated"
            }
        }
    }

    private struct NewFilter: Encodable {
        let name: String
        let enabled = false
        let resolutions: [String] = []
        let codecs: [String] = []
        let sources: [String] = []
        let containers: [String] = []
        let origins: [String] = []
    }

    private struct Created: Decodable {
        let id: Int
    }

    let client: EngineClient

    @Published var sort = Sort.name
    @Published var search = ""
    @Published private(set) var filters: [FilterSummary] = []
    @Published private(set) var hasLoaded = false
    @Published var error: String?

    init(client: EngineClient) {
        self.client = client
    }

    var visibleFilters: [FilterSummary] {
        let search = search.trimmingCharacters(in: .whitespaces)
        guard !search.isEmpty else { return filters }

        return filters.filter { $0.name.localizedCaseInsensitiveContains(search) }
    }

    func load() async {
        do {
            filters = try await client.get("filters", query: [URLQueryItem(name: "sort", value: sort.rawValue)])
            error = nil
        } catch is CancellationError {
            return
        } catch {
            self.error = error.localizedDescription
        }

        hasLoaded = true
    }

    func setEnabled(_ filter: FilterSummary, _ enabled: Bool) async {
        guard let index = filters.firstIndex(where: { $0.id == filter.id }) else { return }

        filters[index].enabled = enabled

        struct Body: Encodable {
            let enabled: Bool
        }

        do {
            try await client.send(.put, "filters/\(filter.id)/enabled", body: Body(enabled: enabled))
        } catch {
            if let index = filters.firstIndex(where: { $0.id == filter.id }) {
                filters[index].enabled = !enabled
            }
            self.error = error.localizedDescription
        }
    }

    func create(name: String) async -> Int? {
        do {
            let created: Created = try await client.send(.post, "filters", body: NewFilter(name: name))
            await load()
            return created.id
        } catch {
            self.error = error.localizedDescription
            return nil
        }
    }

    func duplicate(_ filter: FilterSummary) async {
        do {
            try await client.request(.get, "filters/\(filter.id)/duplicate")
            await load()
        } catch {
            self.error = error.localizedDescription
        }
    }

    func delete(_ filter: FilterSummary) async {
        do {
            try await client.request(.delete, "filters/\(filter.id)")
            filters.removeAll { $0.id == filter.id }
        } catch {
            self.error = error.localizedDescription
        }
    }
}
