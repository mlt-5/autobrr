// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import Foundation

struct Release: Decodable, Identifiable, Hashable {
    struct Indexer: Decodable, Hashable {
        let id: Int
        let name: String
        let identifier: String
    }

    let id: Int64
    let filterStatus: String
    let rejections: [String]?
    let indexer: Indexer
    let filter: String
    let implementation: String
    let timestamp: Date
    let infoUrl: String
    let downloadUrl: String
    let name: String
    let size: UInt64
    var title: String
    let category: String
    var season: Int
    var episode: Int
    var year: Int
    var resolution: String
    var source: String
    var codec: [String]?
    var container: String
    var hdr: [String]?
    var group: String
    /// vocabGroup is true when the vocabulary has the group.
    var vocabGroup: Bool
    /// hdrText is the text in the name the parser read as HDR tags.
    var hdrText: [String]?
    /// resolutionText is the text in the name the parser read as the resolution.
    var resolutionText: [String]?
    /// codecText is the text in the name the parser read as codecs.
    var codecText: [String]?
    /// numberingRanges are where the year, season and episode are in the name, as [start, end) offsets in Unicode
    /// scalars.
    var numberingRanges: [[Int]]?
    let origin: String
    let uploader: String
    var actionStatus: [ActionStatus]?
    var favorite: Bool

    /// takeParsed copies the fields parsed from the name, for a release whose name was parsed again.
    mutating func takeParsed(from other: Release) {
        title = other.title
        season = other.season
        episode = other.episode
        year = other.year
        resolution = other.resolution
        source = other.source
        codec = other.codec
        container = other.container
        hdr = other.hdr
        group = other.group
        vocabGroup = other.vocabGroup
        hdrText = other.hdrText
        resolutionText = other.resolutionText
        codecText = other.codecText
        numberingRanges = other.numberingRanges
    }
}

struct ActionStatus: Decodable, Identifiable, Hashable {
    let id: Int64
    let status: String
    let action: String
    let type: String
    let client: String
    let filter: String
    let rejections: [String]?
    let timestamp: Date

    var pushStatus: PushStatus? { PushStatus(rawValue: status) }
}

enum PushStatus: String, CaseIterable, Identifiable {
    case approved = "PUSH_APPROVED"
    case rejected = "PUSH_REJECTED"
    case error = "PUSH_ERROR"
    case pending = "PENDING"

    var id: String { rawValue }

    var label: String {
        switch self {
        case .approved: "Approved"
        case .rejected: "Rejected"
        case .error: "Error"
        case .pending: "Pending"
        }
    }
}

struct ReleasePage: Decodable {
    let data: [Release]
    let count: Int
    let nextCursor: Int64
}

extension Array where Element == Release {
    /// mergingActions folds the rows the API returns once per action status into one release each, keeping order.
    func mergingActions() -> [Release] {
        var merged: [Release] = []
        var index: [Int64: Int] = [:]

        for release in self {
            if let i = index[release.id] {
                merged[i].actionStatus = (merged[i].actionStatus ?? []) + (release.actionStatus ?? [])
            } else {
                index[release.id] = merged.count
                merged.append(release)
            }
        }

        return merged
    }
}
