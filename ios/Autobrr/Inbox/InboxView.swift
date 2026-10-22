// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import SwiftUI

struct InboxView: View {
    @ObservedObject var model: InboxModel
    let client: EngineClient

    @State private var confirmingDeleteAll = false

    var body: some View {
        List {
            Section {
                Picker("Show", selection: $model.unreadOnly) {
                    Text("All (\(model.allCount))").tag(false)
                    Text("Unread (\(model.unreadCount))").tag(true)
                }
                .pickerStyle(.segmented)
                .listRowBackground(Color.clear)
                .listRowInsets(EdgeInsets())
            }

            Section {
                ForEach(model.messages) { message in
                    NavigationLink {
                        InboxMessageView(model: model, client: client, message: message)
                    } label: {
                        InboxRow(message: message)
                    }
                    .swipeActions {
                        Button(role: .destructive) {
                            Task { await model.delete([message.id]) }
                        } label: {
                            Label("Delete", systemImage: "trash")
                        }

                        if message.isUnread {
                            Button {
                                Task { await model.markRead([message.id]) }
                            } label: {
                                Label("Read", systemImage: "envelope.open")
                            }
                            .tint(.blue)
                        }
                    }
                    .onAppear {
                        if message.id == model.messages.last?.id {
                            Task { await model.loadMore() }
                        }
                    }
                }
            }
        }
        .overlay {
            if !model.hasLoaded {
                ProgressView()
            } else if model.messages.isEmpty {
                Text(model.error ?? (model.unreadOnly ? "No unread notifications." : "No notifications. Choose which events reach the inbox under Settings → Notifications → Built-in."))
                    .font(.callout)
                    .foregroundStyle(.secondary)
                    .multilineTextAlignment(.center)
                    .padding()
            }
        }
        .navigationTitle("Inbox")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .navigationBarTrailing) {
                Menu {
                    Button {
                        Task { await model.markRead([]) }
                    } label: {
                        Label("Mark all read", systemImage: "envelope.open")
                    }
                    .disabled(model.unreadCount == 0)

                    Button(role: .destructive) {
                        confirmingDeleteAll = true
                    } label: {
                        Label("Delete all", systemImage: "trash")
                    }
                    .disabled(model.allCount == 0)
                } label: {
                    Image(systemName: "ellipsis.circle")
                }
            }
        }
        .confirmationDialog("Delete all notifications?", isPresented: $confirmingDeleteAll, titleVisibility: .visible) {
            Button("Delete all", role: .destructive) {
                Task { await model.deleteAll() }
            }
        }
        .refreshable { await model.load() }
        .task {
            await model.load()
            await model.watch()
        }
    }
}

private struct InboxRow: View {
    let message: InboxMessage

    var body: some View {
        HStack(alignment: .top, spacing: 10) {
            Circle()
                .fill(message.isUnread ? Color.accentColor : .clear)
                .frame(width: 8, height: 8)
                .padding(.top, 6)

            VStack(alignment: .leading, spacing: 3) {
                HStack {
                    Text(message.title)
                        .fontWeight(message.isUnread ? .semibold : .regular)
                        .lineLimit(1)
                    Spacer()
                    Text(message.createdAt.formatted(.relative(presentation: .named)))
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }

                if let release = message.releaseName, !release.isEmpty {
                    Text(release)
                        .font(.caption)
                        .lineLimit(1)
                }

                let detail = message.message.isEmpty ? (message.rejections ?? []).joined(separator: ", ") : message.message
                if !detail.isEmpty {
                    Text(detail)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                        .lineLimit(2)
                }
            }
        }
    }
}

private struct InboxMessageView: View {
    @ObservedObject var model: InboxModel
    let client: EngineClient
    let message: InboxMessage

    @Environment(\.dismiss) private var dismiss

    var body: some View {
        Form {
            if !message.message.isEmpty {
                Section {
                    Text(message.message)
                        .textSelection(.enabled)
                }
            }

            Section {
                row("Event", NotificationsView.events[message.event] ?? message.event)
                row("Received", message.createdAt.formatted(date: .abbreviated, time: .standard))
                row("Release", message.releaseName)
                row("Indexer", message.indexer)
                row("Filter", message.filterName)
                row("Action", [message.action, message.actionClient].compactMap { $0 }.filter { !$0.isEmpty }.joined(separator: " · "))
            }

            if let rejections = message.rejections, !rejections.isEmpty {
                Section("Rejections") {
                    ForEach(rejections, id: \.self) { Text($0).font(.callout) }
                }
            }

            Section {
                if let id = message.filterId, id > 0 {
                    NavigationLink("Open filter") {
                        FilterEditorView(id: id, client: client)
                    }
                }

                if let link = message.url, let url = URL(string: link), !link.isEmpty {
                    Link("Open link", destination: url)
                }

                Button("Delete", role: .destructive) {
                    Task {
                        await model.delete([message.id])
                        dismiss()
                    }
                }
            }
        }
        .navigationTitle(message.title)
        .navigationBarTitleDisplayMode(.inline)
        .task {
            if message.isUnread {
                await model.markRead([message.id])
            }
        }
    }

    @ViewBuilder
    private func row(_ label: String, _ value: String?) -> some View {
        if let value, !value.isEmpty {
            LabeledContent(label) {
                Text(value)
                    .multilineTextAlignment(.trailing)
                    .textSelection(.enabled)
            }
        }
    }
}
