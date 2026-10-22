// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import SwiftUI

struct AppConfig: Decodable {
    let configDir: String
    let logLevel: String
    let logPath: String
    let version: String
    let commit: String
    let date: String
}

private struct ConfigUpdate: Encodable {
    var logLevel: String?
    var logPath: String?
}

struct ApplicationSettingsView: View {
    private static let levels = ["TRACE", "DEBUG", "INFO", "WARN", "ERROR"]

    let client: EngineClient

    @State private var config: AppConfig?
    @State private var logLevel = "INFO"
    @State private var writeLogFiles = false
    @State private var error: String?

    var body: some View {
        Form {
            if let error {
                Section {
                    Text(error).foregroundStyle(.red)
                }
            }

            Section("About") {
                LabeledContent("Version", value: config?.version ?? "–")
                if let commit = config?.commit, !commit.isEmpty {
                    LabeledContent("Commit", value: String(commit.prefix(12)))
                }
                if let date = config?.date, !date.isEmpty {
                    LabeledContent("Build date", value: date)
                }
                if let dir = config?.configDir {
                    LabeledContent("Data folder", value: dir)
                        .textSelection(.enabled)
                }
            }

            Section {
                Picker("Log level", selection: $logLevel) {
                    ForEach(Self.levels, id: \.self) { Text($0) }
                }
                .onChange(of: logLevel) { level in
                    guard config != nil, level != config?.logLevel else { return }
                    save(ConfigUpdate(logLevel: level))
                }

                Toggle("Write log files", isOn: $writeLogFiles)
                    .onChange(of: writeLogFiles) { enabled in
                        guard let config, enabled != !config.logPath.isEmpty else { return }
                        save(ConfigUpdate(logPath: enabled ? logFilePath(in: config) : ""))
                    }
            } header: {
                Text("Logging")
            } footer: {
                Text("Log files are written to the data folder and can be shared from the Logs tab. Turning them on or off takes effect the next time autobrr starts.")
            }
        }
        .navigationTitle("Application")
        .navigationBarTitleDisplayMode(.inline)
        .task { await load() }
    }

    private func logFilePath(in config: AppConfig) -> String {
        URL(fileURLWithPath: config.configDir).appendingPathComponent("logs/autobrr.log").path
    }

    private func load() async {
        do {
            let config: AppConfig = try await client.get("config")
            logLevel = config.logLevel.uppercased()
            writeLogFiles = !config.logPath.isEmpty
            self.config = config
            error = nil
        } catch {
            self.error = error.localizedDescription
        }
    }

    private func save(_ update: ConfigUpdate) {
        Task {
            do {
                try await client.send(.patch, "config", body: update)
                await load()
            } catch {
                self.error = error.localizedDescription
            }
        }
    }
}
