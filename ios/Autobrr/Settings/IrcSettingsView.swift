// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import SwiftUI

struct IrcSettingsView: View {
    let client: EngineClient

    private static let kind = RecordKind(
        title: "IRC",
        noun: "network",
        path: "irc",
        itemPath: "irc/network",
        empty: "No IRC networks. Adding an indexer that announces on IRC sets up its network.",
        template: [
            "name": .string(""),
            "enabled": .bool(true),
            "server": .string(""),
            "port": .int(6667),
            "tls": .bool(false),
            "tls_skip_verify": .bool(false),
            "pass": .string(""),
            "nick": .string(""),
            "auth": .object(["mechanism": .string("SASL_PLAIN"), "account": .string("")]),
            "channels": .array([]),
        ],
        actions: [RecordAction(title: "Restart connection", method: .get, suffix: "restart", done: "The network is reconnecting.")],
        summary: { network in
            let state: String
            if network["enabled"]?.boolValue == false {
                state = "Disabled"
            } else if network["healthy"]?.boolValue == true {
                state = "Healthy"
            } else if network["connected"]?.boolValue == true {
                state = "Connected"
            } else {
                state = "Not connected"
            }
            let server = network["server"]?.stringValue ?? ""
            return "\(state) · \(server)"
        }
    )

    var body: some View {
        RecordListView(kind: Self.kind, client: client) { draft in
            IrcNetworkForm(draft: draft, client: client)
        }
    }
}

private struct IrcNetworkForm: View {
    @ObservedObject var draft: RecordDraft
    let client: EngineClient

    @State private var proxies: [FilterDraft.Option] = []

    var body: some View {
        if draft.id != nil {
            status
        }

        Section {
            TextRow(label: "Name", text: draft.string("name"))
            Toggle("Enabled", isOn: draft.bool("enabled"))
        }

        Section("Server") {
            TextRow(label: "Server", text: draft.string("server"), prompt: "irc.example.net")
            NumberRow(label: "Port", value: draft.int("port"))
            Toggle("TLS", isOn: draft.bool("tls"))
            if draft.values["tls"]?.boolValue == true {
                Toggle("Skip TLS verification", isOn: draft.bool("tls_skip_verify"))
            }
            SecretRow(label: "Server password", text: draft.string("pass"))
        }

        Section("Identity") {
            TextRow(label: "Nick", text: draft.string("nick"))
            Picker("Authentication", selection: draft.string(in: "auth", "mechanism")) {
                Text("None").tag("NONE")
                Text("SASL (plain)").tag("SASL_PLAIN")
                Text("NickServ").tag("NICKSERV")
            }
            if draft.values["auth"]?.objectValue?["mechanism"]?.stringValue != "NONE" {
                TextRow(label: "Account", text: draft.string(in: "auth", "account"))
                SecretRow(label: "Password", text: draft.string(in: "auth", "password"))
            }
            TextRow(label: "Invite command", text: draft.string("invite_command"), prompt: "Botname !invite KEY", multiline: true)
        }

        Section("Connection") {
            Toggle("Use bouncer", isOn: draft.bool("use_bouncer"))
            if draft.values["use_bouncer"]?.boolValue == true {
                TextRow(label: "Bouncer address", text: draft.string("bouncer_addr"), prompt: "bouncer.example.net:6697")
            }
            Toggle("Bot mode", isOn: draft.bool("bot_mode"))
            Toggle("Use proxy", isOn: draft.bool("use_proxy"))
            if draft.values["use_proxy"]?.boolValue == true {
                Picker("Proxy", selection: draft.int("proxy_id")) {
                    Text("None").tag(0)
                    ForEach(proxies) { Text($0.name).tag(Int($0.id)) }
                }
            }
        }
        .task { proxies = (try? await client.get("proxy")) ?? [] }

        Section {
            ForEach(Array(draft.items("channels").enumerated()), id: \.offset) { index, channel in
                VStack(alignment: .leading, spacing: 6) {
                    HStack {
                        TextField("#channel", text: draft.string("channels", index, "name"))
                            .autocorrectionDisabled()
                            .textInputAutocapitalization(.never)
                        Spacer()
                        if let state = channelState(channel) {
                            Text(state)
                                .font(.caption)
                                .foregroundStyle(.secondary)
                        }
                    }
                    SecureField("Channel password", text: draft.string("channels", index, "password"))
                        .font(.footnote)
                }
            }
            .onDelete { offsets in
                var channels = draft.items("channels")
                channels.remove(atOffsets: offsets)
                draft.setItems("channels", channels)
            }

            Button {
                draft.setItems("channels", draft.items("channels") + [[
                    "name": .string(""),
                    "password": .string(""),
                    "enabled": .bool(true),
                ]])
            } label: {
                Label("Add channel", systemImage: "plus")
            }
        } header: {
            Text("Channels")
        } footer: {
            Text("IRC announces are read only while the app is open.")
        }
    }

    private var status: some View {
        Section("Status") {
            let connected = draft.values["connected"]?.boolValue == true
            LabeledContent("Connection", value: connected ? "Connected" : "Not connected")
            if let nick = draft.values["current_nick"]?.stringValue, !nick.isEmpty {
                LabeledContent("Current nick", value: nick)
            }
            ForEach(draft.values["connection_errors"]?.arrayValue?.compactMap(\.stringValue) ?? [], id: \.self) { error in
                Text(error)
                    .font(.footnote)
                    .foregroundStyle(.red)
            }
        }
    }

    private func channelState(_ channel: [String: JSONValue]) -> String? {
        guard draft.id != nil, channel["id"] != nil else { return nil }

        return channel["monitoring"]?.boolValue == true ? "Monitoring" : "Not monitoring"
    }
}
