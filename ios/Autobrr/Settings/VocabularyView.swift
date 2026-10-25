// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import SwiftUI

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

    func summary(kindTitle: String) -> String {
        ([kindTitle, title] + aliases).filter { !$0.isEmpty }.joined(separator: " · ")
    }

    func matches(_ search: String) -> Bool {
        search.isEmpty || [tag, title, pattern].contains { $0.localizedCaseInsensitiveContains(search) }
            || aliases.contains { $0.localizedCaseInsensitiveContains(search) }
    }
}

private struct VocabList: Decodable {
    let entries: [VocabEntry]
    let kinds: [String]
    let kindTitles: [String: String]
    let releaseTypes: [String]
}

private struct VocabKey: Encodable {
    let kind: String
    let tag: String
}

/// VocabDraft is an added entry being created or edited; id is set when editing.
struct VocabDraft: Encodable, Equatable {
    var kind = "group"
    var tag = ""
    var title = ""
    var aliases: [String] = []
    var pattern = ""
    var releaseType = ""
    var exclusive = false
    var id: Int64?
}

extension VocabDraft {
    init(editing entry: VocabEntry) {
        self.init(kind: entry.kind, tag: entry.tag, title: entry.title, aliases: entry.aliases, pattern: entry.pattern,
                  releaseType: entry.releaseType, exclusive: entry.exclusive, id: entry.recordID)
    }
}

/// VocabPreview is how the parser reads a release name.
struct VocabPreview: Decodable {
    struct Field: Decodable, Hashable {
        let label: String
        let value: String
    }

    let fields: [Field]
    let unused: [String]
}

private struct Notice {
    let title: String
    let message: String
}

private struct ReparseResult: Decodable {
    let checked: Int
    let updated: Int
}

/// reparseExplanation says what updating existing releases does.
private let reparseExplanation = "Parses the names of stored releases again with the vocabulary. Their parsed fields are replaced, including values indexers announced. Filter results and actions stay as they were."

private struct VocabPreviewRequest: Encodable {
    let name: String
    let entry: VocabDraft?
}

@MainActor
final class VocabularyModel: ObservableObject {
    @Published private(set) var entries: [VocabEntry] = []
    @Published private(set) var kinds: [String] = []
    @Published private(set) var kindTitles: [String: String] = [:]
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
            kindTitles = list.kindTitles
            releaseTypes = list.releaseTypes
        } catch is CancellationError {
            return
        } catch {
            self.error = error.localizedDescription
        }

        hasLoaded = true
    }

    func kindTitle(_ kind: String) -> String {
        kindTitles[kind] ?? kind
    }

    /// export writes the entries of kind (or every kind when empty) that match search to a workbook in the temporary
    /// directory.
    func export(kind: String, search: String) async throws -> URL {
        let query = [URLQueryItem(name: "kind", value: kind), URLQueryItem(name: "q", value: search)]
        let data = try await client.request(.get, "vocab/export", query: query)

        let formatter = DateFormatter()
        formatter.locale = Locale(identifier: "en_US_POSIX")
        formatter.dateFormat = "yyyy-MM-dd"

        let url = FileManager.default.temporaryDirectory.appendingPathComponent("autobrr-vocabulary-\(formatter.string(from: Date())).xlsx")
        try data.write(to: url, options: .atomic)

        return url
    }

    func save(_ draft: VocabDraft) async throws {
        if let id = draft.id {
            let _: VocabEntry = try await client.send(.put, "vocab/\(id)", body: draft)
        } else {
            let _: VocabEntry = try await client.send(.post, "vocab", body: draft)
        }
        await load()
    }

    /// reparse parses the names of the stored releases again with the saved vocabulary and returns a summary.
    func reparse() async throws -> String {
        let data = try await client.request(.post, "release/reparse")
        let result = try JSON.decoder.decode(ReparseResult.self, from: data)
        return "Updated \(result.updated) of \(result.checked) releases."
    }

    /// preview parses name with the saved vocabulary plus draft, without saving anything.
    func preview(name: String, draft: VocabDraft?) async throws -> VocabPreview {
        try await client.send(.post, "vocab/preview", body: VocabPreviewRequest(name: name, entry: draft))
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
    @State private var exporting = false
    @State private var exported: ExportedFile?
    @State private var reparsing = false
    @State private var notice: Notice?

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
            Section {
                Button {
                    reparse()
                } label: {
                    HStack {
                        Text("Update existing releases")
                        Spacer()
                        if reparsing {
                            ProgressView()
                        }
                    }
                }
                .disabled(reparsing)
            } footer: {
                Text(reparseExplanation)
            }

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
                        ForEach(model.kinds, id: \.self) { Text(model.kindTitle($0)).tag($0) }
                    }
                } label: {
                    Image(systemName: kind.isEmpty ? "line.3.horizontal.decrease.circle" : "line.3.horizontal.decrease.circle.fill")
                }
                .accessibilityLabel("Kind")

                if exporting {
                    ProgressView()
                } else {
                    Button {
                        export(isEmpty: visible.isEmpty)
                    } label: {
                        Image(systemName: "square.and.arrow.up")
                    }
                    .accessibilityLabel("Export to Excel")
                    .disabled(!model.hasLoaded)
                }

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
                VocabFormView(model: model, draft: VocabDraft(kind: kind.isEmpty ? "group" : kind))
            }
        }
        .sheet(item: $exported) { file in
            ShareSheet(items: [file.url])
        }
        .alert(notice?.title ?? "", isPresented: Binding(get: { notice != nil }, set: { if !$0 { notice = nil } })) {
            Button("OK") { notice = nil }
        } message: {
            Text(notice?.message ?? "")
        }
        .alert("Vocabulary", isPresented: Binding(get: { model.error != nil && model.hasLoaded && !model.entries.isEmpty }, set: { if !$0 { model.error = nil } })) {
            Button("OK") { model.error = nil }
        } message: {
            Text(model.error ?? "")
        }
        .refreshable { await model.load() }
        .task { await model.load() }
    }

    private func reparse() {
        reparsing = true

        Task {
            defer { reparsing = false }

            do {
                let summary = try await model.reparse()
                notice = Notice(title: "Existing releases", message: summary)
            } catch {
                notice = Notice(title: "Existing releases", message: error.localizedDescription)
            }
        }
    }

    private func export(isEmpty: Bool) {
        guard !isEmpty else {
            notice = Notice(title: "Export", message: "No entries match the current search and kind.")
            return
        }

        exporting = true

        Task {
            defer { exporting = false }

            do {
                let url = try await model.export(kind: kind, search: search)
                exported = ExportedFile(url: url)
            } catch {
                notice = Notice(title: "Export", message: error.localizedDescription)
            }
        }
    }

    private func row(_ entry: VocabEntry) -> some View {
        NavigationLink {
            VocabEntryView(entry: entry, model: model)
        } label: {
            VStack(alignment: .leading, spacing: 2) {
                Text(entry.tag)
                Text(entry.summary(kindTitle: model.kindTitle(entry.kind)))
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
    @State private var draft: VocabDraft?
    @State private var saved = false

    var body: some View {
        Form {
            Section {
                LabeledContent("Kind", value: model.kindTitle(entry.kind))
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

            if entry.source != "custom" && !entry.removed {
                Section {
                    Button("Add other spellings") { draft = VocabDraft(kind: entry.kind, tag: entry.tag) }
                } footer: {
                    Text("Adds an entry with the same tag, so the parser also reads other spellings as \(entry.tag).")
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
        .toolbar {
            if entry.source == "custom" {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button("Edit") { draft = VocabDraft(editing: entry) }
                }
            }
        }
        .sheet(isPresented: Binding(get: { draft != nil }, set: { if !$0 { draft = nil } }), onDismiss: {
            if saved {
                dismiss()
            }
        }) {
            if let draft {
                NavigationStack {
                    VocabFormView(model: model, draft: draft) { saved = true }
                }
            }
        }
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

/// VocabFormView adds or edits an entry. With a release name it offers the name's words as the tag and shows how the
/// parser reads the name with the entry.
struct VocabFormView: View {
    @ObservedObject var model: VocabularyModel
    var releaseName: String?
    var onSaved: () -> Void = {}

    @Environment(\.dismiss) private var dismiss
    @State private var draft: VocabDraft
    @State private var aliases: String
    @State private var preview: VocabPreview?
    @State private var previewError: String?
    @State private var busy = false
    @State private var prompt: Prompt?

    private enum Prompt {
        case error(String)
        case updateReleases
        case updated(String)

        var title: String {
            switch self {
            case .error: return "Vocabulary"
            case .updateReleases: return "Update existing releases?"
            case .updated: return "Existing releases"
            }
        }

        var message: String {
            switch self {
            case .error(let message), .updated(let message): return message
            case .updateReleases: return reparseExplanation
            }
        }
    }

    init(model: VocabularyModel, draft: VocabDraft, releaseName: String? = nil, onSaved: @escaping () -> Void = {}) {
        self.model = model
        self.releaseName = releaseName
        self.onSaved = onSaved
        _draft = State(initialValue: draft)
        _aliases = State(initialValue: draft.aliases.joined(separator: ", "))
    }

    private var current: VocabDraft {
        var current = draft
        current.aliases = aliases.split(separator: ",").map { $0.trimmingCharacters(in: .whitespaces) }.filter { !$0.isEmpty }
        return current
    }

    private var words: [String] {
        var seen = Set<String>()
        return (releaseName ?? "")
            .split(whereSeparator: { " .-_[](){}".contains($0) })
            .map(String.init)
            .filter { $0.count > 1 && seen.insert($0.lowercased()).inserted }
    }

    var body: some View {
        Form {
            if let releaseName {
                Section {
                    Text(releaseName)
                        .font(.footnote)
                        .textSelection(.enabled)
                    ScrollView(.horizontal, showsIndicators: false) {
                        HStack {
                            ForEach(words, id: \.self) { word in
                                Button(word) { draft.tag = word }
                                    .buttonStyle(.bordered)
                                    .tint(draft.tag == word ? Color.accentColor : Color.secondary)
                            }
                        }
                    }
                } header: {
                    Text("Release")
                } footer: {
                    Text("Tap a word to use it as the tag.")
                }
            }

            Section {
                Picker("Kind", selection: $draft.kind) {
                    ForEach(model.kinds, id: \.self) { Text(model.kindTitle($0)).tag($0) }
                }
                TextField("Tag", text: $draft.tag)
                    .autocorrectionDisabled()
                    .textInputAutocapitalization(.never)
                TextField("Other spellings", text: $aliases)
                    .autocorrectionDisabled()
                    .textInputAutocapitalization(.never)
            } footer: {
                Text("The tag is how the parser reports it. Separate other spellings with commas; each is reported as the tag. A - . _ or space in either also matches any of those or none, so WEB-DL finds WEB.DL and WEBDL. To add spellings to a tag the parser already knows, give the same tag.")
            }

            Section {
                TextField("Title", text: $draft.title)
                Picker("Release type", selection: $draft.releaseType) {
                    Text("Any").tag("")
                    ForEach(model.releaseTypes, id: \.self) { Text($0).tag($0) }
                }
                Toggle("Only for this type", isOn: $draft.exclusive)
                    .disabled(draft.releaseType.isEmpty)
            } footer: {
                Text("The title is a longer name, such as High Efficiency Video Coding for HEVC. The release type helps the parser guess what a release is; leave it at Any for tags several kinds of release use.")
            }

            Section {
                TextField("Pattern", text: $draft.pattern)
                    .font(.body.monospaced())
                    .autocorrectionDisabled()
                    .textInputAutocapitalization(.never)
            } header: {
                Text("Advanced")
            } footer: {
                Text("A regular expression used instead of the tag, for spellings plain text can't describe. Matching ignores case unless the pattern says (?-i:…).")
            }

            if releaseName != nil {
                Section {
                    if let previewError {
                        Text(previewError)
                            .font(.footnote)
                            .foregroundStyle(.red)
                    } else if let preview {
                        ForEach(preview.fields, id: \.self) { field in
                            LabeledContent(field.label, value: field.value)
                        }
                        if !preview.unused.isEmpty {
                            LabeledContent("Not recognized", value: preview.unused.joined(separator: ", "))
                        }
                    } else {
                        ProgressView()
                    }
                } header: {
                    Text("Parsed with this entry")
                } footer: {
                    Text("How the parser reads the name with this entry. Releases already received keep what was parsed when they arrived.")
                }
            }
        }
        .navigationTitle(draft.id == nil ? "New entry" : "Edit entry")
        .navigationBarTitleDisplayMode(.inline)
        .onChange(of: draft.releaseType) { type in
            if type.isEmpty {
                draft.exclusive = false
            }
        }
        .task(id: current) { await updatePreview() }
        .toolbar {
            ToolbarItem(placement: .cancellationAction) {
                Button("Cancel") { dismiss() }
            }
            ToolbarItem(placement: .confirmationAction) {
                if busy {
                    ProgressView()
                } else {
                    Button(draft.id == nil ? "Add" : "Save") { save() }
                        .disabled(draft.tag.trimmingCharacters(in: .whitespaces).isEmpty)
                }
            }
        }
        .alert(prompt?.title ?? "", isPresented: Binding(get: { prompt != nil }, set: { if !$0 { prompt = nil } }), presenting: prompt) { prompt in
            switch prompt {
            case .error:
                Button("OK") {}
            case .updateReleases:
                Button("Update") { reparse() }
                Button("Not now", role: .cancel) { finish() }
            case .updated:
                Button("OK") { finish() }
            }
        } message: { prompt in
            Text(prompt.message)
        }
    }

    private func updatePreview() async {
        guard let releaseName else { return }

        do {
            try await Task.sleep(nanoseconds: 300_000_000)
        } catch {
            return
        }

        let entry = current
        do {
            preview = try await model.preview(name: releaseName, draft: entry.tag.trimmingCharacters(in: .whitespaces).isEmpty ? nil : entry)
            previewError = nil
        } catch is CancellationError {
            return
        } catch {
            previewError = error.localizedDescription
        }
    }

    private func save() {
        busy = true
        let entry = current

        Task {
            defer { busy = false }

            do {
                try await model.save(entry)
                prompt = .updateReleases
            } catch {
                prompt = .error(error.localizedDescription)
            }
        }
    }

    private func reparse() {
        busy = true

        Task {
            defer { busy = false }

            do {
                let summary = try await model.reparse()
                prompt = .updated(summary)
            } catch {
                prompt = .updated(error.localizedDescription)
            }
        }
    }

    private func finish() {
        onSaved()
        dismiss()
    }
}

/// VocabAddSheet adds a vocabulary entry from a release name.
struct VocabAddSheet: View {
    let releaseName: String

    @StateObject private var model: VocabularyModel
    @Environment(\.dismiss) private var dismiss

    init(client: EngineClient, releaseName: String) {
        self.releaseName = releaseName
        _model = StateObject(wrappedValue: VocabularyModel(client: client))
    }

    var body: some View {
        NavigationStack {
            if model.kinds.isEmpty {
                Group {
                    if let error = model.error {
                        Text(error)
                            .font(.callout)
                            .foregroundStyle(.secondary)
                            .multilineTextAlignment(.center)
                            .padding()
                    } else {
                        ProgressView()
                    }
                }
                .toolbar {
                    ToolbarItem(placement: .cancellationAction) {
                        Button("Cancel") { dismiss() }
                    }
                }
            } else {
                VocabFormView(model: model, draft: VocabDraft(), releaseName: releaseName)
            }
        }
        .task { await model.load() }
    }
}
