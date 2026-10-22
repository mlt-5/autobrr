// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import SwiftUI

struct LogsView: View {
    @StateObject private var model: LogsModel
    @State private var showFiles = false

    init(client: EngineClient) {
        _model = StateObject(wrappedValue: LogsModel(client: client))
    }

    var body: some View {
        NavigationStack {
            ScrollViewReader { proxy in
                List(model.visibleLines) { line in
                    LogRow(line: line)
                        .listRowInsets(EdgeInsets(top: 2, leading: 12, bottom: 2, trailing: 12))
                        .id(line.id)
                }
                .listStyle(.plain)
                .onChange(of: model.lines.last?.id) { id in
                    guard model.follow, let id else { return }
                    proxy.scrollTo(id, anchor: .bottom)
                }
            }
            .overlay {
                if model.visibleLines.isEmpty {
                    Text(model.lines.isEmpty ? "Waiting for log messages…" : "No messages match the filter")
                        .font(.callout)
                        .foregroundStyle(.secondary)
                }
            }
            .searchable(text: $model.filter, prompt: "Filter (regex)")
            .autocorrectionDisabled()
            .textInputAutocapitalization(.never)
            .navigationTitle("Logs")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .navigationBarLeading) {
                    Circle()
                        .fill(model.isConnected ? Color.green : Color.secondary)
                        .frame(width: 8, height: 8)
                        .accessibilityLabel(model.isConnected ? "Connected" : "Disconnected")
                }

                ToolbarItemGroup(placement: .navigationBarTrailing) {
                    Button {
                        model.follow.toggle()
                    } label: {
                        Image(systemName: model.follow ? "arrow.down.to.line.circle.fill" : "arrow.down.to.line.circle")
                    }
                    .accessibilityLabel("Follow new messages")

                    Menu {
                        Button {
                            showFiles = true
                        } label: {
                            Label("Log files", systemImage: "doc.text")
                        }

                        Button(role: .destructive) {
                            model.clear()
                        } label: {
                            Label("Clear", systemImage: "trash")
                        }
                    } label: {
                        Image(systemName: "ellipsis.circle")
                    }
                }
            }
            .sheet(isPresented: $showFiles) {
                LogFilesView(model: model)
            }
            .task { await model.streamLogs() }
        }
    }
}

private struct LogRow: View {
    let line: LogLine

    var body: some View {
        (Text(line.time + " ").foregroundColor(.secondary)
            + Text(line.level + " ").foregroundColor(color).bold()
            + Text(line.message))
            .font(.system(.caption2, design: .monospaced))
            .textSelection(.enabled)
    }

    private var color: Color {
        switch line.level.uppercased() {
        case "TRC", "TRACE": .purple
        case "DBG", "DEBUG": .yellow
        case "INF", "INFO": .green
        case "WRN", "WARN": .orange
        case "ERR", "ERROR", "FTL", "FATAL", "PNC", "PANIC": .red
        default: .secondary
        }
    }
}

private struct LogFilesView: View {
    @ObservedObject var model: LogsModel
    @Environment(\.dismiss) private var dismiss

    @State private var files: [LogFile] = []
    @State private var exported: [String: URL] = [:]
    @State private var exporting: String?
    @State private var error: String?

    var body: some View {
        NavigationStack {
            List {
                if let error {
                    Text(error).foregroundStyle(.red)
                }

                ForEach(files) { file in
                    HStack {
                        VStack(alignment: .leading, spacing: 2) {
                            Text(file.name)
                            Text("\(ByteCountFormatter.string(fromByteCount: file.sizeBytes, countStyle: .file)) · \(file.updatedAt.formatted(date: .abbreviated, time: .shortened))")
                                .font(.caption)
                                .foregroundStyle(.secondary)
                        }

                        Spacer()

                        if let url = exported[file.name] {
                            ShareLink(item: url) { Image(systemName: "square.and.arrow.up") }
                        } else if exporting == file.name {
                            ProgressView()
                        } else {
                            Button("Prepare") { export(file) }
                                .buttonStyle(.borderless)
                                .disabled(exporting != nil)
                        }
                    }
                }
            }
            .overlay {
                if files.isEmpty && error == nil {
                    Text("No log files. Set a log path in Settings → Application to write them.")
                        .font(.callout)
                        .foregroundStyle(.secondary)
                        .multilineTextAlignment(.center)
                        .padding()
                }
            }
            .navigationTitle("Log files")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .confirmationAction) {
                    Button("Done") { dismiss() }
                }
            }
            .task {
                do {
                    files = try await model.files()
                } catch {
                    self.error = error.localizedDescription
                }
            }
        }
    }

    private func export(_ file: LogFile) {
        exporting = file.name

        Task {
            defer { exporting = nil }

            do {
                exported[file.name] = try await model.export(file)
            } catch {
                self.error = error.localizedDescription
            }
        }
    }
}
