// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import AutobrrEngine
import Foundation
import WebKit

/// SchemeHandler serves every autobrr:// request from the embedded engine, so the web UI never touches the network stack.
final class SchemeHandler: NSObject, WKURLSchemeHandler {
    static let scheme = "autobrr"

    static func url(path: String) -> URL {
        URL(string: "\(scheme)://app\(path)")!
    }

    private let engine: MobileEngine
    private var active: [ObjectIdentifier: (writer: TaskWriter, request: MobileRequest?)] = [:]

    init(engine: MobileEngine) {
        self.engine = engine
    }

    func webView(_ webView: WKWebView, start task: WKURLSchemeTask) {
        let request = task.request

        guard let url = request.url else {
            task.didFailWithError(URLError(.badURL))
            return
        }

        let headers = request.allHTTPHeaderFields ?? [:]
        let headersJSON = (try? JSONSerialization.data(withJSONObject: headers))
            .flatMap { String(data: $0, encoding: .utf8) } ?? ""

        let id = ObjectIdentifier(task)
        let writer = TaskWriter(task: task, url: url) { [weak self] in
            self?.active[id] = nil
        }

        let served = engine.serve(
            request.httpMethod ?? "GET",
            rawURL: url.absoluteString,
            headersJSON: headersJSON,
            body: Self.body(of: request),
            w: writer
        )

        // writer callbacks hop to the main thread, so none can run before this entry exists
        active[id] = (writer, served)
    }

    func webView(_ webView: WKWebView, stop task: WKURLSchemeTask) {
        guard let entry = active.removeValue(forKey: ObjectIdentifier(task)) else { return }

        entry.writer.stop()
        entry.request?.cancel()
    }

    private static func body(of request: URLRequest) -> Data? {
        if let body = request.httpBody {
            return body
        }

        guard let stream = request.httpBodyStream else { return nil }

        stream.open()
        defer { stream.close() }

        var data = Data()
        var buffer = [UInt8](repeating: 0, count: 16 * 1024)

        while stream.hasBytesAvailable {
            let read = stream.read(&buffer, maxLength: buffer.count)
            if read <= 0 { break }
            data.append(buffer, count: read)
        }

        return data
    }
}

/// TaskWriter forwards an engine response to its scheme task. The engine calls it from a background goroutine,
/// and WebKit raises if a task is used after it was stopped, so every call hops to the main thread and checks first.
private final class TaskWriter: NSObject, MobileResponseWriterProtocol {
    private let task: WKURLSchemeTask
    private let url: URL
    private let onDone: () -> Void
    private var stopped = false

    init(task: WKURLSchemeTask, url: URL, onDone: @escaping () -> Void) {
        self.task = task
        self.url = url
        self.onDone = onDone
    }

    func stop() {
        stopped = true
    }

    func writeHeader(_ status: Int, headersJSON: String?) {
        let headers = headersJSON
            .flatMap { $0.data(using: .utf8) }
            .flatMap { try? JSONSerialization.jsonObject(with: $0) as? [String: String] } ?? [:]

        DispatchQueue.main.async { [self] in
            guard !stopped else { return }

            let response = HTTPURLResponse(url: url, statusCode: status, httpVersion: "HTTP/1.1", headerFields: headers)!
            task.didReceive(response)
        }
    }

    func write(_ data: Data?) {
        guard let data else { return }

        DispatchQueue.main.async { [self] in
            guard !stopped else { return }

            task.didReceive(data)
        }
    }

    func finish() {
        DispatchQueue.main.async { [self] in
            guard !stopped else { return }

            stopped = true
            task.didFinish()
            onDone()
        }
    }
}
