// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import SwiftUI

struct ReleaseDetailView: View {
    let client: EngineClient

    @State private var release: Release

    @State private var retrying: Int64?
    @State private var message: String?
    @State private var addingVocab = false

    init(release: Release, client: EngineClient) {
        self.client = client
        _release = State(initialValue: release)
    }

    var body: some View {
        List {
            Section {
                Text(release.name)
                    .font(.headline)
                    .textSelection(.enabled)
            }

            Section("Release") {
                row("Indexer", release.indexer.name.isEmpty ? release.indexer.identifier : release.indexer.name)
                row("Filter", release.filter)
                row("Received", release.timestamp.formatted(date: .abbreviated, time: .standard))
                row("Source", release.implementation.uppercased())
                row("Category", release.category)
                if release.size > 0 {
                    row("Size", ByteCountFormatter.string(fromByteCount: Int64(clamping: release.size), countStyle: .file))
                }
            }

            Section("Parsed") {
                row("Title", release.title)
                if release.year > 0 { row("Year", String(release.year)) }
                if release.season > 0 { row("Season", String(release.season)) }
                if release.episode > 0 { row("Episode", String(release.episode)) }
                row("Resolution", release.resolution)
                row("Source", release.source)
                row("Codec", (release.codec ?? []).joined(separator: ", "))
                row("HDR", (release.hdr ?? []).joined(separator: ", "))
                row("Container", release.container)
                row("Group", release.group)
                row("Origin", release.origin)
                row("Uploader", release.uploader)
            }

            Section {
                Button("Add to vocabulary") { addingVocab = true }
            } footer: {
                Text("Teach the parser a group or tag it missed in this name.")
            }

            if let rejections = release.rejections, !rejections.isEmpty {
                Section("Filter rejections") {
                    ForEach(rejections, id: \.self) { Text($0).font(.footnote) }
                }
            }

            if let actions = release.actionStatus, !actions.isEmpty {
                Section("Actions") {
                    ForEach(actions) { status in
                        actionRow(status)
                    }
                }
            }

            if let url = URL(string: release.infoUrl), url.scheme?.hasPrefix("http") == true {
                Section {
                    Link("Open on indexer", destination: url)
                }
            }
        }
        .navigationTitle(release.title.isEmpty ? "Release" : release.title)
        .navigationBarTitleDisplayMode(.inline)
        .sheet(isPresented: $addingVocab, onDismiss: { Task { await reload() } }) {
            VocabAddSheet(client: client, releaseName: release.name)
        }
        .alert("Retry", isPresented: Binding(get: { message != nil }, set: { if !$0 { message = nil } }), presenting: message) { _ in
            Button("OK") { message = nil }
        } message: {
            Text($0)
        }
    }

    /// reload picks up the release's parsed fields again, in case its name was parsed again.
    private func reload() async {
        guard let fresh: Release = try? await client.get("release/\(release.id)") else { return }

        release.takeParsed(from: fresh)
    }

    @ViewBuilder
    private func row(_ label: String, _ value: String) -> some View {
        if !value.isEmpty {
            LabeledContent(label, value: value)
                .textSelection(.enabled)
        }
    }

    private func actionRow(_ status: ActionStatus) -> some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack {
                StatusBadge(status: status.pushStatus)
                Text(status.action).font(.subheadline)
                Spacer()
                if retrying == status.id {
                    ProgressView()
                } else {
                    Button("Retry") { retry(status) }
                        .buttonStyle(.borderless)
                        .disabled(retrying != nil)
                }
            }

            Text([status.pushStatus?.label ?? status.status, status.client, status.timestamp.formatted(date: .abbreviated, time: .shortened)]
                .filter { !$0.isEmpty }
                .joined(separator: " · "))
                .font(.caption)
                .foregroundStyle(.secondary)

            ForEach(status.rejections ?? [], id: \.self) { rejection in
                Text(rejection)
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
        }
        .padding(.vertical, 2)
    }

    private func retry(_ status: ActionStatus) {
        retrying = status.id

        Task {
            defer { retrying = nil }

            do {
                try await client.request(.post, "release/\(release.id)/actions/\(status.id)/retry")
                message = "\(status.action) was run again. Pull to refresh the list to see the result."
            } catch {
                message = error.localizedDescription
            }
        }
    }
}
