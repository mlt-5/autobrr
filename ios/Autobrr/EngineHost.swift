// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import AutobrrEngine
import Foundation
import UIKit

/// EngineHost starts the embedded Go engine off the main thread and stops it when the app terminates.
@MainActor
final class EngineHost: ObservableObject {
    enum State {
        case starting
        case running(MobileEngine)
        case failed(String)
    }

    @Published private(set) var state: State = .starting

    private var engine: MobileEngine?
    private var terminateObserver: NSObjectProtocol?

    func start() {
        guard engine == nil else { return }

        state = .starting

        Task.detached(priority: .userInitiated) {
            let result = Self.startEngine()

            await MainActor.run {
                switch result {
                case .success(let engine):
                    self.engine = engine
                    self.state = .running(engine)
                    self.observeTermination()
                case .failure(let error):
                    self.state = .failed(error.localizedDescription)
                }
            }
        }
    }

    private nonisolated static func startEngine() -> Result<MobileEngine, Error> {
        do {
            let dir = try configDirectory()

            var error: NSError?
            guard let engine = MobileStart(dir.path, &error) else {
                return .failure(error ?? CocoaError(.featureUnsupported))
            }

            return .success(engine)
        } catch {
            return .failure(error)
        }
    }

    private nonisolated static func configDirectory() throws -> URL {
        let dir = try FileManager.default
            .url(for: .applicationSupportDirectory, in: .userDomainMask, appropriateFor: nil, create: true)
            .appendingPathComponent("autobrr", isDirectory: true)

        try FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)

        return dir
    }

    private func observeTermination() {
        terminateObserver = NotificationCenter.default.addObserver(
            forName: UIApplication.willTerminateNotification,
            object: nil,
            queue: .main
        ) { [weak self] _ in
            MainActor.assumeIsolated {
                try? self?.engine?.stop()
            }
        }
    }
}
