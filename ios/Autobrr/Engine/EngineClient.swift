// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import AutobrrEngine
import Foundation

/// EngineClient calls the engine's API in-process and decodes its JSON responses.
final class EngineClient {
    enum Method: String {
        case get = "GET"
        case post = "POST"
        case put = "PUT"
        case patch = "PATCH"
        case delete = "DELETE"
    }

    private let engine: MobileEngine

    init(engine: MobileEngine) {
        self.engine = engine
    }

    func get<Response: Decodable>(_ path: String, query: [URLQueryItem] = []) async throws -> Response {
        try JSON.decoder.decode(Response.self, from: await request(.get, path, query: query))
    }

    func send<Response: Decodable>(_ method: Method, _ path: String, body: some Encodable) async throws -> Response {
        try JSON.decoder.decode(Response.self, from: await request(method, path, body: JSON.encoder.encode(body)))
    }

    func send(_ method: Method, _ path: String, body: some Encodable) async throws {
        try await request(method, path, body: JSON.encoder.encode(body))
    }

    @discardableResult
    func request(_ method: Method, _ path: String, query: [URLQueryItem] = [], body: Data? = nil) async throws -> Data {
        var components = URLComponents()
        components.scheme = SchemeHandler.scheme
        components.host = "app"
        components.path = "/api/" + path
        components.queryItems = query.isEmpty ? nil : query

        guard let url = components.url else { throw URLError(.badURL) }

        let collector = ResponseCollector()

        let (status, data) = try await withTaskCancellationHandler {
            try await withCheckedThrowingContinuation { continuation in
                collector.begin(continuation) { writer in
                    engine.serve(
                        method.rawValue,
                        rawURL: url.absoluteString,
                        headersJSON: body == nil ? "" : #"{"Content-Type":"application/json"}"#,
                        body: body,
                        w: writer
                    )
                }
            }
        } onCancel: {
            collector.cancel()
        }

        guard (200..<300).contains(status) else {
            throw APIError(status: status, data: data)
        }

        return data
    }
}

/// APIError is a response from the engine with a status outside 2xx.
struct APIError: LocalizedError {
    let status: Int
    let message: String

    init(status: Int, data: Data) {
        struct Body: Decodable {
            let message: String
        }

        self.status = status
        self.message = (try? JSONDecoder().decode(Body.self, from: data).message)
            ?? String(data: data, encoding: .utf8)?.trimmingCharacters(in: .whitespacesAndNewlines)
            ?? HTTPURLResponse.localizedString(forStatusCode: status)
    }

    var errorDescription: String? { message }
}

enum JSON {
    static let decoder: JSONDecoder = {
        let decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase
        decoder.dateDecodingStrategy = .custom { decoder in
            let container = try decoder.singleValueContainer()
            let value = try container.decode(String.self)

            guard let date = parseDate(value) else {
                throw DecodingError.dataCorruptedError(in: container, debugDescription: "invalid date \(value)")
            }

            return date
        }
        return decoder
    }()

    static let encoder: JSONEncoder = {
        let encoder = JSONEncoder()
        encoder.keyEncodingStrategy = .convertToSnakeCase
        encoder.dateEncodingStrategy = .iso8601
        return encoder
    }()

    private static let wholeSeconds: ISO8601DateFormatter = {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime]
        return formatter
    }()

    private static let fractionalSeconds: ISO8601DateFormatter = {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        return formatter
    }()

    static func parseDate(_ value: String) -> Date? {
        if let date = wholeSeconds.date(from: value) {
            return date
        }

        // Go writes up to nine fractional digits, the formatter reads at most three
        let trimmed = value.replacingOccurrences(of: #"(\.\d{3})\d+"#, with: "$1", options: .regularExpression)

        return fractionalSeconds.date(from: trimmed)
    }
}

/// ResponseCollector buffers one engine response and resumes its continuation when the engine finishes it.
private final class ResponseCollector: NSObject, MobileResponseWriterProtocol {
    private let lock = NSLock()
    private var continuation: CheckedContinuation<(Int, Data), Error>?
    private var request: MobileRequest?
    private var cancelled = false
    private var status = 0
    private var data = Data()

    func begin(_ continuation: CheckedContinuation<(Int, Data), Error>, serve: (ResponseCollector) -> MobileRequest?) {
        lock.withLock { self.continuation = continuation }

        let request = serve(self)

        let cancelNow = lock.withLock {
            self.request = request
            return cancelled
        }

        if cancelNow {
            request?.cancel()
        }
    }

    func cancel() {
        let request = lock.withLock {
            cancelled = true
            return self.request
        }

        request?.cancel()
    }

    func writeHeader(_ status: Int, headersJSON: String?) {
        lock.withLock { self.status = status }
    }

    func write(_ data: Data?) {
        guard let data else { return }

        lock.withLock { self.data.append(data) }
    }

    func finish() {
        let (continuation, cancelled, status, data) = lock.withLock {
            defer { self.continuation = nil }
            return (self.continuation, self.cancelled, self.status, self.data)
        }

        if cancelled {
            continuation?.resume(throwing: CancellationError())
        } else {
            continuation?.resume(returning: (status, data))
        }
    }
}
