// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import Foundation

/// IndexerDefinition is a tracker definition, either a template from indexer/schema or a configured indexer from indexer.
struct IndexerDefinition: Decodable, Identifiable {
    struct Setting: Decodable, Hashable {
        let name: String
        let type: String
        let label: String
        var required: Bool?
        var `default`: String?
        var help: String?
        var value: String?

        var isSecret: Bool { type == "secret" }
    }

    struct IRC: Decodable {
        struct Auth: Decodable {
            var mechanism: String?
        }

        struct Channel: Decodable {
            let name: String
        }

        let network: String
        let server: String
        let port: Int
        let tls: Bool
        var auth: Auth?
        var settings: [Setting]?
        var channels: [Channel]?
    }

    struct Feed: Decodable {
        var settings: [Setting]?
    }

    var id: Int64?
    let name: String
    let identifier: String
    var identifierExternal: String?
    let implementation: String
    var baseUrl: String?
    var enabled: Bool?
    var description: String?
    var urls: [String]?
    var supports: [String]?
    var useProxy: Bool?
    var proxyId: Int64?
    var settings: [Setting]?
    var irc: IRC?
    var feed: Feed?

    var isFeed: Bool { implementation == "torznab" || implementation == "rss" }
    var hasApi: Bool { implementation == "irc" && supports?.contains("api") == true }

    var kind: String {
        switch implementation {
        case "irc": return "IRC"
        case "torznab": return "Torznab"
        case "rss": return "RSS"
        default: return implementation
        }
    }
}

/// IndexerDeprecation explains why an archived indexer's definition was removed.
struct IndexerDeprecation: Decodable {
    let identifier: String
    let name: String
    let reason: String
    var issueUrl: String?
    var filterCount: Int?
}

/// IndexerRow is a configured indexer as indexer/options lists it, archived ones included.
struct IndexerRow: Decodable, Identifiable {
    let id: Int64
    let name: String
    let identifier: String
    var archived: Bool?
}
