// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import AutobrrEngine
import SwiftUI

struct FilterEditorView: View {
    let engine: MobileEngine

    @StateObject private var draft: FilterDraft
    @Environment(\.dismiss) private var dismiss

    init(id: Int, engine: MobileEngine, client: EngineClient) {
        self.engine = engine
        _draft = StateObject(wrappedValue: FilterDraft(id: id, client: client))
    }

    var body: some View {
        List {
            Section {
                Toggle("Enabled", isOn: draft.bool("enabled"))
            }

            Section("Rules") {
                page("General", "slider.horizontal.3") { GeneralSection(draft: draft) }
                page("Movies & TV", "film") { MoviesTVSection(draft: draft) }
                page("Music", "music.note") { MusicSection(draft: draft) }
                page("Books", "book") { BooksSection(draft: draft) }
                page("Advanced", "gearshape.2") { AdvancedSection(draft: draft) }
            }

            Section {
                web("External", "external", "arrow.up.forward.app")
                web("Actions", "actions", "bolt")
                web("Notifications", "notifications", "bell")
            } footer: {
                if draft.hasChanges {
                    Text("Save your changes before opening these.")
                }
            }
        }
        .overlay {
            if !draft.hasLoaded { ProgressView() }
        }
        .navigationTitle(draft.name)
        .navigationBarTitleDisplayMode(.inline)
        .navigationBarBackButtonHidden(draft.hasChanges)
        .toolbar {
            if draft.hasChanges {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Discard") { draft.discard() }
                }
            }
            ToolbarItem(placement: .confirmationAction) {
                SaveButton(draft: draft)
            }
        }
        .alert("Filter", isPresented: Binding(get: { draft.error != nil }, set: { if !$0 { draft.error = nil } })) {
            Button("OK") { draft.error = nil }
        } message: {
            Text(draft.error ?? "")
        }
        .task { await draft.load() }
    }

    private func page<Content: View>(_ title: String, _ symbol: String, @ViewBuilder content: @escaping () -> Content) -> some View {
        NavigationLink {
            Form { content() }
                .navigationTitle(title)
                .navigationBarTitleDisplayMode(.inline)
                .toolbar {
                    ToolbarItem(placement: .confirmationAction) {
                        SaveButton(draft: draft)
                    }
                }
        } label: {
            Label(title, systemImage: symbol)
        }
    }

    private func web(_ title: String, _ path: String, _ symbol: String) -> some View {
        NavigationLink {
            WebView(handler: SchemeHandler(engine: engine), path: "/filters/\(draft.id)/\(path)")
                .navigationTitle(title)
                .navigationBarTitleDisplayMode(.inline)
                .onDisappear { Task { await draft.load() } }
        } label: {
            Label(title, systemImage: symbol)
        }
        .disabled(draft.hasChanges)
    }
}

private struct SaveButton: View {
    @ObservedObject var draft: FilterDraft

    var body: some View {
        if draft.isSaving {
            ProgressView()
        } else {
            Button("Save") {
                Task { await draft.save() }
            }
            .disabled(!draft.hasChanges)
        }
    }
}

private struct GeneralSection: View {
    @ObservedObject var draft: FilterDraft

    private static let announceTypes = [
        "NEW": "New", "CHECKED": "Checked", "PROMO": "Promo", "PROMO_GP": "Promo GP", "RESURRECTED": "Resurrected",
    ]

    var body: some View {
        Section {
            TextRow(label: "Name", text: draft.string("name"))
            IndexerSelectRow(draft: draft)
            MultiSelectRow(
                label: "Announce types",
                options: ["NEW", "CHECKED", "PROMO", "PROMO_GP", "RESURRECTED"],
                selection: draft.list("announce_types"),
                titles: Self.announceTypes
            )
        }

        Section {
            TextRow(label: "Min size", text: draft.string("min_size"), prompt: "e.g. 100 MiB")
            TextRow(label: "Max size", text: draft.string("max_size"), prompt: "e.g. 50 GiB")
            NumberRow(label: "Delay (seconds)", value: draft.int("delay"))
            NumberRow(label: "Priority", value: draft.int("priority"))
        } header: {
            Text("Rules")
        } footer: {
            Text("Higher priority filters are checked first. Delay waits before the release is checked.")
        }

        Section("Download limit") {
            NumberRow(label: "Max downloads", value: draft.int("max_downloads"))

            Picker("Per", selection: draft.string("max_downloads_unit")) {
                Text("No limit").tag("")
                ForEach(["MINUTE", "HOUR", "DAY", "WEEK", "MONTH", "EVER"], id: \.self) {
                    Text($0.capitalized).tag($0)
                }
            }

            let unit = draft.string("max_downloads_unit").wrappedValue
            if unit != "" && unit != "EVER" {
                Picker("Window", selection: draft.string("max_downloads_window_type")) {
                    Text("Fixed").tag("FIXED")
                    Text("Rolling").tag("ROLLING")
                }

                if draft.string("max_downloads_window_type").wrappedValue == "ROLLING" {
                    NumberRow(label: "Period", value: draft.int("max_downloads_period"))
                }
            }
        }

        Section("Duplicates") {
            Picker("Skip duplicates", selection: duplicateProfile) {
                Text("Off").tag(Int64?.none)
                ForEach(draft.duplicateProfiles) { profile in
                    Text(profile.name).tag(Int64?.some(profile.id))
                }
            }
        }
    }

    private var duplicateProfile: Binding<Int64?> {
        Binding {
            draft.values["release_profile_duplicate_id"]?.intValue.flatMap { $0 == 0 ? nil : $0 }
        } set: {
            draft.set("release_profile_duplicate_id", $0.map(JSONValue.int) ?? .null)
        }
    }
}

/// IndexerSelectRow picks the filter's indexers. The filter stores whole indexer objects, so the selection keeps them as the engine sent them.
private struct IndexerSelectRow: View {
    @ObservedObject var draft: FilterDraft

    var body: some View {
        NavigationLink {
            List(draft.indexerOptions.indices, id: \.self) { index in
                let option = draft.indexerOptions[index]
                let id = option.objectValue?["id"]?.intValue

                Button {
                    toggle(option, id: id)
                } label: {
                    HStack {
                        Text(option.objectValue?["name"]?.stringValue ?? "").foregroundStyle(Color.primary)
                        Spacer()
                        if selectedIDs.contains(id ?? -1) {
                            Image(systemName: "checkmark").foregroundStyle(Color.accentColor)
                        }
                    }
                }
            }
            .overlay {
                if draft.indexerOptions.isEmpty {
                    Text("No indexers. Add one in Settings → Indexers.")
                        .font(.callout)
                        .foregroundStyle(.secondary)
                }
            }
            .navigationTitle("Indexers")
            .navigationBarTitleDisplayMode(.inline)
        } label: {
            LabeledContent("Indexers") {
                Text(names.isEmpty ? "None" : names.joined(separator: ", ")).lineLimit(1)
            }
        }
    }

    private var selected: [JSONValue] {
        draft.values["indexers"]?.arrayValue ?? []
    }

    private var selectedIDs: Set<Int64> {
        Set(selected.compactMap { $0.objectValue?["id"]?.intValue })
    }

    private var names: [String] {
        selected.compactMap { $0.objectValue?["name"]?.stringValue }
    }

    private func toggle(_ option: JSONValue, id: Int64?) {
        guard let id else { return }

        var indexers = selected
        if selectedIDs.contains(id) {
            indexers.removeAll { $0.objectValue?["id"]?.intValue == id }
        } else {
            indexers.append(option)
        }

        draft.set("indexers", .array(indexers))
    }
}

private struct MoviesTVSection: View {
    @ObservedObject var draft: FilterDraft

    var body: some View {
        Section {
            TextRow(label: "Movies / shows", text: draft.string("shows"), prompt: "That Movie, That Show", multiline: true)
            TextRow(label: "Years", text: draft.string("years"), prompt: "2018,2019-2021")
        } footer: {
            Text("Comma separated. Wildcards * and ? are supported.")
        }

        Section("Seasons and episodes") {
            TextRow(label: "Seasons", text: draft.string("seasons"), prompt: "1,3,5-10")
            TextRow(label: "Episodes", text: draft.string("episodes"), prompt: "1,3,5-10")
            TextRow(label: "Months", text: draft.string("months"), prompt: "1,12")
            TextRow(label: "Days", text: draft.string("days"), prompt: "1,31")
            Toggle("Smart episode", isOn: draft.bool("smart_episode"))
        }

        Section("Quality") {
            MultiSelectRow(label: "Resolutions", options: FilterOptions.resolutions, selection: draft.list("resolutions"))
            MultiSelectRow(label: "Sources", options: FilterOptions.sources, selection: draft.list("sources"))
            MultiSelectRow(label: "Codecs", options: FilterOptions.codecs, selection: draft.list("codecs"))
            MultiSelectRow(label: "Containers", options: FilterOptions.containers, selection: draft.list("containers"))
            MultiSelectRow(label: "Match HDR", options: FilterOptions.hdr, selection: draft.list("match_hdr"))
            MultiSelectRow(label: "Except HDR", options: FilterOptions.hdr, selection: draft.list("except_hdr"))
            MultiSelectRow(label: "Match other", options: FilterOptions.other, selection: draft.list("match_other"))
            MultiSelectRow(label: "Except other", options: FilterOptions.other, selection: draft.list("except_other"))
        }
    }
}

private struct MusicSection: View {
    @ObservedObject var draft: FilterDraft

    var body: some View {
        Section {
            TextRow(label: "Artists", text: draft.string("artists"), multiline: true)
            TextRow(label: "Albums", text: draft.string("albums"), multiline: true)
            TextRow(label: "Match record labels", text: draft.string("match_record_labels"), multiline: true)
            TextRow(label: "Except record labels", text: draft.string("except_record_labels"), multiline: true)
        }

        Section("Release details") {
            MultiSelectRow(label: "Release types", options: FilterOptions.musicReleaseTypes, selection: draft.list("match_release_types"))
            TextRow(label: "Years", text: draft.string("years"), prompt: "2018,2019-2021")
        }

        Section("Quality") {
            MultiSelectRow(label: "Format", options: FilterOptions.musicFormats, selection: draft.list("formats"))
            MultiSelectRow(label: "Quality", options: FilterOptions.musicQuality, selection: draft.list("quality"))
            MultiSelectRow(label: "Media", options: FilterOptions.musicMedia, selection: draft.list("media"))
            Toggle("Cue", isOn: draft.bool("cue"))
            Toggle("Log", isOn: draft.bool("log"))
            NumberRow(label: "Log score", value: draft.int("log_score"))
            Toggle("Perfect FLAC", isOn: draft.bool("perfect_flac"))
        }
    }
}

private struct BooksSection: View {
    @ObservedObject var draft: FilterDraft

    var body: some View {
        Section {
            TextRow(label: "Authors", text: draft.string("artists"), multiline: true)
            TextRow(label: "Titles", text: draft.string("albums"), multiline: true)
            TextRow(label: "Years", text: draft.string("years"), prompt: "2018,2019-2021")
        }

        Section("Options") {
            MultiSelectRow(label: "Match language", options: FilterOptions.languages, selection: draft.list("match_language"))
            MultiSelectRow(label: "Except language", options: FilterOptions.languages, selection: draft.list("except_language"))
            MultiSelectRow(label: "Format", options: FilterOptions.bookFormats, selection: draft.list("containers"))
            TextRow(label: "Match tags", text: draft.string("tags"))
            TextRow(label: "Except tags", text: draft.string("except_tags"))
        }
    }
}

private struct AdvancedSection: View {
    @ObservedObject var draft: FilterDraft

    var body: some View {
        Section("Release names") {
            Toggle("Use regex", isOn: draft.bool("use_regex"))
            TextRow(label: "Match releases", text: draft.string("match_releases"), multiline: true)
            TextRow(label: "Except releases", text: draft.string("except_releases"), multiline: true)
        }

        Section("Groups") {
            TextRow(label: "Match groups", text: draft.string("match_release_groups"), multiline: true)
            TextRow(label: "Except groups", text: draft.string("except_release_groups"), multiline: true)
        }

        Section("Categories") {
            TextRow(label: "Match categories", text: draft.string("match_categories"), multiline: true)
            TextRow(label: "Except categories", text: draft.string("except_categories"), multiline: true)
        }

        Section("Tags") {
            TextRow(label: "Match tags", text: draft.string("tags"), multiline: true)
            logicPicker("Match logic", "tags_match_logic")
            TextRow(label: "Except tags", text: draft.string("except_tags"), multiline: true)
            logicPicker("Except logic", "except_tags_match_logic")
        }

        Section("Uploaders") {
            TextRow(label: "Match uploaders", text: draft.string("match_uploaders"), multiline: true)
            TextRow(label: "Except uploaders", text: draft.string("except_uploaders"), multiline: true)
        }

        Section("Language") {
            MultiSelectRow(label: "Match language", options: FilterOptions.languages, selection: draft.list("match_language"))
            MultiSelectRow(label: "Except language", options: FilterOptions.languages, selection: draft.list("except_language"))
        }

        Section("Origins") {
            MultiSelectRow(label: "Match origins", options: FilterOptions.origins, selection: draft.list("origins"))
            MultiSelectRow(label: "Except origins", options: FilterOptions.origins, selection: draft.list("except_origins"))
        }

        Section("Freeleech") {
            Toggle("Freeleech", isOn: draft.bool("freeleech"))
            TextRow(label: "Freeleech percent", text: draft.string("freeleech_percent"), prompt: "50%,75-100%")
        }

        Section("Feeds") {
            Toggle("Use regex", isOn: draft.bool("use_regex_description"))
            TextRow(label: "Match description", text: draft.string("match_description"), multiline: true)
            TextRow(label: "Except description", text: draft.string("except_description"), multiline: true)
            NumberRow(label: "Min seeders", value: draft.int("min_seeders"))
            NumberRow(label: "Max seeders", value: draft.int("max_seeders"))
            NumberRow(label: "Min leechers", value: draft.int("min_leechers"))
            NumberRow(label: "Max leechers", value: draft.int("max_leechers"))
        }

        Section("Raw release tags") {
            Toggle("Use regex", isOn: draft.bool("use_regex_release_tags"))
            TextRow(label: "Match release tags", text: draft.string("match_release_tags"))
            TextRow(label: "Except release tags", text: draft.string("except_release_tags"))
        }
    }

    private func logicPicker(_ label: String, _ key: String) -> some View {
        Picker(label, selection: draft.string(key)) {
            Text("Any").tag("ANY")
            Text("All").tag("ALL")
        }
    }
}
