// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import SwiftUI

struct IndexerTemplatesView: View {
    let client: EngineClient
    let onCreated: () -> Void

    @Environment(\.dismiss) private var dismiss
    @State private var templates: [IndexerDefinition] = []
    @State private var search = ""
    @State private var error: String?

    private var matches: [IndexerDefinition] {
        guard !search.isEmpty else { return templates }
        return templates.filter { $0.name.localizedCaseInsensitiveContains(search) || $0.identifier.localizedCaseInsensitiveContains(search) }
    }

    var body: some View {
        List(matches, id: \.identifier) { template in
            NavigationLink {
                AddIndexerView(client: client, definition: template, onCreated: onCreated)
            } label: {
                VStack(alignment: .leading, spacing: 2) {
                    Text(template.name)
                    Text([template.kind, template.description ?? ""].filter { !$0.isEmpty }.joined(separator: " · "))
                        .font(.caption)
                        .foregroundStyle(.secondary)
                        .lineLimit(1)
                }
            }
        }
        .overlay {
            if templates.isEmpty {
                if let error {
                    Text(error).foregroundStyle(.secondary).padding()
                } else {
                    ProgressView()
                }
            }
        }
        .searchable(text: $search, placement: .navigationBarDrawer(displayMode: .always))
        .navigationTitle("Add indexer")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .cancellationAction) {
                Button("Cancel") { dismiss() }
            }
        }
        .task {
            do {
                templates = try await client.get("indexer/schema")
            } catch {
                self.error = error.localizedDescription
            }
        }
    }
}

/// AddIndexerView creates an indexer and, like the web form, the IRC network or feed its definition needs.
private struct AddIndexerView: View {
    private struct Created: Decodable {
        let id: Int64
    }

    let client: EngineClient
    let definition: IndexerDefinition
    let onCreated: () -> Void

    @StateObject private var draft: RecordDraft
    @State private var busy = false
    @State private var message: String?

    init(client: EngineClient, definition: IndexerDefinition, onCreated: @escaping () -> Void) {
        self.client = client
        self.definition = definition
        self.onCreated = onCreated

        var irc: [String: JSONValue] = [:]
        for setting in definition.irc?.settings ?? [] {
            irc[setting.name] = .string(setting.default ?? "")
        }

        _draft = StateObject(wrappedValue: RecordDraft([
            "enabled": .bool(true),
            "identifier": .string(definition.identifier),
            "identifier_external": .string(""),
            "implementation": .string(definition.implementation),
            "name": .string(definition.isFeed ? "" : definition.name),
            "base_url": .string(definition.implementation == "irc" ? definition.urls?.first ?? "" : ""),
            "use_proxy": .bool(false),
            "proxy_id": .int(0),
            "settings": .object([:]),
            "irc": .object(irc),
            "feed": .object(["download_type": .string("TORRENT")]),
        ]))
    }

    var body: some View {
        Form {
            if let description = definition.description, !description.isEmpty {
                Section {
                    Text(description)
                        .font(.callout)
                        .foregroundStyle(.secondary)
                }
            }

            IndexerFields(draft: draft, definition: definition, client: client)

            if definition.implementation == "irc", let settings = definition.irc?.settings, !settings.isEmpty {
                Section {
                    ForEach(settings, id: \.name) { setting in
                        settingRow(setting, text: draft.string(in: "irc", setting.name))
                    }
                } header: {
                    Text("IRC")
                } footer: {
                    Text("Creates the \(definition.irc?.network ?? "") network, disabled. Register the nick on the tracker's network first, then enable it under Settings → IRC.")
                }
            }

            if definition.isFeed {
                Section {
                    ForEach(definition.feed?.settings ?? [], id: \.name) { setting in
                        settingRow(setting, text: draft.string(in: "feed", setting.name))
                    }
                    Picker("Download", selection: draft.string(in: "feed", "download_type")) {
                        Text("Torrent").tag("TORRENT")
                        Text("Magnet").tag("MAGNET")
                    }
                } header: {
                    Text("Feed")
                } footer: {
                    Text("Creates the feed, disabled. Pick categories and enable it under Settings → Feeds.")
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
                    Button("Create") { create() }
                }
            }
        }
        .alert("Indexer", isPresented: Binding(get: { message != nil }, set: { if !$0 { message = nil } })) {
            Button("OK") { message = nil }
        } message: {
            Text(message ?? "")
        }
    }

    private func value(_ group: String, _ name: String) -> String {
        draft.values[group]?.objectValue?[name]?.stringValue ?? ""
    }

    /// missing names the first required field left empty or at its placeholder default, as the web form rejects.
    private var missing: String? {
        if definition.isFeed && (draft.values["name"]?.stringValue ?? "").isEmpty {
            return "Name"
        }

        let groups: [(String, [IndexerDefinition.Setting])] = [
            ("settings", definition.settings ?? []),
            ("irc", definition.implementation == "irc" ? definition.irc?.settings ?? [] : []),
            ("feed", definition.isFeed ? definition.feed?.settings ?? [] : []),
        ]

        for (group, settings) in groups {
            for setting in settings where setting.required == true {
                let text = value(group, setting.name)
                if text.isEmpty || (setting.default.map { !$0.isEmpty && $0 == text } ?? false) {
                    return setting.label
                }
            }
        }

        return nil
    }

    private func create() {
        if let missing {
            message = "\(missing) is required."
            return
        }

        busy = true

        Task {
            defer { busy = false }

            do {
                let indexer = draft.values.filter { $0.key != "irc" && $0.key != "feed" }
                let data = try await client.request(.post, "indexer", body: try JSONEncoder().encode(indexer))
                let created = try JSON.decoder.decode(Created.self, from: data)

                if definition.implementation == "irc" {
                    try await client.request(.post, "irc", body: try JSONEncoder().encode(network()))
                } else if definition.isFeed {
                    try await client.request(.post, "feeds", body: try JSONEncoder().encode(feed(indexerID: created.id)))
                }

                onCreated()
            } catch {
                message = error.localizedDescription
            }
        }
    }

    private func network() -> [String: JSONValue] {
        let irc = definition.irc
        let channelPassword = value("irc", "channels.password")

        return [
            "name": .string(irc?.network ?? ""),
            "enabled": .bool(false),
            "connected": .bool(false),
            "server": .string(irc?.server ?? ""),
            "port": .int(Int64(irc?.port ?? 0)),
            "tls": .bool(irc?.tls ?? false),
            "tls_skip_verify": .bool(false),
            "pass": .string(value("irc", "pass")),
            "nick": .string(value("irc", "nick")),
            "auth": .object(auth()),
            "invite_command": .string(value("irc", "invite_command")),
            "channels": .array((irc?.channels ?? []).map { channel in
                JSONValue.object([
                    "id": .int(0),
                    "enabled": .bool(true),
                    "name": .string(channel.name),
                    "password": .string(channelPassword),
                    "detached": .bool(false),
                    "monitoring": .bool(false),
                ])
            }),
        ]
    }

    private func auth() -> [String: JSONValue] {
        let mechanism = definition.irc?.auth?.mechanism ?? "SASL_PLAIN"
        let account = value("irc", "auth.account")
        let password = value("irc", "auth.password")

        if mechanism == "NICKSERV" && !password.isEmpty {
            var auth: [String: JSONValue] = ["mechanism": .string(mechanism), "password": .string(password)]
            if !account.isEmpty {
                auth["account"] = .string(account)
            }
            return auth
        }

        if mechanism == "SASL_PLAIN" && !account.isEmpty && !password.isEmpty {
            return ["mechanism": .string(mechanism), "account": .string(account), "password": .string(password)]
        }

        return ["mechanism": .string("NONE")]
    }

    private func feed(indexerID: Int64) -> [String: JSONValue] {
        let torznab = definition.implementation == "torznab"

        var feed: [String: JSONValue] = [
            "name": .string(draft.values["name"]?.stringValue ?? ""),
            "enabled": .bool(false),
            "type": .string(torznab ? "TORZNAB" : "RSS"),
            "url": .string(value("feed", "url")),
            "interval": .int(30),
            "timeout": .int(60),
            "indexer_id": .int(indexerID),
            "settings": .object(["download_type": .string(value("feed", "download_type"))]),
        ]
        if torznab {
            feed["api_key"] = .string(value("feed", "api_key"))
            feed["categories"] = .array([])
        }
        return feed
    }
}
