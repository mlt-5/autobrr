// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import Foundation

struct LogLine: Decodable, Identifiable {
    let time: String
    let level: String
    let message: String

    var id = UUID()

    private enum CodingKeys: String, CodingKey {
        case time, level, message
    }
}

struct LogFile: Decodable, Identifiable {
    let name: String
    let sizeBytes: Int64
    let updatedAt: Date

    var id: String { name }
}

@MainActor
final class LogsModel: ObservableObject {
    private static let maxLines = 2000

    private struct Files: Decodable {
        let files: [LogFile]
    }

    let client: EngineClient

    @Published private(set) var lines: [LogLine] = []
    @Published private(set) var isConnected = false
    @Published var filter = ""
    @Published var follow = true

    init(client: EngineClient) {
        self.client = client
    }

    var visibleLines: [LogLine] {
        let filter = filter.trimmingCharacters(in: .whitespaces)
        guard !filter.isEmpty else { return lines }

        if let regex = try? NSRegularExpression(pattern: filter, options: .caseInsensitive) {
            return lines.filter { line in
                regex.firstMatch(in: line.message, range: NSRange(line.message.startIndex..., in: line.message)) != nil
            }
        }

        return lines.filter { $0.message.localizedCaseInsensitiveContains(filter) }
    }

    func clear() {
        lines.removeAll()
    }

    /// streamLogs keeps the log stream open, reconnecting with a growing delay, until the calling task is cancelled.
    func streamLogs() async {
        var attempt = 0

        while !Task.isCancelled {
            do {
                var parser = EventStreamParser()

                for try await chunk in client.stream("events", query: [URLQueryItem(name: "stream", value: "logs")]) {
                    isConnected = true
                    attempt = 0

                    let received = parser.append(chunk).compactMap { event in
                        try? JSONDecoder().decode(LogLine.self, from: Data(event.utf8))
                    }

                    lines.append(contentsOf: received)
                    if lines.count > Self.maxLines {
                        lines.removeFirst(lines.count - Self.maxLines)
                    }
                }
            } catch {}

            isConnected = false
            guard !Task.isCancelled else { return }

            try? await Task.sleep(nanoseconds: UInt64(min(30, 1 << min(attempt, 5))) * 1_000_000_000)
            attempt += 1
        }
    }

    func files() async throws -> [LogFile] {
        let response: Files = try await client.get("logs/files")
        return response.files.sorted { $0.updatedAt > $1.updatedAt }
    }

    /// export writes a sanitized copy of a log file to a temporary file for sharing.
    func export(_ file: LogFile) async throws -> URL {
        let data = try await client.request(.get, "logs/files/\(file.name)")
        let url = FileManager.default.temporaryDirectory.appendingPathComponent(file.name)

        try data.write(to: url, options: .atomic)

        return url
    }
}
