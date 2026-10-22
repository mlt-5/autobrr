// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import SwiftUI

/// RecordDraft is one settings record being edited, kept as the engine returns it.
@MainActor
final class RecordDraft: ObservableObject, JSONDocument {
    @Published var values: [String: JSONValue]

    init(_ values: [String: JSONValue]) {
        self.values = values
    }

    var id: Int64? {
        values["id"]?.intValue.flatMap { $0 == 0 ? nil : $0 }
    }
}

/// RecordKind describes a settings list the engine serves as GET/POST on path and PUT/DELETE on path/{id}.
struct RecordKind {
    let title: String
    let noun: String
    let path: String
    let empty: String
    let template: [String: JSONValue]
    var canTest = false
    var canDelete: ([String: JSONValue]) -> Bool = { _ in true }
    let summary: ([String: JSONValue]) -> String
}

struct RecordListView<Fields: View>: View {
    let kind: RecordKind
    let client: EngineClient
    @ViewBuilder let fields: (RecordDraft) -> Fields

    @State private var records: [[String: JSONValue]] = []
    @State private var hasLoaded = false
    @State private var adding = false
    @State private var deleting: [String: JSONValue]?
    @State private var error: String?

    var body: some View {
        List {
            ForEach(Array(records.enumerated()), id: \.offset) { _, record in
                NavigationLink {
                    RecordEditView(kind: kind, client: client, record: record, fields: fields) {
                        Task { await load() }
                    }
                } label: {
                    row(record)
                }
                .swipeActions {
                    if kind.canDelete(record) {
                        Button(role: .destructive) {
                            deleting = record
                        } label: {
                            Label("Delete", systemImage: "trash")
                        }
                    }
                }
            }
        }
        .overlay {
            if !hasLoaded {
                ProgressView()
            } else if records.isEmpty {
                Text(error ?? kind.empty)
                    .font(.callout)
                    .foregroundStyle(.secondary)
                    .multilineTextAlignment(.center)
                    .padding()
            }
        }
        .navigationTitle(kind.title)
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .navigationBarTrailing) {
                Button {
                    adding = true
                } label: {
                    Image(systemName: "plus")
                }
                .accessibilityLabel("Add \(kind.noun)")
            }
        }
        .sheet(isPresented: $adding) {
            NavigationStack {
                RecordEditView(kind: kind, client: client, record: kind.template, fields: fields) {
                    adding = false
                    Task { await load() }
                }
            }
        }
        .confirmationDialog(
            "Delete \(deleting?["name"]?.stringValue ?? kind.noun)?",
            isPresented: Binding(get: { deleting != nil }, set: { if !$0 { deleting = nil } }),
            titleVisibility: .visible,
            presenting: deleting
        ) { record in
            Button("Delete", role: .destructive) {
                Task { await delete(record) }
            }
        }
        .refreshable { await load() }
        .task { await load() }
    }

    private func row(_ record: [String: JSONValue]) -> some View {
        HStack {
            VStack(alignment: .leading, spacing: 2) {
                Text(record["name"]?.stringValue ?? "")
                let summary = kind.summary(record)
                if !summary.isEmpty {
                    Text(summary)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                        .lineLimit(1)
                }
            }

            Spacer()

            if record["enabled"]?.boolValue == false {
                Text("Off")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
        }
    }

    private func load() async {
        do {
            let data = try await client.request(.get, kind.path)
            records = try JSONDecoder().decode([JSONValue].self, from: data).compactMap(\.objectValue)
            error = nil
        } catch is CancellationError {
            return
        } catch {
            self.error = error.localizedDescription
        }

        hasLoaded = true
    }

    private func delete(_ record: [String: JSONValue]) async {
        guard let id = record["id"]?.intValue else { return }

        do {
            try await client.request(.delete, "\(kind.path)/\(id)")
            await load()
        } catch {
            self.error = error.localizedDescription
        }
    }
}

private struct RecordEditView<Fields: View>: View {
    let kind: RecordKind
    let client: EngineClient
    let fields: (RecordDraft) -> Fields
    let onSaved: () -> Void

    @StateObject private var draft: RecordDraft

    @Environment(\.dismiss) private var dismiss
    @State private var busy = false
    @State private var message: String?

    init(kind: RecordKind, client: EngineClient, record: [String: JSONValue], fields: @escaping (RecordDraft) -> Fields, onSaved: @escaping () -> Void) {
        self.kind = kind
        self.client = client
        self.fields = fields
        self.onSaved = onSaved
        _draft = StateObject(wrappedValue: RecordDraft(record))
    }

    var body: some View {
        Form {
            fields(draft)

            if kind.canTest {
                Section {
                    Button("Test") { run(test: true) }
                        .disabled(busy)
                }
            }
        }
        .navigationTitle(draft.id == nil ? "New \(kind.noun)" : draft.values["name"]?.stringValue ?? kind.noun)
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            if draft.id == nil {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
            }

            ToolbarItem(placement: .confirmationAction) {
                if busy {
                    ProgressView()
                } else {
                    Button(draft.id == nil ? "Create" : "Save") { run(test: false) }
                }
            }
        }
        .alert(kind.title, isPresented: Binding(get: { message != nil }, set: { if !$0 { message = nil } })) {
            Button("OK") { message = nil }
        } message: {
            Text(message ?? "")
        }
    }

    private func run(test: Bool) {
        busy = true

        Task {
            defer { busy = false }

            do {
                let body = try JSONEncoder().encode(draft.values)

                if test {
                    try await client.request(.post, "\(kind.path)/test", body: body)
                    message = "Test succeeded."
                } else if let id = draft.id {
                    try await client.request(.put, "\(kind.path)/\(id)", body: body)
                    onSaved()
                    dismiss()
                } else {
                    try await client.request(.post, kind.path, body: body)
                    onSaved()
                }
            } catch {
                message = error.localizedDescription
            }
        }
    }
}
