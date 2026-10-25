// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import SwiftUI

private let kindTitles = [
    "arch": "Architecture", "audio": "Audio", "bitdepth": "Bit depth", "channels": "Audio channels",
    "codec": "Codec", "collection": "Streaming service", "container": "Container", "cut": "Cut",
    "edition": "Edition", "ext": "File extension", "genre": "Genre", "group": "Release group", "hdr": "HDR",
    "language": "Language", "other": "Other", "platform": "Platform", "region": "Region",
    "resolution": "Resolution", "size": "Size", "source": "Source",
]

private func kindTitle(_ kind: String) -> String {
    kindTitles[kind] ?? kind
}

struct VocabEntry: Decodable, Identifiable {
    enum CodingKeys: String, CodingKey {
        case recordID = "id"
        case source, kind, tag, title, aliases, pattern, releaseType, exclusive, removed
    }

    let recordID: Int64?
    let source: String
    let kind: String
    let tag: String
    let title: String
    let aliases: [String]
    let pattern: String
    let releaseType: String
    let exclusive: Bool
    let removed: Bool

    var id: String { "\(source)/\(kind)/\(tag)" }

    var summary: String {
        ([kindTitle(kind), title] + aliases).filter { !$0.isEmpty }.joined(separator: " · ")
    }

    func matches(_ search: String) -> Bool {
        search.isEmpty || [tag, title, pattern].contains { $0.localizedCaseInsensitiveContains(search) }
            || aliases.contains { $0.localizedCaseInsensitiveContains(search) }
    }
}

private struct VocabList: Decodable {
    let entries: [VocabEntry]
    let kinds: [String]
    let releaseTypes: [String]
}

private struct VocabKey: Encodable {
    let kind: String
    let tag: String
}

private struct NewVocabEntry: Encodable {
    var kind = "group"
    var tag = ""
    var title = ""
    var aliases: [String] = []
    var pattern = ""
    var releaseType = ""
    var exclusive = false
}

@MainActor
final class VocabularyModel: ObservableObject {
    @Published private(set) var entries: [VocabEntry] = []
    @Published private(set) var kinds: [String] = []
    @Published private(set) var releaseTypes: [String] = []
    @Published private(set) var hasLoaded = false
    @Published var error: String?

    private let client: EngineClient

    init(client: EngineClient) {
        self.client = client
    }

    func load() async {
        do {
            let list: VocabList = try await client.get("vocab")
            entries = list.entries
            kinds = list.kinds
            releaseTypes = list.releaseTypes
        } catch is CancellationError {
            return
        } catch {
            self.error = error.localizedDescription
        }

        hasLoaded = true
    }

    fileprivate func add(_ entry: NewVocabEntry) async throws {
        let _: VocabEntry = try await client.send(.post, "vocab", body: entry)
        await load()
    }

    func delete(_ entry: VocabEntry) async {
        guard let id = entry.recordID else { return }

        await change { _ = try await self.client.request(.delete, "vocab/\(id)") }
    }

    func remove(_ entry: VocabEntry) async {
        await change { try await self.client.send(.post, "vocab/removed", body: VocabKey(kind: entry.kind, tag: entry.tag)) }
    }

    func restore(_ entry: VocabEntry) async {
        let query = [URLQueryItem(name: "kind", value: entry.kind), URLQueryItem(name: "tag", value: entry.tag)]
        await change { _ = try await self.client.request(.delete, "vocab/removed", query: query) }
    }

    private func change(_ request: @escaping () async throws -> Void) async {
        do {
            try await request()
            await load()
        } catch {
            self.error = error.localizedDescription
        }
    }
}

struct VocabularyView: View {
    @StateObject private var model: VocabularyModel
    @State private var search = ""
    @State private var kind = ""
    @State private var adding = false

    init(client: EngineClient) {
        _model = StateObject(wrappedValue: VocabularyModel(client: client))
    }

    private var shown: [VocabEntry] {
        model.entries.filter { (kind.isEmpty || $0.kind == kind) && $0.matches(search) }
    }

    var body: some View {
        let visible = shown
        let added = visible.filter { $0.source == "custom" }
        let listed = visible.filter { $0.source == "list" && !$0.removed }
        let removed = visible.filter { $0.source == "list" && $0.removed }
        let builtin = visible.filter { $0.source == "builtin" }

        List {
            if !added.isEmpty {
                Section {
                    ForEach(added) { entry in
                        row(entry)
                            .swipeActions {
                                Button(role: .destructive) {
                                    Task { await model.delete(entry) }
                                } label: {
                                    Label("Delete", systemImage: "trash")
                                }
                            }
                    }
                } header: {
                    Text("Added")
                }
            }

            if !listed.isEmpty {
                Section {
                    ForEach(listed) { entry in
                        row(entry)
                            .swipeActions {
                                Button(role: .destructive) {
                                    Task { await model.remove(entry) }
                                } label: {
                                    Label("Remove", systemImage: "minus.circle")
                                }
                            }
                    }
                } header: {
                    Text("Lists")
                } footer: {
                    Text("Tags and groups added to the parser for names it didn't read. Removed ones can be restored.")
                }
            }

            if !removed.isEmpty {
                Section("Removed") {
                    ForEach(removed) { entry in
                        row(entry)
                            .swipeActions {
                                Button {
                                    Task { await model.restore(entry) }
                                } label: {
                                    Label("Restore", systemImage: "arrow.uturn.backward")
                                }
                                .tint(.blue)
                            }
                    }
                }
            }

            if !builtin.isEmpty {
                Section {
                    ForEach(builtin) { row($0) }
                } header: {
                    Text("Built in")
                } footer: {
                    Text("The parser's own tags. They can't be removed, but you can add other spellings for them.")
                }
            }
        }
        .searchable(text: $search, prompt: "Tag, title or pattern")
        .overlay {
            if !model.hasLoaded {
                ProgressView()
            } else if visible.isEmpty {
                Text(model.entries.isEmpty ? (model.error ?? "No vocabulary.") : "No entries match.")
                    .font(.callout)
                    .foregroundStyle(.secondary)
                    .multilineTextAlignment(.center)
                    .padding()
            }
        }
        .navigationTitle("Vocabulary")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItemGroup(placement: .navigationBarTrailing) {
                Menu {
                    Picker("Kind", selection: $kind) {
                        Text("All kinds").tag("")
                        ForEach(model.kinds, id: \.self) { Text(kindTitle($0)).tag($0) }
                    }
                } label: {
                    Image(systemName: kind.isEmpty ? "line.3.horizontal.decrease.circle" : "line.3.horizontal.decrease.circle.fill")
                }
                .accessibilityLabel("Kind")

                Button {
                    adding = true
                } label: {
                    Image(systemName: "plus")
                }
                .accessibilityLabel("Add vocabulary entry")
            }
        }
        .sheet(isPresented: $adding) {
            NavigationStack {
                VocabAddView(model: model, kind: kind.isEmpty ? "group" : kind)
            }
        }
        .alert("Vocabulary", isPresented: Binding(get: { model.error != nil && model.hasLoaded && !model.entries.isEmpty }, set: { if !$0 { model.error = nil } })) {
            Button("OK") { model.error = nil }
        } message: {
            Text(model.error ?? "")
        }
        .refreshable { await model.load() }
        .task { await model.load() }
    }

    private func row(_ entry: VocabEntry) -> some View {
        NavigationLink {
            VocabEntryView(entry: entry, model: model)
        } label: {
            VStack(alignment: .leading, spacing: 2) {
                Text(entry.tag)
                Text(entry.summary)
                    .font(.caption)
                    .foregroundStyle(.secondary)
                    .lineLimit(1)
            }
        }
    }
}

private struct VocabEntryView: View {
    let entry: VocabEntry
    @ObservedObject var model: VocabularyModel

    @Environment(\.dismiss) private var dismiss

    var body: some View {
        Form {
            Section {
                LabeledContent("Kind", value: kindTitle(entry.kind))
                LabeledContent("Tag", value: entry.tag)
                if !entry.title.isEmpty {
                    LabeledContent("Title", value: entry.title)
                }
                if !entry.aliases.isEmpty {
                    LabeledContent("Other spellings", value: entry.aliases.joined(separator: ", "))
                }
                LabeledContent("Release type", value: entry.releaseType.isEmpty ? "Any" : entry.releaseType)
                if entry.exclusive {
                    LabeledContent("Only for this type", value: "Yes")
                }
            }

            if !entry.pattern.isEmpty {
                Section("Pattern") {
                    Text(entry.pattern)
                        .font(.footnote.monospaced())
                        .textSelection(.enabled)
                }
            }

            Section {
                switch (entry.source, entry.removed) {
                case ("custom", _):
                    Button("Delete", role: .destructive) { run { await model.delete(entry) } }
                case ("list", false):
                    Button("Remove", role: .destructive) { run { await model.remove(entry) } }
                case ("list", true):
                    Button("Restore") { run { await model.restore(entry) } }
                default:
                    EmptyView()
                }
            } footer: {
                Text(footer)
            }
        }
        .navigationTitle(entry.tag)
        .navigationBarTitleDisplayMode(.inline)
    }

    private var footer: String {
        switch (entry.source, entry.removed) {
        case ("custom", _): return "Added in the app."
        case ("list", false): return "From the lists shipped with the app."
        case ("list", true): return "Removed from the lists shipped with the app. The parser ignores it."
        default: return "Built into the parser."
        }
    }

    private func run(_ change: @escaping () async -> Void) {
        Task {
            await change()
            dismiss()
        }
    }
}

private struct VocabAddView: View {
    @ObservedObject var model: VocabularyModel

    @Environment(\.dismiss) private var dismiss
    @State private var entry: NewVocabEntry
    @State private var aliases = ""
    @State private var busy = false
    @State private var message: String?

    init(model: VocabularyModel, kind: String) {
        self.model = model
        _entry = State(initialValue: NewVocabEntry(kind: kind))
    }

    var body: some View {
        Form {
            Section {
                Picker("Kind", selection: $entry.kind) {
                    ForEach(model.kinds, id: \.self) { Text(kindTitle($0)).tag($0) }
                }
                TextField("Tag", text: $entry.tag)
                    .autocorrectionDisabled()
                    .textInputAutocapitalization(.never)
                TextField("Other spellings", text: $aliases)
                    .autocorrectionDisabled()
                    .textInputAutocapitalization(.never)
            } footer: {
                Text("The tag is how the parser reports it. Separate other spellings with commas; each is reported as the tag. A - . _ or space in either also matches any of those or none, so WEB-DL finds WEB.DL and WEBDL. To add spellings to a built-in tag, give the same tag.")
            }

            Section {
                TextField("Title", text: $entry.title)
                Picker("Release type", selection: $entry.releaseType) {
                    Text("Any").tag("")
                    ForEach(model.releaseTypes, id: \.self) { Text($0).tag($0) }
                }
                Toggle("Only for this type", isOn: $entry.exclusive)
                    .disabled(entry.releaseType.isEmpty)
            } footer: {
                Text("The title is a longer name, such as High Efficiency Video Coding for HEVC. The release type helps the parser guess what a release is; leave it at Any for tags several kinds of release use.")
            }

            Section {
                TextField("Pattern", text: $entry.pattern)
                    .font(.body.monospaced())
                    .autocorrectionDisabled()
                    .textInputAutocapitalization(.never)
            } header: {
                Text("Advanced")
            } footer: {
                Text("A regular expression used instead of the tag, for spellings plain text can't describe. Matching ignores case unless the pattern says (?-i:…).")
            }
        }
        .navigationTitle("New entry")
        .navigationBarTitleDisplayMode(.inline)
        .onChange(of: entry.releaseType) { type in
            if type.isEmpty {
                entry.exclusive = false
            }
        }
        .toolbar {
            ToolbarItem(placement: .cancellationAction) {
                Button("Cancel") { dismiss() }
            }
            ToolbarItem(placement: .confirmationAction) {
                if busy {
                    ProgressView()
                } else {
                    Button("Add") { add() }
                        .disabled(entry.tag.trimmingCharacters(in: .whitespaces).isEmpty)
                }
            }
        }
        .alert("Vocabulary", isPresented: Binding(get: { message != nil }, set: { if !$0 { message = nil } })) {
            Button("OK") { message = nil }
        } message: {
            Text(message ?? "")
        }
    }

    private func add() {
        busy = true
        entry.aliases = aliases.split(separator: ",").map { $0.trimmingCharacters(in: .whitespaces) }

        Task {
            defer { busy = false }

            do {
                try await model.add(entry)
                dismiss()
            } catch {
                message = error.localizedDescription
            }
        }
    }
}
