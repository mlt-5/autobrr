// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import SwiftUI

private let statusTitles = [
    "PUSH_APPROVED": "Approved",
    "PUSH_REJECTED": "Rejected",
    "PUSH_ERROR": "Error",
    "PENDING": "Pending",
]

private let statusOrder = ["PUSH_APPROVED", "PUSH_REJECTED", "PUSH_ERROR", "PENDING"]

struct ReleaseSettingsView: View {
    let client: EngineClient

    @State private var indexers: [String] = []

    var body: some View {
        List {
            Section {
                NavigationLink("Cleanup jobs") {
                    CleanupJobsView(client: client, indexers: indexers)
                }
                NavigationLink("Duplicate profiles") {
                    DuplicateProfilesView(client: client)
                }
            } footer: {
                Text("Cleanup jobs delete old releases on a schedule. Duplicate profiles let a filter skip releases it already pushed.")
            }

            DeleteReleasesSection(client: client, indexers: indexers)
        }
        .navigationTitle("Releases")
        .navigationBarTitleDisplayMode(.inline)
        .task { indexers = (try? await client.get("release/indexers")) ?? [] }
    }
}

private struct CleanupJobsView: View {
    let client: EngineClient
    let indexers: [String]

    private static let kind = RecordKind(
        title: "Cleanup jobs",
        noun: "cleanup job",
        path: "release/cleanup-jobs",
        empty: "No cleanup jobs. Add one to delete old releases on a schedule.",
        template: [
            "name": .string(""),
            "enabled": .bool(true),
            "schedule": .string("0 3 * * *"),
            "older_than": .int(720),
            "indexers": .string(""),
            "statuses": .string(""),
        ],
        actions: [RecordAction(title: "Run now", suffix: "run", done: "The cleanup job ran.")],
        summary: { job in
            let days = (job["older_than"]?.intValue ?? 0) / 24
            return "\(job["schedule"]?.stringValue ?? "") · older than \(days) days"
        }
    )

    var body: some View {
        RecordListView(kind: Self.kind, client: client) { draft in
            Section {
                TextRow(label: "Name", text: draft.string("name"))
                Toggle("Enabled", isOn: draft.bool("enabled"))
            }

            Section {
                TextRow(label: "Schedule", text: draft.string("schedule"), prompt: "0 3 * * *")
                NumberRow(label: "Older than (hours)", value: draft.int("older_than"))
            } footer: {
                Text("The schedule is a cron expression; 0 3 * * * runs daily at 3:00. Runs happen only while the app is open.")
            }

            Section {
                MultiSelectRow(label: "Indexers", options: indexers, selection: draft.commaList("indexers"))
                MultiSelectRow(label: "Statuses", options: statusOrder, selection: draft.commaList("statuses"), titles: statusTitles)
            } footer: {
                Text("Leave empty to match every indexer or status.")
            }
        }
    }
}

private struct DuplicateProfilesView: View {
    let client: EngineClient

    private static let fields: [(String, String)] = [
        ("release_name", "Release name"), ("hash", "Hash"), ("title", "Title"), ("sub_title", "Subtitle"),
        ("year", "Year"), ("month", "Month"), ("day", "Day"), ("source", "Source"), ("resolution", "Resolution"),
        ("codec", "Codec"), ("container", "Container"), ("dynamic_range", "Dynamic range"), ("audio", "Audio"),
        ("group", "Group"), ("season", "Season"), ("episode", "Episode"), ("website", "Website"),
        ("proper", "Proper"), ("repack", "Repack"), ("edition", "Edition"), ("hybrid", "Hybrid"),
        ("language", "Language"), ("protocol", "Protocol"),
    ]

    private static let kind = RecordKind(
        title: "Duplicate profiles",
        noun: "duplicate profile",
        path: "release/profiles/duplicate",
        empty: "No duplicate profiles. A profile names the release fields that must match for a release to count as a duplicate.",
        template: Dictionary(uniqueKeysWithValues: [("name", JSONValue.string(""))] + DuplicateProfilesView.fields.map { ($0.0, JSONValue.bool(false)) }),
        updatesByPost: true,
        summary: { profile in
            DuplicateProfilesView.fields.filter { profile[$0.0]?.boolValue == true }.map(\.1).joined(separator: ", ")
        }
    )

    var body: some View {
        RecordListView(kind: Self.kind, client: client) { draft in
            Section {
                TextRow(label: "Name", text: draft.string("name"))
            }

            Section("Fields that must match") {
                ForEach(Self.fields, id: \.0) { field in
                    Toggle(field.1, isOn: draft.bool(field.0))
                }
            }
        }
    }
}

private struct DeleteReleasesSection: View {
    let client: EngineClient
    let indexers: [String]

    private static let ages: [(Int, String)] = [
        (0, "Any age"), (1, "1 hour"), (12, "12 hours"), (24, "1 day"), (168, "1 week"),
        (720, "1 month"), (2160, "3 months"), (4320, "6 months"), (8760, "1 year"),
    ]

    @State private var olderThan = 720
    @State private var selectedIndexers: [String] = []
    @State private var statuses: [String] = []
    @State private var confirming = false
    @State private var message: String?

    var body: some View {
        Section {
            Picker("Older than", selection: $olderThan) {
                ForEach(Self.ages, id: \.0) { Text($0.1).tag($0.0) }
            }
            MultiSelectRow(label: "Indexers", options: indexers, selection: $selectedIndexers)
            MultiSelectRow(label: "Statuses", options: statusOrder, selection: $statuses, titles: statusTitles)

            Button("Delete releases", role: .destructive) { confirming = true }
        } header: {
            Text("Delete releases")
        } footer: {
            Text("Leave indexers and statuses empty to match all of them.")
        }
        .confirmationDialog("Delete the matching releases?", isPresented: $confirming, titleVisibility: .visible) {
            Button("Delete", role: .destructive) { delete() }
        } message: {
            Text("This can't be undone.")
        }
        .alert("Releases", isPresented: Binding(get: { message != nil }, set: { if !$0 { message = nil } })) {
            Button("OK") { message = nil }
        } message: {
            Text(message ?? "")
        }
    }

    private func delete() {
        var query = [URLQueryItem(name: "olderThan", value: String(olderThan))]
        query += selectedIndexers.map { URLQueryItem(name: "indexer", value: $0) }
        query += statuses.map { URLQueryItem(name: "releaseStatus", value: $0) }

        Task {
            do {
                try await client.request(.delete, "release", query: query)
                message = "The matching releases were deleted."
            } catch {
                message = error.localizedDescription
            }
        }
    }
}
