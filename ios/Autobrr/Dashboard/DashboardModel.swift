// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import Foundation

struct ReleaseStats: Decodable {
    let filteredCount: Int64
    let pushApprovedCount: Int64
    let pushRejectedCount: Int64
    let pushErrorCount: Int64
}

struct ActivityDay: Decodable, Identifiable {
    let date: String
    let matchedCount: Int64
    let pushApprovedCount: Int64
    let pushRejectedCount: Int64
    let pushErrorCount: Int64

    var id: String { date }
    var day: Date { DayFormat.parse(date) }
}

struct VolumeDay: Decodable, Identifiable {
    let date: String
    let downloadedBytes: Int64

    var id: String { date }
    var day: Date { DayFormat.parse(date) }
}

struct TopEntry: Decodable, Identifiable {
    let indexer: String?
    let filter: String?
    let matchedCount: Int64
    let pushApprovedCount: Int64

    var name: String { indexer ?? filter ?? "" }
    var id: String { name }
}

/// HeatCell is one weekday and hour of the heatmap, moved from UTC to the device's time zone.
struct HeatCell: Identifiable {
    let weekday: Int
    let hour: Int
    let count: Int64

    var id: Int { weekday * 24 + hour }
}

enum DayFormat {
    private static let formatter: DateFormatter = {
        let formatter = DateFormatter()
        formatter.locale = Locale(identifier: "en_US_POSIX")
        formatter.dateFormat = "yyyy-MM-dd"
        return formatter
    }()

    static func parse(_ value: String) -> Date {
        formatter.date(from: value) ?? .distantPast
    }
}

/// EngineAlert is a problem the engine reports until it clears, such as an unhealthy IRC network.
struct EngineAlert: Decodable, Identifiable {
    let kind: String
    let severity: String
    var subjectId: Int64?
    let subject: String
    var message: String?

    var id: String { "\(kind)-\(subjectId ?? 0)-\(subject)" }
}

@MainActor
final class DashboardModel: ObservableObject {
    enum Period: Int, CaseIterable, Identifiable {
        case week = 7
        case month = 30
        case quarter = 90
        case year = 365

        var id: Int { rawValue }

        var label: String {
            switch self {
            case .week: "7 days"
            case .month: "30 days"
            case .quarter: "90 days"
            case .year: "1 year"
            }
        }
    }

    private struct Daily<Day: Decodable>: Decodable {
        let daily: [Day]
    }

    private struct Top: Decodable {
        let top: [TopEntry]
    }

    private struct Heatmap: Decodable {
        let heatmap: [Int64]
    }

    private let client: EngineClient

    @Published var period = Period.month
    @Published private(set) var alerts: [EngineAlert] = []
    @Published private(set) var stats: ReleaseStats?
    @Published private(set) var activity: [ActivityDay] = []
    @Published private(set) var volume: [VolumeDay] = []
    @Published private(set) var heatmap: [HeatCell] = []
    @Published private(set) var topIndexers: [TopEntry] = []
    @Published private(set) var topFilters: [TopEntry] = []
    @Published private(set) var error: String?
    @Published private(set) var hasLoaded = false

    init(client: EngineClient) {
        self.client = client
    }

    func load() async {
        let days = [URLQueryItem(name: "days", value: String(period.rawValue))]

        if let alerts: [EngineAlert] = try? await client.get("alerts") {
            self.alerts = alerts
        }

        do {
            async let stats: ReleaseStats = client.get("release/stats")
            async let activity: Daily<ActivityDay> = client.get("release/stats/activity", query: days)
            async let volume: Daily<VolumeDay> = client.get("release/stats/volume", query: days)
            async let heatmap: Heatmap = client.get("release/stats/heatmap", query: days)
            async let indexers: Top = client.get("release/stats/indexers", query: days)
            async let filters: Top = client.get("release/stats/filters", query: days)

            self.stats = try await stats
            self.activity = try await activity.daily
            self.volume = try await volume.daily
            let counts = try await heatmap.heatmap
            self.heatmap = Self.localCells(counts)
            self.topIndexers = try await indexers.top
            self.topFilters = try await filters.top
            self.error = nil
        } catch is CancellationError {
            return
        } catch {
            self.error = error.localizedDescription
        }

        hasLoaded = true
    }

    /// localCells shifts the engine's UTC weekday and hour buckets (Sunday first) to the device's time zone.
    private static func localCells(_ counts: [Int64]) -> [HeatCell] {
        guard counts.count == 7 * 24 else { return [] }

        let offset = Int((Double(TimeZone.current.secondsFromGMT()) / 3600).rounded())

        return counts.enumerated().map { index, count in
            let shifted = (index + offset + 7 * 24) % (7 * 24)
            return HeatCell(weekday: shifted / 24, hour: shifted % 24, count: count)
        }
    }
}
