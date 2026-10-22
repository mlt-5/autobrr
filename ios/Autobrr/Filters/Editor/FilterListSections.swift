// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import SwiftUI

private extension View {
    func filterDetail(_ title: String, draft: FilterDraft) -> some View {
        navigationTitle(title)
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .confirmationAction) {
                    FilterSaveButton(draft: draft)
                }
            }
    }
}

struct ActionsSection: View {
    @ObservedObject var draft: FilterDraft

    static let types = ["TEST": "Test", "WATCH_FOLDER": "Watch folder", "WEBHOOK": "Webhook"]

    var body: some View {
        Section {
            ForEach(Array(draft.items("actions").enumerated()), id: \.offset) { index, action in
                NavigationLink {
                    ActionDetail(draft: draft, index: index)
                } label: {
                    ItemRow(
                        title: action["name"]?.stringValue ?? "",
                        detail: Self.types[action["type"]?.stringValue ?? ""] ?? "",
                        enabled: action["enabled"]?.boolValue ?? false
                    )
                }
            }
            .onDelete { offsets in
                var actions = draft.items("actions")
                actions.remove(atOffsets: offsets)
                draft.setItems("actions", actions)
            }

            Button {
                draft.setItems("actions", draft.items("actions") + [newAction])
            } label: {
                Label("Add action", systemImage: "plus")
            }
        } footer: {
            Text("Actions run, in order, when a release matches the filter.")
        }
    }

    private var newAction: [String: JSONValue] {
        [
            "id": .int(0),
            "name": .string("new action"),
            "enabled": .bool(true),
            "type": .string("TEST"),
            "watch_folder": .string(""),
            "filter_id": .int(Int64(draft.id)),
            "webhook_host": .string(""),
            "webhook_type": .string(""),
            "webhook_method": .string(""),
            "webhook_data": .string(""),
            "webhook_headers": .array([]),
        ]
    }
}

private struct ActionDetail: View {
    @ObservedObject var draft: FilterDraft
    let index: Int

    var body: some View {
        Form {
            Section {
                TextRow(label: "Name", text: draft.string("actions", index, "name"))
                Toggle("Enabled", isOn: draft.bool("actions", index, "enabled"))
                Picker("Type", selection: draft.string("actions", index, "type")) {
                    ForEach(["TEST", "WATCH_FOLDER", "WEBHOOK"], id: \.self) {
                        Text(ActionsSection.types[$0] ?? $0).tag($0)
                    }
                }
            }

            switch draft.string("actions", index, "type").wrappedValue {
            case "WATCH_FOLDER":
                Section {
                    TextRow(label: "Folder", text: draft.string("actions", index, "watch_folder"), prompt: "/path/to/watch", multiline: true)
                } footer: {
                    Text("The .torrent file is written here for a client that watches the folder.")
                }
            case "WEBHOOK":
                Section {
                    TextRow(label: "Endpoint", text: draft.string("actions", index, "webhook_host"), prompt: "https://example.com/hook")
                    TextRow(label: "Payload", text: draft.string("actions", index, "webhook_data"), prompt: #"{"name": "{{ .TorrentName }}"}"#, multiline: true)
                } footer: {
                    Text("The payload is sent as JSON and may use macros such as {{ .TorrentName }}.")
                }
            default:
                Section {
                    Text("Logs the match without downloading anything.")
                        .foregroundStyle(.secondary)
                }
            }
        }
        .filterDetail(draft.string("actions", index, "name").wrappedValue, draft: draft)
    }
}

struct ExternalSection: View {
    @ObservedObject var draft: FilterDraft

    var body: some View {
        Section {
            ForEach(Array(draft.items("external").enumerated()), id: \.offset) { index, item in
                NavigationLink {
                    ExternalDetail(draft: draft, index: index)
                } label: {
                    ItemRow(
                        title: item["name"]?.stringValue ?? "",
                        detail: item["webhook_host"]?.stringValue ?? "",
                        enabled: item["enabled"]?.boolValue ?? false
                    )
                }
            }
            .onDelete { offsets in
                var items = draft.items("external")
                items.remove(atOffsets: offsets)
                save(items)
            }
            .onMove { from, to in
                var items = draft.items("external")
                items.move(fromOffsets: from, toOffset: to)
                save(items)
            }

            Button {
                let count = draft.items("external").count
                save(draft.items("external") + [[
                    "id": .int(Int64(count + 1)),
                    "index": .int(0),
                    "name": .string("External \(count + 1)"),
                    "enabled": .bool(true),
                    "type": .string("WEBHOOK"),
                    "on_error": .string("REJECT"),
                    "webhook_method": .string("POST"),
                    "webhook_expect_status": .int(200),
                ]])
            } label: {
                Label("Add external filter", systemImage: "plus")
            }
        } footer: {
            Text("A webhook decides whether a matched release is kept. They run top to bottom; tap Edit to reorder.")
        }
        .toolbar {
            ToolbarItem(placement: .navigationBarLeading) {
                EditButton()
            }
        }
    }

    // rows run highest index first, so the order shown is written top-down as n-1...0
    private func save(_ items: [[String: JSONValue]]) {
        let renumbered = items.enumerated().map { position, item in
            var item = item
            item["index"] = .int(Int64(items.count - 1 - position))
            return item
        }

        draft.setItems("external", renumbered)
    }
}

private struct ExternalDetail: View {
    @ObservedObject var draft: FilterDraft
    let index: Int

    var body: some View {
        Form {
            Section {
                TextRow(label: "Name", text: draft.string("external", index, "name"))
                Toggle("Enabled", isOn: draft.bool("external", index, "enabled"))
                Picker("On error", selection: draft.string("external", index, "on_error")) {
                    Text("Reject").tag("REJECT")
                    Text("Continue").tag("CONTINUE")
                }
            }

            Section("Request") {
                TextRow(label: "Endpoint", text: draft.string("external", index, "webhook_host"), prompt: "https://example.com/check")
                Picker("Method", selection: draft.string("external", index, "webhook_method")) {
                    ForEach(["GET", "POST", "PUT", "PATCH", "DELETE"], id: \.self) { Text($0).tag($0) }
                }
                TextRow(label: "Headers", text: draft.string("external", index, "webhook_headers"), prompt: "Authorization=Bearer token;X-Other=value")
                NumberRow(label: "Expected status", value: draft.int("external", index, "webhook_expect_status"))
                TextRow(label: "Payload", text: draft.string("external", index, "webhook_data"), multiline: true)
            }

            Section("Retry") {
                TextRow(label: "Retry on status", text: draft.string("external", index, "webhook_retry_status"), prompt: "500,502,503")
                NumberRow(label: "Attempts", value: draft.int("external", index, "webhook_retry_attempts"))
                NumberRow(label: "Delay (seconds)", value: draft.int("external", index, "webhook_retry_delay_seconds"))
            }
        }
        .filterDetail(draft.string("external", index, "name").wrappedValue, draft: draft)
    }
}

struct NotificationsSection: View {
    @ObservedObject var draft: FilterDraft

    static let events = ["PUSH_APPROVED": "Push approved", "PUSH_REJECTED": "Push rejected", "PUSH_ERROR": "Push error"]

    var body: some View {
        Section {
            ForEach(Array(draft.items("notifications").enumerated()), id: \.offset) { index, item in
                let events = item["events"]?.arrayValue ?? []

                NavigationLink {
                    NotificationDetail(draft: draft, index: index, name: name(of: item))
                } label: {
                    ItemRow(
                        title: name(of: item),
                        detail: events.isEmpty ? "Off for this filter" : "\(events.count) of 3 events",
                        enabled: !events.isEmpty
                    )
                }
            }
            .onDelete { offsets in
                var items = draft.items("notifications")
                items.remove(atOffsets: offsets)
                draft.setItems("notifications", items)
            }

            if !available.isEmpty {
                Menu {
                    ForEach(available) { option in
                        Button(option.name) {
                            draft.setItems("notifications", draft.items("notifications") + [[
                                "notification_id": .int(option.id),
                                "events": .array([.string("PUSH_APPROVED")]),
                            ]])
                        }
                    }
                } label: {
                    Label("Add notification", systemImage: "plus")
                }
            }
        } footer: {
            Text(draft.notificationOptions.isEmpty
                ? "Add a notification service in Settings → Notifications first."
                : "Overrides which events these services send for this filter.")
        }
    }

    private var available: [FilterDraft.NotificationOption] {
        let used = Set(draft.items("notifications").compactMap { $0["notification_id"]?.intValue })
        return draft.notificationOptions.filter { $0.enabled && !used.contains($0.id) }
    }

    private func name(of item: [String: JSONValue]) -> String {
        let id = item["notification_id"]?.intValue
        return draft.notificationOptions.first { $0.id == id }?.name ?? "Notification \(id ?? 0)"
    }
}

private struct NotificationDetail: View {
    @ObservedObject var draft: FilterDraft
    let index: Int
    let name: String

    var body: some View {
        let events = draft.list("notifications", index, "events")

        Form {
            Section {
                Toggle("Send for this filter", isOn: Binding {
                    !events.wrappedValue.isEmpty
                } set: {
                    events.wrappedValue = $0 ? ["PUSH_APPROVED"] : []
                })
            }

            if !events.wrappedValue.isEmpty {
                Section("Events") {
                    ForEach(["PUSH_APPROVED", "PUSH_REJECTED", "PUSH_ERROR"], id: \.self) { event in
                        Toggle(NotificationsSection.events[event] ?? event, isOn: Binding {
                            events.wrappedValue.contains(event)
                        } set: { on in
                            var selected = events.wrappedValue.filter { $0 != event }
                            if on { selected.append(event) }
                            events.wrappedValue = selected
                        })
                    }
                }
            }
        }
        .filterDetail(name, draft: draft)
    }
}

private struct ItemRow: View {
    let title: String
    let detail: String
    let enabled: Bool

    var body: some View {
        HStack {
            VStack(alignment: .leading, spacing: 2) {
                Text(title)
                if !detail.isEmpty {
                    Text(detail)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                        .lineLimit(1)
                }
            }

            Spacer()

            if !enabled {
                Text("Off")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
        }
    }
}
