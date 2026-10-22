// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import AutobrrEngine
import SwiftUI

struct FiltersView: View {
    let engine: MobileEngine

    @StateObject private var model: FiltersModel
    @State private var path: [Int] = []
    @State private var showAdd = false
    @State private var deleting: FilterSummary?

    init(engine: MobileEngine, client: EngineClient) {
        self.engine = engine
        _model = StateObject(wrappedValue: FiltersModel(client: client))
    }

    var body: some View {
        NavigationStack(path: $path) {
            List(model.visibleFilters) { filter in
                NavigationLink(value: filter.id) {
                    FilterRow(filter: filter) { enabled in
                        Task { await model.setEnabled(filter, enabled) }
                    }
                }
                .swipeActions {
                    Button(role: .destructive) {
                        deleting = filter
                    } label: {
                        Label("Delete", systemImage: "trash")
                    }

                    Button {
                        Task { await model.duplicate(filter) }
                    } label: {
                        Label("Duplicate", systemImage: "plus.square.on.square")
                    }
                    .tint(.indigo)
                }
            }
            .listStyle(.plain)
            .overlay { placeholder }
            .navigationTitle("Filters")
            .navigationDestination(for: Int.self) { id in
                FilterEditorView(id: id, engine: engine, client: model.client)
                    .onDisappear { Task { await model.load() } }
            }
            .searchable(text: $model.search, prompt: "Search filters")
            .toolbar {
                ToolbarItemGroup(placement: .navigationBarTrailing) {
                    Menu {
                        Picker("Sort", selection: $model.sort) {
                            ForEach(FiltersModel.Sort.allCases) { Text($0.label).tag($0) }
                        }
                    } label: {
                        Image(systemName: "arrow.up.arrow.down.circle")
                    }

                    Button {
                        showAdd = true
                    } label: {
                        Image(systemName: "plus")
                    }
                    .accessibilityLabel("Add filter")
                }
            }
            .refreshable { await model.load() }
            .task(id: model.sort) { await model.load() }
            .sheet(isPresented: $showAdd) {
                AddFilterView { name in
                    if let id = await model.create(name: name) {
                        path.append(id)
                    }
                }
            }
            .confirmationDialog(
                "Delete \(deleting?.name ?? "filter")?",
                isPresented: Binding(get: { deleting != nil }, set: { if !$0 { deleting = nil } }),
                titleVisibility: .visible,
                presenting: deleting
            ) { filter in
                Button("Delete", role: .destructive) {
                    Task { await model.delete(filter) }
                }
            } message: { _ in
                Text("The filter and its actions are removed.")
            }
            .alert("Filters", isPresented: Binding(get: { model.error != nil }, set: { if !$0 { model.error = nil } })) {
                Button("OK") { model.error = nil }
            } message: {
                Text(model.error ?? "")
            }
        }
    }

    @ViewBuilder
    private var placeholder: some View {
        if !model.hasLoaded {
            ProgressView()
        } else if model.visibleFilters.isEmpty {
            Text(model.filters.isEmpty ? "No filters yet. Tap + to add one." : "No filters match the search")
                .font(.callout)
                .foregroundStyle(.secondary)
        }
    }
}

private struct FilterRow: View {
    let filter: FilterSummary
    let onToggle: (Bool) -> Void

    var body: some View {
        HStack {
            VStack(alignment: .leading, spacing: 4) {
                Text(filter.name)
                    .font(.body)
                    .lineLimit(1)

                Text(details)
                    .font(.caption)
                    .foregroundStyle(.secondary)
                    .lineLimit(1)
            }

            Spacer()

            Toggle("Enabled", isOn: Binding(get: { filter.enabled }, set: onToggle))
                .labelsHidden()
        }
    }

    private var details: String {
        var parts = ["\(filter.actionsEnabledCount)/\(filter.actionsCount) actions"]

        if filter.priority != 0 {
            parts.append("priority \(filter.priority)")
        }

        let indexers = (filter.indexers ?? []).map(\.name)
        parts.append(indexers.isEmpty ? "no indexers" : indexers.joined(separator: ", "))

        return parts.joined(separator: " · ")
    }
}

private struct AddFilterView: View {
    let onCreate: (String) async -> Void

    @Environment(\.dismiss) private var dismiss
    @State private var name = ""
    @State private var creating = false
    @FocusState private var focused: Bool

    var body: some View {
        NavigationStack {
            Form {
                TextField("Name", text: $name)
                    .focused($focused)
                    .submitLabel(.done)
                    .onSubmit(create)
            }
            .navigationTitle("New filter")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Create", action: create)
                        .disabled(trimmed.isEmpty || creating)
                }
            }
            .onAppear { focused = true }
        }
        .presentationDetents([.medium])
    }

    private var trimmed: String {
        name.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    private func create() {
        guard !trimmed.isEmpty, !creating else { return }

        creating = true

        Task {
            dismiss()
            await onCreate(trimmed)
        }
    }
}
