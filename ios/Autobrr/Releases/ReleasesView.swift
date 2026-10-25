// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import SwiftUI

struct ReleasesView: View {
    @StateObject private var model: ReleasesModel
    @State private var exporting = false
    @State private var exported: ExportedFile?
    @State private var exportError: String?
    @State private var favoriteError: String?

    init(client: EngineClient) {
        _model = StateObject(wrappedValue: ReleasesModel(client: client))
    }

    var body: some View {
        NavigationStack {
            List {
                Section {
                    ForEach(model.releases) { release in
                        NavigationLink(value: release) {
                            ReleaseRow(release: release)
                        }
                        .swipeActions(edge: .leading) {
                            Button {
                                toggleFavorite(release)
                            } label: {
                                Label(release.favorite ? "Unfavourite" : "Favourite", systemImage: release.favorite ? "star.slash" : "star")
                            }
                            .tint(.yellow)
                        }
                    }

                    if model.hasMore {
                        ProgressView()
                            .frame(maxWidth: .infinity)
                            .listRowSeparator(.hidden)
                            .task { await model.loadMore() }
                    }
                } header: {
                    Picker("Show", selection: $model.query.favorites) {
                        Text("All").tag(false)
                        Text("Favourites").tag(true)
                    }
                    .pickerStyle(.segmented)
                    .textCase(nil)
                    .padding(.vertical, 4)
                }
            }
            .listStyle(.plain)
            .overlay { placeholder }
            .navigationTitle("Releases")
            .navigationDestination(for: Release.self) { release in
                ReleaseDetailView(release: release, client: model.client) { favorite in
                    model.favoriteChanged(id: release.id, favorite)
                }
            }
            .searchable(text: $model.query.search, prompt: "Name, group, codec, resolution…")
            .autocorrectionDisabled()
            .textInputAutocapitalization(.never)
            .toolbar {
                exportButton
                filterMenu
            }
            .refreshable { await model.reload() }
            .task(id: model.query) {
                // The task also runs each time the list reappears, coming back from a release or another tab;
                // reloading then would replace the scrolled list with its first page.
                guard model.loadedQuery != model.query else { return }

                if model.hasLoaded {
                    try? await Task.sleep(nanoseconds: 300_000_000)
                    guard !Task.isCancelled else { return }
                }

                await model.reload()
            }
            .task { await model.loadIndexers() }
            .sheet(item: $exported) { file in
                ShareSheet(items: [file.url])
            }
            .alert("Favourites", isPresented: Binding(get: { favoriteError != nil }, set: { if !$0 { favoriteError = nil } })) {
                Button("OK") { favoriteError = nil }
            } message: {
                Text(favoriteError ?? "")
            }
            .alert("Export", isPresented: Binding(get: { exportError != nil }, set: { if !$0 { exportError = nil } })) {
                Button("OK") { exportError = nil }
            } message: {
                Text(exportError ?? "")
            }
        }
    }

    private var exportButton: some ToolbarContent {
        ToolbarItem(placement: .navigationBarTrailing) {
            if exporting {
                ProgressView()
            } else {
                Button {
                    export()
                } label: {
                    Image(systemName: "square.and.arrow.up")
                }
                .accessibilityLabel("Export to Excel")
                .disabled(!model.hasLoaded)
            }
        }
    }

    private func export() {
        guard !model.releases.isEmpty else {
            exportError = "No releases match the current search and filters."
            return
        }

        exporting = true

        Task {
            defer { exporting = false }

            do {
                let url = try await model.export()
                exported = ExportedFile(url: url)
            } catch {
                exportError = error.localizedDescription
            }
        }
    }

    @ViewBuilder
    private var placeholder: some View {
        if !model.hasLoaded {
            ProgressView()
        } else if model.releases.isEmpty {
            VStack(spacing: 8) {
                Image(systemName: model.error == nil ? "tray" : "exclamationmark.triangle")
                    .font(.largeTitle)
                    .foregroundStyle(.secondary)
                Text(model.error ?? (model.query.favorites ? "No favourites. Tap the star on a release, or swipe it right, to add it." : "No releases"))
                    .font(.callout)
                    .foregroundStyle(.secondary)
                    .multilineTextAlignment(.center)
            }
            .padding()
        }
    }

    private func toggleFavorite(_ release: Release) {
        Task {
            do {
                try await model.setFavorite(release, !release.favorite)
            } catch {
                favoriteError = error.localizedDescription
            }
        }
    }

    private var filterMenu: some ToolbarContent {
        ToolbarItem(placement: .navigationBarTrailing) {
            Menu {
                Picker("Status", selection: $model.query.status) {
                    Text("Any status").tag(PushStatus?.none)
                    ForEach([PushStatus.approved, .rejected, .error]) { status in
                        Text(status.label).tag(PushStatus?.some(status))
                    }
                }

                Picker("Indexer", selection: $model.query.indexer) {
                    Text("Any indexer").tag(String?.none)
                    ForEach(model.indexers, id: \.self) { indexer in
                        Text(indexer).tag(String?.some(indexer))
                    }
                }
            } label: {
                let filtered = model.query.status != nil || model.query.indexer != nil
                Image(systemName: filtered
                    ? "line.3.horizontal.decrease.circle.fill"
                    : "line.3.horizontal.decrease.circle")
            }
        }
    }
}

private struct ReleaseRow: View {
    let release: Release

    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            HStack(alignment: .firstTextBaseline, spacing: 4) {
                if release.favorite {
                    Image(systemName: "star.fill")
                        .font(.caption)
                        .foregroundStyle(.yellow)
                        .accessibilityLabel("Favourite")
                }
                Text(release.name)
                    .font(.subheadline)
                    .lineLimit(2)
            }

            HStack(spacing: 6) {
                Text(release.indexer.name.isEmpty ? release.indexer.identifier : release.indexer.name)
                Text(release.timestamp.formatted(.relative(presentation: .named)))
                Spacer(minLength: 0)
                ForEach(release.actionStatus ?? []) { status in
                    StatusBadge(status: status.pushStatus)
                }
            }
            .font(.caption)
            .foregroundStyle(.secondary)
        }
        .padding(.vertical, 2)
    }
}

/// StatusBadge shows the outcome of one action run.
struct StatusBadge: View {
    let status: PushStatus?

    var body: some View {
        Image(systemName: symbol)
            .foregroundStyle(color)
            .accessibilityLabel(status?.label ?? "Unknown")
    }

    private var symbol: String {
        switch status {
        case .approved: "checkmark.circle.fill"
        case .rejected: "minus.circle.fill"
        case .error: "exclamationmark.circle.fill"
        case .pending, nil: "clock"
        }
    }

    private var color: Color {
        switch status {
        case .approved: .green
        case .rejected: .orange
        case .error: .red
        case .pending, nil: .secondary
        }
    }
}
