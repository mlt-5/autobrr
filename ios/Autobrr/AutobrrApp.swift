// Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

import SwiftUI

@main
struct AutobrrApp: App {
    @StateObject private var host = EngineHost()

    var body: some Scene {
        WindowGroup {
            ContentView()
                .environmentObject(host)
                .task { host.start() }
        }
    }
}

struct ContentView: View {
    @EnvironmentObject private var host: EngineHost

    var body: some View {
        switch host.state {
        case .starting:
            // ProgressView's UIActivityIndicatorView crashed the iOS 27 simulator at launch
            Text("Starting autobrr…")
                .foregroundStyle(.secondary)
        case .failed(let message):
            VStack(spacing: 12) {
                Text("autobrr could not start").font(.headline)
                Text(message).font(.footnote).foregroundStyle(.secondary)
                Button("Try again") { host.start() }
            }
            .padding()
        case .running(let engine):
            WebView(handler: SchemeHandler(engine: engine))
                .ignoresSafeArea(.container, edges: .bottom)
        }
    }
}
