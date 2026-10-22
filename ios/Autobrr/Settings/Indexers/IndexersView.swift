// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import SwiftUI

struct IndexersView: View {
    let client: EngineClient

    @State private var indexers: [IndexerDefinition] = []
    @State private var archived: [IndexerRow] = []
    @State private var deprecations: [String: IndexerDeprecation] = [:]
    @State private var hasLoaded = false
    @State private var adding = false
    @State private var deleting: IndexerDefinition?
    @State private var error: String?

    var body: some View {
        List {
            Section {
                ForEach(indexers) { indexer in
                    NavigationLink {
                        IndexerEditView(client: client, definition: indexer) {
                            Task { await load() }
                        }
                    } label: {
                        row(indexer)
                    }
                    .swipeActions {
                        Button(role: .destructive) {
                            deleting = indexer
                        } label: {
                            Label("Delete", systemImage: "trash")
                        }
                    }
                }
            }

            if !archived.isEmpty {
                DeprecatedIndexersSection(client: client, archived: archived, deprecations: deprecations) {
                    await load()
                }
            }
        }
        .overlay {
            if !hasLoaded {
                ProgressView()
            } else if indexers.isEmpty && archived.isEmpty {
                Text(error ?? "No indexers. Add the trackers you want to watch announces or feeds for.")
                    .font(.callout)
                    .foregroundStyle(.secondary)
                    .multilineTextAlignment(.center)
                    .padding()
            }
        }
        .navigationTitle("Indexers")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .navigationBarTrailing) {
                Button {
                    adding = true
                } label: {
                    Image(systemName: "plus")
                }
                .accessibilityLabel("Add indexer")
            }
        }
        .sheet(isPresented: $adding) {
            NavigationStack {
                IndexerTemplatesView(client: client) {
                    adding = false
                    Task { await load() }
                }
            }
        }
        .confirmationDialog(
            "Delete \(deleting?.name ?? "indexer")?",
            isPresented: Binding(get: { deleting != nil }, set: { if !$0 { deleting = nil } }),
            titleVisibility: .visible,
            presenting: deleting
        ) { indexer in
            Button("Delete", role: .destructive) {
                Task { await delete(indexer) }
            }
        }
        .refreshable { await load() }
        .task { await load() }
    }

    private func row(_ indexer: IndexerDefinition) -> some View {
        HStack {
            VStack(alignment: .leading, spacing: 2) {
                Text(indexer.name)
                Text(indexer.kind)
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            Spacer()

            Toggle("Enabled", isOn: Binding(get: { indexer.enabled ?? false }, set: { enabled in
                Task { await toggle(indexer, enabled) }
            }))
            .labelsHidden()
        }
    }

    private func load() async {
        do {
            async let definitions: [IndexerDefinition] = client.get("indexer")
            async let rows: [IndexerRow] = client.get("indexer/options")
            async let removed: [IndexerDeprecation] = client.get("indexer/deprecations")

            indexers = try await definitions
            archived = try await rows.filter { $0.archived == true }
            deprecations = Dictionary(try await removed.map { ($0.identifier, $0) }, uniquingKeysWith: { first, _ in first })
            error = nil
        } catch is CancellationError {
            return
        } catch {
            self.error = error.localizedDescription
        }

        hasLoaded = true
    }

    private func toggle(_ indexer: IndexerDefinition, _ enabled: Bool) async {
        guard let id = indexer.id else { return }

        do {
            try await client.send(.patch, "indexer/\(id)/enabled", body: ["enabled": enabled])
        } catch {
            self.error = error.localizedDescription
        }

        await load()
    }

    private func delete(_ indexer: IndexerDefinition) async {
        guard let id = indexer.id else { return }

        do {
            try await client.request(.delete, "indexer/\(id)")
        } catch {
            self.error = error.localizedDescription
        }

        await load()
    }
}

/// IndexerFields edits what an indexer and its definition's settings hold, shared by the add and edit screens.
struct IndexerFields: View {
    @ObservedObject var draft: RecordDraft
    let definition: IndexerDefinition
    let client: EngineClient

    @State private var proxies: [FilterDraft.Option] = []

    var body: some View {
        Section {
            if definition.isFeed || draft.id != nil {
                TextRow(label: "Name", text: draft.string("name"))
            }
            TextRow(label: "External identifier", text: draft.string("identifier_external"), prompt: draft.values["name"]?.stringValue ?? "")
            Toggle("Enabled", isOn: draft.bool("enabled"))
        } footer: {
            Text("The external identifier is the name arr apps and cross-seed tools know this indexer by.")
        }

        if definition.implementation == "irc" {
            Section("Base URL") {
                TextRow(label: "URL", text: draft.string("base_url"), prompt: "https://")
                if let urls = definition.urls, urls.count > 1 {
                    Picker("Known URLs", selection: draft.string("base_url")) {
                        if !urls.contains(draft.values["base_url"]?.stringValue ?? "") {
                            Text("Custom").tag(draft.values["base_url"]?.stringValue ?? "")
                        }
                        ForEach(urls, id: \.self) { Text($0).tag($0) }
                    }
                }
            }
        }

        if let settings = definition.settings, !settings.isEmpty {
            Section {
                ForEach(settings, id: \.name) { setting in
                    settingRow(setting, text: draft.string(in: "settings", setting.name))
                }
            } header: {
                Text("Settings")
            } footer: {
                Text("Secrets are needed to download torrents from announces. Leave a saved secret as is to keep it.")
            }
        }

        Section("Proxy") {
            Toggle("Use proxy", isOn: draft.bool("use_proxy"))
            if draft.values["use_proxy"]?.boolValue == true {
                Picker("Proxy", selection: draft.int("proxy_id")) {
                    Text("None").tag(0)
                    ForEach(proxies) { Text($0.name).tag(Int($0.id)) }
                }
            }
        }
        .task { proxies = (try? await client.get("proxy")) ?? [] }
    }
}

@ViewBuilder
func settingRow(_ setting: IndexerDefinition.Setting, text: Binding<String>) -> some View {
    VStack(alignment: .leading, spacing: 2) {
        if setting.isSecret {
            SecretRow(label: setting.label, text: text)
        } else {
            TextRow(label: setting.label, text: text)
        }

        if let help = setting.help, !help.isEmpty {
            Text(help)
                .font(.caption2)
                .foregroundStyle(.secondary)
        }
    }
}

private struct IndexerEditView: View {
    let client: EngineClient
    let definition: IndexerDefinition
    let onSaved: () -> Void

    @StateObject private var draft: RecordDraft

    @Environment(\.dismiss) private var dismiss
    @State private var busy = false
    @State private var message: String?

    init(client: EngineClient, definition: IndexerDefinition, onSaved: @escaping () -> Void) {
        self.client = client
        self.definition = definition
        self.onSaved = onSaved

        var settings: [String: JSONValue] = [:]
        for setting in definition.settings ?? [] {
            settings[setting.name] = .string(setting.value ?? "")
        }

        _draft = StateObject(wrappedValue: RecordDraft([
            "id": .int(definition.id ?? 0),
            "name": .string(definition.name),
            "enabled": .bool(definition.enabled ?? false),
            "identifier": .string(definition.identifier),
            "identifier_external": .string(definition.identifierExternal ?? ""),
            "implementation": .string(definition.implementation),
            "base_url": .string(definition.baseUrl ?? ""),
            "use_proxy": .bool(definition.useProxy ?? false),
            "proxy_id": .int(definition.proxyId ?? 0),
            "settings": .object(settings),
        ]))
    }

    var body: some View {
        Form {
            IndexerFields(draft: draft, definition: definition, client: client)

            if definition.isFeed {
                Section {
                    Text("The feed URL, API key and categories are edited under Settings → Feeds.")
                        .font(.callout)
                        .foregroundStyle(.secondary)
                }
            }

            if definition.hasApi {
                Section {
                    Button("Test API") { testApi() }
                        .disabled(busy)
                }
            }
        }
        .navigationTitle(definition.name)
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .confirmationAction) {
                if busy {
                    ProgressView()
                } else {
                    Button("Save") { save() }
                }
            }
        }
        .alert("Indexer", isPresented: Binding(get: { message != nil }, set: { if !$0 { message = nil } })) {
            Button("OK") { message = nil }
        } message: {
            Text(message ?? "")
        }
    }

    private func save() {
        guard let id = draft.id else { return }

        busy = true

        Task {
            defer { busy = false }

            do {
                try await client.request(.put, "indexer/\(id)", body: try JSONEncoder().encode(draft.values))
                onSaved()
                dismiss()
            } catch {
                message = error.localizedDescription
            }
        }
    }

    private func testApi() {
        guard let id = draft.id else { return }

        let settings = draft.values["settings"]?.objectValue ?? [:]
        var body: [String: JSONValue] = [
            "id": .int(id),
            "api_key": settings["api_key"] ?? .string(""),
        ]
        if let user = settings["api_user"]?.stringValue, !user.isEmpty {
            body["api_user"] = .string(user)
        }

        busy = true

        Task {
            defer { busy = false }

            do {
                try await client.request(.post, "indexer/\(id)/api/test", body: try JSONEncoder().encode(body))
                message = "API test succeeded."
            } catch {
                message = error.localizedDescription
            }
        }
    }
}

private struct DeprecatedIndexersSection: View {
    let client: EngineClient
    let archived: [IndexerRow]
    let deprecations: [String: IndexerDeprecation]
    let reload: () async -> Void

    @State private var message: String?

    var body: some View {
        Section {
            ForEach(archived) { indexer in
                let deprecation = deprecations[indexer.identifier]

                VStack(alignment: .leading, spacing: 4) {
                    Text(indexer.name)
                    Text(deprecation?.reason ?? "Removed from autobrr.")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                    if let count = deprecation?.filterCount, count > 0 {
                        Text("Used by \(count) filter\(count == 1 ? "" : "s")")
                            .font(.caption)
                            .foregroundStyle(.orange)
                    }
                    if let link = deprecation?.issueUrl, let url = URL(string: link) {
                        Link("More info", destination: url)
                            .font(.caption)
                    }
                }
                .swipeActions {
                    Button(role: .destructive) {
                        Task { await run(.delete, "indexer/deprecations/\(indexer.id)", body: nil, done: "\(indexer.name) deleted.") }
                    } label: {
                        Label("Delete", systemImage: "trash")
                    }

                    Button {
                        Task { await prune([indexer.identifier]) }
                    } label: {
                        Label("Remove from filters", systemImage: "line.3.horizontal.decrease.circle")
                    }
                    .tint(.orange)
                }
            }

            Button("Remove all from filters") {
                Task { await prune([]) }
            }
        } header: {
            Text("Deprecated")
        } footer: {
            Text("These trackers no longer have a definition. Remove them from filters, then delete them. Swipe a row for its actions.")
        }
        .alert("Deprecated indexers", isPresented: Binding(get: { message != nil }, set: { if !$0 { message = nil } })) {
            Button("OK") { message = nil }
        } message: {
            Text(message ?? "")
        }
    }

    private func prune(_ identifiers: [String]) async {
        let body = try? JSONEncoder().encode(["identifiers": identifiers])
        await run(.post, "filters/indexers/prune-deprecated", body: body, done: "Removed from filters.")
    }

    private func run(_ method: EngineClient.Method, _ path: String, body: Data?, done: String) async {
        do {
            try await client.request(method, path, body: body)
            message = done
        } catch {
            message = error.localizedDescription
        }

        await reload()
    }
}
